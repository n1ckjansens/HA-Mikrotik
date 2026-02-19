import type { OnlineFilter } from "@/types/device";
import type { StatusFilter } from "@/hooks/useFilters";
import { DeviceTabs } from "@/components/filters/DeviceTabs";
import { OnlineFilter as OnlineFilterControl } from "@/components/filters/OnlineFilter";
import { Card, CardContent } from "@/components/ui/card";

type Props = {
  status: StatusFilter;
  online: OnlineFilter;
  onStatusChange: (value: StatusFilter) => void;
  onOnlineChange: (value: OnlineFilter) => void;
};

export function DeviceFilters({
  status,
  online,
  onStatusChange,
  onOnlineChange
}: Props) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-4 p-4 md:flex-row md:items-center md:gap-6">
        <div className="grid gap-2">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">Device Type</p>
          <DeviceTabs value={status} onChange={onStatusChange} />
        </div>
        <div className="h-px bg-border md:h-10 md:w-px" aria-hidden />
        <div className="grid gap-2">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">Online State</p>
          <OnlineFilterControl value={online} onChange={onOnlineChange} />
        </div>
      </CardContent>
    </Card>
  );
}
