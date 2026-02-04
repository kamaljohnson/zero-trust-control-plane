package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	organizationv1 "zero-trust-control-plane/backend/api/generated/organization/v1"
	"zero-trust-control-plane/backend/internal/organization/domain"
)

// mockOrgRepo implements organizationrepo.Repository for tests.
type mockOrgRepo struct {
	orgs      map[string]*domain.Org
	getByIDErr error
}

func (m *mockOrgRepo) GetOrganizationByID(ctx context.Context, id string) (*domain.Org, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.orgs[id], nil
}

func (m *mockOrgRepo) CreateOrganization(ctx context.Context, o *domain.Org) error {
	return nil
}

func (m *mockOrgRepo) UpdateOrganization(ctx context.Context, o *domain.Org) error {
	return nil
}

func TestGetOrganization_Success(t *testing.T) {
	now := time.Now().UTC()
	org := &domain.Org{
		ID:        "org-1",
		Name:      "Test Organization",
		Status:    domain.OrgStatusActive,
		CreatedAt: now,
	}
	repo := &mockOrgRepo{
		orgs: map[string]*domain.Org{"org-1": org},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.GetOrganization(ctx, &organizationv1.GetOrganizationRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("GetOrganization: %v", err)
	}
	if resp == nil || resp.Organization == nil {
		t.Fatal("response or organization is nil")
	}
	if resp.Organization.Id != "org-1" {
		t.Errorf("org id = %q, want %q", resp.Organization.Id, "org-1")
	}
	if resp.Organization.Name != "Test Organization" {
		t.Errorf("org name = %q, want %q", resp.Organization.Name, "Test Organization")
	}
	if resp.Organization.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE {
		t.Errorf("org status = %v, want %v", resp.Organization.Status, organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE)
	}
}

func TestGetOrganization_NotFound(t *testing.T) {
	repo := &mockOrgRepo{
		orgs: make(map[string]*domain.Org),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetOrganization(ctx, &organizationv1.GetOrganizationRequest{OrgId: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent organization")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestGetOrganization_InvalidOrgID(t *testing.T) {
	repo := &mockOrgRepo{orgs: make(map[string]*domain.Org)}
	srv := NewServer(repo)
	ctx := context.Background()

	testCases := []struct {
		name  string
		orgID string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"only spaces", "\t\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.GetOrganization(ctx, &organizationv1.GetOrganizationRequest{OrgId: tc.orgID})
			if err == nil {
				t.Fatal("expected error for invalid org_id")
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("error is not a gRPC status: %v", err)
			}
			if st.Code() != codes.InvalidArgument {
				t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
			}
		})
	}
}

func TestGetOrganization_RepositoryError(t *testing.T) {
	repo := &mockOrgRepo{
		orgs:       make(map[string]*domain.Org),
		getByIDErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.GetOrganization(ctx, &organizationv1.GetOrganizationRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for repository error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestGetOrganization_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.GetOrganization(ctx, &organizationv1.GetOrganizationRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for nil repo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestGetOrganization_SuspendedStatus(t *testing.T) {
	now := time.Now().UTC()
	org := &domain.Org{
		ID:        "org-1",
		Name:      "Suspended Org",
		Status:    domain.OrgStatusSuspended,
		CreatedAt: now,
	}
	repo := &mockOrgRepo{
		orgs: map[string]*domain.Org{"org-1": org},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.GetOrganization(ctx, &organizationv1.GetOrganizationRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("GetOrganization: %v", err)
	}
	if resp.Organization.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_SUSPENDED {
		t.Errorf("org status = %v, want %v", resp.Organization.Status, organizationv1.OrganizationStatus_ORGANIZATION_STATUS_SUSPENDED)
	}
}

func TestCreateOrganization_Unimplemented(t *testing.T) {
	repo := &mockOrgRepo{orgs: make(map[string]*domain.Org)}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{Name: "New Org"})
	if err == nil {
		t.Fatal("expected error for unimplemented method")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestListOrganizations_Unimplemented(t *testing.T) {
	repo := &mockOrgRepo{orgs: make(map[string]*domain.Org)}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.ListOrganizations(ctx, &organizationv1.ListOrganizationsRequest{})
	if err == nil {
		t.Fatal("expected error for unimplemented method")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestSuspendOrganization_Unimplemented(t *testing.T) {
	repo := &mockOrgRepo{orgs: make(map[string]*domain.Org)}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.SuspendOrganization(ctx, &organizationv1.SuspendOrganizationRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for unimplemented method")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

// Tests for domainOrgToProto helper function

func TestDomainOrgToProto_NilOrg(t *testing.T) {
	proto := domainOrgToProto(nil)
	if proto != nil {
		t.Error("domainOrgToProto(nil) should return nil")
	}
}

func TestDomainOrgToProto_ActiveStatus(t *testing.T) {
	now := time.Now().UTC()
	org := &domain.Org{
		ID:        "org-1",
		Name:      "Test Org",
		Status:    domain.OrgStatusActive,
		CreatedAt: now,
	}
	proto := domainOrgToProto(org)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE {
		t.Errorf("Status = %v, want ACTIVE", proto.Status)
	}
	if proto.Id != "org-1" {
		t.Errorf("Id = %q, want %q", proto.Id, "org-1")
	}
	if proto.Name != "Test Org" {
		t.Errorf("Name = %q, want %q", proto.Name, "Test Org")
	}
	if proto.CreatedAt == nil {
		t.Error("CreatedAt should be set")
	}
}

func TestDomainOrgToProto_SuspendedStatus(t *testing.T) {
	now := time.Now().UTC()
	org := &domain.Org{
		ID:        "org-1",
		Name:      "Test Org",
		Status:    domain.OrgStatusSuspended,
		CreatedAt: now,
	}
	proto := domainOrgToProto(org)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_SUSPENDED {
		t.Errorf("Status = %v, want SUSPENDED", proto.Status)
	}
}

func TestDomainOrgToProto_UnknownStatus(t *testing.T) {
	now := time.Now().UTC()
	org := &domain.Org{
		ID:        "org-1",
		Name:      "Test Org",
		Status:    domain.OrgStatus("unknown"),
		CreatedAt: now,
	}
	proto := domainOrgToProto(org)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_UNSPECIFIED {
		t.Errorf("Status = %v, want UNSPECIFIED", proto.Status)
	}
}

func TestDomainOrgToProto_EmptyStatus(t *testing.T) {
	now := time.Now().UTC()
	org := &domain.Org{
		ID:        "org-1",
		Name:      "Test Org",
		Status:    "",
		CreatedAt: now,
	}
	proto := domainOrgToProto(org)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_UNSPECIFIED {
		t.Errorf("Status = %v, want UNSPECIFIED", proto.Status)
	}
}

func TestDomainOrgToProto_AllFields(t *testing.T) {
	now := time.Now().UTC()
	org := &domain.Org{
		ID:        "org-123",
		Name:      "My Organization",
		Status:    domain.OrgStatusActive,
		CreatedAt: now,
	}
	proto := domainOrgToProto(org)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Id != "org-123" {
		t.Errorf("Id = %q, want %q", proto.Id, "org-123")
	}
	if proto.Name != "My Organization" {
		t.Errorf("Name = %q, want %q", proto.Name, "My Organization")
	}
	if proto.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE {
		t.Errorf("Status = %v, want ACTIVE", proto.Status)
	}
	if proto.CreatedAt == nil {
		t.Error("CreatedAt should be set")
	}
	if !proto.CreatedAt.AsTime().Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", proto.CreatedAt.AsTime(), now)
	}
}
