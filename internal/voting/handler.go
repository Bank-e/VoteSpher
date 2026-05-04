package voting

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==========================================
// 1. Struct & Constructor
// ==========================================

// VotingHandler ทำหน้าที่รับ Request จากผู้ใช้ ตรวจสอบข้อมูลเบื้องต้น แล้วส่งต่อให้ Service
type VotingHandler struct {
	service VotingService // เรียกใช้งาน Service ผ่าน Interface
}

// NewVotingHandler สร้าง Instance ของ Handler โดยรับ Service เข้ามา (Dependency Injection)
func NewVotingHandler(service VotingService) *VotingHandler {
	return &VotingHandler{
		service: service,
	}
}

// ==========================================
// 2. Handler Methods
// ==========================================

// SubmitBallotHandler POST /ballot/submit
func (h *VotingHandler) SubmitBallotHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse Body
		var req SubmitBallotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message":    "รูปแบบข้อมูลไม่ถูกต้อง",
			})
			return
		}

		// 2. ดึงค่าจาก Gin context (ที่ได้มาจาก Middleware)
		ctxVoterID, existsVoter := c.Get("voter_id")
		ctxAreaID, existsArea := c.Get("area_id")

		if !existsVoter || !existsArea {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message":    "Token ไม่ถูกต้อง หรือหมดอายุ",
			})
			return
		}

		// 3. แปลง Type (uint)
		voterID, ok := ctxVoterID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				// voter_id ใน Token ไม่ใช่รูปแบบตัวเลข (uint)
				"message":    "เกิดข้อผิดพลาดภายในระบบ",
			})
			return
		}

		areaID, ok := ctxAreaID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				// area_id ใน Token ไม่ใช่รูปแบบตัวเลข (uint)
				"message":    "เกิดข้อผิดพลาดภายในระบบ",
			})
			return
		}

		// 4. ส่งไป Service ให้จัดการ Business Logic
		err := h.service.SubmitVote(voterID, areaID, req)
		if err != nil {
			// จัดการ Error แบบใหม่: ตรวจสอบว่าเป็น AppError ที่เราสร้างไว้ใน model.go หรือไม่
			if appErr, ok := err.(*AppError); ok {
				c.JSON(appErr.Code, gin.H{
					"message":    appErr.Message,
				})
				return
			}

			// กรณีเกิด Error อื่นๆ ที่หลุดรอดมา (เช่น Database พังฉับพลัน)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message":    "เกิดข้อผิดพลาดภายในระบบ: " + err.Error(),
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

// GetBallotStatusHandler GET /ballot/status
// ตรวจสอบสถานะระบบและสถานะการลงคะแนนของผู้ใช้งาน
func (h *VotingHandler) GetBallotStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. ดึง VoterID ออกจาก Token
		ctxVoterID, exists := c.Get("voter_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบข้อมูลยืนยันตัวตน"})
			return
		}

		voterID, ok := ctxVoterID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ข้อมูลยืนยันตัวตนไม่ถูกต้อง"})
			return
		}

		// 2. เรียกใช้งาน Service
		result, err := h.service.GetBallotStatus(voterID)
		if err != nil {
			// จัดการ Error ด้วย AppError
			if appErr, ok := err.(*AppError); ok {
				c.JSON(appErr.Code, gin.H{
					"status":  "error",
					"message": appErr.Message,
				})
				return
			}

			// กรณี Error ปกติที่ไม่ได้ทำ Custom ไว้
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "เกิดข้อผิดพลาดภายในระบบ: " + err.Error(),
			})
			return
		}

		// 3. ห่อข้อมูลตอบกลับให้ Frontend
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "ดึงข้อมูลสถานะสำเร็จ",
			"data":    result,
		})
	}
}