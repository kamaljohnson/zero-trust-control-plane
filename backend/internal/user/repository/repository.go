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
}
