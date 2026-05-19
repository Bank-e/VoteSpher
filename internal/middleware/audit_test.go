package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"votespher/internal/models"
)

func setupAuditDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	db.AutoMigrate(&models.AuditLog{})
	return db
}

func TestAuditLog_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAuditDB(t)

	r := gin.New()
	r.POST("/ballot/submit", func(c *gin.Context) {
		c.Set("voter_id", uint(1))
		c.Next()
	}, AuditLog(db, "SUBMIT_VOTE"), func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ballot/submit", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	time.Sleep(50 * time.Millisecond)

	var count int64
	db.Model(&models.AuditLog{}).Count(&count)
	if count == 0 {
		t.Error("expected audit log to be created")
	}
}

func TestAuditLog_MissingVoterID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAuditDB(t)

	r := gin.New()
	r.GET("/test", AuditLog(db, "TEST_ACTION"), func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuditLog_InvalidVoterIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAuditDB(t)

	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("voter_id", "not-a-uint")
		c.Next()
	}, AuditLog(db, "TEST_ACTION"), func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuditLog_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAuditDB(t)

	sqlDB, _ := db.DB()
	sqlDB.Close()

	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("voter_id", uint(1))
		c.Next()
	}, AuditLog(db, "TEST_ACTION"), func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 even with DB error, got %d", w.Code)
	}
}
