import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * POST /api/browser/check-url
 * Body: { url: string, org_id: string }
 * Returns { allowed: boolean, reason?: string } for the URL under org access control policy.
 * Callable by any authenticated org member.
 */
export async function POST(request: NextRequest) {
  const token = getAccessToken(request);
  if (!token) {
    return NextResponse.json({ error: "Authorization required." }, { status: 401 });
  }
  let body: { url?: string; org_id?: string };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON body." }, { status: 400 });
  }
  const orgId = body?.org_id ?? "";
  const url = typeof body?.url === "string" ? body.url.trim() : "";
  if (!orgId) {
    return NextResponse.json({ error: "org_id required." }, { status: 400 });
  }
  if (!url) {
    return NextResponse.json({ error: "url required." }, { status: 400 });
  }
  try {
    const res = await orgAdmin.checkUrlAccess(token, orgId, url);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Could not check access. Please try again." }, { status: 500 });
  }
}
