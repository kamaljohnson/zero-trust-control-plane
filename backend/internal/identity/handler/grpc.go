package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "zero-trust-control-plane/backend/api/generated/auth/v1"
	"zero-trust-control-plane/backend/internal/identity/service"
)

// AuthServer implements AuthService (proto server) for register, login, refresh, logout, and identity linking.
// Proto: auth/auth.proto → internal/identity/handler.
type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	auth *service.AuthService
}

// NewAuthServer returns a new Auth gRPC server. Pass nil for auth to use stub implementations.
func NewAuthServer(auth *service.AuthService) *AuthServer {
	return &AuthServer{auth: auth}
}

// Register creates a new user and local identity.
func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	if s.auth == nil {
		return nil, status.Error(codes.Unimplemented, "method Register not implemented")
	}
	res, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetName())
	if err != nil {
		return nil, authErr(err)
	}
	return authResultToProto(res), nil
}

// Login authenticates the user and returns either tokens or MFA required (challenge_id, phone_mask).
func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if s.auth == nil {
		return nil, status.Error(codes.Unimplemented, "method Login not implemented")
	}
	res, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetOrgId(), req.GetDeviceFingerprint())
	if err != nil {
		return nil, authErr(err)
	}
	return loginResultToProto(res), nil
}

// VerifyMFA verifies the OTP for the given challenge and returns tokens.
func (s *AuthServer) VerifyMFA(ctx context.Context, req *authv1.VerifyMFARequest) (*authv1.AuthResponse, error) {
	if s.auth == nil {
		return nil, status.Error(codes.Unimplemented, "method VerifyMFA not implemented")
	}
	res, err := s.auth.VerifyMFA(ctx, req.GetChallengeId(), req.GetOtp())
	if err != nil {
		return nil, authErr(err)
	}
	return authResultToProto(res), nil
}

// SubmitPhoneAndRequestMFA consumes the intent, creates an MFA challenge for the submitted phone, sends OTP, and returns challenge_id and phone_mask.
func (s *AuthServer) SubmitPhoneAndRequestMFA(ctx context.Context, req *authv1.SubmitPhoneAndRequestMFARequest) (*authv1.SubmitPhoneAndRequestMFAResponse, error) {
	if s.auth == nil {
		return nil, status.Error(codes.Unimplemented, "method SubmitPhoneAndRequestMFA not implemented")
	}
	res, err := s.auth.SubmitPhoneAndRequestMFA(ctx, req.GetIntentId(), req.GetPhone())
	if err != nil {
		return nil, authErr(err)
	}
	return &authv1.SubmitPhoneAndRequestMFAResponse{
		ChallengeId: res.ChallengeID,
		PhoneMask:   res.PhoneMask,
	}, nil
}

// Refresh issues new access and refresh tokens, or returns MFA required / phone required when device-trust policy requires it.
func (s *AuthServer) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if s.auth == nil {
		return nil, status.Error(codes.Unimplemented, "method Refresh not implemented")
	}
	res, err := s.auth.Refresh(ctx, req.GetRefreshToken(), req.GetDeviceFingerprint())
	if err != nil {
		return nil, authErr(err)
	}
	return refreshResultToProto(res), nil
}

// Logout invalidates the session identified by the refresh token.
func (s *AuthServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*emptypb.Empty, error) {
	if s.auth == nil {
		return &emptypb.Empty{}, nil
	}
	if err := s.auth.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, authErr(err)
	}
	return &emptypb.Empty{}, nil
}

// VerifyCredentials validates email/password and returns user_id. Used for org-creation flow.
func (s *AuthServer) VerifyCredentials(ctx context.Context, req *authv1.VerifyCredentialsRequest) (*authv1.VerifyCredentialsResponse, error) {
	if s.auth == nil {
		return nil, status.Error(codes.Unimplemented, "method VerifyCredentials not implemented")
	}
	userID, err := s.auth.VerifyCredentials(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, authErr(err)
	}
	return &authv1.VerifyCredentialsResponse{UserId: userID}, nil
}

