package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/telemetry/domain"
)

// Repository defines persistence for telemetry events.
type Repository interface {
	Save(ctx context.Context, t *domain.Telemetry) error
	GetByID(ctx context.Context, id int64) (*domain.Telemetry, error)
	ListByOrg(ctx context.Context, orgID string, limit, offset int32) ([]*domain.Telemetry, error)
}
