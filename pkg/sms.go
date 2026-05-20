package pkg

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	httpClient  = &http.Client{}
	smsEndpoint = "https://api-v2.thaibulksms.com/sms"
)

// SendSMS sends an SMS via ThaiBulkSMS API
func SendSMS(to string, body string) error {
	apiKey := os.Getenv("THAIBULKSMS_API_KEY")
	apiSecret := os.Getenv("THAIBULKSMS_API_SECRET")
	sender := os.Getenv("THAIBULKSMS_SENDER")

	if apiKey == "" || apiSecret == "" {
		return fmt.Errorf("THAIBULKSMS_API_KEY และ THAIBULKSMS_API_SECRET ต้องตั้งค่าก่อนใช้งาน")
	}
	if sender == "" {
		sender = "SMS"
	}

	data := url.Values{}
	data.Set("msisdn", to)
	data.Set("message", body)
	data.Set("sender", sender)

	req, err := http.NewRequest("POST", smsEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create SMS request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(apiKey, apiSecret)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ส่ง SMS ไม่สำเร็จ: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ThaiBulkSMS error (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}
