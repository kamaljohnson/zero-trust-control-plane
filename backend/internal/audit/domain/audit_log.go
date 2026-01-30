package domain

import "time"

// AuditLog represents an audit event.
type AuditLog struct {
	ID        string
	OrgID     string
	UserID    string
	Action    string
	Resource  string
	IP        string
	Metadata  string
	CreatedAt time.Time
}
