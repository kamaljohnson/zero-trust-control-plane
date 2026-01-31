package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/identity/domain"
)

// Repository defines persistence for identities.
type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.Identity, error)
	GetByUserAndProvider(ctx context.Context, userID string, provider domain.IdentityProvider) (*domain.Identity, error)
	GetByUserAndProviderID(ctx context.Context, userID string, provider domain.IdentityProvider, providerID string) (*domain.Identity, error)
	Create(ctx context.Context, i *domain.Identity) error
	UpdatePasswordHash(ctx context.Context, id string, passwordHash string) error
}
