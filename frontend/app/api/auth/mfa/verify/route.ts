import { NextRequest, NextResponse } from "next/server";
import * as auth from "@/lib/grpc/auth-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";

/**
 * POST /api/auth/mfa/verify — verify MFA OTP and return tokens.
 * Body: { challenge_id, otp }. Returns AuthResponse shape.
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json();
    const challenge_id = typeof raw?.challenge_id === "string" ? raw.challenge_id.trim() : "";
    const otp = typeof raw?.otp === "string" ? raw.otp.trim() : "";
    if (!challenge_id || !otp) {
      return NextResponse.json(
        { error: "challenge_id and otp are required." },
        { status: 400 }
      );
    }
    const res = await auth.verifyMFA(challenge_id, otp);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "MFA verification failed." },
      { status: 500 }
    );
  }
}
