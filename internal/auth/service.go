package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"
	"votespher/internal/models" // DB Models
	"votespher/pkg"
)

type AuthService interface {
	ConfirmOTP(req OTPConfirmRequest) (string, error) // คืนค่าแค่ Token string
	VerifyVoter(req VerifyVoterRequest) (*VerifyVoterResponse, error)
	RequestOTP(req OTPRequestRequest) (*OTPRequestResponse, error)
}

type authService struct {
	repo AuthRepository
}

func NewAuthService(repo AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) ConfirmOTP(req OTPConfirmRequest) (string, error) {
	otp, err := s.repo.FindOTPByRefCode(req.RefCode)
	if err != nil {
		return "", errors.New("ref_code ไม่ถูกต้องหรือ OTP หมดอายุแล้ว")
	}

	if otp.OTPCode != req.OTPCode {
		return "", errors.New("รหัส OTP ไม่ถูกต้อง")
	}

	if err := s.repo.MarkOTPAsUsed(otp.ID); err != nil {
		return "", err
	}

	voter, err := s.repo.FindVoterByID(otp.VoterID)
	if err != nil {
		return "", err
	}

	secretKey := os.Getenv("JWT_SECRET_KEY")
	token, err := pkg.GenerateToken(voter.ID, voter.AreaID, "voter", secretKey)
	if err != nil {
		return "", err
	}

	return token, nil
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
			Name:     "ข้อมูลปกปิด",
			AreaID:   voter.AreaID,
			AreaName: voter.Area.AreaName,
			Province: voter.Area.Province.ProvinceName,
			IsVoted:  voter.IsVoted,
		},
	}, nil
}

func (s *authService) RequestOTP(req OTPRequestRequest) (*OTPRequestResponse, error) {
	_, err := s.repo.FindVoterByID(req.VoterID)
	if err != nil {
		return nil, errors.New("ไม่พบรหัสผู้มีสิทธิ์โหวตนี้")
	}

	otpCode, _ := generateRandomOTP()
	refCode, _ := generateRefCode()

	newOTP := models.OTP{
		VoterID:   req.VoterID,
		OTPCode:   otpCode,
		RefCode:   refCode,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := s.repo.CreateOTP(&newOTP); err != nil {
		return nil, errors.New("สร้าง OTP ไม่สำเร็จ")
	}

	return &OTPRequestResponse{
		RefCode: refCode,
		OTPCode: otpCode,
	}, nil
}

// --- Private Helpers ---
func generateCitizenIDHash(citizenID string) string {
	secretKey := []byte(os.Getenv("HASH_SECRET_KEY"))
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(citizenID))
	return hex.EncodeToString(h.Sum(nil))
}

func generateRandomOTP() (string, error) {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func generateRefCode() (string, error) {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b), nil
}
