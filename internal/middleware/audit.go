package middleware

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"votespher/internal/models"
)

// AuditLog บันทึก action ของผู้ใช้ลง DB แบบ async
func AuditLog(db *gorm.DB, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		var voterID *uint
		if id, ok := c.Get("voter_id"); ok {
			if uid, ok := id.(uint); ok {
				voterID = &uid
			}
		}

		detail := fmt.Sprintf("%s %s → %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
		entry := models.AuditLog{
			VoterID:   voterID,
			Action:    action,
			Detail:    detail,
			IPAddress: c.ClientIP(),
		}

		go func() {
			if err := db.Create(&entry).Error; err != nil {
				log.Printf("audit log write failed: %v", err)
			}
		}()
	}
}
