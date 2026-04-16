package voting

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetActiveConfig ดึงการตั้งค่าระบบที่กำลังเปิดใช้งานอยู่
func GetActiveConfig(db *gorm.DB) (*SystemConfig, error) {
	var config SystemConfig
	if err := db.Where("is_active = ?", true).First(&config).Error; err != nil {
		return nil, errors.New("ไม่พบการตั้งค่าระบบเลือกตั้งที่ใช้งานอยู่")
	}
	return &config, nil
}

// ExecuteVoteTransaction จัดการบันทึกคะแนนโหวตในรูปแบบ Transaction
func ExecuteVoteTransaction(db *gorm.DB, voterID uint, voteRecord Vote) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var voter Voter
		
		// ล็อก Row ข้อมูลผู้โหวตคนนี้ (ป้องกัน Race condition หรือการยิง Request ซ้ำๆ เข้ามาพร้อมกัน)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&voter, voterID).Error; err != nil {
			return errors.New("404: ไม่พบข้อมูลผู้มีสิทธิเลือกตั้ง")
		}

		// เช็คสถานะการโหวต *ต้องทำใน Transaction ที่ถูกล็อกแล้วเท่านั้น*
		if voter.IsVoted {
			return errors.New("403: คุณได้ลงคะแนนไปแล้ว ไม่สามารถลงคะแนนซ้ำได้")
		}

		// บันทึกคะแนน (Secret Ballot - ในตัวแปร voteRecord จะไม่มี voterID)
		if err := tx.Create(&voteRecord).Error; err != nil {
			return errors.New("500: ไม่สามารถบันทึกคะแนนได้")
		}

		// อัปเดตสถานะผู้โหวตว่า "โหวตแล้ว"
		if err := tx.Model(&voter).Updates(map[string]interface{}{
			"is_voted": true,
			"voted_at": time.Now(),
		}).Error; err != nil {
			return errors.New("500: ไม่สามารถอัปเดตสถานะผู้โหวตได้")
		}

		return nil
	})
}