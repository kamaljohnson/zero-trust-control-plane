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
import { MembershipRole, normalizeRole, type MembershipRoleType } from "@/lib/api/membership-roles";

const ROLE_LABEL: Record<number, string> = {
  [MembershipRole.UNSPECIFIED]: "—",
  [MembershipRole.OWNER]: "Owner",
  [MembershipRole.ADMIN]: "Admin",
  [MembershipRole.MEMBER]: "Member",
};

interface Member {
  id: string;
  user_id: string;
  org_id: string;
  /** Role from API: number (1–3) or string enum name (e.g. ROLE_MEMBER) */
  role: number | string;
  created_at?: { seconds?: string; nanos?: number };
}

interface Session {
  id: string;
  user_id: string;
  org_id: string;
  device_id: string;
  ip_address?: string;
  last_seen_at?: { seconds?: string; nanos?: number };
  created_at?: { seconds?: string; nanos?: number };
}

function authHeaders(accessToken: string): HeadersInit {
  return { Authorization: `Bearer ${accessToken}` };
}

export default function DashboardMembersPage() {
  const { user, accessToken, handleSessionInvalid } = useAuth();
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const [addEmail, setAddEmail] = useState("");
  const [addRole, setAddRole] = useState<MembershipRoleType>(MembershipRole.MEMBER);
  const [addSubmitting, setAddSubmitting] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);
  const [expandedUserId, setExpandedUserId] = useState<string | null>(null);
  const [sessionsByUser, setSessionsByUser] = useState<Record<string, Session[]>>({});
  const [revoking, setRevoking] = useState<string | null>(null);

  const orgId = user?.org_id ?? "";

  const loadMembers = useCallback(async () => {
    if (!accessToken || !orgId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(
        `/api/org-admin/members?org_id=${encodeURIComponent(orgId)}&page_size=100`,
        { headers: authHeaders(accessToken) }
      );
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to load members");
      setMembers(data.members ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load members");
    } finally {
      setLoading(false);
    }
  }, [accessToken, orgId, handleSessionInvalid]);

  useEffect(() => {
    loadMembers();
  }, [loadMembers]);

  const loadSessions = useCallback(
    async (userId: string) => {
      if (!accessToken || !orgId) return;
      try {
        const res = await fetch(
          `/api/org-admin/sessions?org_id=${encodeURIComponent(orgId)}&user_id=${encodeURIComponent(userId)}`,
          { headers: authHeaders(accessToken) }
        );
        if (res.status === 401) {
          handleSessionInvalid();
          return;
        }
        const data = await res.json();
        if (!res.ok) throw new Error(data.error ?? "Failed to load sessions");
        setSessionsByUser((prev) => ({ ...prev, [userId]: data.sessions ?? [] }));
      } catch {
        setSessionsByUser((prev) => ({ ...prev, [userId]: [] }));
      }
    },
    [accessToken, orgId, handleSessionInvalid]
  );

  const toggleSessions = (userId: string) => {
    if (expandedUserId === userId) {
      setExpandedUserId(null);
      return;
    }
    setExpandedUserId(userId);
    if (!sessionsByUser[userId]) loadSessions(userId);
  };

  const handleAdd = async () => {
    if (!accessToken || !orgId || !addEmail.trim()) return;
    setAddSubmitting(true);
    setAddError(null);
    try {
      const lookupRes = await fetch(
        `/api/users/by-email?email=${encodeURIComponent(addEmail.trim())}`,
        { headers: authHeaders(accessToken) }
      );
      if (lookupRes.status === 401) {
        handleSessionInvalid();
        return;
      }
      const lookupData = await lookupRes.json();
      if (!lookupRes.ok) throw new Error(lookupData.error ?? "User not found");
      const userId = lookupData.user?.id;
      if (!userId) throw new Error("User not found");

      const addRes = await fetch("/api/org-admin/members/add", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeaders(accessToken) },
        body: JSON.stringify({ org_id: orgId, user_id: userId, role: addRole }),
      });
      if (addRes.status === 401) {
        handleSessionInvalid();
        return;
      }
      const addData = await addRes.json();
      if (!addRes.ok) throw new Error(addData.error ?? "Failed to add member");
      setAddOpen(false);
      setAddEmail("");
      setAddRole(MembershipRole.MEMBER);
      loadMembers();
    } catch (e) {
      setAddError(e instanceof Error ? e.message : "Failed to add member");
    } finally {
      setAddSubmitting(false);
    }
  };

  const removeMember = async (userId: string) => {
    if (!accessToken || !orgId || !confirm("Remove this member from the organization?")) return;
    try {
      const res = await fetch("/api/org-admin/members/remove", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeaders(accessToken) },
        body: JSON.stringify({ org_id: orgId, user_id: userId }),
      });
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to remove");
      loadMembers();
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed to remove member");
    }
  };

  const updateRole = async (userId: string, role: number) => {
    if (!accessToken || !orgId) return;
    try {
      const res = await fetch("/api/org-admin/members/update-role", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeaders(accessToken) },
        body: JSON.stringify({ org_id: orgId, user_id: userId, role }),
      });
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to update role");
      loadMembers();
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed to update role");
    }
  };

  const revokeAllSessions = async (userId: string) => {
    if (!accessToken || !orgId || !confirm("Revoke all sessions for this user?")) return;
    setRevoking(userId);
    try {
      const res = await fetch("/api/org-admin/sessions/revoke-all", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeaders(accessToken) },
        body: JSON.stringify({ org_id: orgId, user_id: userId }),
      });
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? "Failed to revoke");
      setSessionsByUser((prev) => ({ ...prev, [userId]: [] }));
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed to revoke sessions");
    } finally {
      setRevoking(null);
    }
  };

  const revokeSession = async (sessionId: string, memberUserId: string) => {
    if (!accessToken || !confirm("Revoke this session?")) return;
    try {
      const res = await fetch("/api/org-admin/sessions/revoke", {
        method: "POST",
        headers: { "Content-Type": "application/json", ...authHeaders(accessToken) },
        body: JSON.stringify({ session_id: sessionId }),
      });
      if (res.status === 401) {
        handleSessionInvalid();
        return;
      }
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error ?? "Failed to revoke");
      }
      setSessionsByUser((prev) => ({
        ...prev,
        [memberUserId]: (prev[memberUserId] ?? []).filter((s) => s.id !== sessionId),
      }));
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed to revoke session");
    }
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Members</CardTitle>
          <CardDescription>Manage organization members and roles.</CardDescription>
        </div>
        <Button onClick={() => setAddOpen(true)}>Add member</Button>
      </CardHeader>
      <CardContent>
        {error && (
          <p className="mb-4 text-sm text-destructive">{error}</p>
        )}
        {loading ? (
          <p className="text-muted-foreground">Loading members…</p>
        ) : (
          <div className="space-y-2">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b">
                  <th className="pb-2 pr-4 font-medium">User ID</th>
                  <th className="pb-2 pr-4 font-medium">Role</th>
                  <th className="pb-2 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {members.map((m) => (
                  <React.Fragment key={m.id}>
                    <tr className="border-b">
                      <td className="py-2 pr-4 font-mono text-xs">{m.user_id}</td>
                      <td className="py-2 pr-4">
                        <select
                          value={normalizeRole(m.role)}
                          onChange={(e) => updateRole(m.user_id, Number(e.target.value))}
                          className="rounded border bg-background px-2 py-1"
                        >
                          {[MembershipRole.OWNER, MembershipRole.ADMIN, MembershipRole.MEMBER].map(
                            (r) => (
                              <option key={r} value={r}>
                                {ROLE_LABEL[r]}
                              </option>
                            )
                          )}
                        </select>
                      </td>
                      <td className="py-2 flex flex-wrap gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => toggleSessions(m.user_id)}
                        >
                          {expandedUserId === m.user_id ? "Hide" : "Sessions"}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => revokeAllSessions(m.user_id)}
                          disabled={revoking === m.user_id}
                        >
                          {revoking === m.user_id ? "Revoking…" : "Revoke all sessions"}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => removeMember(m.user_id)}
                          className="text-destructive"
                        >
                          Remove
                        </Button>
                      </td>
                    </tr>
                    {expandedUserId === m.user_id && (
                      <tr className="border-b bg-muted/30">
                        <td colSpan={3} className="py-3 pl-4">
                          <div className="text-xs font-medium text-muted-foreground mb-2">
                            Active sessions
                          </div>
                          {sessionsByUser[m.user_id] === undefined ? (
                            <p className="text-muted-foreground">Loading…</p>
                          ) : sessionsByUser[m.user_id].length === 0 ? (
                            <p className="text-muted-foreground">No active sessions.</p>
                          ) : (
                            <ul className="space-y-1">
                              {(sessionsByUser[m.user_id] ?? []).map((s) => (
                                <li
                                  key={s.id}
                                  className="flex items-center justify-between rounded border bg-background px-3 py-2"
                                >
                                  <span className="font-mono text-xs">
                                    {s.id.slice(0, 8)}… · {s.device_id?.slice(0, 8) ?? "—"} ·{" "}
                                    {s.ip_address || "—"}
                                  </span>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => revokeSession(s.id, m.user_id)}
                                  >
                                    Revoke
                                  </Button>
                                </li>
                              ))}
                            </ul>
                          )}
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
            {members.length === 0 && !loading && (
              <p className="py-4 text-muted-foreground">No members yet. Add one above.</p>
            )}
          </div>
        )}

        {addOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <Card className="w-full max-w-md">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Add member</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => setAddOpen(false)}>
                  Close
                </Button>
              </CardHeader>
              <CardContent className="space-y-4">
                {addError && (
                  <p className="text-sm text-destructive">{addError}</p>
                )}
                <div>
                  <Label htmlFor="add-email">Email</Label>
                  <Input
                    id="add-email"
                    type="email"
                    placeholder="user@example.com"
                    value={addEmail}
                    onChange={(e) => setAddEmail(e.target.value)}
                    className="mt-1"
                  />
                </div>
                <div>
                  <Label htmlFor="add-role">Role</Label>
                  <select
                    id="add-role"
                    value={addRole}
                    onChange={(e) => setAddRole(Number(e.target.value) as MembershipRoleType)}
                    className="mt-1 w-full rounded border bg-background px-3 py-2"
                  >
                    <option value={MembershipRole.ADMIN}>{ROLE_LABEL[MembershipRole.ADMIN]}</option>
                    <option value={MembershipRole.MEMBER}>{ROLE_LABEL[MembershipRole.MEMBER]}</option>
                  </select>
                </div>
                <Button onClick={handleAdd} disabled={addSubmitting}>
                  {addSubmitting ? "Adding…" : "Add member"}
                </Button>
              </CardContent>
            </Card>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
