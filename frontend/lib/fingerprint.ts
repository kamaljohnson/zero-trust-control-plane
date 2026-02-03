/**
 * Device fingerprint for login and device-trust. Uses FingerprintJS open-source
 * to compute a stable visitorId in the browser. Falls back to "password-login"
 * on SSR, import/load errors, or when window is undefined so login still works.
 */

const FALLBACK_FINGERPRINT = "password-login";

let cachedFingerprint: string | null = null;

/**
 * Returns a stable device/browser fingerprint for device trust. Runs only in the
 * browser; uses dynamic import so FingerprintJS is not pulled into server bundles.
 * Result is cached for the session to avoid re-running the agent on every login.
 *
 * @returns The FingerprintJS visitorId, or FALLBACK_FINGERPRINT on any failure
 */
export async function getDeviceFingerprint(): Promise<string> {
  if (cachedFingerprint !== null) {
    return cachedFingerprint;
  }
  if (typeof window === "undefined") {
    return FALLBACK_FINGERPRINT;
  }
  try {
    const FingerprintJS = (
      await import("@fingerprintjs/fingerprintjs")
    ).default;
    const agent = await FingerprintJS.load();
    const { visitorId } = await agent.get();
    cachedFingerprint = visitorId;
    return visitorId;
  } catch {
    return FALLBACK_FINGERPRINT;
  }
}
