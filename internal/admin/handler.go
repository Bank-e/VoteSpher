package election

import (
	"encoding/json"
	"net/http"
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
