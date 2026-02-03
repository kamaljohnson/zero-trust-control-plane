import { z } from "zod";

/**
 * Request body schemas for auth API routes. Used to validate and parse
 * incoming JSON before calling the gRPC auth client.
 */

export const loginBodySchema = z.object({
  email: z.string().min(1, "Email is required."),
  password: z.string().min(1, "Password is required."),
  org_id: z.string().min(1, "Organization ID is required."),
  device_fingerprint: z.string().optional(),
});

export const registerBodySchema = z.object({
  email: z.string().min(1, "Email is required."),
  password: z.string().min(1, "Password is required."),
  name: z.string().optional(),
});

export const refreshBodySchema = z.object({
  refresh_token: z.string().min(1, "refresh_token is required."),
  device_fingerprint: z.string().optional(),
});

export const logoutBodySchema = z.object({
  refresh_token: z.string().optional(),
});

export type LoginBody = z.infer<typeof loginBodySchema>;
export type RegisterBody = z.infer<typeof registerBodySchema>;
export type RefreshBody = z.infer<typeof refreshBodySchema>;
export type LogoutBody = z.infer<typeof logoutBodySchema>;
