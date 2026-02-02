"use client";

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import * as authClient from "@/lib/auth-client";

const STORAGE_KEYS = {
  access_token: "ztcp_access_token",
  refresh_token: "ztcp_refresh_token",
  expires_at: "ztcp_expires_at",
  user_id: "ztcp_user_id",
  org_id: "ztcp_org_id",
} as const;

export interface AuthUser {
  user_id: string;
  org_id: string;
}

interface AuthState {
  user: AuthUser;
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
}

interface AuthContextValue {
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  /** Returns login response; if mfa_required, caller should show OTP step and call verifyMFA. */
  login: (
    email: string,
    password: string,
    orgId: string
  ) => Promise<authClient.LoginResponse>;
  verifyMFA: (challengeId: string, otp: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  setAuthFromResponse: (res: authClient.AuthResponse) => void;
  clearAuth: () => void;
}

function loadFromStorage(): AuthState | null {
  if (typeof window === "undefined") return null;
  const accessToken = localStorage.getItem(STORAGE_KEYS.access_token);
  const refreshToken = localStorage.getItem(STORAGE_KEYS.refresh_token);
  const user_id = localStorage.getItem(STORAGE_KEYS.user_id);
  const org_id = localStorage.getItem(STORAGE_KEYS.org_id);
  const expiresAt = localStorage.getItem(STORAGE_KEYS.expires_at) ?? "";
  if (!accessToken || !refreshToken || !user_id || !org_id) return null;
  return {
    user: { user_id, org_id },
    accessToken,
    refreshToken,
    expiresAt,
  };
}

function saveToStorage(state: AuthState): void {
  if (typeof window === "undefined") return;
  localStorage.setItem(STORAGE_KEYS.access_token, state.accessToken);
  localStorage.setItem(STORAGE_KEYS.refresh_token, state.refreshToken);
  localStorage.setItem(STORAGE_KEYS.expires_at, state.expiresAt);
  localStorage.setItem(STORAGE_KEYS.user_id, state.user.user_id);
  localStorage.setItem(STORAGE_KEYS.org_id, state.user.org_id);
}

function clearStorage(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem(STORAGE_KEYS.access_token);
  localStorage.removeItem(STORAGE_KEYS.refresh_token);
  localStorage.removeItem(STORAGE_KEYS.expires_at);
  localStorage.removeItem(STORAGE_KEYS.user_id);
  localStorage.removeItem(STORAGE_KEYS.org_id);
}

/** Refresh access token this many ms before expires_at. */
const REFRESH_MARGIN_MS = 5 * 60 * 1000; // 5 minutes

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const stored = loadFromStorage();
    setState(stored);
    setIsLoading(false);
  }, []);

  const refresh = useCallback(async () => {
    const refreshToken = state?.refreshToken;
    if (!refreshToken) return;
    try {
      const res = await authClient.refresh(refreshToken);
      if (res.access_token && res.refresh_token && res.user_id && res.org_id) {
        const next: AuthState = {
          user: { user_id: res.user_id, org_id: res.org_id },
          accessToken: res.access_token,
          refreshToken: res.refresh_token,
          expiresAt: res.expires_at ?? "",
        };
        saveToStorage(next);
        setState(next);
      }
    } catch {
      clearStorage();
      setState(null);
    }
  }, [state?.refreshToken]);

  useEffect(() => {
    if (!state?.expiresAt || !state?.refreshToken) return;
    const expiresAtMs = new Date(state.expiresAt).getTime();
    if (Number.isNaN(expiresAtMs)) return;
    const refreshAt = expiresAtMs - REFRESH_MARGIN_MS;
    const now = Date.now();
    if (refreshAt <= now) {
      refresh();
      return;
    }
    const delay = refreshAt - now;
    const timeoutId = setTimeout(() => refresh(), delay);
    return () => clearTimeout(timeoutId);
  }, [state?.expiresAt, state?.refreshToken, refresh]);

  const setAuthFromResponse = useCallback((res: authClient.AuthResponse) => {
    if (!res.access_token || !res.refresh_token || !res.user_id || !res.org_id) {
      return;
    }
    const next: AuthState = {
      user: { user_id: res.user_id, org_id: res.org_id },
      accessToken: res.access_token,
      refreshToken: res.refresh_token,
      expiresAt: res.expires_at ?? "",
    };
    saveToStorage(next);
    setState(next);
  }, []);

  const clearAuth = useCallback(() => {
    clearStorage();
    setState(null);
  }, []);

  const login = useCallback(
    async (email: string, password: string, orgId: string): Promise<authClient.LoginResponse> => {
      const res = await authClient.login(email, password, orgId);
      if (res.mfa_required !== true && res.phone_required !== true) {
        setAuthFromResponse(res as authClient.AuthResponse);
      }
      return res;
    },
    [setAuthFromResponse]
  );

  const verifyMFA = useCallback(
    async (challengeId: string, otp: string) => {
      const res = await authClient.verifyMFA(challengeId, otp);
      setAuthFromResponse(res);
    },
    [setAuthFromResponse]
  );

  const logout = useCallback(async () => {
    const refreshToken = state?.refreshToken;
    clearAuth();
    if (refreshToken) {
      try {
        await authClient.logout(refreshToken);
      } catch {
        // ignore
      }
    }
  }, [state?.refreshToken, clearAuth]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user: state?.user ?? null,
      isAuthenticated: state !== null,
      isLoading,
      login,
      verifyMFA,
      logout,
      refresh,
      setAuthFromResponse,
      clearAuth,
    }),
    [state, isLoading, login, verifyMFA, logout, refresh, setAuthFromResponse, clearAuth]
  );

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (ctx == null) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
