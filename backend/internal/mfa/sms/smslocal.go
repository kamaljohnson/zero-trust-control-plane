package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultTimeout = 15 * time.Second

// SMSLocalClient sends OTP SMS via SMS Local API (PoC).
// See https://www.smslocal.in/help/otp-sms/ and https://www.smslocal.com/dev/bulkV2.
type SMSLocalClient struct {
	APIKey     string
	BaseURL    string
	Sender     string
	HTTPClient *http.Client
}

// NewSMSLocalClient returns a client that uses the given API key and optional base URL/sender.
func NewSMSLocalClient(apiKey, baseURL, sender string) *SMSLocalClient {
	if baseURL == "" {
		baseURL = "https://www.smslocal.com/dev/bulkV2"
	}
	return &SMSLocalClient{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Sender:     sender,
		HTTPClient: &http.Client{Timeout: defaultTimeout},
	}
}

// SendOTP sends the OTP to the given phone number via SMS Local (route=otp).
// phone should be digits only (e.g. country code + number). Does not log the OTP.
func (c *SMSLocalClient) SendOTP(phone, otp string) error {
	if c.APIKey == "" {
		return fmt.Errorf("sms: API key not configured")
	}
	body := map[string]interface{}{
		"route":     "otp",
		"numbers":   phone,
		"variables": otp,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sms: request failed status=%d body=%s", resp.StatusCode, string(b))
	}
	return nil
}
