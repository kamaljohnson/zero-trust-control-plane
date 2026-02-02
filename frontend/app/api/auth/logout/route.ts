import { NextRequest, NextResponse } from "next/server";
import * as auth from "@/lib/grpc/auth-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { logoutBodySchema } from "@/lib/api/auth-schemas";

/**
 * POST /api/auth/logout — revoke the session.
 * Body: { refresh_token? }. Returns 200 on success (backend may no-op if token invalid).
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json().catch(() => ({}));
    const parsed = logoutBodySchema.safeParse(raw ?? {});
    const refresh_token = parsed.success ? parsed.data.refresh_token : undefined;
    await auth.logout(refresh_token);
    return NextResponse.json({ ok: true });
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "Logout failed." },
      { status: 500 }
    );
  }
}
