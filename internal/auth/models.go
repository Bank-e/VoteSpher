import auth

type VerifyVoterRequest struct {
    CitizenID string `json:"citizen_id" binding:"required,len=13,numeric"` // บังคับ 13 หลักและเป็นตัวเลข
}

type VerifyVoterResponse struct {
    VoterID   uint      `json:"voter_id"`
    VoterInfo VoterInfo `json:"voter_info"`
}

type VoterInfo struct {
    Name      string `json:"name"`
    AreaID    uint   `json:"area_id"`
    AreaName  string `json:"area_name"`
    Province  string `json:"province"`
    IsVoted   bool   `json:"is_voted"`
}