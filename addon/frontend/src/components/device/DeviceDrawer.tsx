import type { ReactNode } from "react";
import { Cable, Monitor, Wifi } from "lucide-react";

import { DeviceStatusBadge } from "@/components/device/DeviceStatusBadge";
import { DeviceTypeIcon } from "@/components/device/DeviceTypeIcon";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ResponsiveOverlay } from "@/components/ui/responsive-overlay";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import {
  getPrimaryInterface,
  getSourceBreakdown,
  type DeviceType
} from "@/lib/device";
import { formatExactTimestamp } from "@/lib/time";
import type { Device } from "@/types/device";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  device: Device | null;
  isLoading: boolean;
  isMobile: boolean;
  name: string;
  icon: DeviceType;
  comment: string;
  isSaving: boolean;
  onNameChange: (value: string) => void;
  onIconChange: (value: DeviceType) => void;
  onCommentChange: (value: string) => void;
  onSave: () => void;
};

type ItemProps = {
  label: string;
  value: string;
};

function DataItem({ label, value }: ItemProps) {
  return (
    <div className="grid gap-0.5">
      <p className="text-xs uppercase tracking-wide text-muted-foreground">{label}</p>
      <p className="text-sm">{value}</p>
    </div>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3">{children}</CardContent>
    </Card>
  );
}

function SourceBadge({ active, label }: { active: boolean; label: string }) {
  return (
    <Badge
      variant={active ? "secondary" : "outline"}
      className={active ? "border-primary/30 bg-primary/10 text-primary" : "text-muted-foreground"}
      aria-label={`${label} source ${active ? "active" : "inactive"}`}
    >
      {label}
    </Badge>
  );
}

export function DeviceDrawer({
  open,
  onOpenChange,
  device,
  isLoading,
  isMobile,
  name,
  icon,
  comment,
  isSaving,
  onNameChange,
  onIconChange,
  onCommentChange,
  onSave
}: Props) {
  const breakdown = device
    ? getSourceBreakdown(device)
    : { dhcp: false, wifi: false, arp: false, bridge: false };

  return (
    <ResponsiveOverlay
      open={open}
      isMobile={isMobile}
      onOpenChange={onOpenChange}
      title={device?.name ?? "Device"}
      description={device?.mac ?? ""}
    >
      {isLoading ? <p className="text-sm text-muted-foreground">Loading device details...</p> : null}

      <Section title="Identity">
        <div className="flex items-start gap-3">
          <DeviceTypeIcon type={icon} />
          <div className="grid gap-2">
            <DataItem label="Name" value={device?.name ?? "-"} />
            <div className="flex items-center gap-2">
              <DeviceStatusBadge online={Boolean(device?.online)} />
              {device?.status === "new" ? (
                <Badge
                  variant="outline"
                  className="border-blue-500/60 text-blue-700"
                  aria-label="New device"
                >
                  NEW
                </Badge>
              ) : null}
            </div>
          </div>
        </div>
        <DataItem label="Vendor" value={device?.vendor || "Unknown"} />
        <DataItem label="MAC" value={device?.mac ?? "-"} />
      </Section>

      <Section title="Network">
        <DataItem label="IP" value={device?.last_ip ?? "-"} />
        <DataItem label="Subnet" value={device?.last_subnet ?? "-"} />
        <DataItem label="Interface" value={device ? getPrimaryInterface(device) ?? "-" : "-"} />
        <div className="grid gap-1">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">Source Breakdown</p>
          <div className="flex flex-wrap gap-2">
            <SourceBadge active={breakdown.dhcp} label="DHCP" />
            <SourceBadge active={breakdown.wifi} label="WiFi" />
            <SourceBadge active={breakdown.arp} label="ARP" />
            <SourceBadge active={breakdown.bridge} label="Bridge" />
          </div>
        </div>
      </Section>

      <Section title="State">
        <DataItem label="Online" value={device?.online ? "Yes" : "No"} />
        <DataItem label="Last Seen" value={formatExactTimestamp(device?.last_seen_at)} />
        <DataItem
          label="Connected Since"
          value={formatExactTimestamp(device?.connected_since_at)}
        />
        <DataItem label="First Seen" value={formatExactTimestamp(device?.first_seen_at)} />
      </Section>

      <Section title="Registration">
        <div className="grid gap-2">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">Display Name</p>
          <Input
            value={name}
            onChange={(event) => onNameChange(event.target.value)}
            placeholder="Living Room iPhone"
          />
        </div>

        <div className="grid gap-2">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">Icon</p>
          <ToggleGroup
            type="single"
            value={icon}
            variant="outline"
            onValueChange={(next) => {
              if (next === "wifi" || next === "wired" || next === "unknown") {
                onIconChange(next);
              }
            }}
          >
            <ToggleGroupItem value="wifi" aria-label="WiFi icon">
              <Wifi className="mr-1 h-4 w-4" /> WiFi
            </ToggleGroupItem>
            <ToggleGroupItem value="wired" aria-label="Wired icon">
              <Cable className="mr-1 h-4 w-4" /> Wired
            </ToggleGroupItem>
            <ToggleGroupItem value="unknown" aria-label="Unknown icon">
              <Monitor className="mr-1 h-4 w-4" /> Unknown
            </ToggleGroupItem>
          </ToggleGroup>
        </div>

        <div className="grid gap-2">
          <p className="text-xs uppercase tracking-wide text-muted-foreground">Comment</p>
          <Input
            value={comment}
            onChange={(event) => onCommentChange(event.target.value)}
            placeholder="Optional note"
          />
        </div>

        <div className="flex justify-end">
          <Button
            onClick={onSave}
            disabled={isSaving || !device || name.trim() === ""}
            aria-label={device?.status === "new" ? "Register device" : "Save device"}
          >
            {device?.status === "new" ? "Register device" : "Save changes"}
          </Button>
        </div>
      </Section>
    </ResponsiveOverlay>
  );
}
