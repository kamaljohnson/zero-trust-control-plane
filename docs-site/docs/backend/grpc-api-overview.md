---
title: gRPC API Overview
sidebar_label: gRPC API overview
---

# gRPC API Overview

This document summarizes the **gRPC API** of the zero-trust control plane: a single gRPC server, all services and their main RPCs, and how callers use it (backend in-process, frontend via Next.js API routes). For auth, sessions, policy, and other topics, see the linked backend docs.

**Audience**: Developers integrating with the backend or extending the API.

## Overview

- **Server**: One gRPC server (default port **8080**). Wired in [internal/server/grpc.go](../../../backend/internal/server/grpc.go); entry point [cmd/server/main.go](../../../backend/cmd/server/main.go).
- **Protos**: [backend/proto/](../../../backend/proto/) — one directory per service (admin, auth, user, organization, membership, device, session, policy, telemetry, audit, health, orgpolicyconfig, dev, common). Generated Go stubs in [backend/api/generated/](../../../backend/api/generated/).

## Services and RPCs

| Service | Purpose | Main RPCs |
|--------|---------|------------|
| **AdminService** | System admin | GetSystemStats |
| **AuthService** | Auth, MFA, tokens | Register, Login, VerifyCredentials, VerifyMFA, SubmitPhoneAndRequestMFA, Refresh, Logout, LinkIdentity |
| **UserService** | User lookup and lifecycle | GetUser, GetUserByEmail, ListUsers, DisableUser, EnableUser |
| **OrganizationService** | Orgs (tenants) | CreateOrganization (public), GetOrganization, ListOrganizations, SuspendOrganization |
| **MembershipService** | Org membership and roles | AddMember, RemoveMember, UpdateRole, ListMembers |
| **DeviceService** | Device trust | RegisterDevice, GetDevice, ListDevices, RevokeDevice |
| **SessionService** | Sessions | RevokeSession, ListSessions, GetSession, RevokeAllSessionsForUser |
| **PolicyService** | Rego policies (device-trust/MFA) | CreatePolicy, UpdatePolicy, DeletePolicy, ListPolicies |
| **OrgPolicyConfigService** | Org policy config (MFA, device, session, access control) | GetOrgPolicyConfig, UpdateOrgPolicyConfig, GetBrowserPolicy, CheckUrlAccess |
| **TelemetryService** | Telemetry events | EmitTelemetryEvent, BatchEmitTelemetry |
| **AuditService** | Audit logs | ListAuditLogs |
| **HealthService** | Readiness/liveness | HealthCheck |
| **DevService** | Dev-only (e.g. OTP) | GetOTP |

Details: [auth](./auth), [sessions](./sessions), [session-lifecycle](./session-lifecycle), [mfa](./mfa), [device-trust](./device-trust), [policy-engine](./policy-engine), [org-policy-config](./org-policy-config), [audit](./audit), [organization-membership](./organization-membership), [health](./health), [telemetry](./telemetry).

**Public Endpoints**: Most RPCs require a Bearer access token (obtained via Login or Refresh). Public endpoints that do not require authentication include:
- `AuthService.Register`, `AuthService.Login`, `AuthService.VerifyCredentials`, `AuthService.VerifyMFA`, `AuthService.SubmitPhoneAndRequestMFA`, `AuthService.Refresh`
- `OrganizationService.CreateOrganization` (allows newly registered users to create organizations before login)
- `HealthService.HealthCheck`
- `DevService.GetOTP` (dev-only)

## Calling the API

- **From the backend**: Handlers and services use the same process; no network call. Dependencies are injected into [RegisterServices](../../../backend/internal/server/grpc.go); if a dep is nil, that service may return Unimplemented.
- **From the frontend**: The browser does **not** call gRPC. Next.js API routes (e.g. under `frontend/app/api/`) use gRPC clients ([frontend/lib/grpc/](../../../frontend/lib/grpc/)) to call the backend; they map gRPC errors to HTTP status and JSON via [grpc-to-http.ts](../../../frontend/lib/grpc/grpc-to-http.ts). See [Frontend Architecture](../frontend/architecture).
