package repository

import (
	"context"
	"time"

	"zero-trust-control-plane/backend/internal/session/domain"
)

// Repository defines persistence for sessions.
type Repository interface {
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	ListByUserAndOrg(ctx context.Context, userID, orgID string) ([]*domain.Session, error)
	Save(ctx context.Context, s *domain.Session) error
	Revoke(ctx context.Context, id string) error
	UpdateLastSeen(ctx context.Context, id string, at time.Time) error
}
