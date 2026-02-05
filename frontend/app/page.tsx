"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ProjectShowcaseAlert } from "@/components/project-showcase-alert";

export default function Home() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted && !isLoading && isAuthenticated && user) {
      router.replace("/browser");
    }
  }, [mounted, isLoading, isAuthenticated, user, router]);

  // Always render the same structure to avoid hydration mismatch
  // Show loading state only after component has mounted on client
  if (!mounted || isLoading || (isAuthenticated && user)) {
    return (
      <div className="flex min-h-screen flex-col p-4">
        <div className="flex-1 flex items-center justify-center">
          <p className="text-muted-foreground">Loading…</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col p-4">
      <div className="flex-1 flex items-center justify-center">
        <div className="w-full max-w-md space-y-4">
          <ProjectShowcaseAlert />
          <Card>
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
      </div>
    </div>
  );
}
