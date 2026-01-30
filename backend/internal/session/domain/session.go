package domain

import "time"

// Session represents a user session tied to a device.
type Session struct {
	ID         string
	UserID     string
	OrgID      string
	DeviceID   string
	ExpiresAt  time.Time
	RevokedAt  *time.Time // nil when not revoked
	LastSeenAt *time.Time
	IPAddress  string
	CreatedAt  time.Time
}
