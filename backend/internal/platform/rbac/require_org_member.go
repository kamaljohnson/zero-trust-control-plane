package rbac

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zero-trust-control-plane/backend/internal/server/interceptors"
)

// RequireOrgMember ensures the caller is authenticated and is a member of the context org (any role).
// Returns (orgID, userID, nil) on success; returns a gRPC error (Unauthenticated or PermissionDenied) on failure.
func RequireOrgMember(ctx context.Context, getter OrgMembershipGetter) (orgID, userID string, err error) {
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
	return orgID, userID, nil
}
