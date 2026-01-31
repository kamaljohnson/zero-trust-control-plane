package server

import (
	"google.golang.org/grpc"

	adminv1 "zero-trust-control-plane/backend/api/generated/admin/v1"
	auditv1 "zero-trust-control-plane/backend/api/generated/audit/v1"
	authv1 "zero-trust-control-plane/backend/api/generated/auth/v1"
	devicev1 "zero-trust-control-plane/backend/api/generated/device/v1"
	healthv1 "zero-trust-control-plane/backend/api/generated/health/v1"
	membershipv1 "zero-trust-control-plane/backend/api/generated/membership/v1"
	organizationv1 "zero-trust-control-plane/backend/api/generated/organization/v1"
	policyv1 "zero-trust-control-plane/backend/api/generated/policy/v1"
	sessionv1 "zero-trust-control-plane/backend/api/generated/session/v1"
	telemetryv1 "zero-trust-control-plane/backend/api/generated/telemetry/v1"
	userv1 "zero-trust-control-plane/backend/api/generated/user/v1"

	adminhandler "zero-trust-control-plane/backend/internal/admin/handler"
	audithandler "zero-trust-control-plane/backend/internal/audit/handler"
	devicehandler "zero-trust-control-plane/backend/internal/device/handler"
	healthhandler "zero-trust-control-plane/backend/internal/health/handler"
	identityhandler "zero-trust-control-plane/backend/internal/identity/handler"
	membershiphandler "zero-trust-control-plane/backend/internal/membership/handler"
	organizationhandler "zero-trust-control-plane/backend/internal/organization/handler"
	policyhandler "zero-trust-control-plane/backend/internal/policy/handler"
	sessionhandler "zero-trust-control-plane/backend/internal/session/handler"
	telemetryhandler "zero-trust-control-plane/backend/internal/telemetry/handler"
	userhandler "zero-trust-control-plane/backend/internal/user/handler"
)

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
func RegisterServices(s grpc.ServiceRegistrar) {
	adminv1.RegisterAdminServiceServer(s, adminhandler.NewServer())
	authv1.RegisterAuthServiceServer(s, identityhandler.NewAuthServer())
	userv1.RegisterUserServiceServer(s, userhandler.NewServer())
	organizationv1.RegisterOrganizationServiceServer(s, organizationhandler.NewServer())
	devicev1.RegisterDeviceServiceServer(s, devicehandler.NewServer())
	membershipv1.RegisterMembershipServiceServer(s, membershiphandler.NewServer())
	policyv1.RegisterPolicyServiceServer(s, policyhandler.NewServer())
	sessionv1.RegisterSessionServiceServer(s, sessionhandler.NewServer())
	telemetryv1.RegisterTelemetryServiceServer(s, telemetryhandler.NewServer())
	auditv1.RegisterAuditServiceServer(s, audithandler.NewServer())
	healthv1.RegisterHealthServiceServer(s, healthhandler.NewServer())
}
