import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * GET /api/org-admin/sessions?org_id=...&user_id=...&page_size=50&page_token=...
 * Returns list of sessions. user_id optional to filter by member. Requires org admin or owner.
 */
export async function GET(request: NextRequest) {
  const token = getAccessToken(request);
  if (!token) {
    return NextResponse.json({ error: "Authorization required." }, { status: 401 });
  }
  const orgId = request.nextUrl.searchParams.get("org_id") ?? "";
  if (!orgId) {
    return NextResponse.json({ error: "org_id required." }, { status: 400 });
  }
  const userId = request.nextUrl.searchParams.get("user_id") ?? undefined;
  const pageSize = request.nextUrl.searchParams.get("page_size");
  const pageToken = request.nextUrl.searchParams.get("page_token") ?? undefined;
  try {
    const res = await orgAdmin.listSessions(
      token,
      orgId,
      userId,
      pageSize ? parseInt(pageSize, 10) : undefined,
      pageToken
    );
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to list sessions." }, { status: 500 });
  }
}
