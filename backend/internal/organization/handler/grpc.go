package handler

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	organizationv1 "zero-trust-control-plane/backend/api/generated/organization/v1"
	membershipdomain "zero-trust-control-plane/backend/internal/membership/domain"
	membershiprepo "zero-trust-control-plane/backend/internal/membership/repository"
	organizationdomain "zero-trust-control-plane/backend/internal/organization/domain"
	organizationrepo "zero-trust-control-plane/backend/internal/organization/repository"
	userrepo "zero-trust-control-plane/backend/internal/user/repository"
)

// Server implements OrganizationService (proto server) for multi-tenancy and org management.
// Proto: organization/organization.proto → internal/organization/handler.
type Server struct {
	organizationv1.UnimplementedOrganizationServiceServer
	orgRepo        organizationrepo.Repository
	userRepo       userrepo.Repository
	membershipRepo membershiprepo.Repository
}

// NewServer returns a new Organization gRPC server.
// If orgRepo, userRepo, or membershipRepo is nil, CreateOrganization returns Unimplemented.
// Other RPCs may return Unimplemented if orgRepo is nil.
func NewServer(orgRepo organizationrepo.Repository, userRepo userrepo.Repository, membershipRepo membershiprepo.Repository) *Server {
	return &Server{
		orgRepo:        orgRepo,
		userRepo:       userRepo,
		membershipRepo: membershipRepo,
	}
}

// CreateOrganization creates a new organization with the given name and assigns the user as owner.
// The organization is auto-activated (status=active) for PoC. Requires user_id and name.
func (s *Server) CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest) (*organizationv1.CreateOrganizationResponse, error) {
	if s.orgRepo == nil || s.userRepo == nil || s.membershipRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method CreateOrganization not implemented")
	}

	name := strings.TrimSpace(req.GetName())
	userID := strings.TrimSpace(req.GetUserId())

	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to look up user")
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Generate org ID and create organization
	orgID := uuid.New().String()
	now := time.Now().UTC()
	org := &organizationdomain.Org{
		ID:        orgID,
		Name:      name,
		Status:    organizationdomain.OrgStatusActive,
		CreatedAt: now,
	}
	if err := org.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.orgRepo.CreateOrganization(ctx, org); err != nil {
		return nil, status.Error(codes.Internal, "failed to create organization")
	}

	// Create membership with owner role
	membershipID := uuid.New().String()
	membership := &membershipdomain.Membership{
		ID:        membershipID,
		UserID:    userID,
		OrgID:     orgID,
		Role:      membershipdomain.RoleOwner,
		CreatedAt: now,
	}

	if err := s.membershipRepo.CreateMembership(ctx, membership); err != nil {
		// If membership creation fails, we should ideally rollback org creation,
		// but for simplicity in PoC, we'll just return an error.
		// In production, this should be a transaction.
		return nil, status.Error(codes.Internal, "failed to create membership")
	}

	return &organizationv1.CreateOrganizationResponse{
		Organization: domainOrgToProto(org),
	}, nil
}

// GetOrganization returns an organization by ID.
func (s *Server) GetOrganization(ctx context.Context, req *organizationv1.GetOrganizationRequest) (*organizationv1.GetOrganizationResponse, error) {
	if s.orgRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method GetOrganization not implemented")
	}
	orgID := strings.TrimSpace(req.GetOrgId())
	if orgID == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	o, err := s.orgRepo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to look up organization")
	}
	if o == nil {
		return nil, status.Error(codes.NotFound, "organization not found")
	}
	return &organizationv1.GetOrganizationResponse{
		Organization: domainOrgToProto(o),
	}, nil
}

// ListOrganizations returns a paginated list of organizations. TODO: implement.
func (s *Server) ListOrganizations(ctx context.Context, req *organizationv1.ListOrganizationsRequest) (*organizationv1.ListOrganizationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListOrganizations not implemented")
}

// SuspendOrganization suspends an organization. TODO: implement.
func (s *Server) SuspendOrganization(ctx context.Context, req *organizationv1.SuspendOrganizationRequest) (*organizationv1.SuspendOrganizationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method SuspendOrganization not implemented")
}

func domainOrgToProto(o *organizationdomain.Org) *organizationv1.Organization {
	if o == nil {
		return nil
	}
	var status organizationv1.OrganizationStatus
	switch o.Status {
	case organizationdomain.OrgStatusActive:
		status = organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE
	case organizationdomain.OrgStatusSuspended:
		status = organizationv1.OrganizationStatus_ORGANIZATION_STATUS_SUSPENDED
	default:
		status = organizationv1.OrganizationStatus_ORGANIZATION_STATUS_UNSPECIFIED
	}
	return &organizationv1.Organization{
		Id:        o.ID,
		Name:      o.Name,
		Status:    status,
		CreatedAt: timestamppb.New(o.CreatedAt),
	}
}
