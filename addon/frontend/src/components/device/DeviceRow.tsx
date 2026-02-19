import { ChevronRight } from "lucide-react";

import { DeviceLastSeen } from "@/components/device/DeviceLastSeen";
import { DeviceMeta } from "@/components/device/DeviceMeta";
import { DeviceStatusBadge } from "@/components/device/DeviceStatusBadge";
import { DeviceTypeIcon } from "@/components/device/DeviceTypeIcon";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { inferDeviceType } from "@/lib/device";
import { cn } from "@/lib/utils";
import type { Device } from "@/types/device";

type Props = {
  device: Device;
  now: number;
  onOpen: (mac: string) => void;
};

export function DeviceRow({ device, now, onOpen }: Props) {
  return (
    <Card
      className={cn(
        "transition-colors hover:border-primary/60",
        "focus-within:border-primary/70",
        device.status === "new" && "border-l-4 border-l-blue-500",
        !device.online && "opacity-90"
      )}
    >
      <button
        type="button"
        className="w-full cursor-pointer text-left"
        onClick={() => onOpen(device.mac)}
        aria-label={`Open details for ${device.name}`}
      >
        <CardContent className="p-4">
          <div className="flex items-start gap-3">
            <DeviceTypeIcon type={inferDeviceType(device)} />
            <div className="min-w-0 flex-1 space-y-2">
              <div className="flex items-start justify-between gap-3">
                <div className="flex min-w-0 items-center gap-2">
                  <p className="truncate text-sm font-semibold">{device.name}</p>
                  {device.status === "new" ? (
                    <Badge
                      variant="outline"
                      className="border-blue-500/60 text-blue-700"
                      aria-label="New device"
                    >
                      NEW
                    </Badge>
                  ) : null}
                </div>
                <DeviceStatusBadge online={device.online} />
              </div>

              <DeviceMeta device={device} />
              <DeviceLastSeen online={device.online} lastSeenAt={device.last_seen_at} now={now} />
            </div>
            <ChevronRight className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" aria-hidden />
          </div>
        </CardContent>
      </button>
    </Card>
  );
}
