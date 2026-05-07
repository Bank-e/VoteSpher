package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
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
	// ใช้ ShouldBindJSON เพื่อให้ Gin เช็ค binding:"required" ให้เลย
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

	token, err := h.service.ConfirmOTP(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// ใช้ Struct Response ของคุณในการตอบกลับ
	c.JSON(http.StatusOK, OTPConfirmResponse{Token: token})
}
