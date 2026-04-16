package auth

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"os"
	"votespher/pkg"
)

type MockTokenRequest struct {
	VoterID uint   `json:"voter_id"`
	AreaID  uint   `json:"area_id"`
	Role    string `json:"role"`
}

func MockTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 2. บังคับรับเฉพาะ POST Method (ถ้าใช้ Router ที่ดัก Method ให้อยู่แล้ว เอาออกได้ครับ)
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method Not Allowed"})
			return
		}

		var req MockTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			// 3. ปรับ Error ให้คืนค่าเป็น JSON
			json.NewEncoder(w).Encode(map[string]string{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
			return
		}

		// 4. Validate ข้อมูลขั้นต่ำ
		if req.VoterID == 0 || req.AreaID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "กรุณาระบุ voter_id และ area_id (ต้องมากกว่า 0)"})
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
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "สร้าง Token ไม่สำเร็จ: " + err.Error()})
			return
		}

		response := map[string]interface{}{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   7200,
			"mock_data":    req,
		}

		json.NewEncoder(w).Encode(response)
	}
}

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
