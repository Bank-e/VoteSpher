package auth

import (
	"errors"
	"io"
	"math/big"
	"os"
	"testing"

	"votespher/internal/models"
	"votespher/pkg"
)

// --- Mock Repository ---

type mockAuthRepo struct {
	findVoterByIDResult            *models.Voter
	findVoterByIDErr               error
	findVoterByCitizenIDHashResult *models.Voter
	findVoterByCitizenIDHashErr    error
	createOTPErr                   error
	findOTPByRefCodeResult         *models.OTP
	findOTPByRefCodeErr            error
	markOTPAsUsedErr               error
	updateOTPAttemptsErr           error
	checkIsAdminResult             bool
	findVoterWithAreaResult        *models.Voter
	findVoterWithAreaErr           error
}

func (m *mockAuthRepo) FindVoterByID(voterID uint) (*models.Voter, error) {
	return m.findVoterByIDResult, m.findVoterByIDErr
}
func (m *mockAuthRepo) FindVoterByCitizenIDHash(hash string) (*models.Voter, error) {
	return m.findVoterByCitizenIDHashResult, m.findVoterByCitizenIDHashErr
}
func (m *mockAuthRepo) CreateOTP(otp *models.OTP) error { return m.createOTPErr }
func (m *mockAuthRepo) FindOTPByRefCode(refCode string) (*models.OTP, error) {
	return m.findOTPByRefCodeResult, m.findOTPByRefCodeErr
}
func (m *mockAuthRepo) MarkOTPAsUsed(otpID uint) error { return m.markOTPAsUsedErr }
func (m *mockAuthRepo) UpdateOTPAttempts(otpID uint, attempts int, markUsed bool) error {
	return m.updateOTPAttemptsErr
}
func (m *mockAuthRepo) CheckIsAdmin(voterID uint) bool { return m.checkIsAdminResult }
func (m *mockAuthRepo) FindVoterWithArea(voterID uint) (*models.Voter, error) {
	return m.findVoterWithAreaResult, m.findVoterWithAreaErr
}

func newTestAuthService(repo AuthRepository, sendEmail func(string, string, string) error, sendSMS func(string, string) error) *authService {
	svc := &authService{repo: repo}
	if sendEmail != nil {
		svc.sendEmail = sendEmail
	} else {
		svc.sendEmail = func(to, otp, ref string) error { return nil }
	}
	if sendSMS != nil {
		svc.sendSMS = sendSMS
	} else {
		svc.sendSMS = func(to, body string) error { return nil }
	}
	svc.generateToken = pkg.GenerateToken
	return svc
}

// 1. ทดสอบการสุ่ม OTP (ต้องได้ 6 หลักเสมอ)
func TestGenerateRandomOTP(t *testing.T) {
	otp, err := generateRandomOTP()
	if err != nil {
		t.Fatalf("สุ่ม OTP พัง: %v", err)
	}

	if len(otp) != 6 {
		t.Errorf("OTP ต้องมี 6 หลัก แต่ได้มา: %s", otp)
	}
}

// 2. ทดสอบการสุ่ม Ref Code (ต้องได้ 6 ตัวอักษร hex)
func TestGenerateRefCode(t *testing.T) {
	ref, err := generateRefCode()
	if err != nil {
		t.Fatalf("สุ่ม Ref Code พัง: %v", err)
	}

	if len(ref) != 6 {
		t.Errorf("Ref Code ต้องมี 6 ตัว แต่ได้มา: %s", ref)
	}
}

// 3. ทดสอบการ Hash Citizen ID (ต้องได้ค่าเดิมเสมอถ้า Key เดิม)
func TestGenerateCitizenIDHash(t *testing.T) {
	// ตั้งค่า Secret ชั่วคราวสำหรับการเทสต์
	os.Setenv("HASH_SECRET_KEY", "test_secret_key")

	id := "1234567890123"
	hash1 := generateCitizenIDHash(id)
	hash2 := generateCitizenIDHash(id)

	if hash1 == "" {
		t.Error("Hash ที่ได้ต้องไม่ว่างเปล่า")
	}

	if hash1 != hash2 {
		t.Error("Hash ค่าเดิมต้องได้ผลลัพธ์เดิม (Idempotent)")
	}
}

