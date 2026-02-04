/**
 * Maps gRPC status codes from the auth client to HTTP status and user-facing messages.
 */

import { status as GrpcStatus } from "@grpc/grpc-js";

export interface GrpcError {
  code: number;
  message: string;
}

/**
 * Map gRPC error to HTTP status and JSON body message.
 * Returns { status, message } for use in NextResponse.json(..., { status }).
 */
export function grpcErrorToHttp(err: GrpcError): { status: number; message: string } {
  switch (err.code) {
    case GrpcStatus.ALREADY_EXISTS:
      return { status: 409, message: err.message || "Email already registered." };
    case GrpcStatus.UNAUTHENTICATED:
      return { status: 401, message: err.message?.includes("reuse") ? "Session expired or invalid." : err.message || "Invalid credentials." };
    case GrpcStatus.PERMISSION_DENIED:
      return { status: 403, message: err.message || "Permission denied." };
    case GrpcStatus.NOT_FOUND:
      return { status: 404, message: err.message || "Not found." };
    case GrpcStatus.INVALID_ARGUMENT:
      return { status: 400, message: err.message || "Invalid input." };
    case GrpcStatus.FAILED_PRECONDITION:
      return { status: 400, message: err.message || "Precondition failed." };
    case GrpcStatus.UNIMPLEMENTED:
      return { status: 501, message: err.message || "Not configured." };
    default:
      return { status: 500, message: err.message || "Internal error." };
  }
}
