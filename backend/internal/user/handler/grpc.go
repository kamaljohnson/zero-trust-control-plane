package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "zero-trust-control-plane/backend/api/generated/user/v1"
)

// Server implements UserService (proto server) for user lifecycle.
// Proto: user/user.proto → internal/user/handler.
type Server struct {
	userv1.UnimplementedUserServiceServer
}

// NewServer returns a new User gRPC server.
func NewServer() *Server {
	return &Server{}
}

// GetUser returns a user by ID. TODO: implement.
func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
}

// ListUsers returns a paginated list of users. TODO: implement.
func (s *Server) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListUsers not implemented")
}

// DisableUser disables a user. TODO: implement.
func (s *Server) DisableUser(ctx context.Context, req *userv1.DisableUserRequest) (*userv1.DisableUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DisableUser not implemented")
}

// EnableUser re-enables a disabled user. TODO: implement.
func (s *Server) EnableUser(ctx context.Context, req *userv1.EnableUserRequest) (*userv1.EnableUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method EnableUser not implemented")
}
