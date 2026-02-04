package interceptors

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"zero-trust-control-plane/backend/internal/security"
)

func TestAuthUnary_PublicMethod(t *testing.T) {
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	publicMethods := map[string]bool{
		"/test.Service/PublicMethod": true,
	}
	interceptor := AuthUnary(tokens, publicMethods, nil)

	ctx := context.Background()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/PublicMethod",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
}

func TestAuthUnary_ProtectedMethod_NoToken(t *testing.T) {
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	publicMethods := map[string]bool{}
	interceptor := AuthUnary(tokens, publicMethods, nil)

	ctx := context.Background()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	_, err = interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/ProtectedMethod",
	}, handler)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthUnary_ProtectedMethod_ValidToken(t *testing.T) {
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	token, _, _, err := tokens.IssueAccess("session-1", "user-1", "org-1")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	publicMethods := map[string]bool{}
	interceptor := AuthUnary(tokens, publicMethods, nil)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	}))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		userID, ok := GetUserID(ctx)
		if !ok || userID != "user-1" {
			t.Errorf("user_id = %q, ok = %v, want %q", userID, ok, "user-1")
		}
		orgID, ok := GetOrgID(ctx)
		if !ok || orgID != "org-1" {
			t.Errorf("org_id = %q, ok = %v, want %q", orgID, ok, "org-1")
		}
		sessionID, ok := GetSessionID(ctx)
		if !ok || sessionID != "session-1" {
			t.Errorf("session_id = %q, ok = %v, want %q", sessionID, ok, "session-1")
		}
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/ProtectedMethod",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
}

func TestAuthUnary_ProtectedMethod_InvalidToken(t *testing.T) {
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	publicMethods := map[string]bool{}
	interceptor := AuthUnary(tokens, publicMethods, nil)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Bearer invalid-token",
	}))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	_, err = interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/ProtectedMethod",
	}, handler)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthUnary_SessionValidator_ValidSession(t *testing.T) {
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	token, _, _, err := tokens.IssueAccess("session-1", "user-1", "org-1")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	publicMethods := map[string]bool{}
	sessionValidator := func(ctx context.Context, sessionID string) (bool, error) {
		return sessionID == "session-1", nil
	}
	interceptor := AuthUnary(tokens, publicMethods, sessionValidator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	}))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/ProtectedMethod",
	}, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "success" {
		t.Errorf("response = %v, want %q", resp, "success")
	}
}

func TestAuthUnary_SessionValidator_RevokedSession(t *testing.T) {
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	token, _, _, err := tokens.IssueAccess("session-1", "user-1", "org-1")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	publicMethods := map[string]bool{}
	sessionValidator := func(ctx context.Context, sessionID string) (bool, error) {
		return false, nil // session revoked
	}
	interceptor := AuthUnary(tokens, publicMethods, sessionValidator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	}))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	_, err = interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/ProtectedMethod",
	}, handler)
	if err == nil {
		t.Fatal("expected error for revoked session")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthUnary_SessionValidator_Error(t *testing.T) {
	tokens, err := security.NewTestTokenProvider()
	if err != nil {
		t.Fatalf("NewTestTokenProvider: %v", err)
	}
	token, _, _, err := tokens.IssueAccess("session-1", "user-1", "org-1")
	if err != nil {
		t.Fatalf("IssueAccess: %v", err)
	}

	publicMethods := map[string]bool{}
	sessionValidator := func(ctx context.Context, sessionID string) (bool, error) {
		return false, errors.New("database error")
	}
	interceptor := AuthUnary(tokens, publicMethods, sessionValidator)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	}))
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	_, err = interceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/ProtectedMethod",
	}, handler)
	if err == nil {
		t.Fatal("expected error for validator error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestExtractBearer_Valid(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Bearer token123",
	}))
	token := extractBearer(ctx)
	if token != "token123" {
		t.Errorf("token = %q, want %q", token, "token123")
	}
}

func TestExtractBearer_CaseInsensitive(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "bearer token123",
	}))
	token := extractBearer(ctx)
	if token != "token123" {
		t.Errorf("token = %q, want %q", token, "token123")
	}
}

func TestExtractBearer_Missing(t *testing.T) {
	ctx := context.Background()
	token := extractBearer(ctx)
	if token != "" {
		t.Errorf("token = %q, want empty", token)
	}
}

func TestExtractBearer_InvalidPrefix(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Basic token123",
	}))
	token := extractBearer(ctx)
	if token != "" {
		t.Errorf("token = %q, want empty", token)
	}
}

func TestExtractBearer_Whitespace(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "  Bearer   token123  ",
	}))
	token := extractBearer(ctx)
	if token != "token123" {
		t.Errorf("token = %q, want %q", token, "token123")
	}
}
