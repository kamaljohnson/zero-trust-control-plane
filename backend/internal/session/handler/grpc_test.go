package handler

import (
	"context"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "zero-trust-control-plane/backend/api/generated/common/v1"
	sessionv1 "zero-trust-control-plane/backend/api/generated/session/v1"
	membershipdomain "zero-trust-control-plane/backend/internal/membership/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
	sessiondomain "zero-trust-control-plane/backend/internal/session/domain"
)

// mockSessionRepo implements sessionrepo.Repository for tests.
type mockSessionRepo struct {
	sessions   map[string]*sessiondomain.Session
	listByOrg  map[string][]*sessiondomain.Session
	getByIDErr error
	listErr    error
	revokeErr  error
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id string) (*sessiondomain.Session, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.sessions[id], nil
}

func (m *mockSessionRepo) ListByUserAndOrg(ctx context.Context, userID, orgID string) ([]*sessiondomain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) ListByOrg(ctx context.Context, orgID string, userID *string, limit, offset int32) ([]*sessiondomain.Session, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	all := m.listByOrg[orgID]
	if userID != nil {
		filtered := make([]*sessiondomain.Session, 0)
		for _, s := range all {
			if s.UserID == *userID {
				filtered = append(filtered, s)
			}
		}
		all = filtered
	}
	start := int(offset)
	if start > len(all) {
		start = len(all)
	}
	end := start + int(limit)
	if end > len(all) {
		end = len(all)
	}
	if start >= len(all) {
		return []*sessiondomain.Session{}, nil
	}
	return all[start:end], nil
}

func (m *mockSessionRepo) Create(ctx context.Context, s *sessiondomain.Session) error {
	return nil
}

func (m *mockSessionRepo) Revoke(ctx context.Context, id string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	return nil
}

func (m *mockSessionRepo) RevokeAllSessionsByUser(ctx context.Context, userID string) error {
	return nil
}

func (m *mockSessionRepo) RevokeAllSessionsByUserAndOrg(ctx context.Context, userID, orgID string) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	return nil
}

func (m *mockSessionRepo) UpdateLastSeen(ctx context.Context, id string, at time.Time) error {
	return nil
}

func (m *mockSessionRepo) UpdateRefreshToken(ctx context.Context, sessionID, jti, refreshTokenHash string) error {
	return nil
}

// mockMembershipRepoForSession implements membershiprepo.Repository for session handler tests.
type mockMembershipRepoForSession struct {
	memberships map[string]*membershipdomain.Membership
}

func (m *mockMembershipRepoForSession) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*membershipdomain.Membership, error) {
	key := userID + ":" + orgID
	return m.memberships[key], nil
}

func (m *mockMembershipRepoForSession) GetMembershipByID(ctx context.Context, id string) (*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepoForSession) ListMembershipsByOrg(ctx context.Context, orgID string) ([]*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepoForSession) CreateMembership(ctx context.Context, mem *membershipdomain.Membership) error {
	return nil
}

func (m *mockMembershipRepoForSession) DeleteByUserAndOrg(ctx context.Context, userID, orgID string) error {
	return nil
}

func (m *mockMembershipRepoForSession) UpdateRole(ctx context.Context, userID, orgID string, role membershipdomain.Role) (*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepoForSession) CountOwnersByOrg(ctx context.Context, orgID string) (int64, error) {
	return 0, nil
}

// mockAuditLoggerForSession implements audit.AuditLogger for session handler tests.
type mockAuditLoggerForSession struct {
	events []struct {
		orgID, userID, action, resource, resourceID string
	}
}

func (m *mockAuditLoggerForSession) LogEvent(ctx context.Context, orgID, userID, action, resource, resourceID string) {
	m.events = append(m.events, struct {
		orgID, userID, action, resource, resourceID string
	}{orgID, userID, action, resource, resourceID})
}

