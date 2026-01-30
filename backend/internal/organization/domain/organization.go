package domain

import (
	"errors"
	"time"
)

// Org represents an organization/tenant.
type Org struct {
	ID        string
	Name      string
	Status    OrgStatus
	CreatedAt time.Time
}

type OrgStatus string

const (
	OrgStatusActive    OrgStatus = "active"
	OrgStatusSuspended OrgStatus = "suspended"
)

// Validate validates the organization for persistence. Returns an error describing the first validation failure.
func (o *Org) Validate() error {
	if o.Name == "" {
		return errors.New("name is required")
	}
	if o.Status == "" {
		o.Status = OrgStatusActive
	}
	return nil
}
