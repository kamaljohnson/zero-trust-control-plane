"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState, useEffect, Suspense } from "react";
import { useAuth } from "@/contexts/auth-context";
import * as authClient from "@/lib/auth-client";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const EMAIL_REGEX = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
const DEFAULT_ORG_ID = process.env.NEXT_PUBLIC_DEFAULT_ORG_ID ?? "";
const DEV_OTP_ENABLED_BUILD =
  process.env.NEXT_PUBLIC_DEV_OTP_ENABLED === "true" || process.env.NEXT_PUBLIC_DEV_OTP_ENABLED === "1";

function LoginPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, verifyMFA, isAuthenticated, isLoading, setAuthFromResponse } = useAuth();
  const [mode, setMode] = useState<"signin" | "createOrg">("signin");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  // Initialize with DEFAULT_ORG_ID to avoid hydration mismatch; sync from searchParams in useEffect
  const [orgId, setOrgId] = useState(DEFAULT_ORG_ID);
  const [orgName, setOrgName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [creatingOrg, setCreatingOrg] = useState(false);
  const [phoneIntentId, setPhoneIntentId] = useState<string | null>(null);
  const [phone, setPhone] = useState("");
  const [requestingMfa, setRequestingMfa] = useState(false);
  const [mfaChallengeId, setMfaChallengeId] = useState<string | null>(null);
  const [mfaPhoneMask, setMfaPhoneMask] = useState<string | null>(null);
  const [mfaOtp, setMfaOtp] = useState<string | null>(null);
  const [mfaOtpNote, setMfaOtpNote] = useState<string | null>(null);
  const [otp, setOtp] = useState("");
  const [verifying, setVerifying] = useState(false);
  const [mounted, setMounted] = useState(false);
  // Runtime dev OTP flag from GET /api/config (so PoC can enable in prod without rebuild)
  const [devOtpEnabled, setDevOtpEnabled] = useState(DEV_OTP_ENABLED_BUILD);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Fetch runtime config so dev OTP can be enabled in prod via DEV_OTP_ENABLED env
  useEffect(() => {
    if (!mounted) return;
    let cancelled = false;
    fetch("/api/config")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (!cancelled && data && typeof data.devOtpEnabled === "boolean") {
          setDevOtpEnabled(data.devOtpEnabled);
        }
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [mounted]);

  useEffect(() => {
    if (mounted && !isLoading && isAuthenticated) {
      router.replace("/");
    }
  }, [mounted, isLoading, isAuthenticated, router]);

  // Sync orgId from query params
  useEffect(() => {
    const orgIdFromQuery = searchParams?.get("org_id") ?? "";
    if (orgIdFromQuery) {
      setOrgId(orgIdFromQuery);
    }
  }, [searchParams]);

  // On mount: if we were redirected from refresh with MFA required, restore challenge/intent from sessionStorage
  useEffect(() => {
    if (typeof window === "undefined") return;
    const challengeId = window.sessionStorage.getItem("refresh_mfa_challenge_id");
    const phoneMask = window.sessionStorage.getItem("refresh_mfa_phone_mask");
    const intentId = window.sessionStorage.getItem("refresh_mfa_intent_id");
    if (challengeId) {
      window.sessionStorage.removeItem("refresh_mfa_challenge_id");
      window.sessionStorage.removeItem("refresh_mfa_phone_mask");
      setMfaChallengeId(challengeId);
      setMfaPhoneMask(phoneMask ?? null);
      setMfaOtp(null);
      setMfaOtpNote(null);
      setOtp("");
    } else if (intentId) {
      window.sessionStorage.removeItem("refresh_mfa_intent_id");
      setPhoneIntentId(intentId);
    }
  }, []);

  // When MFA step is shown and dev OTP is enabled, fetch OTP from GET /api/dev/mfa/otp
  useEffect(() => {
    if (!mfaChallengeId || !devOtpEnabled) return;
    let cancelled = false;
    (async () => {
      try {
        const res = await fetch(`/api/dev/mfa/otp?challenge_id=${encodeURIComponent(mfaChallengeId)}`);
        if (!cancelled && res.ok) {
          const data = await res.json();
          if (data.otp) {
            setMfaOtp(data.otp);
            setMfaOtpNote(data.note ?? null);
            setOtp(data.otp);
          }
        }
      } catch {
        // ignore; leave mfaOtp null
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [mfaChallengeId]);

  // Use same outer structure as main content to avoid hydration mismatch (isLoading true on server, false on client).
  if (!mounted || isLoading || isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="w-full max-w-sm flex items-center justify-center">
          <p className="text-muted-foreground">Loading…</p>
        </div>
      </div>
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const trimmedEmail = email.trim().toLowerCase();
    if (!EMAIL_REGEX.test(trimmedEmail)) {
      setError("Invalid email format.");
      return;
    }
    
    if (mode === "createOrg") {
      await handleCreateOrganization(trimmedEmail);
      return;
    }
    
    setSubmitting(true);
    try {
      const res = await login(trimmedEmail, password, orgId.trim());
      if (res.phone_required === true && res.intent_id) {
        setPhoneIntentId(res.intent_id);
      } else if (res.mfa_required === true && res.challenge_id) {
        setMfaChallengeId(res.challenge_id);
        setMfaPhoneMask(res.phone_mask ?? null);
        setMfaOtp(null);
        setMfaOtpNote(null);
        setOtp("");
      } else {
        router.replace("/");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed.");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleCreateOrganization(email: string) {
    if (!orgName.trim()) {
      setError("Organization name is required.");
      return;
    }
    setCreatingOrg(true);
    setError(null);
    try {
      // Verify credentials and get user_id (no session created)
      const verifyRes = await fetch("/api/auth/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      const verifyData = await verifyRes.json();
      if (!verifyRes.ok) {
        throw new Error(verifyData.error ?? "Invalid email or password.");
      }
      const userId = verifyData.user_id;
      if (!userId) {
        throw new Error("Verification did not return user_id.");
      }

      // Create organization with the verified user_id
      const createRes = await fetch("/api/organization/create", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: orgName.trim(), user_id: userId }),
      });
      const createData = await createRes.json();
      if (!createRes.ok) {
        throw new Error(createData.error ?? "Organization creation failed.");
      }

      if (createData.organization?.id) {
        // Organization created; sign user in with the new org_id
        const loginRes = await login(email, password, createData.organization.id);
        if (loginRes.phone_required === true && loginRes.intent_id) {
          setPhoneIntentId(loginRes.intent_id);
        } else if (loginRes.mfa_required === true && loginRes.challenge_id) {
          setMfaChallengeId(loginRes.challenge_id);
          setMfaPhoneMask(loginRes.phone_mask ?? null);
          setMfaOtp(null);
          setMfaOtpNote(null);
          setOtp("");
        } else if (loginRes.access_token && loginRes.refresh_token && loginRes.user_id && loginRes.org_id) {
          setAuthFromResponse(loginRes);
          router.replace("/");
        }
      } else {
        setError("Organization created but ID was not returned.");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Organization creation failed.");
    } finally {
      setCreatingOrg(false);
    }
  }

  async function handleSubmitPhone(e: React.FormEvent) {
    e.preventDefault();
    if (!phoneIntentId || !phone.trim()) return;
    setError(null);
    setRequestingMfa(true);
    try {
      const res = await authClient.requestMFAWithPhone(phoneIntentId, phone.trim());
      setPhoneIntentId(null);
      setPhone("");
      setMfaChallengeId(res.challenge_id);
      setMfaPhoneMask(res.phone_mask ?? null);
      setMfaOtp(null);
      setMfaOtpNote(null);
      setOtp("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed.");
    } finally {
      setRequestingMfa(false);
    }
  }

  async function handleVerifyMFA(e: React.FormEvent) {
    e.preventDefault();
    if (!mfaChallengeId || !otp.trim()) return;
    setError(null);
    setVerifying(true);
    try {
      await verifyMFA(mfaChallengeId, otp.trim());
      router.replace("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "MFA verification failed.");
    } finally {
      setVerifying(false);
    }
  }

  if (phoneIntentId != null) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>Enter your phone number</CardTitle>
            <CardDescription>
              We need your phone number to send a verification code. Use digits only (e.g. with country code).
            </CardDescription>
          </CardHeader>
          <form onSubmit={handleSubmitPhone}>
            <CardContent className="space-y-4">
              {error && (
                <p className="text-sm text-destructive" role="alert">
                  {error}
                </p>
              )}
              <div className="space-y-2">
                <Label htmlFor="login-phone">Phone number</Label>
                <Input
                  id="login-phone"
                  type="tel"
                  inputMode="numeric"
                  autoComplete="tel"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value.replace(/\D/g, "").trim())}
                  placeholder="e.g. 15551234567"
                  required
                />
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button type="submit" className="w-full" disabled={requestingMfa || phone.length < 10}>
                {requestingMfa ? "Sending code…" : "Continue"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="w-full"
                disabled={requestingMfa}
                onClick={() => {
                  setPhoneIntentId(null);
                  setPhone("");
                  setError(null);
                }}
              >
                Back to sign in
              </Button>
            </CardFooter>
          </form>
        </Card>
      </div>
    );
  }

  if (mfaChallengeId != null) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>Verify your identity</CardTitle>
            <CardDescription>
              {mfaOtp != null
                ? "Your verification code is shown below (PoC mode; SMS not sent)."
                : `Enter the 6-digit code sent to your phone${mfaPhoneMask ? ` (${mfaPhoneMask})` : ""}.`}
            </CardDescription>
          </CardHeader>
          <form onSubmit={handleVerifyMFA}>
            <CardContent className="space-y-4">
              {mfaOtp != null && (
                <div className="rounded-md border border-muted bg-muted/50 p-3 text-center">
                  <p className="text-2xl font-mono font-semibold tracking-widest">{mfaOtp}</p>
                  {mfaOtpNote != null && (
                    <p className="mt-2 text-xs text-muted-foreground">{mfaOtpNote}</p>
                  )}
                </div>
              )}
              {error && (
                <p className="text-sm text-destructive" role="alert">
                  {error}
                </p>
              )}
              <div className="space-y-2">
                <Label htmlFor="login-otp">Verification code</Label>
                <Input
                  id="login-otp"
                  type="text"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  value={otp}
                  onChange={(e) => setOtp(e.target.value.replace(/\D/g, "").slice(0, 6))}
                  placeholder="000000"
                  maxLength={6}
                  required
                />
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button type="submit" className="w-full" disabled={verifying || otp.length < 6}>
                {verifying ? "Verifying…" : "Verify"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="w-full"
                disabled={verifying}
                onClick={() => {
                  setMfaChallengeId(null);
                  setMfaPhoneMask(null);
                  setMfaOtp(null);
                  setMfaOtpNote(null);
                  setPhoneIntentId(null);
                  setOtp("");
                  setError(null);
                }}
              >
                Back to sign in
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
          <CardTitle>Sign in</CardTitle>
          <CardDescription>
            {mode === "signin"
              ? "Sign in with your email and organization ID."
              : "Create a new organization. Enter your credentials and organization name."}
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
              <Label htmlFor="login-email">Email</Label>
              <Input
                id="login-email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                placeholder="you@example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="login-password">Password</Label>
              <Input
                id="login-password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Tabs
              value={mode}
              onValueChange={(v) => {
                setMode(v as "signin" | "createOrg");
                setError(null);
              }}
            >
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="signin">Existing</TabsTrigger>
                <TabsTrigger value="createOrg">Create new</TabsTrigger>
              </TabsList>
              <TabsContent value="signin" className="space-y-2">
                <Label htmlFor="login-org">Organization ID</Label>
                <Input
                  id="login-org"
                  type="text"
                  value={orgId}
                  onChange={(e) => setOrgId(e.target.value)}
                  required={mode === "signin"}
                  placeholder="your-org-id"
                />
              </TabsContent>
              <TabsContent value="createOrg" className="space-y-2">
                <Label htmlFor="login-org-name">Organization name</Label>
                <Input
                  id="login-org-name"
                  type="text"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  required={mode === "createOrg"}
                  placeholder="My Organization"
                />
              </TabsContent>
            </Tabs>
          </CardContent>
          <CardFooter className="flex flex-col gap-4">
            <Button
              type="submit"
              className="w-full"
              disabled={submitting || creatingOrg}
            >
              {creatingOrg
                ? "Creating organization…"
                : submitting
                  ? "Signing in…"
                  : mode === "createOrg"
                    ? "Create organization"
                    : "Sign in"}
            </Button>
            <p className="text-center text-sm text-muted-foreground">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="font-medium text-primary underline-offset-4 hover:underline">
                Register
              </Link>
            </p>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center">
          <p className="text-muted-foreground">Loading…</p>
        </div>
      }
    >
      <LoginPageContent />
    </Suspense>
  );
}
