import type { Device } from "@/types/device";

type Props = {
  device: Device;
};

export function DeviceMeta({ device }: Props) {
  const ip = device.last_ip ?? "IP unknown";
  const subnet = device.last_subnet ?? "Subnet unknown";
  const vendor = device.vendor || "Unknown vendor";

  return (
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground">
        {ip} • {subnet} • {vendor}
      </p>
      <p className="font-mono text-[11px] text-muted-foreground">{device.mac}</p>
    </div>
  );
}
