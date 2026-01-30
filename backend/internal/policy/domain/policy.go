package domain

import "time"

// Policy represents an org-level policy.
type Policy struct {
	ID        string
	OrgID     string
	Rules     string
	Enabled   bool
	CreatedAt time.Time
}
