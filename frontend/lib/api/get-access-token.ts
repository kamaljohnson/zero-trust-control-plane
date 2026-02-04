import { NextRequest } from "next/server";

/**
 * Returns the access token from the request Authorization header (Bearer <token>).
 * Returns null if missing or invalid.
 */
export function getAccessToken(request: NextRequest): string | null {
  const auth = request.headers.get("authorization");
  if (!auth || !auth.startsWith("Bearer ")) return null;
  const token = auth.slice(7).trim();
  return token || null;
}
