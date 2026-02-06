import { NextRequest, NextResponse } from "next/server";
import { getDevOTP } from "@/lib/grpc/dev-client";

/**
 * GET /api/dev/mfa/otp?challenge_id=... — returns OTP and note for the given MFA challenge when dev OTP is enabled.
 * Uses server-side DEV_OTP_ENABLED (runtime) so PoC can enable in prod without rebuilding; falls back to NEXT_PUBLIC_DEV_OTP_ENABLED (build-time).
 */
export async function GET(request: NextRequest) {
  const devOtpEnabled =
    process.env.DEV_OTP_ENABLED === "true" ||
    process.env.DEV_OTP_ENABLED === "1" ||
    process.env.NEXT_PUBLIC_DEV_OTP_ENABLED === "true" ||
    process.env.NEXT_PUBLIC_DEV_OTP_ENABLED === "1";
  if (!devOtpEnabled) {
    return NextResponse.json({ error: "Not found" }, { status: 404 });
  }

  const challengeId = request.nextUrl.searchParams.get("challenge_id");
  if (!challengeId || !challengeId.trim()) {
    return NextResponse.json({ error: "challenge_id is required" }, { status: 400 });
  }

  const result = await getDevOTP(challengeId.trim());
  if (result == null) {
    return NextResponse.json({ error: "OTP not found or expired" }, { status: 404 });
  }

  return NextResponse.json({ otp: result.otp, note: result.note });
}
