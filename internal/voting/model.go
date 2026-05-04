package voting

import "time"

// SubmitBallotRequest รับข้อมูลจากผู้โหวต
type SubmitBallotRequest struct {
	CandidateNo int `json:"candidate_no"`
	PartyNo     int `json:"party_no"`
}

type BallotStatusResponse struct {
	ElectionStatus string    `json:"election_status"` // PREPARE, OPEN, PAUSED, CLOSED
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	ServerTime     time.Time `json:"server_time"`     // เวลาปัจจุบันของ Server
	IsVoted        bool      `json:"is_voted"`       // สถานะการโหวตของ User คนนี้
}

// ==========================================
// Custom Error สำหรับจัดการ HTTP Status
// ==========================================

// AppError โครงสร้าง Error ที่สามารถระบุ HTTP Status Code ได้
type AppError struct {
	Code    int
	Message string
}

// Error ทำให้ AppError รองรับ interface error มาตรฐานของ Go
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError เป็น Helper function ในการสร้าง AppError ให้ใช้งานง่ายขึ้น
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}