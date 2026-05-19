package pkg

import (
	"errors"
	"net/smtp"
	"sync/atomic"
	"testing"
	"time"
)

func TestEnqueueOTPEmail_NilQueue(t *testing.T) {
	old := emailQueue
	emailQueue = nil
	t.Cleanup(func() { emailQueue = old })

	t.Setenv("SMTP_HOST", "") // SendOTPEmail returns nil (dev skip)

	if err := EnqueueOTPEmail("x@x.com", "111111", "ref"); err != nil {
		t.Errorf("expected nil error from nil-queue fallback, got %v", err)
	}
}

func TestStartEmailWorker_ProcessesJob(t *testing.T) {
	old := emailQueue
	t.Cleanup(func() { emailQueue = old })

	var called atomic.Int32
	oldSend := SMTPSendMail
	SMTPSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		called.Add(1)
		return nil
	}
	t.Cleanup(func() { SMTPSendMail = oldSend })

	t.Setenv("SMTP_HOST", "smtp.test.com")
	t.Setenv("SMTP_USER", "u@t.com")
	t.Setenv("SMTP_PASSWORD", "p")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_FROM", "u@t.com")

	StartEmailWorker(1)

	if err := EnqueueOTPEmail("t@t.com", "222222", "ref"); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if called.Load() > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Error("email worker did not process job within 2s")
}

func TestStartEmailWorker_WorkerErrorLogging(t *testing.T) {
	old := emailQueue
	t.Cleanup(func() { emailQueue = old })

	oldSend := SMTPSendMail
	SMTPSendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("smtp failed")
	}
	t.Cleanup(func() { SMTPSendMail = oldSend })

	t.Setenv("SMTP_HOST", "smtp.test.com")
	t.Setenv("SMTP_USER", "u@t.com")
	t.Setenv("SMTP_PASSWORD", "p")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_FROM", "u@t.com")

	StartEmailWorker(1)

	// Enqueue a job — worker will call SMTPSendMail which returns error → logs
	_ = EnqueueOTPEmail("t@t.com", "999999", "errref")

	// Give the goroutine time to process and log the error
	time.Sleep(100 * time.Millisecond)
	// Test passes as long as it doesn't panic
}

func TestEnqueueOTPEmail_QueueFull(t *testing.T) {
	old := emailQueue
	emailQueue = make(chan emailJob, 0) // zero-capacity → always full
	t.Cleanup(func() { emailQueue = old })

	t.Setenv("SMTP_HOST", "") // fallback sync → dev skip → nil

	if err := EnqueueOTPEmail("x@x.com", "333333", "ref"); err != nil {
		t.Errorf("expected nil from sync fallback, got %v", err)
	}
}
