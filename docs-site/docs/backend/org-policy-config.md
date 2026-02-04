---
title: Org Policy Config
sidebar_label: Org Policy Config
---

# Org Policy Config

This document describes the **OrgPolicyConfigService**: per-org structured policy configuration with five sections, Get/Update API, storage in `org_policy_config`, and sync of Auth & MFA and Device Trust sections to `org_mfa_settings`. The canonical proto is [orgpolicyconfig/orgpolicyconfig.proto](../../../backend/proto/orgpolicyconfig/orgpolicyconfig.proto); the handler is [internal/orgpolicyconfig/handler/grpc.go](../../../backend/internal/orgpolicyconfig/handler/grpc.go).

**Audience**: Developers building the policy UI, integrating with the policy API, or extending enforcement (session TTL, domain/action rules).

## Overview

**OrgPolicyConfigService** provides **GetOrgPolicyConfig** and **UpdateOrgPolicyConfig**. The config is stored as JSON in the **org_policy_config** table (one row per org). The **Auth & MFA** and **Device Trust** sections are synced to **org_mfa_settings** on update so existing [auth_service](../../../backend/internal/identity/service/auth_service.go) and [OPA evaluator](../../../backend/internal/policy/engine/opa_evaluator.go) behavior stay aligned without code changes. All RPCs require **org admin or owner** (RequireOrgAdmin); request `org_id` must match the caller's context org.

## Five sections

Defaults below are from [internal/orgpolicyconfig/domain/config.go](../../../backend/internal/orgpolicyconfig/domain/config.go) (DefaultAuthMfa, DefaultDeviceTrust, etc.).

### 1. Auth & MFA

Org-level when to require MFA and which methods are allowed.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| mfa_requirement | enum/string | new_device | When to require MFA: always, new_device, untrusted. Synced to org_mfa_settings. |
| allowed_mfa_methods | repeated string | ["sms_otp"] | Allowed methods (e.g. sms_otp). Stored; future use for step-up. |
| step_up_sensitive_actions | bool | false | Require step-up MFA for sensitive actions. Stored for future. |
| step_up_policy_violation | bool | false | Require step-up on policy violation. Stored for future. |

### 2. Device Trust

Device registration, auto-trust after MFA, and trust TTL.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| device_registration_allowed | bool | true | Whether devices can be registered. Stored for future. |
| auto_trust_after_mfa | bool | true | After MFA, mark device trusted. Synced to RegisterTrustAfterMFA. |
| max_trusted_devices_per_user | int32 | 0 | 0 = unlimited. Stored for future. |
| reverify_interval_days | int32 | 30 | Trust TTL in days. Synced to TrustTTLDays. |
| admin_revoke_allowed | bool | true | Admins may revoke devices. Stored for future. |

### 3. Session Management

Session lifetime, idle timeout, and concurrent-session limits. Stored for future enforcement.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| session_max_ttl | string | "24h" | Max session lifetime (duration). Stored for future. |
| idle_timeout | string | "30m" | Idle timeout (duration). Stored for future. |
| concurrent_session_limit | int32 | 0 | 0 = unlimited. Stored for future. |
| admin_forced_logout | bool | true | Admins may force logout. Stored for future. |
| reauth_on_policy_change | bool | false | Require reauth when policy changes. Stored for future. |

### 4. Access Control

Allowed/blocked domains and default action for browser. Stored for future enforcement.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| allowed_domains | repeated string | [] | Allowed domains (browser). Stored for future. |
| blocked_domains | repeated string | [] | Blocked domains. Stored for future. |
| wildcard_supported | bool | false | Whether wildcards are supported. Stored for future. |
| default_action | enum/string | allow | allow or deny when no rule matches. Stored for future. |

### 5. Action Restrictions

Allowed actions and read-only mode. Stored for future enforcement.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| allowed_actions | repeated string | navigate, download, upload, copy_paste | Allowed actions. Stored for future. |
| read_only_mode | bool | false | Restrict to read-only. Stored for future. |

## API

### Request and response

- **GetOrgPolicyConfigRequest**: `org_id` (optional; defaults to context org).
- **GetOrgPolicyConfigResponse**: `config` (OrgPolicyConfig with five sections; nil sections are merged with defaults when returned).
- **UpdateOrgPolicyConfigRequest**: `org_id`, `config` (full or partial; merged with defaults before save).
- **UpdateOrgPolicyConfigResponse**: `config` (merged result).
- **RBAC**: Caller must be org admin or owner (RequireOrgAdmin). If request `org_id` is empty, context org is used; if non-empty, it must equal context org.

### GetOrgPolicyConfig behavior

