package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/orgpolicyconfig/domain"
)

// Repository persists org policy config.
type Repository interface {
	// GetByOrgID returns the config for the org, or nil if not found (caller applies defaults).
	GetByOrgID(ctx context.Context, orgID string) (*domain.OrgPolicyConfig, error)
	// Upsert saves or replaces the config for the org.
	Upsert(ctx context.Context, orgID string, config *domain.OrgPolicyConfig) error
}
