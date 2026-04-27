package election

import (
	"encoding/json"
	"net/http"
<<<<<<< Updated upstream
	"os"
	"strings"
	"votespher/pkg"

	"gorm.io/gorm"
)

// PATCH /election/config
// อัปเดตการตั้งค่าการเลือกตั้ง (admin เท่านั้น)
func UpdateConfigHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// ตรวจสอบ JWT token จาก Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "กรุณา login ก่อน", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := pkg.ValidateToken(tokenStr, os.Getenv("JWT_SECRET"))
		if err != nil {
			http.Error(w, "token ไม่ถูกต้องหรือหมดอายุ", http.StatusUnauthorized)
			return
		}

		// เฉพาะ admin เท่านั้นที่แก้ config ได้
		if claims.Role != "admin" {
			http.Error(w, "ไม่มีสิทธิ์เข้าถึง", http.StatusForbidden)
			return
		}

		// อ่าน request body
		var req UpdateConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "request body ไม่ถูกต้อง", http.StatusBadRequest)
			return
		}

		// ตรวจว่า field ครบไหม
		if req.Status == "" {
			http.Error(w, "กรุณาระบุ status", http.StatusBadRequest)
			return
		}

		// อัปเดต config
		result, err := UpdateElectionConfig(db, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}
=======

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
// ถ้าไม่มีหรือ type ผิด จะคืน *AppError ที่ map กับ HTTP code ให้เรียบร้อย
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
// — ถ้าเป็น *AppError จะใช้ HTTPStatus ที่ติดมากับ error
// — ถ้าไม่ใช่ จะ fallback เป็น 500
//
// ทำให้ handler ไม่ต้อง parse string เพื่อหา status code อีกต่อไป
func respondError(c *gin.Context, err error) {
	if appErr, ok := AsAppError(err); ok {
		c.JSON(appErr.HTTPStatus(), gin.H{"error": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

>>>>>>> Stashed changes