If no row exists for the org, `GetByOrgID` returns nil. The handler then calls `MergeWithDefaults(nil)` and returns that merged config (all five sections filled with defaults). So the client always receives a full config.

### UpdateOrgPolicyConfig behavior

The request may contain a full or partial config (any section may be omitted). The handler uses `protoToDomain` (partial OK), then `Upsert` with that domain config. Before returning and before sync, the handler uses `MergeWithDefaults(config)` so stored JSON and response are consistent with defaults for missing sections. Sync to org_mfa_settings runs only when `config.AuthMfa != nil` or `config.DeviceTrust != nil`, using the merged config.

## Storage

- **Table**: `org_policy_config` — `org_id` (VARCHAR PK, REFERENCES organizations), `config_json` (TEXT NOT NULL, default `'{}'`), `updated_at` (TIMESTAMPTZ NOT NULL). One row per org.  
- **Domain**: Structs and defaults in [internal/orgpolicyconfig/domain/config.go](../../../backend/internal/orgpolicyconfig/domain/config.go); `MergeWithDefaults` fills nil sections.  
- **Repository**: GetByOrgID (returns nil when no row), Upsert (JSON marshal); see [internal/orgpolicyconfig/repository](../../../backend/internal/orgpolicyconfig/repository).

## Sync to org_mfa_settings

On **UpdateOrgPolicyConfig**, if `auth_mfa` or `device_trust` is present, the handler maps to **OrgMFASettings** and calls the org_mfa_settings repository **Upsert** via [domainToOrgMFASettings](../../../backend/internal/orgpolicyconfig/handler/grpc.go). Only **Auth & MFA** and **Device Trust** are synced; Session Management, Access Control, and Action Restrictions are **not** written to org_mfa_settings.

| Policy config field | org_mfa_settings column | Notes |
|---------------------|-------------------------|-------|
| auth_mfa.mfa_requirement = always | MFARequiredAlways = true, MFARequiredForNewDevice = false, MFARequiredForUntrusted = false | |
| auth_mfa.mfa_requirement = new_device | MFARequiredForNewDevice = true, MFARequiredForUntrusted = true, MFARequiredAlways = false | |
| auth_mfa.mfa_requirement = untrusted | MFARequiredForUntrusted = true, MFARequiredForNewDevice = false, MFARequiredAlways = false | |
| device_trust.auto_trust_after_mfa | RegisterTrustAfterMFA | |
| device_trust.reverify_interval_days | TrustTTLDays | Only applied when > 0 |

Other OrgMFASettings columns (e.g. when a section is absent) get default values in the handler (MFARequiredForNewDevice true, MFARequiredForUntrusted true, RegisterTrustAfterMFA true, TrustTTLDays 30). Consumers ([auth_service](../../../backend/internal/identity/service/auth_service.go), [OPA evaluator](../../../backend/internal/policy/engine/opa_evaluator.go)) read from org_mfa_settings only; updating policy config keeps that table in sync so no code changes are needed there.

## Default values reference

When a section is absent (nil), the handler and domain use these defaults (from [domain/config.go](../../../backend/internal/orgpolicyconfig/domain/config.go)):

| Section | Defaults |
|---------|----------|
| Auth & MFA | mfa_requirement = new_device, allowed_mfa_methods = ["sms_otp"], step_up_sensitive_actions = false, step_up_policy_violation = false |
| Device Trust | device_registration_allowed = true, auto_trust_after_mfa = true, max_trusted_devices_per_user = 0, reverify_interval_days = 30, admin_revoke_allowed = true |
| Session Management | session_max_ttl = "24h", idle_timeout = "30m", concurrent_session_limit = 0, admin_forced_logout = true, reauth_on_policy_change = false |
| Access Control | allowed_domains = [], blocked_domains = [], wildcard_supported = false, default_action = allow |
| Action Restrictions | allowed_actions = ["navigate", "download", "upload", "copy_paste"], read_only_mode = false |

## Dashboard and enforcement

**Dashboard**: The org admin dashboard [Policy page](/docs/frontend/dashboard) loads config via GET and saves via PUT. Clients may need to normalize proto enums (e.g. mfa_requirement, default_action) when they are returned as strings.

**Enforcement**: Auth & MFA and Device Trust are effectively enforced today because they are synced to org_mfa_settings and used by auth_service and the policy engine. Session Management, Access Control, and Action Restrictions are stored for future enforcement (e.g. session service, browser/agent); the current implementation does not enforce them.

## Wiring

OrgPolicyConfigService is registered in [internal/server/grpc.go](../../../backend/internal/server/grpc.go). The handler is constructed in [cmd/server/main.go](../../../backend/cmd/server/main.go) with the org policy config repo, membershipRepo (for RequireOrgAdmin), and orgMfaSettingsRepo (for sync).
