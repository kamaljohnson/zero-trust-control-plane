"use client";

import {
  Card,
  CardContent,
} from "@/components/ui/card";

interface SessionInfoCardProps {
  userName: string;
  userEmail: string;
  orgName: string;
  orgId: string;
}

/**
 * SessionInfoCard displays user and organization information in a card format.
 */
export function SessionInfoCard({
  userName,
  userEmail,
  orgName,
  orgId,
}: SessionInfoCardProps) {
  return (
    <Card className="w-80">
      <CardContent className="space-y-4 pt-6">
        {/* User Section */}
        <div className="space-y-1">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            User
          </div>
          <div className="space-y-0.5">
            <div className="text-sm font-medium text-foreground">{userName}</div>
            <div className="text-xs text-muted-foreground">{userEmail}</div>
          </div>
        </div>

        {/* Organization Section */}
        <div className="space-y-1 border-t pt-4">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Organization
          </div>
          <div className="space-y-0.5">
            <div className="text-sm font-medium text-foreground">{orgName}</div>
            <div className="text-xs text-muted-foreground font-mono">{orgId}</div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
