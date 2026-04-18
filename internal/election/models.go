package election

import "time"

// PATCH /election/config request body
type UpdateConfigRequest struct {
	Status    string    `json:"status"`     // สถานะการเลือกตั้ง เช่น "open", "closed", "paused"
	StartTime time.Time `json:"start_time"` // เวลาเริ่มเลือกตั้ง
	EndTime   time.Time `json:"end_time"`   // เวลาสิ้นสุดเลือกตั้ง
}

// response หลังอัปเดต config สำเร็จ
type ConfigResponse struct {
	ConfigID  uint      `json:"config_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}