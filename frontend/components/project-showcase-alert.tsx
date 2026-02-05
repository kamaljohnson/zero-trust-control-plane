"use client";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

const DOCS_URL = process.env.NEXT_PUBLIC_DOCS_URL ?? "/docs";

/**
 * Project showcase alert component that displays project details and a Docs button.
 * Designed to be shown below the navbar as a showcase banner.
 */
export function ProjectShowcaseAlert() {
  return (
    <div className="flex justify-center">
      <Alert className="w-full max-w-5xl m-4 border-primary/50 bg-primary/5">
        <div className="flex items-center justify-between gap-4">
          <div className="flex-1">
            <AlertTitle className="text-base font-semibold mb-1">
              Zero Trust Control Plane
            </AlertTitle>
            <AlertDescription className="text-sm text-muted-foreground">
              A proof-of-concept zero-trust session and policy control plane with backend (Go gRPC) 
              and web client (Next.js). Features authentication, session management, policy engine, 
              multi-tenancy, org admin dashboard, and telemetry.
            </AlertDescription>
          </div>
          <Button
            asChild
            variant="outline"
            size="sm"
            className="shrink-0"
          >
            <a
              href={DOCS_URL}
              target="_blank"
              rel="noopener noreferrer"
            >
              View Docs
            </a>
          </Button>
        </div>
      </Alert>
    </div>
  );
}
