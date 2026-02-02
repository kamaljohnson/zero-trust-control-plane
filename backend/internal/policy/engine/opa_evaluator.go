package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"

	devicedomain "zero-trust-control-plane/backend/internal/device/domain"
	orgmfasettingsdomain "zero-trust-control-plane/backend/internal/orgmfasettings/domain"
	platformdomain "zero-trust-control-plane/backend/internal/platformsettings/domain"
	"zero-trust-control-plane/backend/internal/policy/repository"
	userdomain "zero-trust-control-plane/backend/internal/user/domain"
)

const defaultPolicyPackage = "ztcp.device_trust"

// Default Rego policy that matches current hardcoded logic (backward compatibility).
const defaultRegoPolicy = `package ztcp.device_trust

default mfa_required = false
default register_trust_after_mfa = true
default trust_ttl_days = 30

mfa_required if {
	input.platform.mfa_required_always
}

mfa_required if {
	input.device.is_new
	input.org.mfa_required_for_new_device
}

mfa_required if {
	not input.device.is_effectively_trusted
	input.org.mfa_required_for_untrusted
}

register_trust_after_mfa = input.org.register_trust_after_mfa if {
	input.org.register_trust_after_mfa != null
}
register_trust_after_mfa = true if {
	not input.org.register_trust_after_mfa
}

trust_ttl_days = input.org.trust_ttl_days if {
	input.org.trust_ttl_days > 0
}
trust_ttl_days = input.platform.default_trust_ttl_days if {
	input.org.trust_ttl_days <= 0
	input.platform.default_trust_ttl_days > 0
}
`

// OPAEvaluator evaluates device-trust/MFA policies using OPA Rego.
type OPAEvaluator struct {
	policyRepo repository.Repository
}

// NewOPAEvaluator returns an OPA-based policy evaluator.
func NewOPAEvaluator(policyRepo repository.Repository) *OPAEvaluator {
	return &OPAEvaluator{policyRepo: policyRepo}
}

// HealthCheck verifies that the in-process OPA Rego engine can compile and evaluate the default policy.
// Does not call the policy repo or database. Returns nil on success.
func (e *OPAEvaluator) HealthCheck(ctx context.Context) error {
	modules := map[string]string{"policy_0.rego": defaultRegoPolicy}
	compiler, err := ast.CompileModules(modules)
	if err != nil {
		return fmt.Errorf("compile default policy: %w", err)
	}
	minimalInput := map[string]interface{}{
		"platform": map[string]interface{}{
			"mfa_required_always":    false,
			"default_trust_ttl_days": 30,
		},
		"org": map[string]interface{}{
			"mfa_required_for_new_device": true,
			"mfa_required_for_untrusted":  true,
			"mfa_required_always":         false,
			"register_trust_after_mfa":    true,
			"trust_ttl_days":              30,
		},
		"device": map[string]interface{}{
			"id":                     "",
			"trusted":                false,
			"trusted_until":          nil,
			"revoked_at":             nil,
			"is_new":                 false,
			"is_effectively_trusted": false,
		},
		"user": map[string]interface{}{
			"id":        "",
			"has_phone": false,
		},
	}
	q := rego.New(
		rego.Query("data.ztcp.device_trust.mfa_required"),
		rego.Compiler(compiler),
		rego.Input(minimalInput),
	)
	rs, err := q.Eval(ctx)
	if err != nil {
		return fmt.Errorf("eval default policy: %w", err)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return fmt.Errorf("policy query returned no result")
	}
	return nil
}

// EvaluateMFA evaluates MFA policy using OPA Rego policies.
func (e *OPAEvaluator) EvaluateMFA(
	ctx context.Context,
	platformSettings *platformdomain.PlatformDeviceTrustSettings,
	orgSettings *orgmfasettingsdomain.OrgMFASettings,
	device *devicedomain.Device,
	user *userdomain.User,
	isNewDevice bool,
) (MFAResult, error) {
	// Build input JSON for OPA
	input, err := e.buildInput(platformSettings, orgSettings, device, user, isNewDevice)
	if err != nil {
		return e.defaultResult(platformSettings), fmt.Errorf("build input: %w", err)
	}

	// Load enabled policies for org
	var policies []string
	if orgSettings != nil {
		enabledPolicies, err := e.policyRepo.GetEnabledPoliciesByOrg(ctx, orgSettings.OrgID)
		if err != nil {
			log.Printf("policy: failed to load policies for org %s: %v", orgSettings.OrgID, err)
		} else {
			for _, p := range enabledPolicies {
				if p.Enabled && p.Rules != "" {
					policies = append(policies, p.Rules)
				}
			}
		}
	}

	// Use default policy if no org policies exist
	if len(policies) == 0 {
		policies = []string{defaultRegoPolicy}
	}

	// Compile and evaluate policies
	result, err := e.evaluatePolicies(ctx, policies, input)
	if err != nil {
		log.Printf("policy: evaluation failed: %v, using defaults", err)
		return e.defaultResult(platformSettings), nil
	}

	return result, nil
}

