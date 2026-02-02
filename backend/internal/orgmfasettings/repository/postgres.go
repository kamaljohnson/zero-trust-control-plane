package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/orgmfasettings/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns an org MFA settings repository that uses the given db.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByOrgID returns org MFA settings for the given org, or nil if not found.
func (r *PostgresRepository) GetByOrgID(ctx context.Context, orgID string) (*domain.OrgMFASettings, error) {
	row, err := r.queries.GetOrgMFASettings(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.OrgMFASettings{
		OrgID:                   row.OrgID,
		MFARequiredForNewDevice: row.MfaRequiredForNewDevice,
		MFARequiredForUntrusted: row.MfaRequiredForUntrusted,
		MFARequiredAlways:       row.MfaRequiredAlways,
		RegisterTrustAfterMFA:   row.RegisterTrustAfterMfa,
		TrustTTLDays:            int(row.TrustTtlDays),
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}, nil
}
