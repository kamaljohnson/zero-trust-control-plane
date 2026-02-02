package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/mfaintent/domain"
)

// Repository defines persistence for MFA intents (one-time phone-collect binding).
type Repository interface {
	Create(ctx context.Context, i *domain.Intent) error
	GetByID(ctx context.Context, id string) (*domain.Intent, error)
	Delete(ctx context.Context, id string) error
}
