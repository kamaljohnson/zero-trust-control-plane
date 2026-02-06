---
title: Web Dashboard (Org Admin)
sidebar_label: Dashboard
---

# Web Dashboard (Org Admin)

This document describes the **authenticated org admin dashboard**: layout, Members, Audit log, and Policy pages, API route pattern, and 401 handling. The dashboard lives under [app/dashboard/](../../../frontend/app/dashboard/).

**Audience**: Frontend developers extending the dashboard or integrating with org-admin APIs.

## Overview

The dashboard is an authenticated area for **org admins** (role owner or admin). It provides a layout with navigation: **Members**, **Audit log**, **Policy**. Access requires a valid session; a **401** on any dashboard API call triggers clear auth and redirect to login via `handleSessionInvalid` from the auth context.

## Layout

- **Route**: `/dashboard` (and children). Layout in [app/dashboard/layout.tsx](../../../frontend/app/dashboard/layout.tsx): requires auth (redirects to sign-in when not authenticated), shows nav links to Members, Audit log, Policy.
- **Nav**: Members → `/dashboard`, Audit log → `/dashboard/audit`, Policy → `/dashboard/policy`.

## Members

- **Page**: [app/dashboard/page.tsx](../../../frontend/app/dashboard/page.tsx). List members (paginated), add member by email (UserService.GetUserByEmail + MembershipService.AddMember), remove member, update role (dropdown). Per-member: expand to show active sessions; revoke one session or revoke all sessions for that user.
- **API routes**: GET/POST under `/api/org-admin/members`, `/api/org-admin/sessions` (list, revoke, revoke-all), `/api/users/by-email`. RBAC enforced by backend (RequireOrgAdmin).

## Audit log

- **Page**: [app/dashboard/audit/page.tsx](../../../frontend/app/dashboard/audit/page.tsx). Paginated list of org audit events (time, user, action, resource, IP).
- **API**: GET `/api/org-admin/audit`. Backend: AuditService.ListAuditLogs.

## Policy

- **Page**: [app/dashboard/policy/page.tsx](../../../frontend/app/dashboard/policy/page.tsx). Five sections: Auth & MFA, Device Trust, Session Management, Access Control, Action Restrictions. Load on mount (GET `/api/org-admin/policy-config`), edit in form, "Save all" (PUT same route with full config).
- **API**: GET/PUT `/api/org-admin/policy-config`. Backend: OrgPolicyConfigService. Enum normalization for `mfa_requirement` and `default_action` when loading (proto may return string enums).

## API route pattern

All org-admin routes:

1. Use `getAccessToken(request)`; return **401** when the token is missing.
2. Call gRPC via [lib/grpc/org-admin-clients.ts](../../../frontend/lib/grpc/org-admin-clients.ts).
3. Map gRPC errors with `grpcErrorToHttp`.

Frontend pages pass the Bearer token in `fetch` headers; on `res.status === 401` they call **handleSessionInvalid()** from the auth context (clear storage, redirect to `/login`).

## Auth context

**handleSessionInvalid()** is exposed by the auth context for 401 handling. It clears token and user state and redirects to `/login`. Used by dashboard Members, Audit, and Policy pages so that revoked or expired sessions log the user out immediately.
