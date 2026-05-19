package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"time"
	"votespher/internal/models"
	"votespher/pkg"
)

var (
	cryptoRandInt  = rand.Int
	cryptoRandRead = io.ReadFull
)

type AuthService interface {
	VerifyVoter(req VerifyVoterRequest) (*VerifyVoterResponse, error)
	RequestOTP(req OTPRequestRequest) (*OTPRequestResponse, error)
	ConfirmOTP(req OTPConfirmRequest) (*OTPConfirmResult, error)
}

type authService struct {
	repo          AuthRepository
	sendEmail     func(to, otp, refCode string) error
	sendSMS       func(to, body string) error
	generateToken func(voterID, areaID uint, role, secret string) (string, error)
}

func NewAuthService(repo AuthRepository) AuthService {
	return &authService{
		repo:          repo,
		sendEmail:     pkg.EnqueueOTPEmail,
		sendSMS:       pkg.SendSMS,
		generateToken: pkg.GenerateToken,
	}
}

func (s *authService) VerifyVoter(req VerifyVoterRequest) (*VerifyVoterResponse, error) {
	hashedID := generateCitizenIDHash(req.CitizenID)
	voter, err := s.repo.FindVoterByCitizenIDHash(hashedID)
	if err != nil {
		return nil, errors.New("ไม่พบผู้มีสิทธิ์เลือกตั้ง")
	}
	return &VerifyVoterResponse{
		VoterID: voter.ID,
		VoterInfo: VoterInfo{
			Name:        "ข้อมูลปกปิด",
			AreaID:      voter.AreaID,
			AreaName:    voter.Area.Province.ProvinceName + " " + voter.Area.AreaName,
			Province:    voter.Area.Province.ProvinceName,
			IsVoted:     voter.IsVoted,
			MaskedEmail: maskEmail(voter.Email),
			MaskedPhone: maskPhone(voter.PhoneNumber),
		},
	}, nil
}

