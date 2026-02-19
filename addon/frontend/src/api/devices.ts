import { z } from "zod";

import { apiRequest } from "@/api/client";
import {
  deviceSchema,
  listDevicesResponseSchema,
  type Device,
  type OnlineFilter
} from "@/types/device";

export type DevicesFilterParams = {
  status: "all" | "new" | "registered";
  online: OnlineFilter;
  query: string;
};

export async function fetchDevices(params: DevicesFilterParams): Promise<Device[]> {
  const query = new URLSearchParams();
  if (params.status !== "all") {
    query.set("status", params.status);
  }
  if (params.online === "online") {
    query.set("online", "true");
  }
  if (params.online === "offline") {
    query.set("online", "false");
  }
  if (params.query.trim() !== "") {
    query.set("query", params.query.trim());
  }

  const suffix = query.toString();
  const raw = await apiRequest<unknown>(suffix ? `/api/devices?${suffix}` : "/api/devices");
  return listDevicesResponseSchema.parse(raw).items;
}

export async function fetchDevice(mac: string): Promise<Device> {
  const raw = await apiRequest<unknown>(`/api/devices/${encodeURIComponent(mac)}`);
  return deviceSchema.parse(raw);
}

const registerDeviceSchema = z.object({
  name: z.string().min(1).optional(),
  icon: z.string().min(1).optional(),
  comment: z.string().optional()
});

export type RegisterDeviceInput = z.infer<typeof registerDeviceSchema>;

export async function registerDevice(mac: string, input: RegisterDeviceInput) {
  const payload = registerDeviceSchema.parse(input);
  return apiRequest<{ ok: true }>(`/api/devices/${encodeURIComponent(mac)}/register`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function patchDevice(mac: string, input: RegisterDeviceInput) {
  const payload = registerDeviceSchema.partial().parse(input);
  return apiRequest<{ ok: true }>(`/api/devices/${encodeURIComponent(mac)}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export async function refreshDevices() {
  return apiRequest<{ ok: true }>("/api/refresh", { method: "POST", body: "{}" });
}
