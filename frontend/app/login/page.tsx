"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useState, useEffect, Suspense } from "react";
import Link from "next/link";
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

const EMAIL_REGEX = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
const DEFAULT_ORG_ID = process.env.NEXT_PUBLIC_DEFAULT_ORG_ID ?? "";
const DEV_OTP_ENABLED = process.env.NEXT_PUBLIC_DEV_OTP_ENABLED === "true" || process.env.NEXT_PUBLIC_DEV_OTP_ENABLED === "1";

function LoginPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, verifyMFA, isAuthenticated, isLoading } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  // Initialize with DEFAULT_ORG_ID to avoid hydration mismatch; sync from searchParams in useEffect
  const [orgId, setOrgId] = useState(DEFAULT_ORG_ID);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [phoneIntentId, setPhoneIntentId] = useState<string | null>(null);
  const [phone, setPhone] = useState("");
  const [requestingMfa, setRequestingMfa] = useState(false);
  const [mfaChallengeId, setMfaChallengeId] = useState<string | null>(null);
  const [mfaPhoneMask, setMfaPhoneMask] = useState<string | null>(null);
  const [mfaOtp, setMfaOtp] = useState<string | null>(null);
  const [mfaOtpNote, setMfaOtpNote] = useState<string | null>(null);
  const [otp, setOtp] = useState("");
  const [verifying, setVerifying] = useState(false);

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace("/");
    }
  }, [isLoading, isAuthenticated, router]);

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
    if (!mfaChallengeId || !DEV_OTP_ENABLED) return;
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

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading…</p>
      </div>
    );
  }

  if (isAuthenticated) {
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
            Sign in with your email and organization ID.
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
            <div className="space-y-2">
              <Label htmlFor="login-org">Organization ID</Label>
              <Input
                id="login-org"
                type="text"
                value={orgId}
                onChange={(e) => setOrgId(e.target.value)}
                required
                placeholder="your-org-id"
              />
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-4">
            <Button type="submit" className="w-full" disabled={submitting}>
              {submitting ? "Signing in…" : "Sign in"}
            </Button>
            <p className="text-center text-sm text-muted-foreground">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="text-primary underline-offset-4 hover:underline">
                Register
              </Link>
            </p>
            <p className="text-center text-sm text-muted-foreground">
              Need to create an organization?{" "}
              <Link href="/register" className="text-primary underline-offset-4 hover:underline">
                Register and create one
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
