import { DeviceRow } from "@/components/device/DeviceRow";
import type { Device } from "@/types/device";

type Props = {
  devices: Device[];
  now: number;
  onOpenDevice: (mac: string) => void;
};

export function DeviceList({ devices, now, onOpenDevice }: Props) {
  return (
    <div className="grid gap-3 md:gap-4">
      {devices.map((device) => (
        <DeviceRow key={device.mac} device={device} now={now} onOpen={onOpenDevice} />
      ))}
    </div>
  );
}
