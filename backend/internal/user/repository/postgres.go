package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"zero-trust-control-plane/backend/internal/db/sqlc/gen"
	"zero-trust-control-plane/backend/internal/user/domain"
)

type PostgresRepository struct {
	queries *gen.Queries
}

// NewPostgresRepository returns a user repository that uses the given db for persistence.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{queries: gen.New(db)}
}

// GetByID returns the user for id, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := r.queries.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genUserToDomain(&u), nil
}

// GetByEmail returns the user with the given email, or nil if not found.
// It returns an error only for database failures, not for missing rows.
func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return genUserToDomain(&u), nil
}

// Create persists the user to the database. The user must have ID set; it is not assigned by this method.
func (r *PostgresRepository) Create(ctx context.Context, u *domain.User) error {
	name := sql.NullString{String: u.Name, Valid: u.Name != ""}
	phone := sql.NullString{String: u.Phone, Valid: u.Phone != ""}
	_, err := r.queries.CreateUser(ctx, gen.CreateUserParams{
		ID:            u.ID,
		Email:         u.Email,
		Name:          name,
		Phone:         phone,
		PhoneVerified: u.PhoneVerified,
		Status:        gen.UserStatus(u.Status),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	})
	return err
}

// Update updates the existing user record in the database. If the user has PhoneVerified true, phone is preserved and not overwritten.
func (r *PostgresRepository) Update(ctx context.Context, u *domain.User) error {
	current, err := r.queries.GetUser(ctx, u.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	name := sql.NullString{String: u.Name, Valid: u.Name != ""}
	phone := sql.NullString{String: u.Phone, Valid: u.Phone != ""}
	if current.PhoneVerified {
		phone = current.Phone
	}
	_, err = r.queries.UpdateUser(ctx, gen.UpdateUserParams{
		ID:            u.ID,
		Email:         u.Email,
		Name:          name,
		Phone:         phone,
		PhoneVerified: current.PhoneVerified,
		Status:        gen.UserStatus(u.Status),
		UpdatedAt:     u.UpdatedAt,
	})
	return err
}

// SetPhoneVerified sets the user's phone and phone_verified only when phone is currently empty and not verified. Returns nil if no row was updated.
func (r *PostgresRepository) SetPhoneVerified(ctx context.Context, userID, phone string) error {
	phoneVal := sql.NullString{String: phone, Valid: phone != ""}
	_, err := r.queries.SetPhoneVerified(ctx, gen.SetPhoneVerifiedParams{
		ID:        userID,
		Phone:     phoneVal,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}

func genUserToDomain(u *gen.User) *domain.User {
	if u == nil {
		return nil
	}
	name := ""
	if u.Name.Valid {
		name = u.Name.String
	}
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	return &domain.User{
		ID:            u.ID,
		Email:         u.Email,
		Name:          name,
		Phone:         phone,
		PhoneVerified: u.PhoneVerified,
		Status:        domain.UserStatus(u.Status),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
