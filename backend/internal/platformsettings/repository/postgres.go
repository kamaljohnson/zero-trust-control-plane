package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/platformsettings/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns a platform settings repository that uses the given db.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetDeviceTrustSettings returns platform-level MFA/device trust settings from DB, or defaults.
func (r *PostgresRepository) GetDeviceTrustSettings(ctx context.Context, defaultTrustTTLDays int) (*domain.PlatformDeviceTrustSettings, error) {
	out := &domain.PlatformDeviceTrustSettings{
		MFARequiredAlways:   false,
		DefaultTrustTTLDays: defaultTrustTTLDays,
	}
	mfa, err := r.queries.GetPlatformSetting(ctx, "mfa_required_always")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, nil
		}
		return nil, err
	}
	if v, err := parseBool(mfa.ValueJson); err == nil {
		out.MFARequiredAlways = v
	}
	ttl, err := r.queries.GetPlatformSetting(ctx, "default_trust_ttl_days")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, nil
		}
		return nil, err
	}
	if v, err := strconv.Atoi(ttl.ValueJson); err == nil && v > 0 {
		out.DefaultTrustTTLDays = v
	}
	return out, nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1":
		return true, nil
	case "false", "0", "":
		return false, nil
	default:
		return false, strconv.ErrSyntax
	}
}
