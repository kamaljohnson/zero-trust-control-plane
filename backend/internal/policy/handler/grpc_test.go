package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	policyv1 "zero-trust-control-plane/backend/api/generated/policy/v1"
	"zero-trust-control-plane/backend/internal/policy/domain"
)

// mockPolicyRepo implements repository.Repository for tests.
type mockPolicyRepo struct {
	policies  map[string]*domain.Policy
	byOrg     map[string][]*domain.Policy
	createErr error
	updateErr error
	deleteErr error
	listErr   error
	getByIDErr error
}

func (m *mockPolicyRepo) GetByID(ctx context.Context, id string) (*domain.Policy, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.policies[id], nil
}

func (m *mockPolicyRepo) ListByOrg(ctx context.Context, orgID string) ([]*domain.Policy, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.byOrg[orgID], nil
}

func (m *mockPolicyRepo) GetEnabledPoliciesByOrg(ctx context.Context, orgID string) ([]*domain.Policy, error) {
	return nil, nil
}

func (m *mockPolicyRepo) Create(ctx context.Context, p *domain.Policy) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.policies[p.ID] = p
	if m.byOrg[p.OrgID] == nil {
		m.byOrg[p.OrgID] = []*domain.Policy{}
	}
	m.byOrg[p.OrgID] = append(m.byOrg[p.OrgID], p)
	return nil
}

func (m *mockPolicyRepo) Update(ctx context.Context, p *domain.Policy) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.policies[p.ID] = p
	return nil
}

func (m *mockPolicyRepo) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.policies, id)
	return nil
}

