import { Cable, Monitor, Wifi } from "lucide-react";

import type { DeviceType } from "@/lib/device";

type Props = {
  type: DeviceType;
};

export function DeviceTypeIcon({ type }: Props) {
  if (type === "wifi") {
    return (
      <span className="rounded-md border bg-muted p-2" aria-label="WiFi device">
        <Wifi className="h-4 w-4" />
      </span>
    );
  }

  if (type === "wired") {
    return (
      <span className="rounded-md border bg-muted p-2" aria-label="Wired device">
        <Cable className="h-4 w-4" />
      </span>
    );
  }

  return (
    <span className="rounded-md border bg-muted p-2" aria-label="Unknown device type">
      <Monitor className="h-4 w-4" />
    </span>
  );
}
