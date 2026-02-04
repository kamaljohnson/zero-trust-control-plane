---
title: Backend Testing
sidebar_label: Testing
---

# Backend Testing

This document describes the comprehensive test suite for the zero-trust control plane backend. All tests use Go's standard `testing` package and follow consistent patterns for mocking dependencies and testing both success and error scenarios.

**Audience**: Developers writing or maintaining backend tests, and contributors understanding test coverage.

## Overview

The backend test suite consists of **32 test files** covering:

- **gRPC handlers** (12 handlers) - API layer tests
- **Services** - Business logic tests (AuthService)
- **Interceptors** - Middleware tests (auth, audit, context)
- **RBAC utilities** - Authorization helper tests
- **Security utilities** - Token, hashing, keys, and refresh token hash tests
- **Policy engine** - OPA/Rego evaluation tests
- **Telemetry** - OTel adapter and async emission tests
- **MFA utilities** - OTP generation, SMS client, and dev OTP store tests
- **Audit utilities** - Mapping and logger tests
- **Configuration** - Config loading and validation tests

All tests use **in-memory mocks** for repositories and dependencies, ensuring fast execution and isolation. Tests follow a consistent pattern: setup mocks, execute handler/service method, assert results and error codes.

## Test Organization

Tests are co-located with the code they test, following Go conventions:

```
backend/
├── internal/
│   ├── admin/handler/grpc_test.go
│   ├── user/handler/grpc_test.go
│   ├── organization/handler/grpc_test.go
│   ├── membership/handler/grpc_test.go
│   ├── device/handler/grpc_test.go
│   ├── session/handler/grpc_test.go
│   ├── policy/handler/grpc_test.go
│   ├── audit/
│   │   ├── handler/grpc_test.go
│   │   ├── mapping_test.go
│   │   └── logger_test.go
│   ├── orgpolicyconfig/handler/grpc_test.go
│   ├── telemetry/
│   │   ├── handler/grpc_test.go
│   │   ├── async_test.go
│   │   └── otel/adapter_test.go
│   ├── health/handler/grpc_test.go
│   ├── devotp/
│   │   ├── handler/grpc_test.go
│   │   └── store_test.go
│   ├── identity/
│   │   ├── handler/grpc_test.go
│   │   └── service/auth_service_test.go
│   ├── mfa/
│   │   ├── otp_test.go
│   │   └── sms/smslocal_test.go
│   ├── server/interceptors/
│   │   ├── auth_test.go
│   │   ├── audit_test.go
│   │   └── context_test.go
│   ├── platform/rbac/
│   │   ├── require_org_admin_test.go
│   │   └── require_org_member_test.go
│   ├── security/
│   │   ├── tokens_test.go
│   │   ├── hashing_test.go
│   │   ├── keys_test.go
│   │   └── refresh_hash_test.go
│   ├── config/config_test.go
│   └── policy/engine/opa_evaluator_test.go
```

## Test Categories

### Handler Tests (gRPC Layer)

Handler tests verify the gRPC API layer, including request validation, error mapping, authorization checks, and proto conversion.

#### Admin Handler Tests
**File**: [`backend/internal/admin/handler/grpc_test.go`](../../../backend/internal/admin/handler/grpc_test.go)

**Purpose**: Tests the AdminService gRPC handler for system-level admin operations.

**Test Scenarios**:
- `GetSystemStats`: Returns Unimplemented status
- `NewServer`: Creates server instance

**Key Test Cases**:
- Unimplemented method returns correct gRPC status code
- Server initialization

**Dependencies**: None (simple stub handler)

#### User Handler Tests
**File**: [`backend/internal/user/handler/grpc_test.go`](../../../backend/internal/user/handler/grpc_test.go)

**Purpose**: Tests the UserService gRPC handler for user lookup operations.

**Test Scenarios**:
- `GetUser`: Success, not found, invalid user_id, repository errors, nil repo (Unimplemented)
- `GetUserByEmail`: Success, not found, invalid email, repository errors, nil repo
- `GetUser`: Disabled user status conversion
- `ListUsers`: Unimplemented stub
- `DisableUser`: Unimplemented stub
- `EnableUser`: Unimplemented stub

