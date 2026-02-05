import { NextRequest, NextResponse } from "next/server";
import * as orgAdmin from "@/lib/grpc/org-admin-clients";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { getAccessToken } from "@/lib/api/get-access-token";

/**
 * GET /api/browser/session?org_id=...&user_id=...
 * Returns the current user's active session. Callable by any authenticated org member.
 * Finds the active session (not revoked, not expired) from the user's sessions.
 */
export async function GET(request: NextRequest) {
  const token = getAccessToken(request);
  if (!token) {
    return NextResponse.json({ error: "Authorization required." }, { status: 401 });
  }
  const orgId = request.nextUrl.searchParams.get("org_id") ?? "";
  const userId = request.nextUrl.searchParams.get("user_id") ?? "";
  if (!orgId) {
    return NextResponse.json({ error: "org_id required." }, { status: 400 });
  }
  if (!userId) {
    return NextResponse.json({ error: "user_id required." }, { status: 400 });
  }
  type SessionRow = {
    id: string;
    user_id: string;
    org_id: string;
    device_id?: string;
    expires_at?: { seconds?: number; nanos?: number };
    revoked_at?: { seconds?: number; nanos?: number };
    last_seen_at?: { seconds?: number; nanos?: number };
    ip_address?: string;
    created_at?: { seconds?: number; nanos?: number };
  };
  try {
    // List sessions for this user (listSessions returns unknown from gRPC)
    const res = (await orgAdmin.listSessions(token, orgId, userId, 50)) as {
      sessions?: SessionRow[];
    };
    const sessions: SessionRow[] = res.sessions ?? [];

    // Find active session (not revoked, not expired)
    const now = Date.now() / 1000; // Convert to seconds
    const activeSession = sessions.find((s) => {
      // Skip if revoked
      if (s.revoked_at?.seconds) {
        return false;
      }
      // Skip if expired
      if (s.expires_at?.seconds && s.expires_at.seconds < now) {
        return false;
      }
      return true;
    });

    if (!activeSession) {
      return NextResponse.json({ error: "No active session found." }, { status: 404 });
    }

    return NextResponse.json({ session: activeSession });
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json({ error: "Failed to get session." }, { status: 500 });
  }
}
