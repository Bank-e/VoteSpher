package voting

// SubmitBallotRequest รับข้อมูลจากผู้โหวต
type SubmitBallotRequest struct {
	CandidateNo int `json:"candidate_no"`
	PartyNo     int `json:"party_no"`
}