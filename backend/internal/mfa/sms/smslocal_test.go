package sms

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewSMSLocalClient_Defaults(t *testing.T) {
	client := NewSMSLocalClient("api-key", "", "")
	if client.APIKey != "api-key" {
		t.Errorf("APIKey = %q, want %q", client.APIKey, "api-key")
	}
	if client.BaseURL != "https://www.smslocal.com/dev/bulkV2" {
		t.Errorf("BaseURL = %q, want default", client.BaseURL)
	}
	if client.Sender != "" {
		t.Errorf("Sender = %q, want empty", client.Sender)
	}
	if client.HTTPClient == nil {
		t.Fatal("HTTPClient should be set")
	}
	if client.HTTPClient.Timeout != defaultTimeout {
		t.Errorf("HTTPClient.Timeout = %v, want %v", client.HTTPClient.Timeout, defaultTimeout)
	}
}

func TestNewSMSLocalClient_CustomBaseURL(t *testing.T) {
	customURL := "https://custom.sms.local/api"
	client := NewSMSLocalClient("api-key", customURL, "")
	if client.BaseURL != customURL {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, customURL)
	}
}

func TestNewSMSLocalClient_CustomSender(t *testing.T) {
	sender := "TEST"
	client := NewSMSLocalClient("api-key", "", sender)
	if client.Sender != sender {
		t.Errorf("Sender = %q, want %q", client.Sender, sender)
	}
}

func TestSendOTP_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "test-api-key" {
			t.Errorf("Authorization = %q, want test-api-key", r.Header.Get("Authorization"))
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode body: %v", err)
		}
		if body["route"] != "otp" {
			t.Errorf("route = %v, want otp", body["route"])
		}
		if body["numbers"] != "1234567890" {
			t.Errorf("numbers = %v, want 1234567890", body["numbers"])
		}
		if body["variables"] != "123456" {
			t.Errorf("variables = %v, want 123456", body["variables"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := NewSMSLocalClient("test-api-key", server.URL, "")
	err := client.SendOTP("1234567890", "123456")
	if err != nil {
		t.Fatalf("SendOTP: %v", err)
	}
}

func TestSendOTP_MissingAPIKey(t *testing.T) {
	client := NewSMSLocalClient("", "", "")
	err := client.SendOTP("1234567890", "123456")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "API key not configured") {
		t.Errorf("error message = %q, want to contain 'API key not configured'", err.Error())
	}
}

func TestSendOTP_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate connection error by closing connection
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	client := NewSMSLocalClient("api-key", server.URL, "")
	// Use a very short timeout to trigger error faster
	client.HTTPClient = &http.Client{Timeout: 1 * time.Millisecond}

	err := client.SendOTP("1234567890", "123456")
	if err == nil {
		t.Fatal("expected error for HTTP failure")
	}
}

func TestSendOTP_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid request"}`))
	}))
	defer server.Close()

	client := NewSMSLocalClient("api-key", server.URL, "")
	err := client.SendOTP("1234567890", "123456")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "status=400") {
		t.Errorf("error message = %q, want to contain 'status=400'", err.Error())
	}
	if !strings.Contains(err.Error(), "invalid request") {
		t.Errorf("error message = %q, want to contain response body", err.Error())
	}
}

func TestSendOTP_500Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	client := NewSMSLocalClient("api-key", server.URL, "")
	err := client.SendOTP("1234567890", "123456")
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "status=500") {
		t.Errorf("error message = %q, want to contain 'status=500'", err.Error())
	}
}

func TestSendOTP_RequestFormat(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := NewSMSLocalClient("api-key", server.URL, "")
	err := client.SendOTP("9876543210", "654321")
	if err != nil {
		t.Fatalf("SendOTP: %v", err)
	}

	if receivedBody == nil {
		t.Fatal("request body was not received")
	}
	if receivedBody["route"] != "otp" {
		t.Errorf("route = %v, want otp", receivedBody["route"])
	}
	if receivedBody["numbers"] != "9876543210" {
		t.Errorf("numbers = %v, want 9876543210", receivedBody["numbers"])
	}
	if receivedBody["variables"] != "654321" {
		t.Errorf("variables = %v, want 654321", receivedBody["variables"])
	}
}
