package handler

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	auditv1 "zero-trust-control-plane/backend/api/generated/audit/v1"
	commonv1 "zero-trust-control-plane/backend/api/generated/common/v1"
	auditdomain "zero-trust-control-plane/backend/internal/audit/domain"
	membershipdomain "zero-trust-control-plane/backend/internal/membership/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
)

// mockAuditRepo implements Repository for tests.
type mockAuditRepo struct {
	logs    map[string][]*auditdomain.AuditLog
	listErr error
}

func (m *mockAuditRepo) ListByOrgFiltered(ctx context.Context, orgID string, limit, offset int32, userID, action, resource *string) ([]*auditdomain.AuditLog, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	all := m.logs[orgID]
	if all == nil {
		return []*auditdomain.AuditLog{}, nil
	}
	var filtered []*auditdomain.AuditLog
	for _, log := range all {
		if userID != nil && log.UserID != *userID {
			continue
		}
		if action != nil && log.Action != *action {
			continue
		}
		if resource != nil && log.Resource != *resource {
			continue
		}
		filtered = append(filtered, log)
	}
	start := int(offset)
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + int(limit)
	if end > len(filtered) {
		end = len(filtered)
	}
	if start >= len(filtered) {
		return []*auditdomain.AuditLog{}, nil
	}
	return filtered[start:end], nil
}

// mockMembershipRepoForAudit implements rbac.OrgMembershipGetter for audit handler tests.
type mockMembershipRepoForAudit struct {
	memberships map[string]*membershipdomain.Membership
}

func (m *mockMembershipRepoForAudit) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*membershipdomain.Membership, error) {
	key := userID + ":" + orgID
	return m.memberships[key], nil
}

func ctxWithAdminForAudit(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func ctxWithMemberForAudit(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func TestListAuditLogs_Success(t *testing.T) {
	now := time.Now().UTC()
	logs := []*auditdomain.AuditLog{
		{ID: "log-1", OrgID: "org-1", UserID: "user-1", Action: "create", Resource: "policy", IP: "1.2.3.4", CreatedAt: now},
		{ID: "log-2", OrgID: "org-1", UserID: "user-2", Action: "update", Resource: "policy", IP: "1.2.3.5", CreatedAt: now},
	}
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": logs},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	resp, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(resp.Logs) != 2 {
		t.Errorf("logs count = %d, want 2", len(resp.Logs))
	}
}

func TestListAuditLogs_FilterByUserID(t *testing.T) {
	now := time.Now().UTC()
	logs := []*auditdomain.AuditLog{
		{ID: "log-1", OrgID: "org-1", UserID: "user-1", Action: "create", Resource: "policy", IP: "1.2.3.4", CreatedAt: now},
		{ID: "log-2", OrgID: "org-1", UserID: "user-2", Action: "update", Resource: "policy", IP: "1.2.3.5", CreatedAt: now},
		{ID: "log-3", OrgID: "org-1", UserID: "user-1", Action: "delete", Resource: "policy", IP: "1.2.3.6", CreatedAt: now},
	}
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": logs},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	resp, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{
		OrgId:  "org-1",
		UserId: "user-1",
	})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(resp.Logs) != 2 {
		t.Errorf("logs count = %d, want 2", len(resp.Logs))
	}
	for _, log := range resp.Logs {
		if log.UserId != "user-1" {
			t.Errorf("log user_id = %q, want %q", log.UserId, "user-1")
		}
	}
}

func TestListAuditLogs_FilterByAction(t *testing.T) {
	now := time.Now().UTC()
	logs := []*auditdomain.AuditLog{
		{ID: "log-1", OrgID: "org-1", UserID: "user-1", Action: "create", Resource: "policy", IP: "1.2.3.4", CreatedAt: now},
		{ID: "log-2", OrgID: "org-1", UserID: "user-2", Action: "update", Resource: "policy", IP: "1.2.3.5", CreatedAt: now},
		{ID: "log-3", OrgID: "org-1", UserID: "user-1", Action: "create", Resource: "user", IP: "1.2.3.6", CreatedAt: now},
	}
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": logs},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	resp, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{
		OrgId:  "org-1",
		Action: "create",
	})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(resp.Logs) != 2 {
		t.Errorf("logs count = %d, want 2", len(resp.Logs))
	}
	for _, log := range resp.Logs {
		if log.Action != "create" {
			t.Errorf("log action = %q, want %q", log.Action, "create")
		}
	}
}

