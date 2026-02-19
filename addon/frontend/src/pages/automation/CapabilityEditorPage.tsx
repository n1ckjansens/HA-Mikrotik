import { useEffect, useMemo, useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";

import { ActionInstanceDialog } from "@/components/automation/ActionInstanceDialog";
import { CapabilityStateCard } from "@/components/automation/CapabilityStateCard";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { useActionTypes } from "@/hooks/useActionTypes";
import { useCapabilityEditor } from "@/hooks/useCapabilityEditor";
import { useStateSourceTypes } from "@/hooks/useStateSourceTypes";
import {
  buildStateMapFromOptions,
  createEmptyCapabilityTemplate,
  defaultValueForActionField,
  hasScopeViolationForGlobal,
  normalizeControl,
  resolveVisibleFields,
  scopeParamSchema
} from "@/lib/automation";
import type { CapabilityTemplate, ControlType, StateSourceType } from "@/types/automation";

const categoryOptions = [
  "General",
  "Routing",
  "Security",
  "Parental",
  "Connectivity",
  "QoS"
];

function defaultSyncParams(sourceType: StateSourceType | undefined) {
  const defaults: Record<string, unknown> = {};
  for (const field of sourceType?.param_schema ?? []) {
    defaults[field.key] = defaultValueForActionField(field);
  }
  return defaults;
}

function defaultSyncConfig(
  template: CapabilityTemplate,
  sourceTypeId: string,
  sourceParams: Record<string, unknown>
): NonNullable<CapabilityTemplate["sync"]> {
  const whenTrue = template.control.options[0]?.value ?? template.default_state;
  const whenFalse = template.control.options[1]?.value ?? whenTrue;

  return {
    enabled: true,
    source: {
      type_id: sourceTypeId,
      params: sourceParams
    },
    mapping: {
      when_true: whenTrue,
      when_false: whenFalse
    },
    mode: "external_truth",
    trigger_actions_on_sync: false
  };
}

function applyScopeToStateSourceTypes(
  sourceTypes: StateSourceType[],
  scope: CapabilityTemplate["scope"]
) {
  return sourceTypes.map((item) => ({
    ...item,
    param_schema: scopeParamSchema(item.param_schema, scope)
  }));
}

function hasGlobalScopeParamViolations(template: CapabilityTemplate) {
  if (template.scope !== "global") {
    return false;
  }

  for (const state of Object.values(template.states)) {
    for (const action of state.actions_on_enter) {
      if (hasScopeViolationForGlobal(action.params)) {
        return true;
      }
    }
  }

  return hasScopeViolationForGlobal(template.sync?.source.params ?? {});
}

export function CapabilityEditorPage() {
  const navigate = useNavigate();
  const params = useParams();
  const capabilityId = params.id ?? "new";

  const actionTypesQuery = useActionTypes();
  const stateSourceTypesQuery = useStateSourceTypes();
  const editor = useCapabilityEditor(capabilityId);

  const [draft, setDraft] = useState<CapabilityTemplate>(createEmptyCapabilityTemplate());
  const [actionDialogStateId, setActionDialogStateId] = useState<string | null>(null);

  const title = editor.isNew ? "New capability" : `Edit ${capabilityId}`;
  const actionTypes = useMemo(
    () =>
      (actionTypesQuery.data ?? []).map((item) => ({
        ...item,
        param_schema: scopeParamSchema(item.param_schema, draft.scope)
      })),
    [actionTypesQuery.data, draft.scope]
  );

  const stateSourceTypes = useMemo(
    () => applyScopeToStateSourceTypes(stateSourceTypesQuery.data ?? [], draft.scope),
    [stateSourceTypesQuery.data, draft.scope]
  );

  useEffect(() => {
    if (!editor.isNew && editor.capabilityQuery.data) {
      setDraft(editor.capabilityQuery.data);
    }
  }, [editor.isNew, editor.capabilityQuery.data]);

  const stateOrder = useMemo(
    () => draft.control.options.map((option) => option.value),
    [draft.control.options]
  );

  const selectedSyncSourceType = useMemo(
    () => stateSourceTypes.find((item) => item.id === draft.sync?.source.type_id),
    [stateSourceTypes, draft.sync?.source.type_id]
  );

  const visibleSyncFields = useMemo(
    () => resolveVisibleFields(selectedSyncSourceType?.param_schema ?? [], draft.sync?.source.params ?? {}),
    [selectedSyncSourceType, draft.sync?.source.params]
  );

  const setControlType = (controlType: ControlType) => {
    setDraft((current) => {
      const options = normalizeControl(controlType, current.control.options);
      const states = buildStateMapFromOptions(options, current.states);
      const defaultState = states[current.default_state]
        ? current.default_state
        : options[0]?.value ?? current.default_state;
      return {
        ...current,
        control: {
          ...current.control,
          type: controlType,
          options
        },
        states,
        default_state: defaultState,
        ha_expose: {
          ...current.ha_expose,
          entity_type: controlType === "switch" ? "switch" : "select"
        }
      };
    });
  };

  const updateControlOption = (index: number, field: "value" | "label", value: string) => {
    setDraft((current) => {
      const options = current.control.options.map((option, optionIndex) =>
        optionIndex === index ? { ...option, [field]: value } : option
      );
      const normalizedOptions =
        current.control.type === "switch"
          ? [
              { value: "on", label: options[0]?.label || "On" },
              { value: "off", label: options[1]?.label || "Off" }
            ]
          : options;

      return {
        ...current,
        control: {
          ...current.control,
          options: normalizedOptions
        },
        states: buildStateMapFromOptions(normalizedOptions, current.states)
      };
    });
  };

  const addSelectOption = () => {
    setDraft((current) => {
      const nextIndex = current.control.options.length + 1;
      const nextValue = `option_${nextIndex}`;
      const options = [...current.control.options, { value: nextValue, label: `Option ${nextIndex}` }];
      return {
        ...current,
        control: {
          ...current.control,
          options
        },
        states: buildStateMapFromOptions(options, current.states)
      };
    });
  };

  const removeSelectOption = (index: number) => {
    setDraft((current) => {
      if (current.control.type !== "select" || current.control.options.length <= 2) {
        return current;
      }
      const options = current.control.options.filter((_, optionIndex) => optionIndex !== index);
      const states = buildStateMapFromOptions(options, current.states);
      const defaultState = states[current.default_state]
        ? current.default_state
        : options[0]?.value ?? "";
      return {
        ...current,
        control: {
          ...current.control,
          options
        },
        states,
        default_state: defaultState
      };
    });
  };

  const updateStateLabel = (stateId: string, label: string) => {
    setDraft((current) => ({
      ...current,
      states: {
        ...current.states,
        [stateId]: {
          ...current.states[stateId],
          label
        }
      }
    }));
  };

  const addActionToState = (stateId: string) => {
    setActionDialogStateId(stateId);
  };

  const handleSaveAction = (stateId: string, action: CapabilityTemplate["states"][string]["actions_on_enter"][number]) => {
    setDraft((current) => ({
      ...current,
      states: {
        ...current.states,
        [stateId]: {
          ...current.states[stateId],
          actions_on_enter: [...current.states[stateId].actions_on_enter, action]
        }
      }
    }));
  };

  const removeActionFromState = (stateId: string, actionId: string) => {
    setDraft((current) => ({
      ...current,
      states: {
        ...current.states,
        [stateId]: {
          ...current.states[stateId],
          actions_on_enter: current.states[stateId].actions_on_enter.filter(
            (action) => action.id !== actionId
          )
        }
      }
    }));
  };

  const setSyncEnabled = (enabled: boolean) => {
    setDraft((current) => {
      const firstSourceType = stateSourceTypes[0];
      const currentSync = current.sync;

      if (!enabled) {
        if (!currentSync) {
          return current;
        }
        return {
          ...current,
          sync: {
            ...currentSync,
            enabled: false
          }
        };
      }

      if (currentSync) {
        let nextSync = { ...currentSync, enabled: true };
        if (!nextSync.source.type_id && firstSourceType) {
          nextSync = {
            ...nextSync,
            source: {
              type_id: firstSourceType.id,
              params: defaultSyncParams(firstSourceType)
            }
          };
        }
        return {
          ...current,
          sync: nextSync
        };
      }

      return {
        ...current,
        sync: defaultSyncConfig(
          current,
          firstSourceType?.id ?? "",
          defaultSyncParams(firstSourceType)
        )
      };
    });
  };

  const setSyncSourceType = (typeId: string) => {
    setDraft((current) => {
      if (!current.sync) {
        return current;
      }
      const sourceType = stateSourceTypes.find((item) => item.id === typeId);
      return {
        ...current,
        sync: {
          ...current.sync,
          source: {
            type_id: typeId,
            params: defaultSyncParams(sourceType)
          }
        }
      };
    });
  };

  const setSyncParam = (key: string, value: unknown) => {
    setDraft((current) => {
      if (!current.sync) {
        return current;
      }
      return {
        ...current,
        sync: {
          ...current.sync,
          source: {
            ...current.sync.source,
            params: {
              ...(current.sync.source.params ?? {}),
              [key]: value
            }
          }
        }
      };
    });
  };

  useEffect(() => {
    if (!draft.sync) {
      return;
    }

    const allowed = new Set(draft.control.options.map((option) => option.value));
    const fallbackTrue = draft.control.options[0]?.value ?? draft.default_state;
    const fallbackFalse = draft.control.options[1]?.value ?? fallbackTrue;
    const nextWhenTrue = allowed.has(draft.sync.mapping.when_true)
      ? draft.sync.mapping.when_true
      : fallbackTrue;
    const nextWhenFalse = allowed.has(draft.sync.mapping.when_false)
      ? draft.sync.mapping.when_false
      : fallbackFalse;

    if (
      nextWhenTrue === draft.sync.mapping.when_true &&
      nextWhenFalse === draft.sync.mapping.when_false
    ) {
      return;
    }

    setDraft((current) => {
      if (!current.sync) {
        return current;
      }
      return {
        ...current,
        sync: {
          ...current.sync,
          mapping: {
            ...current.sync.mapping,
            when_true: nextWhenTrue,
            when_false: nextWhenFalse
          }
        }
      };
    });
  }, [draft.control.options, draft.default_state, draft.sync]);

  const handleSave = async () => {
    if (hasGlobalScopeParamViolations(draft)) {
      toast.error("Global scope does not allow device placeholders in actions or sync params.");
      return;
    }
    try {
      await editor.saveCapability(draft);
      toast.success("Capability saved");
      navigate(`/automation/capabilities/${encodeURIComponent(draft.id)}`);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Save failed";
      toast.error(message);
    }
  };

  if (!editor.isNew && editor.capabilityQuery.isPending) {
    return <p className="text-sm text-muted-foreground">Loading capability...</p>;
  }

  return (
    <div className="space-y-4 pb-24 md:pb-6">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{title}</h1>
          <p className="text-sm text-muted-foreground">
            Configure states, actions, and Home Assistant exposure for this capability.
          </p>
        </div>
        <div className="hidden items-center gap-2 md:flex">
          <Button variant="outline" onClick={() => navigate("/automation/capabilities")}>
            Cancel
          </Button>
          <Button onClick={() => void handleSave()} disabled={editor.isSaving}>
            Save
          </Button>
        </div>
      </header>

      <Tabs defaultValue="basic" className="space-y-4">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="basic">Basic</TabsTrigger>
          <TabsTrigger value="control">Control</TabsTrigger>
          <TabsTrigger value="states">States & Actions</TabsTrigger>
          <TabsTrigger value="sync">Sync</TabsTrigger>
          <TabsTrigger value="ha">Home Assistant</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Basic settings</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="space-y-2">
                <Label>ID</Label>
                <Input
                  value={draft.id}
                  onChange={(event) => setDraft((current) => ({ ...current, id: event.target.value }))}
                  placeholder="routing.vpn"
                  disabled={!editor.isNew}
                />
              </div>

              <div className="space-y-2">
                <Label>Label</Label>
                <Input
                  value={draft.label}
                  onChange={(event) => setDraft((current) => ({ ...current, label: event.target.value }))}
                  placeholder="VPN routing"
                />
              </div>

              <div className="space-y-2">
                <Label>Description</Label>
                <Textarea
                  value={draft.description}
                  onChange={(event) =>
                    setDraft((current) => ({ ...current, description: event.target.value }))
                  }
                />
              </div>

              <div className="space-y-2">
                <Label>Category</Label>
                <Select
                  value={draft.category || "General"}
                  onValueChange={(value) =>
                    setDraft((current) => ({
                      ...current,
                      category: value
                    }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Category" />
                  </SelectTrigger>
                  <SelectContent>
                    {categoryOptions.map((item) => (
                      <SelectItem key={item} value={item}>
                        {item}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Scope</Label>
                <RadioGroup
                  value={draft.scope}
                  onValueChange={(value) => {
                    if (value === "device" || value === "global") {
                      setDraft((current) => ({
                        ...current,
                        scope: value
                      }));
                    }
                  }}
                >
                  <div className="flex items-center gap-2">
                    <RadioGroupItem value="device" id="scope-device" />
                    <Label htmlFor="scope-device">Per-device</Label>
                  </div>
                  <div className="flex items-center gap-2">
                    <RadioGroupItem value="global" id="scope-global" />
                    <Label htmlFor="scope-global">Global</Label>
                  </div>
                </RadioGroup>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="control" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Control type</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <RadioGroup
                value={draft.control.type}
                onValueChange={(value) => {
                  if (value === "switch" || value === "select") {
                    setControlType(value);
                  }
                }}
              >
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="switch" id="control-switch" />
                  <Label htmlFor="control-switch">Switch</Label>
                </div>
                <div className="flex items-center gap-2">
                  <RadioGroupItem value="select" id="control-select" />
                  <Label htmlFor="control-select">Select</Label>
                </div>
              </RadioGroup>

              <div className="space-y-2">
                <Label>Options</Label>
                <div className="space-y-2">
                  {draft.control.options.map((option, index) => (
                    <div key={option.value + index} className="grid gap-2 sm:grid-cols-[1fr_1fr_auto]">
                      <Input
                        value={option.value}
                        onChange={(event) =>
                          updateControlOption(index, "value", event.target.value)
                        }
                        disabled={draft.control.type === "switch"}
                      />
                      <Input
                        value={option.label}
                        onChange={(event) =>
                          updateControlOption(index, "label", event.target.value)
                        }
                      />
                      {draft.control.type === "select" ? (
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => removeSelectOption(index)}
                          disabled={draft.control.options.length <= 2}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      ) : null}
                    </div>
                  ))}
                </div>

                {draft.control.type === "select" ? (
                  <Button variant="outline" size="sm" onClick={addSelectOption}>
                    <Plus className="mr-2 h-4 w-4" />
                    Add option
                  </Button>
                ) : null}
              </div>

              <div className="space-y-2">
                <Label>Default state</Label>
                <Select
                  value={draft.default_state}
                  onValueChange={(value) =>
                    setDraft((current) => ({
                      ...current,
                      default_state: value
                    }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select default" />
                  </SelectTrigger>
                  <SelectContent>
                    {draft.control.options.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="states" className="space-y-4">
          {stateOrder.map((stateId) => {
            const state = draft.states[stateId];
            if (!state) {
              return null;
            }
            return (
              <CapabilityStateCard
                key={stateId}
                stateId={stateId}
                stateConfig={state}
                actionTypes={actionTypes}
                onStateLabelChange={updateStateLabel}
                onAddAction={addActionToState}
                onRemoveAction={removeActionFromState}
              />
            );
          })}
        </TabsContent>

        <TabsContent value="sync" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">External Sync</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between gap-2 rounded-md border p-3">
                <div>
                  <p className="text-sm font-medium">Enable state synchronization</p>
                  <p className="text-xs text-muted-foreground">
                    Read external state from a StateSource and map it to capability state.
                  </p>
                </div>
                <Switch
                  checked={Boolean(draft.sync?.enabled)}
                  onCheckedChange={setSyncEnabled}
                />
              </div>

              {draft.sync?.enabled ? (
                <>
                  <div className="space-y-2">
                    <Label>State source type</Label>
                    <Select
                      value={draft.sync.source.type_id}
                      onValueChange={setSyncSourceType}
                      disabled={stateSourceTypes.length === 0}
                    >
                      <SelectTrigger>
                        <SelectValue
                          placeholder={
                            stateSourceTypes.length === 0
                              ? "No state sources available"
                              : "Choose state source type"
                          }
                        />
                      </SelectTrigger>
                      <SelectContent>
                        {stateSourceTypes.map((item) => (
                          <SelectItem key={item.id} value={item.id}>
                            {item.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    {selectedSyncSourceType ? (
                      <p className="text-xs text-muted-foreground">
                        {selectedSyncSourceType.description} (output: {selectedSyncSourceType.output_type})
                      </p>
                    ) : null}
                  </div>

                  {visibleSyncFields.length > 0 ? (
                    <div className="space-y-3 rounded-md border p-3">
                      <p className="text-sm font-medium">State source params</p>
                      {visibleSyncFields.map((field) => (
                        <div key={field.key} className="space-y-2">
                          <Label>
                            {field.label}
                            {field.required ? " *" : ""}
                          </Label>

                          {field.kind === "string" ? (
                            <Input
                              value={String(draft.sync?.source.params?.[field.key] ?? "")}
                              onChange={(event) => setSyncParam(field.key, event.target.value)}
                            />
                          ) : null}

                          {field.kind === "enum" ? (
                            <Select
                              value={String(draft.sync?.source.params?.[field.key] ?? "")}
                              onValueChange={(value) => setSyncParam(field.key, value)}
                            >
                              <SelectTrigger>
                                <SelectValue placeholder="Select value" />
                              </SelectTrigger>
                              <SelectContent>
                                {(field.options ?? []).map((option) => (
                                  <SelectItem key={option} value={option}>
                                    {option}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          ) : null}

                          {field.kind === "bool" ? (
                            <div className="flex items-center gap-2">
                              <Switch
                                checked={Boolean(draft.sync?.source.params?.[field.key])}
                                onCheckedChange={(checked) => setSyncParam(field.key, checked)}
                              />
                              <span className="text-sm text-muted-foreground">
                                {draft.sync?.source.params?.[field.key] ? "True" : "False"}
                              </span>
                            </div>
                          ) : null}

                          {field.description ? (
                            <p className="text-xs text-muted-foreground">{field.description}</p>
                          ) : null}
                        </div>
                      ))}
                    </div>
                  ) : null}

                  <div className="grid gap-3 sm:grid-cols-2">
                    <div className="space-y-2">
                      <Label>Mapping when true</Label>
                      <Select
                        value={draft.sync.mapping.when_true}
                        onValueChange={(value) =>
                          setDraft((current) =>
                            current.sync
                              ? {
                                  ...current,
                                  sync: {
                                    ...current.sync,
                                    mapping: { ...current.sync.mapping, when_true: value }
                                  }
                                }
                              : current
                          )
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Select state" />
                        </SelectTrigger>
                        <SelectContent>
                          {draft.control.options.map((option) => (
                            <SelectItem key={option.value} value={option.value}>
                              {option.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label>Mapping when false</Label>
                      <Select
                        value={draft.sync.mapping.when_false}
                        onValueChange={(value) =>
                          setDraft((current) =>
                            current.sync
                              ? {
                                  ...current,
                                  sync: {
                                    ...current.sync,
                                    mapping: { ...current.sync.mapping, when_false: value }
                                  }
                                }
                              : current
                          )
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Select state" />
                        </SelectTrigger>
                        <SelectContent>
                          {draft.control.options.map((option) => (
                            <SelectItem key={option.value} value={option.value}>
                              {option.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label>Sync mode</Label>
                    <Select
                      value={draft.sync.mode}
                      onValueChange={(value) =>
                        setDraft((current) =>
                          current.sync
                            ? {
                                ...current,
                                sync: {
                                  ...current.sync,
                                  mode: value === "internal_truth" ? "internal_truth" : "external_truth"
                                }
                              }
                            : current
                        )
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="external_truth">external_truth</SelectItem>
                        <SelectItem value="internal_truth">internal_truth</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="flex items-center justify-between gap-2 rounded-md border p-3">
                    <div>
                      <p className="text-sm font-medium">Trigger actions on sync</p>
                      <p className="text-xs text-muted-foreground">
                        If enabled, sync uses state transition flow and executes ActionsOnEnter.
                      </p>
                    </div>
                    <Switch
                      checked={draft.sync.trigger_actions_on_sync}
                      onCheckedChange={(checked) =>
                        setDraft((current) =>
                          current.sync
                            ? {
                                ...current,
                                sync: {
                                  ...current.sync,
                                  trigger_actions_on_sync: checked
                                }
                              }
                            : current
                        )
                      }
                    />
                  </div>
                </>
              ) : null}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="ha" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Home Assistant exposure</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between gap-2 rounded-md border p-3">
                <div>
                  <p className="text-sm font-medium">Expose to Home Assistant</p>
                  <p className="text-xs text-muted-foreground">
                    Stores entity metadata in capability template for future sync.
                  </p>
                </div>
                <Switch
                  checked={draft.ha_expose.enabled}
                  onCheckedChange={(checked) =>
                    setDraft((current) => ({
                      ...current,
                      ha_expose: {
                        ...current.ha_expose,
                        enabled: checked
                      }
                    }))
                  }
                />
              </div>

              {draft.ha_expose.enabled ? (
                <>
                  <div className="space-y-2">
                    <Label>Entity type</Label>
                    <Select
                      value={draft.ha_expose.entity_type || (draft.control.type === "switch" ? "switch" : "select")}
                      onValueChange={(value) =>
                        setDraft((current) => ({
                          ...current,
                          ha_expose: {
                            ...current.ha_expose,
                            entity_type: value
                          }
                        }))
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="switch">switch</SelectItem>
                        <SelectItem value="select">select</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label>Entity suffix</Label>
                    <Input
                      value={draft.ha_expose.entity_suffix}
                      onChange={(event) =>
                        setDraft((current) => ({
                          ...current,
                          ha_expose: {
                            ...current.ha_expose,
                            entity_suffix: event.target.value
                          }
                        }))
                      }
                      placeholder="vpn"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label>Name template</Label>
                    <Input
                      value={draft.ha_expose.name_template}
                      onChange={(event) =>
                        setDraft((current) => ({
                          ...current,
                          ha_expose: {
                            ...current.ha_expose,
                            name_template: event.target.value
                          }
                        }))
                      }
                      placeholder="{{device.name}} VPN"
                    />
                    <p className="text-xs text-muted-foreground">
                      Supported placeholders include <code>{"{{device.name}}"}</code>.
                    </p>
                  </div>
                </>
              ) : null}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <ActionInstanceDialog
        open={actionDialogStateId !== null}
        onOpenChange={(open) => {
          if (!open) {
            setActionDialogStateId(null);
          }
        }}
        actionTypes={actionTypes}
        scope={draft.scope}
        onSave={(action) => {
          if (!actionDialogStateId) {
            return;
          }
          handleSaveAction(actionDialogStateId, action);
        }}
      />

      <footer className="fixed inset-x-0 bottom-0 z-20 border-t bg-background/95 p-3 backdrop-blur md:hidden">
        <div className="mx-auto flex max-w-6xl gap-2">
          <Button variant="outline" className="flex-1" onClick={() => navigate("/automation/capabilities")}>
            Cancel
          </Button>
          <Button className="flex-1" onClick={() => void handleSave()} disabled={editor.isSaving}>
            Save
          </Button>
        </div>
      </footer>
    </div>
  );
}
