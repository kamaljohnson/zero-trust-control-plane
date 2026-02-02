import { NextRequest, NextResponse } from "next/server";
import * as auth from "@/lib/grpc/auth-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";

/**
 * POST /api/auth/mfa/request-with-phone — submit phone for MFA (when Login returned phone_required). Returns challenge_id and phone_mask.
 * Body: { intent_id, phone }.
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json();
    const intent_id = typeof raw?.intent_id === "string" ? raw.intent_id.trim() : "";
    const phone = typeof raw?.phone === "string" ? raw.phone.trim() : "";
    if (!intent_id || !phone) {
      return NextResponse.json(
        { error: "intent_id and phone are required." },
        { status: 400 }
      );
    }
    const res = await auth.requestMFAWithPhone(intent_id, phone);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "Request MFA with phone failed." },
      { status: 500 }
    );
  }
}
