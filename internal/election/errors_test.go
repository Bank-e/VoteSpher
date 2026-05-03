package election

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// TestAppError_Error_WithCause ตรวจกรณีที่ AppError มี inner error
// คาดว่าจะคืน "{message}: {cause}"
func TestAppError_Error_WithCause(t *testing.T) {
	cause := errors.New("db down")
	appErr := &AppError{
		Code:    500,
		Message: "ไม่สามารถบันทึก",
		Err:     cause,
	}

	got := appErr.Error()
	if !strings.Contains(got, "ไม่สามารถบันทึก") {
		t.Errorf("expected message in output, got %q", got)
	}
	if !strings.Contains(got, "db down") {
		t.Errorf("expected cause in output, got %q", got)
	}
}

// TestAppError_Error_NoCause ตรวจกรณีที่ Err = nil
// คาดว่าจะคืนแค่ Message
func TestAppError_Error_NoCause(t *testing.T) {
	appErr := &AppError{
		Code:    400,
		Message: "input ไม่ถูก",
		Err:     nil,
	}

	got := appErr.Error()
	if got != "input ไม่ถูก" {
		t.Errorf("expected exact message, got %q", got)
	}
}

// TestAppError_HTTPStatus_ZeroFallback ตรวจว่าถ้า Code = 0 จะ fallback เป็น 500
func TestAppError_HTTPStatus_ZeroFallback(t *testing.T) {
	appErr := &AppError{Code: 0}

	if got := appErr.HTTPStatus(); got != http.StatusInternalServerError {
		t.Errorf("expected 500 fallback, got %d", got)
	}
}

// TestAppError_HTTPStatus_ExplicitCode ตรวจว่า Code ที่กำหนดถูกส่งคืน
func TestAppError_HTTPStatus_ExplicitCode(t *testing.T) {
	appErr := &AppError{Code: 418}

	if got := appErr.HTTPStatus(); got != 418 {
		t.Errorf("expected 418, got %d", got)
	}
}

// TestAppError_Unwrap ตรวจว่า Unwrap คืน inner error ถูกตัว
func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("inner")
	appErr := &AppError{Err: cause}

	if got := appErr.Unwrap(); got != cause {
		t.Errorf("expected unwrap to return inner error, got %v", got)
	}
}

// TestAsAppError_NonAppError ตรวจว่า error ที่ไม่ใช่ AppError จะคืน (nil, false)
func TestAsAppError_NonAppError(t *testing.T) {
	plainErr := errors.New("plain")
	appErr, ok := AsAppError(plainErr)

	if ok {
		t.Error("expected ok=false for plain error")
	}
	if appErr != nil {
		t.Errorf("expected nil AppError, got %+v", appErr)
	}
}
