import { NextRequest, NextResponse } from "next/server";
import * as auth from "@/lib/grpc/auth-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { refreshBodySchema } from "@/lib/api/auth-schemas";

/**
 * POST /api/auth/refresh — issue new access and refresh tokens.
 * Body: { refresh_token }. Returns AuthResponse shape.
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json();
    const parsed = refreshBodySchema.safeParse(raw);
    if (!parsed.success) {
      const message =
        parsed.error.issues[0]?.message ?? "refresh_token is required.";
      return NextResponse.json({ error: message }, { status: 400 });
    }
    const { refresh_token } = parsed.data;
    const res = await auth.refresh(refresh_token);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "Refresh failed." },
      { status: 500 }
    );
  }
}
