package domain

import "time"

// Device represents a registered device for a user in an org.
// Effective trust is Trusted && (TrustedUntil == nil || TrustedUntil.After(now)) && RevokedAt == nil.
type Device struct {
	ID           string
	UserID       string
	OrgID        string
	Fingerprint  string
	Trusted      bool
	TrustedUntil *time.Time
	RevokedAt    *time.Time
	LastSeenAt   *time.Time
	CreatedAt    time.Time
}

// IsEffectivelyTrusted returns true if the device is trusted, not revoked, and trust has not expired.
func (d *Device) IsEffectivelyTrusted(now time.Time) bool {
	if !d.Trusted || d.RevokedAt != nil {
		return false
	}
	if d.TrustedUntil != nil && !d.TrustedUntil.After(now) {
		return false
	}
	return true
}
