package server

import (
	"google.golang.org/grpc"

	adminv1 "zero-trust-control-plane/backend/api/generated/admin/v1"
	auditv1 "zero-trust-control-plane/backend/api/generated/audit/v1"
	authv1 "zero-trust-control-plane/backend/api/generated/auth/v1"
	devv1 "zero-trust-control-plane/backend/api/generated/dev/v1"
	devicev1 "zero-trust-control-plane/backend/api/generated/device/v1"
	healthv1 "zero-trust-control-plane/backend/api/generated/health/v1"
	membershipv1 "zero-trust-control-plane/backend/api/generated/membership/v1"
	organizationv1 "zero-trust-control-plane/backend/api/generated/organization/v1"
	orgpolicyconfigv1 "zero-trust-control-plane/backend/api/generated/orgpolicyconfig/v1"
	policyv1 "zero-trust-control-plane/backend/api/generated/policy/v1"
	sessionv1 "zero-trust-control-plane/backend/api/generated/session/v1"
	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	userv1 "zero-trust-control-plane/backend/api/generated/user/v1"

	adminhandler "zero-trust-control-plane/backend/internal/admin/handler"
	"zero-trust-control-plane/backend/internal/audit"
	audithandler "zero-trust-control-plane/backend/internal/audit/handler"
	auditrepo "zero-trust-control-plane/backend/internal/audit/repository"
	devicehandler "zero-trust-control-plane/backend/internal/device/handler"
	devicerepo "zero-trust-control-plane/backend/internal/device/repository"
	healthhandler "zero-trust-control-plane/backend/internal/health/handler"
	identityhandler "zero-trust-control-plane/backend/internal/identity/handler"
	identityservice "zero-trust-control-plane/backend/internal/identity/service"
	membershiphandler "zero-trust-control-plane/backend/internal/membership/handler"
	membershiprepo "zero-trust-control-plane/backend/internal/membership/repository"
	organizationhandler "zero-trust-control-plane/backend/internal/organization/handler"
	organizationrepo "zero-trust-control-plane/backend/internal/organization/repository"
	orgmfasettingsrepo "zero-trust-control-plane/backend/internal/orgmfasettings/repository"
	orgpolicyconfighandler "zero-trust-control-plane/backend/internal/orgpolicyconfig/handler"
	orgpolicyconfigrepo "zero-trust-control-plane/backend/internal/orgpolicyconfig/repository"
	policyhandler "zero-trust-control-plane/backend/internal/policy/handler"
	policyrepo "zero-trust-control-plane/backend/internal/policy/repository"
	sessionhandler "zero-trust-control-plane/backend/internal/session/handler"
	sessionrepo "zero-trust-control-plane/backend/internal/session/repository"
	"zero-trust-control-plane/backend/internal/telemetry"
	telemetryhandler "zero-trust-control-plane/backend/internal/telemetry/handler"
	userhandler "zero-trust-control-plane/backend/internal/user/handler"
	userrepo "zero-trust-control-plane/backend/internal/user/repository"
)

