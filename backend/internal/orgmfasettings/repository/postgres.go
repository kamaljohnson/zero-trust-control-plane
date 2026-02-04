package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

// Upsert creates or updates org MFA settings for the given org.
func (r *PostgresRepository) Upsert(ctx context.Context, settings *domain.OrgMFASettings) error {
	now := settings.UpdatedAt
	if now.IsZero() {
		now = settings.CreatedAt
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	created := settings.CreatedAt
	if created.IsZero() {
		created = now
	}
	_, err := r.queries.UpsertOrgMFASettings(ctx, gen.UpsertOrgMFASettingsParams{
		OrgID:                   settings.OrgID,
		MfaRequiredForNewDevice: settings.MFARequiredForNewDevice,
		MfaRequiredForUntrusted: settings.MFARequiredForUntrusted,
		MfaRequiredAlways:       settings.MFARequiredAlways,
		RegisterTrustAfterMfa:   settings.RegisterTrustAfterMFA,
		TrustTtlDays:            int32(settings.TrustTTLDays),
		CreatedAt:               created,
		UpdatedAt:               now,
	})
	return err
}