func (e *OPAEvaluator) buildInput(
	platformSettings *platformdomain.PlatformDeviceTrustSettings,
	orgSettings *orgmfasettingsdomain.OrgMFASettings,
	device *devicedomain.Device,
	user *userdomain.User,
	isNewDevice bool,
) (map[string]interface{}, error) {
	now := time.Now().UTC()
	platform := map[string]interface{}{
		"mfa_required_always":    false,
		"default_trust_ttl_days": 30,
	}
	if platformSettings != nil {
		platform["mfa_required_always"] = platformSettings.MFARequiredAlways
		platform["default_trust_ttl_days"] = platformSettings.DefaultTrustTTLDays
	}

	org := map[string]interface{}{
		"mfa_required_for_new_device": true,
		"mfa_required_for_untrusted":  true,
		"mfa_required_always":         false,
		"register_trust_after_mfa":    true,
		"trust_ttl_days":              30,
	}
	if orgSettings != nil {
		org["mfa_required_for_new_device"] = orgSettings.MFARequiredForNewDevice
		org["mfa_required_for_untrusted"] = orgSettings.MFARequiredForUntrusted
		org["mfa_required_always"] = orgSettings.MFARequiredAlways
		org["register_trust_after_mfa"] = orgSettings.RegisterTrustAfterMFA
		org["trust_ttl_days"] = orgSettings.TrustTTLDays
		if orgSettings.TrustTTLDays <= 0 && platformSettings != nil {
			org["trust_ttl_days"] = platformSettings.DefaultTrustTTLDays
		}
	}

	deviceMap := map[string]interface{}{
		"id":                     "",
		"trusted":                false,
		"trusted_until":          nil,
		"revoked_at":             nil,
		"is_new":                 isNewDevice,
		"is_effectively_trusted": false,
	}
	if device != nil {
		deviceMap["id"] = device.ID
		deviceMap["trusted"] = device.Trusted
		if device.TrustedUntil != nil {
			deviceMap["trusted_until"] = device.TrustedUntil.Format(time.RFC3339)
		}
		if device.RevokedAt != nil {
			deviceMap["revoked_at"] = device.RevokedAt.Format(time.RFC3339)
		}
		deviceMap["is_effectively_trusted"] = device.IsEffectivelyTrusted(now)
	}

	userMap := map[string]interface{}{
		"id":        "",
		"has_phone": false,
	}
	if user != nil {
		userMap["id"] = user.ID
		userMap["has_phone"] = user.Phone != ""
	}

	return map[string]interface{}{
		"platform": platform,
		"org":      org,
		"device":   deviceMap,
		"user":     userMap,
	}, nil
}

func (e *OPAEvaluator) evaluatePolicies(ctx context.Context, policies []string, input map[string]interface{}) (MFAResult, error) {
	// Compile all policies
	modules := make(map[string]string)
	for i, policy := range policies {
		modules[fmt.Sprintf("policy_%d.rego", i)] = policy
	}

	compiler, err := ast.CompileModules(modules)
	if err != nil {
		return MFAResult{}, fmt.Errorf("compile policies: %w", err)
	}

	// Prepare queries for each value
	out := MFAResult{
		MFARequired:           false,
		RegisterTrustAfterMFA: true,
		TrustTTLDays:          30,
	}

	// Query mfa_required
	mfaQuery := rego.New(
		rego.Query("data.ztcp.device_trust.mfa_required"),
		rego.Compiler(compiler),
		rego.Input(input),
	)
	mfaRS, err := mfaQuery.Eval(ctx)
	if err == nil && len(mfaRS) > 0 && len(mfaRS[0].Expressions) > 0 {
		if v, ok := mfaRS[0].Expressions[0].Value.(bool); ok {
			out.MFARequired = v
		}
	}

	// Query register_trust_after_mfa
	registerQuery := rego.New(
		rego.Query("data.ztcp.device_trust.register_trust_after_mfa"),
		rego.Compiler(compiler),
		rego.Input(input),
	)
	registerRS, err := registerQuery.Eval(ctx)
	if err == nil && len(registerRS) > 0 && len(registerRS[0].Expressions) > 0 {
		if v, ok := registerRS[0].Expressions[0].Value.(bool); ok {
			out.RegisterTrustAfterMFA = v
		}
	}

	// Query trust_ttl_days
	ttlQuery := rego.New(
		rego.Query("data.ztcp.device_trust.trust_ttl_days"),
		rego.Compiler(compiler),
		rego.Input(input),
	)
	ttlRS, err := ttlQuery.Eval(ctx)
	if err == nil && len(ttlRS) > 0 && len(ttlRS[0].Expressions) > 0 {
		switch v := ttlRS[0].Expressions[0].Value.(type) {
		case json.Number:
			if days, err := v.Int64(); err == nil && days > 0 {
				out.TrustTTLDays = int(days)
			}
		case float64:
			if days := int(v); days > 0 {
				out.TrustTTLDays = days
			}
		case int64:
			if v > 0 {
				out.TrustTTLDays = int(v)
			}
		}
	}

	return out, nil
}

func (e *OPAEvaluator) defaultResult(platformSettings *platformdomain.PlatformDeviceTrustSettings) MFAResult {
	ttl := 30
	if platformSettings != nil {
		ttl = platformSettings.DefaultTrustTTLDays
	}
	return MFAResult{
		MFARequired:           false,
		RegisterTrustAfterMFA: true,
		TrustTTLDays:          ttl,
	}
}
