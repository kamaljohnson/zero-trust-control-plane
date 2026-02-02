package domain

import "time"

// Intent represents a one-time binding for "collect phone then send OTP" when the user has no phone.
// Consumed (deleted) when SubmitPhoneAndRequestMFA is called.
type Intent struct {
	ID        string
	UserID    string
	OrgID     string
	DeviceID  string
	ExpiresAt time.Time
}
