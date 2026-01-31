package repository

import (
	"context"
	"database/sql"
	"errors"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/identity/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns an identity repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByID returns the identity for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Identity, error) {
	i, err := r.queries.GetIdentity(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genIdentityToDomain(&i), nil
}

// GetByUserAndProvider returns the identity for the given user and provider, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByUserAndProvider(ctx context.Context, userID string, provider domain.IdentityProvider) (*domain.Identity, error) {
	i, err := r.queries.GetIdentityByUserAndProvider(ctx, gen.GetIdentityByUserAndProviderParams{UserID: userID, Provider: gen.IdentityProvider(provider)})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genIdentityToDomain(&i), nil
}

// GetByUserAndProviderID returns the identity for the given user, provider, and providerID, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByUserAndProviderID(ctx context.Context, userID string, provider domain.IdentityProvider, providerID string) (*domain.Identity, error) {
	i, err := r.queries.GetIdentityByUserAndProviderID(ctx, gen.GetIdentityByUserAndProviderIDParams{
		UserID: userID, Provider: gen.IdentityProvider(provider), ProviderID: providerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genIdentityToDomain(&i), nil
}

// Create persists the identity to the database. The identity must have ID set.
func (r *PostgresRepository) Create(ctx context.Context, i *domain.Identity) error {
	ph := sql.NullString{String: i.PasswordHash, Valid: i.PasswordHash != ""}
	_, err := r.queries.CreateIdentity(ctx, gen.CreateIdentityParams{
		ID:           i.ID,
		UserID:       i.UserID,
		Provider:     gen.IdentityProvider(i.Provider),
		ProviderID:   i.ProviderID,
		PasswordHash: ph,
		CreatedAt:    i.CreatedAt,
	})
	return err
}

// UpdatePasswordHash updates the password hash for the identity with the given id. Returns an error if the update fails.
func (r *PostgresRepository) UpdatePasswordHash(ctx context.Context, id string, passwordHash string) error {
	ph := sql.NullString{String: passwordHash, Valid: passwordHash != ""}
	_, err := r.queries.UpdateIdentityPasswordHash(ctx, gen.UpdateIdentityPasswordHashParams{ID: id, PasswordHash: ph})
	return err
}

func genIdentityToDomain(i *gen.Identity) *domain.Identity {
	if i == nil {
		return nil
	}
	ph := ""
	if i.PasswordHash.Valid {
		ph = i.PasswordHash.String
	}
	return &domain.Identity{
		ID:           i.ID,
		UserID:       i.UserID,
		Provider:     domain.IdentityProvider(i.Provider),
		ProviderID:   i.ProviderID,
		PasswordHash: ph,
		CreatedAt:    i.CreatedAt,
	}
}
