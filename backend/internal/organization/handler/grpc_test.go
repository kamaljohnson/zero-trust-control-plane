package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	organizationv1 "zero-trust-control-plane/backend/api/generated/organization/v1"
	membershipdomain "zero-trust-control-plane/backend/internal/membership/domain"
	organizationdomain "zero-trust-control-plane/backend/internal/organization/domain"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

// mockOrgRepo implements organizationrepo.Repository for tests.
type mockOrgRepo struct {
	orgs           map[string]*organizationdomain.Org
	getByIDErr     error
	createErr      error
	createdOrgs    map[string]*organizationdomain.Org
}

func (m *mockOrgRepo) GetOrganizationByID(ctx context.Context, id string) (*organizationdomain.Org, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.orgs[id], nil
}

func (m *mockOrgRepo) CreateOrganization(ctx context.Context, o *organizationdomain.Org) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.createdOrgs == nil {
		m.createdOrgs = make(map[string]*organizationdomain.Org)
	}
	m.createdOrgs[o.ID] = o
	return nil
}

func (m *mockOrgRepo) UpdateOrganization(ctx context.Context, o *organizationdomain.Org) error {
	return nil
}

// mockUserRepo implements userrepo.Repository for tests.
type mockUserRepo struct {
	users map[string]*userdomain.User
	err   error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.users[id], nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, u *userdomain.User) error {
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *userdomain.User) error {
	return nil
}

func (m *mockUserRepo) SetPhoneVerified(ctx context.Context, userID, phone string) error {
	return nil
}

// mockMembershipRepo implements membershiprepo.Repository for tests.
type mockMembershipRepo struct {
	memberships map[string]*membershipdomain.Membership // key: userID:orgID
	createErr   error
}

func (m *mockMembershipRepo) GetMembershipByID(ctx context.Context, id string) (*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepo) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*membershipdomain.Membership, error) {
	key := userID + ":" + orgID
	return m.memberships[key], nil
}

func (m *mockMembershipRepo) ListMembershipsByOrg(ctx context.Context, orgID string) ([]*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepo) CreateMembership(ctx context.Context, mem *membershipdomain.Membership) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.memberships == nil {
		m.memberships = make(map[string]*membershipdomain.Membership)
	}
	key := mem.UserID + ":" + mem.OrgID
	m.memberships[key] = mem
	return nil
}

func (m *mockMembershipRepo) DeleteByUserAndOrg(ctx context.Context, userID, orgID string) error {
	return nil
}

func (m *mockMembershipRepo) UpdateRole(ctx context.Context, userID, orgID string, role membershipdomain.Role) (*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepo) CountOwnersByOrg(ctx context.Context, orgID string) (int64, error) {
	return 0, nil
}

func TestGetOrganization_Success(t *testing.T) {
	now := time.Now().UTC()
	org := &organizationdomain.Org{
		ID:        "org-1",
		Name:      "Test Organization",
		Status:    organizationdomain.OrgStatusActive,
		CreatedAt: now,
	}
	repo := &mockOrgRepo{
		orgs: map[string]*organizationdomain.Org{"org-1": org},
	}
	srv := NewServer(repo, nil, nil)
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
		orgs: make(map[string]*organizationdomain.Org),
	}
	srv := NewServer(repo, nil, nil)
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
	repo := &mockOrgRepo{orgs: make(map[string]*organizationdomain.Org)}
	srv := NewServer(repo, nil, nil)
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
		orgs:       make(map[string]*organizationdomain.Org),
		getByIDErr: errors.New("database error"),
	}
	srv := NewServer(repo, nil, nil)
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
	srv := NewServer(nil, nil, nil)
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
	org := &organizationdomain.Org{
		ID:        "org-1",
		Name:      "Suspended Org",
		Status:    organizationdomain.OrgStatusSuspended,
		CreatedAt: now,
	}
	repo := &mockOrgRepo{
		orgs: map[string]*organizationdomain.Org{"org-1": org},
	}
	srv := NewServer(repo, nil, nil)
	ctx := context.Background()

	resp, err := srv.GetOrganization(ctx, &organizationv1.GetOrganizationRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("GetOrganization: %v", err)
	}
	if resp.Organization.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_SUSPENDED {
		t.Errorf("org status = %v, want %v", resp.Organization.Status, organizationv1.OrganizationStatus_ORGANIZATION_STATUS_SUSPENDED)
	}
}

