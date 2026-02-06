"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import { SessionInfoDropdown } from "@/components/session-info-dropdown";
import { ProjectShowcaseAlert } from "@/components/project-showcase-alert";
import { MembershipRole, normalizeRole, type MembershipRoleType } from "@/lib/api/membership-roles";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface UserDetails {
  user: {
    id: string;
    email: string;
    name: string;
  };
}

interface OrgDetails {
  organization: {
    id: string;
    name: string;
  };
}

/**
 * Dashboard layout: requires auth, shows org-admin nav (Members, Audit, Policy).
 */
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, isAuthenticated, isLoading, logout, accessToken, handleSessionInvalid } = useAuth();
  const pathname = usePathname();
  const [userDetails, setUserDetails] = useState<UserDetails | null>(null);
  const [orgDetails, setOrgDetails] = useState<OrgDetails | null>(null);
  const [userRole, setUserRole] = useState<MembershipRoleType | null>(null);
  const [roleLoading, setRoleLoading] = useState(true);

  useEffect(() => {
    if (!isAuthenticated || !user || !accessToken) {
      return;
    }

    const fetchDetails = async () => {
      try {
        const [userRes, orgRes] = await Promise.all([
          fetch(`/api/users/${user.user_id}`, {
            headers: { Authorization: `Bearer ${accessToken}` },
          }),
          fetch(`/api/organizations/${user.org_id}`, {
            headers: { Authorization: `Bearer ${accessToken}` },
          }),
        ]);

        if (userRes.status === 401 || orgRes.status === 401) {
          handleSessionInvalid();
          return;
        }

        if (userRes.ok) {
          const userData = await userRes.json();
          setUserDetails(userData);
        }
        if (orgRes.ok) {
          const orgData = await orgRes.json();
          setOrgDetails(orgData);
        }
      } catch {
        // Silently fail - we'll show IDs as fallback
      }
    };

    fetchDetails();
  }, [isAuthenticated, user, accessToken, handleSessionInvalid]);

  // Fetch current user's membership role
  useEffect(() => {
    if (!isAuthenticated || !user || !accessToken) {
      setRoleLoading(false);
      return;
    }

    const fetchUserRole = async () => {
      setRoleLoading(true);
      try {
        const res = await fetch(
          `/api/org-admin/members?org_id=${encodeURIComponent(user.org_id)}&page_size=100`,
          { headers: { Authorization: `Bearer ${accessToken}` } }
        );

        if (res.status === 401) {
          handleSessionInvalid();
          return;
        }

        if (res.status === 403) {
          // Permission denied - user is not admin/owner
          setUserRole(MembershipRole.MEMBER);
          setRoleLoading(false);
          return;
        }

        if (!res.ok) {
          const data = await res.json();
          throw new Error(data.error ?? "Failed to fetch membership");
        }

        const data = await res.json();
        const members = data.members ?? [];
        
        // Find current user in members list
        const currentUserMember = members.find(
          (m: { user_id: string }) => m.user_id === user.user_id
        );

        if (currentUserMember) {
          const role = normalizeRole(currentUserMember.role);
          setUserRole(role);
        } else {
          // User not found in members list - treat as member
          setUserRole(MembershipRole.MEMBER);
        }
      } catch {
        // Default to member role on error to be safe
        setUserRole(MembershipRole.MEMBER);
      } finally {
        setRoleLoading(false);
      }
    };

    fetchUserRole();
  }, [isAuthenticated, user, accessToken, handleSessionInvalid]);

  if (isLoading || roleLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading…</p>
      </div>
    );
  }

  if (!isAuthenticated || !user) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <p className="text-muted-foreground">
          You must be signed in to access the dashboard.{" "}
          <Link href="/login" className="text-foreground underline">
            Sign in
          </Link>
        </p>
      </div>
    );
  }

  // Check if user is admin or owner
  const isAdminOrOwner = userRole === MembershipRole.ADMIN || userRole === MembershipRole.OWNER;

  // Use email from API if present and non-empty, otherwise fall back to user_id
  const userEmail = (userDetails?.user?.email && userDetails.user.email.trim()) 
    ? userDetails.user.email 
    : user.user_id;
  
  // Use name from backend API response (user.name field)
  // If name is not available or empty, extract from email as fallback
  const userName = (userDetails?.user?.name && userDetails.user.name.trim())
    ? userDetails.user.name
    : (userEmail.includes("@") ? userEmail.split("@")[0] : userEmail);
  
  const orgName = orgDetails?.organization?.name || "Organization";
  const orgId = user.org_id;

  const nav = [
    { href: "/dashboard", label: "Members" },
    { href: "/dashboard/audit", label: "Audit log" },
    { href: "/dashboard/policy", label: "Policy" },
  ];

  return (
    <div className="min-h-screen">
      <nav className="border-b bg-muted/30">
        <div className="mx-auto flex h-12 max-w-5xl items-center gap-6 px-4">
          <Link href="/" className="text-foreground font-medium hover:underline">
            Home
          </Link>
          <Link href="/dashboard" className="font-medium text-foreground hover:underline">
            Org admin
          </Link>
          {nav.map(({ href, label }) => (
            <Link
              key={href}
              href={href}
              className={
                pathname === href
                  ? "text-foreground font-medium underline"
                  : "text-muted-foreground hover:text-foreground"
              }
            >
              {label}
            </Link>
          ))}
          <div className="ml-auto flex items-center gap-6">
            <SessionInfoDropdown
              userName={userName}
              userEmail={userEmail}
              orgName={orgName}
              orgId={orgId}
            />
            <button
              type="button"
              onClick={() => logout()}
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              Sign out
            </button>
          </div>
        </div>
      </nav>
      <ProjectShowcaseAlert />
      <main className="mx-auto max-w-5xl p-4">
        {isAdminOrOwner ? (
          children
        ) : (
          <Alert variant="destructive" className="mt-4">
            <AlertTitle>Restricted Access</AlertTitle>
            <AlertDescription>
              Only admins and owners have access to this page.
            </AlertDescription>
          </Alert>
        )}
      </main>
    </div>
  );
}
