package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendSMS_MissingCredentials(t *testing.T) {
	t.Setenv("THAIBULKSMS_API_KEY", "")
	t.Setenv("THAIBULKSMS_API_SECRET", "")
	err := SendSMS("0812345678", "test message")
	if err == nil {
		t.Error("expected error for missing credentials")
	}
}

func TestSendSMS_DefaultSender(t *testing.T) {
	t.Setenv("THAIBULKSMS_API_KEY", "key")
	t.Setenv("THAIBULKSMS_API_SECRET", "secret")
	t.Setenv("THAIBULKSMS_SENDER", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	old := smsEndpoint
	smsEndpoint = srv.URL
	t.Cleanup(func() { smsEndpoint = old })

	if err := SendSMS("0812345678", "test"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendSMS_Non200Response(t *testing.T) {
	t.Setenv("THAIBULKSMS_API_KEY", "key")
	t.Setenv("THAIBULKSMS_API_SECRET", "secret")
	t.Setenv("THAIBULKSMS_SENDER", "TEST")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	old := smsEndpoint
	smsEndpoint = srv.URL
	t.Cleanup(func() { smsEndpoint = old })

	if err := SendSMS("0812345678", "test"); err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestSendSMS_HTTPError(t *testing.T) {
	t.Setenv("THAIBULKSMS_API_KEY", "key")
	t.Setenv("THAIBULKSMS_API_SECRET", "secret")
	t.Setenv("THAIBULKSMS_SENDER", "TEST")

	old := smsEndpoint
	smsEndpoint = "http://127.0.0.1:1" // nothing listening
	t.Cleanup(func() { smsEndpoint = old })

	if err := SendSMS("0812345678", "test"); err == nil {
		t.Error("expected error for connection failure")
	}
}

func TestSendSMS_InvalidURL(t *testing.T) {
	t.Setenv("THAIBULKSMS_API_KEY", "key")
	t.Setenv("THAIBULKSMS_API_SECRET", "secret")
	t.Setenv("THAIBULKSMS_SENDER", "TEST")

	old := smsEndpoint
	smsEndpoint = "://invalid-url" // malformed URL causes http.NewRequest to fail
	t.Cleanup(func() { smsEndpoint = old })

	if err := SendSMS("0812345678", "test"); err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestSendSMS_Success(t *testing.T) {
	t.Setenv("THAIBULKSMS_API_KEY", "key")
	t.Setenv("THAIBULKSMS_API_SECRET", "secret")
	t.Setenv("THAIBULKSMS_SENDER", "TEST")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	old := smsEndpoint
	smsEndpoint = srv.URL
	t.Cleanup(func() { smsEndpoint = old })

	if err := SendSMS("0812345678", "hello world"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
