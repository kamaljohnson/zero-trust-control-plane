package domain

import "time"

// Device represents a registered device for a user in an org.
type Device struct {
	ID          string
	UserID      string
	OrgID       string
	Fingerprint string
	Trusted     bool
	LastSeenAt  *time.Time
	CreatedAt   time.Time
}
