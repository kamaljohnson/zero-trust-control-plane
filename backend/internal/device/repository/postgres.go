package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/device/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns a device repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByID returns the device for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	d, err := r.queries.GetDevice(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genDeviceToDomain(&d), nil
}

// GetByUserOrgAndFingerprint returns the device for the given user, org, and fingerprint, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByUserOrgAndFingerprint(ctx context.Context, userID, orgID, fingerprint string) (*domain.Device, error) {
	d, err := r.queries.GetDeviceByUserAndFingerprint(ctx, gen.GetDeviceByUserAndFingerprintParams{UserID: userID, OrgID: orgID, Fingerprint: fingerprint})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genDeviceToDomain(&d), nil
}

// ListByOrg returns all devices for the given org. Returns (nil, error) only on database errors.
func (r *PostgresRepository) ListByOrg(ctx context.Context, orgID string) ([]*domain.Device, error) {
	list, err := r.queries.ListDevicesByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Device, len(list))
	for i := range list {
		out[i] = genDeviceToDomain(&list[i])
	}
	return out, nil
}

// Create persists the device to the database. The device must have ID set.
func (r *PostgresRepository) Create(ctx context.Context, d *domain.Device) error {
	lastSeen := sql.NullTime{}
	if d.LastSeenAt != nil {
		lastSeen = sql.NullTime{Time: *d.LastSeenAt, Valid: true}
	}
	trustedUntil := sql.NullTime{}
	if d.TrustedUntil != nil {
		trustedUntil = sql.NullTime{Time: *d.TrustedUntil, Valid: true}
	}
	revokedAt := sql.NullTime{}
	if d.RevokedAt != nil {
		revokedAt = sql.NullTime{Time: *d.RevokedAt, Valid: true}
	}
	_, err := r.queries.CreateDevice(ctx, gen.CreateDeviceParams{
		ID: d.ID, UserID: d.UserID, OrgID: d.OrgID, Fingerprint: d.Fingerprint,
		Trusted: d.Trusted, TrustedUntil: trustedUntil, RevokedAt: revokedAt,
		LastSeenAt: lastSeen, CreatedAt: d.CreatedAt,
	})
	return err
}

// UpdateTrusted sets the device's trusted flag for the given id. Returns an error if the update fails.
func (r *PostgresRepository) UpdateTrusted(ctx context.Context, id string, trusted bool) error {
	_, err := r.queries.UpdateDeviceTrusted(ctx, gen.UpdateDeviceTrustedParams{ID: id, Trusted: trusted})
	return err
}

// UpdateTrustedWithExpiry sets the device's trusted flag and trusted_until for the given id; clears revoked_at.
// Pass nil for trustedUntil to set no expiry.
func (r *PostgresRepository) UpdateTrustedWithExpiry(ctx context.Context, id string, trusted bool, trustedUntil *time.Time) error {
	tu := sql.NullTime{}
	if trustedUntil != nil {
		tu = sql.NullTime{Time: *trustedUntil, Valid: true}
	}
	_, err := r.queries.UpdateDeviceTrustedWithExpiry(ctx, gen.UpdateDeviceTrustedWithExpiryParams{
		ID: id, Trusted: trusted, TrustedUntil: tu,
	})
	return err
}

// Revoke sets revoked_at to now and clears trusted and trusted_until for the given device id.
func (r *PostgresRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := r.queries.RevokeDevice(ctx, gen.RevokeDeviceParams{ID: id, RevokedAt: sql.NullTime{Time: now, Valid: true}})
	return err
}

// UpdateLastSeen sets the device's last-seen timestamp for the given id. Returns an error if the update fails.
func (r *PostgresRepository) UpdateLastSeen(ctx context.Context, id string, at time.Time) error {
	_, err := r.queries.UpdateDeviceLastSeen(ctx, gen.UpdateDeviceLastSeenParams{ID: id, LastSeenAt: sql.NullTime{Time: at, Valid: true}})
	return err
}

func genDeviceToDomain(d *gen.Device) *domain.Device {
	if d == nil {
		return nil
	}
	var lastSeen, trustedUntil, revokedAt *time.Time
	if d.LastSeenAt.Valid {
		lastSeen = &d.LastSeenAt.Time
	}
	if d.TrustedUntil.Valid {
		trustedUntil = &d.TrustedUntil.Time
	}
	if d.RevokedAt.Valid {
		revokedAt = &d.RevokedAt.Time
	}
	return &domain.Device{
		ID: d.ID, UserID: d.UserID, OrgID: d.OrgID, Fingerprint: d.Fingerprint,
		Trusted: d.Trusted, TrustedUntil: trustedUntil, RevokedAt: revokedAt,
		LastSeenAt: lastSeen, CreatedAt: d.CreatedAt,
	}
}
