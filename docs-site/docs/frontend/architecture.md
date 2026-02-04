---
title: Frontend Architecture
sidebar_label: Architecture
---

# Frontend Architecture

This document describes the **frontend architecture** of the zero-trust control plane: Next.js App Router structure, auth context, gRPC/HTTP bridge, and session handling. For dashboard pages and browser flows, see [Dashboard](./dashboard) and [User Browser](./user-browser).

**Audience**: Frontend developers onboarding or extending the web client.

## Overview

The frontend is a **Next.js** application (App Router) that talks to the backend only via **Next.js API routes**. The browser never calls gRPC directly; API routes run on the server and use gRPC clients to call the backend, then return JSON (or redirects). Auth state (tokens, user/org ids) is stored in **localStorage** and exposed through an **auth context**; a **401** from any API triggers clear auth and redirect to login.

## App structure

- **Routes**: [app/](../../../frontend/app/) — App Router. Public: `/`, `/login`, `/register`. Authenticated: `/dashboard` (and children), `/browser`. API routes under `app/api/`.
- **API routes**:
  - **auth**: `login`, `logout`, `refresh`, `register`, `mfa/request-with-phone`, `mfa/verify` — AuthService, token issuance, MFA flow.
  - **org-admin**: `members`, `audit`, `policy-config`, `sessions` — Org admin dashboard backend (Membership, Audit, OrgPolicyConfig, Session).
  - **browser**: `check-url`, `policy` — CheckUrlAccess, GetBrowserPolicy for the user browser flow.
  - **users**: `by-email` — UserService.GetUserByEmail (e.g. add member by email).
  - **dev**: `mfa/otp` — Dev OTP for development.

## Auth context

- **Provider**: [contexts/auth-context.tsx](../../../frontend/contexts/auth-context.tsx). Wraps the app and exposes `useAuth()`.
- **Exposed**: `user` (user_id, org_id), `accessToken`, `isAuthenticated`, `isLoading`, `login`, `verifyMFA`, `logout`, `refresh`, `setAuthFromResponse`, `clearAuth`, **handleSessionInvalid**.
- **Storage**: Tokens and user/org ids in localStorage; keys prefixed with `ztcp_`.
- **401 handling**: When any API returns 401, the caller should invoke **handleSessionInvalid()** — it clears auth state and redirects to `/login`. Used by dashboard and other protected pages.

## gRPC/HTTP bridge

The browser does **not** call gRPC. Flow:

1. Page or client code calls the API with the Bearer token in the `Authorization` header (e.g. `fetch('/api/...', { headers: { Authorization: 'Bearer ' + accessToken } })`).
2. The API route uses [getAccessToken(request)](../../../frontend/lib/api/get-access-token.ts) to read the Bearer token; if missing, returns 401.
3. The route instantiates gRPC clients from [lib/grpc/](../../../frontend/lib/grpc/): [auth-client.ts](../../../frontend/lib/grpc/auth-client.ts), [dev-client.ts](../../../frontend/lib/grpc/dev-client.ts), [org-admin-clients.ts](../../../frontend/lib/grpc/org-admin-clients.ts). These connect to the backend gRPC server (URL from env, e.g. `NEXT_PUBLIC_GRPC_URL` or similar).
4. Errors from gRPC are mapped to HTTP status and JSON with [grpc-to-http.ts](../../../frontend/lib/grpc/grpc-to-http.ts) (`grpcErrorToHttp`): e.g. UNAUTHENTICATED → 401, PERMISSION_DENIED → 403, NOT_FOUND → 404.

## Session handling

- **Refresh**: The auth context exposes `refresh()`; pages or middleware can call it to refresh the access token before expiry. API routes that need a valid token use `getAccessToken(request)` and return 401 if absent.
- **Proactive refresh**: Client code can refresh before long operations or on a timer; after 401, **handleSessionInvalid()** clears storage and redirects to login so the user re-authenticates.
