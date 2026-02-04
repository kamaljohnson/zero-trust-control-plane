package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	devicedomain "zero-trust-control-plane/backend/internal/device/domain"
	orgmfasettingsdomain "zero-trust-control-plane/backend/internal/orgmfasettings/domain"
	platformdomain "zero-trust-control-plane/backend/internal/platformsettings/domain"
	"zero-trust-control-plane/backend/internal/policy/domain"
	"zero-trust-control-plane/backend/internal/policy/repository"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

func TestOPAEvaluator_HealthCheck(t *testing.T) {
	// OPAEvaluator needs a policy repo for NewOPAEvaluator; HealthCheck does not use it.
	e := NewOPAEvaluator(nil)
	ctx := context.Background()
	if err := e.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

// mockPolicyRepo implements repository.Repository for tests.
type mockPolicyRepo struct {
	policies map[string][]*domain.Policy
	err      error
}

var _ repository.Repository = (*mockPolicyRepo)(nil)

func (m *mockPolicyRepo) GetByID(ctx context.Context, id string) (*domain.Policy, error) {
	return nil, nil
}

func (m *mockPolicyRepo) ListByOrg(ctx context.Context, orgID string) ([]*domain.Policy, error) {
	return nil, nil
}

func (m *mockPolicyRepo) GetEnabledPoliciesByOrg(ctx context.Context, orgID string) ([]*domain.Policy, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.policies == nil {
		return nil, nil
	}
	return m.policies[orgID], nil
}

func (m *mockPolicyRepo) Create(ctx context.Context, p *domain.Policy) error {
	return nil
}

func (m *mockPolicyRepo) Update(ctx context.Context, p *domain.Policy) error {
	return nil
}

func (m *mockPolicyRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func TestOPAEvaluator_EvaluateMFA_DefaultPolicy(t *testing.T) {
	// Need a mock repo (can be empty) to avoid nil pointer dereference
	repo := &mockPolicyRepo{
		policies: make(map[string][]*domain.Policy),
	}
	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	// Test with default policy (org settings with no enabled policies)
	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}
	result, err := e.EvaluateMFA(ctx, nil, orgSettings, nil, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	if result.MFARequired {
		t.Error("MFARequired should be false with default policy and no triggers")
	}
	if !result.RegisterTrustAfterMFA {
		t.Error("RegisterTrustAfterMFA should be true by default")
	}
	if result.TrustTTLDays != 30 {
		t.Errorf("TrustTTLDays = %d, want 30", result.TrustTTLDays)
	}
}

func TestOPAEvaluator_EvaluateMFA_NewDevice(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string][]*domain.Policy),
	}
	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: true,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	// New device should require MFA
	result, err := e.EvaluateMFA(ctx, nil, orgSettings, nil, nil, true)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	if !result.MFARequired {
		t.Error("MFARequired should be true for new device")
	}
}

func TestOPAEvaluator_EvaluateMFA_UntrustedDevice(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string][]*domain.Policy),
	}
	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: true,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	device := &devicedomain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp1",
		Trusted:     false,
		CreatedAt:   time.Now().UTC(),
	}

	// Untrusted device should require MFA
	result, err := e.EvaluateMFA(ctx, nil, orgSettings, device, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	if !result.MFARequired {
		t.Error("MFARequired should be true for untrusted device")
	}
}

func TestOPAEvaluator_EvaluateMFA_PlatformMFAAlways(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string][]*domain.Policy),
	}
	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	platformSettings := &platformdomain.PlatformDeviceTrustSettings{
		MFARequiredAlways:   true,
		DefaultTrustTTLDays: 30,
	}

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	// Platform MFA always should require MFA
	result, err := e.EvaluateMFA(ctx, platformSettings, orgSettings, nil, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	if !result.MFARequired {
		t.Error("MFARequired should be true when platform requires MFA always")
	}
}

