import { NextRequest, NextResponse } from "next/server";
import * as auth from "@/lib/grpc/auth-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { loginBodySchema } from "@/lib/api/auth-schemas";

/**
 * POST /api/auth/login — authenticate and return tokens.
 * Body: { email, password, org_id, device_fingerprint? }. Returns AuthResponse shape.
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json();
    const parsed = loginBodySchema.safeParse(raw);
    if (!parsed.success) {
      const message =
        parsed.error.issues[0]?.message ?? "Email, password, and org_id are required.";
      return NextResponse.json({ error: message }, { status: 400 });
    }
    const { email, password, org_id, device_fingerprint } = parsed.data;
    const res = await auth.login(email, password, org_id, device_fingerprint);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "Login failed." },
      { status: 500 }
    );
  }
}
