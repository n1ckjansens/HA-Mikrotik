import { ResponsiveOverlay } from "@/components/ui/responsive-overlay";
import type { Device } from "@/types/device";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  device: Device | null;
  isMobile: boolean;
};

function formatDate(value?: string | null) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

function DeviceDetailsBody({ device }: { device: Device | null }) {
  return (
    <div className="grid gap-2 text-sm">
      <p>Vendor: {device?.vendor ?? "-"}</p>
      <p>Status: {device?.status ?? "-"}</p>
      <p>Online: {device?.online ? "Yes" : "No"}</p>
      <p>IP: {device?.last_ip ?? "-"}</p>
      <p>Subnet: {device?.last_subnet ?? "-"}</p>
      <p>Last seen: {formatDate(device?.last_seen_at)}</p>
      <p>Connected since: {formatDate(device?.connected_since_at)}</p>
      <p>Sources: {(device?.last_sources ?? []).join(", ") || "-"}</p>
      <p>Comment: {device?.comment ?? "-"}</p>
      <p>Raw MAC: {device?.mac ?? "-"}</p>
    </div>
  );
}

export function DeviceCard({ open, onOpenChange, device, isMobile }: Props) {
  return (
    <ResponsiveOverlay
      open={open}
      isMobile={isMobile}
      onOpenChange={onOpenChange}
      title={device?.name ?? "Device"}
      description={device?.mac ?? ""}
    >
      <DeviceDetailsBody device={device} />
    </ResponsiveOverlay>
  );
}
