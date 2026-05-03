package election

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler คือ HTTP layer ของ election
// รับ Service เป็น dependency เพื่อให้ test สามารถใส่ mock service เข้ามาได้
type Handler struct {
	svc Service
}

// NewHandler สร้าง Handler ใหม่จาก Service
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// UpdateConfig — PATCH /election/config
// อัปเดตการตั้งค่าการเลือกตั้ง (สิทธิ์ admin ถูกควบคุมจาก Middleware แล้ว)
func (h *Handler) UpdateConfig(c *gin.Context) {
	// 1. parse request body
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, badRequest(ErrInvalidRequest))
		return
	}

	// 2. ดึง voter_id จาก Token (ที่ Middleware ใส่ไว้ให้)
	voterID, err := extractVoterID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	// 3. ส่งต่อให้ service
	result, err := h.svc.UpdateElectionConfig(c.Request.Context(), voterID, req)
	if err != nil {
		respondError(c, err)
		return
	}

	// 4. success
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "อัปเดตการตั้งค่าระบบเลือกตั้งเรียบร้อยแล้ว",
		"data":    result,
	})
}

// ============================================================
// Helpers
// ============================================================

// extractVoterID ดึง voter_id ที่ middleware ใส่ไว้ใน gin.Context
func extractVoterID(c *gin.Context) (uint, error) {
	ctxVoterID, exists := c.Get("voter_id")
	if !exists {
		return 0, unauthorized(ErrUnauthorized)
	}

	voterID, ok := ctxVoterID.(uint)
	if !ok {
		return 0, internal(ErrInvalidToken, nil)
	}

	return voterID, nil
}

// respondError แปลง error เป็น HTTP response
func respondError(c *gin.Context, err error) {
	if appErr, ok := AsAppError(err); ok {
		c.JSON(appErr.HTTPStatus(), gin.H{"error": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
