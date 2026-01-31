package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	membershipv1 "zero-trust-control-plane/backend/api/generated/membership/v1"
)

// Server implements MembershipService (proto server) for org membership and roles.
// Proto: membership/membership.proto → internal/membership/handler.
type Server struct {
	membershipv1.UnimplementedMembershipServiceServer
}

// NewServer returns a new Membership gRPC server.
func NewServer() *Server {
	return &Server{}
}

// AddMember adds a member to an organization. TODO: implement.
func (s *Server) AddMember(ctx context.Context, req *membershipv1.AddMemberRequest) (*membershipv1.AddMemberResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method AddMember not implemented")
}

// RemoveMember removes a member from an organization. TODO: implement.
func (s *Server) RemoveMember(ctx context.Context, req *membershipv1.RemoveMemberRequest) (*membershipv1.RemoveMemberResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RemoveMember not implemented")
}

// UpdateRole updates a member's role. TODO: implement.
func (s *Server) UpdateRole(ctx context.Context, req *membershipv1.UpdateRoleRequest) (*membershipv1.UpdateRoleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateRole not implemented")
}

// ListMembers returns a paginated list of members. TODO: implement.
func (s *Server) ListMembers(ctx context.Context, req *membershipv1.ListMembersRequest) (*membershipv1.ListMembersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListMembers not implemented")
}
