package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/audit/domain"
	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns an audit log repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByID returns the audit log for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.AuditLog, error) {
	a, err := r.queries.GetAuditLog(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genAuditLogToDomain(&a), nil
}

// ListByOrg returns audit logs for the given org, paginated by limit and offset.
// Returns (nil, error) only on database errors.
func (r *PostgresRepository) ListByOrg(ctx context.Context, orgID string, limit, offset int32) ([]*domain.AuditLog, error) {
	list, err := r.queries.ListAuditLogsByOrg(ctx, gen.ListAuditLogsByOrgParams{OrgID: orgID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.AuditLog, len(list))
	for i := range list {
		out[i] = genAuditLogToDomain(&list[i])
	}
	return out, nil
}

// Save persists the audit log to the database. The audit log must have ID set.
func (r *PostgresRepository) Save(ctx context.Context, a *domain.AuditLog) error {
	uid := sql.NullString{String: a.UserID, Valid: a.UserID != ""}
	meta := sql.NullString{String: a.Metadata, Valid: a.Metadata != ""}
	_, err := r.queries.CreateAuditLog(ctx, gen.CreateAuditLogParams{
		ID: a.ID, OrgID: a.OrgID, UserID: uid, Action: a.Action, Resource: a.Resource,
		Ip: a.IP, Metadata: meta, CreatedAt: a.CreatedAt,
	})
	return err
}

func genAuditLogToDomain(a *gen.AuditLog) *domain.AuditLog {
	if a == nil {
		return nil
	}
	uid := ""
	if a.UserID.Valid {
		uid = a.UserID.String
	}
	meta := ""
	if a.Metadata.Valid {
		meta = a.Metadata.String
	}
	return &domain.AuditLog{
		ID: a.ID, OrgID: a.OrgID, UserID: uid, Action: a.Action, Resource: a.Resource,
		IP: a.Ip, Metadata: meta, CreatedAt: a.CreatedAt,
	}
}
