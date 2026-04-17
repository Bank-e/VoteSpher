package election

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PATCH /election/config
// อัปเดตการตั้งค่าการเลือกตั้ง (ถูกควบคุมสิทธิ์ admin จาก Middleware แล้ว)
func UpdateConfigHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. อ่าน request body ได้เลย 
		// ใช้ ShouldBindJSON ของ Gin เพื่อแปลง JSON เป็น Struct อัตโนมัติ
		var req UpdateConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body ไม่ถูกต้อง"})
			return
		}

		// 2. ตรวจสอบความถูกต้องของข้อมูล (Validation)
		// (ถ้าต้องการลดโค้ดส่วนนี้ สามารถใช้ binding tags ใน Struct UpdateConfigRequest ได้ เช่น `binding:"required"`)
		if req.Status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ status"})
			return
		}

		// 3. เรียกใช้ Service เพื่ออัปเดต config ลง Database
		result, err := UpdateElectionConfig(db, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 4. ส่งผลลัพธ์กลับเป็น JSON
		// Gin จะจัดการ set Content-Type เป็น application/json ให้อัตโนมัติ
		c.JSON(http.StatusOK, result)
	}
}