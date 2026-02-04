/**
 * gRPC client for the backend OrganizationService. Loads the organization proto at runtime
 * and exposes CreateOrganization. Used only by Next.js API routes (BFF).
 */

import * as grpc from "@grpc/grpc-js";
import type { ServiceClientConstructor } from "@grpc/grpc-js/build/src/make-client";
import * as protoLoader from "@grpc/proto-loader";
import path from "path";

const PROTO_ROOT = path.join(process.cwd(), "lib", "grpc", "proto");
const ORGANIZATION_PROTO = path.join(PROTO_ROOT, "organization", "organization.proto");

const packageDefinition = protoLoader.loadSync(ORGANIZATION_PROTO, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
  includeDirs: [PROTO_ROOT],
});

const proto = grpc.loadPackageDefinition(packageDefinition) as unknown as {
  ztcp: { organization: { v1: { OrganizationService: ServiceClientConstructor } } };
};

const OrganizationServiceClient = proto.ztcp.organization.v1.OrganizationService;

export interface OrganizationJson {
  id?: string;
  name?: string;
  status?: string;
  created_at?: string;
}

export interface CreateOrganizationResponseJson {
  organization?: OrganizationJson;
}

interface CreateOrganizationResponseProto {
  organization?: {
    id?: string;
    name?: string;
    status?: number;
    created_at?: { seconds?: string | number; nanos?: number };
  };
}

function organizationResponseToJson(res: CreateOrganizationResponseProto): CreateOrganizationResponseJson {
  const out: CreateOrganizationResponseJson = {};
  if (res.organization) {
    const org: OrganizationJson = {};
    if (res.organization.id != null) org.id = res.organization.id;
    if (res.organization.name != null) org.name = res.organization.name;
    if (res.organization.status != null) {
      // Map status enum to string
      const statusMap: Record<number, string> = {
        0: "UNSPECIFIED",
        1: "ACTIVE",
        2: "SUSPENDED",
      };
      org.status = statusMap[res.organization.status] || "UNSPECIFIED";
    }
    if (res.organization.created_at != null && typeof res.organization.created_at === "object") {
      const s = Number(res.organization.created_at.seconds ?? 0);
      const n = Number(res.organization.created_at.nanos ?? 0);
      if (s || n) {
        const d = new Date(s * 1000 + n / 1e6);
        org.created_at = d.toISOString();
      }
    }
    out.organization = org;
  }
  return out;
}

function promisifyCreateOrganization(
  client: grpc.Client,
  req: { name: string; user_id: string }
): Promise<CreateOrganizationResponseProto> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & {
      CreateOrganization: (
        r: unknown,
        c: (e: grpc.ServiceError | null, r: CreateOrganizationResponseProto) => void
      ) => void;
    }).CreateOrganization(
      req,
      (err: grpc.ServiceError | null, res: CreateOrganizationResponseProto) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
}

/** Create OrganizationService client. Address should be host:port (e.g. localhost:8080). Use insecure for local dev. */
export function createOrganizationClient(address: string): grpc.Client {
  const creds = address.startsWith("https://")
    ? grpc.credentials.createSsl()
    : grpc.credentials.createInsecure();
  const target = address.replace(/^https?:\/\//, "");
  return new OrganizationServiceClient(target, creds);
}

let cachedClient: grpc.Client | null = null;

/**
 * Get or create the OrganizationService client. Uses BACKEND_GRPC_URL (default localhost:8080).
 */
export function getOrganizationClient(): grpc.Client {
  if (cachedClient) return cachedClient;
  const url = process.env.BACKEND_GRPC_URL ?? "localhost:8080";
  cachedClient = createOrganizationClient(url);
  return cachedClient;
}

/**
 * CreateOrganization creates a new organization with the given name and user_id.
 *
 * @param name - Organization name (required, non-empty)
 * @param user_id - User ID from registration (required, non-empty)
 * @returns The created organization with id, name, status (ACTIVE), and created_at
 * @throws {Error} When the request fails (e.g., user not found, validation error, server error)
 *   Error object has `code` (gRPC status code) and `message` properties
 */
export async function createOrganization(
  name: string,
  user_id: string
): Promise<CreateOrganizationResponseJson> {
  const client = getOrganizationClient();
  const res = await promisifyCreateOrganization(client, { name, user_id });
  return organizationResponseToJson(res);
}