func (s *authService) RequestOTP(req OTPRequestRequest) (*OTPRequestResponse, error) {
	voter, err := s.repo.FindVoterByID(req.VoterID)
	if err != nil {
		return nil, errors.New("ไม่พบรหัสผู้มีสิทธิ์โหวตนี้")
	}

	// dev mode always mocks — ENABLE_DEV_ENDPOINTS takes absolute priority
	var mode string
	if os.Getenv("ENABLE_DEV_ENDPOINTS") == "true" {
		mode = "mock"
	} else {
		mode = req.DeliveryChannel
		if mode == "" {
			mode = os.Getenv("OTP_DELIVERY_MODE")
		}
		if mode != "email" && mode != "sms" {
			mode = "email"
		}
	}

	var otpCode string
	if mode == "mock" {
		otpCode = "111111"
	} else {
		otpCode, err = generateRandomOTP()
		if err != nil {
			return nil, errors.New("สร้าง OTP ไม่สำเร็จ")
		}
	}

	refCode, err := generateRefCode()
	if err != nil {
		return nil, errors.New("สร้าง OTP ไม่สำเร็จ")
	}

	newOTP := models.OTP{
		VoterID:   req.VoterID,
		OTPCode:   otpCode,
		RefCode:   refCode,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := s.repo.CreateOTP(&newOTP); err != nil {
		return nil, fmt.Errorf("สร้าง OTP ไม่สำเร็จ: %w", err)
	}

	// determine target address: custom override or stored value
	targetEmail := req.DeliveryAddress
	targetPhone := req.DeliveryAddress

	switch mode {
	case "mock":
		// mock mode — return OTP in response, no real delivery
		maskedContact := ""
		if req.DeliveryChannel == "sms" {
			if targetPhone == "" {
				targetPhone = voter.PhoneNumber
			}
			maskedContact = maskPhone(targetPhone)
		} else {
			if targetEmail == "" {
				targetEmail = voter.Email
			}
			maskedContact = maskEmail(targetEmail)
		}
		return &OTPRequestResponse{OTPCode: otpCode, RefCode: refCode, MaskedContact: maskedContact}, nil

	case "sms":
		if targetPhone == "" {
			targetPhone = voter.PhoneNumber
		}
		if targetPhone == "" {
			return nil, errors.New("ไม่มีเบอร์โทรสำหรับส่ง OTP กรุณาใส่เบอร์โทร")
		}
		smsBody := fmt.Sprintf("รหัส OTP VoteSpher: %s (Ref: %s) หมดอายุใน 5 นาที", otpCode, refCode)
		if err := s.sendSMS(targetPhone, smsBody); err != nil {
			return nil, fmt.Errorf("ส่ง SMS ไม่สำเร็จ กรุณาลองใหม่")
		}
		return &OTPRequestResponse{RefCode: refCode, MaskedContact: maskPhone(targetPhone)}, nil

	default: // email
		if targetEmail == "" {
			targetEmail = voter.Email
		}
		if targetEmail == "" {
			return nil, errors.New("ไม่มี email สำหรับส่ง OTP กรุณาใส่ email")
		}
		if err := s.sendEmail(targetEmail, otpCode, refCode); err != nil {
			return nil, errors.New("ส่ง OTP ไม่สำเร็จ กรุณาลองใหม่")
		}
		return &OTPRequestResponse{RefCode: refCode, MaskedContact: maskEmail(targetEmail)}, nil
	}
}

func (s *authService) ConfirmOTP(req OTPConfirmRequest) (*OTPConfirmResult, error) {
	otp, err := s.repo.FindOTPByRefCode(req.RefCode)
	if err != nil {
		return nil, errors.New("ref_code ไม่ถูกต้องหรือ OTP หมดอายุแล้ว")
	}

	if otp.OTPCode != req.OTPCode {
		newAttempts := otp.Attempts + 1
		markUsed := newAttempts >= 5
		_ = s.repo.UpdateOTPAttempts(otp.ID, newAttempts, markUsed)
		if markUsed {
			return nil, errors.New("กรอก OTP ผิดเกิน 5 ครั้ง รหัสถูกยกเลิกแล้ว กรุณาขอ OTP ใหม่")
		}
		return nil, fmt.Errorf("รหัส OTP ไม่ถูกต้อง (ครั้งที่ %d/5)", newAttempts)
	}

	if err := s.repo.MarkOTPAsUsed(otp.ID); err != nil {
		return nil, err
	}

	voter, err := s.repo.FindVoterByID(otp.VoterID)
	if err != nil {
		return nil, err
	}

	role := "voter"
	if s.repo.CheckIsAdmin(voter.ID) {
		role = "admin"
	}

	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("ระบบผิดพลาด กรุณาติดต่อผู้ดูแล")
	}
	token, err := s.generateToken(voter.ID, voter.AreaID, role, secretKey)
	if err != nil {
		return nil, err
	}

	return &OTPConfirmResult{Token: token, Role: role}, nil
}

// --- Private helpers ---

func generateCitizenIDHash(citizenID string) string {
	secretKey := []byte(os.Getenv("HASH_SECRET_KEY"))
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(citizenID))
	return hex.EncodeToString(h.Sum(nil))
}

func generateRandomOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := cryptoRandInt(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func generateRefCode() (string, error) {
	b := make([]byte, 3)
	if _, err := cryptoRandRead(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// maskEmail: "piyachat.sal@dome.tu.ac.th" → "p***@dome.tu.ac.th"
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "***"
	}
	local := parts[0]
	if len(local) <= 1 {
		return "***@" + parts[1]
	}
	return string(local[0]) + "***@" + parts[1]
}

// maskPhone: "0929400592" → "092*****92"
func maskPhone(phone string) string {
	if len(phone) < 6 {
		return "***"
	}
	return phone[:3] + strings.Repeat("*", len(phone)-5) + phone[len(phone)-2:]
}
