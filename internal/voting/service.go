package voting

import (
	"errors"
	"time"
	"strings"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// SubmitVoteService จัดการขั้นตอนและกฎเกณฑ์ก่อนการลงคะแนน
// (สมมติว่า SubmitBallotRequest ยังอยู่ใน package voting ถ้าย้ายไป models แล้วก็ต้องเติม models. ด้วยครับ)
func SubmitVoteService(db *gorm.DB, voterID uint, areaID uint, req SubmitBallotRequest) error {
	
	// 1. ดึงการตั้งค่าระบบจาก Repository
	config, err := GetActiveConfig(db)
	if err != nil {
		return err
	}

	// 2. ด่านตรวจสอบสถานะระบบ
	now := time.Now()
	statusLower := strings.ToLower(config.Status) // แปลงเป็นตัวพิมพ์เล็กป้องกันปัญหา OPEN vs open
	
	if statusLower != "open" || now.Before(config.StartTime) || now.After(config.EndTime) {
		return errors.New("403: อยู่นอกเวลาการลงคะแนน หรือระบบปิดรับคะแนนแล้ว")
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
	return ExecuteVoteTransaction(db, voterID, voteRecord)
}

// GetBallotStatusService รวบรวมข้อมูลสถานะระบบและสถานะผู้ใช้เข้าด้วยกัน
func GetBallotStatusService(db *gorm.DB, voterID uint) (*BallotStatusResponse, error) {
	
	// 1. ตรวจสอบสถานะว่าโหวตหรือยัง
	isVoted, err := CheckUserHasVoted(db, voterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("404: ไม่พบข้อมูลผู้ใช้งานในระบบ")
		}
		return nil, errors.New("500: ไม่สามารถตรวจสอบประวัติการลงคะแนนของผู้ใช้ได้")
	}

	// 2. ตรวจสอบการตั้งค่าระบบเลือกตั้ง
	config, err := GetActiveElectionConfig(db)
	if err != nil {
		// กรณีที่ฐานข้อมูลเพิ่งสร้างใหม่ ยังไม่มีแอดมินมาตั้งค่าใดๆ เลย
		// ให้ถือว่าระบบอยู่ในสถานะ "เตรียมการ (PREPARE)"
		return &BallotStatusResponse{
			ElectionStatus: "PREPARE",
			ServerTime:     time.Now(),
			IsVoted:       isVoted,
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