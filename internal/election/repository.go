package election

import (
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// หา config ที่ active อยู่ตัวล่าสุด
func GetActiveConfig(db *gorm.DB) (*models.SystemConfig, error) {
	var cfg models.SystemConfig
	err := db.Where("is_active = true").First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// อัปเดต status, start_time, end_time ของ config
func UpdateConfig(db *gorm.DB, cfg *models.SystemConfig, status string, startTime time.Time, endTime time.Time) error {
	return db.Model(cfg).Updates(map[string]interface{}{
		"status":     status,
		"start_time": startTime,
		"end_time":   endTime,
		"updated_at": time.Now(),
	}).Error
}
