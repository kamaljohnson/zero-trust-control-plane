package domain

import "time"

// Telemetry represents a telemetry event (org-scoped, optional user/device/session).
type Telemetry struct {
	ID        int64
	OrgID     string
	UserID    *string // nil if not set
	DeviceID  *string
	SessionID *string
	EventType string
	Source    string
	Metadata  []byte // JSONB
	CreatedAt time.Time
}
