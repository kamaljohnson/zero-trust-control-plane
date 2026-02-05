"use client";

import Link from "next/link";

/**
 * Global error boundary. Catches errors in the root layout (e.g. AuthProvider)
 * and replaces the entire document. Must define its own html and body.
 * reset() re-renders the root.
 */
export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body style={{ fontFamily: "system-ui, sans-serif", padding: "2rem" }}>
        <div style={{ maxWidth: "28rem", margin: "0 auto" }}>
          <h1 style={{ fontSize: "1.5rem", marginBottom: "0.5rem" }}>
            Something went wrong
          </h1>
          <p style={{ color: "#737373", marginBottom: "1.5rem" }}>
            An unexpected error occurred. You can try again or go to the home
            page.
          </p>
          {error.message && (
            <p style={{ fontSize: "0.875rem", color: "#737373", marginBottom: "1rem" }}>
              {error.message}
            </p>
          )}
          <div style={{ display: "flex", gap: "0.75rem" }}>
            <button
              type="button"
              onClick={() => reset()}
              style={{
                padding: "0.5rem 1rem",
                background: "#171717",
                color: "#fafafa",
                border: "none",
                borderRadius: "0.375rem",
                cursor: "pointer",
              }}
            >
              Try again
            </button>
            <Link
              href="/"
              style={{
                padding: "0.5rem 1rem",
                border: "1px solid #e5e5e5",
                borderRadius: "0.375rem",
                color: "inherit",
                textDecoration: "none",
              }}
            >
              Go to home
            </Link>
          </div>
        </div>
      </body>
    </html>
  );
}
