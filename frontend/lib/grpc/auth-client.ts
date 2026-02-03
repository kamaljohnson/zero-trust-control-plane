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

/** Login response: either tokens, MFA required (challenge_id, phone_mask), or phone required (intent_id). */
export interface LoginResponseJson {
  access_token?: string;
  refresh_token?: string;
  expires_at?: string;
  user_id?: string;
  org_id?: string;
  mfa_required?: boolean;
  challenge_id?: string;
  phone_mask?: string;
  phone_required?: boolean;
  intent_id?: string;
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

interface LoginResponseProto {
  tokens?: AuthResponseProto;
  mfa_required?: { challenge_id?: string; phone_mask?: string };
  phone_required?: { intent_id?: string };
}

function promisifyLogin(
  client: grpc.Client,
  req: { email: string; password: string; org_id: string; device_fingerprint?: string }
): Promise<LoginResponseProto> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { Login: (r: unknown, c: (e: grpc.ServiceError | null, r: LoginResponseProto) => void) => void }).Login(
      req,
      (err: grpc.ServiceError | null, res: LoginResponseProto) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
}

function promisifyVerifyMFA(
  client: grpc.Client,
  req: { challenge_id: string; otp: string }
): Promise<AuthResponseProto> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { VerifyMFA: (r: unknown, c: (e: grpc.ServiceError | null, r: AuthResponseProto) => void) => void }).VerifyMFA(
      req,
      (err: grpc.ServiceError | null, res: AuthResponseProto) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
}

interface RefreshResponseProto {
  tokens?: AuthResponseProto;
  mfa_required?: { challenge_id?: string; phone_mask?: string };
  phone_required?: { intent_id?: string };
}

function promisifyRefresh(
  client: grpc.Client,
  req: { refresh_token: string; device_fingerprint?: string }
): Promise<RefreshResponseProto> {
  return new Promise((resolve, reject) => {
    (client as grpc.Client & { Refresh: (r: unknown, c: (e: grpc.ServiceError | null, r: RefreshResponseProto) => void) => void }).Refresh(
      req,
      (err: grpc.ServiceError | null, res: RefreshResponseProto) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
}

function refreshResponseToJson(res: RefreshResponseProto): LoginResponseJson {
  if (res.mfa_required != null) {
    return {
      mfa_required: true,
      challenge_id: res.mfa_required.challenge_id ?? "",
      phone_mask: res.mfa_required.phone_mask ?? "",
    };
  }
  if (res.phone_required != null) {
    return {
      phone_required: true,
      intent_id: res.phone_required.intent_id ?? "",
    };
  }
  if (res.tokens != null) {
    return authResponseToJson(res.tokens);
  }
  return {};
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
 * Login authenticates and returns either tokens or MFA required (challenge_id, phone_mask).
 */
export async function login(
  email: string,
  password: string,
  org_id: string,
  device_fingerprint?: string
): Promise<LoginResponseJson> {
  const client = getAuthClient();
  const res = await promisifyLogin(client, {
    email,
    password,
    org_id,
    device_fingerprint: device_fingerprint ?? "password-login",
  });
  if (res.mfa_required != null) {
    return {
      mfa_required: true,
      challenge_id: res.mfa_required.challenge_id ?? "",
      phone_mask: res.mfa_required.phone_mask ?? "",
    };
  }
  if (res.phone_required != null) {
    return {
      phone_required: true,
      intent_id: res.phone_required.intent_id ?? "",
    };
  }
  if (res.tokens != null) {
    return authResponseToJson(res.tokens);
  }
  return {};
}

/**
 * RequestMFAWithPhone consumes the intent, creates an MFA challenge for the submitted phone, sends OTP, and returns challenge_id and phone_mask.
 */
export async function requestMFAWithPhone(
  intent_id: string,
  phone: string
): Promise<{ challenge_id: string; phone_mask: string }> {
  const client = getAuthClient();
  const res = await new Promise<{ challenge_id?: string; phone_mask?: string }>((resolve, reject) => {
    (client as grpc.Client & {
      SubmitPhoneAndRequestMFA: (
        r: { intent_id: string; phone: string },
        c: (e: grpc.ServiceError | null, r: { challenge_id?: string; phone_mask?: string }) => void
      ) => void;
    }).SubmitPhoneAndRequestMFA(
      { intent_id, phone },
      (err: grpc.ServiceError | null, res: { challenge_id?: string; phone_mask?: string }) => {
        if (err) reject({ code: err.code, message: err.details || err.message });
        else resolve(res ?? {});
      }
    );
  });
  return {
    challenge_id: res.challenge_id ?? "",
    phone_mask: res.phone_mask ?? "",
  };
}

/**
 * VerifyMFA verifies the OTP for the given challenge and returns tokens.
 */
export async function verifyMFA(challenge_id: string, otp: string): Promise<AuthResponseJson> {
  const client = getAuthClient();
  const res = await promisifyVerifyMFA(client, { challenge_id, otp });
  return authResponseToJson(res);
}

/**
 * Refresh returns new access and refresh tokens, or MFA required / phone required when device-trust policy requires it (same shape as LoginResponseJson).
 */
export async function refresh(
  refresh_token: string,
  device_fingerprint?: string
): Promise<LoginResponseJson> {
  const client = getAuthClient();
  const res = await promisifyRefresh(client, {
    refresh_token,
    device_fingerprint: device_fingerprint ?? "password-login",
  });
  return refreshResponseToJson(res);
}

/**
 * Logout revokes the session. Pass refresh_token or leave empty if revoking from context (not used from BFF).
 */
export async function logout(refresh_token?: string): Promise<void> {
  const client = getAuthClient();
  await promisifyLogout(client, { refresh_token: refresh_token ?? "" });
}
