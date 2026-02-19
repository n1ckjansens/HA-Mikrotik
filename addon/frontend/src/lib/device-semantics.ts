import type { Device } from "@/types/device";

export type RegistrationScope = "all" | "new" | "registered" | "unregistered";
export type OnlineScope = "any" | "online" | "offline";

export function isNewDevice(device: Device) {
  return device.status === "new";
}

export function isRegisteredDevice(device: Device) {
  return device.status === "registered";
}

export function isUnregisteredDevice(device: Device) {
  return !isRegisteredDevice(device);
}

export function isDeviceOnline(device: Device, nowMs = Date.now()) {
  void nowMs;
  return device.online;
}

export function matchRegistrationScope(
  device: Device,
  scope: RegistrationScope
) {
  if (scope === "all") {
    return true;
  }
  if (scope === "new") {
    return isNewDevice(device);
  }
  if (scope === "registered") {
    return isRegisteredDevice(device);
  }
  return isUnregisteredDevice(device);
}

export function matchOnlineScope(
  device: Device,
  scope: OnlineScope,
  nowMs = Date.now()
) {
  const online = isDeviceOnline(device, nowMs);
  if (scope === "any") {
    return true;
  }
  if (scope === "online") {
    return online;
  }
  return !online;
}

export function matchSearch(device: Device, search: string) {
  const query = search.trim().toLowerCase();
  if (!query) {
    return true;
  }

  const parts = [
    device.name,
    device.mac,
    device.vendor,
    device.last_ip ?? "",
    device.last_subnet ?? "",
    ...(device.last_sources ?? [])
  ];

  return parts.join(" ").toLowerCase().includes(query);
}

export function matchFacet(value: string | null | undefined, selected: string[]) {
  if (selected.length === 0) {
    return true;
  }
  const normalized = (value ?? "Unknown").trim();
  return selected.includes(normalized);
}

export function uniqueSorted(values: Array<string | null | undefined>) {
  return Array.from(
    new Set(
      values
        .map((value) => value?.trim() || "Unknown")
        .filter((value) => value !== "")
    )
  ).sort((a, b) => a.localeCompare(b));
}
