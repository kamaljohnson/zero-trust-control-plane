import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebars: SidebarsConfig = {
  docsSidebar: [
    "intro",
    {
      type: "category",
      label: "Backend",
      link: {
        type: "generated-index",
        title: "Backend",
        description: "Backend services, auth, database, and observability.",
      },
      items: [
        "backend/grpc-api-overview",
        "backend/auth",
        "backend/audit",
        "backend/database",
        "backend/device-trust",
        "backend/health",
        "backend/mfa",
        "backend/org-policy-config",
        "backend/organization-membership",
        "backend/policy-engine",
        "backend/sessions",
        "backend/session-lifecycle",
        "backend/telemetry",
        "backend/testing",
      ],
    },
    {
      type: "category",
      label: "Frontend",
      link: {
        type: "generated-index",
        title: "Frontend",
        description: "Web client and org admin dashboard.",
      },
      items: ["frontend/architecture", "frontend/dashboard", "frontend/user-browser"],
    },
    {
      type: "category",
      label: "Operations",
      link: {
        type: "generated-index",
        title: "Operations",
        description: "Deployment, environment, and production.",
      },
      items: ["operations/deployment"],
    },
  ],
};

export default sidebars;
