"use client";

import Link from "next/link";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function Home() {
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
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-2xl">Zero Trust Control Plane</CardTitle>
            <CardDescription>
              Sign in or create an account to continue.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <Button asChild>
              <Link href="/login">Sign in</Link>
            </Button>
            <Button asChild variant="outline">
              <Link href="/register">Register</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Zero Trust Control Plane</CardTitle>
          <CardDescription>
            Logged in as user {user.user_id} in organization {user.org_id}.
          </CardDescription>
        </CardHeader>
        <CardFooter>
          <Button variant="outline" onClick={() => logout()}>
            Sign out
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
