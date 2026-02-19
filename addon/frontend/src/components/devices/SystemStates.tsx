import { PlugZap, Router, SearchX, TriangleAlert } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

export function AddonConfigurationRequiredState() {
  return (
    <Card>
      <CardContent className="space-y-4 p-6">
        <div className="flex items-start gap-2">
          <Router className="mt-0.5 h-4 w-4 text-muted-foreground" />
          <div>
            <p className="font-medium">Add-on is not configured</p>
            <p className="text-sm text-muted-foreground">
              Open Settings {" > "}Add-ons {" > "}MikroTik Presence {" > "}
              Configuration and set router credentials.
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function DisconnectedState({
  message,
  onRetry,
  lastSuccessfulLabel
}: {
  message: string;
  onRetry: () => void;
  lastSuccessfulLabel: string;
}) {
  return (
    <Alert variant="destructive">
      <TriangleAlert className="h-4 w-4" />
      <AlertTitle>Router disconnected</AlertTitle>
      <AlertDescription className="flex flex-wrap items-center gap-2">
        <span>{message}</span>
        <span>Last successful update: {lastSuccessfulLabel}</span>
        <Button variant="destructive" size="sm" onClick={onRetry}>
          Retry
        </Button>
      </AlertDescription>
    </Alert>
  );
}

export function NoDevicesState() {
  return (
    <Card>
      <CardContent className="grid place-items-center gap-2 p-8 text-center">
        <div className="rounded-full border bg-muted p-3">
          <PlugZap className="h-6 w-6 text-muted-foreground" />
        </div>
        <p className="font-medium">No devices detected yet</p>
        <p className="text-sm text-muted-foreground">Devices will appear after router sync.</p>
      </CardContent>
    </Card>
  );
}

export function NoResultsState({ onClear }: { onClear: () => void }) {
  return (
    <Card>
      <CardContent className="grid place-items-center gap-2 p-8 text-center">
        <div className="rounded-full border bg-muted p-3">
          <SearchX className="h-6 w-6 text-muted-foreground" />
        </div>
        <p className="font-medium">No results for current filters</p>
        <Button variant="outline" onClick={onClear}>
          Clear filters
        </Button>
      </CardContent>
    </Card>
  );
}
