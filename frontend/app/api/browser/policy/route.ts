import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * GET /api/browser/policy?org_id=...
 * Returns browser-relevant policy (access_control, action_restrictions) for the caller's org.
 * Callable by any authenticated org member.
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
    const res = await orgAdmin.getBrowserPolicy(token, orgId);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to get browser policy." }, { status: 500 });
  }
}
