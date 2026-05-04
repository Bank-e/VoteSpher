package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"votespher/internal/models"
	"votespher/pkg"
)

type MockTokenRequest struct {
	VoterID uint   `json:"voter_id"`
	AreaID  uint   `json:"area_id"`
	Role    string `json:"role"`
}

// POST /dev/mock-token
func MockTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse Request Body ด้วย Gin
		var req MockTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
			return
		}

		// 2. Validate ข้อมูลขั้นต่ำ
		if req.VoterID == 0 || req.AreaID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ voter_id และ area_id (ต้องมากกว่า 0)"})
			return
		}

		if req.Role == "" {
			req.Role = "voter"
		}

		secretKey := os.Getenv("JWT_SECRET_KEY")
		if secretKey == "" {
			secretKey = "dev_secret_key"
		}

		token, err := pkg.GenerateToken(req.VoterID, req.AreaID, req.Role, secretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้าง Token ไม่สำเร็จ: " + err.Error()})
			return
		}

		response := map[string]interface{}{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   7200,
			"mock_data":    req,
		}

		// 3. ส่ง Response กลับ
		c.JSON(http.StatusOK, response)
	}
}

// POST /voter/otp-confirm
// รับ otp_code และ ref_code แล้วยืนยัน OTP
// ถ้าถูกต้องจะคืน JWT token กลับไป
func OTPConfirmHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. อ่าน request body
		var req OTPConfirmRequest // สมมติว่ามี struct นี้ประกาศไว้ที่อื่นใน package
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body ไม่ถูกต้อง"})
			return
		}

		// 2. ตรวจว่า field ครบไหม
		if req.OTPCode == "" || req.RefCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ otp_code และ ref_code"})
			return
		}

		// 3. ส่งไปให้ service ยืนยัน OTP
		result, err := ConfirmOTP(db, req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// 4. ส่ง Token กลับไป
		c.JSON(http.StatusOK, OTPConfirmResponse{Token: result.Token})
	}
}

// / POST /voter/verify
// ฟังก์ชันสำหรับตรวจสอบสิทธิ์ผู้เลือกตั้ง
func VerifyVoterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyVoterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง (ต้องการ citizen_id)"})
			return
		}

		// เรียกใช้ฟังก์ชัน Hash ที่เพื่อนทำไว้ใน service.go
		hashedID := generateCitizenIDHash(req.CitizenID)

		// เรียกใช้ฟังก์ชันหาข้อมูลที่เราเพิ่มไว้ใน repository.go
		voter, err := FindVoterByCitizenIDHash(db, hashedID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้มีสิทธิ์เลือกตั้ง"})
			return
		}

		// เตรียมข้อมูลส่งกลับตามโครงสร้าง VerifyVoterResponse ใน models.go
		res := VerifyVoterResponse{
			VoterID: voter.ID,
			VoterInfo: VoterInfo{
				Name:     "ข้อมูลปกปิด",
				AreaID:   voter.AreaID,
				AreaName: voter.Area.AreaName,
				Province: voter.Area.Province.ProvinceName, // 👈 แก้ตรงนี้: ดึงชื่อจังหวัดมาโชว์ของจริงแล้ว!
				IsVoted:  voter.IsVoted,
			},
		}
		c.JSON(http.StatusOK, res)
	}
}

// POST /voter/otp-request
// ฟังก์ชันสำหรับขอรหัส OTP
func OTPRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OTPRequestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ต้องการ voter_id"})
			return
		}

		// ตรวจสอบก่อนว่ามี Voter นี้จริงไหม
		_, err := FindVoterByID(db, req.VoterID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรหัสผู้มีสิทธิ์โหวตนี้"})
			return
		}

		// สุ่มรหัสจากฟังก์ชันที่เราเพิ่มไว้ใน service.go
		otpCode, _ := generateRandomOTP()
		refCode, _ := generateRefCode()

		// บันทึกลงฐานข้อมูล (ตาราง otps)
		newOTP := models.OTP{
			VoterID:   req.VoterID,
			OTPCode:   otpCode,
			RefCode:   refCode,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}

		if err := db.Create(&newOTP).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้าง OTP ไม่สำเร็จ"})
			return
		}

		c.JSON(http.StatusOK, OTPRequestResponse{
			RefCode: refCode,
			OTPCode: otpCode,
		})
	}
}
