package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/orgmfasettings/domain"
)

// Repository defines read access to org MFA/device trust settings.
type Repository interface {
	// GetByOrgID returns org MFA settings for the given org, or nil if not found (caller uses defaults).
	GetByOrgID(ctx context.Context, orgID string) (*domain.OrgMFASettings, error)
}
