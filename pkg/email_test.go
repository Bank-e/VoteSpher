package pkg

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"testing"
)

func TestBuildOTPEmailHTML(t *testing.T) {
	html := buildOTPEmailHTML("123456", "abc123")
	if !strings.Contains(html, "123456") {
		t.Error("HTML must contain OTP code")
	}
	if !strings.Contains(html, "abc123") {
		t.Error("HTML must contain ref code")
	}
	if !strings.Contains(html, "VoteSpher") {
		t.Error("HTML must contain VoteSpher branding")
	}
}

func TestSendOTPEmail_SkipWhenNoHost(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	if err := SendOTPEmail("test@example.com", "123456", "ref001"); err != nil {
		t.Errorf("expected nil for missing host, got %v", err)
	}
}

func TestSendOTPEmail_MissingCreds(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.test.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USER", "")
	t.Setenv("SMTP_PASSWORD", "")
	err := SendOTPEmail("test@example.com", "123456", "ref001")
	if err == nil {
		t.Error("expected error for missing credentials")
	}
}

func TestSendOTPEmail_DefaultPortAndFrom(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.test.com")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USER", "user@test.com")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM", "")

	var gotAddr string
	SMTPSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotAddr = addr
		if from != "user@test.com" {
			return fmt.Errorf("expected from=user@test.com, got %s", from)
		}
		return nil
	}
	t.Cleanup(func() { SMTPSendMail = smtp.SendMail })

	if err := SendOTPEmail("to@test.com", "654321", "ref002"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(gotAddr, ":587") {
		t.Errorf("expected default port 587 in addr, got %s", gotAddr)
	}
}

func TestSendOTPEmail_SMTPError(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.test.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USER", "user@test.com")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM", "from@test.com")

	SMTPSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("connection refused")
	}
	t.Cleanup(func() { SMTPSendMail = smtp.SendMail })

	err := SendOTPEmail("to@test.com", "654321", "ref003")
	if err == nil {
		t.Error("expected error from SMTP failure")
	}
}

func TestSendOTPEmail_Success(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.test.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USER", "user@test.com")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM", "from@test.com")

	SMTPSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}
	t.Cleanup(func() { SMTPSendMail = smtp.SendMail })

	if err := SendOTPEmail("to@test.com", "111111", "ref004"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
