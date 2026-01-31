package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/policy/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns a policy repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByID returns the policy for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Policy, error) {
	p, err := r.queries.GetPolicy(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genPolicyToDomain(&p), nil
}

// ListByOrg returns all policies for the given org. Returns (nil, error) only on database errors.
func (r *PostgresRepository) ListByOrg(ctx context.Context, orgID string) ([]*domain.Policy, error) {
	list, err := r.queries.ListPoliciesByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Policy, len(list))
	for i := range list {
		out[i] = genPolicyToDomain(&list[i])
	}
	return out, nil
}

// Create persists the policy to the database. The policy must have ID set.
func (r *PostgresRepository) Create(ctx context.Context, p *domain.Policy) error {
	_, err := r.queries.CreatePolicy(ctx, gen.CreatePolicyParams{
		ID: p.ID, OrgID: p.OrgID, Rules: p.Rules, Enabled: p.Enabled, CreatedAt: p.CreatedAt,
	})
	return err
}

// Update updates the existing policy record in the database. Returns an error if the update fails.
func (r *PostgresRepository) Update(ctx context.Context, p *domain.Policy) error {
	_, err := r.queries.UpdatePolicy(ctx, gen.UpdatePolicyParams{
		ID: p.ID, Rules: p.Rules, Enabled: p.Enabled,
	})
	return err
}

func genPolicyToDomain(p *gen.Policy) *domain.Policy {
	if p == nil {
		return nil
	}
	return &domain.Policy{
		ID: p.ID, OrgID: p.OrgID, Rules: p.Rules, Enabled: p.Enabled, CreatedAt: p.CreatedAt,
	}
}
