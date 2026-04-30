package voting

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==========================================
// 🟢 1. สร้าง Mock Service
// ==========================================
type MockVotingService struct {
	mock.Mock
}

func (m *MockVotingService) SubmitVote(voterID uint, areaID uint, req SubmitBallotRequest) error {
	args := m.Called(voterID, areaID, req)
	return args.Error(0)
}

func (m *MockVotingService) GetBallotStatus(voterID uint) (*BallotStatusResponse, error) {
	args := m.Called(voterID)
	if args.Get(0) != nil {
		return args.Get(0).(*BallotStatusResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// ==========================================
// 🟢 2. Test Cases สำหรับ Handler
// ==========================================

func TestSubmitBallotHandler_Success(t *testing.T) {
	// ปิด Log ของ Gin ชั่วคราวตอนรันเทส
	gin.SetMode(gin.TestMode)

	// 1. Setup Mock Service
	mockService := new(MockVotingService)
	handler := NewVotingHandler(mockService)

	// สอน Mock ว่าถ้ามีการเรียก SubmitVote ให้ตอบว่าไม่มี Error (สำเร็จ)
	mockService.On("SubmitVote", uint(123), uint(10), mock.Anything).Return(nil)

	// 2. จำลอง HTTP Request & Response (เหมือน Postman)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// จำลองค่าที่ได้จาก Middleware (Token)
	c.Set("voter_id", uint(123))
	c.Set("area_id", uint(10))

	// จำลอง Body (JSON)
	reqBody := SubmitBallotRequest{CandidateNo: 1, PartyNo: 2}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/ballot/submit", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	// 3. รัน Handler
	handler.SubmitBallotHandler()(c)

	// 4. ตรวจสอบผลลัพธ์
	assert.Equal(t, http.StatusCreated, w.Code, "ถ้าสำเร็จต้องตอบ 201 Created")
	
	// ตรวจสอบ JSON Response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	
	mockService.AssertExpectations(t)
}

func TestSubmitBallotHandler_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVotingService)
	handler := NewVotingHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ❌ ไม่ได้ Set voter_id เข้าไปใน Context (จำลองเคส Token ไม่มีค่า)
	
	reqBody := SubmitBallotRequest{CandidateNo: 1, PartyNo: 2}
	jsonValue, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/ballot/submit", bytes.NewBuffer(jsonValue))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SubmitBallotHandler()(c)

	// ต้องตอบ 401 ทันที โดยที่ยังไม่ทันเรียก Service ด้วยซ้ำ
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetBallotStatusHandler_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVotingService)
	handler := NewVotingHandler(mockService)

	// สอน Mock ให้พ่น AppError 404 ออกมา (จำลองหา User ไม่เจอ)
	mockService.On("GetBallotStatus", uint(999)).Return(nil, NewAppError(404, "ไม่พบข้อมูลผู้ใช้งาน"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("voter_id", uint(999)) // ใส่ User 999 

	c.Request, _ = http.NewRequest(http.MethodGet, "/ballot/status", nil)

	handler.GetBallotStatusHandler()(c)

	// ตรวจสอบว่า Handler แปลง AppError กลับเป็น HTTP Status ได้ถูกต้องไหม
	assert.Equal(t, http.StatusNotFound, w.Code, "Handler ต้องแกะโค้ด 404 ออกมาจาก AppError ได้")
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "ไม่พบข้อมูลผู้ใช้งาน", response["message"])
}

// *********************************************************
func TestSubmitBallotHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewVotingHandler(new(MockVotingService))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ส่ง JSON พังๆ เข้าไป (ขาดปีกกาปิด)
	c.Request, _ = http.NewRequest(http.MethodPost, "/ballot/submit", bytes.NewBufferString(`{"candidate_no": 1`))
	
	handler.SubmitBallotHandler()(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitBallotHandler_InvalidTypeInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewVotingHandler(new(MockVotingService))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// แกล้งใส่ voter_id เป็นข้อความ (String) แทนที่จะเป็นตัวเลข (uint)
	c.Set("voter_id", "NOT_A_NUMBER")
	c.Set("area_id", uint(10))
	c.Request, _ = http.NewRequest(http.MethodPost, "/ballot/submit", bytes.NewBufferString(`{"candidate_no": 1}`))
	
	handler.SubmitBallotHandler()(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetBallotStatusHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVotingService)
	handler := NewVotingHandler(mockService)

	// สอน Mock ให้คืนค่า Success
	mockData := &BallotStatusResponse{ElectionStatus: "OPEN", IsVoted: true}
	mockService.On("GetBallotStatus", uint(123)).Return(mockData, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("voter_id", uint(123)) 
	c.Request, _ = http.NewRequest(http.MethodGet, "/ballot/status", nil)

	handler.GetBallotStatusHandler()(c)

	assert.Equal(t, http.StatusOK, w.Code) // ทดสอบบรรทัดสีแดง 200 OK
}

func TestSubmitBallotHandler_InvalidAreaIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewVotingHandler(new(MockVotingService))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("voter_id", uint(123))
	// แกล้งใส่ area_id เป็นข้อความแทนที่จะเป็นตัวเลข
	c.Set("area_id", "NOT_A_NUMBER") 
	c.Request, _ = http.NewRequest(http.MethodPost, "/ballot/submit", bytes.NewBufferString(`{"candidate_no": 1}`))

	handler.SubmitBallotHandler()(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSubmitBallotHandler_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVotingService)
	handler := NewVotingHandler(mockService)

	// สอนให้ Mock คืนค่า Error ธรรมดา ที่ไม่ใช่ AppError
	mockService.On("SubmitVote", uint(123), uint(10), mock.Anything).Return(errors.New("fatal server crash"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("voter_id", uint(123))
	c.Set("area_id", uint(10))
	c.Request, _ = http.NewRequest(http.MethodPost, "/ballot/submit", bytes.NewBufferString(`{"candidate_no": 1}`))

	handler.SubmitBallotHandler()(c)
	
	// ต้องตกไปเข้าเงื่อนไข Error 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetBallotStatusHandler_NoVoterID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewVotingHandler(new(MockVotingService))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// ไม่ Set ข้อมูลอะไรใน Context เลย (จำลองกรณีไม่มี Token)
	c.Request, _ = http.NewRequest(http.MethodGet, "/ballot/status", nil)

	handler.GetBallotStatusHandler()(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetBallotStatusHandler_InvalidVoterIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewVotingHandler(new(MockVotingService))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// แกล้งใส่ voter_id เป็นข้อความ
	c.Set("voter_id", "INVALID_ID") 
	c.Request, _ = http.NewRequest(http.MethodGet, "/ballot/status", nil)

	handler.GetBallotStatusHandler()(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetBallotStatusHandler_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockVotingService)
	handler := NewVotingHandler(mockService)

	// สอนให้ Mock คืนค่า Error ธรรมดา
	mockService.On("GetBallotStatus", uint(123)).Return(nil, errors.New("unexpected error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("voter_id", uint(123))
	c.Request, _ = http.NewRequest(http.MethodGet, "/ballot/status", nil)

	handler.GetBallotStatusHandler()(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}