func TestListAuditLogs_FilterByResource(t *testing.T) {
	now := time.Now().UTC()
	logs := []*auditdomain.AuditLog{
		{ID: "log-1", OrgID: "org-1", UserID: "user-1", Action: "create", Resource: "policy", IP: "1.2.3.4", CreatedAt: now},
		{ID: "log-2", OrgID: "org-1", UserID: "user-2", Action: "update", Resource: "user", IP: "1.2.3.5", CreatedAt: now},
		{ID: "log-3", OrgID: "org-1", UserID: "user-1", Action: "delete", Resource: "policy", IP: "1.2.3.6", CreatedAt: now},
	}
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": logs},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	resp, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{
		OrgId:    "org-1",
		Resource: "policy",
	})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(resp.Logs) != 2 {
		t.Errorf("logs count = %d, want 2", len(resp.Logs))
	}
	for _, log := range resp.Logs {
		if log.Resource != "policy" {
			t.Errorf("log resource = %q, want %q", log.Resource, "policy")
		}
	}
}

func TestListAuditLogs_Pagination(t *testing.T) {
	now := time.Now().UTC()
	logs := make([]*auditdomain.AuditLog, 60)
	for i := 0; i < 60; i++ {
		logs[i] = &auditdomain.AuditLog{
			ID:        "log-" + strconv.Itoa(i),
			OrgID:     "org-1",
			UserID:    "user-1",
			Action:    "create",
			Resource:  "policy",
			IP:        "1.2.3.4",
			CreatedAt: now,
		}
	}
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": logs},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	resp, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{
		OrgId: "org-1",
		Pagination: &commonv1.Pagination{
			PageSize:  20,
			PageToken: "",
		},
	})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(resp.Logs) != 20 {
		t.Errorf("logs count = %d, want 20", len(resp.Logs))
	}
	if resp.Pagination.NextPageToken == "" {
		t.Error("expected next page token")
	}
}

func TestListAuditLogs_MaxPageSize(t *testing.T) {
	now := time.Now().UTC()
	logs := make([]*auditdomain.AuditLog, 200)
	for i := 0; i < 200; i++ {
		logs[i] = &auditdomain.AuditLog{
			ID:        "log-" + strconv.Itoa(i),
			OrgID:     "org-1",
			UserID:    "user-1",
			Action:    "create",
			Resource:  "policy",
			IP:        "1.2.3.4",
			CreatedAt: now,
		}
	}
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": logs},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	resp, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{
		OrgId: "org-1",
		Pagination: &commonv1.Pagination{
			PageSize:  150, // exceeds maxPageSize
			PageToken: "",
		},
	})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(resp.Logs) > maxPageSize {
		t.Errorf("logs count = %d, want <= %d", len(resp.Logs), maxPageSize)
	}
}

func TestListAuditLogs_NonAdminCaller(t *testing.T) {
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": {}},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithMemberForAudit("org-1", "member-1")

	_, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{OrgId: "org-1"})
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

func TestListAuditLogs_OrgIDMismatch(t *testing.T) {
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": {}},
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	_, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{OrgId: "org-2"})
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

func TestListAuditLogs_RepositoryError(t *testing.T) {
	repo := &mockAuditRepo{
		logs:    map[string][]*auditdomain.AuditLog{"org-1": {}},
		listErr: errors.New("database error"),
	}
	membershipRepo := &mockMembershipRepoForAudit{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	_, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{OrgId: "org-1"})
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

func TestListAuditLogs_NilRepo(t *testing.T) {
	srv := NewServer(nil, nil)
	ctx := ctxWithAdminForAudit("org-1", "admin-1")

	_, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{OrgId: "org-1"})
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

func TestListAuditLogs_NoOrgAdminChecker(t *testing.T) {
	now := time.Now().UTC()
	logs := []*auditdomain.AuditLog{
		{ID: "log-1", OrgID: "org-1", UserID: "user-1", Action: "create", Resource: "policy", IP: "1.2.3.4", CreatedAt: now},
	}
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": logs},
	}
	srv := NewServer(repo, nil)
	ctx := ctxWithAdminForAudit("org-1", "user-1")

	resp, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(resp.Logs) != 1 {
		t.Errorf("logs count = %d, want 1", len(resp.Logs))
	}
}

func TestListAuditLogs_NoOrgContext(t *testing.T) {
	repo := &mockAuditRepo{
		logs: map[string][]*auditdomain.AuditLog{"org-1": {}},
	}
	srv := NewServer(repo, nil)
	ctx := context.Background()

	_, err := srv.ListAuditLogs(ctx, &auditv1.ListAuditLogsRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for missing org context")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}
