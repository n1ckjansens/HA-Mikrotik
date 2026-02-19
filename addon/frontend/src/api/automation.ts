import { z } from "zod";

import { apiRequest } from "@/api/client";
import {
  actionTypeSchema,
  capabilityDeviceAssignmentSchema,
  capabilityTemplateSchema,
  capabilityUIModelSchema,
  stateSourceTypeSchema,
  setStateResultSchema,
  type CapabilityTemplate
} from "@/types/automation";

const actionTypesSchema = z.array(actionTypeSchema);
const stateSourceTypesSchema = z.array(stateSourceTypeSchema);
const capabilitiesSchema = z.array(capabilityTemplateSchema);
const capabilityAssignmentsSchema = z.array(capabilityDeviceAssignmentSchema);
const capabilityUIModelsSchema = z.array(capabilityUIModelSchema);

export type CapabilitiesQuery = {
  search?: string;
  category?: string;
};

export async function fetchActionTypes() {
  const raw = await apiRequest<unknown>("/api/automation/action-types");
  return actionTypesSchema.parse(raw);
}

export async function fetchStateSourceTypes() {
  const raw = await apiRequest<unknown>("/api/automation/state-source-types");
  return stateSourceTypesSchema.parse(raw);
}

export async function fetchCapabilities(query: CapabilitiesQuery) {
  const params = new URLSearchParams();
  if (query.search?.trim()) {
    params.set("search", query.search.trim());
  }
  if (query.category?.trim()) {
    params.set("category", query.category.trim());
  }
  const suffix = params.toString();
  const raw = await apiRequest<unknown>(
    suffix ? `/api/automation/capabilities?${suffix}` : "/api/automation/capabilities"
  );
  return capabilitiesSchema.parse(raw);
}

export async function fetchCapability(capabilityId: string) {
  const raw = await apiRequest<unknown>(
    `/api/automation/capabilities/${encodeURIComponent(capabilityId)}`
  );
  return capabilityTemplateSchema.parse(raw);
}

export async function createCapability(template: CapabilityTemplate) {
  const payload = capabilityTemplateSchema.parse(template);
  const raw = await apiRequest<unknown>("/api/automation/capabilities", {
    method: "POST",
    body: JSON.stringify(payload)
  });
  return capabilityTemplateSchema.parse(raw);
}

export async function updateCapability(capabilityId: string, template: CapabilityTemplate) {
  const payload = capabilityTemplateSchema.parse(template);
  const raw = await apiRequest<unknown>(
    `/api/automation/capabilities/${encodeURIComponent(capabilityId)}`,
    {
      method: "PUT",
      body: JSON.stringify(payload)
    }
  );
  return capabilityTemplateSchema.parse(raw);
}

export async function deleteCapability(capabilityId: string) {
  await apiRequest<void>(`/api/automation/capabilities/${encodeURIComponent(capabilityId)}`, {
    method: "DELETE"
  });
}

const patchCapabilityPayloadSchema = z
  .object({
    state: z.string().min(1).optional(),
    enabled: z.boolean().optional()
  })
  .refine((value) => value.state !== undefined || value.enabled !== undefined, {
    message: "Either state or enabled must be provided"
  });

export async function fetchDeviceCapabilities(deviceId: string) {
  const raw = await apiRequest<unknown>(
    `/api/devices/${encodeURIComponent(deviceId)}/capabilities`
  );
  return capabilityUIModelsSchema.parse(raw);
}

export async function patchDeviceCapability(
  deviceId: string,
  capabilityId: string,
  payload: z.infer<typeof patchCapabilityPayloadSchema>
) {
  const body = patchCapabilityPayloadSchema.parse(payload);
  const raw = await apiRequest<unknown>(
    `/api/devices/${encodeURIComponent(deviceId)}/capabilities/${encodeURIComponent(capabilityId)}`,
    {
      method: "PATCH",
      body: JSON.stringify(body)
    }
  );
  return setStateResultSchema.parse(raw);
}

export async function fetchCapabilityAssignments(capabilityId: string) {
  const raw = await apiRequest<unknown>(
    `/api/automation/capabilities/${encodeURIComponent(capabilityId)}/devices`
  );
  return capabilityAssignmentsSchema.parse(raw);
}

export async function patchCapabilityDevice(
  capabilityId: string,
  deviceId: string,
  payload: z.infer<typeof patchCapabilityPayloadSchema>
) {
  const body = patchCapabilityPayloadSchema.parse(payload);
  const raw = await apiRequest<unknown>(
    `/api/automation/capabilities/${encodeURIComponent(capabilityId)}/devices/${encodeURIComponent(deviceId)}`,
    {
      method: "PATCH",
      body: JSON.stringify(body)
    }
  );
  return setStateResultSchema.parse(raw);
}