**Key Test Cases**:
- Validates user_id and email trimming/validation
- Tests gRPC status code mapping (NotFound, InvalidArgument, Internal, Unimplemented)
- Verifies proto conversion (domain.User → userv1.User)

**Dependencies**: `mockUserRepo` implementing `userrepo.Repository`

#### Organization Handler Tests
**File**: [`backend/internal/organization/handler/grpc_test.go`](../../../backend/internal/organization/handler/grpc_test.go)

**Purpose**: Tests the OrganizationService gRPC handler for organization lookup.

**Test Scenarios**:
- `GetOrganization`: Success, not found, invalid org_id, repository errors, nil repo
- `GetOrganization`: Suspended status conversion
- `CreateOrganization`: Unimplemented stub
- `ListOrganizations`: Unimplemented stub
- `SuspendOrganization`: Unimplemented stub

**Key Test Cases**:
- Validates org_id trimming and validation
- Tests status enum conversion (Active, Suspended)

**Dependencies**: `mockOrgRepo` implementing `organizationrepo.Repository`

#### Membership Handler Tests
**File**: [`backend/internal/membership/handler/grpc_test.go`](../../../backend/internal/membership/handler/grpc_test.go)

**Purpose**: Tests the MembershipService gRPC handler for org membership and role management.

**Test Scenarios**:
- `AddMember`: Success, duplicate member, invalid user_id, user not found, non-admin caller, org_id mismatch, default role assignment, nil repo
- `RemoveMember`: Success, membership not found, last owner protection, non-admin caller, org_id mismatch, nil repo
- `UpdateRole`: Success, membership not found, last owner demotion protection, invalid role, non-admin caller, org_id mismatch, nil repo
- `ListMembers`: Success, pagination (page size, offset, next token), max page size enforcement, non-admin caller, org_id mismatch, nil repo

**Key Test Cases**:
- RBAC enforcement (RequireOrgAdmin)
- Last owner protection (cannot remove/demote last owner)
- Pagination with page tokens and size limits
- Audit logging verification

**Dependencies**: `mockMembershipRepo`, `mockUserRepo`, `mockAuditLogger`, RBAC context helpers

#### Device Handler Tests
**File**: [`backend/internal/device/handler/grpc_test.go`](../../../backend/internal/device/handler/grpc_test.go)

**Purpose**: Tests the DeviceService gRPC handler for device trust operations.

**Test Scenarios**:
- `GetDevice`: Success, not found, repository errors, nil repo, timestamp handling (LastSeenAt, TrustedUntil, RevokedAt)
- `ListDevices`: Success, filtered by user_id, empty list, repository errors, nil repo
- `RevokeDevice`: Success, repository errors, nil repo
- `RegisterDevice`: Unimplemented stub

**Key Test Cases**:
- User filtering in ListDevices
- Optional timestamp field handling
- Device trust status verification

**Dependencies**: `mockDeviceRepo` implementing `repository.Repository`

#### Session Handler Tests
**File**: [`backend/internal/session/handler/grpc_test.go`](../../../backend/internal/session/handler/grpc_test.go)

**Purpose**: Tests the SessionService gRPC handler for session management.

**Test Scenarios**:
- `RevokeSession`: Success, session not found, wrong org, non-admin caller, invalid session_id, nil repo
- `ListSessions`: Success, pagination, filtered by user_id, non-admin caller, org_id mismatch, nil repo
- `GetSession`: Success, session not found, wrong org, non-admin caller, nil repo
- `RevokeAllSessionsForUser`: Success, invalid user_id, non-admin caller, org_id mismatch, nil repo

**Key Test Cases**:
- Multi-tenant isolation (org_id validation)
- RBAC enforcement (RequireOrgAdmin)
- Pagination with next page tokens
- Audit logging for revocation events

**Dependencies**: `mockSessionRepo`, `mockMembershipRepoForSession`, `mockAuditLoggerForSession`

#### Policy Handler Tests
**File**: [`backend/internal/policy/handler/grpc_test.go`](../../../backend/internal/policy/handler/grpc_test.go)

**Purpose**: Tests the PolicyService gRPC handler for Rego policy CRUD operations.

