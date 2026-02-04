"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/contexts/auth-context";

/**
 * Dashboard layout: requires auth, shows org-admin nav (Members, Audit, Telemetry).
 */
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, isAuthenticated, isLoading } = useAuth();
  const pathname = usePathname();

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
        </div>
      </nav>
      <main className="mx-auto max-w-5xl p-4">{children}</main>
    </div>
  );
}
