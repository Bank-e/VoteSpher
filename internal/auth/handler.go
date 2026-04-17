package auth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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