// 4. ทดสอบกรณีที่ Secret Key เปลี่ยนไป (Hash ต้องเปลี่ยน)
func TestGenerateCitizenIDHash_DifferentKey(t *testing.T) {
	id := "1234567890123"

	os.Setenv("HASH_SECRET_KEY", "key_1")
	hash1 := generateCitizenIDHash(id)

	os.Setenv("HASH_SECRET_KEY", "key_2")
	hash2 := generateCitizenIDHash(id)

	if hash1 == hash2 {
		t.Error("Secret Key ต่างกัน Hash ไม่ควรจะเหมือนกันนะ!")
	}
}

// --- VerifyVoter ---

func TestVerifyVoter_Success(t *testing.T) {
	t.Setenv("HASH_SECRET_KEY", "test_key")
	repo := &mockAuthRepo{
		findVoterByCitizenIDHashResult: &models.Voter{
			ID:     1,
			AreaID: 1,
			Area: models.Area{
				AreaName: "Area 1",
				Province: models.Province{ProvinceName: "Bangkok"},
			},
			Email:       "test@example.com",
			PhoneNumber: "0812345678",
		},
	}
	svc := newTestAuthService(repo, nil, nil)
	res, err := svc.VerifyVoter(VerifyVoterRequest{CitizenID: "1234567890123"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.VoterID != 1 {
		t.Errorf("expected voterID=1, got %d", res.VoterID)
	}
}

func TestVerifyVoter_NotFound(t *testing.T) {
	t.Setenv("HASH_SECRET_KEY", "test_key")
	repo := &mockAuthRepo{findVoterByCitizenIDHashErr: errors.New("not found")}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.VerifyVoter(VerifyVoterRequest{CitizenID: "1234567890123"}); err == nil {
		t.Error("expected error")
	}
}

// --- RequestOTP ---

func TestRequestOTP_MockMode(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "true")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, Email: "test@example.com"}}
	svc := newTestAuthService(repo, nil, nil)
	res, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.OTPCode != "111111" {
		t.Errorf("expected mock OTP 111111, got %s", res.OTPCode)
	}
}

func TestRequestOTP_MockMode_SMSChannel(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "true")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, PhoneNumber: "0812345678"}}
	svc := newTestAuthService(repo, nil, nil)
	res, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1, DeliveryChannel: "sms"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.OTPCode != "111111" {
		t.Errorf("expected mock OTP, got %s", res.OTPCode)
	}
}

func TestRequestOTP_VoterNotFound(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	repo := &mockAuthRepo{findVoterByIDErr: errors.New("not found")}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 99}); err == nil {
		t.Error("expected error")
	}
}

func TestRequestOTP_EmailMode_Success(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "email")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, Email: "test@example.com"}}
	emailCalled := false
	svc := newTestAuthService(repo, func(to, otp, ref string) error {
		emailCalled = true
		return nil
	}, nil)
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !emailCalled {
		t.Error("expected email to be called")
	}
}

func TestRequestOTP_EmailMode_NoEmail(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "email")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, Email: ""}}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1}); err == nil {
		t.Error("expected error for no email")
	}
}

func TestRequestOTP_EmailMode_SendFail(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "email")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, Email: "test@example.com"}}
	svc := newTestAuthService(repo, func(to, otp, ref string) error {
		return errors.New("smtp down")
	}, nil)
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1}); err == nil {
		t.Error("expected error from email failure")
	}
}

func TestRequestOTP_SMSMode_Success(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, PhoneNumber: "0812345678"}}
	smsCalled := false
	svc := newTestAuthService(repo, nil, func(to, body string) error {
		smsCalled = true
		return nil
	})
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1, DeliveryChannel: "sms"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !smsCalled {
		t.Error("expected SMS to be called")
	}
}

