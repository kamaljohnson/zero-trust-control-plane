"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import Link from "next/link";
import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import * as authClient from "@/lib/auth-client";

const EMAIL_REGEX = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

function validatePassword(password: string): string | null {
  if (password.length < 12) {
    return "Password must be at least 12 characters.";
  }
  let hasUpper = false;
  let hasLower = false;
  let hasNumber = false;
  let hasSymbol = false;
  for (const ch of password) {
    if (ch >= "A" && ch <= "Z") hasUpper = true;
    else if (ch >= "a" && ch <= "z") hasLower = true;
    else if (ch >= "0" && ch <= "9") hasNumber = true;
    else hasSymbol = true;
  }
  if (!hasUpper) return "Password must contain at least one uppercase letter.";
  if (!hasLower) return "Password must contain at least one lowercase letter.";
  if (!hasNumber) return "Password must contain at least one number.";
  if (!hasSymbol) return "Password must contain at least one symbol.";
  return null;
}

export default function RegisterPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [userId, setUserId] = useState<string | null>(null);
  const [orgName, setOrgName] = useState("");
  const [creatingOrg, setCreatingOrg] = useState(false);
  const [orgCreated, setOrgCreated] = useState(false);
  const [createdOrgId, setCreatedOrgId] = useState<string | null>(null);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading…</p>
      </div>
    );
  }

  if (isAuthenticated) {
    router.replace("/");
    return null;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const trimmedEmail = email.trim().toLowerCase();
    if (!EMAIL_REGEX.test(trimmedEmail)) {
      setError("Invalid email format.");
      return;
    }
    const pwdError = validatePassword(password);
    if (pwdError) {
      setError(pwdError);
      return;
    }
    setSubmitting(true);
    try {
      const res = await authClient.register(trimmedEmail, password, name.trim() || undefined);
      if (res.user_id) {
        setUserId(res.user_id);
        setSuccess(true);
      } else {
        setError("Registration succeeded but user_id was not returned.");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed.");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleCreateOrganization(e: React.FormEvent) {
    e.preventDefault();
    if (!userId || !orgName.trim()) {
      setError("Organization name is required.");
      return;
    }
    setError(null);
    setCreatingOrg(true);
    try {
      const res = await fetch("/api/organization/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: orgName.trim(), user_id: userId }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "Organization creation failed.");
      }
      if (data.organization?.id) {
        setCreatedOrgId(data.organization.id);
        setOrgCreated(true);
      } else {
        setError("Organization created but ID was not returned.");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Organization creation failed.");
    } finally {
      setCreatingOrg(false);
    }
  }

  if (orgCreated && createdOrgId) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>Organization created</CardTitle>
            <CardDescription>
              Your organization has been created. You can now sign in with your organization ID.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-md border border-muted bg-muted/50 p-3">
              <p className="text-sm font-medium">Organization ID:</p>
              <p className="text-sm font-mono text-muted-foreground break-all">{createdOrgId}</p>
            </div>
          </CardContent>
          <CardFooter>
            <Button
              asChild
              className="w-full"
              onClick={() => {
                router.push(`/login?org_id=${encodeURIComponent(createdOrgId)}`);
              }}
            >
              <Link href={`/login?org_id=${encodeURIComponent(createdOrgId)}`}>
                Go to sign in
              </Link>
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  if (success && userId) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>Account created</CardTitle>
            <CardDescription>
              Create an organization to get started, or sign in to an existing organization.
            </CardDescription>
          </CardHeader>
          <form onSubmit={handleCreateOrganization}>
            <CardContent className="space-y-4">
              {error && (
                <p className="text-sm text-destructive" role="alert">
                  {error}
                </p>
              )}
              <div className="space-y-2">
                <Label htmlFor="org-name">Organization name</Label>
                <Input
                  id="org-name"
                  type="text"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  required
                  placeholder="My Organization"
                  disabled={creatingOrg}
                />
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button type="submit" className="w-full" disabled={creatingOrg || !orgName.trim()}>
                {creatingOrg ? "Creating organization…" : "Create organization"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="w-full"
                disabled={creatingOrg}
                asChild
              >
                <Link href="/login">Sign in to existing organization</Link>
              </Button>
            </CardFooter>
          </form>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Create an account</CardTitle>
          <CardDescription>
            Password: 12+ characters, upper, lower, number, symbol.
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
            {error && (
              <p className="text-sm text-destructive" role="alert">
                {error}
              </p>
            )}
            <div className="space-y-2">
              <Label htmlFor="register-email">Email</Label>
              <Input
                id="register-email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                placeholder="you@example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="register-password">Password</Label>
              <Input
                id="register-password"
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                placeholder="••••••••••••"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="register-name">Name (optional)</Label>
              <Input
                id="register-name"
                type="text"
                autoComplete="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Your name"
              />
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-4">
            <Button type="submit" className="w-full" disabled={submitting}>
              {submitting ? "Creating account…" : "Create account"}
            </Button>
            <p className="text-center text-sm text-muted-foreground">
              Already have an account?{" "}
              <Link href="/login" className="text-primary underline-offset-4 hover:underline">
                Sign in
              </Link>
            </p>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
