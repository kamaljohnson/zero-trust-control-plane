# Zero Trust Control Plane — Frontend

Web UI for the Zero Trust Control Plane. Built with Next.js 16 (App Router), React 19, Tailwind CSS 4, and [shadcn/ui](https://ui.shadcn.com) components. Includes a full auth flow (register, login, refresh, logout) that talks to the backend via a Next.js BFF (API routes → gRPC).

## Prerequisites

- **Node.js** `>= 20.9.0` (see [engines](package.json)). Use [nvm](https://github.com/nvm-sh/nvm) or [fnm](https://github.com/Schniz/fnm) if needed: `nvm use 20`.
- A running **backend** gRPC server (default `localhost:8080`) with auth enabled for login/register to work.

## Getting Started

1. **Install dependencies**

   ```bash
   npm install
   ```

2. **Configure environment** (optional)

   Copy [.env.example](.env.example) to `.env.local` and set:

   - `BACKEND_GRPC_URL` — backend gRPC address (default: `localhost:8080`).
   - `NEXT_PUBLIC_DEFAULT_ORG_ID` — optional; prefill org ID on the login form.

3. **Run the development server**

   ```bash
   npm run dev
   ```

   Open [http://localhost:3000](http://localhost:3000). The app shows the home page; sign in and register links when unauthenticated, or a simple “logged in” view and sign out when authenticated.

## Scripts

| Script   | Description                    |
|----------|--------------------------------|
| `npm run dev`   | Start Next.js dev server (Turbopack). |
| `npm run build` | Production build.              |
| `npm run start` | Start production server.        |
| `npm run lint`  | Run Next.js lint (ESLint).     |

## Project Structure

```
frontend/
├── app/
│   ├── api/auth/          # BFF: register, login, refresh, logout (JSON → gRPC, Zod validation)
│   ├── error.tsx          # Error boundary for app content (Try again / Go home)
│   ├── global-error.tsx   # Error boundary for root layout
│   ├── login/page.tsx     # Sign-in form (email, password, org ID)
│   ├── register/page.tsx  # Registration form (email, password, name)
│   ├── page.tsx           # Home: sign in/register or logged-in + sign out
│   ├── layout.tsx         # Root layout + AuthProvider
│   └── globals.css        # Tailwind + shadcn theme variables
├── components/ui/         # shadcn: Button, Input, Label, Card
├── contexts/
│   └── auth-context.tsx   # Auth state, login/logout/refresh, token storage, proactive refresh
├── hooks/                 # Custom hooks (shadcn alias @/hooks)
├── lib/
│   ├── api/auth-schemas.ts # Zod schemas for auth API request bodies
│   ├── auth-client.ts     # Browser: fetch to /api/auth/* (login, register, etc.)
│   ├── utils.ts           # cn() for Tailwind class merging
│   └── grpc/              # Server-only: gRPC client for backend AuthService
│       ├── auth-client.ts # register, login, refresh, logout → backend
│       ├── grpc-to-http.ts # Map gRPC status codes to HTTP + messages
│       └── proto/         # auth.proto + google protobuf stubs
├── components.json        # shadcn CLI config
├── .env.example           # Env template
└── README.md             # This file
```

## Authentication

- **Backend**: Auth is gRPC-only (Register, Login, Refresh, Logout). The browser cannot call gRPC directly.
- **BFF**: Next.js API routes under `app/api/auth/` accept JSON, validate request bodies with [Zod](https://zod.dev) ([lib/api/auth-schemas.ts](lib/api/auth-schemas.ts)), and forward to the backend via a Node gRPC client (`lib/grpc/auth-client.ts`). They return JSON and map gRPC errors to HTTP status and user-facing messages.
- **Browser**: The app uses `lib/auth-client.ts` to call `/api/auth/register`, `/api/auth/login`, etc. Tokens and user/org are stored in `localStorage` and exposed via `contexts/auth-context.tsx` (`useAuth()`). For production, consider httpOnly cookies set by the BFF.
- **Flows**:
  - **Register**: email, password, optional name → backend creates user; no tokens until the user is in an org and logs in.
  - **Login**: email (trimmed and lowercased, format validated), password, org ID → returns access + refresh tokens, user_id, org_id.
  - **Refresh**: The auth context refreshes the access token before it expires (5‑minute margin). The client sends **device_fingerprint** (from [lib/fingerprint.ts](lib/fingerprint.ts)) with `POST /api/auth/refresh` (body: `refresh_token`, optional `device_fingerprint`). The backend may return new tokens or, when device-trust policy requires MFA, **mfa_required** or **phone_required**. In the MFA case, the app clears auth state, stores the challenge/intent in sessionStorage, and redirects to `/login`, where the user completes the same MFA flow (OTP or phone then OTP); after VerifyMFA, the new tokens are stored and the user is redirected home.
  - **Logout**: `POST /api/auth/logout` with **access_token** and optional **refresh_token** in the body. The BFF forwards the access token as `Authorization: Bearer` to the backend so the call is authorized and logout is audited; the backend revokes the session, then the client clears storage.

Password policy (backend): 12+ characters, at least one uppercase, one lowercase, one number, one symbol. The register form validates this on the client.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `BACKEND_GRPC_URL` | No (default: `localhost:8080`) | Backend gRPC server address (host:port). Use TLS (`https://...`) in production. |
| `NEXT_PUBLIC_DEFAULT_ORG_ID` | No | Default organization ID for the login form (e.g. single-tenant). |

See [.env.example](.env.example).

## UI and Styling

- **Tailwind CSS 4** with PostCSS ([postcss.config.mjs](postcss.config.mjs)).
- **shadcn/ui**-style components in `components/ui/` (Button, Input, Label, Card). Theme variables and light/dark support are in [app/globals.css](app/globals.css). Config: [components.json](components.json).
- **Fonts**: [Geist](https://vercel.com/font) (sans and mono) via `next/font` in the root layout.

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [shadcn/ui](https://ui.shadcn.com)
- Backend auth: [backend/docs/auth.md](../backend/docs/auth.md)
