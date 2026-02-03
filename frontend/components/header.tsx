"use client";

import Link from "next/link";

const DOCS_URL = process.env.NEXT_PUBLIC_DOCS_URL ?? "/docs";

/**
 * Global header with app title and Docs link. Rendered in the root layout.
 */
export function Header() {
  return (
    <header className="border-b bg-background">
      <div className="mx-auto flex h-14 max-w-4xl items-center justify-between px-4">
        <Link href="/" className="font-semibold text-foreground hover:underline">
          Zero Trust Control Plane
        </Link>
        <a
          href={DOCS_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm text-muted-foreground hover:text-foreground hover:underline"
        >
          Docs
        </a>
      </div>
    </header>
  );
}
