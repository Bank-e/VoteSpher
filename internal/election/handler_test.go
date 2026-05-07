package election

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// mockService — implement Service สำหรับทดสอบ Handler
// ============================================================

type mockService struct {
	resp *ConfigResponse
	err  error

	calledWithVoter uint
	calledWithReq   UpdateConfigRequest
}

func (m *mockService) UpdateElectionConfig(_ context.Context, voterID uint, req UpdateConfigRequest) (*ConfigResponse, error) {
	m.calledWithVoter = voterID
	m.calledWithReq = req
	return m.resp, m.err
}

// ============================================================
// helpers
// ============================================================

// setupHandlerRouter สร้าง gin engine ที่ใส่ voter_id ใน context จำลอง middleware
func setupHandlerRouter(svc Service, injectVoterID interface{}) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewHandler(svc)

	// middleware จำลองที่ฉีด voter_id ลง context (เลียนแบบ JWT middleware ของจริง)
	r.PATCH("/election/config", func(c *gin.Context) {
		if injectVoterID != nil {
			c.Set("voter_id", injectVoterID)
		}
		c.Next()
	}, h.UpdateConfig)

	return r
}

func validRequestBody() []byte {
	body, _ := json.Marshal(UpdateConfigRequest{
		Status:    "OPEN",
		StartTime: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 5, 1, 17, 0, 0, 0, time.UTC),
	})
	return body
}

func sendPatch(r *gin.Engine, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/election/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeError(t *testing.T, body []byte) string {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode error response: %v (body=%s)", err, string(body))
	}
	msg, _ := resp["error"].(string)
	return msg
}

// ============================================================
// tests
// ============================================================

func TestHandler_Success(t *testing.T) {
	svc := &mockService{
		resp: &ConfigResponse{
			ConfigID: 42,
			Status:   "OPEN",
			IsActive: true,
		},
	}
	r := setupHandlerRouter(svc, uint(7))

	w := sendPatch(r, validRequestBody())

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}
	if svc.calledWithVoter != 7 {
		t.Errorf("expected service called with voterID=7, got %d", svc.calledWithVoter)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "success" {
		t.Errorf("expected status=success, got %v", resp["status"])
	}
	if resp["data"] == nil {
		t.Error("expected data field in response")
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	r := setupHandlerRouter(svc, uint(7))

	w := sendPatch(r, []byte("{not valid json"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if got := decodeError(t, w.Body.Bytes()); got != ErrInvalidRequest.Error() {
		t.Errorf("expected error %q, got %q", ErrInvalidRequest.Error(), got)
	}
}

func TestHandler_MissingStatus_FailsValidation(t *testing.T) {
	svc := &mockService{}
	r := setupHandlerRouter(svc, uint(7))

	body, _ := json.Marshal(map[string]interface{}{
		"start_time": time.Now(),
		"end_time":   time.Now().Add(time.Hour),
	})

	w := sendPatch(r, body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing status, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestHandler_MissingVoterID(t *testing.T) {
	svc := &mockService{}
	r := setupHandlerRouter(svc, nil) // ไม่ set voter_id

	w := sendPatch(r, validRequestBody())

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if got := decodeError(t, w.Body.Bytes()); got != ErrUnauthorized.Error() {
		t.Errorf("expected %q, got %q", ErrUnauthorized.Error(), got)
	}
}

func TestHandler_VoterIDWrongType(t *testing.T) {
	svc := &mockService{}
	r := setupHandlerRouter(svc, "not-a-uint") // ใส่เป็น string

	w := sendPatch(r, validRequestBody())

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if got := decodeError(t, w.Body.Bytes()); got != ErrInvalidToken.Error() {
		t.Errorf("expected %q, got %q", ErrInvalidToken.Error(), got)
	}
}

func TestHandler_ServiceForbiddenError(t *testing.T) {
	svc := &mockService{err: forbidden(ErrNotAdmin)}
	r := setupHandlerRouter(svc, uint(7))

	w := sendPatch(r, validRequestBody())

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if got := decodeError(t, w.Body.Bytes()); got != ErrNotAdmin.Error() {
		t.Errorf("expected %q, got %q", ErrNotAdmin.Error(), got)
	}
}

func TestHandler_ServiceBadRequestError(t *testing.T) {
	svc := &mockService{err: badRequest(ErrInvalidTimeRange)}
	r := setupHandlerRouter(svc, uint(7))

	w := sendPatch(r, validRequestBody())

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if got := decodeError(t, w.Body.Bytes()); got != ErrInvalidTimeRange.Error() {
		t.Errorf("expected %q, got %q", ErrInvalidTimeRange.Error(), got)
	}
}

func TestHandler_ServiceInternalError(t *testing.T) {
	svc := &mockService{err: internal(ErrConfigNotFound, nil)}
	r := setupHandlerRouter(svc, uint(7))

	w := sendPatch(r, validRequestBody())

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestRespondError_NonAppErrorFallback ตรวจ branch fallback ของ respondError
// — เมื่อได้ error ที่ไม่ใช่ *AppError ต้องตอบ 500 พร้อมข้อความจาก err.Error()
//
// ใช้ test ตรงๆ แทนการบังคับ service คืน plain error เพราะใน flow ปกติ
// service คืน *AppError เสมอ (จึงไม่มีทางเข้า branch นี้ผ่าน HTTP request)
func TestRespondError_NonAppErrorFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	plainErr := errors.New("ฐานข้อมูลล่ม")
	respondError(c, plainErr)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	if got := decodeError(t, w.Body.Bytes()); got != "ฐานข้อมูลล่ม" {
		t.Errorf("expected error %q in response, got %q", "ฐานข้อมูลล่ม", got)
	}
}
