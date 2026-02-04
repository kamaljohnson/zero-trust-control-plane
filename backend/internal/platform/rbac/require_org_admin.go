package rbac

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zero-trust-control-plane/backend/internal/membership/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
)

// OrgMembershipGetter returns a user's membership in an org. Used by RequireOrgAdmin to resolve caller role.
type OrgMembershipGetter interface {
	GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*domain.Membership, error)
}

// RequireOrgAdmin ensures the caller is authenticated and has role owner or admin in the context org.
// Returns (orgID, userID, nil) on success; returns a gRPC error (Unauthenticated or PermissionDenied) on failure.
func RequireOrgAdmin(ctx context.Context, getter OrgMembershipGetter) (orgID, userID string, err error) {
	orgID, okOrg := interceptors.GetOrgID(ctx)
	userID, okUser := interceptors.GetUserID(ctx)
	if !okOrg || orgID == "" || !okUser || userID == "" {
		return "", "", status.Error(codes.Unauthenticated, "org and user context required")
	}
	m, err := getter.GetMembershipByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return "", "", status.Error(codes.Internal, "failed to resolve membership")
	}
	if m == nil {
		return "", "", status.Error(codes.PermissionDenied, "not a member of this organization")
	}
	if m.Role != domain.RoleOwner && m.Role != domain.RoleAdmin {
		return "", "", status.Error(codes.PermissionDenied, "organization admin or owner required")
	}
	return orgID, userID, nil
}