func TestOPAEvaluator_EvaluateMFA_CustomPolicy(t *testing.T) {
	customPolicy := `package ztcp.device_trust

default mfa_required = true
default register_trust_after_mfa = false
default trust_ttl_days = 60
`

	repo := &mockPolicyRepo{
		policies: map[string][]*domain.Policy{
			"org-1": {
				{
					ID:      "policy-1",
					OrgID:   "org-1",
					Enabled: true,
					Rules:   customPolicy,
				},
			},
		},
	}

	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	result, err := e.EvaluateMFA(ctx, nil, orgSettings, nil, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	if !result.MFARequired {
		t.Error("MFARequired should be true with custom policy")
	}
	if result.RegisterTrustAfterMFA {
		t.Error("RegisterTrustAfterMFA should be false with custom policy")
	}
	if result.TrustTTLDays != 60 {
		t.Errorf("TrustTTLDays = %d, want 60", result.TrustTTLDays)
	}
}

func TestOPAEvaluator_EvaluateMFA_PolicyRepoError(t *testing.T) {
	repo := &mockPolicyRepo{
		err: errors.New("database error"),
	}

	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	// Should fallback to default policy on error
	result, err := e.EvaluateMFA(ctx, nil, orgSettings, nil, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA should not return error on repo error: %v", err)
	}
	if result.MFARequired {
		t.Error("MFARequired should be false with default policy")
	}
}

func TestOPAEvaluator_EvaluateMFA_DeviceWithTimestamps(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string][]*domain.Policy),
	}
	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	now := time.Now().UTC()
	trustedUntil := now.Add(24 * time.Hour)
	revokedAt := now.Add(-1 * time.Hour)

	device := &devicedomain.Device{
		ID:          "device-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		Fingerprint: "fp1",
		Trusted:     true,
		TrustedUntil: &trustedUntil,
		RevokedAt:   &revokedAt,
		CreatedAt:   now,
	}

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: true,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	// Revoked device should require MFA (is_effectively_trusted = false)
	result, err := e.EvaluateMFA(ctx, nil, orgSettings, device, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	if !result.MFARequired {
		t.Error("MFARequired should be true for revoked device")
	}
}

func TestOPAEvaluator_EvaluateMFA_UserWithPhone(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string][]*domain.Policy),
	}
	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	user := &userdomain.User{
		ID:            "user-1",
		Email:         "user@example.com",
		Name:          "Test User",
		Phone:         "+1234567890",
		PhoneVerified: true,
		Status:        userdomain.UserStatusActive,
		CreatedAt:     time.Now().UTC(),
	}

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: true,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	result, err := e.EvaluateMFA(ctx, nil, orgSettings, nil, user, true)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	if !result.MFARequired {
		t.Error("MFARequired should be true for new device")
	}
}

func TestOPAEvaluator_EvaluateMFA_PlatformTTLOverride(t *testing.T) {
	repo := &mockPolicyRepo{
		policies: make(map[string][]*domain.Policy),
	}
	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	platformSettings := &platformdomain.PlatformDeviceTrustSettings{
		MFARequiredAlways:   false,
		DefaultTrustTTLDays: 60,
	}

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            0, // Should use platform default
	}

	result, err := e.EvaluateMFA(ctx, platformSettings, orgSettings, nil, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA: %v", err)
	}
	// The default policy uses org.trust_ttl_days which should fallback to platform default
	// But since org.trust_ttl_days is 0, it should use platform default
	if result.TrustTTLDays != 60 {
		t.Errorf("TrustTTLDays = %d, want 60 (platform default)", result.TrustTTLDays)
	}
}

func TestOPAEvaluator_EvaluateMFA_InvalidPolicy(t *testing.T) {
	invalidPolicy := `package ztcp.device_trust

invalid syntax here
`

	repo := &mockPolicyRepo{
		policies: map[string][]*domain.Policy{
			"org-1": {
				{
					ID:      "policy-1",
					OrgID:   "org-1",
					Enabled: true,
					Rules:   invalidPolicy,
				},
			},
		},
	}

	e := NewOPAEvaluator(repo)
	ctx := context.Background()

	orgSettings := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   "org-1",
		MFARequiredForNewDevice: false,
		MFARequiredForUntrusted: false,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
	}

	// Should fallback to default result on invalid policy
	result, err := e.EvaluateMFA(ctx, nil, orgSettings, nil, nil, false)
	if err != nil {
		t.Fatalf("EvaluateMFA should not return error on invalid policy: %v", err)
	}
	if result.MFARequired {
		t.Error("MFARequired should be false with default fallback")
	}
}

func TestOPAEvaluator_defaultResult(t *testing.T) {
	e := NewOPAEvaluator(&mockPolicyRepo{})

	// Test with nil platform settings
	result := e.defaultResult(nil)
	if result.MFARequired {
		t.Error("MFARequired should be false")
	}
	if !result.RegisterTrustAfterMFA {
		t.Error("RegisterTrustAfterMFA should be true")
	}
	if result.TrustTTLDays != 30 {
		t.Errorf("TrustTTLDays = %d, want 30", result.TrustTTLDays)
	}

	// Test with platform settings
	platformSettings := &platformdomain.PlatformDeviceTrustSettings{
		MFARequiredAlways:   false,
		DefaultTrustTTLDays: 60,
	}
	result = e.defaultResult(platformSettings)
	if result.TrustTTLDays != 60 {
		t.Errorf("TrustTTLDays = %d, want 60", result.TrustTTLDays)
	}
}
