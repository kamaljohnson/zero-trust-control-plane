import { NextRequest, NextResponse } from "next/server";
import * as auth from "@/lib/grpc/auth-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { verifyBodySchema } from "@/lib/api/auth-schemas";

/**
 * POST /api/auth/verify — verify email/password and return user_id (no session).
 * Used by the org-creation flow so registered users can create an organization from the sign-in page.
 * Body: { email, password }. Returns { user_id }.
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json();
    const parsed = verifyBodySchema.safeParse(raw);
    if (!parsed.success) {
      const message = parsed.error.issues[0]?.message ?? "Email and password are required.";
      return NextResponse.json({ error: message }, { status: 400 });
    }
    const { email, password } = parsed.data;
    const res = await auth.verifyCredentials(email, password);
    if (!res.user_id) {
      return NextResponse.json({ error: "user_id not returned" }, { status: 500 });
    }
    return NextResponse.json({ user_id: res.user_id });
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "Verification failed." },
      { status: 500 }
    );
  }
}
