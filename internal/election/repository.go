package election

import (
	"errors"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// GetAdminByVoterID ค้นหาข้อมูลแอดมินจากรหัสผู้โหวต
func GetAdminByVoterID(db *gorm.DB, voterID uint) (*models.Admin, error) {
	var admin models.Admin
	if err := db.Where("voter_id = ?", voterID).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// หา config ที่ active อยู่ตัวล่าสุด
func GetActiveConfig(db *gorm.DB) (*models.SystemConfig, error) {
	var cfg models.SystemConfig
	// แนะนำให้ใช้ ? แทน true เผื่อกรณีสลับใช้ Database ต่างค่ายกัน (เช่น MySQL ใช้ 1/0)
	err := db.Where("is_active = ?", true).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// CreateConfigVersion จัดการยกเลิกของเก่าและสร้างของใหม่แบบ Transaction
func CreateConfigVersion(db *gorm.DB, oldConfig *models.SystemConfig, newConfig *models.SystemConfig) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. ยกเลิกของเก่า (set is_active = false)
		if err := tx.Model(oldConfig).Update("is_active", false).Error; err != nil {
			return errors.New("ไม่สามารถยกเลิกการตั้งค่าเดิมได้")
		}

		// 2. บันทึกของใหม่ (Insert row ใหม่)
		if err := tx.Create(newConfig).Error; err != nil {
			return errors.New("ไม่สามารถบันทึกการตั้งค่าใหม่ได้")
		}

		return nil
	})
}