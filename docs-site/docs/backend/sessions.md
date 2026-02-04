---
title: Session Management
sidebar_label: Sessions
---

# Session Management

This document describes the **SessionService**: listing and revoking sessions for an organization, and how revocation invalidates both refresh and access tokens. The canonical proto is [session/session.proto](../../../backend/proto/session/session.proto); the handler is [internal/session/handler/grpc.go](../../../backend/internal/session/handler/grpc.go).

**Audience**: Developers integrating with session APIs or implementing clients that react to session revocation (e.g. org admin dashboard, logout on 401).

## Overview

**SessionService** provides RPCs to list sessions for an org (with optional user filter), revoke a single session, and revoke all sessions for a user. All RPCs require the caller to be an **org admin or owner** (RBAC via [RequireOrgAdmin](../../../backend/internal/platform/rbac/require_org_admin.go)). Session data is read from the **sessions** table; revocation sets `sessions.revoked_at` and is enforced immediately for both refresh and access tokens (see [Token invalidation](#token-invalidation)).

## RPCs

| RPC | Request | Response | Notes |
|-----|---------|----------|-------|
| **ListSessions** | `org_id`, optional `user_id`, `pagination` (page_size, page_token) | `sessions[]`, `pagination` (next_page_token, total_count when supported) | Returns only **non-revoked** sessions for the org; optional filter by user. |
| **RevokeSession** | `session_id` | empty | Session must belong to caller's org. Sets `sessions.revoked_at`. |
| **RevokeAllSessionsForUser** | `org_id`, `user_id` | empty | Revokes all sessions for that user in the org. |
| **GetSession** | `session_id` | `session` | Returns the session (including `revoked_at` when set). Used by SessionValidator; callers can use it to check session state. |

**Request/response shapes**: See [session.proto](../../../backend/proto/session/session.proto). `ListSessionsRequest` uses `ztcp.common.v1.Pagination` (e.g. page_size, page_token); `ListSessionsResponse` includes `sessions` and `pagination` (PaginationResult). Session message includes `id`, `user_id`, `org_id`, `device_id`, `expires_at`, `revoked_at`, `last_seen_at`, `ip_address`, `created_at`.

## Session revocation semantics

- **Revoke** (single or all for user) sets `sessions.revoked_at` to the current time. The row remains; only the timestamp is updated.
- **ListSessions** returns only sessions where `revoked_at IS NULL` (active sessions).
- **GetSession** returns the session by ID; the response includes `revoked_at`, so callers can distinguish active vs revoked.

## Token invalidation

Revoking a session invalidates both refresh and access tokens for that session so the user is effectively logged out.

1. **Refresh token**  
   [AuthService.Refresh](../../../backend/internal/identity/service/auth_service.go) loads the session by `session_id` from the refresh JWT. If `sess.RevokedAt != nil`, it returns **ErrInvalidRefreshToken** (gRPC Unauthenticated). So any attempt to refresh using that session’s refresh token fails immediately after revocation.

2. **Access token**  
   The auth interceptor supports an optional **SessionValidator**. When auth is enabled, the server wires a validator that looks up the session by `session_id` (from the access token) and checks that the session exists and is not revoked (`sess.RevokedAt == nil`). If the session is missing or revoked, the interceptor returns **Unauthenticated** before the RPC handler runs. So the next API call that sends that access token receives 401; clients (e.g. web dashboard) should treat 401 as session invalid and clear auth state and redirect to login.

**Summary**: Revoking a session invalidates both refresh and access immediately. The frontend receives 401 on the next authenticated request and can clear storage and redirect to login.

## Wiring

- **SessionValidator** is built in [cmd/server/main.go](../../../backend/cmd/server/main.go): when `deps.SessionRepo != nil`, a closure is created that calls `SessionRepo.GetByID(ctx, sessionID)` and returns `active = (sess != nil && sess.RevokedAt == nil)`. This validator is passed into `interceptors.AuthUnary(tokens, publicMethods, sessionValidator)`.
- **Audit**: Revoke actions (RevokeSession, RevokeAllSessionsForUser) are audited via the handler’s audit logger (e.g. action `revoke`, resource `session`).

## Database

The **sessions** table is described in [database.md](./database). Revocation only updates `revoked_at`; no other columns are changed. For schema and migrations, see [database.md](./database).

## See also

- For session creation, heartbeats (last_seen_at), and client behavior, see [Session lifecycle](./session-lifecycle).
- [database.md](./database) — Schema and migrations for the sessions table.
