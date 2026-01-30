package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/organization/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns an organization repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetOrganizationByID returns the organization for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetOrganizationByID(ctx context.Context, id string) (*domain.Org, error) {
	o, err := r.queries.GetOrganization(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genOrgToDomain(&o), nil
}

// CreateOrganization persists the organization to the database. The organization must have ID set.
func (r *PostgresRepository) CreateOrganization(ctx context.Context, o *domain.Org) error {
	_, err := r.queries.CreateOrganization(ctx, gen.CreateOrganizationParams{
		ID: o.ID, Name: o.Name, Status: gen.OrgStatus(o.Status), CreatedAt: o.CreatedAt,
	})
	return err
}

// UpdateOrganization updates the existing organization record in the database. Returns an error if the update fails.
func (r *PostgresRepository) UpdateOrganization(ctx context.Context, o *domain.Org) error {
	_, err := r.queries.UpdateOrganization(ctx, gen.UpdateOrganizationParams{
		ID: o.ID, Name: o.Name, Status: gen.OrgStatus(o.Status),
	})
	return err
}

func genOrgToDomain(o *gen.Organization) *domain.Org {
	if o == nil {
		return nil
	}
	return &domain.Org{
		ID: o.ID, Name: o.Name,
		Status: domain.OrgStatus(o.Status), CreatedAt: o.CreatedAt,
	}
}
