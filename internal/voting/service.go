package voting

import (
	"errors"
	"strings"
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// ==========================================
// 1. Interface
// ==========================================

// VotingService กำหนดสัญญา (Contract) ว่า Service นี้ทำอะไรได้บ้าง
type VotingService interface {
	SubmitVote(voterID uint, areaID uint, req SubmitBallotRequest) error
	GetBallotStatus(voterID uint) (*BallotStatusResponse, error)
}

// ==========================================
// 2. Struct & Constructor
// ==========================================

// votingService เป็น Implementation ของ VotingService Interface
type votingService struct {
	repo VotingRepository // เรียกใช้งาน Repository ผ่าน Interface
}

// NewVotingService สร้าง Instance ของ Service โดยรับ Repository เข้ามา (Dependency Injection)
func NewVotingService(repo VotingRepository) VotingService {
	return &votingService{
		repo: repo,
	}
}

// ==========================================
// 3. Methods (Business Logic)
// ==========================================

// SubmitVote จัดการขั้นตอนและกฎเกณฑ์ก่อนการลงคะแนน
func (s *votingService) SubmitVote(voterID uint, areaID uint, req SubmitBallotRequest) error {

	// 1. ดึงการตั้งค่าระบบจาก Repository
	config, err := s.repo.GetActiveConfig()
	if err != nil {
		// กรณีไม่พบการตั้งค่า ส่ง 500 กลับไป
		return NewAppError(500, "ไม่พบการตั้งค่าระบบเลือกตั้งที่ใช้งานอยู่")
	}

	// 2. ด่านตรวจสอบสถานะระบบและเวลาการลงคะแนน
	now := time.Now()
	statusLower := strings.ToLower(config.Status) // แปลงเป็นตัวพิมพ์เล็กป้องกันปัญหา OPEN vs open

	if statusLower != "open" || now.Before(config.StartTime) || now.After(config.EndTime) {
		return NewAppError(403, "อยู่นอกเวลาการลงคะแนน หรือระบบปิดรับคะแนนแล้ว")
	}

	// 3. เตรียมข้อมูลผลโหวต (ประกอบร่าง Data)
	var candidateID *uint
	if req.CandidateNo > 0 {
		cID := uint(req.CandidateNo)
		candidateID = &cID
	}

	var partyID *uint
	if req.PartyNo > 0 {
		pID := uint(req.PartyNo)
		partyID = &pID
	}

	// สร้าง struct models.Vote รอไว้ โดยไม่ใส่ VoterID เพื่อรักษา Secret Ballot
	voteRecord := models.Vote{
		AreaID:      areaID,
		CandidateID: candidateID,
		PartyID:     partyID,
		CreatedAt:   time.Now(),
	}

	// 4. สั่งให้ Repository ทำการบันทึกข้อมูลแบบ Transaction
	// Repository มีการจัดการส่ง AppError มาให้แล้วหากเกิดข้อผิดพลาด
	return s.repo.ExecuteVoteTransaction(voterID, voteRecord)
}

// GetBallotStatus รวบรวมข้อมูลสถานะระบบและสถานะผู้ใช้เข้าด้วยกัน
func (s *votingService) GetBallotStatus(voterID uint) (*BallotStatusResponse, error) {

	// 1. ตรวจสอบสถานะว่าโหวตหรือยัง
	isVoted, err := s.repo.CheckUserHasVoted(voterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewAppError(404, "ไม่พบข้อมูลผู้ใช้งานในระบบ")
		}
		return nil, NewAppError(500, "ไม่สามารถตรวจสอบประวัติการลงคะแนนของผู้ใช้ได้")
	}

	// 2. ตรวจสอบการตั้งค่าระบบเลือกตั้ง
	// หมายเหตุ: ยุบรวมไปเรียกใช้ GetActiveConfig ตัวเดียวกับตอน Submit เพื่อลดโค้ดซ้ำซ้อน
	config, err := s.repo.GetActiveConfig()
	if err != nil {
		// กรณีที่ฐานข้อมูลเพิ่งสร้างใหม่ ยังไม่มีแอดมินมาตั้งค่าใดๆ เลย
		// ให้ถือว่าระบบอยู่ในสถานะ "เตรียมการ (PREPARE)"
		return &BallotStatusResponse{
			ElectionStatus: "PREPARE",
			ServerTime:     time.Now(),
			IsVoted:        isVoted,
		}, nil
	}

	// 3. ประกอบร่างข้อมูล (BFF Pattern) เพื่อส่งกลับไปให้ด่านหน้า
	return &BallotStatusResponse{
		ElectionStatus: config.Status,
		StartTime:      config.StartTime,
		EndTime:        config.EndTime,
		ServerTime:     time.Now(), // ประทับตราเวลา Server ณ วินาทีนี้
		IsVoted:        isVoted,
	}, nil
}