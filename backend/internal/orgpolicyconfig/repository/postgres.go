package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/orgpolicyconfig/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns an org policy config repository that uses the given db.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByOrgID returns the config for the org, or nil if not found.
func (r *PostgresRepository) GetByOrgID(ctx context.Context, orgID string) (*domain.OrgPolicyConfig, error) {
	row, err := r.queries.GetOrgPolicyConfig(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var config domain.OrgPolicyConfig
	if err := json.Unmarshal([]byte(row.ConfigJson), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Upsert saves or replaces the config for the org.
func (r *PostgresRepository) Upsert(ctx context.Context, orgID string, config *domain.OrgPolicyConfig) error {
	if config == nil {
		config = &domain.OrgPolicyConfig{}
	}
	merged := domain.MergeWithDefaults(config)
	raw, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	_, err = r.queries.UpsertOrgPolicyConfig(ctx, gen.UpsertOrgPolicyConfigParams{
		OrgID:      orgID,
		ConfigJson: string(raw),
		UpdatedAt:  time.Now().UTC(),
	})
	return err
}
