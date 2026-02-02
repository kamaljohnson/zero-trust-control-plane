/**
 * gRPC client for the backend AuthService. Loads the auth proto at runtime
 * and exposes Register, Login, Refresh, Logout. Used only by Next.js API routes (BFF).
 */

import * as grpc from "@grpc/grpc-js";
import type { ServiceClientConstructor } from "@grpc/grpc-js/build/src/make-client";
import * as protoLoader from "@grpc/proto-loader";
import path from "path";

const PROTO_ROOT = path.join(process.cwd(), "lib", "grpc", "proto");
const AUTH_PROTO = path.join(PROTO_ROOT, "auth", "auth.proto");

const packageDefinition = protoLoader.loadSync(AUTH_PROTO, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
  includeDirs: [PROTO_ROOT],
});

const proto = grpc.loadPackageDefinition(packageDefinition) as unknown as {
  ztcp: { auth: { v1: { AuthService: ServiceClientConstructor } } };
};

const AuthServiceClient = proto.ztcp.auth.v1.AuthService;

export interface AuthResponseJson {
  access_token?: string;
  refresh_token?: string;
  expires_at?: string;
  user_id?: string;
  org_id?: string;
}

interface AuthResponseProto {
  access_token?: string;
  refresh_token?: string;
  expires_at?: { seconds?: string | number; nanos?: number };
  user_id?: string;
  org_id?: string;
}

function authResponseToJson(res: AuthResponseProto): AuthResponseJson {
  const out: AuthResponseJson = {};
  if (res.user_id != null) out.user_id = res.user_id;
  if (res.org_id != null) out.org_id = res.org_id;
  if (res.access_token != null) out.access_token = res.access_token;
  if (res.refresh_token != null) out.refresh_token = res.refresh_token;
  if (res.expires_at != null && typeof res.expires_at === "object") {
    const s = Number(res.expires_at.seconds ?? 0);
    const n = Number(res.expires_at.nanos ?? 0);
    if (s || n) {
      const d = new Date(s * 1000 + n / 1e6);
      out.expires_at = d.toISOString();
    }
  }
  return out;
}

function promisifyRegister(
  client: grpc.Client,
  req: { email: string; password: string; name?: string }
): Promise<AuthResponseProto> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { Register: (r: unknown, c: (e: grpc.ServiceError | null, r: AuthResponseProto) => void) => void }).Register(
      req,
      (err: grpc.ServiceError | null, res: AuthResponseProto) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
}

function promisifyLogin(
  client: grpc.Client,
  req: { email: string; password: string; org_id: string; device_fingerprint?: string }
): Promise<AuthResponseProto> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { Login: (r: unknown, c: (e: grpc.ServiceError | null, r: AuthResponseProto) => void) => void }).Login(
      req,
      (err: grpc.ServiceError | null, res: AuthResponseProto) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
}

function promisifyRefresh(
  client: grpc.Client,
  req: { refresh_token: string }
): Promise<AuthResponseProto> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { Refresh: (r: unknown, c: (e: grpc.ServiceError | null, r: AuthResponseProto) => void) => void }).Refresh(
      req,
      (err: grpc.ServiceError | null, res: AuthResponseProto) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
}

function promisifyLogout(
  client: grpc.Client,
  req: { refresh_token?: string }
): Promise<void> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { Logout: (r: unknown, c: (e: grpc.ServiceError | null) => void) => void }).Logout(
      req,
      (err: grpc.ServiceError | null) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve();
      }
    );
  });
}

/** Create AuthService client. Address should be host:port (e.g. localhost:8080). Use insecure for local dev. */
export function createAuthClient(address: string): grpc.Client {
  const creds = address.startsWith("https://")
    ? grpc.credentials.createSsl()
    : grpc.credentials.createInsecure();
  const target = address.replace(/^https?:\/\//, "");
  return new AuthServiceClient(target, creds);
}

let cachedClient: grpc.Client | null = null;

/**
 * Get or create the AuthService client. Uses BACKEND_GRPC_URL (default localhost:8080).
 */
export function getAuthClient(): grpc.Client {
  if (cachedClient) return cachedClient;
  const url = process.env.BACKEND_GRPC_URL ?? "localhost:8080";
  cachedClient = createAuthClient(url);
  return cachedClient;
}

/**
 * Register creates a user and local identity. Returns AuthResponse with user_id only.
 */
export async function register(
  email: string,
  password: string,
  name?: string
): Promise<AuthResponseJson> {
  const client = getAuthClient();
  const res = await promisifyRegister(client, { email, password, name });
  return authResponseToJson(res);
}

/**
 * Login authenticates and returns tokens and user/org context.
 */
export async function login(
  email: string,
  password: string,
  org_id: string,
  device_fingerprint?: string
): Promise<AuthResponseJson> {
  const client = getAuthClient();
  const res = await promisifyLogin(client, {
    email,
    password,
    org_id,
    device_fingerprint: device_fingerprint ?? "password-login",
  });
  return authResponseToJson(res);
}

/**
 * Refresh returns new access and refresh tokens.
 */
export async function refresh(refresh_token: string): Promise<AuthResponseJson> {
  const client = getAuthClient();
  const res = await promisifyRefresh(client, { refresh_token });
  return authResponseToJson(res);
}

/**
 * Logout revokes the session. Pass refresh_token or leave empty if revoking from context (not used from BFF).
 */
export async function logout(refresh_token?: string): Promise<void> {
  const client = getAuthClient();
  await promisifyLogout(client, { refresh_token: refresh_token ?? "" });
}
