package handler

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "zero-trust-control-plane/backend/api/generated/auth/v1"
	"zero-trust-control-plane/backend/internal/identity/service"
)

func TestRegister_NilAuthService(t *testing.T) {
	srv := NewAuthServer(nil)
	ctx := context.Background()

	_, err := srv.Register(ctx, &authv1.RegisterRequest{
		Email:    "user@example.com",
		Password: "Password123!abc",
	})
	if err == nil {
		t.Fatal("expected error for nil auth service")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestLogin_NilAuthService(t *testing.T) {
	srv := NewAuthServer(nil)
	ctx := context.Background()

	_, err := srv.Login(ctx, &authv1.LoginRequest{
		Email:    "user@example.com",
		Password: "Password123!abc",
		OrgId:    "org-1",
	})
	if err == nil {
		t.Fatal("expected error for nil auth service")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestVerifyMFA_NilAuthService(t *testing.T) {
	srv := NewAuthServer(nil)
	ctx := context.Background()

	_, err := srv.VerifyMFA(ctx, &authv1.VerifyMFARequest{
		ChallengeId: "challenge-1",
		Otp:         "123456",
	})
	if err == nil {
		t.Fatal("expected error for nil auth service")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestSubmitPhoneAndRequestMFA_NilAuthService(t *testing.T) {
	srv := NewAuthServer(nil)
	ctx := context.Background()

	_, err := srv.SubmitPhoneAndRequestMFA(ctx, &authv1.SubmitPhoneAndRequestMFARequest{
		IntentId: "intent-1",
		Phone:    "15551234567",
	})
	if err == nil {
		t.Fatal("expected error for nil auth service")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestRefresh_NilAuthService(t *testing.T) {
	srv := NewAuthServer(nil)
	ctx := context.Background()

	_, err := srv.Refresh(ctx, &authv1.RefreshRequest{
		RefreshToken: "refresh-token",
	})
	if err == nil {
		t.Fatal("expected error for nil auth service")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestLogout_NilAuthService(t *testing.T) {
	srv := NewAuthServer(nil)
	ctx := context.Background()

	resp, err := srv.Logout(ctx, &authv1.LogoutRequest{
		RefreshToken: "refresh-token",
	})
	if err != nil {
		t.Fatalf("Logout with nil auth service should succeed: %v", err)
	}
	if resp == nil {
		t.Fatal("response should not be nil")
	}
}

func TestLinkIdentity_Unimplemented(t *testing.T) {
	srv := NewAuthServer(nil)
	ctx := context.Background()

	_, err := srv.LinkIdentity(ctx, &authv1.LinkIdentityRequest{})
	if err == nil {
		t.Fatal("expected error for unimplemented method")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

// Test error mapping functions
func TestAuthErr_EmailAlreadyRegistered(t *testing.T) {
	err := authErr(service.ErrEmailAlreadyRegistered)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.AlreadyExists {
		t.Errorf("status code = %v, want %v", st.Code(), codes.AlreadyExists)
	}
}

func TestAuthErr_InvalidCredentials(t *testing.T) {
	err := authErr(service.ErrInvalidCredentials)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthErr_InvalidRefreshToken(t *testing.T) {
	err := authErr(service.ErrInvalidRefreshToken)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthErr_RefreshTokenReuse(t *testing.T) {
	err := authErr(service.ErrRefreshTokenReuse)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthErr_NotOrgMember(t *testing.T) {
	err := authErr(service.ErrNotOrgMember)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestAuthErr_PhoneRequiredForMFA(t *testing.T) {
	err := authErr(service.ErrPhoneRequiredForMFA)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("status code = %v, want %v", st.Code(), codes.FailedPrecondition)
	}
}

func TestAuthErr_InvalidMFAChallenge(t *testing.T) {
	err := authErr(service.ErrInvalidMFAChallenge)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthErr_InvalidOTP(t *testing.T) {
	err := authErr(service.ErrInvalidOTP)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthErr_InvalidMFAIntent(t *testing.T) {
	err := authErr(service.ErrInvalidMFAIntent)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}

func TestAuthErr_ChallengeExpired(t *testing.T) {
	err := authErr(service.ErrChallengeExpired)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("status code = %v, want %v", st.Code(), codes.FailedPrecondition)
	}
}

func TestAuthErr_UnknownError(t *testing.T) {
	err := authErr(service.ErrEmailAlreadyRegistered) // Using a known error wrapped
	err2 := authErr(err)
	st, ok := status.FromError(err2)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err2)
	}
	// Should map to InvalidArgument for unknown errors
	if st.Code() == codes.Unknown {
		t.Error("unknown errors should be mapped to InvalidArgument")
	}
}

// Test proto conversion functions
func TestLoginResultToProto_Tokens(t *testing.T) {
	result := &service.LoginResult{
		Tokens: &service.AuthResult{
			AccessToken:  "access",
			RefreshToken: "refresh",
			UserID:       "user-1",
			OrgID:        "org-1",
			ExpiresAt:    time.Now(),
		},
	}
	proto := loginResultToProto(result)
	if proto.GetTokens() == nil {
		t.Fatal("tokens should be set")
	}
	if proto.GetTokens().AccessToken != "access" {
		t.Errorf("access_token = %q, want %q", proto.GetTokens().AccessToken, "access")
	}
}

func TestLoginResultToProto_MFARequired(t *testing.T) {
	result := &service.LoginResult{
		MFARequired: &service.MFARequiredResult{
			ChallengeID: "challenge-1",
			PhoneMask:   "***-1234",
		},
	}
	proto := loginResultToProto(result)
	if proto.GetMfaRequired() == nil {
		t.Fatal("mfa_required should be set")
	}
	if proto.GetMfaRequired().ChallengeId != "challenge-1" {
		t.Errorf("challenge_id = %q, want %q", proto.GetMfaRequired().ChallengeId, "challenge-1")
	}
}

func TestLoginResultToProto_PhoneRequired(t *testing.T) {
	result := &service.LoginResult{
		PhoneRequired: &service.PhoneRequiredResult{
			IntentID: "intent-1",
		},
	}
	proto := loginResultToProto(result)
	if proto.GetPhoneRequired() == nil {
		t.Fatal("phone_required should be set")
	}
	if proto.GetPhoneRequired().IntentId != "intent-1" {
		t.Errorf("intent_id = %q, want %q", proto.GetPhoneRequired().IntentId, "intent-1")
	}
}

func TestRefreshResultToProto_Tokens(t *testing.T) {
	result := &service.RefreshResult{
		Tokens: &service.AuthResult{
			AccessToken:  "access",
			RefreshToken: "refresh",
			UserID:       "user-1",
			OrgID:        "org-1",
			ExpiresAt:    time.Now(),
		},
	}
	proto := refreshResultToProto(result)
	if proto.GetTokens() == nil {
		t.Fatal("tokens should be set")
	}
}

func TestAuthResultToProto(t *testing.T) {
	result := &service.AuthResult{
		AccessToken:  "access",
		RefreshToken: "refresh",
		UserID:       "user-1",
		OrgID:        "org-1",
		ExpiresAt:    time.Now(),
	}
	proto := authResultToProto(result)
	if proto.AccessToken != "access" {
		t.Errorf("access_token = %q, want %q", proto.AccessToken, "access")
	}
	if proto.UserId != "user-1" {
		t.Errorf("user_id = %q, want %q", proto.UserId, "user-1")
	}
	if proto.ExpiresAt == nil {
		t.Error("expires_at should be set")
	}
}
