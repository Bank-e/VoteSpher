package auth

// POST /voter/verify
type VerifyVoterRequest struct {
	CitizenID string `json:"citizen_id"`
}

type VoterInfo struct {
	Name        string `json:"name"`
	AreaID      uint   `json:"area_id"`
	AreaName    string `json:"area_name"`
	Province    string `json:"province"`
	IsVoted     bool   `json:"is_voted"`
	MaskedEmail string `json:"masked_email"`
	MaskedPhone string `json:"masked_phone"`
}

type VerifyVoterResponse struct {
	VoterID   uint      `json:"voter_id"`
	VoterInfo VoterInfo `json:"voter_info"`
}

// POST /voter/otp-confirm
type OTPConfirmRequest struct {
	OTPCode string `json:"otp_code"`
	RefCode string `json:"ref_code"`
}

type OTPConfirmResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

// POST /voter/otp-request
type OTPRequestRequest struct {
	VoterID         uint   `json:"voter_id" binding:"required"`
	DeliveryChannel string `json:"delivery_channel"` // "email" | "sms"
	DeliveryAddress string `json:"delivery_address"` // custom email/phone — overrides stored value
}

type OTPRequestResponse struct {
	OTPCode       string `json:"otp_code,omitempty"` // แสดงเฉพาะโหมด mock
	RefCode       string `json:"ref_code"`
	MaskedContact string `json:"masked_contact"` // where OTP was sent (masked)
}

// ผลลัพธ์หลังยืนยัน OTP สำเร็จ
type OTPConfirmResult struct {
	Token string
	Role  string
}
