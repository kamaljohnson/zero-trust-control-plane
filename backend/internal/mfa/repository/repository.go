package repository

import (
	"context"
	"time"

	"zero-trust-control-plane/backend/internal/mfa/domain"
)

// Repository defines persistence for MFA challenges.
type Repository interface {
	Create(ctx context.Context, c *domain.Challenge) error
	GetByID(ctx context.Context, id string) (*domain.Challenge, error)
	Delete(ctx context.Context, id string) error
}

// DefaultChallengeTTL is the default MFA challenge expiry (e.g. 10 minutes).
const DefaultChallengeTTL = 10 * time.Minute
