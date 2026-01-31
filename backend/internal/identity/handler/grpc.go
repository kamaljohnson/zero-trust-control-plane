package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "zero-trust-control-plane/backend/api/generated/auth/v1"
)

// AuthServer implements AuthService (proto server) for login, logout, session validation, and identity linking.
// Proto: auth/auth.proto → internal/identity/handler.
type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
}

// NewAuthServer returns a new Auth gRPC server.
func NewAuthServer() *AuthServer {
	return &AuthServer{}
}

// Login authenticates the user and returns a session. TODO: implement.
func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Login not implemented")
}

// Logout invalidates the current session. TODO: implement.
func (s *AuthServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Logout not implemented")
}

// ValidateSession checks whether the session is valid and returns session info. TODO: implement.
func (s *AuthServer) ValidateSession(ctx context.Context, req *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ValidateSession not implemented")
}

// LinkIdentity associates an external identity with the current user. TODO: implement.
func (s *AuthServer) LinkIdentity(ctx context.Context, req *authv1.LinkIdentityRequest) (*authv1.LinkIdentityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method LinkIdentity not implemented")
}
