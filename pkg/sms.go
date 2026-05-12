package pkg

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SendSMS sends an SMS using ThaiBulkSMS
func SendSMS(to string, body string) error {
	apiKey := os.Getenv("THAIBULKSMS_API_KEY")
	apiSecret := os.Getenv("THAIBULKSMS_API_SECRET")
	sender := os.Getenv("THAIBULKSMS_SENDER")

	if apiKey == "" || apiSecret == "" {
		return fmt.Errorf("THAIBULKSMS credentials are not set in environment variables")
	}
	if sender == "" {
		sender = "SMS" // Default sender
	}

	apiUrl := "https://api-v2.thaibulksms.com/sms"

	// ThaiBulkSMS Payload
	data := url.Values{}
	data.Set("msisdn", to) // ThaiBulkSMS รองรับเบอร์ 08xxxxxxxx ได้โดยตรง
	data.Set("message", body)
	data.Set("sender", sender)

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(apiKey, apiSecret) // ส่ง Auth ผ่าน Header แบบ Basic Auth

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to ThaiBulkSMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ThaiBulkSMS error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
