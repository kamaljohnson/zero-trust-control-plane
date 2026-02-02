package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/audit/domain"
)

// Repository defines persistence for audit logs.
type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.AuditLog, error)
	ListByOrg(ctx context.Context, orgID string, limit, offset int32) ([]*domain.AuditLog, error)
	// ListByOrgFiltered returns audit logs for the org with optional filters; nil filter means no filter.
	ListByOrgFiltered(ctx context.Context, orgID string, limit, offset int32, userID, action, resource *string) ([]*domain.AuditLog, error)
	Create(ctx context.Context, a *domain.AuditLog) error
}
