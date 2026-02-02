package interceptors

import "context"

type contextKey struct{ name string }

var (
	userIDKey    = contextKey{"user_id"}
	orgIDKey     = contextKey{"org_id"}
	sessionIDKey = contextKey{"session_id"}
)

// WithIdentity returns a context with user_id, org_id, and session_id set.
// Handlers and the auth service can read these via GetUserID, GetOrgID, GetSessionID.
func WithIdentity(ctx context.Context, userID, orgID, sessionID string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, orgIDKey, orgID)
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)
	return ctx
}

// GetUserID returns the user_id from context and true if set; otherwise "", false.
func GetUserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

// GetOrgID returns the org_id from context and true if set; otherwise "", false.
func GetOrgID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(orgIDKey).(string)
	return v, ok
}

// GetSessionID returns the session_id from context and true if set; otherwise "", false.
func GetSessionID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionIDKey).(string)
	return v, ok
}
