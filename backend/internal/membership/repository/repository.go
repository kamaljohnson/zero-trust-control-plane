package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/membership/domain"
)

// Repository defines persistence for memberships.
type Repository interface {
	GetMembershipByID(ctx context.Context, id string) (*domain.Membership, error)
	GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*domain.Membership, error)
	ListMembershipsByOrg(ctx context.Context, orgID string) ([]*domain.Membership, error)
	CreateMembership(ctx context.Context, m *domain.Membership) error
}
