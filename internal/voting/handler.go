package voting

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SubmitBallotHandler POST /ballot/submit
func SubmitBallotHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse Body — เปลี่ยนจาก json.NewDecoder → ShouldBindJSON
		var req SubmitBallotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error_code": "BAD_REQUEST",
				"message":    "รูปแบบข้อมูลไม่ถูกต้อง",
			})
			return
		}

		// 2. ดึงค่าจาก Gin context — เปลี่ยนจาก r.Context().Value() → c.Get()
		ctxVoterID, existsVoter := c.Get("voter_id")
		ctxAreaID, existsArea := c.Get("area_id")

		if !existsVoter || !existsArea {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error_code": "UNAUTHORIZED",
				"message":    "Token ไม่ถูกต้อง หรือหมดอายุ",
			})
			return
		}

		// 3. แปลง Type — logic เดิมเลย
		voterID, ok := ctxVoterID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_code": "SERVER_ERROR",
				"message":    "voter_id ใน Token ไม่ใช่รูปแบบตัวเลข (uint)",
			})
			return
		}

		areaID, ok := ctxAreaID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_code": "SERVER_ERROR",
				"message":    "area_id ใน Token ไม่ใช่รูปแบบตัวเลข (uint)",
			})
			return
		}

		// 4. ส่งไป Service — เหมือนเดิมทุกอย่าง
		err := SubmitVoteService(db, voterID, areaID, req)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if strings.Contains(err.Error(), "403") {
				statusCode = http.StatusForbidden
			} else if strings.Contains(err.Error(), "404") {
				statusCode = http.StatusNotFound
			}
			c.JSON(statusCode, gin.H{
				"error_code": "VOTE_FAILED",
				"message":    err.Error(),
			})
			return
		}

		// 5. สำเร็จ
		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "บันทึกคะแนนสำเร็จ",
		})
	}
}