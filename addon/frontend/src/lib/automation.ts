import type {
  CapabilityTemplate,
  ControlType,
  ActionType,
  ActionParamField,
  CapabilityScope
} from "@/types/automation";

export function createEmptyCapabilityTemplate(): CapabilityTemplate {
  return {
    id: "",
    label: "",
    description: "",
    category: "General",
    scope: "device",
    control: {
      type: "switch",
      options: [
        { value: "on", label: "On" },
        { value: "off", label: "Off" }
      ]
    },
    states: {
      on: { label: "On", actions_on_enter: [] },
      off: { label: "Off", actions_on_enter: [] }
    },
    default_state: "off",
    ha_expose: {
      enabled: false,
      entity_type: "switch",
      entity_suffix: "",
      name_template: "{{device.name}} capability"
    }
  };
}

export function buildStateMapFromOptions(
  options: Array<{ value: string; label: string }>,
  currentStates: CapabilityTemplate["states"]
): CapabilityTemplate["states"] {
  const result: CapabilityTemplate["states"] = {};
  for (const option of options) {
    const previous = currentStates[option.value];
    result[option.value] = previous ?? {
      label: option.label,
      actions_on_enter: []
    };
  }
  return result;
}

export function normalizeControl(
  controlType: ControlType,
  existingOptions: Array<{ value: string; label: string }>
) {
  if (controlType === "switch") {
    return [
      { value: "on", label: "On" },
      { value: "off", label: "Off" }
    ];
  }

  const unique = new Map<string, { value: string; label: string }>();
  for (const option of existingOptions) {
    const value = option.value.trim();
    if (!value || unique.has(value)) {
      continue;
    }
    unique.set(value, { value, label: option.label || value });
  }
  if (unique.size < 2) {
    unique.set("allow", { value: "allow", label: "Allow" });
    unique.set("deny", { value: "deny", label: "Deny" });
  }
  return Array.from(unique.values());
}

export function countActions(template: CapabilityTemplate) {
  return Object.values(template.states).reduce(
    (total, state) => total + state.actions_on_enter.length,
    0
  );
}

export function categoriesFromCapabilities(capabilities: CapabilityTemplate[]) {
  const categories = new Set<string>();
  for (const capability of capabilities) {
    if (capability.category.trim()) {
      categories.add(capability.category.trim());
    }
  }
  return Array.from(categories).sort((a, b) => a.localeCompare(b));
}

export function defaultValueForActionField(field: ActionParamField): unknown {
  if (field.kind === "bool") {
    return false;
  }
  if (field.kind === "enum") {
    return field.options?.[0] ?? "";
  }
  return "";
}

export function resolveVisibleFields(fields: ActionParamField[], params: Record<string, unknown>) {
  return fields.filter((field) => {
    if (!field.visible_if) {
      return true;
    }
    return params[field.visible_if.key] === field.visible_if.equals;
  });
}

export function findActionType(actionTypes: ActionType[], typeId: string) {
  return actionTypes.find((item) => item.id === typeId);
}

export function scopeParamSchema(fields: ActionParamField[], scope: CapabilityScope) {
  if (scope !== "global") {
    return fields;
  }
  return fields.map((field) => {
    if (field.kind !== "enum") {
      return field;
    }
    return {
      ...field,
      options: (field.options ?? []).filter((option) => !isDeviceRef(option))
    };
  });
}

export function hasScopeViolationForGlobal(value: unknown): boolean {
  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    return normalized.startsWith("device.") || normalized.includes("{{device.");
  }
  if (Array.isArray(value)) {
    return value.some((item) => hasScopeViolationForGlobal(item));
  }
  if (value && typeof value === "object") {
    return Object.values(value as Record<string, unknown>).some((item) =>
      hasScopeViolationForGlobal(item)
    );
  }
  return false;
}

function isDeviceRef(value: string) {
  return value.trim().toLowerCase().startsWith("device.");
}
