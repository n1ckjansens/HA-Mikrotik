import type { Device } from "@/types/device";

export type DeviceType = "wifi" | "wired" | "unknown";

function normalize(value: string | null | undefined) {
  return (value ?? "").trim().toLowerCase();
}

export function inferDeviceType(device: Device): DeviceType {
  const icon = normalize(device.icon);
  if (["wifi", "wireless", "wlan"].includes(icon)) {
    return "wifi";
  }
  if (["wired", "ethernet", "eth", "lan"].includes(icon)) {
    return "wired";
  }

  if (device.last_sources.includes("wifi")) {
    return "wifi";
  }
  if (device.last_sources.some((source) => ["dhcp", "arp", "bridge"].includes(source))) {
    return "wired";
  }

  return "unknown";
}

export function parseStoredIcon(
  icon: string | null | undefined,
  fallback: DeviceType = "unknown"
): DeviceType {
  const normalized = normalize(icon);
  if (["wifi", "wireless", "wlan"].includes(normalized)) {
    return "wifi";
  }
  if (["wired", "ethernet", "eth", "lan"].includes(normalized)) {
    return "wired";
  }
  return fallback;
}

export function toStoredIcon(type: DeviceType): string | undefined {
  if (type === "unknown") {
    return undefined;
  }
  return type;
}

export function getSourceBreakdown(device: Device) {
  const sources = new Set(device.last_sources.map((value) => value.toLowerCase()));
  return {
    dhcp: sources.has("dhcp"),
    wifi: sources.has("wifi"),
    arp: sources.has("arp"),
    bridge: sources.has("bridge")
  };
}

function readInterface(source: unknown): string | null {
  if (!source || typeof source !== "object") {
    return null;
  }
  const maybeRecord = source as Record<string, unknown>;
  const value = maybeRecord.Interface ?? maybeRecord.interface;
  if (typeof value !== "string") {
    return null;
  }
  const trimmed = value.trim();
  return trimmed === "" ? null : trimmed;
}

export function getPrimaryInterface(device: Device): string | null {
  const directInterface = [device.interface, device.wifi_interface, device.arp_interface, device.bridge_host_port]
    .map((value) => (value ?? "").trim())
    .find((value) => value !== "");
  if (directInterface) {
    return directInterface;
  }

  if (!device.raw_sources || typeof device.raw_sources !== "object") {
    return null;
  }

  const sources = device.raw_sources as Record<string, unknown>;
  return (
    readInterface(sources.wifi) ??
    readInterface(sources.bridge) ??
    readInterface(sources.arp) ??
    readInterface(sources.dhcp) ??
    null
  );
}