// LinkIdentity associates an external identity with the current user. Not implemented for password-only auth.
func (s *AuthServer) LinkIdentity(ctx context.Context, req *authv1.LinkIdentityRequest) (*authv1.LinkIdentityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method LinkIdentity not implemented for password-only auth")
}

func authErr(err error) error {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyRegistered):
		return status.Error(codes.AlreadyExists, "email already registered")
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, service.ErrInvalidRefreshToken):
		return status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	case errors.Is(err, service.ErrRefreshTokenReuse):
		return status.Error(codes.Unauthenticated, "refresh token reuse detected; all sessions revoked")
	case errors.Is(err, service.ErrNotOrgMember):
		return status.Error(codes.PermissionDenied, "user is not a member of the organization")
	case errors.Is(err, service.ErrPhoneRequiredForMFA):
		return status.Error(codes.FailedPrecondition, "phone number required for MFA; add in profile")
	case errors.Is(err, service.ErrInvalidMFAChallenge), errors.Is(err, service.ErrInvalidOTP):
		return status.Error(codes.Unauthenticated, "invalid or expired MFA challenge")
	case errors.Is(err, service.ErrInvalidMFAIntent):
		return status.Error(codes.Unauthenticated, "invalid or expired MFA intent")
	case errors.Is(err, service.ErrChallengeExpired):
		return status.Error(codes.FailedPrecondition, "MFA challenge expired")
	default:
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return nil
	}
}

func loginResultToProto(r *service.LoginResult) *authv1.LoginResponse {
	if r == nil {
		return &authv1.LoginResponse{}
	}
	if r.Tokens != nil {
		return &authv1.LoginResponse{
			Result: &authv1.LoginResponse_Tokens{Tokens: authResultToProto(r.Tokens)},
		}
	}
	if r.MFARequired != nil {
		return &authv1.LoginResponse{
			Result: &authv1.LoginResponse_MfaRequired{
				MfaRequired: &authv1.MFARequired{
					ChallengeId: r.MFARequired.ChallengeID,
					PhoneMask:   r.MFARequired.PhoneMask,
				},
			},
		}
	}
	if r.PhoneRequired != nil {
		return &authv1.LoginResponse{
			Result: &authv1.LoginResponse_PhoneRequired{
				PhoneRequired: &authv1.PhoneRequired{
					IntentId: r.PhoneRequired.IntentID,
				},
			},
		}
	}
	return &authv1.LoginResponse{}
}

func refreshResultToProto(r *service.RefreshResult) *authv1.RefreshResponse {
	if r == nil {
		return &authv1.RefreshResponse{}
	}
	if r.Tokens != nil {
		return &authv1.RefreshResponse{
			Result: &authv1.RefreshResponse_Tokens{Tokens: authResultToProto(r.Tokens)},
		}
	}
	if r.MFARequired != nil {
		return &authv1.RefreshResponse{
			Result: &authv1.RefreshResponse_MfaRequired{
				MfaRequired: &authv1.MFARequired{
					ChallengeId: r.MFARequired.ChallengeID,
					PhoneMask:   r.MFARequired.PhoneMask,
				},
			},
		}
	}
	if r.PhoneRequired != nil {
		return &authv1.RefreshResponse{
			Result: &authv1.RefreshResponse_PhoneRequired{
				PhoneRequired: &authv1.PhoneRequired{
					IntentId: r.PhoneRequired.IntentID,
				},
			},
		}
	}
	return &authv1.RefreshResponse{}
}

func authResultToProto(r *service.AuthResult) *authv1.AuthResponse {
	if r == nil {
		return &authv1.AuthResponse{}
	}
	out := &authv1.AuthResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		UserId:       r.UserID,
		OrgId:        r.OrgID,
	}
	if !r.ExpiresAt.IsZero() {
		out.ExpiresAt = timestamppb.New(r.ExpiresAt)
	}
	return out
}
