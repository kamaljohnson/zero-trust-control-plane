---
title: Introduction
sidebar_label: Introduction
---

# Introduction

Welcome to the **Zero Trust Control Plane** documentation. This project is a proof-of-concept zero-trust session and policy control plane: backend (Go gRPC), web client (Next.js), and CLI.

## What you'll find here

- **Backend** — Authentication, audit logging, database schema, device trust, health checks, MFA, and telemetry (Kafka → Loki → Grafana).
- **Contributing** — Planned documentation and how to extend the docs.

## Quick links

- [Auth](/docs/backend/auth) — Register, login, refresh, logout, and JWT flows.
- [Database](/docs/backend/database) — Schema, migrations, and codegen.
- [Telemetry](/docs/backend/telemetry) — gRPC interceptor, Kafka, worker, Loki, Grafana.

Run the backend from `backend/`, the frontend from `frontend/`, and this docs site from `docs-site/` (see [docs-site README](../../docs-site/README.md)).
