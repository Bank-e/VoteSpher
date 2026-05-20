package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	netsmtp "net/smtp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"votespher/internal/models"
	"votespher/pkg"
)

// --- MockAuthService ---

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyVoter(req VerifyVoterRequest) (*VerifyVoterResponse, error) {
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*VerifyVoterResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) RequestOTP(req OTPRequestRequest) (*OTPRequestResponse, error) {
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*OTPRequestResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthService) ConfirmOTP(req OTPConfirmRequest) (*OTPConfirmResult, error) {
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*OTPConfirmResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func setupHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	db.AutoMigrate(&models.Province{}, &models.Area{}, &models.Voter{}, &models.OTP{}, &models.Admin{})
	return db
}

// --- VerifyVoter ---

func TestVerifyVoterHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(MockAuthService)
	h := NewAuthHandler(svc, nil)
	svc.On("VerifyVoter", mock.Anything).Return(&VerifyVoterResponse{VoterID: 1}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(VerifyVoterRequest{CitizenID: "1234567890123"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/verify", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.VerifyVoter(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestVerifyVoterHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(new(MockAuthService), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/verify", bytes.NewBufferString("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.VerifyVoter(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyVoterHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(MockAuthService)
	h := NewAuthHandler(svc, nil)
	svc.On("VerifyVoter", mock.Anything).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(VerifyVoterRequest{CitizenID: "1234567890123"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/verify", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.VerifyVoter(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- OTPRequest ---

func TestOTPRequestHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(MockAuthService)
	h := NewAuthHandler(svc, nil)
	svc.On("RequestOTP", mock.Anything).Return(&OTPRequestResponse{RefCode: "abc123"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.OTPRequest(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOTPRequestHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(new(MockAuthService), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBufferString("bad"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.OTPRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPRequestHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(MockAuthService)
	h := NewAuthHandler(svc, nil)
	svc.On("RequestOTP", mock.Anything).Return(nil, errors.New("service error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.OTPRequest(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- OTPConfirm ---

func TestOTPConfirmHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(MockAuthService)
	h := NewAuthHandler(svc, nil)
	svc.On("ConfirmOTP", mock.Anything).Return(&OTPConfirmResult{Token: "tok", Role: "voter"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPConfirmRequest{OTPCode: "123456", RefCode: "abc123"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.OTPConfirm(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOTPConfirmHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(new(MockAuthService), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBufferString("bad"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.OTPConfirm(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPConfirmHandler_EmptyOTPCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(new(MockAuthService), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPConfirmRequest{OTPCode: "", RefCode: "abc123"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.OTPConfirm(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOTPConfirmHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(MockAuthService)
	h := NewAuthHandler(svc, nil)
	svc.On("ConfirmOTP", mock.Anything).Return(nil, errors.New("invalid OTP"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPConfirmRequest{OTPCode: "123456", RefCode: "abc123"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.OTPConfirm(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- MockTokenHandler ---

func TestMockTokenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET_KEY", "test_secret")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(MockTokenRequest{VoterID: 1, AreaID: 1, Role: "voter"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/dev/mock-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	MockTokenHandler()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMockTokenHandler_DefaultRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET_KEY", "test_secret")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(MockTokenRequest{VoterID: 1, AreaID: 1, Role: ""})
	c.Request, _ = http.NewRequest(http.MethodPost, "/dev/mock-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	MockTokenHandler()(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMockTokenHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/dev/mock-token", bytes.NewBufferString("bad"))
	c.Request.Header.Set("Content-Type", "application/json")
	MockTokenHandler()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMockTokenHandler_ZeroVoterID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(MockTokenRequest{VoterID: 0, AreaID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/dev/mock-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	MockTokenHandler()(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMockTokenHandler_GenerateTokenError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET_KEY", "test_secret")

	old := tokenGenerator
	tokenGenerator = func(voterID, areaID uint, role, secret string) (string, error) {
		return "", errors.New("token gen failed")
	}
	t.Cleanup(func() { tokenGenerator = old })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(MockTokenRequest{VoterID: 1, AreaID: 1, Role: "voter"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/dev/mock-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	MockTokenHandler()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMockTokenHandler_MissingJWTKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET_KEY", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(MockTokenRequest{VoterID: 1, AreaID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/dev/mock-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	MockTokenHandler()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- VoterMeHandler ---

func TestVoterMeHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', 'test@example.com')")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("voter_id", uint(1))
	c.Request, _ = http.NewRequest(http.MethodGet, "/voter/me", nil)
	VoterMeHandler(db)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVoterMeHandler_NoVoterID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/voter/me", nil)
	VoterMeHandler(db)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVoterMeHandler_InvalidVoterIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("voter_id", "not-a-uint")
	c.Request, _ = http.NewRequest(http.MethodGet, "/voter/me", nil)
	VoterMeHandler(db)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestVoterMeHandler_VoterNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("voter_id", uint(999))
	c.Request, _ = http.NewRequest(http.MethodGet, "/voter/me", nil)
	VoterMeHandler(db)(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Legacy: OTPRequestHandler ---

func TestLegacyOTPRequestHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBufferString("bad"))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLegacyOTPRequestHandler_VoterNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 99})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLegacyOTPRequestHandler_NoEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', '')")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLegacyOTPRequestHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', 'test@example.com')")
	t.Setenv("ENABLE_DEV_ENDPOINTS", "true")
	t.Setenv("SMTP_HOST", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLegacyOTPRequestHandler_DBCreateFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', 'test@example.com')")
	db.Exec("DROP TABLE IF EXISTS otps") // OTP creation will fail
	t.Setenv("ENABLE_DEV_ENDPOINTS", "true")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLegacyOTPRequestHandler_EmailFail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', 'test@example.com')")
	t.Setenv("ENABLE_DEV_ENDPOINTS", "true")
	t.Setenv("SMTP_HOST", "smtp.test.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USER", "u@t.com")
	t.Setenv("SMTP_PASSWORD", "p")

	old := pkg.SMTPSendMail
	pkg.SMTPSendMail = func(addr string, a netsmtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("smtp failed")
	}
	t.Cleanup(func() { pkg.SMTPSendMail = old })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLegacyOTPRequestHandler_RealOTPGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', 'test@example.com')")
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "")
	t.Setenv("SMTP_HOST", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLegacyOTPRequestHandler_OTPRandError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', 'test@example.com')")
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "")

	old := cryptoRandInt
	cryptoRandInt = func(r io.Reader, max *big.Int) (*big.Int, error) {
		return nil, errors.New("forced rand error")
	}
	t.Cleanup(func() { cryptoRandInt = old })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLegacyOTPRequestHandler_RefRandError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number, email) VALUES (1, 'hash1', 1, '0812345678', 'test@example.com')")
	t.Setenv("ENABLE_DEV_ENDPOINTS", "false")
	t.Setenv("OTP_DELIVERY_MODE", "")

	oldInt := cryptoRandInt
	cryptoRandInt = func(r io.Reader, max *big.Int) (*big.Int, error) {
		return big.NewInt(111111), nil
	}
	oldRead := cryptoRandRead
	cryptoRandRead = func(r io.Reader, b []byte) (int, error) {
		return 0, errors.New("forced read error")
	}
	t.Cleanup(func() {
		cryptoRandInt = oldInt
		cryptoRandRead = oldRead
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPRequestRequest{VoterID: 1})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPRequestHandler(db)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Legacy: VerifyVoterHandler ---

func TestLegacyVerifyVoterHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/verify", bytes.NewBufferString("bad"))
	c.Request.Header.Set("Content-Type", "application/json")
	VerifyVoterHandler(db)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLegacyVerifyVoterHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	t.Setenv("HASH_SECRET_KEY", "test_key")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(VerifyVoterRequest{CitizenID: "1234567890123"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/verify", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	VerifyVoterHandler(db)(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLegacyVerifyVoterHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	t.Setenv("HASH_SECRET_KEY", "test_key")

	hash := generateCitizenIDHash("1234567890123")
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number) VALUES (1, ?, 1, '0812345678')", hash)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(VerifyVoterRequest{CitizenID: "1234567890123"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/verify", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	VerifyVoterHandler(db)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Legacy: OTPConfirmHandler ---

func TestLegacyOTPConfirmHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBufferString("bad"))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPConfirmHandler(db)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLegacyOTPConfirmHandler_EmptyOTPCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPConfirmRequest{OTPCode: "", RefCode: "abc"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPConfirmHandler(db)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLegacyOTPConfirmHandler_OTPNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	t.Setenv("JWT_SECRET_KEY", "test_secret")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPConfirmRequest{OTPCode: "123456", RefCode: "badref"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPConfirmHandler(db)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLegacyOTPConfirmHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerDB(t)
	t.Setenv("JWT_SECRET_KEY", "test_secret")

	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number) VALUES (1, 'hash1', 1, '0812345678')")
	exp := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO otps (otp_id, voter_id, otp_code, ref_code, expires_at, is_used, attempts) VALUES (1, 1, '123456', 'ref001', ?, 0, 0)", exp)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(OTPConfirmRequest{OTPCode: "123456", RefCode: "ref001"})
	c.Request, _ = http.NewRequest(http.MethodPost, "/voter/otp-confirm", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	OTPConfirmHandler(db)(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
