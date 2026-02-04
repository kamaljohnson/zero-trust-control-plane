import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * POST /api/org-admin/members/remove
 * Body: { org_id, user_id }. Requires org admin or owner.
 */
export async function POST(request: NextRequest) {
  const token = getAccessToken(request);
  if (!token) {
    return NextResponse.json({ error: "Authorization required." }, { status: 401 });
  }
  let body: { org_id?: string; user_id?: string };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON body." }, { status: 400 });
  }
  const orgId = body.org_id ?? "";
  const userId = body.user_id ?? "";
  if (!orgId || !userId) {
    return NextResponse.json({ error: "org_id and user_id required." }, { status: 400 });
  }
  try {
    await orgAdmin.removeMember(token, orgId, userId);
    return NextResponse.json({});
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to remove member." }, { status: 500 });
  }
}
