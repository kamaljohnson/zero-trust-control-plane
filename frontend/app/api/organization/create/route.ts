import { NextRequest, NextResponse } from "next/server";
import * as organization from "@/lib/grpc/organization-client";
import { grpcErrorToHttp } from "@/lib/grpc/grpc-to-http";
import { z } from "zod";

const createOrganizationBodySchema = z.object({
  name: z.string().min(1, "Organization name is required"),
  user_id: z.string().min(1, "User ID is required"),
});

/**
 * POST /api/organization/create — create a new organization.
 *
 * Creates a new organization and assigns the user as owner. The organization is
 * auto-activated (status=ACTIVE) for PoC. This endpoint does not require authentication
 * as users need to create organizations before they can log in. Callers obtain user_id
 * via /api/auth/verify (sign-in page create-org flow) or from registration response.
 *
 * @param request - Next.js request object
 * @param request.body - Request body with:
 *   - name: string (required, min length 1) - Organization name
 *   - user_id: string (required, min length 1) - User ID from registration or /api/auth/verify
 * @returns JSON response with:
 *   - organization: { id, name, status, created_at } on success
 *   - error: string on failure
 * @throws Returns HTTP status codes:
 *   - 400: Missing or invalid name/user_id
 *   - 404: User not found
 *   - 500: Server error during creation
 */
export async function POST(request: NextRequest) {
  try {
    const raw = await request.json();
    const parsed = createOrganizationBodySchema.safeParse(raw);
    if (!parsed.success) {
      const message =
        parsed.error.issues[0]?.message ?? "Name and user_id are required.";
      return NextResponse.json({ error: message }, { status: 400 });
    }
    const { name, user_id } = parsed.data;
    const res = await organization.createOrganization(name, user_id);
    return NextResponse.json(res);
  } catch (err) {
    const e = err as { code?: number; message?: string };
    if (typeof e.code === "number" && e.message) {
      const { status, message } = grpcErrorToHttp({ code: e.code, message: e.message });
      return NextResponse.json({ error: message }, { status });
    }
    return NextResponse.json(
      { error: "Organization creation failed." },
      { status: 500 }
    );
  }
}
