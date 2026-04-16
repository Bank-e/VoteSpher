package voting

import (
	"encoding/json"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// SubmitBallotHandler POST /ballot/submit
func SubmitBallotHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. ตรวจสอบและ Parse ข้อมูลจาก Body
		var req SubmitBallotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error_code": "BAD_REQUEST", "message": "รูปแบบข้อมูลไม่ถูกต้อง"}`, http.StatusBadRequest)
			return
		}

		// 2. ดึงค่าจาก Context (สมมติว่าคุณมี jwt_middleware.go คอยเช็คและยัดค่าใส่ Request Context ไว้แล้ว)
		ctxVoterID := r.Context().Value("voter_id")
		ctxAreaID := r.Context().Value("area_id")

		if ctxVoterID == nil || ctxAreaID == nil {
			http.Error(w, `{"error_code": "UNAUTHORIZED", "message": "Token ไม่ถูกต้อง หรือหมดอายุ"}`, http.StatusUnauthorized)
			return
		}

		// แปลง Type ให้เป็น uint ตรงๆ พร้อมดักจับ Error ป้องกัน Server พัง (Panic)
        voterID, ok := ctxVoterID.(uint)
        if !ok {
            http.Error(w, `{"error_code": "SERVER_ERROR", "message": "voter_id ใน Token ไม่ใช่รูปแบบตัวเลข (uint)"}`, http.StatusInternalServerError)
            return
        }
        
        areaID, ok := ctxAreaID.(uint)
        if !ok {
            http.Error(w, `{"error_code": "SERVER_ERROR", "message": "area_id ใน Token ไม่ใช่รูปแบบตัวเลข (uint)"}`, http.StatusInternalServerError)
            return
        }

		// 3. ส่งไปให้ Service จัดการ
		err := SubmitVoteService(db, voterID, areaID, req)
		if err != nil {
			// จัดการแยก HTTP Status Code ตาม Message Error เพื่อความง่าย
			statusCode := http.StatusInternalServerError
			if strings.Contains(err.Error(), "403") {
				statusCode = http.StatusForbidden
			} else if strings.Contains(err.Error(), "404") {
				statusCode = http.StatusNotFound
			}
			
			// ส่ง Error กลับไป
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(map[string]string{
				"error_code": "VOTE_FAILED",
				"message":    err.Error(),
			})
			return
		}

		// 4. บันทึกสำเร็จ (201 Created)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
			"message": "บันทึกคะแนนสำเร็จ",
		})
	}
}