func TestRequestOTP_SMSMode_NoPhone(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, PhoneNumber: ""}}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1, DeliveryChannel: "sms"}); err == nil {
		t.Error("expected error for no phone")
	}
}

func TestRequestOTP_SMSMode_SendFail(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, PhoneNumber: "0812345678"}}
	svc := newTestAuthService(repo, nil, func(to, body string) error {
		return errors.New("sms down")
	})
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1, DeliveryChannel: "sms"}); err == nil {
		t.Error("expected error from SMS failure")
	}
}

func TestRequestOTP_CreateOTPFail(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "true")
	repo := &mockAuthRepo{
		findVoterByIDResult: &models.Voter{ID: 1, Email: "test@example.com"},
		createOTPErr:        errors.New("db error"),
	}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1}); err == nil {
		t.Error("expected error")
	}
}

// --- ConfirmOTP ---

func TestConfirmOTP_Success_VoterRole(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test_secret")
	otp := &models.OTP{ID: 1, VoterID: 1, OTPCode: "123456", Attempts: 0}
	voter := &models.Voter{ID: 1, AreaID: 1}
	repo := &mockAuthRepo{findOTPByRefCodeResult: otp, findVoterByIDResult: voter, checkIsAdminResult: false}
	svc := newTestAuthService(repo, nil, nil)
	res, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Token == "" {
		t.Error("expected token")
	}
	if res.Role != "voter" {
		t.Errorf("expected voter role, got %s", res.Role)
	}
}

func TestConfirmOTP_Success_AdminRole(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test_secret")
	otp := &models.OTP{ID: 1, VoterID: 1, OTPCode: "123456", Attempts: 0}
	voter := &models.Voter{ID: 1, AreaID: 1}
	repo := &mockAuthRepo{findOTPByRefCodeResult: otp, findVoterByIDResult: voter, checkIsAdminResult: true}
	svc := newTestAuthService(repo, nil, nil)
	res, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"})
	if err != nil || res.Role != "admin" {
		t.Errorf("expected admin role, got %v, err=%v", res, err)
	}
}

func TestConfirmOTP_RefCodeNotFound(t *testing.T) {
	repo := &mockAuthRepo{findOTPByRefCodeErr: errors.New("not found")}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "badref"}); err == nil {
		t.Error("expected error")
	}
}

func TestConfirmOTP_WrongCode(t *testing.T) {
	otp := &models.OTP{ID: 1, VoterID: 1, OTPCode: "999999", Attempts: 0}
	repo := &mockAuthRepo{findOTPByRefCodeResult: otp}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"}); err == nil {
		t.Error("expected error for wrong OTP")
	}
}

func TestConfirmOTP_MaxAttempts(t *testing.T) {
	otp := &models.OTP{ID: 1, VoterID: 1, OTPCode: "999999", Attempts: 4}
	repo := &mockAuthRepo{findOTPByRefCodeResult: otp}
	svc := newTestAuthService(repo, nil, nil)
	_, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"})
	if err == nil {
		t.Fatal("expected error for max attempts")
	}
	if err.Error() != "กรอก OTP ผิดเกิน 5 ครั้ง รหัสถูกยกเลิกแล้ว กรุณาขอ OTP ใหม่" {
		t.Errorf("unexpected message: %s", err.Error())
	}
}

func TestConfirmOTP_VoterNotFound(t *testing.T) {
	otp := &models.OTP{ID: 1, VoterID: 99, OTPCode: "123456", Attempts: 0}
	repo := &mockAuthRepo{findOTPByRefCodeResult: otp, findVoterByIDErr: errors.New("not found")}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"}); err == nil {
		t.Error("expected error")
	}
}

func TestConfirmOTP_MissingJWTKey(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "")
	otp := &models.OTP{ID: 1, VoterID: 1, OTPCode: "123456", Attempts: 0}
	voter := &models.Voter{ID: 1, AreaID: 1}
	repo := &mockAuthRepo{findOTPByRefCodeResult: otp, findVoterByIDResult: voter}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"}); err == nil {
		t.Error("expected error for missing JWT key")
	}
}

