import { z } from "zod";

export const actionParamFieldKindSchema = z.enum(["string", "enum", "bool"]);
export const capabilityScopeSchema = z.enum(["device", "global"]);

export const visibleIfConditionSchema = z.object({
  key: z.string(),
  equals: z.string()
});

export const actionParamFieldSchema = z.object({
  key: z.string(),
  label: z.string(),
  kind: actionParamFieldKindSchema,
  required: z.boolean(),
  description: z.string().optional(),
  options: z.array(z.string()).optional().default([]),
  visible_if: visibleIfConditionSchema.optional()
});

export const actionTypeSchema = z.object({
  id: z.string(),
  label: z.string(),
  description: z.string(),
  param_schema: z.array(actionParamFieldSchema)
});

export const stateSourceTypeSchema = z.object({
  id: z.string(),
  label: z.string(),
  description: z.string(),
  output_type: z.string(),
  param_schema: z.array(actionParamFieldSchema)
});

export const actionInstanceSchema = z.object({
  id: z.string(),
  type_id: z.string(),
  params: z.record(z.unknown()).default({})
});

export const controlTypeSchema = z.enum(["switch", "select"]);

export const capabilityControlOptionSchema = z.object({
  value: z.string(),
  label: z.string()
});

export const capabilityControlSchema = z.object({
  type: controlTypeSchema,
  options: z.array(capabilityControlOptionSchema)
});

export const capabilityStateConfigSchema = z.object({
  label: z.string(),
  actions_on_enter: z.array(actionInstanceSchema)
});

export const capabilitySyncSourceSchema = z.object({
  type_id: z.string(),
  params: z.record(z.unknown()).default({})
});

export const capabilitySyncMappingSchema = z.object({
  when_true: z.string(),
  when_false: z.string()
});

export const capabilitySyncConfigSchema = z.object({
  enabled: z.boolean(),
  source: capabilitySyncSourceSchema,
  mapping: capabilitySyncMappingSchema,
  mode: z.enum(["external_truth", "internal_truth"]),
  trigger_actions_on_sync: z.boolean()
});

export const haExposeSchema = z.object({
  enabled: z.boolean(),
  entity_type: z.string(),
  entity_suffix: z.string(),
  name_template: z.string()
});

export const capabilityTemplateSchema = z.object({
  id: z.string(),
  label: z.string(),
  description: z.string(),
  category: z.string(),
  scope: capabilityScopeSchema.optional().default("device"),
  control: capabilityControlSchema,
  states: z.record(capabilityStateConfigSchema),
  default_state: z.string(),
  sync: capabilitySyncConfigSchema.optional(),
  ha_expose: haExposeSchema
});

export const capabilityUIModelSchema = z.object({
  id: z.string(),
  label: z.string(),
  description: z.string(),
  control: capabilityControlSchema,
  state: z.string(),
  enabled: z.boolean()
});

export const capabilityDeviceAssignmentSchema = z.object({
  device_id: z.string(),
  device_name: z.string(),
  device_ip: z.string().optional(),
  online: z.boolean(),
  enabled: z.boolean(),
  state: z.string()
});

export const actionExecutionWarningSchema = z.object({
  action_id: z.string().optional(),
  type_id: z.string(),
  message: z.string()
});

export const setStateResultSchema = z.object({
  ok: z.boolean(),
  warnings: z.array(actionExecutionWarningSchema).optional().default([])
});

export type ActionParamField = z.infer<typeof actionParamFieldSchema>;
export type ActionType = z.infer<typeof actionTypeSchema>;
export type StateSourceType = z.infer<typeof stateSourceTypeSchema>;
export type CapabilityScope = z.infer<typeof capabilityScopeSchema>;
export type ActionInstance = z.infer<typeof actionInstanceSchema>;
export type CapabilityStateConfig = z.infer<typeof capabilityStateConfigSchema>;
export type CapabilitySyncConfig = z.infer<typeof capabilitySyncConfigSchema>;
export type CapabilityTemplate = z.infer<typeof capabilityTemplateSchema>;
export type CapabilityUIModel = z.infer<typeof capabilityUIModelSchema>;
export type CapabilityDeviceAssignment = z.infer<typeof capabilityDeviceAssignmentSchema>;
export type SetStateResult = z.infer<typeof setStateResultSchema>;
export type ControlType = z.infer<typeof controlTypeSchema>;
