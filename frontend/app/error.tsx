"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

/**
 * Root error boundary for the app. Catches uncaught errors in the tree below
 * the root layout and shows a fallback UI instead of a white screen.
 * reset() re-renders the segment so the user can try again.
 */
export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log to console or reporting service in production.
    console.error("App error boundary caught:", error.message, error.digest);
  }, [error]);

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Something went wrong</CardTitle>
          <CardDescription>
            An unexpected error occurred. You can try again or return to the home
            page.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error.message && (
            <p className="text-sm text-muted-foreground" role="alert">
              {error.message}
            </p>
          )}
        </CardContent>
        <CardFooter className="flex gap-3">
          <Button onClick={() => reset()}>Try again</Button>
          <Button variant="outline" asChild>
            <a href="/">Go to home</a>
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
