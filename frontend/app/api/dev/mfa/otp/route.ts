import { NextRequest, NextResponse } from "next/server";
import { getDevOTP } from "@/lib/grpc/dev-client";

/**
 * GET /api/dev/mfa/otp?challenge_id=... — dev-only: returns OTP and note for the given MFA challenge.
 * Guards: returns 404 when NODE_ENV is production or neither DEV_OTP_ENABLED nor NEXT_PUBLIC_DEV_OTP_ENABLED is set.
 */
export async function GET(request: NextRequest) {
  if (process.env.NODE_ENV === "production") {
    return NextResponse.json({ error: "Not found" }, { status: 404 });
  }
  const devOtpEnabled =
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