**Test Scenarios**:
- `CreatePolicy`: Success, invalid org_id, empty rules, invalid Rego syntax, repository errors, nil repo
- `UpdatePolicy`: Success, invalid policy_id, policy not found, invalid Rego syntax, empty rules allowed, nil repo
- `DeletePolicy`: Success, invalid policy_id, repository errors, nil repo
- `ListPolicies`: Success, empty list, invalid org_id, repository errors, nil repo

**Key Test Cases**:
- Rego syntax validation using OPA parser
- Policy enabled/disabled state
- Empty rules handling (allowed in UpdatePolicy)

**Dependencies**: `mockPolicyRepo` implementing `repository.Repository`

#### Audit Handler Tests
**File**: [`backend/internal/audit/handler/grpc_test.go`](../../../backend/internal/audit/handler/grpc_test.go)

**Purpose**: Tests the AuditService gRPC handler for audit log retrieval.

**Test Scenarios**:
- `ListAuditLogs`: Success, pagination, filters (user_id, action, resource), max page size, non-admin caller, org_id mismatch, repository errors, nil repo, no org admin checker, missing org context

**Key Test Cases**:
- Multi-filter support (user_id, action, resource)
- Pagination with next page tokens
- RBAC enforcement (optional org admin checker)
- Fallback to context org_id when no checker

**Dependencies**: `mockAuditRepo`, `mockMembershipRepoForAudit`

#### OrgPolicyConfig Handler Tests
**File**: [`backend/internal/orgpolicyconfig/handler/grpc_test.go`](../../../backend/internal/orgpolicyconfig/handler/grpc_test.go)

**Purpose**: Tests the OrgPolicyConfigService gRPC handler for org policy configuration and URL access evaluation.

**Test Scenarios**:
- `GetOrgPolicyConfig`: Success, defaults merging, non-admin caller, org_id mismatch, nil repo
- `UpdateOrgPolicyConfig`: Success, sync to org_mfa_settings, non-admin caller, org_id mismatch, nil repo
- `GetBrowserPolicy`: Success, non-member caller, org_id mismatch, nil repo
- `CheckUrlAccess`: Success, blocked domain, allowed domain, wildcard matching, invalid URL, default deny/allow, URL without protocol, case insensitive matching, non-member caller, org_id mismatch, nil repo

**Key Test Cases**:
- URL access evaluation logic (domain matching, wildcards, defaults)
- Default policy merging (MergeWithDefaults)
- MFA settings synchronization
- RBAC enforcement (RequireOrgAdmin vs RequireOrgMember)

**Dependencies**: `mockOrgPolicyConfigRepo`, `mockMembershipRepoForOrgPolicyConfig`, `mockOrgMFASettingsRepo`

#### Telemetry Handler Tests
**File**: [`backend/internal/telemetry/handler/grpc_test.go`](../../../backend/internal/telemetry/handler/grpc_test.go)

**Purpose**: Tests the TelemetryService gRPC handler for telemetry event emission.

**Test Scenarios**:
- `EmitTelemetryEvent`: Nil request handling, nil emitter (no-op), valid request with all fields
- `BatchEmitTelemetry`: Nil request, nil emitter, nil events in batch, batch truncation (maxBatchSize=500)

**Key Test Cases**:
- Graceful nil handling (no-op behavior)
- Batch size enforcement
- Event field validation

**Dependencies**: `mockEmitter` with channel-based event capture

#### Health Handler Tests
**File**: [`backend/internal/health/handler/grpc_test.go`](../../../backend/internal/health/handler/grpc_test.go)

**Purpose**: Tests the HealthService gRPC handler for health checks.

**Test Scenarios**:
- `HealthCheck`: Nil pinger, pinger success, pinger failure, policy checker success, policy checker failure, both checks with policy failure

**Key Test Cases**:
- Health check aggregation (database + policy engine)
- Failure handling (does not fail RPC, returns NOT_SERVING status)
- Optional dependencies (nil pinger/checker)

**Dependencies**: `mockPinger`, `mockPolicyChecker`

#### Identity/Auth Handler Tests
**File**: [`backend/internal/identity/handler/grpc_test.go`](../../../backend/internal/identity/handler/grpc_test.go)

**Purpose**: Tests the AuthService gRPC handler for error mapping and proto conversion.

