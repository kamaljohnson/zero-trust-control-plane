import React from "react";
import Link from "@docusaurus/Link";
import Layout from "@theme/Layout";

const quickLinks = [
  { to: "/docs/backend/auth", label: "Auth", description: "Register, login, refresh, JWT flows" },
  { to: "/docs/backend/sessions", label: "Sessions", description: "Session management, revocation, token invalidation" },
  { to: "/docs/backend/session-lifecycle", label: "Session lifecycle", description: "Creation, heartbeats, revocation, client behavior" },
  { to: "/docs/backend/policy-engine", label: "Policy engine", description: "OPA/Rego, policy structure, evaluation flow" },
  { to: "/docs/backend/org-policy-config", label: "Org policy config", description: "Five sections, sync to org_mfa_settings" },
  { to: "/docs/frontend/dashboard", label: "Web dashboard", description: "Org admin: Members, Audit, Policy, Telemetry" },
  { to: "/docs/backend/database", label: "Database", description: "Schema, migrations, codegen" },
  { to: "/docs/backend/telemetry", label: "Telemetry", description: "OpenTelemetry, Collector, Grafana" },
];

export default function Home(): React.ReactElement {
  return (
    <Layout
      title="Zero Trust Control Plane — Documentation"
      description="Documentation for the zero-trust session and policy control plane: backend (Go gRPC), web client (Next.js), and CLI. Auth, sessions, policy engine, org admin dashboard, database, and telemetry."
    >
      <main style={{ padding: "2rem 1rem", maxWidth: 900, margin: "0 auto" }}>
        <section style={{ textAlign: "center", marginBottom: "3rem" }}>
          <h1>Zero Trust Control Plane — Documentation</h1>
          <p style={{ fontSize: "1.15rem", marginBottom: "1.5rem" }}>
            This site documents a proof-of-concept <strong>zero-trust session and policy control plane</strong>: a backend (Go gRPC), web client (Next.js), and CLI. Here you’ll find backend services (auth, sessions, policy engine, org policy config, database, telemetry) and the org admin dashboard.
          </p>
          <Link to="/docs/intro" className="button button--primary button--lg">
            Get started
          </Link>
        </section>

        <section style={{ marginBottom: "2.5rem" }}>
          <h2>What's in the docs</h2>
          <ul style={{ paddingLeft: "1.25rem" }}>
            <li>
              <strong>Backend</strong> — Authentication (register, login, refresh, MFA), session management and lifecycle, policy engine (OPA/Rego for device-trust/MFA), org policy config (five sections), database schema, audit, health, and telemetry (OpenTelemetry → Collector → Loki / Prometheus / Tempo → Grafana).
            </li>
            <li>
              <strong>Frontend</strong> — Org admin dashboard (Members, Audit log, Policy, Telemetry); how it uses the backend and handles 401 / session invalidation.
            </li>
            <li>
              <strong>Contributing</strong> — Planned documentation and how to extend the docs (see the sidebar).
            </li>
          </ul>
        </section>

        <section style={{ marginBottom: "2.5rem" }}>
          <h2>Quick links</h2>
          <ul style={{ listStyle: "none", paddingLeft: 0 }}>
            {quickLinks.map(({ to, label, description }) => (
              <li key={to} style={{ marginBottom: "0.75rem" }}>
                <Link to={to} className="button button--secondary button--sm" style={{ marginRight: "0.5rem" }}>
                  {label}
                </Link>
                <span style={{ color: "var(--ifm-color-content-secondary)" }}> — {description}</span>
              </li>
            ))}
          </ul>
        </section>

        <section style={{ fontSize: "0.9rem", color: "var(--ifm-color-content-secondary)" }}>
          <h2>How to run</h2>
          <p>
            Run the backend from <code>backend/</code>, the frontend from <code>frontend/</code>, and this docs site from <code>docs-site/</code> (see the docs-site README).
          </p>
        </section>
      </main>
    </Layout>
  );
}