// Deps holds optional service dependencies for gRPC handlers.
type Deps struct {
	// Auth is the auth service for Register/Login/Refresh/Logout. If nil, auth RPCs return Unimplemented.
	Auth *identityservice.AuthService
	// DeviceRepo is the device repository for DeviceService. If nil, device RPCs return Unimplemented.
	DeviceRepo devicerepo.Repository
	// PolicyRepo is the policy repository for PolicyService. If nil, policy RPCs return Unimplemented.
	PolicyRepo policyrepo.Repository
	// AuditRepo is the audit log repository for AuditService and the audit interceptor. If nil, ListAuditLogs returns Unimplemented and no RPCs are audited.
	AuditRepo auditrepo.Repository
	// HealthPinger is used by HealthService for readiness (e.g. *sql.DB). If nil, HealthCheck skips DB ping.
	HealthPinger healthhandler.Pinger
	// HealthPolicyChecker is used by HealthService for readiness (e.g. OPA evaluator). If nil, HealthCheck skips policy check.
	HealthPolicyChecker healthhandler.PolicyChecker
	// DevOTPHandler is the dev-only DevService (GetOTP). If nil, DevService is not registered. Set only when dev OTP is enabled and not production.
	DevOTPHandler devv1.DevServiceServer
	// TelemetryEmitter emits telemetry events (e.g. via OTel Logs). If nil, TelemetryService Emit RPCs no-op.
	TelemetryEmitter telemetry.EventEmitter
	// MembershipRepo is used by MembershipService. If nil, membership RPCs return Unimplemented.
	MembershipRepo membershiprepo.Repository
	// SessionRepo is used by SessionService. If nil, session RPCs return Unimplemented.
	SessionRepo sessionrepo.Repository
	// UserRepo is used by UserService (e.g. GetUserByEmail). If nil, user RPCs return Unimplemented.
	UserRepo userrepo.Repository
	// AuditLogger logs org-admin actions (membership/session). If nil, admin actions are not audited.
	AuditLogger audit.AuditLogger
	// OrgPolicyConfigRepo is used by OrgPolicyConfigService. If nil, org policy config RPCs return Unimplemented.
	OrgPolicyConfigRepo orgpolicyconfigrepo.Repository
	// OrgMFASettingsRepo is used by OrgPolicyConfigService to sync auth_mfa and device_trust on update. If nil, sync is skipped.
	OrgMFASettingsRepo orgmfasettingsrepo.Repository
	// OrgRepo is used by OrganizationService. If nil, organization RPCs return Unimplemented.
	OrgRepo organizationrepo.Repository
}

// RegisterServices registers all proto gRPC services with the given server.
//
// Proto → handler mapping:
//   - AdminService       → internal/admin/handler
//   - AuthService        → internal/identity/handler
//   - UserService        → internal/user/handler
//   - OrganizationService → internal/organization/handler
//   - DeviceService      → internal/device/handler
//   - MembershipService  → internal/membership/handler
//   - PolicyService      → internal/policy/handler
//   - SessionService     → internal/session/handler
//   - TelemetryService   → internal/telemetry/handler
//   - AuditService       → internal/audit/handler
//   - HealthService      → internal/health/handler
func RegisterServices(s grpc.ServiceRegistrar, deps Deps) {
	adminv1.RegisterAdminServiceServer(s, adminhandler.NewServer())
	var authSvc *identityservice.AuthService
	if deps.Auth != nil {
		authSvc = deps.Auth
	}
	authv1.RegisterAuthServiceServer(s, identityhandler.NewAuthServer(authSvc))
	userv1.RegisterUserServiceServer(s, userhandler.NewServer(deps.UserRepo))
	organizationv1.RegisterOrganizationServiceServer(s, organizationhandler.NewServer(deps.OrgRepo, deps.UserRepo, deps.MembershipRepo))
	devicev1.RegisterDeviceServiceServer(s, devicehandler.NewServer(deps.DeviceRepo))
	membershipv1.RegisterMembershipServiceServer(s, membershiphandler.NewServer(deps.MembershipRepo, deps.UserRepo, deps.AuditLogger))
	policyv1.RegisterPolicyServiceServer(s, policyhandler.NewServer(deps.PolicyRepo))
	orgpolicyconfigv1.RegisterOrgPolicyConfigServiceServer(s, orgpolicyconfighandler.NewServer(deps.OrgPolicyConfigRepo, deps.MembershipRepo, deps.OrgMFASettingsRepo))
	sessionv1.RegisterSessionServiceServer(s, sessionhandler.NewServer(deps.SessionRepo, deps.MembershipRepo, deps.AuditLogger))
	telemetryv1.RegisterTelemetryServiceServer(s, telemetryhandler.NewServer(deps.TelemetryEmitter))
	auditv1.RegisterAuditServiceServer(s, audithandler.NewServer(deps.AuditRepo, deps.MembershipRepo))
	healthv1.RegisterHealthServiceServer(s, healthhandler.NewServer(deps.HealthPinger, deps.HealthPolicyChecker))
	if deps.DevOTPHandler != nil {
		devv1.RegisterDevServiceServer(s, deps.DevOTPHandler)
	}
}
