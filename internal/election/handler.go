package election

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PATCH /election/config
// อัปเดตการตั้งค่าการเลือกตั้ง (ถูกควบคุมสิทธิ์ admin จาก Middleware แล้ว)
func UpdateConfigHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body ไม่ถูกต้อง"})
			return
		}

		if req.Status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ status"})
			return
		}

		// 1. ดึง voter_id จาก Token (ที่ Middleware ใส่ไว้ให้)
		ctxVoterID, exists := c.Get("voter_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบข้อมูลยืนยันตัวตนใน Token"})
			return
		}

		voterID, ok := ctxVoterID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ข้อมูลยืนยันตัวตนไม่ถูกต้อง"})
			return
		}

		// ส่ง voterID เข้าไปใน Service
		result, err := UpdateElectionConfig(db, voterID, req)
		if err != nil {
			statusCode := http.StatusInternalServerError

			// ถ้าเป็น Error จากฝั่งผู้ใช้ (400) ให้ตอบกลับเป็น Bad Request
			if strings.Contains(err.Error(), "403") {
				statusCode = http.StatusForbidden
			} else if strings.Contains(err.Error(), "400") {
				statusCode = http.StatusBadRequest
			}

			c.JSON(statusCode, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
            "status":  "success",
            "message": "อัปเดตการตั้งค่าระบบเลือกตั้งเรียบร้อยแล้ว",
            "data":    result,
        })
	}
}