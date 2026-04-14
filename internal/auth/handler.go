package auth

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

// POST /voter/otp-confirm
// รับ otp_code และ ref_code แล้วยืนยัน OTP
// ถ้าถูกต้องจะคืน JWT token กลับไป
func OTPConfirmHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// อ่าน request body
		var req OTPConfirmRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "request body ไม่ถูกต้อง", http.StatusBadRequest)
			return
		}

		// ตรวจว่า field ครบไหม
		if req.OTPCode == "" || req.RefCode == "" {
			http.Error(w, "กรุณาระบุ otp_code และ ref_code", http.StatusBadRequest)
			return
		}

		// ส่งไปให้ service ยืนยัน OTP
		result, err := ConfirmOTP(db, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(OTPConfirmResponse{Token: result.Token})
	}
}
