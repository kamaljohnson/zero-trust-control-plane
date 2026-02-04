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

func TestDefaultActionToProto(t *testing.T) {
	testCases := []struct {
		input    string
		expected orgpolicyconfigv1.DefaultAction
	}{
		{"allow", orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_ALLOW},
		{"deny", orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_DENY},
		{"invalid", orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_UNSPECIFIED},
		{"", orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_UNSPECIFIED},
	}

	for _, tc := range testCases {
		result := defaultActionToProto(tc.input)
		if result != tc.expected {
			t.Errorf("defaultActionToProto(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestDefaultActionToDomain(t *testing.T) {
	testCases := []struct {
		input    orgpolicyconfigv1.DefaultAction
		expected string
	}{
		{orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_ALLOW, "allow"},
		{orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_DENY, "deny"},
		{orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_UNSPECIFIED, "allow"},
	}

	for _, tc := range testCases {
		result := defaultActionToDomain(tc.input)
		if result != tc.expected {
			t.Errorf("defaultActionToDomain(%v) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestMfaRequirementToProto(t *testing.T) {
	testCases := []struct {
		input    string
		expected orgpolicyconfigv1.MfaRequirement
	}{
		{"always", orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_ALWAYS},
		{"new_device", orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_NEW_DEVICE},
		{"untrusted", orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNTRUSTED},
		{"invalid", orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNSPECIFIED},
		{"", orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNSPECIFIED},
	}

	for _, tc := range testCases {
		result := mfaRequirementToProto(tc.input)
		if result != tc.expected {
			t.Errorf("mfaRequirementToProto(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestMfaRequirementToDomain(t *testing.T) {
	testCases := []struct {
		input    orgpolicyconfigv1.MfaRequirement
		expected string
	}{
		{orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_ALWAYS, "always"},
		{orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_NEW_DEVICE, "new_device"},
		{orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNTRUSTED, "untrusted"},
		{orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNSPECIFIED, "new_device"},
	}

	for _, tc := range testCases {
		result := mfaRequirementToDomain(tc.input)
		if result != tc.expected {
			t.Errorf("mfaRequirementToDomain(%v) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestPtr(t *testing.T) {
	// Test with string
	strVal := "test"
	strPtr := ptr(strVal)
	if strPtr == nil {
		t.Fatal("ptr should return non-nil pointer")
	}
	if *strPtr != strVal {
		t.Errorf("ptr string = %q, want %q", *strPtr, strVal)
	}

	// Test with int
	intVal := 42
	intPtr := ptr(intVal)
	if *intPtr != intVal {
		t.Errorf("ptr int = %d, want %d", *intPtr, intVal)
	}

	// Test with bool
	boolVal := true
	boolPtr := ptr(boolVal)
	if *boolPtr != boolVal {
		t.Errorf("ptr bool = %v, want %v", *boolPtr, boolVal)
	}
}

func TestEvaluateURLAccess_BlockedTakesPrecedence(t *testing.T) {
	ac := &domain.AccessControl{
		AllowedDomains:    []string{"example.com"},
		BlockedDomains:    []string{"example.com"},
		WildcardSupported: false,
		DefaultAction:     "allow",
	}
	allowed, reason := evaluateURLAccess("https://example.com", ac)
	if allowed {
		t.Error("blocked domain should take precedence over allowed")
	}
	if reason == "" {
		t.Error("reason should be set when blocked")
	}
}

func TestEvaluateURLAccess_WildcardBlocked(t *testing.T) {
	ac := &domain.AccessControl{
		AllowedDomains:    []string{},
		BlockedDomains:    []string{"*.malicious.com"},
		WildcardSupported: true,
		DefaultAction:     "allow",
	}
	allowed, reason := evaluateURLAccess("https://sub.malicious.com", ac)
	if allowed {
		t.Error("wildcard blocked domain should be denied")
	}
	if reason == "" {
		t.Error("reason should be set when blocked")
	}
}

func TestEvaluateURLAccess_WildcardAllowed(t *testing.T) {
	ac := &domain.AccessControl{
		AllowedDomains:    []string{"*.example.com"},
		BlockedDomains:    []string{},
		WildcardSupported: true,
		DefaultAction:     "deny",
	}
	allowed, reason := evaluateURLAccess("https://sub.example.com", ac)
	if !allowed {
		t.Error("wildcard allowed domain should be allowed")
	}
	if reason != "" {
		t.Errorf("reason should be empty when allowed, got %q", reason)
	}
}

func TestEvaluateURLAccess_ExactMatchTakesPrecedence(t *testing.T) {
	ac := &domain.AccessControl{
		AllowedDomains:    []string{"example.com", "*.example.com"},
		BlockedDomains:    []string{},
		WildcardSupported: true,
		DefaultAction:     "deny",
	}
	allowed, _ := evaluateURLAccess("https://example.com", ac)
	if !allowed {
		t.Error("exact match should work")
	}
}

// Tests for helper functions: domainToOrgMFASettings, domainToProto, protoToDomain, extractHost, matchWildcard

func TestDomainToOrgMFASettings_Always(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement: "always",
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	if settings.OrgID != "org-1" {
		t.Errorf("OrgID = %q, want %q", settings.OrgID, "org-1")
	}
	if !settings.MFARequiredAlways {
		t.Error("MFARequiredAlways should be true for 'always'")
	}
	if settings.MFARequiredForNewDevice {
		t.Error("MFARequiredForNewDevice should be false for 'always'")
	}
	if settings.MFARequiredForUntrusted {
		t.Error("MFARequiredForUntrusted should be false for 'always'")
	}
}

func TestDomainToOrgMFASettings_NewDevice(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement: "new_device",
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	if !settings.MFARequiredForNewDevice {
		t.Error("MFARequiredForNewDevice should be true for 'new_device'")
	}
	if !settings.MFARequiredForUntrusted {
		t.Error("MFARequiredForUntrusted should be true for 'new_device'")
	}
	if settings.MFARequiredAlways {
		t.Error("MFARequiredAlways should be false for 'new_device'")
	}
}

func TestDomainToOrgMFASettings_Untrusted(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement: "untrusted",
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	if !settings.MFARequiredForUntrusted {
		t.Error("MFARequiredForUntrusted should be true for 'untrusted'")
	}
	if settings.MFARequiredForNewDevice {
		t.Error("MFARequiredForNewDevice should be false for 'untrusted'")
	}
	if settings.MFARequiredAlways {
		t.Error("MFARequiredAlways should be false for 'untrusted'")
	}
}

func TestDomainToOrgMFASettings_DefaultMFARequirement(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement: "invalid",
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	// Default should be new_device behavior
	if !settings.MFARequiredForNewDevice {
		t.Error("MFARequiredForNewDevice should be true by default")
	}
	if !settings.MFARequiredForUntrusted {
		t.Error("MFARequiredForUntrusted should be true by default")
	}
}

func TestDomainToOrgMFASettings_DeviceTrust_AutoTrust(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		DeviceTrust: &domain.DeviceTrust{
			AutoTrustAfterMfa: false,
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	if settings.RegisterTrustAfterMFA {
		t.Error("RegisterTrustAfterMFA should be false when AutoTrustAfterMfa is false")
	}
}

func TestDomainToOrgMFASettings_DeviceTrust_ReverifyInterval(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		DeviceTrust: &domain.DeviceTrust{
			ReverifyIntervalDays: 60,
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	if settings.TrustTTLDays != 60 {
		t.Errorf("TrustTTLDays = %d, want 60", settings.TrustTTLDays)
	}
}

func TestDomainToOrgMFASettings_DeviceTrust_ZeroReverifyInterval(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		DeviceTrust: &domain.DeviceTrust{
			ReverifyIntervalDays: 0,
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	// Should use default 30 when zero
	if settings.TrustTTLDays != 30 {
		t.Errorf("TrustTTLDays = %d, want 30 (default)", settings.TrustTTLDays)
	}
}

func TestDomainToOrgMFASettings_NilAuthMfaAndDeviceTrust(t *testing.T) {
	config := &domain.OrgPolicyConfig{}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	// Should use defaults
	if !settings.MFARequiredForNewDevice {
		t.Error("MFARequiredForNewDevice should be true by default")
	}
	if !settings.MFARequiredForUntrusted {
		t.Error("MFARequiredForUntrusted should be true by default")
	}
	if settings.MFARequiredAlways {
		t.Error("MFARequiredAlways should be false by default")
	}
	if !settings.RegisterTrustAfterMFA {
		t.Error("RegisterTrustAfterMFA should be true by default")
	}
	if settings.TrustTTLDays != 30 {
		t.Errorf("TrustTTLDays = %d, want 30", settings.TrustTTLDays)
	}
}

func TestDomainToOrgMFASettings_Combined(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement: "always",
		},
		DeviceTrust: &domain.DeviceTrust{
			AutoTrustAfterMfa:         true,
			ReverifyIntervalDays:      45,
		},
	}
	settings := domainToOrgMFASettings("org-1", config)
	if settings == nil {
		t.Fatal("settings should not be nil")
	}
	if !settings.MFARequiredAlways {
		t.Error("MFARequiredAlways should be true")
	}
	if !settings.RegisterTrustAfterMFA {
		t.Error("RegisterTrustAfterMFA should be true")
	}
	if settings.TrustTTLDays != 45 {
		t.Errorf("TrustTTLDays = %d, want 45", settings.TrustTTLDays)
	}
}

func TestDomainToProto_NilConfig(t *testing.T) {
	proto := domainToProto(nil)
	if proto != nil {
		t.Error("domainToProto(nil) should return nil")
	}
}

func TestDomainToProto_AllSections(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement:         "always",
			AllowedMfaMethods:      []string{"sms_otp", "totp"},
			StepUpSensitiveActions: true,
			StepUpPolicyViolation:  true,
		},
		DeviceTrust: &domain.DeviceTrust{
			DeviceRegistrationAllowed: true,
			AutoTrustAfterMfa:         false,
			MaxTrustedDevicesPerUser:  5,
			ReverifyIntervalDays:      60,
			AdminRevokeAllowed:        true,
		},
		SessionMgmt: &domain.SessionMgmt{
			SessionMaxTtl:          "12h",
			IdleTimeout:            "15m",
			ConcurrentSessionLimit: 3,
			AdminForcedLogout:      false,
			ReauthOnPolicyChange:   true,
		},
		AccessControl: &domain.AccessControl{
			AllowedDomains:    []string{"example.com", "test.com"},
			BlockedDomains:    []string{"malicious.com"},
			WildcardSupported: true,
			DefaultAction:     "deny",
		},
		ActionRestrictions: &domain.ActionRestrictions{
			AllowedActions: []string{"navigate", "download"},
			ReadOnlyMode:   true,
		},
	}
	proto := domainToProto(config)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.AuthMfa == nil {
		t.Fatal("AuthMfa should not be nil")
	}
	if proto.AuthMfa.MfaRequirement != orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_ALWAYS {
		t.Errorf("MfaRequirement = %v, want ALWAYS", proto.AuthMfa.MfaRequirement)
	}
	if len(proto.AuthMfa.AllowedMfaMethods) != 2 {
		t.Errorf("AllowedMfaMethods length = %d, want 2", len(proto.AuthMfa.AllowedMfaMethods))
	}
	if !proto.AuthMfa.StepUpSensitiveActions {
		t.Error("StepUpSensitiveActions should be true")
	}
	if proto.DeviceTrust == nil {
		t.Fatal("DeviceTrust should not be nil")
	}
	if proto.DeviceTrust.MaxTrustedDevicesPerUser != 5 {
		t.Errorf("MaxTrustedDevicesPerUser = %d, want 5", proto.DeviceTrust.MaxTrustedDevicesPerUser)
	}
	if proto.SessionMgmt == nil {
		t.Fatal("SessionMgmt should not be nil")
	}
	if proto.SessionMgmt.ConcurrentSessionLimit != 3 {
		t.Errorf("ConcurrentSessionLimit = %d, want 3", proto.SessionMgmt.ConcurrentSessionLimit)
	}
	if proto.AccessControl == nil {
		t.Fatal("AccessControl should not be nil")
	}
	if len(proto.AccessControl.AllowedDomains) != 2 {
		t.Errorf("AllowedDomains length = %d, want 2", len(proto.AccessControl.AllowedDomains))
	}
	if proto.ActionRestrictions == nil {
		t.Fatal("ActionRestrictions should not be nil")
	}
	if !proto.ActionRestrictions.ReadOnlyMode {
		t.Error("ReadOnlyMode should be true")
	}
}

func TestDomainToProto_PartialConfig(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement: "new_device",
		},
		// Other sections are nil
	}
	proto := domainToProto(config)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.AuthMfa == nil {
		t.Fatal("AuthMfa should not be nil")
	}
	if proto.DeviceTrust != nil {
		t.Error("DeviceTrust should be nil")
	}
	if proto.SessionMgmt != nil {
		t.Error("SessionMgmt should be nil")
	}
}

func TestDomainToProto_EmptySlices(t *testing.T) {
	config := &domain.OrgPolicyConfig{
		AuthMfa: &domain.AuthMfa{
			MfaRequirement:    "always",
			AllowedMfaMethods: []string{},
		},
		AccessControl: &domain.AccessControl{
			AllowedDomains: []string{},
			BlockedDomains: []string{},
		},
		ActionRestrictions: &domain.ActionRestrictions{
			AllowedActions: []string{},
		},
	}
	proto := domainToProto(config)
	if proto == nil {
		t.Fatal("proto should not be nil")
	}
	if proto.AuthMfa == nil {
		t.Fatal("AuthMfa should not be nil")
	}
	if len(proto.AuthMfa.AllowedMfaMethods) != 0 {
		t.Errorf("AllowedMfaMethods length = %d, want 0", len(proto.AuthMfa.AllowedMfaMethods))
	}
	if len(proto.AccessControl.AllowedDomains) != 0 {
		t.Errorf("AllowedDomains length = %d, want 0", len(proto.AccessControl.AllowedDomains))
	}
}

func TestProtoToDomain_NilProto(t *testing.T) {
	domain := protoToDomain(nil)
	if domain != nil {
		t.Error("protoToDomain(nil) should return nil")
	}
}

func TestProtoToDomain_AllSections(t *testing.T) {
	proto := &orgpolicyconfigv1.OrgPolicyConfig{
		AuthMfa: &orgpolicyconfigv1.AuthMfa{
			MfaRequirement:         orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_ALWAYS,
			AllowedMfaMethods:      []string{"sms_otp"},
			StepUpSensitiveActions: true,
			StepUpPolicyViolation:  true,
		},
		DeviceTrust: &orgpolicyconfigv1.DeviceTrust{
			DeviceRegistrationAllowed: true,
			AutoTrustAfterMfa:         false,
			MaxTrustedDevicesPerUser:  10,
			ReverifyIntervalDays:      90,
			AdminRevokeAllowed:        false,
		},
		SessionMgmt: &orgpolicyconfigv1.SessionMgmt{
			SessionMaxTtl:          "48h",
			IdleTimeout:            "1h",
			ConcurrentSessionLimit: 5,
			AdminForcedLogout:      true,
			ReauthOnPolicyChange:   false,
		},
		AccessControl: &orgpolicyconfigv1.AccessControl{
			AllowedDomains:    []string{"allowed.com"},
			BlockedDomains:    []string{"blocked.com"},
			WildcardSupported: false,
			DefaultAction:     orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_DENY,
		},
		ActionRestrictions: &orgpolicyconfigv1.ActionRestrictions{
			AllowedActions: []string{"navigate"},
			ReadOnlyMode:   false,
		},
	}
	domainConfig := protoToDomain(proto)
	if domainConfig == nil {
		t.Fatal("domain config should not be nil")
	}
	if domainConfig.AuthMfa == nil {
		t.Fatal("AuthMfa should not be nil")
	}
	if domainConfig.AuthMfa.MfaRequirement != "always" {
		t.Errorf("MfaRequirement = %q, want %q", domainConfig.AuthMfa.MfaRequirement, "always")
	}
	if domainConfig.DeviceTrust == nil {
		t.Fatal("DeviceTrust should not be nil")
	}
	if domainConfig.DeviceTrust.MaxTrustedDevicesPerUser != 10 {
		t.Errorf("MaxTrustedDevicesPerUser = %d, want 10", domainConfig.DeviceTrust.MaxTrustedDevicesPerUser)
	}
	if domainConfig.SessionMgmt == nil {
		t.Fatal("SessionMgmt should not be nil")
	}
	if domainConfig.SessionMgmt.ConcurrentSessionLimit != 5 {
		t.Errorf("ConcurrentSessionLimit = %d, want 5", domainConfig.SessionMgmt.ConcurrentSessionLimit)
	}
	if domainConfig.AccessControl == nil {
		t.Fatal("AccessControl should not be nil")
	}
	if domainConfig.AccessControl.DefaultAction != "deny" {
		t.Errorf("DefaultAction = %q, want %q", domainConfig.AccessControl.DefaultAction, "deny")
	}
	if domainConfig.ActionRestrictions == nil {
		t.Fatal("ActionRestrictions should not be nil")
	}
	if len(domainConfig.ActionRestrictions.AllowedActions) != 1 {
		t.Errorf("AllowedActions length = %d, want 1", len(domainConfig.ActionRestrictions.AllowedActions))
	}
}

func TestProtoToDomain_PartialProto(t *testing.T) {
	proto := &orgpolicyconfigv1.OrgPolicyConfig{
		AuthMfa: &orgpolicyconfigv1.AuthMfa{
			MfaRequirement: orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_NEW_DEVICE,
		},
		// Other sections are nil
	}
	domainConfig := protoToDomain(proto)
	if domainConfig == nil {
		t.Fatal("domain config should not be nil")
	}
	if domainConfig.AuthMfa == nil {
		t.Fatal("AuthMfa should not be nil")
	}
	if domainConfig.DeviceTrust != nil {
		t.Error("DeviceTrust should be nil")
	}
}

func TestProtoToDomain_TypeConversions(t *testing.T) {
	proto := &orgpolicyconfigv1.OrgPolicyConfig{
		DeviceTrust: &orgpolicyconfigv1.DeviceTrust{
			MaxTrustedDevicesPerUser:  int32(100),
			ReverifyIntervalDays:      int32(120),
		},
		SessionMgmt: &orgpolicyconfigv1.SessionMgmt{
			ConcurrentSessionLimit: int32(50),
		},
	}
	domainConfig := protoToDomain(proto)
	if domainConfig == nil {
		t.Fatal("domain config should not be nil")
	}
	if domainConfig.DeviceTrust.MaxTrustedDevicesPerUser != 100 {
		t.Errorf("MaxTrustedDevicesPerUser = %d, want 100", domainConfig.DeviceTrust.MaxTrustedDevicesPerUser)
	}
	if domainConfig.DeviceTrust.ReverifyIntervalDays != 120 {
		t.Errorf("ReverifyIntervalDays = %d, want 120", domainConfig.DeviceTrust.ReverifyIntervalDays)
	}
	if domainConfig.SessionMgmt.ConcurrentSessionLimit != 50 {
		t.Errorf("ConcurrentSessionLimit = %d, want 50", domainConfig.SessionMgmt.ConcurrentSessionLimit)
	}
}

func TestExtractHost_WithProtocol(t *testing.T) {
	host, err := extractHost("https://example.com")
	if err != nil {
		t.Fatalf("extractHost: %v", err)
	}
	if host != "example.com" {
		t.Errorf("host = %q, want %q", host, "example.com")
	}
}

func TestExtractHost_WithoutProtocol(t *testing.T) {
	host, err := extractHost("example.com")
	if err != nil {
		t.Fatalf("extractHost: %v", err)
	}
	if host != "example.com" {
		t.Errorf("host = %q, want %q", host, "example.com")
	}
}

func TestExtractHost_WithPort(t *testing.T) {
	host, err := extractHost("https://example.com:8080")
	if err != nil {
		t.Fatalf("extractHost: %v", err)
	}
	if host != "example.com" {
		t.Errorf("host = %q, want %q", host, "example.com")
	}
}

func TestExtractHost_WithPath(t *testing.T) {
	host, err := extractHost("https://example.com/path/to/resource")
	if err != nil {
		t.Fatalf("extractHost: %v", err)
	}
	if host != "example.com" {
		t.Errorf("host = %q, want %q", host, "example.com")
	}
}

func TestExtractHost_WithQuery(t *testing.T) {
	host, err := extractHost("https://example.com?param=value")
	if err != nil {
		t.Fatalf("extractHost: %v", err)
	}
	if host != "example.com" {
		t.Errorf("host = %q, want %q", host, "example.com")
	}
}

func TestExtractHost_InvalidURL(t *testing.T) {
	_, err := extractHost("://invalid")
	if err == nil {
		t.Error("extractHost with invalid URL should return error")
	}
}

func TestExtractHost_EmptyHostname(t *testing.T) {
	host, err := extractHost("https://")
	if err != nil {
		t.Fatalf("extractHost: %v", err)
	}
	if host != "" {
		t.Errorf("host = %q, want empty string", host)
	}
}

func TestExtractHost_EmptyString(t *testing.T) {
	host, err := extractHost("")
	if err != nil {
		t.Fatalf("extractHost: %v", err)
	}
	if host != "" {
		t.Errorf("host = %q, want empty string", host)
	}
}

func TestMatchWildcard_ExactMatch(t *testing.T) {
	// Pattern "*.example.com" should match ".example.com" (exact suffix match)
	if !matchWildcard(".example.com", "*.example.com") {
		t.Error("matchWildcard should match exact suffix")
	}
}

func TestMatchWildcard_SubdomainMatch(t *testing.T) {
	if !matchWildcard("sub.example.com", "*.example.com") {
		t.Error("matchWildcard should match subdomain")
	}
	if !matchWildcard("deep.sub.example.com", "*.example.com") {
		t.Error("matchWildcard should match nested subdomain")
	}
}

func TestMatchWildcard_NonWildcardPattern(t *testing.T) {
	if matchWildcard("example.com", "example.com") {
		t.Error("matchWildcard should return false for non-wildcard pattern")
	}
	if matchWildcard("sub.example.com", "example.com") {
		t.Error("matchWildcard should return false for non-wildcard pattern")
	}
}

func TestMatchWildcard_NoWildcardPrefix(t *testing.T) {
	if matchWildcard("sub.example.com", "example.com") {
		t.Error("matchWildcard should return false when pattern doesn't start with *.")
	}
}

func TestMatchWildcard_EmptyStrings(t *testing.T) {
	if matchWildcard("", "*.example.com") {
		t.Error("matchWildcard should return false for empty host")
	}
	if matchWildcard("example.com", "") {
		t.Error("matchWildcard should return false for empty pattern")
	}
	if matchWildcard("", "") {
		t.Error("matchWildcard should return false for both empty")
	}
}

func TestMatchWildcard_ExactDomainMatch(t *testing.T) {
	// "*.example.com" should match ".example.com" but not "example.com"
	if matchWildcard("example.com", "*.example.com") {
		t.Error("matchWildcard should not match exact domain (only subdomains)")
	}
	// But it should match ".example.com"
	if !matchWildcard(".example.com", "*.example.com") {
		t.Error("matchWildcard should match .example.com")
	}
}