func TestCreatePolicy_Success(t *testing.T) {
	validRego := `package mfa

default mfa_required = false

mfa_required if {
    input.is_new_device
}`
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.CreatePolicy(ctx, &policyv1.CreatePolicyRequest{
		OrgId:  "org-1",
		Rules:  validRego,
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	if resp == nil || resp.Policy == nil {
		t.Fatal("response or policy is nil")
	}
	if resp.Policy.OrgId != "org-1" {
		t.Errorf("policy org_id = %q, want %q", resp.Policy.OrgId, "org-1")
	}
	if resp.Policy.Rules != validRego {
		t.Errorf("policy rules mismatch")
	}
	if !resp.Policy.Enabled {
		t.Error("policy enabled = false, want true")
	}
}

func TestCreatePolicy_InvalidOrgID(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.CreatePolicy(ctx, &policyv1.CreatePolicyRequest{
		OrgId:  "",
		Rules:  "package test",
	})
	if err == nil {
		t.Fatal("expected error for empty org_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestCreatePolicy_EmptyRules(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.CreatePolicy(ctx, &policyv1.CreatePolicyRequest{
		OrgId:  "org-1",
		Rules:  "",
	})
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestCreatePolicy_InvalidRegoSyntax(t *testing.T) {
	invalidRego := `package test
invalid syntax {`
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.CreatePolicy(ctx, &policyv1.CreatePolicyRequest{
		OrgId:  "org-1",
		Rules:  invalidRego,
	})
	if err == nil {
		t.Fatal("expected error for invalid Rego syntax")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestCreatePolicy_RepositoryError(t *testing.T) {
	validRego := `package test`
	repo := &mockPolicyRepo{
		policies:  make(map[string]*domain.Policy),
		byOrg:     make(map[string][]*domain.Policy),
		createErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.CreatePolicy(ctx, &policyv1.CreatePolicyRequest{
		OrgId:  "org-1",
		Rules:  validRego,
	})
	if err == nil {
		t.Fatal("expected error for repository error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestCreatePolicy_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.CreatePolicy(ctx, &policyv1.CreatePolicyRequest{
		OrgId:  "org-1",
		Rules:  "package test",
	})
	if err == nil {
		t.Fatal("expected error for nil repo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}

func TestUpdatePolicy_Success(t *testing.T) {
	now := time.Now().UTC()
	existing := &domain.Policy{
		ID:        "policy-1",
		OrgID:     "org-1",
		Rules:     "package old",
		Enabled:   false,
		CreatedAt: now,
	}
	validRego := `package new

default allow = false

allow if {
    input.user.role == "admin"
}`
	repo := &mockPolicyRepo{
		policies: map[string]*domain.Policy{"policy-1": existing},
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.UpdatePolicy(ctx, &policyv1.UpdatePolicyRequest{
		PolicyId: "policy-1",
		Rules:    validRego,
		Enabled:  true,
	})
	if err != nil {
		t.Fatalf("UpdatePolicy: %v", err)
	}
	if resp.Policy.Rules != validRego {
		t.Errorf("policy rules mismatch")
	}
	if !resp.Policy.Enabled {
		t.Error("policy enabled = false, want true")
	}
}

func TestUpdatePolicy_InvalidPolicyID(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.UpdatePolicy(ctx, &policyv1.UpdatePolicyRequest{
		PolicyId: "",
		Rules:    "package test",
	})
	if err == nil {
		t.Fatal("expected error for empty policy_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestUpdatePolicy_NotFound(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.UpdatePolicy(ctx, &policyv1.UpdatePolicyRequest{
		PolicyId: "nonexistent",
		Rules:    "package test",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent policy")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestUpdatePolicy_InvalidRegoSyntax(t *testing.T) {
	now := time.Now().UTC()
	existing := &domain.Policy{
		ID:        "policy-1",
		OrgID:     "org-1",
		Rules:     "package old",
		Enabled:   false,
		CreatedAt: now,
	}
	invalidRego := `package test
invalid {`
	repo := &mockPolicyRepo{
		policies: map[string]*domain.Policy{"policy-1": existing},
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.UpdatePolicy(ctx, &policyv1.UpdatePolicyRequest{
		PolicyId: "policy-1",
		Rules:    invalidRego,
	})
	if err == nil {
		t.Fatal("expected error for invalid Rego syntax")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestUpdatePolicy_EmptyRulesAllowed(t *testing.T) {
	now := time.Now().UTC()
	existing := &domain.Policy{
		ID:        "policy-1",
		OrgID:     "org-1",
		Rules:     "package old",
		Enabled:   false,
		CreatedAt: now,
	}
	repo := &mockPolicyRepo{
		policies: map[string]*domain.Policy{"policy-1": existing},
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.UpdatePolicy(ctx, &policyv1.UpdatePolicyRequest{
		PolicyId: "policy-1",
		Rules:    "",
		Enabled:  true,
	})
	if err != nil {
		t.Fatalf("UpdatePolicy: %v", err)
	}
	if resp.Policy.Rules != "" {
		t.Errorf("policy rules = %q, want empty", resp.Policy.Rules)
	}
}

func TestDeletePolicy_Success(t *testing.T) {
	now := time.Now().UTC()
	existing := &domain.Policy{
		ID:        "policy-1",
		OrgID:     "org-1",
		Rules:     "package test",
		Enabled:   true,
		CreatedAt: now,
	}
	repo := &mockPolicyRepo{
		policies: map[string]*domain.Policy{"policy-1": existing},
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.DeletePolicy(ctx, &policyv1.DeletePolicyRequest{PolicyId: "policy-1"})
	if err != nil {
		t.Fatalf("DeletePolicy: %v", err)
	}
}

func TestDeletePolicy_InvalidPolicyID(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.DeletePolicy(ctx, &policyv1.DeletePolicyRequest{PolicyId: ""})
	if err == nil {
		t.Fatal("expected error for empty policy_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestDeletePolicy_RepositoryError(t *testing.T) {
	repo := &mockPolicyRepo{
		policies:  make(map[string]*domain.Policy),
		byOrg:     make(map[string][]*domain.Policy),
		deleteErr: errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.DeletePolicy(ctx, &policyv1.DeletePolicyRequest{PolicyId: "policy-1"})
	if err == nil {
		t.Fatal("expected error for repository error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestListPolicies_Success(t *testing.T) {
	now := time.Now().UTC()
	policies := []*domain.Policy{
		{ID: "policy-1", OrgID: "org-1", Rules: "package p1", Enabled: true, CreatedAt: now},
		{ID: "policy-2", OrgID: "org-1", Rules: "package p2", Enabled: false, CreatedAt: now},
	}
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    map[string][]*domain.Policy{"org-1": policies},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.ListPolicies(ctx, &policyv1.ListPoliciesRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListPolicies: %v", err)
	}
	if len(resp.Policies) != 2 {
		t.Errorf("policies count = %d, want 2", len(resp.Policies))
	}
}

func TestListPolicies_EmptyList(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    map[string][]*domain.Policy{"org-1": {}},
	}
	srv := NewServer(repo)
	ctx := context.Background()

	resp, err := srv.ListPolicies(ctx, &policyv1.ListPoliciesRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("ListPolicies: %v", err)
	}
	if len(resp.Policies) != 0 {
		t.Errorf("policies count = %d, want 0", len(resp.Policies))
	}
}

func TestListPolicies_InvalidOrgID(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.ListPolicies(ctx, &policyv1.ListPoliciesRequest{OrgId: ""})
	if err == nil {
		t.Fatal("expected error for empty org_id")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", st.Code(), codes.InvalidArgument)
	}
}

func TestListPolicies_RepositoryError(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string]*domain.Policy),
		byOrg:    make(map[string][]*domain.Policy),
		listErr:  errors.New("database error"),
	}
	srv := NewServer(repo)
	ctx := context.Background()

	_, err := srv.ListPolicies(ctx, &policyv1.ListPoliciesRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for repository error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestListPolicies_NilRepo(t *testing.T) {
	srv := NewServer(nil)
	ctx := context.Background()

	_, err := srv.ListPolicies(ctx, &policyv1.ListPoliciesRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for nil repo")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unimplemented)
	}
}
