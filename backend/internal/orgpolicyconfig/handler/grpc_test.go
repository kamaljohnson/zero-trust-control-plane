package handler

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orgpolicyconfigv1 "zero-trust-control-plane/backend/api/generated/orgpolicyconfig/v1"
	membershipdomain "zero-trust-control-plane/backend/internal/membership/domain"
	orgmfasettingsdomain "zero-trust-control-plane/backend/internal/orgmfasettings/domain"
	"zero-trust-control-plane/backend/internal/orgpolicyconfig/domain"
	"zero-trust-control-plane/backend/internal/server/interceptors"
)

// mockOrgPolicyConfigRepo implements repository.Repository for tests.
type mockOrgPolicyConfigRepo struct {
	configs map[string]*domain.OrgPolicyConfig
	err     error
}

func (m *mockOrgPolicyConfigRepo) GetByOrgID(ctx context.Context, orgID string) (*domain.OrgPolicyConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.configs[orgID], nil
}

func (m *mockOrgPolicyConfigRepo) Upsert(ctx context.Context, orgID string, config *domain.OrgPolicyConfig) error {
	if m.err != nil {
		return m.err
	}
	if m.configs == nil {
		m.configs = make(map[string]*domain.OrgPolicyConfig)
	}
	m.configs[orgID] = config
	return nil
}

// mockMembershipRepoForOrgPolicyConfig implements membershiprepo.Repository for tests.
type mockMembershipRepoForOrgPolicyConfig struct {
	memberships map[string]*membershipdomain.Membership
}

func (m *mockMembershipRepoForOrgPolicyConfig) GetMembershipByUserAndOrg(ctx context.Context, userID, orgID string) (*membershipdomain.Membership, error) {
	key := userID + ":" + orgID
	return m.memberships[key], nil
}

func (m *mockMembershipRepoForOrgPolicyConfig) GetMembershipByID(ctx context.Context, id string) (*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepoForOrgPolicyConfig) ListMembershipsByOrg(ctx context.Context, orgID string) ([]*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepoForOrgPolicyConfig) CreateMembership(ctx context.Context, mem *membershipdomain.Membership) error {
	return nil
}

func (m *mockMembershipRepoForOrgPolicyConfig) DeleteByUserAndOrg(ctx context.Context, userID, orgID string) error {
	return nil
}

func (m *mockMembershipRepoForOrgPolicyConfig) UpdateRole(ctx context.Context, userID, orgID string, role membershipdomain.Role) (*membershipdomain.Membership, error) {
	return nil, nil
}

func (m *mockMembershipRepoForOrgPolicyConfig) CountOwnersByOrg(ctx context.Context, orgID string) (int64, error) {
	return 0, nil
}

// mockOrgMFASettingsRepo implements orgmfasettingsrepo.Repository for tests.
type mockOrgMFASettingsRepo struct {
	settings map[string]*orgmfasettingsdomain.OrgMFASettings
	err      error
}

func (m *mockOrgMFASettingsRepo) GetByOrgID(ctx context.Context, orgID string) (*orgmfasettingsdomain.OrgMFASettings, error) {
	return m.settings[orgID], nil
}

func (m *mockOrgMFASettingsRepo) Upsert(ctx context.Context, s *orgmfasettingsdomain.OrgMFASettings) error {
	if m.err != nil {
		return m.err
	}
	if m.settings == nil {
		m.settings = make(map[string]*orgmfasettingsdomain.OrgMFASettings)
	}
	m.settings[s.OrgID] = s
	return nil
}

func ctxWithAdminForOrgPolicyConfig(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func ctxWithMemberForOrgPolicyConfig(orgID, userID string) context.Context {
	return interceptors.WithIdentity(context.Background(), userID, orgID, "session-1")
}

func TestGetOrgPolicyConfig_Success(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"example.com"},
			BlockedDomains:    []string{"malicious.com"},
			WildcardSupported: true,
			DefaultAction:     "deny",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithAdminForOrgPolicyConfig("org-1", "admin-1")

	resp, err := srv.GetOrgPolicyConfig(ctx, &orgpolicyconfigv1.GetOrgPolicyConfigRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("GetOrgPolicyConfig: %v", err)
	}
	if resp.Config == nil {
		t.Fatal("config is nil")
	}
	if resp.Config.AccessControl == nil {
		t.Fatal("access_control is nil")
	}
	if len(resp.Config.AccessControl.AllowedDomains) != 1 {
		t.Errorf("allowed domains count = %d, want 1", len(resp.Config.AccessControl.AllowedDomains))
	}
}

func TestGetOrgPolicyConfig_DefaultsMerging(t *testing.T) {
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": nil},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithAdminForOrgPolicyConfig("org-1", "admin-1")

	resp, err := srv.GetOrgPolicyConfig(ctx, &orgpolicyconfigv1.GetOrgPolicyConfigRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("GetOrgPolicyConfig: %v", err)
	}
	if resp.Config.AuthMfa == nil {
		t.Error("auth_mfa should have defaults")
	}
	if resp.Config.AccessControl == nil {
		t.Error("access_control should have defaults")
	}
}

func TestGetOrgPolicyConfig_NonAdminCaller(t *testing.T) {
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": {}},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	_, err := srv.GetOrgPolicyConfig(ctx, &orgpolicyconfigv1.GetOrgPolicyConfigRequest{OrgId: "org-1"})
	if err == nil {
		t.Fatal("expected error for non-admin caller")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestCheckUrlAccess_AllowedDomain(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"example.com"},
			BlockedDomains:    []string{},
			WildcardSupported: false,
			DefaultAction:     "deny",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "https://example.com/path",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if !resp.Allowed {
		t.Error("url should be allowed")
	}
}

func TestCheckUrlAccess_BlockedDomain(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"example.com"},
			BlockedDomains:    []string{"malicious.com"},
			WildcardSupported: false,
			DefaultAction:     "allow",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "https://malicious.com",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if resp.Allowed {
		t.Error("url should be blocked")
	}
	if resp.Reason == "" {
		t.Error("reason should be set")
	}
}

