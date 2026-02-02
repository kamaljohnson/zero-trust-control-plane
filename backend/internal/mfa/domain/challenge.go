package domain

import "time"

// Challenge represents an MFA OTP challenge (stored in mfa_challenges table).
type Challenge struct {
	ID        string
	UserID    string
	OrgID     string
	DeviceID  string
	Phone     string
	CodeHash  string
	ExpiresAt time.Time
	CreatedAt time.Time
}
