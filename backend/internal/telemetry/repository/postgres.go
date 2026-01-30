package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/telemetry/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns a telemetry repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// Save persists the telemetry event to the database. It sets t.ID on success.
func (r *PostgresRepository) Save(ctx context.Context, t *domain.Telemetry) error {
	row, err := r.queries.CreateTelemetry(ctx, gen.CreateTelemetryParams{
		OrgID:     t.OrgID,
		UserID:    nullStringFromPtr(t.UserID),
		DeviceID:  nullStringFromPtr(t.DeviceID),
		SessionID: nullStringFromPtr(t.SessionID),
		EventType: t.EventType,
		Source:    t.Source,
		Metadata:  telemetryMetadata(t.Metadata),
		CreatedAt: t.CreatedAt,
	})
	if err != nil {
		return err
	}
	t.ID = row.ID
	return nil
}

// GetByID returns the telemetry record for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Telemetry, error) {
	row, err := r.queries.GetTelemetry(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genTelemetryToDomain(&row), nil
}

// ListByOrg returns telemetry events for the given org, paginated by limit and offset.
// Returns (nil, error) only on database errors.
func (r *PostgresRepository) ListByOrg(ctx context.Context, orgID string, limit, offset int32) ([]*domain.Telemetry, error) {
	list, err := r.queries.ListTelemetryByOrg(ctx, gen.ListTelemetryByOrgParams{
		OrgID: orgID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Telemetry, len(list))
	for i := range list {
		out[i] = genTelemetryToDomain(&list[i])
	}
	return out, nil
}

func nullStringFromPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func ptrFromNullString(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	return &n.String
}

func telemetryMetadata(b []byte) json.RawMessage {
	if b == nil {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}

func genTelemetryToDomain(t *gen.Telemetry) *domain.Telemetry {
	if t == nil {
		return nil
	}
	meta := t.Metadata
	if meta == nil {
		meta = []byte("{}")
	}
	return &domain.Telemetry{
		ID:        t.ID,
		OrgID:     t.OrgID,
		UserID:    ptrFromNullString(t.UserID),
		DeviceID:  ptrFromNullString(t.DeviceID),
		SessionID: ptrFromNullString(t.SessionID),
		EventType: t.EventType,
		Source:    t.Source,
		Metadata:  meta,
		CreatedAt: t.CreatedAt,
	}
}
