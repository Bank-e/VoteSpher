package voting

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// SubmitVoteService จัดการขั้นตอนและกฎเกณฑ์ก่อนการลงคะแนน
func SubmitVoteService(db *gorm.DB, voterID uint, areaID uint, req SubmitBallotRequest) error {
	
	// 1. ดึงการตั้งค่าระบบจาก Repository
	config, err := GetActiveConfig(db)
	if err != nil {
		return err // ส่งต่อ error (เช่น ไม่พบ config) ไปให้ Handler
	}

	// 2. ด่านตรวจสอบสถานะระบบ (Business Rule: ตรวจสอบเวลา)
	now := time.Now()
	if config.Status != "open" || now.Before(config.StartTime) || now.After(config.EndTime) {
		return errors.New("403: อยู่นอกเวลาการลงคะแนน หรือระบบปิดรับคะแนนแล้ว")
	}

	// 3. เตรียมข้อมูลผลโหวต (ประกอบร่าง Data)
	// สร้าง struct Vote รอไว้ โดยไม่ใส่ VoterID เข้าไปเพื่อรักษา Secret Ballot
	voteRecord := Vote{
		AreaID:      areaID,
		CandidateNo: req.CandidateNo,
		PartyNo:     req.PartyNo,
		CreatedAt:   time.Now(),
	}

	// 4. สั่งให้ Repository ทำการบันทึกข้อมูลแบบ Transaction
	return ExecuteVoteTransaction(db, voterID, voteRecord)
}