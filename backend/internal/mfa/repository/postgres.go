package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/mfa/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns an MFA challenge repository that uses the given db.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// Create persists the MFA challenge. The challenge must have ID set.
func (r *PostgresRepository) Create(ctx context.Context, c *domain.Challenge) error {
	_, err := r.queries.CreateMFAChallenge(ctx, gen.CreateMFAChallengeParams{
		ID: c.ID, UserID: c.UserID, OrgID: c.OrgID, DeviceID: c.DeviceID,
		Phone: c.Phone, CodeHash: c.CodeHash, ExpiresAt: c.ExpiresAt, CreatedAt: c.CreatedAt,
	})
	return err
}

// GetByID returns the MFA challenge for id, or nil if not found.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Challenge, error) {
	row, err := r.queries.GetMFAChallenge(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Challenge{
		ID: row.ID, UserID: row.UserID, OrgID: row.OrgID, DeviceID: row.DeviceID,
		Phone: row.Phone, CodeHash: row.CodeHash, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
	}, nil
}

// Delete removes the MFA challenge by id.
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteMFAChallenge(ctx, id)
}
