package handler

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	organizationv1 "zero-trust-control-plane/backend/api/generated/organization/v1"
	"zero-trust-control-plane/backend/internal/organization/domain"
	organizationrepo "zero-trust-control-plane/backend/internal/organization/repository"
)

// Server implements OrganizationService (proto server) for multi-tenancy and org management.
// Proto: organization/organization.proto → internal/organization/handler.
type Server struct {
	organizationv1.UnimplementedOrganizationServiceServer
	orgRepo organizationrepo.Repository
}

// NewServer returns a new Organization gRPC server. orgRepo may be nil; then all RPCs return Unimplemented.
func NewServer(orgRepo organizationrepo.Repository) *Server {
	return &Server{orgRepo: orgRepo}
}

// CreateOrganization creates a new organization. TODO: implement.
func (s *Server) CreateOrganization(ctx context.Context, req *organizationv1.CreateOrganizationRequest) (*organizationv1.CreateOrganizationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateOrganization not implemented")
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

func domainOrgToProto(o *domain.Org) *organizationv1.Organization {
	if o == nil {
		return nil
	}
	var status organizationv1.OrganizationStatus
	switch o.Status {
	case domain.OrgStatusActive:
		status = organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE
	case domain.OrgStatusSuspended:
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
