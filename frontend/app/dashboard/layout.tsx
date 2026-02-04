"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import { SessionInfoDropdown } from "@/components/session-info-dropdown";
import { ProjectShowcaseAlert } from "@/components/project-showcase-alert";

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
 * Dashboard layout: requires auth, shows org-admin nav (Members, Audit, Telemetry).
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
      } catch (err) {
        // Silently fail - we'll show IDs as fallback
      }
    };

    fetchDetails();
  }, [isAuthenticated, user, accessToken, handleSessionInvalid]);

  if (isLoading) {
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
    { href: "/dashboard/telemetry", label: "Telemetry" },
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
      <main className="mx-auto max-w-5xl p-4">{children}</main>
    </div>
  );
}
