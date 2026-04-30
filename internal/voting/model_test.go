package voting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	// ทดสอบการสร้าง AppError
	err := NewAppError(404, "Not Found")

	// ตรวจสอบค่าภายใน
	assert.Equal(t, 404, err.Code)
	
	// ตรวจสอบว่า Implement interface `error` ได้ถูกต้อง
	assert.Equal(t, "Not Found", err.Error())
}