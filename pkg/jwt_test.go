package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-for-unit-tests"

func TestGenerateToken_Valid(t *testing.T) {
	token, err := GenerateToken(1, 2, "voter", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Error("token must not be empty")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	token, _ := GenerateToken(42, 3, "admin", testSecret)
	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.VoterID != 42 {
		t.Errorf("expected voter_id=42, got %d", claims.VoterID)
	}
	if claims.AreaID != 3 {
		t.Errorf("expected area_id=3, got %d", claims.AreaID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role=admin, got %s", claims.Role)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken(1, 1, "voter", testSecret)
	_, err := ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error with wrong secret, got nil")
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	_, err := ValidateToken("not.a.valid.token", testSecret)
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestValidateToken_Empty(t *testing.T) {
	_, err := ValidateToken("", testSecret)
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	// สร้าง token ที่หมดอายุแล้วโดยตรง
	claims := &JWTCustomClaims{
		VoterID: 1,
		AreaID:  1,
		Role:    "voter",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "VoteSpher",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := tok.SignedString([]byte(testSecret))

	_, err := ValidateToken(tokenStr, testSecret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestGenerateToken_RolesRoundtrip(t *testing.T) {
	for _, role := range []string{"voter", "admin"} {
		tok, err := GenerateToken(99, 5, role, testSecret)
		if err != nil {
			t.Fatalf("role=%s generate failed: %v", role, err)
		}
		c, err := ValidateToken(tok, testSecret)
		if err != nil {
			t.Fatalf("role=%s validate failed: %v", role, err)
		}
		if c.Role != role || c.VoterID != 99 || c.AreaID != 5 {
			t.Errorf("role=%s roundtrip mismatch: %+v", role, c)
		}
		if c.ExpiresAt == nil || c.ExpiresAt.Before(time.Now()) {
			t.Errorf("role=%s token should not be expired", role)
		}
	}
}

func TestValidateToken_TamperedPayload(t *testing.T) {
	// token ถูก sign ด้วย secret อื่น แต่ใช้ secret เดิม validate — ต้อง error
	tok1, _ := GenerateToken(1, 1, "voter", "secret-A")
	_, err := ValidateToken(tok1, "secret-B")
	if err == nil {
		t.Error("tampered token should not validate")
	}
}

func TestGenerateToken_CustomExpiry(t *testing.T) {
	t.Setenv("JWT_EXPIRY_HOURS", "3")
	tok, err := GenerateToken(1, 1, "voter", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerateToken_ZeroExpiry(t *testing.T) {
	// JWT_EXPIRY_HOURS=0 → n <= 0 → uses default 2h
	t.Setenv("JWT_EXPIRY_HOURS", "0")
	tok, err := GenerateToken(1, 1, "voter", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerateToken_InvalidExpiry(t *testing.T) {
	// JWT_EXPIRY_HOURS=abc → parse error → uses default 2h
	t.Setenv("JWT_EXPIRY_HOURS", "abc")
	tok, err := GenerateToken(1, 1, "voter", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok == "" {
		t.Error("expected non-empty token")
	}
}

func TestValidateToken_UnexpectedAlgorithm(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	claims := &JWTCustomClaims{
		VoterID: 1,
		AreaID:  1,
		Role:    "voter",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenStr, _ := tok.SignedString(rsaKey)

	_, err = ValidateToken(tokenStr, testSecret)
	if err == nil {
		t.Error("expected error for unexpected signing algorithm")
	}
}
