package rbac

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zero-trust-control-plane/backend/internal/membership/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
)

// mockMembershipGetter implements OrgMembershipGetter for tests.
type mockMembershipGetter struct {
	memberships map[string]*domain.Membership
	err         error
}

func (m *mockMembershipGetter) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*domain.Membership, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := userID + ":" + orgID
	return m.memberships[key], nil
}

func TestRequireOrgAdmin_Success_Owner(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: map[string]*domain.Membership{
			"user-1:org-1": {ID: "m1", UserID: "user-1", OrgID: "org-1", Role: domain.RoleOwner},
		},
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

	orgID, userID, err := RequireOrgAdmin(ctx, getter)
	if err != nil {
		t.Fatalf("RequireOrgAdmin: %v", err)
	}
	if orgID != "org-1" {
		t.Errorf("org_id = %q, want %q", orgID, "org-1")
	}
	if userID != "user-1" {
		t.Errorf("user_id = %q, want %q", userID, "user-1")
	}
}

func TestRequireOrgAdmin_Success_Admin(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: map[string]*domain.Membership{
			"user-1:org-1": {ID: "m1", UserID: "user-1", OrgID: "org-1", Role: domain.RoleAdmin},
		},
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

	orgID, userID, err := RequireOrgAdmin(ctx, getter)
	if err != nil {
		t.Fatalf("RequireOrgAdmin: %v", err)
	}
	if orgID != "org-1" {
		t.Errorf("org_id = %q, want %q", orgID, "org-1")
	}
	if userID != "user-1" {
		t.Errorf("user_id = %q, want %q", userID, "user-1")
	}
}

func TestRequireOrgAdmin_Failure_Member(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: map[string]*domain.Membership{
			"user-1:org-1": {ID: "m1", UserID: "user-1", OrgID: "org-1", Role: domain.RoleMember},
		},
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

	_, _, err := RequireOrgAdmin(ctx, getter)
	if err == nil {
		t.Fatal("expected error for member role")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestRequireOrgAdmin_Failure_NotMember(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

	_, _, err := RequireOrgAdmin(ctx, getter)
	if err == nil {
		t.Fatal("expected error for non-member")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestRequireOrgAdmin_Failure_NoContext(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := context.Background()

	_, _, err := RequireOrgAdmin(ctx, getter)
	if err == nil {
		t.Fatal("expected error for missing context")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestRequireOrgAdmin_Failure_RepositoryError(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: make(map[string]*domain.Membership),
		err:         errors.New("database error"),
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

	_, _, err := RequireOrgAdmin(ctx, getter)
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

func TestRequireOrgAdmin_Failure_EmptyOrgID(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "", "session-1")

	_, _, err := RequireOrgAdmin(ctx, getter)
	if err == nil {
		t.Fatal("expected error for empty org_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestRequireOrgAdmin_Failure_EmptyUserID(t *testing.T) {
	getter := &mockMembershipGetter{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := interceptors.WithIdentity(context.Background(), "", "org-1", "session-1")

	_, _, err := RequireOrgAdmin(ctx, getter)
	if err == nil {
		t.Fatal("expected error for empty user_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}
