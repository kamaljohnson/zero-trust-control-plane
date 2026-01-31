package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/audit/domain"
)

// Repository defines persistence for audit logs.
type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.AuditLog, error)
	ListByOrg(ctx context.Context, orgID string, limit, offset int32) ([]*domain.AuditLog, error)
	Create(ctx context.Context, a *domain.AuditLog) error
}
