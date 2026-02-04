"use client";

import React, { useCallback, useEffect, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { DomainListEditor } from "@/components/domain-list-editor";

interface AuthMfa {
  mfa_requirement?: number;
  allowed_mfa_methods?: string[];
  step_up_sensitive_actions?: boolean;
  step_up_policy_violation?: boolean;
}

interface DeviceTrust {
  device_registration_allowed?: boolean;
  auto_trust_after_mfa?: boolean;
  max_trusted_devices_per_user?: number;
  reverify_interval_days?: number;
  admin_revoke_allowed?: boolean;
}

interface SessionMgmt {
  session_max_ttl?: string;
  idle_timeout?: string;
  concurrent_session_limit?: number;
  admin_forced_logout?: boolean;
  reauth_on_policy_change?: boolean;
}

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

interface OrgPolicyConfig {
  auth_mfa?: AuthMfa;
  device_trust?: DeviceTrust;
  session_mgmt?: SessionMgmt;
  access_control?: AccessControl;
  action_restrictions?: ActionRestrictions;
}

function authHeaders(accessToken: string): HeadersInit {
  return { Authorization: `Bearer ${accessToken}` };
}

const MFA_REQUIREMENT_OPTIONS = [
  { value: 1, label: "Always" },
  { value: 2, label: "On new device" },
  { value: 3, label: "On untrusted device" },
];

const DEFAULT_ACTION_OPTIONS = [
  { value: 1, label: "Allow" },
  { value: 2, label: "Deny" },
];

const ACTION_OPTIONS = ["navigate", "download", "upload", "copy_paste"];

function emptyConfig(): OrgPolicyConfig {
  return {
    auth_mfa: {
      mfa_requirement: 2,
      allowed_mfa_methods: ["sms_otp"],
      step_up_sensitive_actions: false,
      step_up_policy_violation: false,
    },
    device_trust: {
      device_registration_allowed: true,
      auto_trust_after_mfa: true,
      max_trusted_devices_per_user: 0,
      reverify_interval_days: 30,
      admin_revoke_allowed: true,
    },
    session_mgmt: {
      session_max_ttl: "24h",
      idle_timeout: "30m",
      concurrent_session_limit: 0,
      admin_forced_logout: true,
      reauth_on_policy_change: false,
    },
    access_control: {
      allowed_domains: [],
      blocked_domains: [],
      wildcard_supported: false,
      default_action: 1,
    },
    action_restrictions: {
      allowed_actions: ["navigate", "download", "upload", "copy_paste"],
      read_only_mode: false,
    },
  };
}

function mfaRequirementToNumber(v: number | string | undefined): number {
  if (typeof v === "number" && v >= 1 && v <= 3) return v;
  if (typeof v === "string") {
    if (v === "MFA_REQUIREMENT_ALWAYS" || v === "always") return 1;
    if (v === "MFA_REQUIREMENT_NEW_DEVICE" || v === "new_device") return 2;
    if (v === "MFA_REQUIREMENT_UNTRUSTED" || v === "untrusted") return 3;
  }
  return 2;
}

function defaultActionToNumber(v: number | string | undefined): number {
  if (typeof v === "number" && (v === 1 || v === 2)) return v;
  if (typeof v === "string") {
    if (v === "DEFAULT_ACTION_ALLOW" || v === "allow") return 1;
    if (v === "DEFAULT_ACTION_DENY" || v === "deny") return 2;
  }
  return 1;
}

function mergeConfig(loaded: OrgPolicyConfig | null): OrgPolicyConfig {
  const def = emptyConfig();
  if (!loaded) return def;
  const auth = { ...def.auth_mfa, ...loaded.auth_mfa } as AuthMfa;
  auth.mfa_requirement = mfaRequirementToNumber(loaded.auth_mfa?.mfa_requirement ?? auth.mfa_requirement);
  const access = { ...def.access_control, ...loaded.access_control } as AccessControl;
  access.default_action = defaultActionToNumber(loaded.access_control?.default_action ?? access.default_action);
  return {
    auth_mfa: auth,
    device_trust: { ...def.device_trust, ...loaded.device_trust },
    session_mgmt: { ...def.session_mgmt, ...loaded.session_mgmt },
    access_control: access,
    action_restrictions: { ...def.action_restrictions, ...loaded.action_restrictions },
  };
}

export default function DashboardPolicyPage() {
  const { user, accessToken, handleSessionInvalid } = useAuth();
  const [config, setConfig] = useState<OrgPolicyConfig>(emptyConfig());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saveMessage, setSaveMessage] = useState<string | null>(null);

  const orgId = user?.org_id ?? "";

  const loadConfig = useCallback(async () => {
    if (!accessToken || !orgId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(
        `/api/org-admin/policy-config?org_id=${encodeURIComponent(orgId)}`,
        { headers: authHeaders(accessToken) }
      );
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to load policy config");
      setConfig(mergeConfig(data.config ?? null));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load policy config");
    } finally {
      setLoading(false);
    }
  }, [accessToken, orgId, handleSessionInvalid]);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const saveConfig = async () => {
    if (!accessToken || !orgId) return;
    setSaving(true);
    setSaveMessage(null);
    setError(null);
    try {
      const res = await fetch("/api/org-admin/policy-config", {
        method: "PUT",
        headers: { "Content-Type": "application/json", ...authHeaders(accessToken) },
        body: JSON.stringify({ org_id: orgId, config }),
      });
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to save");
      setConfig(mergeConfig(data.config ?? null));
      setSaveMessage("Policy config saved.");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  };

  const updateAuthMfa = (patch: Partial<AuthMfa>) => {
    setConfig((c) => ({ ...c, auth_mfa: { ...c.auth_mfa, ...patch } as AuthMfa }));
  };
  const updateDeviceTrust = (patch: Partial<DeviceTrust>) => {
    setConfig((c) => ({ ...c, device_trust: { ...c.device_trust, ...patch } as DeviceTrust }));
  };
  const updateSessionMgmt = (patch: Partial<SessionMgmt>) => {
    setConfig((c) => ({ ...c, session_mgmt: { ...c.session_mgmt, ...patch } as SessionMgmt }));
  };
  const updateAccessControl = (patch: Partial<AccessControl>) => {
    setConfig((c) => ({ ...c, access_control: { ...c.access_control, ...patch } as AccessControl }));
  };
  const updateActionRestrictions = (patch: Partial<ActionRestrictions>) => {
    setConfig((c) => ({
      ...c,
      action_restrictions: { ...c.action_restrictions, ...patch } as ActionRestrictions,
    }));
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-muted-foreground">Loading policy config…</p>
        </CardContent>
      </Card>
    );
  }

  const auth = config.auth_mfa ?? emptyConfig().auth_mfa!;
  const device = config.device_trust ?? emptyConfig().device_trust!;
  const session = config.session_mgmt ?? emptyConfig().session_mgmt!;
  const access = config.access_control ?? emptyConfig().access_control!;
  const actions = config.action_restrictions ?? emptyConfig().action_restrictions!;

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Policy config</CardTitle>
            <CardDescription>Organization policy: auth, device trust, session, access control, actions.</CardDescription>
          </div>
          <Button onClick={saveConfig} disabled={saving}>
            {saving ? "Saving…" : "Save all"}
          </Button>
        </CardHeader>
        <CardContent className="space-y-6">
          {error && <p className="text-sm text-destructive">{error}</p>}
          {saveMessage && <p className="text-sm text-green-600 dark:text-green-400">{saveMessage}</p>}

          {/* 1. Authentication & MFA */}
          <div className="space-y-3 rounded-lg border p-4">
            <h3 className="font-medium">Authentication & MFA</h3>
            <div>
              <Label>MFA requirement</Label>
              <select
                value={auth.mfa_requirement ?? 2}
                onChange={(e) => updateAuthMfa({ mfa_requirement: Number(e.target.value) })}
                className="mt-1 w-full rounded border bg-background px-3 py-2"
              >
                {MFA_REQUIREMENT_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <Label className="block mb-2">Allowed MFA methods</Label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={(auth.allowed_mfa_methods ?? []).includes("sms_otp")}
                  onChange={(e) =>
                    updateAuthMfa({
                      allowed_mfa_methods: e.target.checked
                        ? [...(auth.allowed_mfa_methods ?? []), "sms_otp"]
                        : (auth.allowed_mfa_methods ?? []).filter((m) => m !== "sms_otp"),
                    })
                  }
                />
                SMS OTP
              </label>
            </div>
            <div className="flex gap-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={auth.step_up_sensitive_actions ?? false}
                  onChange={(e) => updateAuthMfa({ step_up_sensitive_actions: e.target.checked })}
                />
                Step-up on sensitive actions
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={auth.step_up_policy_violation ?? false}
                  onChange={(e) => updateAuthMfa({ step_up_policy_violation: e.target.checked })}
                />
                Step-up on policy violation
              </label>
            </div>
          </div>

          {/* 2. Device Trust */}
          <div className="space-y-3 rounded-lg border p-4">
            <h3 className="font-medium">Device Trust</h3>
            <div className="flex flex-wrap gap-6">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={device.device_registration_allowed ?? true}
                  onChange={(e) => updateDeviceTrust({ device_registration_allowed: e.target.checked })}
                />
                Device registration allowed
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={device.auto_trust_after_mfa ?? true}
                  onChange={(e) => updateDeviceTrust({ auto_trust_after_mfa: e.target.checked })}
                />
                Auto-trust after MFA
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={device.admin_revoke_allowed ?? true}
                  onChange={(e) => updateDeviceTrust({ admin_revoke_allowed: e.target.checked })}
                />
                Admin revoke allowed
              </label>
            </div>
            <div className="grid grid-cols-2 gap-4 max-w-md">
              <div>
                <Label>Max trusted devices per user (0 = unlimited)</Label>
                <Input
                  type="number"
                  min={0}
                  value={device.max_trusted_devices_per_user ?? 0}
                  onChange={(e) => updateDeviceTrust({ max_trusted_devices_per_user: parseInt(e.target.value, 10) || 0 })}
                  className="mt-1"
                />
              </div>
              <div>
                <Label>Re-verification interval (days)</Label>
                <Input
                  type="number"
                  min={1}
                  value={device.reverify_interval_days ?? 30}
                  onChange={(e) => updateDeviceTrust({ reverify_interval_days: parseInt(e.target.value, 10) || 30 })}
                  className="mt-1"
                />
              </div>
            </div>
          </div>

          {/* 3. Session Management */}
          <div className="space-y-3 rounded-lg border p-4">
            <h3 className="font-medium">Session Management</h3>
            <div className="grid grid-cols-2 gap-4 max-w-md">
              <div>
                <Label>Session max TTL (e.g. 24h)</Label>
                <Input
                  value={session.session_max_ttl ?? "24h"}
                  onChange={(e) => updateSessionMgmt({ session_max_ttl: e.target.value })}
                  className="mt-1"
                />
              </div>
              <div>
                <Label>Idle timeout (e.g. 30m)</Label>
                <Input
                  value={session.idle_timeout ?? "30m"}
                  onChange={(e) => updateSessionMgmt({ idle_timeout: e.target.value })}
                  className="mt-1"
                />
              </div>
              <div>
                <Label>Concurrent session limit (0 = unlimited)</Label>
                <Input
                  type="number"
                  min={0}
                  value={session.concurrent_session_limit ?? 0}
                  onChange={(e) => updateSessionMgmt({ concurrent_session_limit: parseInt(e.target.value, 10) || 0 })}
                  className="mt-1"
                />
              </div>
            </div>
            <div className="flex gap-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={session.admin_forced_logout ?? true}
                  onChange={(e) => updateSessionMgmt({ admin_forced_logout: e.target.checked })}
                />
                Admin forced logout
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={session.reauth_on_policy_change ?? false}
                  onChange={(e) => updateSessionMgmt({ reauth_on_policy_change: e.target.checked })}
                />
                Re-auth on policy change
              </label>
            </div>
          </div>

          {/* 4. Access Control (Browser) */}
          <div className="space-y-3 rounded-lg border p-4">
            <h3 className="font-medium">Access Control (Browser)</h3>
            <DomainListEditor
              label="Allowed domains"
              value={access.allowed_domains ?? []}
              onChange={(allowed_domains) => updateAccessControl({ allowed_domains })}
              aria-label="Allowed domains list"
            />
            <DomainListEditor
              label="Blocked domains"
              value={access.blocked_domains ?? []}
              onChange={(blocked_domains) => updateAccessControl({ blocked_domains })}
              aria-label="Blocked domains list"
            />
            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={access.wildcard_supported ?? false}
                  onChange={(e) => updateAccessControl({ wildcard_supported: e.target.checked })}
                />
                Wildcard support
              </label>
              <div className="flex items-center gap-2">
                <Label className="mb-0">Default action</Label>
                <select
                  value={access.default_action ?? 1}
                  onChange={(e) => updateAccessControl({ default_action: Number(e.target.value) })}
                  className="rounded border bg-background px-2 py-1"
                >
                  {DEFAULT_ACTION_OPTIONS.map((o) => (
                    <option key={o.value} value={o.value}>
                      {o.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* 5. Action Restrictions */}
          <div className="space-y-3 rounded-lg border p-4">
            <h3 className="font-medium">Action Restrictions</h3>
            <div className="flex flex-wrap gap-4">
              {ACTION_OPTIONS.map((action) => (
                <label key={action} className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={(actions.allowed_actions ?? []).includes(action)}
                    onChange={(e) => {
                      const current = actions.allowed_actions ?? [];
                      updateActionRestrictions({
                        allowed_actions: e.target.checked
                          ? [...current, action]
                          : current.filter((a) => a !== action),
                      });
                    }}
                  />
                  {action.replace("_", " / ")}
                </label>
              ))}
            </div>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={actions.read_only_mode ?? false}
                onChange={(e) => updateActionRestrictions({ read_only_mode: e.target.checked })}
              />
              Read-only mode (no uploads/downloads)
            </label>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
