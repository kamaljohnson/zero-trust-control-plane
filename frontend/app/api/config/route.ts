import { NextResponse } from "next/server";

/**
 * GET /api/config — returns public runtime config (e.g. dev OTP enabled).
 * Used so PoC can enable OTP return to client via env without rebuilding the frontend.
 */
export async function GET() {
  const devOtpEnabled =
    process.env.DEV_OTP_ENABLED === "true" || process.env.DEV_OTP_ENABLED === "1";
  return NextResponse.json({ devOtpEnabled });
}
