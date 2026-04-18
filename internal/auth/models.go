package auth

// POST /voter/verify
type VerifyVoterRequest struct {
	CitizenID string `json:"citizen_id"` // บัตรประชาชน 13 หลัก
}

type VoterInfo struct {
	Name     string `json:"name"`
	AreaID   uint   `json:"area_id"`
	AreaName string `json:"area_name"`
	Province string `json:"province"`
	IsVoted  bool   `json:"is_voted"`
}

type VerifyVoterResponse struct {
	VoterID   uint      `json:"voter_id"`
	VoterInfo VoterInfo `json:"voter_info"`
}

// POST /voter/otp-confirm
type OTPConfirmRequest struct {
	OTPCode string `json:"otp_code"` // รหัส OTP 6 หลัก
	RefCode string `json:"ref_code"` // ref code ที่ได้รับตอน request OTP
}

type OTPConfirmResponse struct {
	Token string `json:"token"` // JWT token
}

// POST /voter/otp-request
type OTPRequestRequest struct {
	VoterID uint `json:"voter_id" binding:"required"`
}

type OTPRequestResponse struct {
	RefCode string `json:"ref_code"`
	OTPCode string `json:"otp_code"` // คืนให้เพื่อใช้ทดสอบ (ในระบบจริงจะส่ง SMS)
}
