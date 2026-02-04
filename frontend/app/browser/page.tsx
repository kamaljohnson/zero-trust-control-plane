"use client";

import React, { useCallback, useEffect, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

interface AccessControl {
  allowed_domains?: string[];
  blocked_domains?: string[];
  wildcard_supported?: boolean;
  default_action?: number;
}

interface ActionRestrictions {
  allowed_actions?: string[];
  read_only_mode?: boolean;
}

interface BrowserPolicy {
  access_control?: AccessControl;
  action_restrictions?: ActionRestrictions;
}

const DEFAULT_POLICY: BrowserPolicy = {
  access_control: {},
  action_restrictions: {
    allowed_actions: ["navigate", "download", "upload", "copy_paste"],
    read_only_mode: false,
  },
};

/**
 * Normalize URL: add https if no scheme.
 */
function normalizeUrl(input: string): string {
  const s = input.trim();
  if (!s) return "";
  if (!/^[a-zA-Z][a-zA-Z0-9+.-]*:\/\//.test(s)) {
    return "https://" + s;
  }
  return s;
}

export default function BrowserPage() {
  const { user, accessToken, handleSessionInvalid } = useAuth();
  const [policy, setPolicy] = useState<BrowserPolicy | null>(null);
  const [policyError, setPolicyError] = useState<string | null>(null);
  const [urlInput, setUrlInput] = useState("");
  const [currentUrl, setCurrentUrl] = useState<string | null>(null);
  const [checkError, setCheckError] = useState<string | null>(null);
  const [checkReason, setCheckReason] = useState<string | null>(null);
  const [isChecking, setIsChecking] = useState(false);
  const [actionMessage, setActionMessage] = useState<string | null>(null);

  const orgId = user?.org_id ?? "";

  const fetchPolicy = useCallback(async () => {
    if (!accessToken || !orgId) return;
    setPolicyError(null);
    try {
      const res = await fetch(
        `/api/browser/policy?org_id=${encodeURIComponent(orgId)}`,
        { headers: { Authorization: `Bearer ${accessToken}` } }
      );
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        setPolicyError((data.error as string) || "Failed to load policy.");
        return;
      }
      const data = (await res.json()) as BrowserPolicy;
      setPolicy(data);
    } catch {
      setPolicyError("Could not load policy. Please try again.");
    }
  }, [accessToken, orgId, handleSessionInvalid]);

  useEffect(() => {
    if (orgId && accessToken) fetchPolicy();
  }, [orgId, accessToken, fetchPolicy]);

  const handleNavigate = useCallback(async () => {
    const raw = urlInput.trim();
    if (!raw) return;
    if (!accessToken || !orgId) return;
    const url = normalizeUrl(raw);
    setCheckError(null);
    setCheckReason(null);
    setIsChecking(true);
    setCurrentUrl(null);
    try {
      const res = await fetch("/api/browser/check-url", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({ url, org_id: orgId }),
      });
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      const data = (await res.json()) as { allowed?: boolean; reason?: string };
      if (!res.ok) {
        setCheckError((data as { error?: string }).error || "Could not check access. Please try again.");
        setCurrentUrl(null);
        return;
      }
      if (!data.allowed) {
        setCheckReason(data.reason || "Access to this URL is denied by your organization's policy.");
        setCurrentUrl(null);
        return;
      }
      setCheckReason(null);
      setCurrentUrl(url);
      setUrlInput(url);
    } catch {
      setCheckError("Could not check access. Please try again.");
      setCurrentUrl(null);
    } finally {
      setIsChecking(false);
    }
  }, [urlInput, accessToken, orgId, handleSessionInvalid]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleNavigate();
  };

  const allowedActions = policy?.action_restrictions?.allowed_actions ?? DEFAULT_POLICY.action_restrictions!.allowed_actions!;
  const readOnlyMode = policy?.action_restrictions?.read_only_mode ?? false;

  const canDownload = allowedActions.includes("download");
  const canUpload = allowedActions.includes("upload") && !readOnlyMode;
  const canCopyPaste = allowedActions.includes("copy_paste") && !readOnlyMode;

  const handleDownload = () => setActionMessage("Download is not yet implemented. Policy allows download.");
  const handleUpload = () => setActionMessage("Upload is not yet implemented. Policy allows upload.");
  const handleCopyPaste = () => setActionMessage("Copy/paste is not yet implemented. Policy allows copy/paste.");

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-4 p-4">
      {readOnlyMode && (
        <div className="rounded-md border border-amber-500/50 bg-amber-500/10 px-4 py-2 text-sm text-amber-800 dark:text-amber-200">
          Read-only mode — upload and copy/paste are disabled by your organization&apos;s policy.
        </div>
      )}

      <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
        <div className="flex-1 space-y-2">
          <Label htmlFor="browser-url">URL</Label>
          <Input
            id="browser-url"
            type="url"
            placeholder="https://example.com"
            value={urlInput}
            onChange={(e) => setUrlInput(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={!orgId}
            className="font-mono"
          />
        </div>
        <Button onClick={handleNavigate} disabled={isChecking || !urlInput.trim() || !orgId}>
          {isChecking ? "Checking…" : "Go"}
        </Button>
      </div>

      <div className="flex flex-wrap gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={!canDownload}
          onClick={handleDownload}
        >
          Download
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={!canUpload}
          onClick={handleUpload}
        >
          Upload
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={!canCopyPaste}
          onClick={handleCopyPaste}
        >
          Copy/paste
        </Button>
      </div>

      {actionMessage && (
        <p className="text-sm text-muted-foreground">{actionMessage}</p>
      )}

      {policyError && (
        <Card className="border-destructive/50 bg-destructive/5">
          <CardHeader className="pb-2">
            <p className="text-sm font-medium text-destructive">Policy error</p>
          </CardHeader>
          <CardContent className="pt-0 text-sm text-muted-foreground">
            {policyError}
          </CardContent>
        </Card>
      )}

      {(checkError || checkReason) && (
        <Card className="border-destructive/50 bg-destructive/5">
          <CardHeader className="pb-2">
            <p className="text-sm font-medium text-destructive">Access denied</p>
          </CardHeader>
          <CardContent className="pt-0 text-sm text-muted-foreground">
            {checkReason || checkError || "Access to this URL is denied by your organization's policy."}
          </CardContent>
        </Card>
      )}

      <div className="min-h-[400px] flex-1 rounded-lg border bg-muted/20 overflow-hidden flex flex-col">
        {!currentUrl && !isChecking && (
          <div className="flex flex-1 items-center justify-center p-8 text-muted-foreground">
            Enter a URL to browse. Access is checked against your organization&apos;s policy; allowed URLs open in a new tab.
          </div>
        )}
        {isChecking && (
          <div className="flex flex-1 items-center justify-center p-8 text-muted-foreground">
            Checking access…
          </div>
        )}
        {currentUrl && !isChecking && (
          <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
            <p className="text-muted-foreground">
              Access allowed. Most sites block being shown inside this page, so open the link in a new tab to browse.
            </p>
            <a
              href={currentUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              Open in new tab
            </a>
            <p className="text-muted-foreground text-sm break-all max-w-lg">
              {currentUrl}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
