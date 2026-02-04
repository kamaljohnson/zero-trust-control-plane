package handler

import (
	"context"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "zero-trust-control-plane/backend/api/generated/common/v1"
	membershipv1 "zero-trust-control-plane/backend/api/generated/membership/v1"
	"zero-trust-control-plane/backend/internal/membership/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

// mockMembershipRepo implements membershiprepo.Repository for tests.
type mockMembershipRepo struct {
	memberships map[string]*domain.Membership // key: userID:orgID
	byID        map[string]*domain.Membership
	ownerCounts map[string]int64
	createErr   error
	deleteErr   error
	updateErr   error
	listErr     error
	countErr    error
}

func (m *mockMembershipRepo) GetMembershipByID(ctx context.Context, id string) (*domain.Membership, error) {
	return m.byID[id], nil
}

func (m *mockMembershipRepo) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*domain.Membership, error) {
	key := userID + ":" + orgID
	return m.memberships[key], nil
}

func (m *mockMembershipRepo) ListMembershipsByOrg(ctx context.Context, orgID string) ([]*domain.Membership, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*domain.Membership
	for _, m := range m.memberships {
		if m.OrgID == orgID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (m *mockMembershipRepo) CreateMembership(ctx context.Context, mem *domain.Membership) error {
	if m.createErr != nil {
		return m.createErr
	}
	key := mem.UserID + ":" + mem.OrgID
	m.memberships[key] = mem
	m.byID[mem.ID] = mem
	return nil
}

func (m *mockMembershipRepo) DeleteByUserAndOrg(ctx context.Context, userID, orgID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	key := userID + ":" + orgID
	delete(m.memberships, key)
	return nil
}

func (m *mockMembershipRepo) UpdateRole(ctx context.Context, userID, orgID string, role domain.Role) (*domain.Membership, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	key := userID + ":" + orgID
	mem := m.memberships[key]
	if mem == nil {
		return nil, nil
	}
	updated := *mem
	updated.Role = role
	m.memberships[key] = &updated
	return &updated, nil
}

func (m *mockMembershipRepo) CountOwnersByOrg(ctx context.Context, orgID string) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.ownerCounts[orgID], nil
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
	if m.err != nil {
		return nil, m.err
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
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

// mockAuditLogger implements audit.AuditLogger for tests.
type mockAuditLogger struct {
	events []struct {
		orgID, userID, action, resource, resourceID string
	}
}

func (m *mockAuditLogger) LogEvent(ctx context.Context, orgID, userID, action, resource, resourceID string) {
	m.events = append(m.events, struct {
		orgID, userID, action, resource, resourceID string
	}{orgID, userID, action, resource, resourceID})
}

func ctxWithAdmin(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func ctxWithMember(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func TestAddMember_Success(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	userRepo := &mockUserRepo{
		users: map[string]*userdomain.User{
			"user-2": {ID: "user-2", Email: "user2@example.com"},
		},
	}
	auditLogger := &mockAuditLogger{}
	srv := NewServer(membershipRepo, userRepo, auditLogger)
	ctx := ctxWithAdmin("org-1", "admin-1")

	resp, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "user-2",
		OrgId:  "org-1",
		Role:   membershipv1.Role_ROLE_MEMBER,
	})
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if resp == nil || resp.Member == nil {
		t.Fatal("response or member is nil")
	}
	if resp.Member.UserId != "user-2" {
		t.Errorf("member user_id = %q, want %q", resp.Member.UserId, "user-2")
	}
	if resp.Member.OrgId != "org-1" {
		t.Errorf("member org_id = %q, want %q", resp.Member.OrgId, "org-1")
	}
	if resp.Member.Role != membershipv1.Role_ROLE_MEMBER {
		t.Errorf("member role = %v, want %v", resp.Member.Role, membershipv1.Role_ROLE_MEMBER)
	}
	if len(auditLogger.events) != 1 {
		t.Errorf("audit events = %d, want 1", len(auditLogger.events))
	}
}

func TestAddMember_DuplicateMember(t *testing.T) {
	existing := &domain.Membership{
		ID:     "m1",
		UserID: "user-2",
		OrgID:  "org-1",
		Role:   domain.RoleMember,
	}
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
			"user-2:org-1":  existing,
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "user-2",
		OrgId:  "org-1",
	})
	if err == nil {
		t.Fatal("expected error for duplicate member")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.AlreadyExists {
		t.Errorf("status code = %v, want %v", st.Code(), codes.AlreadyExists)
	}
}

func TestAddMember_InvalidUserID(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "",
		OrgId:  "org-1",
	})
	if err == nil {
		t.Fatal("expected error for empty user_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

// Additional tests for RemoveMember, UpdateRole, and domainMemberToProto

func TestRemoveMember_RepositoryError(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"user-1:org-1": {ID: "m1", UserID: "user-1", OrgID: "org-1", Role: domain.RoleMember},
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: map[string]int64{"org-1": 1},
		deleteErr:   status.Error(codes.Internal, "database error"),
	}
	userRepo := &mockUserRepo{
		users: make(map[string]*userdomain.User),
	}
	srv := NewServer(membershipRepo, userRepo, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.RemoveMember(ctx, &membershipv1.RemoveMemberRequest{
		UserId: "user-1",
		OrgId:  "org-1",
	})
	if err == nil {
		t.Fatal("expected error when repository fails")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestUpdateRole_NotFound(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	userRepo := &mockUserRepo{
		users: make(map[string]*userdomain.User),
	}
	srv := NewServer(membershipRepo, userRepo, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.UpdateRole(ctx, &membershipv1.UpdateRoleRequest{
		UserId: "nonexistent",
		OrgId:  "org-1",
		Role:   membershipv1.Role_ROLE_ADMIN,
	})
	if err == nil {
		t.Fatal("expected error for nonexistent membership")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestUpdateRole_RepositoryError(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"user-1:org-1": {ID: "m1", UserID: "user-1", OrgID: "org-1", Role: domain.RoleMember},
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
		updateErr:   status.Error(codes.Internal, "database error"),
	}
	userRepo := &mockUserRepo{
		users: make(map[string]*userdomain.User),
	}
	srv := NewServer(membershipRepo, userRepo, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.UpdateRole(ctx, &membershipv1.UpdateRoleRequest{
		UserId: "user-1",
		OrgId:  "org-1",
		Role:   membershipv1.Role_ROLE_ADMIN,
	})
	if err == nil {
		t.Fatal("expected error when repository fails")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestDomainMemberToProto_Owner(t *testing.T) {
	now := time.Now().UTC()
	member := &domain.Membership{
		ID:        "m1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Role:      domain.RoleOwner,
		CreatedAt: now,
	}

	proto := domainMemberToProto(member)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Id != "m1" {
		t.Errorf("id = %q, want %q", proto.Id, "m1")
	}
	if proto.UserId != "user-1" {
		t.Errorf("user_id = %q, want %q", proto.UserId, "user-1")
	}
	if proto.OrgId != "org-1" {
		t.Errorf("org_id = %q, want %q", proto.OrgId, "org-1")
	}
	if proto.Role != membershipv1.Role_ROLE_OWNER {
		t.Errorf("role = %v, want %v", proto.Role, membershipv1.Role_ROLE_OWNER)
	}
	if proto.CreatedAt == nil {
		t.Error("created_at should be set")
	}
}

func TestDomainMemberToProto_Admin(t *testing.T) {
	now := time.Now().UTC()
	member := &domain.Membership{
		ID:        "m1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Role:      domain.RoleAdmin,
		CreatedAt: now,
	}

	proto := domainMemberToProto(member)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Role != membershipv1.Role_ROLE_ADMIN {
		t.Errorf("role = %v, want %v", proto.Role, membershipv1.Role_ROLE_ADMIN)
	}
}

func TestDomainMemberToProto_Member(t *testing.T) {
	now := time.Now().UTC()
	member := &domain.Membership{
		ID:        "m1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Role:      domain.RoleMember,
		CreatedAt: now,
	}

	proto := domainMemberToProto(member)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Role != membershipv1.Role_ROLE_MEMBER {
		t.Errorf("role = %v, want %v", proto.Role, membershipv1.Role_ROLE_MEMBER)
	}
}

func TestDomainMemberToProto_NilMember(t *testing.T) {
	proto := domainMemberToProto(nil)
	if proto != nil {
		t.Error("proto should be nil for nil member")
	}
}

func TestAddMember_UserNotFound(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	userRepo := &mockUserRepo{users: make(map[string]*userdomain.User)}
	srv := NewServer(membershipRepo, userRepo, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "nonexistent",
		OrgId:  "org-1",
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

func TestAddMember_NonAdminCaller(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: make(map[string]*domain.Membership),
		byID:        make(map[string]*domain.Membership),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithMember("org-1", "member-1")

	_, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "user-2",
		OrgId:  "org-1",
	})
	if err == nil {
		t.Fatal("expected error for non-admin caller")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestAddMember_OrgIDMismatch(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: make(map[string]*domain.Membership),
		byID:        make(map[string]*domain.Membership),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "user-2",
		OrgId:  "org-2",
	})
	if err == nil {
		t.Fatal("expected error for org_id mismatch")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestAddMember_NilRepo(t *testing.T) {
	srv := NewServer(nil, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "user-2",
		OrgId:  "org-1",
	})
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

func TestAddMember_DefaultRole(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	resp, err := srv.AddMember(ctx, &membershipv1.AddMemberRequest{
		UserId: "user-2",
		OrgId:  "org-1",
		Role:   membershipv1.Role_ROLE_UNSPECIFIED,
	})
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if resp.Member.Role != membershipv1.Role_ROLE_MEMBER {
		t.Errorf("member role = %v, want %v", resp.Member.Role, membershipv1.Role_ROLE_MEMBER)
	}
}

func TestRemoveMember_Success(t *testing.T) {
	existing := &domain.Membership{
		ID:     "m1",
		UserID: "user-2",
		OrgID:  "org-1",
		Role:   domain.RoleMember,
	}
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
			"user-2:org-1":  existing,
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: map[string]int64{"org-1": 1},
	}
	auditLogger := &mockAuditLogger{}
	srv := NewServer(membershipRepo, nil, auditLogger)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.RemoveMember(ctx, &membershipv1.RemoveMemberRequest{
		UserId: "user-2",
		OrgId:  "org-1",
	})
	if err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	if len(auditLogger.events) != 1 {
		t.Errorf("audit events = %d, want 1", len(auditLogger.events))
	}
}

func TestRemoveMember_NotFound(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.RemoveMember(ctx, &membershipv1.RemoveMemberRequest{
		UserId: "nonexistent",
		OrgId:  "org-1",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent membership")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestRemoveMember_LastOwnerProtection(t *testing.T) {
	existing := &domain.Membership{
		ID:     "m1",
		UserID: "owner-1",
		OrgID:  "org-1",
		Role:   domain.RoleOwner,
	}
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
			"owner-1:org-1": existing,
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: map[string]int64{"org-1": 1},
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.RemoveMember(ctx, &membershipv1.RemoveMemberRequest{
		UserId: "owner-1",
		OrgId:  "org-1",
	})
	if err == nil {
		t.Fatal("expected error for removing last owner")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("status code = %v, want %v", st.Code(), codes.FailedPrecondition)
	}
}

func TestRemoveMember_NonAdminCaller(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: make(map[string]*domain.Membership),
		byID:        make(map[string]*domain.Membership),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithMember("org-1", "member-1")

	_, err := srv.RemoveMember(ctx, &membershipv1.RemoveMemberRequest{
		UserId: "user-2",
		OrgId:  "org-1",
	})
	if err == nil {
		t.Fatal("expected error for non-admin caller")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestUpdateRole_Success(t *testing.T) {
	existing := &domain.Membership{
		ID:     "m1",
		UserID: "user-2",
		OrgID:  "org-1",
		Role:   domain.RoleMember,
	}
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
			"user-2:org-1":  existing,
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: map[string]int64{"org-1": 1},
	}
	auditLogger := &mockAuditLogger{}
	srv := NewServer(membershipRepo, nil, auditLogger)
	ctx := ctxWithAdmin("org-1", "admin-1")

	resp, err := srv.UpdateRole(ctx, &membershipv1.UpdateRoleRequest{
		UserId: "user-2",
		OrgId:  "org-1",
		Role:   membershipv1.Role_ROLE_ADMIN,
	})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if resp.Member.Role != membershipv1.Role_ROLE_ADMIN {
		t.Errorf("member role = %v, want %v", resp.Member.Role, membershipv1.Role_ROLE_ADMIN)
	}
	if len(auditLogger.events) != 1 {
		t.Errorf("audit events = %d, want 1", len(auditLogger.events))
	}
}

func TestUpdateRole_LastOwnerDemotionProtection(t *testing.T) {
	existing := &domain.Membership{
		ID:     "m1",
		UserID: "owner-1",
		OrgID:  "org-1",
		Role:   domain.RoleOwner,
	}
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
			"owner-1:org-1": existing,
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: map[string]int64{"org-1": 1},
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.UpdateRole(ctx, &membershipv1.UpdateRoleRequest{
		UserId: "owner-1",
		OrgId:  "org-1",
		Role:   membershipv1.Role_ROLE_MEMBER,
	})
	if err == nil {
		t.Fatal("expected error for demoting last owner")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("status code = %v, want %v", st.Code(), codes.FailedPrecondition)
	}
}

func TestUpdateRole_InvalidRole(t *testing.T) {
	existing := &domain.Membership{
		ID:     "m1",
		UserID: "user-2",
		OrgID:  "org-1",
		Role:   domain.RoleMember,
	}
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
			"user-2:org-1":  existing,
		},
		byID:        make(map[string]*domain.Membership),
		ownerCounts: make(map[string]int64),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.UpdateRole(ctx, &membershipv1.UpdateRoleRequest{
		UserId: "user-2",
		OrgId:  "org-1",
		Role:   membershipv1.Role_ROLE_UNSPECIFIED,
	})
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestListMembers_Success(t *testing.T) {
	now := time.Now().UTC()
	memberships := []*domain.Membership{
		{ID: "m1", UserID: "user-1", OrgID: "org-1", Role: domain.RoleOwner, CreatedAt: now},
		{ID: "m2", UserID: "user-2", OrgID: "org-1", Role: domain.RoleAdmin, CreatedAt: now},
		{ID: "m3", UserID: "user-3", OrgID: "org-1", Role: domain.RoleMember, CreatedAt: now},
	}
	membershipRepo := &mockMembershipRepo{
		memberships: map[string]*domain.Membership{
			"admin-1:org-1": {ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin},
			"user-1:org-1":  memberships[0],
			"user-2:org-1":  memberships[1],
			"user-3:org-1":  memberships[2],
		},
		byID: make(map[string]*domain.Membership),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	resp, err := srv.ListMembers(ctx, &membershipv1.ListMembersRequest{
		OrgId: "org-1",
	})
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(resp.Members) != 4 {
		t.Errorf("members count = %d, want 4", len(resp.Members))
	}
}

func TestListMembers_Pagination(t *testing.T) {
	now := time.Now().UTC()
	memberships := make([]*domain.Membership, 60)
	membershipMap := make(map[string]*domain.Membership)
	membershipMap["admin-1:org-1"] = &domain.Membership{ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin, CreatedAt: now}
	for i := 0; i < 60; i++ {
		id := strconv.Itoa(i)
		mem := &domain.Membership{
			ID:        "m" + id,
			UserID:    "user-" + id,
			OrgID:     "org-1",
			Role:      domain.RoleMember,
			CreatedAt: now,
		}
		memberships[i] = mem
		membershipMap["user-"+id+":org-1"] = mem
	}
	membershipRepo := &mockMembershipRepo{
		memberships: membershipMap,
		byID:        make(map[string]*domain.Membership),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	resp, err := srv.ListMembers(ctx, &membershipv1.ListMembersRequest{
		OrgId: "org-1",
		Pagination: &commonv1.Pagination{
			PageSize:  20,
			PageToken: "",
		},
	})
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(resp.Members) != 20 {
		t.Errorf("members count = %d, want 20", len(resp.Members))
	}
	if resp.Pagination.NextPageToken == "" {
		t.Error("expected next page token")
	}
}

func TestListMembers_MaxPageSize(t *testing.T) {
	now := time.Now().UTC()
	memberships := make([]*domain.Membership, 200)
	membershipMap := make(map[string]*domain.Membership)
	membershipMap["admin-1:org-1"] = &domain.Membership{ID: "m-admin", UserID: "admin-1", OrgID: "org-1", Role: domain.RoleAdmin, CreatedAt: now}
	for i := 0; i < 200; i++ {
		id := strconv.Itoa(i)
		mem := &domain.Membership{
			ID:        "m" + id,
			UserID:    "user-" + id,
			OrgID:     "org-1",
			Role:      domain.RoleMember,
			CreatedAt: now,
		}
		memberships[i] = mem
		membershipMap["user-"+id+":org-1"] = mem
	}
	membershipRepo := &mockMembershipRepo{
		memberships: membershipMap,
		byID:        make(map[string]*domain.Membership),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	resp, err := srv.ListMembers(ctx, &membershipv1.ListMembersRequest{
		OrgId: "org-1",
		Pagination: &commonv1.Pagination{
			PageSize:  150, // exceeds maxPageSize
			PageToken: "",
		},
	})
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(resp.Members) > maxPageSize {
		t.Errorf("members count = %d, want <= %d", len(resp.Members), maxPageSize)
	}
}

func TestListMembers_NonAdminCaller(t *testing.T) {
	membershipRepo := &mockMembershipRepo{
		memberships: make(map[string]*domain.Membership),
		byID:        make(map[string]*domain.Membership),
	}
	srv := NewServer(membershipRepo, nil, nil)
	ctx := ctxWithMember("org-1", "member-1")

	_, err := srv.ListMembers(ctx, &membershipv1.ListMembersRequest{
		OrgId: "org-1",
	})
	if err == nil {
		t.Fatal("expected error for non-admin caller")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestListMembers_NilRepo(t *testing.T) {
	srv := NewServer(nil, nil, nil)
	ctx := ctxWithAdmin("org-1", "admin-1")

	_, err := srv.ListMembers(ctx, &membershipv1.ListMembersRequest{
		OrgId: "org-1",
	})
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

func TestProtoRoleToDomain(t *testing.T) {
	testCases := []struct {
		input    membershipv1.Role
		expected domain.Role
	}{
		{membershipv1.Role_ROLE_OWNER, domain.RoleOwner},
		{membershipv1.Role_ROLE_ADMIN, domain.RoleAdmin},
		{membershipv1.Role_ROLE_MEMBER, domain.RoleMember},
		{membershipv1.Role_ROLE_UNSPECIFIED, ""},
		{999, ""}, // Invalid enum value
	}

	for _, tc := range testCases {
		result := protoRoleToDomain(tc.input)
		if result != tc.expected {
			t.Errorf("protoRoleToDomain(%v) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestDomainRoleToProto(t *testing.T) {
	testCases := []struct {
		input    domain.Role
		expected membershipv1.Role
	}{
		{domain.RoleOwner, membershipv1.Role_ROLE_OWNER},
		{domain.RoleAdmin, membershipv1.Role_ROLE_ADMIN},
		{domain.RoleMember, membershipv1.Role_ROLE_MEMBER},
		{"", membershipv1.Role_ROLE_UNSPECIFIED},
		{"invalid", membershipv1.Role_ROLE_UNSPECIFIED},
	}

	for _, tc := range testCases {
		result := domainRoleToProto(tc.input)
		if result != tc.expected {
			t.Errorf("domainRoleToProto(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}
