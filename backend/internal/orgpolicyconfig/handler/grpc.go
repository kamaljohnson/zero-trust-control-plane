package handler

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orgpolicyconfigv1 "zero-trust-control-plane/backend/api/generated/orgpolicyconfig/v1"
	membershiprepo "zero-trust-control-plane/backend/internal/membership/repository"
	orgmfasettingsdomain "zero-trust-control-plane/backend/internal/orgmfasettings/domain"
	orgmfasettingsrepo "zero-trust-control-plane/backend/internal/orgmfasettings/repository"
	"zero-trust-control-plane/backend/internal/orgpolicyconfig/domain"
	"zero-trust-control-plane/backend/internal/orgpolicyconfig/repository"
	"zero-trust-control-plane/backend/internal/platform/rbac"
)

// Server implements OrgPolicyConfigService. Caller must be org admin or owner.
type Server struct {
	orgpolicyconfigv1.UnimplementedOrgPolicyConfigServiceServer
	repo               repository.Repository
	membershipRepo     membershiprepo.Repository
	orgMfaSettingsRepo orgmfasettingsrepo.Repository
}

// NewServer returns a new OrgPolicyConfig gRPC server.
func NewServer(
	repo repository.Repository,
	membershipRepo membershiprepo.Repository,
	orgMfaSettingsRepo orgmfasettingsrepo.Repository,
) *Server {
	return &Server{
		repo:               repo,
		membershipRepo:     membershipRepo,
		orgMfaSettingsRepo: orgMfaSettingsRepo,
	}
}

// GetOrgPolicyConfig returns the org policy config for the caller's org. Caller must be org admin or owner.
func (s *Server) GetOrgPolicyConfig(ctx context.Context, req *orgpolicyconfigv1.GetOrgPolicyConfigRequest) (*orgpolicyconfigv1.GetOrgPolicyConfigResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method GetOrgPolicyConfig not implemented")
	}
	orgID, _, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	requestOrgID := req.GetOrgId()
	if requestOrgID != "" && requestOrgID != orgID {
		return nil, status.Error(codes.PermissionDenied, "org_id does not match your organization")
	}
	useOrgID := orgID
	if useOrgID == "" {
		useOrgID = requestOrgID
	}
	if useOrgID == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	config, err := s.repo.GetByOrgID(ctx, useOrgID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	merged := domain.MergeWithDefaults(config)
	return &orgpolicyconfigv1.GetOrgPolicyConfigResponse{
		Config: domainToProto(merged),
	}, nil
}

