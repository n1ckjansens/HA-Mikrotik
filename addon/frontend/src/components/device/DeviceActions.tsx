import { Button } from "@/components/ui/button";
import type { Device } from "@/types/device";

type Props = {
  device: Device;
  busy: boolean;
  onOpenDetails: (mac: string) => void;
  onRegister: (mac: string) => void;
};

export function DeviceActions({ device, busy, onOpenDetails, onRegister }: Props) {
  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={() => onOpenDetails(device.mac)}>
        Details
      </Button>
      {device.status === "new" ? (
        <Button size="sm" disabled={busy} onClick={() => onRegister(device.mac)}>
          Register
        </Button>
      ) : null}
    </div>
  );
}
