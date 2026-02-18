import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import type { Device } from "@/types/device";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  device: Device | null;
};

function formatDate(value?: string | null) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

export function DeviceCard({ open, onOpenChange, device }: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{device?.name ?? "Device"}</DialogTitle>
          <DialogDescription>{device?.mac ?? ""}</DialogDescription>
        </DialogHeader>
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
      </DialogContent>
    </Dialog>
  );
}
