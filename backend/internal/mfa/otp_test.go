package mfa

import (
	"testing"
)

func TestGenerateOTP_ReturnsSixDigits(t *testing.T) {
	otp, err := GenerateOTP()
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}
	if len(otp) != 6 {
		t.Errorf("OTP length = %d, want 6", len(otp))
	}
	for _, c := range otp {
		if c < '0' || c > '9' {
			t.Errorf("OTP contains non-digit: %c", c)
		}
	}
}

func TestGenerateOTP_Randomness(t *testing.T) {
	// Generate multiple OTPs and verify they're different (very unlikely to be same)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		otp, err := GenerateOTP()
		if err != nil {
			t.Fatalf("GenerateOTP: %v", err)
		}
		if seen[otp] {
			t.Errorf("duplicate OTP generated: %s", otp)
		}
		seen[otp] = true
	}
}

func TestHashOTP_Consistent(t *testing.T) {
	otp := "123456"
	hash1 := HashOTP(otp)
	hash2 := HashOTP(otp)

	if hash1 != hash2 {
		t.Errorf("HashOTP not consistent: hash1 = %q, hash2 = %q", hash1, hash2)
	}
	if len(hash1) != 64 {
		t.Errorf("hash length = %d, want 64 (SHA-256 hex)", len(hash1))
	}
}

func TestHashOTP_DifferentInputs(t *testing.T) {
	hash1 := HashOTP("123456")
	hash2 := HashOTP("654321")

	if hash1 == hash2 {
		t.Error("HashOTP produced same hash for different inputs")
	}
}

func TestOTPEqual_CorrectMatch(t *testing.T) {
	otp := "123456"
	storedHash := HashOTP(otp)

	if !OTPEqual(otp, storedHash) {
		t.Error("OTPEqual should match correct OTP")
	}
}

func TestOTPEqual_RejectsIncorrect(t *testing.T) {
	correctOTP := "123456"
	wrongOTP := "654321"
	storedHash := HashOTP(correctOTP)

	if OTPEqual(wrongOTP, storedHash) {
		t.Error("OTPEqual should reject incorrect OTP")
	}
}

func TestOTPEqual_ConstantTime(t *testing.T) {
	// This is a basic test - full timing attack resistance requires more sophisticated testing
	otp := "123456"
	storedHash := HashOTP(otp)

	// Test that it works correctly
	if !OTPEqual(otp, storedHash) {
		t.Error("OTPEqual should match correct OTP")
	}

	// Test with different length inputs to ensure constant-time comparison
	wrongHash := "a" + storedHash
	if OTPEqual(otp, wrongHash) {
		t.Error("OTPEqual should reject hash with different length")
	}
}

func TestOTPEqual_EmptyInputs(t *testing.T) {
	if OTPEqual("", "") {
		t.Error("OTPEqual should not match empty inputs")
	}

	hash := HashOTP("123456")
	if OTPEqual("", hash) {
		t.Error("OTPEqual should not match empty OTP")
	}
}
