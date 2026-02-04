"use client";

import Link from "next/link";
import { useAuth } from "@/contexts/auth-context";

/**
 * Browser layout: requires auth, shows nav (Home, Org admin, user/org, Sign out).
 */
export default function BrowserLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, isAuthenticated, isLoading, logout } = useAuth();

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
          You must be signed in to use the browser.{" "}
          <Link href="/login" className="text-foreground underline">
            Sign in
          </Link>
        </p>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col">
      <nav className="border-b bg-muted/30 shrink-0">
        <div className="mx-auto flex h-12 max-w-5xl items-center gap-6 px-4">
          <Link href="/" className="text-foreground font-medium hover:underline">
            Home
          </Link>
          <Link href="/dashboard" className="text-muted-foreground hover:text-foreground">
            Org admin
          </Link>
          <div className="ml-auto flex items-center gap-4">
            <span className="text-muted-foreground text-sm">
              User {user.user_id} · Org {user.org_id}
            </span>
            <button
              type="button"
              onClick={() => logout()}
              className="text-muted-foreground hover:text-foreground"
            >
              Sign out
            </button>
          </div>
        </div>
      </nav>
      <main className="flex-1 flex flex-col min-h-0">{children}</main>
    </div>
  );
}
