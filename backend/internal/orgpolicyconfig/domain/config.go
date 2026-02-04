package domain

// AuthMfa holds org-level auth/MFA policy.
type AuthMfa struct {
	MfaRequirement         string   `json:"mfa_requirement"`     // always, new_device, untrusted
	AllowedMfaMethods      []string `json:"allowed_mfa_methods"` // e.g. sms_otp
	StepUpSensitiveActions bool     `json:"step_up_sensitive_actions"`
	StepUpPolicyViolation  bool     `json:"step_up_policy_violation"`
}

// DeviceTrust holds org-level device trust policy.
type DeviceTrust struct {
	DeviceRegistrationAllowed bool `json:"device_registration_allowed"`
	AutoTrustAfterMfa         bool `json:"auto_trust_after_mfa"`
	MaxTrustedDevicesPerUser  int  `json:"max_trusted_devices_per_user"` // 0 = unlimited
	ReverifyIntervalDays      int  `json:"reverify_interval_days"`
	AdminRevokeAllowed        bool `json:"admin_revoke_allowed"`
}

// SessionMgmt holds org-level session policy.
type SessionMgmt struct {
	SessionMaxTtl          string `json:"session_max_ttl"`          // e.g. "24h"
	IdleTimeout            string `json:"idle_timeout"`             // e.g. "30m"
	ConcurrentSessionLimit int    `json:"concurrent_session_limit"` // 0 = unlimited
	AdminForcedLogout      bool   `json:"admin_forced_logout"`
	ReauthOnPolicyChange   bool   `json:"reauth_on_policy_change"`
}

// AccessControl holds org-level access control (browser) policy.
type AccessControl struct {
	AllowedDomains    []string `json:"allowed_domains"`
	BlockedDomains    []string `json:"blocked_domains"`
	WildcardSupported bool     `json:"wildcard_supported"`
	DefaultAction     string   `json:"default_action"` // allow, deny
}

// ActionRestrictions holds org-level action restrictions.
type ActionRestrictions struct {
	AllowedActions []string `json:"allowed_actions"` // navigate, download, upload, copy_paste
	ReadOnlyMode   bool     `json:"read_only_mode"`
}

// OrgPolicyConfig holds all five sections. Used for JSON storage and API.
type OrgPolicyConfig struct {
	AuthMfa            *AuthMfa            `json:"auth_mfa,omitempty"`
	DeviceTrust        *DeviceTrust        `json:"device_trust,omitempty"`
	SessionMgmt        *SessionMgmt        `json:"session_mgmt,omitempty"`
	AccessControl      *AccessControl      `json:"access_control,omitempty"`
	ActionRestrictions *ActionRestrictions `json:"action_restrictions,omitempty"`
}

// DefaultAuthMfa returns default AuthMfa (MFA on new device, SMS OTP allowed).
func DefaultAuthMfa() AuthMfa {
	return AuthMfa{
		MfaRequirement:         "new_device",
		AllowedMfaMethods:      []string{"sms_otp"},
		StepUpSensitiveActions: false,
		StepUpPolicyViolation:  false,
	}
}

// DefaultDeviceTrust returns default DeviceTrust (registration allowed, auto-trust after MFA).
func DefaultDeviceTrust() DeviceTrust {
	return DeviceTrust{
		DeviceRegistrationAllowed: true,
		AutoTrustAfterMfa:         true,
		MaxTrustedDevicesPerUser:  0,
		ReverifyIntervalDays:      30,
		AdminRevokeAllowed:        true,
	}
}

// DefaultSessionMgmt returns default SessionMgmt.
func DefaultSessionMgmt() SessionMgmt {
	return SessionMgmt{
		SessionMaxTtl:          "24h",
		IdleTimeout:            "30m",
		ConcurrentSessionLimit: 0,
		AdminForcedLogout:      true,
		ReauthOnPolicyChange:   false,
	}
}

// DefaultAccessControl returns default AccessControl (allow).
func DefaultAccessControl() AccessControl {
	return AccessControl{
		AllowedDomains:    nil,
		BlockedDomains:    nil,
		WildcardSupported: false,
		DefaultAction:     "allow",
	}
}

// DefaultActionRestrictions returns default ActionRestrictions.
func DefaultActionRestrictions() ActionRestrictions {
	return ActionRestrictions{
		AllowedActions: []string{"navigate", "download", "upload", "copy_paste"},
		ReadOnlyMode:   false,
	}
}

// MergeWithDefaults returns a copy of c with nil sections replaced by defaults.
func MergeWithDefaults(c *OrgPolicyConfig) *OrgPolicyConfig {
	if c == nil {
		return &OrgPolicyConfig{
			AuthMfa:            ptr(DefaultAuthMfa()),
			DeviceTrust:        ptr(DefaultDeviceTrust()),
			SessionMgmt:        ptr(DefaultSessionMgmt()),
			AccessControl:      ptr(DefaultAccessControl()),
			ActionRestrictions: ptr(DefaultActionRestrictions()),
		}
	}
	out := *c
	if out.AuthMfa == nil {
		out.AuthMfa = ptr(DefaultAuthMfa())
	}
	if out.DeviceTrust == nil {
		out.DeviceTrust = ptr(DefaultDeviceTrust())
	}
	if out.SessionMgmt == nil {
		out.SessionMgmt = ptr(DefaultSessionMgmt())
	}
	if out.AccessControl == nil {
		out.AccessControl = ptr(DefaultAccessControl())
	}
	if out.ActionRestrictions == nil {
		out.ActionRestrictions = ptr(DefaultActionRestrictions())
	}
	return &out
}

func ptr[T any](v T) *T { return &v }
