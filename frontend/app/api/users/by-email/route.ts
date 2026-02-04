import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * GET /api/users/by-email?email=...
 * Returns user by email (for invite flow). Requires authentication.
 */
export async function GET(request: NextRequest) {
  const token = getAccessToken(request);
  if (!token) {
    return NextResponse.json({ error: "Authorization required." }, { status: 401 });
  }
  const email = request.nextUrl.searchParams.get("email") ?? "";
  if (!email.trim()) {
    return NextResponse.json({ error: "email required." }, { status: 400 });
  }
  try {
    const res = await orgAdmin.getUserByEmail(token, email.trim());
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to look up user." }, { status: 500 });
  }
}