func ctxWithAdminForSession(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func ctxWithMemberForSession(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func TestRevokeSession_Success(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	sessionRepo := &mockSessionRepo{
		sessions:  map[string]*sessiondomain.Session{"session-1": session},
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	auditLogger := &mockAuditLoggerForSession{}
	srv := NewServer(sessionRepo, membershipRepo, auditLogger)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeSession(ctx, &sessionv1.RevokeSessionRequest{SessionId: "session-1"})
	if err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	if len(auditLogger.events) != 1 {
		t.Errorf("audit events = %d, want 1", len(auditLogger.events))
	}
}

func TestRevokeSession_NotFound(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeSession(ctx, &sessionv1.RevokeSessionRequest{SessionId: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestRevokeSession_WrongOrg(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-2",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	sessionRepo := &mockSessionRepo{
		sessions:  map[string]*sessiondomain.Session{"session-1": session},
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeSession(ctx, &sessionv1.RevokeSessionRequest{SessionId: "session-1"})
	if err == nil {
		t.Fatal("expected error for wrong org")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestRevokeSession_NonAdminCaller(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithMemberForSession("org-1", "member-1")

	_, err := srv.RevokeSession(ctx, &sessionv1.RevokeSessionRequest{SessionId: "session-1"})
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

func TestRevokeSession_InvalidSessionID(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeSession(ctx, &sessionv1.RevokeSessionRequest{SessionId: ""})
	if err == nil {
		t.Fatal("expected error for empty session_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestRevokeSession_NilRepo(t *testing.T) {
	srv := NewServer(nil, nil, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeSession(ctx, &sessionv1.RevokeSessionRequest{SessionId: "session-1"})
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

func TestListSessions_Success(t *testing.T) {
	now := time.Now().UTC()
	sessions := []*sessiondomain.Session{
		{ID: "session-1", UserID: "user-1", OrgID: "org-1", DeviceID: "device-1", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now},
		{ID: "session-2", UserID: "user-2", OrgID: "org-1", DeviceID: "device-2", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now},
	}
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: map[string][]*sessiondomain.Session{"org-1": sessions},
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	resp, err := srv.ListSessions(ctx, &sessionv1.ListSessionsRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Sessions) != 2 {
		t.Errorf("sessions count = %d, want 2", len(resp.Sessions))
	}
}

func TestListSessions_FilteredByUserID(t *testing.T) {
	now := time.Now().UTC()
	sessions := []*sessiondomain.Session{
		{ID: "session-1", UserID: "user-1", OrgID: "org-1", DeviceID: "device-1", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now},
		{ID: "session-2", UserID: "user-2", OrgID: "org-1", DeviceID: "device-2", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now},
		{ID: "session-3", UserID: "user-1", OrgID: "org-1", DeviceID: "device-3", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now},
	}
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: map[string][]*sessiondomain.Session{"org-1": sessions},
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	resp, err := srv.ListSessions(ctx, &sessionv1.ListSessionsRequest{
		OrgId:  "org-1",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Sessions) != 2 {
		t.Errorf("sessions count = %d, want 2", len(resp.Sessions))
	}
}

func TestListSessions_Pagination(t *testing.T) {
	now := time.Now().UTC()
	sessions := make([]*sessiondomain.Session, 60)
	for i := 0; i < 60; i++ {
		sessions[i] = &sessiondomain.Session{
			ID:        "session-" + strconv.Itoa(i),
			UserID:    "user-1",
			OrgID:     "org-1",
			DeviceID:  "device-1",
			ExpiresAt: now.Add(24 * time.Hour),
			CreatedAt: now,
		}
	}
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: map[string][]*sessiondomain.Session{"org-1": sessions},
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	resp, err := srv.ListSessions(ctx, &sessionv1.ListSessionsRequest{
		OrgId: "org-1",
		Pagination: &commonv1.Pagination{
			PageSize:  20,
			PageToken: "",
		},
	})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Sessions) != 20 {
		t.Errorf("sessions count = %d, want 20", len(resp.Sessions))
	}
	if resp.Pagination.NextPageToken == "" {
		t.Error("expected next page token")
	}
}

func TestListSessions_RepositoryError(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
		listErr:   status.Error(codes.Internal, "database error"),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.ListSessions(ctx, &sessionv1.ListSessionsRequest{OrgId: "org-1"})
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

func TestListSessions_EmptyResults(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: map[string][]*sessiondomain.Session{"org-1": {}},
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	resp, err := srv.ListSessions(ctx, &sessionv1.ListSessionsRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Sessions) != 0 {
		t.Errorf("sessions count = %d, want 0", len(resp.Sessions))
	}
}

func TestListSessions_OffsetBeyondResults(t *testing.T) {
	now := time.Now().UTC()
	sessions := []*sessiondomain.Session{
		{ID: "session-1", UserID: "user-1", OrgID: "org-1", DeviceID: "device-1", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now},
	}
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: map[string][]*sessiondomain.Session{"org-1": sessions},
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	resp, err := srv.ListSessions(ctx, &sessionv1.ListSessionsRequest{
		OrgId: "org-1",
		Pagination: &commonv1.Pagination{
			PageSize:  10,
			PageToken: "100", // Beyond available results (offset 100)
		},
	})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Sessions) != 0 {
		t.Errorf("sessions count = %d, want 0 when offset beyond results", len(resp.Sessions))
	}
}

func TestRevokeSession_RepositoryError(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	sessionRepo := &mockSessionRepo{
		sessions:  map[string]*sessiondomain.Session{"session-1": session},
		listByOrg: make(map[string][]*sessiondomain.Session),
		revokeErr: status.Error(codes.Internal, "database error"),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeSession(ctx, &sessionv1.RevokeSessionRequest{SessionId: "session-1"})
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

func TestListSessions_NonAdminCaller(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithMemberForSession("org-1", "member-1")

	_, err := srv.ListSessions(ctx, &sessionv1.ListSessionsRequest{OrgId: "org-1"})
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

func TestGetSession_Success(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	sessionRepo := &mockSessionRepo{
		sessions:  map[string]*sessiondomain.Session{"session-1": session},
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	resp, err := srv.GetSession(ctx, &sessionv1.GetSessionRequest{SessionId: "session-1"})
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if resp.Session.Id != "session-1" {
		t.Errorf("session id = %q, want %q", resp.Session.Id, "session-1")
	}
}

func TestGetSession_WrongOrg(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-2",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	sessionRepo := &mockSessionRepo{
		sessions:  map[string]*sessiondomain.Session{"session-1": session},
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.GetSession(ctx, &sessionv1.GetSessionRequest{SessionId: "session-1"})
	if err == nil {
		t.Fatal("expected error for wrong org")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestRevokeAllSessionsForUser_Success(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	auditLogger := &mockAuditLoggerForSession{}
	srv := NewServer(sessionRepo, membershipRepo, auditLogger)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeAllSessionsForUser(ctx, &sessionv1.RevokeAllSessionsForUserRequest{
		UserId: "user-1",
		OrgId:  "org-1",
	})
	if err != nil {
		t.Fatalf("RevokeAllSessionsForUser: %v", err)
	}
	if len(auditLogger.events) != 1 {
		t.Errorf("audit events = %d, want 1", len(auditLogger.events))
	}
}

func TestRevokeAllSessionsForUser_InvalidUserID(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeAllSessionsForUser(ctx, &sessionv1.RevokeAllSessionsForUserRequest{
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

// Additional tests for GetSession, RevokeAllSessionsForUser, and domainSessionToProto

func TestGetSession_NotFound(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.GetSession(ctx, &sessionv1.GetSessionRequest{SessionId: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestGetSession_RepositoryError(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
		getByIDErr: status.Error(codes.Internal, "database error"),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.GetSession(ctx, &sessionv1.GetSessionRequest{SessionId: "session-1"})
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

func TestRevokeAllSessionsForUser_RepositoryError(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		sessions:  make(map[string]*sessiondomain.Session),
		listByOrg: make(map[string][]*sessiondomain.Session),
		revokeErr: status.Error(codes.Internal, "database error"),
	}
	membershipRepo := &mockMembershipRepoForSession{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(sessionRepo, membershipRepo, nil)
	ctx := ctxWithAdminForSession("org-1", "admin-1")

	_, err := srv.RevokeAllSessionsForUser(ctx, &sessionv1.RevokeAllSessionsForUserRequest{
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

func TestDomainSessionToProto_WithRevokedAt(t *testing.T) {
	now := time.Now().UTC()
	revokedAt := now.Add(1 * time.Hour)
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: &revokedAt,
		CreatedAt: now,
	}

	proto := domainSessionToProto(session)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.Id != "session-1" {
		t.Errorf("id = %q, want %q", proto.Id, "session-1")
	}
	if proto.RevokedAt == nil {
		t.Error("revoked_at should be set")
	}
	if !proto.RevokedAt.AsTime().Equal(revokedAt) {
		t.Errorf("revoked_at = %v, want %v", proto.RevokedAt.AsTime(), revokedAt)
	}
}

func TestDomainSessionToProto_WithoutRevokedAt(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: nil, // Not revoked
		CreatedAt: now,
	}

	proto := domainSessionToProto(session)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.RevokedAt != nil {
		t.Error("revoked_at should be nil for non-revoked session")
	}
}

func TestDomainSessionToProto_WithLastSeenAt(t *testing.T) {
	now := time.Now().UTC()
	lastSeenAt := now.Add(30 * time.Minute)
	session := &sessiondomain.Session{
		ID:         "session-1",
		UserID:     "user-1",
		OrgID:      "org-1",
		DeviceID:   "device-1",
		ExpiresAt:  now.Add(24 * time.Hour),
		LastSeenAt: &lastSeenAt,
		CreatedAt:  now,
	}

	proto := domainSessionToProto(session)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.LastSeenAt == nil {
		t.Error("last_seen_at should be set")
	}
	if !proto.LastSeenAt.AsTime().Equal(lastSeenAt) {
		t.Errorf("last_seen_at = %v, want %v", proto.LastSeenAt.AsTime(), lastSeenAt)
	}
}

func TestDomainSessionToProto_WithoutLastSeenAt(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:         "session-1",
		UserID:     "user-1",
		OrgID:      "org-1",
		DeviceID:   "device-1",
		ExpiresAt:  now.Add(24 * time.Hour),
		LastSeenAt: nil,
		CreatedAt:  now,
	}

	proto := domainSessionToProto(session)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.LastSeenAt != nil {
		t.Error("last_seen_at should be nil when not set")
	}
}

func TestDomainSessionToProto_NilSession(t *testing.T) {
	proto := domainSessionToProto(nil)
	if proto != nil {
		t.Error("proto should be nil for nil session")
	}
}

func TestDomainSessionToProto_WithIPAddress(t *testing.T) {
	now := time.Now().UTC()
	session := &sessiondomain.Session{
		ID:        "session-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		DeviceID:  "device-1",
		ExpiresAt: now.Add(24 * time.Hour),
		IPAddress: "192.168.1.1",
		CreatedAt: now,
	}

	proto := domainSessionToProto(session)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.IpAddress != "192.168.1.1" {
		t.Errorf("ip_address = %q, want %q", proto.IpAddress, "192.168.1.1")
	}
}
