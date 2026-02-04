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

// mockMembershipGetterForMember implements OrgMembershipGetter for RequireOrgMember tests.
type mockMembershipGetterForMember struct {
	memberships map[string]*domain.Membership
	err         error
}

func (m *mockMembershipGetterForMember) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*domain.Membership, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := userID + ":" + orgID
	return m.memberships[key], nil
}

func TestRequireOrgMember_Success_AnyRole(t *testing.T) {
	testCases := []struct {
		name string
		role domain.Role
	}{
		{"owner", domain.RoleOwner},
		{"admin", domain.RoleAdmin},
		{"member", domain.RoleMember},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getter := &mockMembershipGetterForMember{
				memberships: map[string]*domain.Membership{
					"user-1:org-1": {ID: "m1", UserID: "user-1", OrgID: "org-1", Role: tc.role},
				},
			}
			ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

			orgID, userID, err := RequireOrgMember(ctx, getter)
			if err != nil {
				t.Fatalf("RequireOrgMember: %v", err)
			}
			if orgID != "org-1" {
				t.Errorf("org_id = %q, want %q", orgID, "org-1")
			}
			if userID != "user-1" {
				t.Errorf("user_id = %q, want %q", userID, "user-1")
			}
		})
	}
}

func TestRequireOrgMember_Failure_NotMember(t *testing.T) {
	getter := &mockMembershipGetterForMember{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

	_, _, err := RequireOrgMember(ctx, getter)
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

func TestRequireOrgMember_Failure_NoContext(t *testing.T) {
	getter := &mockMembershipGetterForMember{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := context.Background()

	_, _, err := RequireOrgMember(ctx, getter)
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

func TestRequireOrgMember_Failure_RepositoryError(t *testing.T) {
	getter := &mockMembershipGetterForMember{
		memberships: make(map[string]*domain.Membership),
		err:         errors.New("database error"),
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "org-1", "session-1")

	_, _, err := RequireOrgMember(ctx, getter)
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

func TestRequireOrgMember_Failure_EmptyOrgID(t *testing.T) {
	getter := &mockMembershipGetterForMember{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := interceptors.WithIdentity(context.Background(), "user-1", "", "session-1")

	_, _, err := RequireOrgMember(ctx, getter)
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

func TestRequireOrgMember_Failure_EmptyUserID(t *testing.T) {
	getter := &mockMembershipGetterForMember{
		memberships: make(map[string]*domain.Membership),
	}
	ctx := interceptors.WithIdentity(context.Background(), "", "org-1", "session-1")

	_, _, err := RequireOrgMember(ctx, getter)
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
