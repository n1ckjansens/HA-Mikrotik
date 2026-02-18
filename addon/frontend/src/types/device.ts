import { z } from "zod";

export const deviceSchema = z.object({
  mac: z.string(),
  name: z.string(),
  vendor: z.string(),
  icon: z.string().nullable().optional(),
  comment: z.string().nullable().optional(),
  status: z.enum(["new", "registered"]),
  online: z.boolean(),
  last_seen_at: z.string().datetime().nullable().optional(),
  connected_since_at: z.string().datetime().nullable().optional(),
  last_ip: z.string().nullable().optional(),
  last_subnet: z.string().nullable().optional(),
  last_sources: z.array(z.string()),
  raw_sources: z.unknown().optional(),
  created_at: z.string().datetime().nullable().optional(),
  updated_at: z.string().datetime(),
  first_seen_at: z.string().datetime().nullable().optional()
});

export const listDevicesResponseSchema = z.object({
  items: z.array(deviceSchema)
});

export type Device = z.infer<typeof deviceSchema>;
export type DeviceStatus = Device["status"];
export type OnlineFilter = "all" | "online" | "offline";
