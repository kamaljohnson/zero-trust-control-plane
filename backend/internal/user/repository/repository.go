package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/user/domain"
)

// Repository defines persistence for users.
type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	// SetPhoneVerified sets phone and phone_verified only when user has no phone and not yet verified. No-op if already set.
	SetPhoneVerified(ctx context.Context, userID, phone string) error
}
