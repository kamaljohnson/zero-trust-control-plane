package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/v1/ast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	policyv1 "zero-trust-control-plane/backend/api/generated/policy/v1"
	"zero-trust-control-plane/backend/internal/policy/domain"
	"zero-trust-control-plane/backend/internal/policy/repository"
)

// Server implements PolicyService (proto server) for policy CRUD and evaluation.
// Proto: policy/policy.proto → internal/policy/handler.
type Server struct {
	policyv1.UnimplementedPolicyServiceServer
	repo repository.Repository
}

// NewServer returns a new Policy gRPC server. Pass nil repo for stub (Unimplemented).
func NewServer(repo repository.Repository) *Server {
	return &Server{repo: repo}
}

// CreatePolicy creates a new policy with Rego validation.
func (s *Server) CreatePolicy(ctx context.Context, req *policyv1.CreatePolicyRequest) (*policyv1.CreatePolicyResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method CreatePolicy not implemented")
	}
	if req.GetOrgId() == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id is required")
	}
	if req.GetRules() == "" {
		return nil, status.Error(codes.InvalidArgument, "rules (Rego policy) is required")
	}
	if err := validateRego(req.GetRules()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid Rego syntax: "+err.Error())
	}
	policy := &domain.Policy{
		ID:        uuid.New().String(),
		OrgID:     req.GetOrgId(),
		Rules:     req.GetRules(),
		Enabled:   req.GetEnabled(),
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, policy); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &policyv1.CreatePolicyResponse{Policy: policyToProto(policy)}, nil
}

// UpdatePolicy updates an existing policy with Rego validation.
func (s *Server) UpdatePolicy(ctx context.Context, req *policyv1.UpdatePolicyRequest) (*policyv1.UpdatePolicyResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method UpdatePolicy not implemented")
	}
	if req.GetPolicyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "policy_id is required")
	}
	if req.GetRules() != "" {
		if err := validateRego(req.GetRules()); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid Rego syntax: "+err.Error())
		}
	}
	existing, err := s.repo.GetByID(ctx, req.GetPolicyId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing == nil {
		return nil, status.Error(codes.NotFound, "policy not found")
	}
	existing.Rules = req.GetRules()
	existing.Enabled = req.GetEnabled()
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &policyv1.UpdatePolicyResponse{Policy: policyToProto(existing)}, nil
}

// DeletePolicy deletes a policy.
func (s *Server) DeletePolicy(ctx context.Context, req *policyv1.DeletePolicyRequest) (*policyv1.DeletePolicyResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method DeletePolicy not implemented")
	}
	if req.GetPolicyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "policy_id is required")
	}
	if err := s.repo.Delete(ctx, req.GetPolicyId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &policyv1.DeletePolicyResponse{}, nil
}

// ListPolicies returns a paginated list of policies for an org.
func (s *Server) ListPolicies(ctx context.Context, req *policyv1.ListPoliciesRequest) (*policyv1.ListPoliciesResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method ListPolicies not implemented")
	}
	if req.GetOrgId() == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id is required")
	}
	list, err := s.repo.ListByOrg(ctx, req.GetOrgId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	policies := make([]*policyv1.Policy, len(list))
	for i := range list {
		policies[i] = policyToProto(list[i])
	}
	return &policyv1.ListPoliciesResponse{Policies: policies}, nil
}

func validateRego(regoCode string) error {
	_, err := ast.ParseModule("", regoCode)
	return err
}

func policyToProto(p *domain.Policy) *policyv1.Policy {
	if p == nil {
		return nil
	}
	return &policyv1.Policy{
		Id:        p.ID,
		OrgId:     p.OrgID,
		Rules:     p.Rules,
		Enabled:   p.Enabled,
		CreatedAt: timestamppb.New(p.CreatedAt),
	}
}
