package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"votespher/pkg"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testSecret = "middleware-test-secret"

func makeToken(t *testing.T, voterID uint, areaID uint, role string) string {
	t.Helper()
	tok, err := pkg.GenerateToken(voterID, areaID, role, testSecret)
	if err != nil {
		t.Fatalf("makeToken: %v", err)
	}
	return tok
}

func newRouter(mw ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(mw...)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// ─── RequireAuth ─────────────────────────────────────────────────────────────

func TestRequireAuth_NoHeader(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", testSecret)
	r := newRouter(RequireAuth())

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_MissingBearer(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", testSecret)
	r := newRouter(RequireAuth())

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", testSecret)
	r := newRouter(RequireAuth())

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", testSecret)
	token := makeToken(t, 7, 2, "voter")
	r := newRouter(RequireAuth())

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAuth_SetsContext(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", testSecret)
	token := makeToken(t, 55, 3, "admin")

	var gotVoterID interface{}
	var gotRole interface{}

	r := gin.New()
	r.Use(RequireAuth())
	r.GET("/test", func(c *gin.Context) {
		gotVoterID, _ = c.Get("voter_id")
		gotRole, _ = c.Get("role")
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	httptest.NewRecorder()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if gotVoterID.(uint) != 55 {
		t.Errorf("expected voter_id=55, got %v", gotVoterID)
	}
	if gotRole.(string) != "admin" {
		t.Errorf("expected role=admin, got %v", gotRole)
	}
}

// ─── RequireRole ─────────────────────────────────────────────────────────────

func TestRequireRole_WrongRole(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", testSecret)
	token := makeToken(t, 1, 1, "voter")

	r := gin.New()
	r.Use(RequireAuth(), RequireRole("admin"))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_CorrectRole(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", testSecret)
	token := makeToken(t, 1, 1, "admin")

	r := gin.New()
	r.Use(RequireAuth(), RequireRole("admin"))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ─── RateLimit ────────────────────────────────────────────────────────────────

func TestRateLimit_AllowsNormalRequests(t *testing.T) {
	// reset limiter state
	limiter = &ipRateLimiter{requests: make(map[string][]time.Time)}

	r := gin.New()
	r.Use(RateLimit())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_BlocksAfterLimit(t *testing.T) {
	// ใช้ IP ใหม่เพื่อไม่กระทบ test อื่น
	limiter = &ipRateLimiter{requests: make(map[string][]time.Time)}

	r := gin.New()
	r.Use(RateLimit())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	// ส่ง request เกิน maxRequests (60) โดยตรง
	ip := "10.99.99.99:0"
	for i := 0; i < maxRequests; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// request ที่ 61 ต้อง 429
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = ip
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after exceeding limit, got %d", w.Code)
	}
}