func TestCheckUrlAccess_WildcardMatching(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"*.example.com"},
			BlockedDomains:    []string{},
			WildcardSupported: true,
			DefaultAction:     "deny",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "https://sub.example.com",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if !resp.Allowed {
		t.Error("wildcard should match subdomain")
	}
}

func TestCheckUrlAccess_DefaultDeny(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"example.com"},
			BlockedDomains:    []string{},
			WildcardSupported: false,
			DefaultAction:     "deny",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "https://other.com",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if resp.Allowed {
		t.Error("url should be denied by default")
	}
}

func TestCheckUrlAccess_DefaultAllow(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{},
			BlockedDomains:    []string{},
			WildcardSupported: false,
			DefaultAction:     "allow",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "https://any.com",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if !resp.Allowed {
		t.Error("url should be allowed by default")
	}
}

func TestCheckUrlAccess_InvalidURL(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{},
			BlockedDomains:    []string{},
			WildcardSupported: false,
			DefaultAction:     "deny", // Use deny to ensure invalid URLs are rejected
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "not-a-url",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if resp.Allowed {
		t.Error("invalid URL should be denied")
	}
	if resp.Reason == "" {
		t.Error("reason should be set for invalid URL")
	}
}

func TestCheckUrlAccess_EmptyURL(t *testing.T) {
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": {}},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if resp.Allowed {
		t.Error("empty URL should be denied")
	}
}

func TestCheckUrlAccess_URLWithoutProtocol(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"example.com"},
			BlockedDomains:    []string{},
			WildcardSupported: false,
			DefaultAction:     "deny",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "example.com/path",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if !resp.Allowed {
		t.Error("URL without protocol should be handled")
	}
}

func TestCheckUrlAccess_CaseInsensitive(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"Example.COM"},
			BlockedDomains:    []string{},
			WildcardSupported: false,
			DefaultAction:     "deny",
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "https://EXAMPLE.com",
	})
	if err != nil {
		t.Fatalf("CheckUrlAccess: %v", err)
	}
	if !resp.Allowed {
		t.Error("domain matching should be case insensitive")
	}
}

func TestCheckUrlAccess_NonMemberCaller(t *testing.T) {
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": {}},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "nonmember-1")

	_, err := srv.CheckUrlAccess(ctx, &orgpolicyconfigv1.CheckUrlAccessRequest{
		OrgId: "org-1",
		Url:   "https://example.com",
	})
	if err == nil {
		t.Fatal("expected error for non-member caller")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("status code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestGetBrowserPolicy_Success(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"example.com"},
			BlockedDomains:    []string{},
			WildcardSupported: true,
			DefaultAction:     "allow",
		},
		ActionRestrictions: &domain.ActionRestrictions{
			AllowedActions: []string{"navigate", "download"},
			ReadOnlyMode:   false,
		},
	}
	repo := &mockOrgPolicyConfigRepo{
		configs: map[string]*domain.OrgPolicyConfig{"org-1": config},
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"member-1:org-1": {ID: "m1", UserID: "member-1", OrgID: "org-1", Role: membershipdomain.RoleMember},
		},
	}
	srv := NewServer(repo, membershipRepo, nil)
	ctx := ctxWithMemberForOrgPolicyConfig("org-1", "member-1")

	resp, err := srv.GetBrowserPolicy(ctx, &orgpolicyconfigv1.GetBrowserPolicyRequest{OrgId: "org-1"})
	if err != nil {
		t.Fatalf("GetBrowserPolicy: %v", err)
	}
	if resp.AccessControl == nil {
		t.Fatal("access_control is nil")
	}
	if resp.ActionRestrictions == nil {
		t.Fatal("action_restrictions is nil")
	}
}

func TestUpdateOrgPolicyConfig_SyncToMFASettings(t *testing.T) {
	repo := &mockOrgPolicyConfigRepo{
		configs: make(map[string]*domain.OrgPolicyConfig),
	}
	membershipRepo := &mockMembershipRepoForOrgPolicyConfig{
		memberships: map[string]*membershipdomain.Membership{
			"admin-1:org-1": {ID: "m1", UserID: "admin-1", OrgID: "org-1", Role: membershipdomain.RoleAdmin},
		},
	}
	mfaSettingsRepo := &mockOrgMFASettingsRepo{
		settings: make(map[string]*orgmfasettingsdomain.OrgMFASettings),
	}
	srv := NewServer(repo, membershipRepo, mfaSettingsRepo)
	ctx := ctxWithAdminForOrgPolicyConfig("org-1", "admin-1")

	config := &orgpolicyconfigv1.OrgPolicyConfig{
		AuthMfa: &orgpolicyconfigv1.AuthMfa{
			MfaRequirement: orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_ALWAYS,
		},
	}
	_, err := srv.UpdateOrgPolicyConfig(ctx, &orgpolicyconfigv1.UpdateOrgPolicyConfigRequest{
		OrgId:  "org-1",
		Config: config,
	})
	if err != nil {
		t.Fatalf("UpdateOrgPolicyConfig: %v", err)
	}
	if mfaSettingsRepo.settings["org-1"] == nil {
		t.Error("MFA settings should be synced")
	}
}
