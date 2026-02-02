import { NextRequest, NextResponse } from "next/server";
import * as auth from "@/lib/grpc/auth-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { registerBodySchema } from "@/lib/api/auth-schemas";

/**
 * POST /api/auth/register — create user and local identity.
 * Body: { email, password, name? }. Returns { user_id }.
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json();
    const parsed = registerBodySchema.safeParse(raw);
    if (!parsed.success) {
      const message =
        parsed.error.issues[0]?.message ?? "Email and password are required.";
      return NextResponse.json({ error: message }, { status: 400 });
    }
    const { email, password, name } = parsed.data;
    const res = await auth.register(email, password, name);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "Registration failed." },
      { status: 500 }
    );
  }
}