// --- maskEmail / maskPhone ---

func TestMaskEmail(t *testing.T) {
	tests := []struct{ input, want string }{
		{"test@example.com", "t***@example.com"},
		{"a@b.com", "***@b.com"},
		{"", ""},
		{"noemail", "***"},
	}
	for _, tt := range tests {
		got := maskEmail(tt.input)
		if got != tt.want {
			t.Errorf("maskEmail(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0812345678", "081*****78"},
		{"12345", "***"},
		{"123456", "123*56"},
	}
	for _, tt := range tests {
		got := maskPhone(tt.input)
		if got != tt.want {
			t.Errorf("maskPhone(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Crypto error branches ---

func TestRequestOTP_OTPRandError(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "email")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, Email: "test@example.com"}}
	svc := newTestAuthService(repo, nil, nil)

	old := cryptoRandInt
	cryptoRandInt = func(r io.Reader, max *big.Int) (*big.Int, error) {
		return nil, errors.New("rand error")
	}
	t.Cleanup(func() { cryptoRandInt = old })

	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1}); err == nil {
		t.Error("expected error from OTP rand failure")
	}
}

func TestRequestOTP_RefRandError(t *testing.T) {
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "email")
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, Email: "test@example.com"}}
	svc := newTestAuthService(repo, nil, nil)

	oldRandInt := cryptoRandInt
	cryptoRandInt = func(r io.Reader, max *big.Int) (*big.Int, error) {
		return big.NewInt(123456), nil
	}
	oldRandRead := cryptoRandRead
	cryptoRandRead = func(r io.Reader, b []byte) (int, error) {
		return 0, errors.New("rand read error")
	}
	t.Cleanup(func() {
		cryptoRandInt = oldRandInt
		cryptoRandRead = oldRandRead
	})

	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1}); err == nil {
		t.Error("expected error from ref code rand failure")
	}
}

func TestRequestOTP_InvalidDeliveryMode(t *testing.T) {
	// mode != "email" && mode != "sms" → fallback to "email"
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "fax") // invalid mode → defaults to email
	repo := &mockAuthRepo{findVoterByIDResult: &models.Voter{ID: 1, Email: "test@example.com"}}
	emailCalled := false
	svc := newTestAuthService(repo, func(to, otp, ref string) error {
		emailCalled = true
		return nil
	}, nil)
	if _, err := svc.RequestOTP(OTPRequestRequest{VoterID: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !emailCalled {
		t.Error("expected email fallback for invalid delivery mode")
	}
}

func TestConfirmOTP_GenerateTokenError(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test_secret")
	otp := &models.OTP{ID: 1, VoterID: 1, OTPCode: "123456", Attempts: 0}
	voter := &models.Voter{ID: 1, AreaID: 1}
	repo := &mockAuthRepo{findOTPByRefCodeResult: otp, findVoterByIDResult: voter}
	svc := newTestAuthService(repo, nil, nil)
	svc.generateToken = func(voterID, areaID uint, role, secret string) (string, error) {
		return "", errors.New("token generation failed")
	}
	if _, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"}); err == nil {
		t.Error("expected error from generateToken failure")
	}
}

func TestConfirmOTP_MarkOTPAsUsedError(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test_secret")
	otp := &models.OTP{ID: 1, VoterID: 1, OTPCode: "123456", Attempts: 0}
	voter := &models.Voter{ID: 1, AreaID: 1}
	repo := &mockAuthRepo{
		findOTPByRefCodeResult: otp,
		findVoterByIDResult:    voter,
		markOTPAsUsedErr:       errors.New("db error"),
	}
	svc := newTestAuthService(repo, nil, nil)
	if _, err := svc.ConfirmOTP(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"}); err == nil {
		t.Error("expected error from MarkOTPAsUsed failure")
	}
}
