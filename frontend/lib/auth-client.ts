/**
 * Client for the auth BFF (Next.js API routes). Used by the browser to call
 * /api/auth/register, login, refresh, logout. Tokens are stored by the auth context.
 */

const API = "/api/auth";

export interface AuthResponse {
  access_token?: string;
  refresh_token?: string;
  expires_at?: string;
  user_id?: string;
  org_id?: string;
}

export interface RegisterResponse {
  user_id?: string;
}

/**
 * Register creates a user and local identity. Returns user_id only.
 */
export async function register(
  email: string,
  password: string,
  name?: string
): Promise<RegisterResponse> {
  const res = await fetch(`${API}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, name }),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error ?? "Registration failed.");
  }
  return data;
}

/**
 * Login authenticates and returns tokens and user/org context.
 */
export async function login(
  email: string,
  password: string,
  orgId: string,
  deviceFingerprint?: string
): Promise<AuthResponse> {
  const res = await fetch(`${API}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      email,
      password,
      org_id: orgId,
      device_fingerprint: deviceFingerprint,
    }),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error ?? "Login failed.");
  }
  return data;
}

/**
 * Refresh returns new access and refresh tokens.
 */
export async function refresh(refreshToken: string): Promise<AuthResponse> {
  const res = await fetch(`${API}/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error ?? "Refresh failed.");
  }
  return data;
}

/**
 * Logout revokes the session.
 */
export async function logout(refreshToken?: string): Promise<void> {
  const res = await fetch(`${API}/logout`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken ?? "" }),
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error ?? "Logout failed.");
  }
}
