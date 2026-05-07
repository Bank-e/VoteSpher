package voting

import (
	"errors"
	"testing"
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. สร้าง Mock Repository เพื่อใช้ในการทดสอบ Service โดยไม่ต้องพึ่งพา Database จริง

// MockVotingRepository เป็นตัวแทน (Mock) ของ VotingRepository เพื่อใช้สำหรับเทส
type MockVotingRepository struct {
	mock.Mock
}

func (m *MockVotingRepository) GetActiveConfig() (*models.SystemConfig, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*models.SystemConfig), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockVotingRepository) ExecuteVoteTransaction(voterID uint, voteRecord models.Vote) error {
	args := m.Called(voterID, voteRecord)
	return args.Error(0)
}

func (m *MockVotingRepository) CheckUserHasVoted(voterID uint) (bool, error) {
	args := m.Called(voterID)
	return args.Bool(0), args.Error(1)
}


// 2. Test Cases สำหรับ SubmitVote

func TestSubmitVote_Success(t *testing.T) {
	// 1. Setup Mock
	mockRepo := new(MockVotingRepository)

	// จำลองว่าระบบเปิดให้โหวต (เวลาครอบคลุมปัจจุบัน)
	activeConfig := &models.SystemConfig{
		Status:    "OPEN",
		StartTime: time.Now().Add(-1 * time.Hour), // เปิดมาแล้ว 1 ชั่วโมง
		EndTime:   time.Now().Add(1 * time.Hour),  // จะปิดในอีก 1 ชั่วโมง
	}

	// สอน Mock ว่าถ้ามีการเรียกฟังก์ชันเหล่านี้ ให้ตอบกลับไปว่าอะไร
	mockRepo.On("GetActiveConfig").Return(activeConfig, nil)
	// mock.Anything แปลว่ารับค่าอะไรมาก็ได้ในพารามิเตอร์นั้น
	mockRepo.On("ExecuteVoteTransaction", uint(123), mock.Anything).Return(nil)

	// 2. สร้าง Service โดยยัด Mock Repo เข้าไป
	service := NewVotingService(mockRepo)

	// 3. เตรียมข้อมูล Request และทดสอบรัน
	req := SubmitBallotRequest{CandidateNo: 1, PartyNo: 2}
	err := service.SubmitVote(123, 10, req)

	// 4. Assert (ตรวจสอบผล)
	assert.NoError(t, err, "การโหวตที่ถูกต้อง ต้องไม่มี Error เกิดขึ้น")
	mockRepo.AssertExpectations(t) // ยืนยันว่า Mock ถูกเรียกใช้งานครบตามที่ตั้งไว้
}

func TestSubmitVote_Fail_SystemClosed(t *testing.T) {
	// 1. Setup Mock
	mockRepo := new(MockVotingRepository)

	// จำลองว่าระบบปิดแล้ว (PAUSED)
	closedConfig := &models.SystemConfig{
		Status:    "PAUSED",
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
	}

	mockRepo.On("GetActiveConfig").Return(closedConfig, nil)
	// สังเกตว่าเราไม่ต้องสอน Mock เรื่อง ExecuteVoteTransaction เพราะ Logic ควรจะตีตกไปก่อนที่จะถึงขั้นบันทึกลง DB

	// 2. เตรียม Service
	service := NewVotingService(mockRepo)

	// 3. รันเทส
	req := SubmitBallotRequest{CandidateNo: 1, PartyNo: 2}
	err := service.SubmitVote(123, 10, req)

	// 4. Assert
	assert.Error(t, err, "ต้องมี Error กลับมาถ้าระบบปิด")
	
	// ตรวจสอบว่าเป็น AppError เบอร์ 403 ตามที่เขียนไว้ใน Business Logic หรือไม่
	appErr, ok := err.(*AppError)
	assert.True(t, ok, "Error ต้องเป็น Type AppError")
	assert.Equal(t, 403, appErr.Code, "HTTP Status ต้องเป็น 403 Forbidden")
	assert.Contains(t, appErr.Message, "อยู่นอกเวลา", "ข้อความ Error ต้องสื่อความหมายถูกต้อง")
}


// 3. Test Cases สำหรับ GetBallotStatus

func TestGetBallotStatus_UserAlreadyVoted(t *testing.T) {
	mockRepo := new(MockVotingRepository)

	// จำลองว่า User คนนี้ (ID: 999) เคยโหวตไปแล้ว (Return true)
	mockRepo.On("CheckUserHasVoted", uint(999)).Return(true, nil)
	
	activeConfig := &models.SystemConfig{
		Status:    "OPEN",
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
	}
	mockRepo.On("GetActiveConfig").Return(activeConfig, nil)

	service := NewVotingService(mockRepo)

	// รันเทส
	result, err := service.GetBallotStatus(999)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsVoted, "สถานะ IsVoted ต้องเป็น true")
	assert.Equal(t, "OPEN", result.ElectionStatus)
}

// ***************************************************************
func TestSubmitVote_Fail_ConfigError(t *testing.T) {
	mockRepo := new(MockVotingRepository)
	// จำลองว่า Database พังตอนดึง Config
	mockRepo.On("GetActiveConfig").Return(nil, errors.New("db error"))
	service := NewVotingService(mockRepo)

	err := service.SubmitVote(123, 10, SubmitBallotRequest{})
	assert.Error(t, err)
	appErr, _ := err.(*AppError)
	assert.Equal(t, 500, appErr.Code)
}

func TestGetBallotStatus_UserNotFound(t *testing.T) {
	mockRepo := new(MockVotingRepository)
	// จำลองว่าหา User ไม่เจอใน DB (ErrRecordNotFound)
	mockRepo.On("CheckUserHasVoted", uint(999)).Return(false, gorm.ErrRecordNotFound)
	service := NewVotingService(mockRepo)

	result, err := service.GetBallotStatus(999)
	assert.Error(t, err)
	assert.Nil(t, result)
	appErr, _ := err.(*AppError)
	assert.Equal(t, 404, appErr.Code)
}

func TestGetBallotStatus_PrepareState(t *testing.T) {
	mockRepo := new(MockVotingRepository)
	// จำลองว่า User ปกติ
	mockRepo.On("CheckUserHasVoted", uint(123)).Return(false, nil)
	// จำลองว่ายังไม่มีใครตั้งค่าระบบเลือกตั้ง (Config = error)
	mockRepo.On("GetActiveConfig").Return(nil, errors.New("not found config"))
	
	service := NewVotingService(mockRepo)
	result, err := service.GetBallotStatus(123)
	
	// ระบบต้องไม่พัง แต่ต้องตอบกลับว่าสถานะคือ PREPARE
	assert.NoError(t, err)
	assert.Equal(t, "PREPARE", result.ElectionStatus)
}

func TestGetBallotStatus_Fail_GenericDBError(t *testing.T) {
	mockRepo := new(MockVotingRepository)
	
	// แกล้งให้ Database พังแบบปกติ (ไม่ใช่แบบหาไม่เจอ)
	mockRepo.On("CheckUserHasVoted", uint(123)).Return(false, errors.New("database connection lost"))
	service := NewVotingService(mockRepo)

	result, err := service.GetBallotStatus(123)
	
	assert.Error(t, err)
	assert.Nil(t, result)
	
	// ต้องคืนค่าเป็น 500 (เกิดข้อผิดพลาดภายในระบบ)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, 500, appErr.Code)
}