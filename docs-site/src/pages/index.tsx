import React from "react";
import Link from "@docusaurus/Link";
import Layout from "@theme/Layout";

const quickLinks = [
  { to: "/backend/auth", label: "Auth", description: "Register, login, refresh, JWT flows" },
  { to: "/backend/sessions", label: "Sessions", description: "Session management, revocation, token invalidation" },
  { to: "/backend/session-lifecycle", label: "Session lifecycle", description: "Creation, heartbeats, revocation, client behavior" },
  { to: "/backend/policy-engine", label: "Policy engine", description: "OPA/Rego, policy structure, evaluation flow" },
  { to: "/backend/org-policy-config", label: "Org policy config", description: "Five sections, sync to org_mfa_settings" },
  { to: "/frontend/dashboard", label: "Web dashboard", description: "Org admin: Members, Audit, Policy" },
  { to: "/backend/database", label: "Database", description: "Schema, migrations, codegen" },
];

export default function Home(): React.ReactElement {
  return (
    <Layout
      title="Zero Trust Control Plane — Documentation"
      description="Documentation for the zero-trust session and policy control plane: backend (Go gRPC) and web client (Next.js). Auth, sessions, policy engine, org admin dashboard, database."
    >
      <main style={{ padding: "2rem 1rem", maxWidth: 900, margin: "0 auto" }}>
        <section style={{ textAlign: "center", marginBottom: "3rem" }}>
          <h1>Zero Trust Control Plane — Documentation</h1>
          <p style={{ fontSize: "1.15rem", marginBottom: "1.5rem" }}>
            This site documents a proof-of-concept <strong>zero-trust session and policy control plane</strong>: a backend (Go gRPC), web client (Next.js). Here you’ll find backend services (auth, sessions, policy engine, org policy config, database) and the org admin dashboard.
          </p>
          <Link to="/intro" className="button button--primary button--lg">
            Get started
          </Link>
        </section>

        <section style={{ marginBottom: "2.5rem" }}>
          <h2>What's in the docs</h2>
          <ul style={{ paddingLeft: "1.25rem" }}>
            <li>
              <strong>Backend</strong> — Authentication (register, login, refresh, MFA), session management and lifecycle, policy engine (OPA/Rego for device-trust/MFA), org policy config (five sections), database schema, audit, health.
            </li>
            <li>
              <strong>Frontend</strong> — Org admin dashboard (Members, Audit log, Policy); how it uses the backend and handles 401 / session invalidation.
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
