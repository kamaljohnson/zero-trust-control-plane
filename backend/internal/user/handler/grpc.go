package handler

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "zero-trust-control-plane/backend/api/generated/user/v1"
	"zero-trust-control-plane/backend/internal/user/domain"
	userrepo "zero-trust-control-plane/backend/internal/user/repository"
)

// Server implements UserService (proto server) for user lifecycle.
// Proto: user/user.proto → internal/user/handler.
type Server struct {
	userv1.UnimplementedUserServiceServer
	userRepo userrepo.Repository
}

// NewServer returns a new User gRPC server. userRepo may be nil; then all RPCs return Unimplemented.
func NewServer(userRepo userrepo.Repository) *Server {
	return &Server{userRepo: userRepo}
}

// GetUser returns a user by ID.
func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if s.userRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
	}
	userID := strings.TrimSpace(req.GetUserId())
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to look up user")
	}
	if u == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &userv1.GetUserResponse{
		User: domainUserToProto(u),
	}, nil
}

// GetUserByEmail returns a user by email. Caller must be authenticated.
func (s *Server) GetUserByEmail(ctx context.Context, req *userv1.GetUserByEmailRequest) (*userv1.GetUserByEmailResponse, error) {
	if s.userRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method GetUserByEmail not implemented")
	}
	email := strings.TrimSpace(req.GetEmail())
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email required")
	}
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to look up user")
	}
	if u == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &userv1.GetUserByEmailResponse{
		User: domainUserToProto(u),
	}, nil
}

// ListUsers returns a paginated list of users.
func (s *Server) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	if s.userRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method ListUsers not implemented")
	}
	return nil, status.Error(codes.Unimplemented, "method ListUsers not implemented")
}

// DisableUser disables a user.
func (s *Server) DisableUser(ctx context.Context, req *userv1.DisableUserRequest) (*userv1.DisableUserResponse, error) {
	if s.userRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method DisableUser not implemented")
	}
	return nil, status.Error(codes.Unimplemented, "method DisableUser not implemented")
}

// EnableUser re-enables a disabled user.
func (s *Server) EnableUser(ctx context.Context, req *userv1.EnableUserRequest) (*userv1.EnableUserResponse, error) {
	if s.userRepo == nil {
		return nil, status.Error(codes.Unimplemented, "method EnableUser not implemented")
	}
	return nil, status.Error(codes.Unimplemented, "method EnableUser not implemented")
}

func domainUserToProto(u *domain.User) *userv1.User {
	if u == nil {
		return nil
	}
	var status userv1.UserStatus
	switch u.Status {
	case domain.UserStatusActive:
		status = userv1.UserStatus_USER_STATUS_ACTIVE
	case domain.UserStatusDisabled:
		status = userv1.UserStatus_USER_STATUS_DISABLED
	default:
		status = userv1.UserStatus_USER_STATUS_UNSPECIFIED
	}
	return &userv1.User{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Status:    status,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
