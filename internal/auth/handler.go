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

var tokenGenerator = pkg.GenerateToken

type AuthHandler struct {
	service AuthService
	db      *gorm.DB
}

func NewAuthHandler(service AuthService, db *gorm.DB) *AuthHandler {
	return &AuthHandler{service: service, db: db}
}

// POST /voter/verify
func (h *AuthHandler) VerifyVoter(c *gin.Context) {
	var req VerifyVoterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง (ต้องการ citizen_id)"})
		return
	}
	res, err := h.service.VerifyVoter(req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// POST /voter/otp-request
func (h *AuthHandler) OTPRequest(c *gin.Context) {
	var req OTPRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ต้องการ voter_id"})
		return
	}
	res, err := h.service.RequestOTP(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// POST /voter/otp-confirm
func (h *AuthHandler) OTPConfirm(c *gin.Context) {
	var req OTPConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body ไม่ถูกต้อง"})
		return
	}
	if req.OTPCode == "" || req.RefCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ otp_code และ ref_code"})
		return
	}
	result, err := h.service.ConfirmOTP(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, OTPConfirmResponse{Token: result.Token, Role: result.Role})
}

// ============================================================
// Standalone handlers (not part of AuthService — use db directly)
// ============================================================

type MockTokenRequest struct {
	VoterID uint   `json:"voter_id"`
	AreaID  uint   `json:"area_id"`
	Role    string `json:"role"`
}

// POST /dev/mock-token
func MockTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MockTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
			return
		}
		if req.VoterID == 0 || req.AreaID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ voter_id และ area_id (ต้องมากกว่า 0)"})
			return
		}
		if req.Role == "" {
			req.Role = "voter"
		}
		secretKey := os.Getenv("JWT_SECRET_KEY")
		if secretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ระบบผิดพลาด กรุณาติดต่อผู้ดูแล"})
			return
		}
		token, err := tokenGenerator(req.VoterID, req.AreaID, req.Role, secretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้าง Token ไม่สำเร็จ"})
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   7200,
			"mock_data":    req,
		})
	}
}

// GET /voter/me
func VoterMeHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxVoterID, exists := c.Get("voter_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบข้อมูลยืนยันตัวตน"})
			return
		}
		voterID, ok := ctxVoterID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดภายในระบบ"})
			return
		}
		var voter models.Voter
		if err := db.Preload("Area.Province").First(&voter, voterID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลผู้มีสิทธิ์"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"voter_id":     voter.ID,
			"area_id":      voter.AreaID,
			"area_name":    voter.Area.Province.ProvinceName + " " + voter.Area.AreaName,
			"province":     voter.Area.Province.ProvinceName,
			"email":        voter.Email,
			"phone_number": voter.PhoneNumber,
			"is_voted":     voter.IsVoted,
			"voted_at":     voter.VotedAt,
		})
	}
}

// OTPRequestHandler kept for backward compat
func OTPRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OTPRequestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ต้องการ voter_id"})
			return
		}
		var voter models.Voter
		if err := db.First(&voter, req.VoterID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรหัสผู้มีสิทธิ์โหวตนี้"})
			return
		}
		if voter.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ผู้มีสิทธิ์โหวตรายนี้ยังไม่มี email ในระบบ"})
			return
		}
		var otpCode string
		mode := os.Getenv("OTP_DELIVERY_MODE")
		if mode == "mock" || os.Getenv("ENABLE_DEV_ENDPOINTS") == "true" {
			otpCode = "111111"
		} else {
			var err2 error
			otpCode, err2 = generateRandomOTP()
			if err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้าง OTP ไม่สำเร็จ"})
				return
			}
		}
		refCode, err := generateRefCode()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้าง OTP ไม่สำเร็จ"})
			return
		}
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
		if err := pkg.EnqueueOTPEmail(voter.Email, otpCode, refCode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ส่ง OTP ไม่สำเร็จ กรุณาลองใหม่"})
			return
		}
		c.JSON(http.StatusOK, OTPRequestResponse{RefCode: refCode})
	}
}

// VerifyVoterHandler kept for backward compat
func VerifyVoterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyVoterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง (ต้องการ citizen_id)"})
			return
		}
		hashedID := generateCitizenIDHash(req.CitizenID)
		var voter models.Voter
		if err := db.Preload("Area.Province").Where("citizen_id_hash = ?", hashedID).First(&voter).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้มีสิทธิ์เลือกตั้ง"})
			return
		}
		res := VerifyVoterResponse{
			VoterID: voter.ID,
			VoterInfo: VoterInfo{
				Name:     "ข้อมูลปกปิด",
				AreaID:   voter.AreaID,
				AreaName: voter.Area.Province.ProvinceName + " " + voter.Area.AreaName,
				Province: voter.Area.Province.ProvinceName,
				IsVoted:  voter.IsVoted,
			},
		}
		c.JSON(http.StatusOK, res)
	}
}

// OTPConfirmHandler kept for backward compat
func OTPConfirmHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OTPConfirmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request body ไม่ถูกต้อง"})
			return
		}
		if req.OTPCode == "" || req.RefCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุ otp_code และ ref_code"})
			return
		}
		repo := NewAuthRepository(db)
		svc := NewAuthService(repo)
		result, err := svc.ConfirmOTP(req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, OTPConfirmResponse{Token: result.Token, Role: result.Role})
	}
}
