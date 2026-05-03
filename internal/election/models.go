package election

import "time"

// UpdateConfigRequest — body ของ PATCH /election/config
//
// validate ด้วย gin's binding tag (เช็คเฉพาะค่าว่างเปล่า/รูปแบบ):
//   - status ต้องไม่ว่าง
//   - start_time / end_time ต้องไม่ว่าง
//
// ค่าที่ status รับได้ (case-insensitive) และการเช็ค start < end
// ทำใน service layer เพื่อแยก concern ระหว่าง syntactic vs semantic validation
type UpdateConfigRequest struct {
	// สถานะการเลือกตั้ง: PREPARE / OPEN / PAUSED / CLOSED / COUNTING (case-insensitive)
	Status string `json:"status" binding:"required"`

	// เวลาเริ่มเลือกตั้ง
	StartTime time.Time `json:"start_time" binding:"required"`

	// เวลาสิ้นสุดเลือกตั้ง
	EndTime time.Time `json:"end_time" binding:"required"`
}

// ConfigResponse — response หลังอัปเดต config สำเร็จ
type ConfigResponse struct {
	ConfigID  uint      `json:"config_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}