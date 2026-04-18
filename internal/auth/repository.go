package auth

import (
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// หา OTP จาก ref_code
// เช็คว่ายังไม่หมดอายุ และยังไม่ถูกใช้
func FindOTPByRefCode(db *gorm.DB, refCode string) (*models.OTP, error) {
	var otp models.OTP
	err := db.Where("ref_code = ? AND is_used = false AND expires_at > ?", refCode, time.Now()).
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

// mark OTP ว่าถูกใช้แล้ว เพื่อป้องกันการใช้ซ้ำ
func MarkOTPAsUsed(db *gorm.DB, otpID uint) error {
	return db.Model(&models.OTP{}).
		Where("otp_id = ?", otpID).
		Update("is_used", true).Error
}

// หา Voter จาก voter_id เพื่อเอา area_id ไปสร้าง JWT
func FindVoterByID(db *gorm.DB, voterID uint) (*models.Voter, error) {
	var voter models.Voter
	err := db.First(&voter, voterID).Error
	if err != nil {
		return nil, err
	}
	return &voter, nil
}

// ค้นหา Voter พร้อมดึงข้อมูล Area (เขต/จังหวัด) มาด้วย
func FindVoterByCitizenIDHash(db *gorm.DB, citizenIDHash string) (*models.Voter, error) {
	var voter models.Voter
	// Preload("Area") จะทำให้ GORM ไปดึงชื่อเขตและจังหวัดมาจากตาราง areas ให้เราอัตโนมัติ
	err := db.Preload("Area").Where("citizen_id_hash = ?", citizenIDHash).First(&voter).Error
	if err != nil {
		return nil, err
	}
	return &voter, nil
}
