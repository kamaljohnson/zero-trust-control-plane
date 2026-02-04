---
title: User Browser
sidebar_label: User Browser
---

# User Browser

This document describes the **policy-enforced user browser**: layout, URL flow, action buttons, BFF API routes, backend RPCs, and how Access Control and Action Restrictions are applied. The browser lives under [app/browser/](../../../frontend/app/browser/).

**Audience**: Frontend developers extending the browser or integrating with browser APIs; org admins configuring Access Control and Action Restrictions.

## Overview

The user browser is a policy-enforced UI for **any authenticated org member** (not just org admins). The home page (`/`) redirects logged-in users to `/browser`. Users enter a URL; the app checks it against the org's **Access Control** policy; allowed URLs are opened in a new tab. **Action Restrictions** (allowed_actions, read_only_mode) control which action buttons are enabled. The app does not embed third-party sites in an iframe because most sites send X-Frame-Options or Content-Security-Policy headers that cause the browser to refuse to connect; opening in a new tab is the supported flow.

## Layout and routes

- **Route**: `/browser`. Layout in [app/browser/layout.tsx](../../../frontend/app/browser/layout.tsx): requires auth (redirects to sign-in when not authenticated); nav shows **Home**, **Org admin**, user/org info, and **Sign out**.
- **Home**: [app/page.tsx](../../../frontend/app/page.tsx). Unauthenticated: login/register card. Authenticated: redirect to `/browser` so the browser is the effective home for logged-in users.

## Page behavior

- **Page**: [app/browser/page.tsx](../../../frontend/app/browser/page.tsx). On mount, fetches browser policy (GET `/api/browser/policy`) to get `access_control` and `action_restrictions`. URL input and **Go** (or Enter): input is normalized (adds `https://` if no scheme), then POST `/api/browser/check-url` is called. If the response is **denied**, an error card is shown (API `reason` or default message); the URL is not opened. If **allowed**, the page shows "Access allowed" and an **Open in new tab** link; no iframe is used, since most sites block embedding (X-Frame-Options / CSP).
- **Action buttons**: Download, Upload, Copy/paste. Each is enabled only when the corresponding action is in `allowed_actions` and not overridden by `read_only_mode` (upload and copy_paste are disabled when `read_only_mode` is true). Buttons currently show placeholder messages; full download/upload/copy-paste flows are for future implementation.
- **Read-only mode**: When `read_only_mode` is true, a banner is shown and Upload and Copy/paste are disabled.
- **Error messaging**: Policy load failure, check-url failure, and access denied (with optional `reason` from the API) are shown in cards. On **401** from any browser API, the page calls `handleSessionInvalid()` from the auth context (clear storage, redirect to `/login`).

## BFF API routes

### GET /api/browser/policy

- **Query**: `org_id` (required).
- **Headers**: `Authorization: Bearer <access_token>`.
- **Response**: `{ access_control?, action_restrictions? }`. Used to drive action button state and the read-only indicator.
- **401**: Client should call `handleSessionInvalid()`.
- **Implementation**: [app/api/browser/policy/route.ts](../../../frontend/app/api/browser/policy/route.ts). Uses [lib/grpc/org-admin-clients.ts](../../../frontend/lib/grpc/org-admin-clients.ts) `getBrowserPolicy` and `grpcErrorToHttp`.

### POST /api/browser/check-url

- **Body**: `{ url: string, org_id: string }`.
- **Headers**: `Authorization: Bearer <access_token>`.
- **Response**: `{ allowed: boolean, reason?: string }`. Called before allowing the user to open a URL.
- **401**: Same handling as above.
- **Implementation**: [app/api/browser/check-url/route.ts](../../../frontend/app/api/browser/check-url/route.ts). Uses `checkUrlAccess` and `grpcErrorToHttp`.

## Backend RPCs (reference)

- **GetBrowserPolicy(org_id)**: Returns only `access_control` and `action_restrictions` (no auth_mfa, device_trust, session_mgmt). Callable by **any org member** (RequireOrgMember). Implemented in [backend/internal/orgpolicyconfig/handler/grpc.go](../../../backend/internal/orgpolicyconfig/handler/grpc.go).
- **CheckUrlAccess(org_id, url)**: Evaluates the URL host against the org's Access Control policy: blocked list first, then allowed list and default_action; optional wildcard matching when `wildcard_supported` is true. Returns `allowed` and optional user-facing `reason` when denied. Callable by any org member.

## Policy

- **Access Control**: `allowed_domains`, `blocked_domains`, `default_action`, `wildcard_supported`. Used by CheckUrlAccess to allow or deny a URL.
- **Action Restrictions**: `allowed_actions` (e.g. navigate, download, upload, copy_paste), `read_only_mode`. Used to enable/disable action buttons and show the read-only banner.

Org admins configure these on the [Policy](/docs/frontend/dashboard) page (Access Control and Action Restrictions sections). Full field reference: [Org policy config](/docs/backend/org-policy-config).

## Auth and 401

Browser pages pass the Bearer token in `fetch` headers. On `res.status === 401` they call **handleSessionInvalid()** from the auth context (clear storage, redirect to `/login`), consistent with the dashboard and other authenticated areas.
