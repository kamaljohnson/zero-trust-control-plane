---
title: Introduction
sidebar_label: Introduction
---

# Introduction

Welcome to the **Zero Trust Control Plane** documentation. This project is a proof-of-concept zero-trust session and policy control plane: backend (Go gRPC), web client (Next.js), and CLI.

## What you'll find here

- **Backend** — Authentication, audit logging, database schema, device trust, health checks, MFA, policy engine (OPA/Rego for device-trust/MFA), session management (list/revoke sessions, token invalidation on revoke), session lifecycle (creation, heartbeats, revocation, client behavior), org policy config (five sections, sync to org MFA settings), and telemetry (OpenTelemetry → Collector → Loki / Prometheus / Tempo → Grafana).
- **Frontend** — Web dashboard for org admins: Members, Audit log, Policy, Telemetry.
- **Contributing** — Planned documentation and how to extend the docs.

## Quick links

- [Auth](/docs/backend/auth) — Register, login, refresh, logout, and JWT flows.
- [Sessions](/docs/backend/sessions) — Session management, revocation, and token invalidation.
- [Session lifecycle](/docs/backend/session-lifecycle) — Session creation, heartbeats, revocation, client behavior.
- [Org policy config](/docs/backend/org-policy-config) — Per-org policy (five sections) and sync to org_mfa_settings.
- [Policy engine](/docs/backend/policy-engine) — OPA/Rego integration, policy structure, evaluation flow.
- [Web dashboard](/docs/frontend/dashboard) — Org admin dashboard: Members, Audit, Policy, Telemetry.
- [Database](/docs/backend/database) — Schema, migrations, and codegen.
- [Telemetry](/docs/backend/telemetry) — OpenTelemetry SDK, Collector, Loki, Prometheus, Tempo, Grafana.

Run the backend from `backend/`, the frontend from `frontend/`, and this docs site from `docs-site/` (see [docs-site README](../../docs-site/README.md)).
