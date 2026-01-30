package repository

import (
	"context"
	"time"

	"zero-trust-control-plane/backend/internal/device/domain"
)

// Repository defines persistence for devices.
type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.Device, error)
	GetByUserOrgAndFingerprint(ctx context.Context, userID, orgID, fingerprint string) (*domain.Device, error)
	ListByOrg(ctx context.Context, orgID string) ([]*domain.Device, error)
	Save(ctx context.Context, d *domain.Device) error
	UpdateTrusted(ctx context.Context, id string, trusted bool) error
	UpdateLastSeen(ctx context.Context, id string, at time.Time) error
}
