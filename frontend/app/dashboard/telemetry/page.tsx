"use client";

import { useAuth } from "@/contexts/auth-context";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

/**
 * Scoped telemetry: link to Grafana dashboard with org_id variable.
 * If NEXT_PUBLIC_GRAFANA_URL and org filter are configured, show link; else show placeholder.
 */
export default function DashboardTelemetryPage() {
  const { user } = useAuth();
  const orgId = user?.org_id ?? "";
  const baseUrl = process.env.NEXT_PUBLIC_GRAFANA_URL ?? "";

  const grafanaUrl = baseUrl
    ? `${baseUrl.replace(/\/$/, "")}/d/ztcp-telemetry?var-org_id=${encodeURIComponent(orgId)}`
    : null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Scoped telemetry</CardTitle>
        <CardDescription>
          View organization-scoped metrics and logs in Grafana.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {grafanaUrl ? (
          <p className="mb-4">
            <a
              href={grafanaUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary underline"
            >
              Open Grafana dashboard (org: {orgId || "—"})
            </a>
          </p>
        ) : (
          <p className="text-muted-foreground">
            Set <code className="rounded bg-muted px-1">NEXT_PUBLIC_GRAFANA_URL</code> to your
            Grafana base URL to link to the org-scoped telemetry dashboard. The dashboard
            should define a variable <code className="rounded bg-muted px-1">org_id</code> to
            filter by organization.
          </p>
        )}
      </CardContent>
    </Card>
  );
}
