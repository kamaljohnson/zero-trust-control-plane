import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * GET /api/org-admin/policy-config?org_id=...
 * Returns org policy config for the org. Requires org admin or owner.
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
  try {
    const res = await orgAdmin.getOrgPolicyConfig(token, orgId);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to get policy config." }, { status: 500 });
  }
}

/**
 * PUT /api/org-admin/policy-config
 * Body: { org_id: string, config: OrgPolicyConfig }
 * Updates org policy config. Requires org admin or owner.
 */
export async function PUT(request: NextRequest) {
  const token = getAccessToken(request);
  if (!token) {
    return NextResponse.json({ error: "Authorization required." }, { status: 401 });
  }
  let body: { org_id?: string; config?: unknown };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON body." }, { status: 400 });
  }
  const orgId = body?.org_id ?? "";
  if (!orgId) {
    return NextResponse.json({ error: "org_id required." }, { status: 400 });
  }
  const config = body?.config ?? {};
  try {
    const res = await orgAdmin.updateOrgPolicyConfig(token, orgId, config);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to update policy config." }, { status: 500 });
  }
}
