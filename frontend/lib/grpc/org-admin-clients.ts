/**
 * gRPC clients for org-admin: membership, session, audit, user.
 * All calls require Bearer token (access_token). Used by Next.js API routes.
 */

import * as grpc from "@grpc/grpc-js";
import type { ServiceClientConstructor } from "@grpc/grpc-js/build/src/make-client";
import * as protoLoader from "@grpc/proto-loader";
import path from "path";

const PROTO_ROOT = path.join(process.cwd(), "lib", "grpc", "proto");

const loadOptions = {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
  includeDirs: [PROTO_ROOT],
};

function getChannel(): string {
  const url = process.env.BACKEND_GRPC_URL ?? "localhost:8080";
  return url.startsWith("https://") ? url.replace(/^https?:\/\//, "") : url;
}

function getCredentials(): grpc.ChannelCredentials {
  const url = process.env.BACKEND_GRPC_URL ?? "localhost:8080";
  return url.startsWith("https://") ? grpc.credentials.createSsl() : grpc.credentials.createInsecure();
}

export function metadataWithAuth(accessToken: string): grpc.Metadata {
  const md = new grpc.Metadata();
  md.add("authorization", `Bearer ${accessToken.trim()}`);
  return md;
}

// Membership
const membershipDef = protoLoader.loadSync(path.join(PROTO_ROOT, "membership", "membership.proto"), loadOptions);
const membershipPkg = grpc.loadPackageDefinition(membershipDef) as unknown as {
  ztcp: { membership: { v1: { MembershipService: ServiceClientConstructor } } };
};
const MembershipServiceClient = membershipPkg.ztcp.membership.v1.MembershipService;

let membershipClient: grpc.Client | null = null;
export function getMembershipClient(): grpc.Client {
  if (!membershipClient) {
    membershipClient = new MembershipServiceClient(getChannel(), getCredentials());
  }
  return membershipClient;
}

// Session
const sessionDef = protoLoader.loadSync(path.join(PROTO_ROOT, "session", "session.proto"), loadOptions);
const sessionPkg = grpc.loadPackageDefinition(sessionDef) as unknown as {
  ztcp: { session: { v1: { SessionService: ServiceClientConstructor } } };
};
const SessionServiceClient = sessionPkg.ztcp.session.v1.SessionService;

let sessionClient: grpc.Client | null = null;
export function getSessionClient(): grpc.Client {
  if (!sessionClient) {
    sessionClient = new SessionServiceClient(getChannel(), getCredentials());
  }
  return sessionClient;
}

// Audit
const auditDef = protoLoader.loadSync(path.join(PROTO_ROOT, "audit", "audit.proto"), loadOptions);
const auditPkg = grpc.loadPackageDefinition(auditDef) as unknown as {
  ztcp: { audit: { v1: { AuditService: ServiceClientConstructor } } };
};
const AuditServiceClient = auditPkg.ztcp.audit.v1.AuditService;

let auditClient: grpc.Client | null = null;
export function getAuditClient(): grpc.Client {
  if (!auditClient) {
    auditClient = new AuditServiceClient(getChannel(), getCredentials());
  }
  return auditClient;
}

// User
const userDef = protoLoader.loadSync(path.join(PROTO_ROOT, "user", "user.proto"), loadOptions);
const userPkg = grpc.loadPackageDefinition(userDef) as unknown as {
  ztcp: { user: { v1: { UserService: ServiceClientConstructor } } };
};
const UserServiceClient = userPkg.ztcp.user.v1.UserService;

let userClient: grpc.Client | null = null;
export function getUserClient(): grpc.Client {
  if (!userClient) {
    userClient = new UserServiceClient(getChannel(), getCredentials());
  }
  return userClient;
}

function promisifyWithMeta<TReq, TRes>(
  client: grpc.Client,
  method: string,
  req: TReq,
  metadata: grpc.Metadata
): Promise<TRes> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { [k: string]: (r: TReq, m: grpc.Metadata, c: (e: grpc.ServiceError | null, r: TRes) => void) => void })[method](
      req,
      metadata,
      (err: grpc.ServiceError | null, res: TRes) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? ({} as TRes));
      }
    );
  });
}

// Membership RPCs
export async function listMembers(accessToken: string, orgId: string, pageSize?: number, pageToken?: string) {
  return promisifyWithMeta(
    getMembershipClient(),
    "ListMembers",
    {
      org_id: orgId,
      pagination: { page_size: pageSize ?? 50, page_token: pageToken ?? "" },
    },
    metadataWithAuth(accessToken)
  );
}

export async function addMember(accessToken: string, orgId: string, userId: string, role: number) {
  return promisifyWithMeta(
    getMembershipClient(),
    "AddMember",
    { org_id: orgId, user_id: userId, role },
    metadataWithAuth(accessToken)
  );
}

export async function removeMember(accessToken: string, orgId: string, userId: string) {
  return promisifyWithMeta(
    getMembershipClient(),
    "RemoveMember",
    { org_id: orgId, user_id: userId },
    metadataWithAuth(accessToken)
  );
}

export async function updateRole(accessToken: string, orgId: string, userId: string, role: number) {
  return promisifyWithMeta(
    getMembershipClient(),
    "UpdateRole",
    { org_id: orgId, user_id: userId, role },
    metadataWithAuth(accessToken)
  );
}

// Session RPCs
export async function listSessions(
  accessToken: string,
  orgId: string,
  userId?: string,
  pageSize?: number,
  pageToken?: string
) {
  return promisifyWithMeta(
    getSessionClient(),
    "ListSessions",
    {
      org_id: orgId,
      user_id: userId ?? "",
      pagination: { page_size: pageSize ?? 50, page_token: pageToken ?? "" },
    },
    metadataWithAuth(accessToken)
  );
}

export async function revokeSession(accessToken: string, sessionId: string) {
  return promisifyWithMeta(
    getSessionClient(),
    "RevokeSession",
    { session_id: sessionId },
    metadataWithAuth(accessToken)
  );
}

export async function revokeAllSessionsForUser(accessToken: string, orgId: string, userId: string) {
  return promisifyWithMeta(
    getSessionClient(),
    "RevokeAllSessionsForUser",
    { org_id: orgId, user_id: userId },
    metadataWithAuth(accessToken)
  );
}

// Audit RPCs
export async function listAuditLogs(
  accessToken: string,
  orgId: string,
  pageSize?: number,
  pageToken?: string,
  userId?: string,
  action?: string,
  resource?: string
) {
  return promisifyWithMeta(
    getAuditClient(),
    "ListAuditLogs",
    {
      org_id: orgId,
      pagination: { page_size: pageSize ?? 50, page_token: pageToken ?? "" },
      user_id: userId ?? "",
      action: action ?? "",
      resource: resource ?? "",
    },
    metadataWithAuth(accessToken)
  );
}

// User RPCs
export async function getUserByEmail(accessToken: string, email: string) {
  return promisifyWithMeta(getUserClient(), "GetUserByEmail", { email }, metadataWithAuth(accessToken));
}

// Re-export for API routes; client components should use @/lib/api/membership-roles
export { MembershipRole } from "@/lib/api/membership-roles";
