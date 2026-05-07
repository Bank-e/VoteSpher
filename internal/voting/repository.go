package voting

import (
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ==========================================
// ************** Interface **************
// ==========================================

// VotingRepository กำหนดสัญญา (Contract) ว่า Repository นี้ทำอะไรได้บ้าง
type VotingRepository interface {
	GetActiveConfig() (*models.SystemConfig, error)
	ExecuteVoteTransaction(voterID uint, voteRecord models.Vote) error
	CheckUserHasVoted(voterID uint) (bool, error)
}

// ==========================================
// ********* Struct & Constructor *********
// ==========================================

// votingRepository เป็น Implementation ของ VotingRepository Interface
type votingRepository struct {
	db *gorm.DB
}

// NewVotingRepository สร้าง Instance ของ Repository โดยรับ Database Connection เข้ามา
func NewVotingRepository(db *gorm.DB) VotingRepository {
	return &votingRepository{
		db: db,
	}
}

// ==========================================
// **************** Methods ****************
// ==========================================

// GetActiveConfig ดึงการตั้งค่าระบบที่กำลังเปิดใช้งานอยู่ (ใช้แทนทั้ง GetActiveConfig และ GetActiveElectionConfig เดิม)
func (r *votingRepository) GetActiveConfig() (*models.SystemConfig, error) {
	var config models.SystemConfig
	if err := r.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		return nil, err // ส่ง Error ดิบกลับไปให้ Service ตัดสินใจจัดการต่อ
	}
	return &config, nil
}

// ExecuteVoteTransaction จัดการบันทึกคะแนนโหวตในรูปแบบ Transaction
func (r *votingRepository) ExecuteVoteTransaction(voterID uint, voteRecord models.Vote) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var voter models.Voter

		// ล็อก Row ข้อมูลผู้โหวตคนนี้ (ป้องกัน Race condition หรือการยิง Request ซ้ำๆ เข้ามาพร้อมกัน)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&voter, voterID).Error; err != nil {
			// ใช้ Custom Error ที่เราสร้างไว้ใน model.go
			return NewAppError(404, "ไม่พบข้อมูลผู้มีสิทธิเลือกตั้ง")
		}

		// เช็คสถานะการโหวต *ต้องทำใน Transaction ที่ถูกล็อกแล้วเท่านั้น*
		if voter.IsVoted {
			return NewAppError(403, "คุณได้ลงคะแนนไปแล้ว ไม่สามารถลงคะแนนซ้ำได้")
		}

		// บันทึกคะแนน (Secret Ballot - ในตัวแปร voteRecord จะไม่มี voterID)
		if err := tx.Create(&voteRecord).Error; err != nil {
			return NewAppError(500, "ไม่สามารถบันทึกคะแนนได้")
		}

		// อัปเดตสถานะผู้โหวตว่า "โหวตแล้ว"
		if err := tx.Model(&voter).Updates(map[string]interface{}{
			"is_voted": true,
			"voted_at": time.Now(),
		}).Error; err != nil {
			return NewAppError(500, "ไม่สามารถอัปเดตสถานะผู้โหวตได้")
		}

		return nil
	})
}

// CheckUserHasVoted ตรวจสอบว่าผู้ใช้งานคนนี้เคยลงคะแนนไปแล้วหรือยัง
func (r *votingRepository) CheckUserHasVoted(voterID uint) (bool, error) {
	var voter models.Voter

	// ส่ง ID เข้าไปใน First โดยตรง GORM จะใช้ Primary Key ให้เอง
	if err := r.db.Select("is_voted").First(&voter, voterID).Error; err != nil {
		return false, err
	}
	return voter.IsVoted, nil
}