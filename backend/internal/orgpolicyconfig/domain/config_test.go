package domain

import (
	"testing"
)

func TestDefaultAuthMfa(t *testing.T) {
	authMfa := DefaultAuthMfa()
	if authMfa.MfaRequirement != "new_device" {
		t.Errorf("MfaRequirement = %q, want %q", authMfa.MfaRequirement, "new_device")
	}
	if len(authMfa.AllowedMfaMethods) != 1 || authMfa.AllowedMfaMethods[0] != "sms_otp" {
		t.Errorf("AllowedMfaMethods = %v, want [sms_otp]", authMfa.AllowedMfaMethods)
	}
	if authMfa.StepUpSensitiveActions {
		t.Error("StepUpSensitiveActions should be false by default")
	}
	if authMfa.StepUpPolicyViolation {
		t.Error("StepUpPolicyViolation should be false by default")
	}
}

func TestDefaultDeviceTrust(t *testing.T) {
	deviceTrust := DefaultDeviceTrust()
	if !deviceTrust.DeviceRegistrationAllowed {
		t.Error("DeviceRegistrationAllowed should be true by default")
	}
	if !deviceTrust.AutoTrustAfterMfa {
		t.Error("AutoTrustAfterMfa should be true by default")
	}
	if deviceTrust.MaxTrustedDevicesPerUser != 0 {
		t.Errorf("MaxTrustedDevicesPerUser = %d, want 0 (unlimited)", deviceTrust.MaxTrustedDevicesPerUser)
	}
	if deviceTrust.ReverifyIntervalDays != 30 {
		t.Errorf("ReverifyIntervalDays = %d, want 30", deviceTrust.ReverifyIntervalDays)
	}
	if !deviceTrust.AdminRevokeAllowed {
		t.Error("AdminRevokeAllowed should be true by default")
	}
}

func TestDefaultSessionMgmt(t *testing.T) {
	sessionMgmt := DefaultSessionMgmt()
	if sessionMgmt.SessionMaxTtl != "24h" {
		t.Errorf("SessionMaxTtl = %q, want %q", sessionMgmt.SessionMaxTtl, "24h")
	}
	if sessionMgmt.IdleTimeout != "30m" {
		t.Errorf("IdleTimeout = %q, want %q", sessionMgmt.IdleTimeout, "30m")
	}
	if sessionMgmt.ConcurrentSessionLimit != 0 {
		t.Errorf("ConcurrentSessionLimit = %d, want 0 (unlimited)", sessionMgmt.ConcurrentSessionLimit)
	}
	if !sessionMgmt.AdminForcedLogout {
		t.Error("AdminForcedLogout should be true by default")
	}
	if sessionMgmt.ReauthOnPolicyChange {
		t.Error("ReauthOnPolicyChange should be false by default")
	}
}

func TestDefaultAccessControl(t *testing.T) {
	accessControl := DefaultAccessControl()
	if accessControl.AllowedDomains != nil && len(accessControl.AllowedDomains) != 0 {
		t.Errorf("AllowedDomains = %v, want nil or empty", accessControl.AllowedDomains)
	}
	if accessControl.BlockedDomains != nil && len(accessControl.BlockedDomains) != 0 {
		t.Errorf("BlockedDomains = %v, want nil or empty", accessControl.BlockedDomains)
	}
	if accessControl.WildcardSupported {
		t.Error("WildcardSupported should be false by default")
	}
	if accessControl.DefaultAction != "allow" {
		t.Errorf("DefaultAction = %q, want %q", accessControl.DefaultAction, "allow")
	}
}

func TestDefaultActionRestrictions(t *testing.T) {
	actionRestrictions := DefaultActionRestrictions()
	expectedActions := []string{"navigate", "download", "upload", "copy_paste"}
	if len(actionRestrictions.AllowedActions) != len(expectedActions) {
		t.Errorf("AllowedActions length = %d, want %d", len(actionRestrictions.AllowedActions), len(expectedActions))
	}
	for i, action := range expectedActions {
		if i >= len(actionRestrictions.AllowedActions) || actionRestrictions.AllowedActions[i] != action {
			t.Errorf("AllowedActions[%d] = %q, want %q", i, actionRestrictions.AllowedActions[i], action)
		}
	}
	if actionRestrictions.ReadOnlyMode {
		t.Error("ReadOnlyMode should be false by default")
	}
}

func TestMergeWithDefaults_NilConfig(t *testing.T) {
	result := MergeWithDefaults(nil)
	if result == nil {
		t.Fatal("MergeWithDefaults(nil) should return non-nil config")
	}
	if result.AuthMfa == nil {
		t.Error("AuthMfa should be set")
	}
	if result.DeviceTrust == nil {
		t.Error("DeviceTrust should be set")
	}
	if result.SessionMgmt == nil {
		t.Error("SessionMgmt should be set")
	}
	if result.AccessControl == nil {
		t.Error("AccessControl should be set")
	}
	if result.ActionRestrictions == nil {
		t.Error("ActionRestrictions should be set")
	}
	// Verify defaults
	if result.AuthMfa.MfaRequirement != "new_device" {
		t.Errorf("AuthMfa.MfaRequirement = %q, want %q", result.AuthMfa.MfaRequirement, "new_device")
	}
	if result.AccessControl.DefaultAction != "allow" {
		t.Errorf("AccessControl.DefaultAction = %q, want %q", result.AccessControl.DefaultAction, "allow")
	}
}

