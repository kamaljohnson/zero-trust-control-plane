package domain

import (
	"time"
)

// Membership links a user to an organization with a role.
type Membership struct {
	ID        string
	UserID    string
	OrgID     string
	Role      Role
	CreatedAt time.Time
}

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)
