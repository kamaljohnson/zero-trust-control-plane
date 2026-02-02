/**
 * gRPC client for the backend DevService (dev-only, e.g. GetOTP). Used only by Next.js API routes (BFF).
 */

import * as grpc from "@grpc/grpc-js";
import type { ServiceClientConstructor } from "@grpc/grpc-js/build/src/make-client";
import * as protoLoader from "@grpc/proto-loader";
import path from "path";

const PROTO_ROOT = path.join(process.cwd(), "lib", "grpc", "proto");
const DEV_PROTO = path.join(PROTO_ROOT, "dev", "dev.proto");

const packageDefinition = protoLoader.loadSync(DEV_PROTO, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
  includeDirs: [PROTO_ROOT],
});

const proto = grpc.loadPackageDefinition(packageDefinition) as unknown as {
  ztcp: { dev: { v1: { DevService: ServiceClientConstructor } } };
};

const DevServiceClient = proto.ztcp.dev.v1.DevService;

function getDevClient(): grpc.Client {
  const url = process.env.BACKEND_GRPC_URL ?? "localhost:8080";
  const creds = url.startsWith("https://")
    ? grpc.credentials.createSsl()
    : grpc.credentials.createInsecure();
  const target = url.replace(/^https?:\/\//, "");
  return new DevServiceClient(target, creds);
}

/**
 * GetOTP returns the plain OTP for the given challenge_id from the backend dev store. Returns null if not found or expired (e.g. 404).
 */
export async function getDevOTP(challengeId: string): Promise<{ otp: string; note: string } | null> {
  const client = getDevClient();
  return new Promise((resolve) => {
    (client as grpc.Client & { GetOTP: (r: { challenge_id: string }, c: (e: grpc.ServiceError | null, r: { otp?: string; note?: string }) => void) => void }).GetOTP(
      { challenge_id: challengeId },
      (err: grpc.ServiceError | null, res: { otp?: string; note?: string }) => {
        if (err) {
          resolve(null);
          return;
        }
        if (res?.otp != null) {
          resolve({ otp: res.otp, note: res.note ?? "DEV MODE ONLY" });
        } else {
          resolve(null);
        }
      }
    );
  });
}
