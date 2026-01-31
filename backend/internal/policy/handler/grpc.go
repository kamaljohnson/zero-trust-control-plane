package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	policyv1 "zero-trust-control-plane/backend/api/generated/policy/v1"
)

// Server implements PolicyService (proto server) for policy CRUD and evaluation.
// Proto: policy/policy.proto → internal/policy/handler.
type Server struct {
	policyv1.UnimplementedPolicyServiceServer
}

// NewServer returns a new Policy gRPC server.
func NewServer() *Server {
	return &Server{}
}

// CreatePolicy creates a new policy. TODO: implement.
func (s *Server) CreatePolicy(ctx context.Context, req *policyv1.CreatePolicyRequest) (*policyv1.CreatePolicyResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreatePolicy not implemented")
}

// UpdatePolicy updates an existing policy. TODO: implement.
func (s *Server) UpdatePolicy(ctx context.Context, req *policyv1.UpdatePolicyRequest) (*policyv1.UpdatePolicyResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdatePolicy not implemented")
}

// DeletePolicy deletes a policy. TODO: implement.
func (s *Server) DeletePolicy(ctx context.Context, req *policyv1.DeletePolicyRequest) (*policyv1.DeletePolicyResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeletePolicy not implemented")
}

// ListPolicies returns a paginated list of policies. TODO: implement.
func (s *Server) ListPolicies(ctx context.Context, req *policyv1.ListPoliciesRequest) (*policyv1.ListPoliciesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListPolicies not implemented")
}
