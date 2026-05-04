package election

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError คือ error type ที่ผูก HTTP status code มากับข้อความให้เลย
// ทำให้ handler ไม่ต้อง parse string เพื่อหาว่า error นี้ควรตอบ HTTP code อะไร
type AppError struct {
	Code    int    // HTTP status code (เช่น 400, 403, 500)
	Message string // ข้อความที่ปลอดภัยที่จะส่งให้ client
	Err     error  // error ตัวจริง (เก็บไว้ wrap ไม่ส่งออกให้ client)
}

// Error ทำให้ AppError implement interface error ได้
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap ทำให้ใช้ errors.Is / errors.As ตรวจกับ inner error ได้
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus คืนค่า status code (ไว้ให้ handler หยิบมาใช้)
func (e *AppError) HTTPStatus() int {
	if e.Code == 0 {
		return http.StatusInternalServerError
	}
	return e.Code
}

// helper สร้าง AppError แบบเร็ว
func newAppError(code int, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Err: cause}
}

// AsAppError ดึง *AppError ออกจาก error chain
// คืน (nil, false) ถ้าไม่ใช่ AppError
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// ============================================================
// Sentinel errors — ใช้เป็น "ชนิด" ของ error เพื่อให้ test/compare ได้
// ============================================================

var (
	// 400 — input ไม่ถูก
	ErrInvalidRequest         = errors.New("request body ไม่ถูกต้อง")
	ErrStatusRequired         = errors.New("กรุณาระบุ status")
	ErrInvalidStatus          = errors.New("status ไม่ถูกต้อง: รองรับเฉพาะ PREPARE, OPEN, PAUSED, CLOSED, COUNTING")
	ErrInvalidTimeRange       = errors.New("เวลาเปิดหีบต้องมาก่อนเวลาปิดหีบเสมอ")
	ErrInvalidStateTransition = errors.New("ไม่อนุญาตให้เปิดหีบใหม่ เนื่องจากระบบถูกปิด (CLOSED) ไปแล้ว")

	// 401/403 — สิทธิ์
	ErrUnauthorized = errors.New("ไม่พบข้อมูลยืนยันตัวตนใน Token")
	ErrInvalidToken = errors.New("ข้อมูลยืนยันตัวตนไม่ถูกต้อง")
	ErrNotAdmin     = errors.New("คุณไม่มีสิทธิ์ผู้ดูแลระบบ")

	// 500 — ระบบ
	ErrConfigNotFound     = errors.New("ไม่พบการตั้งค่าระบบที่กำลังใช้งานอยู่")
	ErrConfigUpdateFailed = errors.New("ไม่สามารถบันทึกการตั้งค่าใหม่ได้")
	ErrConfigDeactivate   = errors.New("ไม่สามารถยกเลิกการตั้งค่าเดิมได้")
)

// ============================================================
// Constructor helpers — รวมตำแหน่งสร้าง AppError ที่ใช้บ่อย
// ============================================================

func badRequest(sentinel error) *AppError {
	return newAppError(http.StatusBadRequest, sentinel.Error(), sentinel)
}

func unauthorized(sentinel error) *AppError {
	return newAppError(http.StatusUnauthorized, sentinel.Error(), sentinel)
}

func forbidden(sentinel error) *AppError {
	return newAppError(http.StatusForbidden, sentinel.Error(), sentinel)
}

func internal(sentinel error, cause error) *AppError {
	// เก็บ sentinel ไว้ใน chain เสมอ เพื่อให้ errors.Is(err, sentinel) ทำงานได้
	// ถ้ามี cause ด้วยให้ wrap sentinel + cause เข้าด้วยกัน
	err := error(sentinel)
	if cause != nil {
		err = fmt.Errorf("%w: %v", sentinel, cause)
	}
	return newAppError(http.StatusInternalServerError, sentinel.Error(), err)
}