func TestMergeWithDefaults_PartialConfig(t *testing.T) {
	customAuthMfa := AuthMfa{
		MfaRequirement:    "always",
		AllowedMfaMethods: []string{"sms_otp", "totp"},
	}
	config := &OrgPolicyConfig{
		AuthMfa: &customAuthMfa,
		// Other sections are nil
	}

	result := MergeWithDefaults(config)
	if result == nil {
		t.Fatal("MergeWithDefaults should return non-nil config")
	}
	// Custom AuthMfa should be preserved
	if result.AuthMfa.MfaRequirement != "always" {
		t.Errorf("AuthMfa.MfaRequirement = %q, want %q", result.AuthMfa.MfaRequirement, "always")
	}
	if len(result.AuthMfa.AllowedMfaMethods) != 2 {
		t.Errorf("AllowedMfaMethods length = %d, want 2", len(result.AuthMfa.AllowedMfaMethods))
	}
	// Other sections should use defaults
	if result.DeviceTrust == nil {
		t.Error("DeviceTrust should be set from defaults")
	}
	if result.SessionMgmt == nil {
		t.Error("SessionMgmt should be set from defaults")
	}
	if result.AccessControl == nil {
		t.Error("AccessControl should be set from defaults")
	}
	if result.ActionRestrictions == nil {
		t.Error("ActionRestrictions should be set from defaults")
	}
}

func TestMergeWithDefaults_FullConfig(t *testing.T) {
	customAuthMfa := AuthMfa{MfaRequirement: "always"}
	customDeviceTrust := DeviceTrust{MaxTrustedDevicesPerUser: 5}
	customSessionMgmt := SessionMgmt{SessionMaxTtl: "12h"}
	customAccessControl := AccessControl{DefaultAction: "deny"}
	customActionRestrictions := ActionRestrictions{ReadOnlyMode: true}

	config := &OrgPolicyConfig{
		AuthMfa:            &customAuthMfa,
		DeviceTrust:        &customDeviceTrust,
		SessionMgmt:        &customSessionMgmt,
		AccessControl:      &customAccessControl,
		ActionRestrictions: &customActionRestrictions,
	}

	result := MergeWithDefaults(config)
	if result == nil {
		t.Fatal("MergeWithDefaults should return non-nil config")
	}
	// All custom values should be preserved
	if result.AuthMfa.MfaRequirement != "always" {
		t.Errorf("AuthMfa.MfaRequirement = %q, want %q", result.AuthMfa.MfaRequirement, "always")
	}
	if result.DeviceTrust.MaxTrustedDevicesPerUser != 5 {
		t.Errorf("DeviceTrust.MaxTrustedDevicesPerUser = %d, want 5", result.DeviceTrust.MaxTrustedDevicesPerUser)
	}
	if result.SessionMgmt.SessionMaxTtl != "12h" {
		t.Errorf("SessionMgmt.SessionMaxTtl = %q, want %q", result.SessionMgmt.SessionMaxTtl, "12h")
	}
	if result.AccessControl.DefaultAction != "deny" {
		t.Errorf("AccessControl.DefaultAction = %q, want %q", result.AccessControl.DefaultAction, "deny")
	}
	if !result.ActionRestrictions.ReadOnlyMode {
		t.Error("ActionRestrictions.ReadOnlyMode should be true")
	}
}

func TestMergeWithDefaults_EmptySections(t *testing.T) {
	// Config with empty (zero-value) sections should still use defaults
	config := &OrgPolicyConfig{
		AuthMfa:            &AuthMfa{},
		DeviceTrust:        &DeviceTrust{},
		SessionMgmt:        &SessionMgmt{},
		AccessControl:      &AccessControl{},
		ActionRestrictions: &ActionRestrictions{},
	}

	result := MergeWithDefaults(config)
	if result == nil {
		t.Fatal("MergeWithDefaults should return non-nil config")
	}
	// Empty sections should be preserved (not replaced with defaults)
	if result.AuthMfa == nil {
		t.Error("AuthMfa should be preserved even if empty")
	}
	if result.DeviceTrust == nil {
		t.Error("DeviceTrust should be preserved even if empty")
	}
}

func TestPtr(t *testing.T) {
	// Test the ptr helper function
	val := "test"
	result := ptr(val)
	if result == nil {
		t.Fatal("ptr should return non-nil pointer")
	}
	if *result != val {
		t.Errorf("ptr result = %q, want %q", *result, val)
	}

	// Test with int
	intVal := 42
	intResult := ptr(intVal)
	if *intResult != intVal {
		t.Errorf("ptr int result = %d, want %d", *intResult, intVal)
	}

	// Test with struct
	structVal := AuthMfa{MfaRequirement: "test"}
	structResult := ptr(structVal)
	if structResult.MfaRequirement != "test" {
		t.Errorf("ptr struct result = %q, want %q", structResult.MfaRequirement, "test")
	}
}
