package mfa

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

const otpDigits = 6

// GenerateOTP returns a 6-digit numeric OTP string (e.g. "123456").
// Uses crypto/rand for randomness.
func GenerateOTP() (string, error) {
	b := make([]byte, otpDigits)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := make([]byte, otpDigits)
	for i := 0; i < otpDigits; i++ {
		s[i] = '0' + (b[i] % 10)
	}
	return string(s), nil
}

// HashOTP returns a SHA-256 hash of the OTP string, hex-encoded.
func HashOTP(otp string) string {
	h := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(h[:])
}

// OTPEqual performs constant-time comparison of the provided OTP's hash with the stored hash.
func OTPEqual(providedOTP, storedHash string) bool {
	providedHash := HashOTP(providedOTP)
	return subtle.ConstantTimeCompare([]byte(providedHash), []byte(storedHash)) == 1
}
