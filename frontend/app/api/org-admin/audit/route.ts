import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * GET /api/org-admin/audit?org_id=...&page_size=50&page_token=...&user_id=...&action=...&resource=...
 * Returns audit logs for the org. Requires org admin or owner.
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
  const pageSize = request.nextUrl.searchParams.get("page_size");
  const pageToken = request.nextUrl.searchParams.get("page_token") ?? undefined;
  const userId = request.nextUrl.searchParams.get("user_id") ?? undefined;
  const action = request.nextUrl.searchParams.get("action") ?? undefined;
  const resource = request.nextUrl.searchParams.get("resource") ?? undefined;
  try {
    const res = await orgAdmin.listAuditLogs(
      token,
      orgId,
      pageSize ? parseInt(pageSize, 10) : undefined,
      pageToken,
      userId,
      action,
      resource
    );
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to list audit logs." }, { status: 500 });
  }
}
