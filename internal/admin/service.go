package election

import "gorm.io/gorm"

// UpdateElectionConfig อัปเดตการตั้งค่าการเลือกตั้ง
func UpdateElectionConfig(db *gorm.DB, req UpdateConfigRequest) (*ConfigResponse, error) {
	// ดึง config ที่ active อยู่มาก่อน
	cfg, err := GetActiveConfig(db)
	if err != nil {
		return nil, err
	}

	// อัปเดตค่าตาม request
	if err := UpdateConfig(db, cfg, req.Status, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// คืนค่า config ที่อัปเดตแล้ว
	return &ConfigResponse{
		ConfigID:  cfg.ConfigID,
		Status:    cfg.Status,
		StartTime: cfg.StartTime,
		EndTime:   cfg.EndTime,
		UpdatedAt: cfg.UpdatedAt,
		IsActive:  cfg.IsActive,
	}, nil
}
