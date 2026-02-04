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
        "backend/auth",
        "backend/audit",
        "backend/database",
        "backend/device-trust",
        "backend/health",
        "backend/mfa",
        "backend/org-policy-config",
        "backend/policy-engine",
        "backend/sessions",
        "backend/session-lifecycle",
        "backend/telemetry",
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
      items: ["frontend/dashboard", "frontend/user-browser"],
    },
    {
      type: "category",
      label: "Contributing",
      items: ["contributing/pending"],
    },
  ],
};

export default sidebars;
