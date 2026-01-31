package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/policy/domain"
)

// Repository defines persistence for policies.
type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.Policy, error)
	ListByOrg(ctx context.Context, orgID string) ([]*domain.Policy, error)
	Create(ctx context.Context, p *domain.Policy) error
	Update(ctx context.Context, p *domain.Policy) error
}
