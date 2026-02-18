import { Wifi } from "lucide-react";

import { DeviceActions } from "@/components/device/DeviceActions";
import { DeviceOnlineBadge } from "@/components/device/DeviceOnlineBadge";
import { DeviceStatusBadge } from "@/components/device/DeviceStatusBadge";
import { Card, CardContent } from "@/components/ui/card";
import type { Device } from "@/types/device";

type Props = {
  device: Device;
  busy: boolean;
  onOpenDetails: (mac: string) => void;
  onRegister: (mac: string) => void;
};

function formatLastSeen(value?: string | null) {
  if (!value) {
    return "never";
  }
  return new Date(value).toLocaleString();
}

export function DeviceRow({ device, busy, onOpenDetails, onRegister }: Props) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-4 pt-6 md:flex-row md:items-center md:justify-between">
        <div className="flex items-start gap-3">
          <Wifi className="mt-1 h-4 w-4" />
          <div className="space-y-1">
            <p className="text-sm font-semibold">{device.name}</p>
            <p className="font-mono text-xs text-muted-foreground">{device.mac}</p>
            <p className="text-xs text-muted-foreground">{device.vendor}</p>
            <p className="text-xs text-muted-foreground">
              {device.last_ip ?? "-"} {device.last_subnet ? `(${device.last_subnet})` : ""}
            </p>
            <p className="text-xs text-muted-foreground">Last seen: {formatLastSeen(device.last_seen_at)}</p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <DeviceStatusBadge status={device.status} />
          <DeviceOnlineBadge online={device.online} />
          <DeviceActions
            device={device}
            busy={busy}
            onOpenDetails={onOpenDetails}
            onRegister={onRegister}
          />
        </div>
      </CardContent>
    </Card>
  );
}
