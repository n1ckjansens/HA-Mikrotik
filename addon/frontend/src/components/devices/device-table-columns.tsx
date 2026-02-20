import type { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, MoreHorizontal } from "lucide-react";
import { toast } from "sonner";

import { CopyValue } from "@/components/devices/CopyValue";
import { DeviceTypeIcon } from "@/components/device/DeviceTypeIcon";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from "@/components/ui/tooltip";
import { inferDeviceType } from "@/lib/device";
import {
  isDeviceOnline,
  isNewDevice,
  isUnregisteredDevice
} from "@/lib/device-semantics";
import { copyTextToClipboard } from "@/lib/clipboard";
import { formatExactTimestamp, formatLastSeenLabel } from "@/lib/time";
import type { Device } from "@/types/device";

async function copyToClipboard(value: string, label: string) {
  const copied = await copyTextToClipboard(value);
  if (copied) {
    toast.success(`${label} copied`);
  } else {
    toast.error(`Unable to copy ${label.toLowerCase()}`);
  }
}

type BuildColumnsParams = {
  isMobile: boolean;
  onOpenDevice: (device: Device, trigger: HTMLElement | null) => void;
};

export function buildDeviceColumns({
  isMobile,
  onOpenDevice
}: BuildColumnsParams): Array<ColumnDef<Device>> {
  const columns: Array<ColumnDef<Device>> = [
    {
      accessorKey: "name",
      header: () => <span>Name</span>,
      cell: ({ row }) => {
        const device = row.original;
        const isNew = isNewDevice(device);
        const isOnline = isDeviceOnline(device);

        return (
          <div className="flex min-w-0 items-start gap-3">
            <DeviceTypeIcon type={inferDeviceType(device)} />
            <div className="min-w-0 flex-1 space-y-1">
              <Button
                type="button"
                variant="ghost"
                className="h-auto w-full min-w-0 justify-start overflow-hidden p-0 hover:bg-transparent"
                onClick={(event) => onOpenDevice(device, event.currentTarget)}
              >
                <div className="flex min-w-0 items-center gap-2">
                  <p className="min-w-0 flex-1 truncate text-left font-medium">{device.name}</p>
                  {isNew ? (
                    <Badge variant="outline" aria-label="New device" className="shrink-0">
                      NEW
                    </Badge>
                  ) : null}
                  <Badge
                    variant={isOnline ? "default" : "outline"}
                    className={`shrink-0 ${
                      isOnline ? "bg-emerald-600 hover:bg-emerald-600" : "text-muted-foreground"
                    }`}
                    aria-label={isOnline ? "Online" : "Offline"}
                  >
                    {isOnline ? "Online" : "Offline"}
                  </Badge>
                </div>
              </Button>

              <CopyValue
                label="MAC"
                value={device.mac}
                mono
                className="h-6 w-fit max-w-full px-1 text-xs text-muted-foreground"
                onClick={(event) => {
                  event.stopPropagation();
                }}
              />
            </div>
          </div>
        );
      }
    },
    {
      accessorKey: "last_ip",
      header: ({ column }) => (
        <Button
          variant="ghost"
          size="sm"
          className="px-0"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          IP
          <ArrowUpDown className="ml-1 h-3.5 w-3.5" />
        </Button>
      ),
      cell: ({ row }) => (
        <CopyValue
          label="IP"
          value={row.original.last_ip}
          mono
          className="h-7 px-1 text-xs text-muted-foreground"
        />
      )
    },
    {
      accessorKey: "last_subnet",
      header: "Subnet",
      cell: ({ row }) => (
        <CopyValue
          label="Subnet"
          value={row.original.last_subnet}
          mono
          className="h-7 px-1 text-xs text-muted-foreground"
        />
      )
    },
    {
      accessorKey: "vendor",
      header: "Vendor",
      cell: ({ row }) => <span className="block truncate">{row.original.vendor || "Unknown"}</span>
    },
    {
      id: "source",
      header: "Source",
      cell: ({ row }) => {
        const source = row.original.last_sources[0] ?? "-";
        return <span className="block truncate uppercase text-xs text-muted-foreground">{source}</span>;
      }
    },
    {
      id: "last_seen",
      accessorFn: (row) => row.last_seen_at ?? "",
      header: ({ column }) => (
        <Button
          variant="ghost"
          size="sm"
          className="px-0"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Last seen
          <ArrowUpDown className="ml-1 h-3.5 w-3.5" />
        </Button>
      ),
      cell: ({ row }) => {
        const device = row.original;
        const relative = formatLastSeenLabel(
          isDeviceOnline(device),
          device.last_seen_at,
          Date.now()
        );
        return (
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="cursor-help text-sm">{relative}</span>
            </TooltipTrigger>
            <TooltipContent>{formatExactTimestamp(device.last_seen_at)}</TooltipContent>
          </Tooltip>
        );
      }
    },
    {
      id: "registration",
      header: "Registration",
      cell: ({ row }) => {
        const device = row.original;
        const unregistered = isUnregisteredDevice(device);
        return unregistered ? (
          <Badge variant="outline" aria-label="Unregistered">
            Unregistered
          </Badge>
        ) : (
          <Badge variant="secondary" aria-label="Registered">
            Registered
          </Badge>
        );
      }
    },
    {
      id: "actions",
      enableHiding: false,
      cell: ({ row }) => {
        const device = row.original;
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" aria-label={`Actions for ${device.name}`}>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={(event) => {
                  onOpenDevice(device, event.currentTarget as HTMLElement);
                }}
              >
                Open details
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => void copyToClipboard(device.mac, "MAC")}>
                Copy MAC
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => {
                  if (device.last_ip) {
                    void copyToClipboard(device.last_ip, "IP");
                  }
                }}
                disabled={!device.last_ip}
              >
                Copy IP
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={(event) => {
                  onOpenDevice(device, event.currentTarget as HTMLElement);
                }}
              >
                {isUnregisteredDevice(device) ? "Register" : "Edit"}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      }
    }
  ];

  if (isMobile) {
    return columns.map((column) => {
      if (
        column.id === "last_subnet" ||
        column.id === "vendor" ||
        column.id === "source" ||
        column.id === "registration"
      ) {
        return { ...column, enableHiding: true };
      }
      return column;
    });
  }

  return columns;
}