// UpdateOrgPolicyConfig updates the org policy config. Caller must be org admin or owner. Syncs auth_mfa and device_trust to org_mfa_settings.
func (s *Server) UpdateOrgPolicyConfig(ctx context.Context, req *orgpolicyconfigv1.UpdateOrgPolicyConfigRequest) (*orgpolicyconfigv1.UpdateOrgPolicyConfigResponse, error) {
	if s.repo == nil {
		return nil, status.Error(codes.Unimplemented, "method UpdateOrgPolicyConfig not implemented")
	}
	orgID, _, err := rbac.RequireOrgAdmin(ctx, s.membershipRepo)
	if err != nil {
		return nil, err
	}
	requestOrgID := req.GetOrgId()
	if requestOrgID != "" && requestOrgID != orgID {
		return nil, status.Error(codes.PermissionDenied, "org_id does not match your organization")
	}
	useOrgID := orgID
	if useOrgID == "" {
		useOrgID = requestOrgID
	}
	if useOrgID == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	config := protoToDomain(req.GetConfig())
	if err := s.repo.Upsert(ctx, useOrgID, config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Sync auth_mfa and device_trust to org_mfa_settings so auth_service and policy engine keep working.
	if s.orgMfaSettingsRepo != nil && (config.AuthMfa != nil || config.DeviceTrust != nil) {
		merged := domain.MergeWithDefaults(config)
		settings := domainToOrgMFASettings(useOrgID, merged)
		if err := s.orgMfaSettingsRepo.Upsert(ctx, settings); err != nil {
			return nil, status.Error(codes.Internal, "failed to sync org MFA settings: "+err.Error())
		}
	}
	updated := domain.MergeWithDefaults(config)
	return &orgpolicyconfigv1.UpdateOrgPolicyConfigResponse{
		Config: domainToProto(updated),
	}, nil
}

// domainToOrgMFASettings maps policy config auth_mfa and device_trust to OrgMFASettings for upsert.
func domainToOrgMFASettings(orgID string, c *domain.OrgPolicyConfig) *orgmfasettingsdomain.OrgMFASettings {
	now := time.Now().UTC()
	s := &orgmfasettingsdomain.OrgMFASettings{
		OrgID:                   orgID,
		MFARequiredForNewDevice: true,
		MFARequiredForUntrusted: true,
		MFARequiredAlways:       false,
		RegisterTrustAfterMFA:   true,
		TrustTTLDays:            30,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if c.AuthMfa != nil {
		switch c.AuthMfa.MfaRequirement {
		case "always":
			s.MFARequiredAlways = true
			s.MFARequiredForNewDevice = false
			s.MFARequiredForUntrusted = false
		case "new_device":
			s.MFARequiredForNewDevice = true
			s.MFARequiredForUntrusted = true
			s.MFARequiredAlways = false
		case "untrusted":
			s.MFARequiredForUntrusted = true
			s.MFARequiredForNewDevice = false
			s.MFARequiredAlways = false
		}
	}
	if c.DeviceTrust != nil {
		s.RegisterTrustAfterMFA = c.DeviceTrust.AutoTrustAfterMfa
		if c.DeviceTrust.ReverifyIntervalDays > 0 {
			s.TrustTTLDays = c.DeviceTrust.ReverifyIntervalDays
		}
	}
	return s
}

func domainToProto(c *domain.OrgPolicyConfig) *orgpolicyconfigv1.OrgPolicyConfig {
	if c == nil {
		return nil
	}
	out := &orgpolicyconfigv1.OrgPolicyConfig{}
	if c.AuthMfa != nil {
		out.AuthMfa = &orgpolicyconfigv1.AuthMfa{
			MfaRequirement:         mfaRequirementToProto(c.AuthMfa.MfaRequirement),
			AllowedMfaMethods:      append([]string(nil), c.AuthMfa.AllowedMfaMethods...),
			StepUpSensitiveActions: c.AuthMfa.StepUpSensitiveActions,
			StepUpPolicyViolation:  c.AuthMfa.StepUpPolicyViolation,
		}
	}
	if c.DeviceTrust != nil {
		out.DeviceTrust = &orgpolicyconfigv1.DeviceTrust{
			DeviceRegistrationAllowed: c.DeviceTrust.DeviceRegistrationAllowed,
			AutoTrustAfterMfa:         c.DeviceTrust.AutoTrustAfterMfa,
			MaxTrustedDevicesPerUser:  int32(c.DeviceTrust.MaxTrustedDevicesPerUser),
			ReverifyIntervalDays:      int32(c.DeviceTrust.ReverifyIntervalDays),
			AdminRevokeAllowed:        c.DeviceTrust.AdminRevokeAllowed,
		}
	}
	if c.SessionMgmt != nil {
		out.SessionMgmt = &orgpolicyconfigv1.SessionMgmt{
			SessionMaxTtl:          c.SessionMgmt.SessionMaxTtl,
			IdleTimeout:            c.SessionMgmt.IdleTimeout,
			ConcurrentSessionLimit: int32(c.SessionMgmt.ConcurrentSessionLimit),
			AdminForcedLogout:      c.SessionMgmt.AdminForcedLogout,
			ReauthOnPolicyChange:   c.SessionMgmt.ReauthOnPolicyChange,
		}
	}
	if c.AccessControl != nil {
		out.AccessControl = &orgpolicyconfigv1.AccessControl{
			AllowedDomains:    append([]string(nil), c.AccessControl.AllowedDomains...),
			BlockedDomains:    append([]string(nil), c.AccessControl.BlockedDomains...),
			WildcardSupported: c.AccessControl.WildcardSupported,
			DefaultAction:     defaultActionToProto(c.AccessControl.DefaultAction),
		}
	}
	if c.ActionRestrictions != nil {
		out.ActionRestrictions = &orgpolicyconfigv1.ActionRestrictions{
			AllowedActions: append([]string(nil), c.ActionRestrictions.AllowedActions...),
			ReadOnlyMode:   c.ActionRestrictions.ReadOnlyMode,
		}
	}
	return out
}

func mfaRequirementToProto(s string) orgpolicyconfigv1.MfaRequirement {
	switch s {
	case "always":
		return orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_ALWAYS
	case "new_device":
		return orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_NEW_DEVICE
	case "untrusted":
		return orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNTRUSTED
	default:
		return orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNSPECIFIED
	}
}

func defaultActionToProto(s string) orgpolicyconfigv1.DefaultAction {
	switch s {
	case "deny":
		return orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_DENY
	case "allow":
		return orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_ALLOW
	default:
		return orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_UNSPECIFIED
	}
}

func protoToDomain(p *orgpolicyconfigv1.OrgPolicyConfig) *domain.OrgPolicyConfig {
	if p == nil {
		return nil
	}
	out := &domain.OrgPolicyConfig{}
	if p.AuthMfa != nil {
		out.AuthMfa = &domain.AuthMfa{
			MfaRequirement:         mfaRequirementToDomain(p.AuthMfa.GetMfaRequirement()),
			AllowedMfaMethods:      append([]string(nil), p.AuthMfa.GetAllowedMfaMethods()...),
			StepUpSensitiveActions: p.AuthMfa.GetStepUpSensitiveActions(),
			StepUpPolicyViolation:  p.AuthMfa.GetStepUpPolicyViolation(),
		}
	}
	if p.DeviceTrust != nil {
		out.DeviceTrust = &domain.DeviceTrust{
			DeviceRegistrationAllowed: p.DeviceTrust.GetDeviceRegistrationAllowed(),
			AutoTrustAfterMfa:         p.DeviceTrust.GetAutoTrustAfterMfa(),
			MaxTrustedDevicesPerUser:  int(p.DeviceTrust.GetMaxTrustedDevicesPerUser()),
			ReverifyIntervalDays:      int(p.DeviceTrust.GetReverifyIntervalDays()),
			AdminRevokeAllowed:        p.DeviceTrust.GetAdminRevokeAllowed(),
		}
	}
	if p.SessionMgmt != nil {
		out.SessionMgmt = &domain.SessionMgmt{
			SessionMaxTtl:          p.SessionMgmt.GetSessionMaxTtl(),
			IdleTimeout:            p.SessionMgmt.GetIdleTimeout(),
			ConcurrentSessionLimit: int(p.SessionMgmt.GetConcurrentSessionLimit()),
			AdminForcedLogout:      p.SessionMgmt.GetAdminForcedLogout(),
			ReauthOnPolicyChange:   p.SessionMgmt.GetReauthOnPolicyChange(),
		}
	}
	if p.AccessControl != nil {
		out.AccessControl = &domain.AccessControl{
			AllowedDomains:    append([]string(nil), p.AccessControl.GetAllowedDomains()...),
			BlockedDomains:    append([]string(nil), p.AccessControl.GetBlockedDomains()...),
			WildcardSupported: p.AccessControl.GetWildcardSupported(),
			DefaultAction:     defaultActionToDomain(p.AccessControl.GetDefaultAction()),
		}
	}
	if p.ActionRestrictions != nil {
		out.ActionRestrictions = &domain.ActionRestrictions{
			AllowedActions: append([]string(nil), p.ActionRestrictions.GetAllowedActions()...),
			ReadOnlyMode:   p.ActionRestrictions.GetReadOnlyMode(),
		}
	}
	return out
}

func mfaRequirementToDomain(e orgpolicyconfigv1.MfaRequirement) string {
	switch e {
	case orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_ALWAYS:
		return "always"
	case orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_NEW_DEVICE:
		return "new_device"
	case orgpolicyconfigv1.MfaRequirement_MFA_REQUIREMENT_UNTRUSTED:
		return "untrusted"
	default:
		return "new_device"
	}
}

func defaultActionToDomain(e orgpolicyconfigv1.DefaultAction) string {
	switch e {
	case orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_DENY:
		return "deny"
	case orgpolicyconfigv1.DefaultAction_DEFAULT_ACTION_ALLOW:
		return "allow"
	default:
		return "allow"
	}
}
