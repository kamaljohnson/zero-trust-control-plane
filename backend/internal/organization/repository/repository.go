package repository

import (
	"context"

	"zero-trust-control-plane/backend/internal/organization/domain"
)

// Repository defines persistence for organizations.
type Repository interface {
	GetOrganizationByID(ctx context.Context, id string) (*domain.Org, error)
	CreateOrganization(ctx context.Context, o *domain.Org) error
	UpdateOrganization(ctx context.Context, o *domain.Org) error
}