**Test Scenarios**:
- `Register`: Nil auth service (Unimplemented)
- `Login`: Nil auth service (Unimplemented)
- `VerifyMFA`: Nil auth service (Unimplemented)
- `SubmitPhoneAndRequestMFA`: Nil auth service (Unimplemented)
- `Refresh`: Nil auth service (Unimplemented)
- `Logout`: Nil auth service (no-op, returns success)
- `LinkIdentity`: Unimplemented
- Error mapping tests: EmailAlreadyRegistered, InvalidCredentials, InvalidRefreshToken, RefreshTokenReuse, NotOrgMember, PhoneRequiredForMFA, InvalidMFAChallenge, InvalidOTP, InvalidMFAIntent, ChallengeExpired
- Proto conversion tests: LoginResultToProto (tokens, MFARequired, PhoneRequired), RefreshResultToProto, AuthResultToProto

**Key Test Cases**:
- gRPC status code mapping (AlreadyExists, Unauthenticated, PermissionDenied, FailedPrecondition)
- Proto conversion for all result types
- Nil service handling (Unimplemented vs no-op)

**Dependencies**: Tests error mapping functions, not actual AuthService (service tests are separate)

#### DevOTP Handler Tests
**File**: [`backend/internal/devotp/handler/grpc_test.go`](../../../backend/internal/devotp/handler/grpc_test.go)

**Purpose**: Tests the DevService gRPC handler for dev-only OTP retrieval.

**Test Scenarios**:
- `GetOTP`: Success (returns OTP and note), not found (expired or missing challenge_id), invalid challenge_id (empty string), nil store handling

**Key Test Cases**:
- OTP retrieval from dev store
- Challenge ID validation
- Error handling (NotFound, InvalidArgument)
- Dev mode note in response

**Dependencies**: Mock `devotp.Store` implementation

### Service Tests (Business Logic)

#### AuthService Tests
**File**: [`backend/internal/identity/service/auth_service_test.go`](../../../backend/internal/identity/service/auth_service_test.go)

**Purpose**: Tests the core authentication business logic including registration, login, MFA flows, token refresh, and logout.

**Test Scenarios**:
- `Register`: Success, email already registered, validation errors (email format, password strength)
- `Login`: Success, wrong password, requires membership, MFA required (new device), phone required, OTP return to client
- `LoginAndRefreshAndLogout`: Full flow with trusted device
- `Refresh`: Success, token reuse detection, revoked session, empty token, untrusted device, new device
- `VerifyMFA`: Device trust registration, expired challenge
- `SubmitPhoneAndRequestMFA`: Expired intent
- `LogoutFromContext`: Context-based logout

**Key Test Cases**:
- Password validation (length, uppercase, lowercase, number, symbol)
- Refresh token reuse detection (revokes all sessions)
- Device trust policy evaluation
- MFA challenge/OTP flow
- Session lifecycle (creation, refresh, revocation)
- Phone verification workflow

**Dependencies**: In-memory mock repositories (`memUserRepo`, `memIdentityRepo`, `memSessionRepo`, `memDeviceRepo`, `memMembershipRepo`, `memMFAChallengeRepo`, `memMFAIntentRepo`), `memPolicyEvaluator`, `recordingOTPSender`, test token provider

**Test Helpers**: `newTestAuthService`, `newTestAuthServiceOpt` (for OTP return to client testing)

### Audit Utility Tests

#### Audit Mapping Tests
**File**: [`backend/internal/audit/mapping_test.go`](../../../backend/internal/audit/mapping_test.go)

**Purpose**: Tests gRPC full method name parsing for audit logging.

**Test Scenarios**:
- `ParseFullMethod`: Standard methods (Get, List, Create, Update, Delete), membership overrides (AddMember → user_added, RemoveMember → user_removed, UpdateRole → role_changed), unknown format, edge cases
- `serviceToResource`: Service name conversion (UserService → user)
- `methodToAction`: Method name to action verb mapping

**Key Test Cases**:
- Standard CRUD method mapping
- Membership service overrides
- Service name to resource conversion
- Method name to action verb conversion
- Unknown format handling
- Edge cases (no slash, no dot, etc.)

**Dependencies**: None (pure functions)

#### Audit Logger Tests
**File**: [`backend/internal/audit/logger_test.go`](../../../backend/internal/audit/logger_test.go)

**Purpose**: Tests audit logger for event persistence and telemetry emission.

