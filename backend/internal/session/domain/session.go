package domain

import "time"

// Session represents a user session tied to a device.
type Session struct {
	ID                string
	UserID            string
	OrgID             string
	DeviceID          string
	ExpiresAt         time.Time
	RevokedAt         *time.Time // nil when not revoked
	LastSeenAt        *time.Time
	IPAddress         string
	RefreshJti        string // current refresh token jti for rotation; empty if not set
	RefreshTokenHash  string // SHA-256 hash of current refresh token; empty for legacy sessions
	CreatedAt         time.Time
}
