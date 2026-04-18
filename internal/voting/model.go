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