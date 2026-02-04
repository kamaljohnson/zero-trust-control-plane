"use client";

import { useCallback, useEffect, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface AuditEvent {
  id: string;
  org_id: string;
  user_id: string;
  action: string;
  resource: string;
  ip: string;
  metadata?: string;
  created_at?: { seconds?: string; nanos?: number };
}

function authHeaders(accessToken: string): HeadersInit {
  return { Authorization: `Bearer ${accessToken}` };
}

function formatTime(ev: AuditEvent): string {
  const c = ev.created_at;
  if (!c?.seconds) return "—";
  const ms = Number(c.seconds) * 1000 + Number(c.nanos ?? 0) / 1e6;
  return new Date(ms).toLocaleString();
}

const PAGE_SIZE = 50;

export default function DashboardAuditPage() {
  const { user, accessToken, handleSessionInvalid } = useAuth();
  const [logs, setLogs] = useState<AuditEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pageToken, setPageToken] = useState<string>("");
  const [tokenHistory, setTokenHistory] = useState<string[]>([""]);
  const [currentPageIndex, setCurrentPageIndex] = useState(0);

  const orgId = user?.org_id ?? "";

  const loadLogs = useCallback(
    async (token?: string): Promise<boolean> => {
      if (!accessToken || !orgId) return false;
      setLoading(true);
      setError(null);
      try {
        const params = new URLSearchParams({ org_id: orgId, page_size: String(PAGE_SIZE) });
        if (token) params.set("page_token", token);
        const res = await fetch(`/api/org-admin/audit?${params}`, {
          headers: authHeaders(accessToken),
        });
        if (res.status === 401) {
          handleSessionInvalid();
          return false;
        }
        const data = await res.json();
        if (!res.ok) throw new Error(data.error ?? "Failed to load audit logs");
        setLogs(data.logs ?? []);
        setPageToken(data.pagination?.next_page_token ?? "");
        return true;
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load audit logs");
        return false;
      } finally {
        setLoading(false);
      }
    },
    [accessToken, orgId, handleSessionInvalid]
  );

  useEffect(() => {
    loadLogs().then((ok) => {
      if (ok) {
        setTokenHistory([""]);
        setCurrentPageIndex(0);
      }
    });
  }, [loadLogs]);

  const handleNext = useCallback(async () => {
    const tokenToUse = pageToken;
    const ok = await loadLogs(tokenToUse);
    if (ok) {
      setTokenHistory((prev) => [...prev, tokenToUse]);
      setCurrentPageIndex((prev) => prev + 1);
    }
  }, [loadLogs, pageToken]);

  const handlePrevious = useCallback(async () => {
    if (currentPageIndex <= 0) return;
    const tokenToUse = tokenHistory[currentPageIndex - 1];
    const ok = await loadLogs(tokenToUse);
    if (ok) {
      setCurrentPageIndex((prev) => prev - 1);
    }
  }, [loadLogs, tokenHistory, currentPageIndex]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit log</CardTitle>
        <CardDescription>Organization activity. Admin or owner only.</CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <p className="mb-4 text-sm text-destructive">{error}</p>
        )}
        {loading ? (
          <p className="text-muted-foreground">Loading…</p>
        ) : (
          <div className="space-y-2">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b">
                  <th className="pb-2 pr-4 font-medium">Time</th>
                  <th className="pb-2 pr-4 font-medium">User</th>
                  <th className="pb-2 pr-4 font-medium">Action</th>
                  <th className="pb-2 pr-4 font-medium">Resource</th>
                  <th className="pb-2 font-medium">IP</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((ev) => (
                  <tr key={ev.id} className="border-b">
                    <td className="py-2 pr-4 text-muted-foreground">{formatTime(ev)}</td>
                    <td className="py-2 pr-4 font-mono text-xs">{ev.user_id || "—"}</td>
                    <td className="py-2 pr-4">{ev.action}</td>
                    <td className="py-2 pr-4">{ev.resource}</td>
                    <td className="py-2">{ev.ip || "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {logs.length === 0 && !loading && (
              <p className="py-4 text-muted-foreground">No audit events.</p>
            )}
            <div className="mt-4 flex items-center gap-4">
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={currentPageIndex <= 0 || loading}
                onClick={handlePrevious}
              >
                Previous page
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {currentPageIndex + 1}
              </span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={!pageToken || loading}
                onClick={handleNext}
              >
                Next page
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
