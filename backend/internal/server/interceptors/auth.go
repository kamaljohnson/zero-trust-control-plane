package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"zero-trust-control-plane/backend/internal/security"
)

const bearerPrefix = "bearer "

// SessionValidator returns true if the session is active (exists and not revoked).
// When non-nil, AuthUnary calls it after ValidateAccess; if it returns false or an error, the request is rejected with Unauthenticated.
type SessionValidator func(ctx context.Context, sessionID string) (active bool, err error)

// AuthUnary returns a unary server interceptor that validates the Bearer (access) token
// from gRPC metadata and sets user_id, org_id, session_id in context for protected RPCs.
// publicMethods is the set of full method names that do not require a Bearer token
// (e.g. AuthService Register, Login, Refresh; HealthService HealthCheck).
// If sessionValidator is non-nil, it is called after token validation; revoked or missing sessions are rejected with Unauthenticated.
func AuthUnary(tokens *security.TokenProvider, publicMethods map[string]bool, sessionValidator SessionValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token := extractBearer(ctx)
		public := publicMethods[info.FullMethod]

		if token == "" {
			if public {
				return handler(ctx, req)
			}
			return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization")
		}

		sessionID, userID, orgID, err := tokens.ValidateAccess(token)
		if err != nil {
			if public {
				return handler(ctx, req)
			}
			return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization")
		}

		if sessionValidator != nil {
			active, err := sessionValidator(ctx, sessionID)
			if err != nil || !active {
				return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization")
			}
		}

		ctx = WithIdentity(ctx, userID, orgID, sessionID)
		return handler(ctx, req)
	}
}

// extractBearer returns the Bearer token from ctx metadata, or "" if missing or malformed.
func extractBearer(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return ""
	}
	v := strings.TrimSpace(vals[0])
	if len(v) < len(bearerPrefix) {
		return ""
	}
	if !strings.EqualFold(v[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(v[len(bearerPrefix):])
}
