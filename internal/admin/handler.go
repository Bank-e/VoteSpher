package election

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

// PATCH /election/config
// อัปเดตการตั้งค่าการเลือกตั้ง (ถูกควบคุมสิทธิ์ admin จาก Middleware แล้ว)
func UpdateConfigHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. อ่าน request body ได้เลย 
		// (ถ้าหลุดเข้ามาถึงบรรทัดนี้ได้ แปลว่า Middleware ยืนยันแล้วว่าคนเรียกคือ Admin ตัวจริง)
		var req UpdateConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "request body ไม่ถูกต้อง", http.StatusBadRequest)
			return
		}

		// 2. ตรวจสอบความถูกต้องของข้อมูล (Validation)
		if req.Status == "" {
			http.Error(w, "กรุณาระบุ status", http.StatusBadRequest)
			return
		}

		// 3. เรียกใช้ Service เพื่ออัปเดต config ลง Database
		result, err := UpdateElectionConfig(db, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 4. ส่งผลลัพธ์กลับ
		json.NewEncoder(w).Encode(result)
	}
}