**Test Scenarios**:
- `Logger.LogEvent`: Success, nil repo (no-op), IP extraction, sentinel org_id, telemetry emission, repository error handling
- `auditActionToEventType`: Mapping for login_success, login_failure, logout, session_created, unknown action

**Key Test Cases**:
- Audit log entry creation with all fields
- IP extractor integration
- Sentinel org_id for events without org
- Telemetry event emission (async)
- Error resilience (repository errors don't fail caller)
- Action to event type mapping
- Non-mapped actions (no telemetry emission)

**Dependencies**: Mock `auditrepo.Repository`, mock `IPExtractor`, mock `telemetry.EventEmitter`

### Interceptor Tests (Middleware)

#### Auth Interceptor Tests
**File**: [`backend/internal/server/interceptors/auth_test.go`](../../../backend/internal/server/interceptors/auth_test.go)

**Purpose**: Tests the authentication interceptor that validates Bearer tokens and sets identity context.

**Test Scenarios**:
- `AuthUnary`: Public methods (allow without token), protected methods (require valid token, reject invalid/missing token), session validation (accept valid session, reject revoked session, handle validator error), context setting (user_id, org_id, session_id)
- `ExtractBearer`: Valid token, case insensitive prefix, missing metadata, invalid prefix, whitespace handling

**Key Test Cases**:
- Public method bypass (no token required)
- Token validation and error handling
- Session validator integration
- Context identity injection
- Bearer token extraction from metadata

**Dependencies**: `security.NewTestTokenProvider`, mock session validator

#### Audit Interceptor Tests
**File**: [`backend/internal/server/interceptors/audit_test.go`](../../../backend/internal/server/interceptors/audit_test.go)

**Purpose**: Tests the audit logging interceptor that records RPC events.

**Test Scenarios**:
- `AuditUnary`: Skip methods (no audit log), authenticated requests (audit log created), unauthenticated requests (no audit log), repository errors (does not fail RPC), handler errors (still logs), full method parsing
- `ClientIP`: X-Forwarded-For header, X-Real-IP header, precedence (X-Forwarded-For first), peer address fallback, unknown fallback, whitespace handling, comma-separated IPs

**Key Test Cases**:
- Skip list functionality
- Audit log creation with correct fields (org_id, user_id, action, resource, IP)
- IP extraction priority (X-Forwarded-For > X-Real-IP > peer > unknown)
- Error resilience (audit failures don't fail RPCs)

**Dependencies**: `mockAuditRepoForInterceptor` implementing `auditrepo.Repository`

#### Context Helper Tests
**File**: [`backend/internal/server/interceptors/context_test.go`](../../../backend/internal/server/interceptors/context_test.go)

**Purpose**: Tests context helpers for identity information (user_id, org_id, session_id).

**Test Scenarios**:
- `WithIdentity`: Sets user_id, org_id, session_id
- `GetUserID`: Returns value when set, returns false when not set
- `GetOrgID`: Returns value when set, returns false when not set
- `GetSessionID`: Returns value when set, returns false when not set

**Key Test Cases**:
- Context value setting and retrieval
- Missing value handling (returns false)
- Context isolation (different contexts don't interfere)
- Context chaining (overriding values)
- Empty value handling

**Dependencies**: None (context operations)

### Configuration Tests

#### Config Loading Tests
**File**: [`backend/internal/config/config_test.go`](../../../backend/internal/config/config_test.go)

**Purpose**: Tests configuration loading from environment variables and .env files.

**Test Scenarios**:
- `Load`: Default values, env var override, validation (GRPC_ADDR required, BCRYPT_COST range, OTP_RETURN_TO_CLIENT + production validation)
- `AccessTTL`: Valid duration, invalid duration (defaults to 15m), zero/negative (defaults to 15m)
- `RefreshTTL`: Valid duration, invalid duration (defaults to 168h), zero/negative (defaults to 168h)

**Key Test Cases**:
- Default value loading
- Environment variable override
- Validation (required fields, ranges)
- Production safety (OTP_RETURN_TO_CLIENT validation)
- Duration parsing and defaults
- Error handling

**Dependencies**: Environment variable manipulation, temporary .env files

### RBAC Utility Tests

#### RequireOrgAdmin Tests
**File**: [`backend/internal/platform/rbac/require_org_admin_test.go`](../../../backend/internal/platform/rbac/require_org_admin_test.go)

**Purpose**: Tests the RBAC utility that enforces org admin or owner role requirement.

**Test Scenarios**:
- Success: Owner role, Admin role
- Failure: Member role, not a member, no context, repository error, empty org_id, empty user_id

**Key Test Cases**:
- Role hierarchy (Owner and Admin allowed, Member denied)
- Context validation
- Error code mapping (Unauthenticated, PermissionDenied, Internal)

**Dependencies**: `mockMembershipGetter` implementing `OrgMembershipGetter`

#### RequireOrgMember Tests
**File**: [`backend/internal/platform/rbac/require_org_member_test.go`](../../../backend/internal/platform/rbac/require_org_member_test.go)

**Purpose**: Tests the RBAC utility that enforces org membership (any role).

**Test Scenarios**:
- Success: Any role (owner, admin, member)
- Failure: Not a member, no context, repository error, empty org_id, empty user_id

**Key Test Cases**:
- Any role acceptance (less restrictive than RequireOrgAdmin)
- Context validation
- Error code mapping

**Dependencies**: `mockMembershipGetterForMember` implementing `OrgMembershipGetter`

### MFA Utility Tests

#### MFA OTP Function Tests
**File**: [`backend/internal/mfa/otp_test.go`](../../../backend/internal/mfa/otp_test.go)

**Purpose**: Tests OTP generation, hashing, and constant-time comparison functions.

**Test Scenarios**:
- `GenerateOTP`: Returns 6-digit string, uses crypto/rand for randomness
- `HashOTP`: SHA-256 hash, hex encoding, deterministic output
- `OTPEqual`: Constant-time comparison, correct matches, rejects wrong OTPs

**Key Test Cases**:
- OTP format validation (6 digits, numeric only)
- Randomness verification (no duplicates in multiple generations)
- Hash consistency (same input produces same hash)
- Hash uniqueness (different inputs produce different hashes)
- Constant-time comparison (timing attack resistance)
- Empty input handling

**Dependencies**: None (pure functions)

#### MFA SMS Client Tests
**File**: [`backend/internal/mfa/sms/smslocal_test.go`](../../../backend/internal/mfa/sms/smslocal_test.go)

**Purpose**: Tests SMS Local API client for OTP delivery.

**Test Scenarios**:
- `NewSMSLocalClient`: Default base URL, custom base URL, HTTP client timeout
- `SendOTP`: Success (mocked HTTP), missing API key, HTTP errors, non-200 status codes, request format validation

**Key Test Cases**:
- Default configuration
- Custom base URL and sender
- HTTP request format (route, numbers, variables, headers)
- Error handling (missing API key, network errors, API errors)
- Response parsing

**Dependencies**: Mock HTTP server (`httptest`)

#### DevOTP Store Tests
**File**: [`backend/internal/devotp/store_test.go`](../../../backend/internal/devotp/store_test.go)

**Purpose**: Tests in-memory OTP store for dev mode.

**Test Scenarios**:
- `MemoryStore.Put`: Stores OTP with expiration
- `MemoryStore.Get`: Returns OTP when valid, returns false when expired/missing, cleans up expired entries

**Key Test Cases**:
- OTP storage and retrieval
- Expiration handling (boundary conditions)
- Expired entry cleanup
- Concurrent access safety (race detection)
- Multiple OTPs management

**Dependencies**: None (in-memory implementation)

### Security Utility Tests

#### Token Provider Tests
**File**: [`backend/internal/security/tokens_test.go`](../../../backend/internal/security/tokens_test.go)

**Purpose**: Tests JWT token issuance and validation for access and refresh tokens.

**Test Scenarios**:
- `IssueAccessAndRefresh`: Token issuance, jti generation, expiration times, validation
- `ValidateRefresh`: Valid token, invalid token
- `ValidateAccess`: Valid token, invalid token

**Key Test Cases**:
- Token structure (claims, expiration, jti)
- Token validation (signature, expiration, claims)
- Error handling (ErrInvalidToken)

**Dependencies**: `security.NewTestTokenProvider` (uses embedded test RSA keys)

#### Password Hashing Tests
**File**: [`backend/internal/security/hashing_test.go`](../../../backend/internal/security/hashing_test.go)

**Purpose**: Tests bcrypt password hashing and comparison.

**Test Scenarios**:
- `HashAndCompare`: Successful hash and verify
- `CompareWrongPassword`: Rejection of incorrect passwords
- `Cost`: Cost parameter validation, zero cost clamping

**Key Test Cases**:
- Bcrypt hashing and verification
- Cost parameter handling
- Security (wrong passwords rejected)

**Dependencies**: `security.NewHasher`

#### Security Keys Tests
**File**: [`backend/internal/security/keys_test.go`](../../../backend/internal/security/keys_test.go)

**Purpose**: Tests PEM key loading and parsing utilities.

**Test Scenarios**:
- `LoadPEM`: Inline PEM, file path, literal `\n` conversion, empty string, invalid file
- `ParsePrivateKey`: RSA PKCS1, RSA PKCS8, ECDSA, invalid PEM, invalid key type
- `ParsePublicKey`: RSA PKCS1, RSA PKIX, ECDSA, invalid PEM
- `KeyAlg`: RS256 for RSA, ES256 for ECDSA P-256, empty for unsupported

**Key Test Cases**:
- Inline PEM handling (with `\n` conversion)
- File path reading
- Multiple key formats (PKCS1, PKCS8, PKIX)
- Error handling (invalid PEM, invalid key type, missing files)
- Algorithm detection

**Dependencies**: Test key files or embedded test keys (`testPrivateKeyPEM`, `testPublicKeyPEM`)

#### Refresh Token Hash Tests
**File**: [`backend/internal/security/refresh_hash_test.go`](../../../backend/internal/security/refresh_hash_test.go)

**Purpose**: Tests refresh token hashing and constant-time comparison.

**Test Scenarios**:
- `HashRefreshToken`: SHA-256 hash, hex encoding, deterministic output
- `RefreshTokenHashEqual`: Constant-time comparison, correct matches, rejects wrong tokens

**Key Test Cases**:
- Hash consistency (same token produces same hash)
- Hash uniqueness (different tokens produce different hashes)
- Constant-time comparison (timing attack resistance)
- Empty input handling
- Different hash format handling

**Dependencies**: None (pure functions)

### Policy Engine Tests

#### OPA Evaluator Tests
**File**: [`backend/internal/policy/engine/opa_evaluator_test.go`](../../../backend/internal/policy/engine/opa_evaluator_test.go)

**Purpose**: Tests the OPA/Rego policy evaluator health check.

**Test Scenarios**:
- `HealthCheck`: Evaluator initialization and health check

**Key Test Cases**:
- OPA evaluator initialization
- Health check functionality

**Dependencies**: OPA evaluator (can be nil for health check)

### Telemetry Tests

#### OTel Adapter Tests
**File**: [`backend/internal/telemetry/otel/adapter_test.go`](../../../backend/internal/telemetry/otel/adapter_test.go)

**Purpose**: Tests the OpenTelemetry adapter for telemetry event emission.

**Test Scenarios**:
- `NewEventEmitter`: Nil provider returns noop emitter
- `Emit`: Nil event handling, attribute mapping, body mapping (metadata), empty metadata handling

**Key Test Cases**:
- Noop emitter for nil provider
- Event attribute mapping (org_id, user_id, device_id, session_id, event_type, source)
- Body mapping from metadata bytes
- Empty metadata handling

**Dependencies**: OpenTelemetry SDK (`sdklog.NewLoggerProvider`), `recordCapture` helper

#### Telemetry Async Tests
**File**: [`backend/internal/telemetry/async_test.go`](../../../backend/internal/telemetry/async_test.go)

**Purpose**: Tests asynchronous telemetry event emission.

**Test Scenarios**:
- `EmitAsync`: Nil emitter (no-op), nil event (no-op), successful emit, timeout handling, error handling (logged, doesn't panic)

**Key Test Cases**:
- No-op behavior for nil inputs
- Goroutine-based async execution
- Context.Background() usage (not request context)
- Timeout handling
- Error logging without panicking
- Concurrent access safety

**Dependencies**: Mock `EventEmitter` with mutex-protected event capture

## Testing Patterns

### Mock Repositories

All handler tests use in-memory mock repositories that implement the repository interfaces. Mocks store data in maps keyed by ID or composite keys (e.g., `userID:orgID`).

**Example Pattern**:
```go
type mockUserRepo struct {
    usersByID    map[string]*domain.User
    usersByEmail map[string]*domain.User
    getByIDErr   error
}
```

### Test Helpers

Common test helpers include:
- `ctxWithAdmin(orgID, userID)` - Creates context with admin identity
- `ctxWithMember(orgID, userID)` - Creates context with member identity
- `security.NewTestTokenProvider()` - Creates token provider with test keys
- `newTestAuthService(t)` - Creates AuthService with all mock dependencies

### Table-Driven Tests

Some tests use table-driven patterns for multiple similar scenarios:

```go
testCases := []struct {
    name  string
    input string
    want  error
}{
    {"empty", "", ErrInvalid},
    {"whitespace", "   ", ErrInvalid},
}
```

### Error Assertions

Tests verify gRPC status codes:

```go
st, ok := status.FromError(err)
if !ok {
    t.Fatalf("error is not a gRPC status: %v", err)
}
if st.Code() != codes.NotFound {
    t.Errorf("status code = %v, want %v", st.Code(), codes.NotFound)
}
```

## Running Tests

### Run All Tests

From the `backend/` directory:

```bash
go test ./...
```

### Run Tests for Specific Package

```bash
go test ./internal/user/handler/...
go test ./internal/identity/service/...
```

### Run Tests with Verbose Output

```bash
go test -v ./...
```

### Run Tests with Race Detection

```bash
go test -race ./...
```

### Run Specific Test

```bash
go test -v -run TestGetUser_Success ./internal/user/handler/...
```

### Generate Coverage Report

See [Coverage](#coverage) section below.

## Coverage

### Generate Coverage Profile

From the `backend/` directory:

```bash
go test -race -covermode=atomic -coverprofile=coverage.out ./...
```

This generates `coverage.out` containing coverage data for all packages.

### View Coverage Summary

```bash
go tool cover -func=coverage.out
```

Shows coverage percentage per function.

### Generate HTML Coverage Report

```bash
go tool cover -html=coverage.out -o coverage.html
```

Opens `coverage.html` in your browser showing line-by-line coverage.

### Coverage Script

Use the provided script for convenience:

```bash
./scripts/test-coverage.sh
```

This script:
1. Runs tests with coverage
2. Generates HTML report
3. Outputs coverage summary

### CI Coverage

Coverage is automatically generated and uploaded to Codecov on every push and pull request. See the [GitHub Actions workflow](../../../.github/workflows/backend-tests.yml) for details.

View coverage reports at: https://codecov.io/gh/OWNER/REPO (replace OWNER/REPO with your repository)

## Test Coverage Summary

The test suite covers:

- **gRPC Handlers**: All 12 handlers with success, error, and edge cases
- **Services**: AuthService with full authentication flows
- **Interceptors**: Auth, audit, and context middleware
- **RBAC**: Authorization utilities
- **Security**: Token provider, password hashing, key parsing, and refresh token hashing
- **Policy Engine**: OPA evaluator health check
- **Telemetry**: OTel adapter and async emission
- **MFA**: OTP generation, SMS client, and dev OTP store
- **Audit**: Mapping functions and logger
- **Configuration**: Config loading and validation

**Total Test Files**: 32  
**Test Patterns**: Unit tests with in-memory mocks, table-driven tests, error code verification, HTTP mocking, concurrent access testing

## Best Practices

1. **Use mocks**: All tests use in-memory mocks for repositories and dependencies
2. **Test error codes**: Verify gRPC status codes match expected values
3. **Test edge cases**: Empty inputs, nil values, boundary conditions
4. **Test authorization**: Verify RBAC enforcement (admin vs member vs non-member)
5. **Test pagination**: Verify page size limits, offsets, next tokens
6. **Test validation**: Input validation and error messages
7. **Use helpers**: Leverage test helpers for common setup (contexts, mocks)
8. **Isolate tests**: Each test should be independent and not rely on other tests

## CI Integration

Tests run automatically on:

- **Push to main**: Full test suite with coverage
- **Pull requests**: Full test suite with coverage and PR comments
- **Manual trigger**: Can be triggered manually from GitHub Actions

See `.github/workflows/backend-tests.yml` for the complete workflow configuration.