func TestCreateOrganization_Success(t *testing.T) {
	userID := "user-1"
	orgName := "Test Organization"
	now := time.Now().UTC()
	user := &userdomain.User{
		ID:        userID,
		Email:     "user@example.com",
		Name:      "Test User",
		Status:    userdomain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	orgRepo := &mockOrgRepo{
		orgs:        make(map[string]*organizationdomain.Org),
		createdOrgs: make(map[string]*organizationdomain.Org),
	}
	userRepo := &mockUserRepo{
		users: map[string]*userdomain.User{userID: user},
	}
	membershipRepo := &mockMembershipRepo{
		memberships: make(map[string]*membershipdomain.Membership),
	}

	srv := NewServer(orgRepo, userRepo, membershipRepo)
	ctx := context.Background()

	resp, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   orgName,
		UserId: userID,
	})
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	if resp == nil || resp.Organization == nil {
		t.Fatal("response or organization is nil")
	}
	if resp.Organization.Id == "" {
		t.Error("organization id should not be empty")
	}
	if resp.Organization.Name != orgName {
		t.Errorf("org name = %q, want %q", resp.Organization.Name, orgName)
	}
	if resp.Organization.Status != organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE {
		t.Errorf("org status = %v, want %v", resp.Organization.Status, organizationv1.OrganizationStatus_ORGANIZATION_STATUS_ACTIVE)
	}
	if resp.Organization.CreatedAt == nil {
		t.Error("created_at should be set")
	}

	// Verify organization was created
	createdOrg := orgRepo.createdOrgs[resp.Organization.Id]
	if createdOrg == nil {
		t.Fatal("organization was not created in repository")
	}
	if createdOrg.Name != orgName {
		t.Errorf("created org name = %q, want %q", createdOrg.Name, orgName)
	}
	if createdOrg.Status != organizationdomain.OrgStatusActive {
		t.Errorf("created org status = %v, want %v", createdOrg.Status, organizationdomain.OrgStatusActive)
	}

	// Verify membership was created with owner role
	membershipKey := userID + ":" + resp.Organization.Id
	membership := membershipRepo.memberships[membershipKey]
	if membership == nil {
		t.Fatal("membership was not created")
	}
	if membership.UserID != userID {
		t.Errorf("membership user_id = %q, want %q", membership.UserID, userID)
	}
	if membership.OrgID != resp.Organization.Id {
		t.Errorf("membership org_id = %q, want %q", membership.OrgID, resp.Organization.Id)
	}
	if membership.Role != membershipdomain.RoleOwner {
		t.Errorf("membership role = %v, want %v", membership.Role, membershipdomain.RoleOwner)
	}
}

