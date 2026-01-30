package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/session/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns a session repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByID returns the session for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	s, err := r.queries.GetSession(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genSessionToDomain(&s), nil
}

// ListByUserAndOrg returns all sessions for the given user and org. Returns (nil, error) only on database errors.
func (r *PostgresRepository) ListByUserAndOrg(ctx context.Context, userID, orgID string) ([]*domain.Session, error) {
	list, err := r.queries.ListSessionsByUserAndOrg(ctx, gen.ListSessionsByUserAndOrgParams{UserID: userID, OrgID: orgID})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Session, len(list))
	for i := range list {
		out[i] = genSessionToDomain(&list[i])
	}
	return out, nil
}

// Save persists the session to the database. The session must have ID set.
func (r *PostgresRepository) Save(ctx context.Context, s *domain.Session) error {
	_, err := r.queries.CreateSession(ctx, gen.CreateSessionParams{
		ID:         s.ID,
		UserID:     s.UserID,
		OrgID:      s.OrgID,
		DeviceID:   s.DeviceID,
		ExpiresAt:  s.ExpiresAt,
		RevokedAt:  timeToNullTime(s.RevokedAt),
		LastSeenAt: timeToNullTime(s.LastSeenAt),
		IpAddress:  sql.NullString{String: s.IPAddress, Valid: s.IPAddress != ""},
		CreatedAt:  s.CreatedAt,
	})
	return err
}

// Revoke marks the session with the given id as revoked. Returns an error if the update fails.
func (r *PostgresRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.queries.RevokeSession(ctx, gen.RevokeSessionParams{
		ID:        id,
		RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
	})
	return err
}

// UpdateLastSeen sets the session's last-seen timestamp for the given id. Returns an error if the update fails.
func (r *PostgresRepository) UpdateLastSeen(ctx context.Context, id string, at time.Time) error {
	_, err := r.queries.UpdateSessionLastSeen(ctx, gen.UpdateSessionLastSeenParams{
		ID:         id,
		LastSeenAt: sql.NullTime{Time: at, Valid: true},
	})
	return err
}

func timeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func nullTimeToPtr(n sql.NullTime) *time.Time {
	if !n.Valid {
		return nil
	}
	return &n.Time
}

func genSessionToDomain(s *gen.Session) *domain.Session {
	if s == nil {
		return nil
	}
	ip := ""
	if s.IpAddress.Valid {
		ip = s.IpAddress.String
	}
	return &domain.Session{
		ID:         s.ID,
		UserID:     s.UserID,
		OrgID:      s.OrgID,
		DeviceID:   s.DeviceID,
		ExpiresAt:  s.ExpiresAt,
		RevokedAt:  nullTimeToPtr(s.RevokedAt),
		LastSeenAt: nullTimeToPtr(s.LastSeenAt),
		IPAddress:  ip,
		CreatedAt:  s.CreatedAt,
	}
}
