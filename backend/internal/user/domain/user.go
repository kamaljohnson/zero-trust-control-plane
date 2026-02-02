package domain

import (
	"errors"
	"time"
)

// User is the core user entity.
type User struct {
	ID            string
	Email         string
	Name          string
	Phone         string // optional; used for MFA (PoC); immutable after PhoneVerified
	PhoneVerified bool   // true after first successful MFA verification
	Status        UserStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

// Validate validates the user for persistence. Returns an error describing the first validation failure.
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	return nil
}