func TestCreateOrganization_MissingName(t *testing.T) {
	userID := "user-1"
	userRepo := &mockUserRepo{
		users: map[string]*userdomain.User{userID: {ID: userID}},
	}
	srv := NewServer(&mockOrgRepo{}, userRepo, &mockMembershipRepo{})
	ctx := context.Background()

	testCases := []struct {
		name string
		req  *organizationv1.CreateOrganizationRequest
	}{
		{"empty name", &organizationv1.CreateOrganizationRequest{Name: "", UserId: userID}},
		{"whitespace name", &organizationv1.CreateOrganizationRequest{Name: "   ", UserId: userID}},
		{"only spaces", &organizationv1.CreateOrganizationRequest{Name: "\t\n", UserId: userID}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.CreateOrganization(ctx, tc.req)
			if err == nil {
				t.Fatal("expected error for missing name")
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

func TestCreateOrganization_MissingUserID(t *testing.T) {
	srv := NewServer(&mockOrgRepo{}, &mockUserRepo{}, &mockMembershipRepo{})
	ctx := context.Background()

	testCases := []struct {
		name string
		req  *organizationv1.CreateOrganizationRequest
	}{
		{"empty user_id", &organizationv1.CreateOrganizationRequest{Name: "Test Org", UserId: ""}},
		{"whitespace user_id", &organizationv1.CreateOrganizationRequest{Name: "Test Org", UserId: "   "}},
		{"only spaces", &organizationv1.CreateOrganizationRequest{Name: "Test Org", UserId: "\t\n"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.CreateOrganization(ctx, tc.req)
			if err == nil {
				t.Fatal("expected error for missing user_id")
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

func TestCreateOrganization_UserNotFound(t *testing.T) {
	userID := "nonexistent-user"
	userRepo := &mockUserRepo{
		users: make(map[string]*userdomain.User),
	}
	srv := NewServer(&mockOrgRepo{}, userRepo, &mockMembershipRepo{})
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   "Test Org",
		UserId: userID,
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestCreateOrganization_UserRepoError(t *testing.T) {
	userID := "user-1"
	userRepo := &mockUserRepo{
		users: make(map[string]*userdomain.User),
		err:   errors.New("database error"),
	}
	srv := NewServer(&mockOrgRepo{}, userRepo, &mockMembershipRepo{})
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   "Test Org",
		UserId: userID,
	})
	if err == nil {
		t.Fatal("expected error for user repo error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestCreateOrganization_OrgRepoError(t *testing.T) {
	userID := "user-1"
	now := time.Now().UTC()
	user := &userdomain.User{
		ID:        userID,
		Email:     "user@example.com",
		Status:    userdomain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	orgRepo := &mockOrgRepo{
		orgs:     make(map[string]*organizationdomain.Org),
		createErr: errors.New("database error"),
	}
	userRepo := &mockUserRepo{
		users: map[string]*userdomain.User{userID: user},
	}
	srv := NewServer(orgRepo, userRepo, &mockMembershipRepo{})
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   "Test Org",
		UserId: userID,
	})
	if err == nil {
		t.Fatal("expected error for org repo error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestCreateOrganization_MembershipRepoError(t *testing.T) {
	userID := "user-1"
	now := time.Now().UTC()
	user := &userdomain.User{
		ID:        userID,
		Email:     "user@example.com",
		Status:    userdomain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	orgRepo := &mockOrgRepo{
		orgs:        make(map[string]*organizationdomain.Org),
		createdOrgs: make(map[string]*organizationdomain.Org),
	}
	userRepo := &mockUserRepo{
		users: map[string]*userdomain.User{userID: user},
	}
	membershipRepo := &mockMembershipRepo{
		memberships: make(map[string]*membershipdomain.Membership),
		createErr:   errors.New("database error"),
	}
	srv := NewServer(orgRepo, userRepo, membershipRepo)
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   "Test Org",
		UserId: userID,
	})
	if err == nil {
		t.Fatal("expected error for membership repo error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestCreateOrganization_NilOrgRepo(t *testing.T) {
	srv := NewServer(nil, &mockUserRepo{}, &mockMembershipRepo{})
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   "Test Org",
		UserId: "user-1",
	})
	if err == nil {
		t.Fatal("expected error for nil orgRepo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestCreateOrganization_NilUserRepo(t *testing.T) {
	srv := NewServer(&mockOrgRepo{}, nil, &mockMembershipRepo{})
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   "Test Org",
		UserId: "user-1",
	})
	if err == nil {
		t.Fatal("expected error for nil userRepo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestCreateOrganization_NilMembershipRepo(t *testing.T) {
	userID := "user-1"
	now := time.Now().UTC()
	user := &userdomain.User{
		ID:        userID,
		Email:     "user@example.com",
		Status:    userdomain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	srv := NewServer(&mockOrgRepo{}, &mockUserRepo{users: map[string]*userdomain.User{userID: user}}, nil)
	ctx := context.Background()

	_, err := srv.CreateOrganization(ctx, &organizationv1.CreateOrganizationRequest{
		Name:   "Test Org",
		UserId: userID,
	})
	if err == nil {
		t.Fatal("expected error for nil membershipRepo")
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
	repo := &mockOrgRepo{orgs: make(map[string]*organizationdomain.Org)}
	srv := NewServer(repo, nil, nil)
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
	repo := &mockOrgRepo{orgs: make(map[string]*organizationdomain.Org)}
	srv := NewServer(repo, nil, nil)
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
	org := &organizationdomain.Org{
		ID:        "org-1",
		Name:      "Test Org",
		Status:    organizationdomain.OrgStatusActive,
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
	org := &organizationdomain.Org{
		ID:        "org-1",
		Name:      "Test Org",
		Status:    organizationdomain.OrgStatusSuspended,
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
	org := &organizationdomain.Org{
		ID:        "org-1",
		Name:      "Test Org",
		Status:    organizationdomain.OrgStatus("unknown"),
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
	org := &organizationdomain.Org{
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
	org := &organizationdomain.Org{
		ID:        "org-123",
		Name:      "My Organization",
		Status:    organizationdomain.OrgStatusActive,
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
