package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/mfaintent/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns an MFA intent repository that uses the given db.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// Create persists the MFA intent. The intent must have ID set.
func (r *PostgresRepository) Create(ctx context.Context, i *domain.Intent) error {
	_, err := r.queries.CreateMFAIntent(ctx, gen.CreateMFAIntentParams{
		ID:        i.ID,
		UserID:    i.UserID,
		OrgID:     i.OrgID,
		DeviceID:  i.DeviceID,
		ExpiresAt: i.ExpiresAt,
	})
	return err
}

// GetByID returns the MFA intent for id, or nil if not found.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Intent, error) {
	row, err := r.queries.GetMFAIntent(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Intent{
		ID:        row.ID,
		UserID:    row.UserID,
		OrgID:     row.OrgID,
		DeviceID:  row.DeviceID,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

// Delete removes the MFA intent by id.
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteMFAIntent(ctx, id)
}
