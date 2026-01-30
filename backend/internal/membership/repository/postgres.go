package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/membership/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns a membership repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetMembershipByID returns the membership for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetMembershipByID(ctx context.Context, id string) (*domain.Membership, error) {
	m, err := r.queries.GetMembership(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genMembershipToDomain(&m), nil
}

// GetMembershipByUserAndOrg returns the membership for the given user and org, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*domain.Membership, error) {
	m, err := r.queries.GetMembershipByUserAndOrg(ctx, gen.GetMembershipByUserAndOrgParams{UserID: userID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genMembershipToDomain(&m), nil
}

// ListMembershipsByOrg returns all memberships for the given org. Returns (nil, error) only on database errors.
func (r *PostgresRepository) ListMembershipsByOrg(ctx context.Context, orgID string) ([]*domain.Membership, error) {
	list, err := r.queries.ListMembershipsByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Membership, len(list))
	for i := range list {
		out[i] = genMembershipToDomain(&list[i])
	}
	return out, nil
}

// CreateMembership persists the membership to the database. The membership must have ID set.
func (r *PostgresRepository) CreateMembership(ctx context.Context, m *domain.Membership) error {
	_, err := r.queries.CreateMembership(ctx, gen.CreateMembershipParams{
		ID: m.ID, UserID: m.UserID, OrgID: m.OrgID, Role: gen.Role(m.Role), CreatedAt: m.CreatedAt,
	})
	return err
}

func genMembershipToDomain(m *gen.Membership) *domain.Membership {
	if m == nil {
		return nil
	}
	return &domain.Membership{
		ID: m.ID, UserID: m.UserID, OrgID: m.OrgID, Role: domain.Role(m.Role), CreatedAt: m.CreatedAt,
	}
}
