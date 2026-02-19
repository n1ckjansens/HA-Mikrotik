import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import {
  defaultValueForActionField,
  resolveVisibleFields
} from "@/lib/automation";
import type { ActionInstance, ActionType } from "@/types/automation";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  actionTypes: ActionType[];
  onSave: (action: ActionInstance) => void;
};

function createActionID() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `action_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`;
}

export function ActionInstanceDialog({ open, onOpenChange, actionTypes, onSave }: Props) {
  const [selectedTypeId, setSelectedTypeId] = useState("");
  const [params, setParams] = useState<Record<string, unknown>>({});

  const selectedType = useMemo(
    () => actionTypes.find((item) => item.id === selectedTypeId),
    [actionTypes, selectedTypeId]
  );

  const visibleFields = useMemo(
    () => resolveVisibleFields(selectedType?.param_schema ?? [], params),
    [selectedType, params]
  );

  const resetDialog = () => {
    setSelectedTypeId("");
    setParams({});
  };

  const handleTypeChange = (typeId: string) => {
    setSelectedTypeId(typeId);
    const type = actionTypes.find((item) => item.id === typeId);
    if (!type) {
      setParams({});
      return;
    }
    const defaults: Record<string, unknown> = {};
    for (const field of type.param_schema) {
      defaults[field.key] = defaultValueForActionField(field);
    }
    setParams(defaults);
  };

  const handleSave = () => {
    if (!selectedType) {
      return;
    }
    onSave({
      id: createActionID(),
      type_id: selectedType.id,
      params
    });
    onOpenChange(false);
    resetDialog();
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        onOpenChange(nextOpen);
        if (!nextOpen) {
          resetDialog();
        }
      }}
    >
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add action</DialogTitle>
          <DialogDescription>
            Choose an action type and provide parameter values for this state transition.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label>Action type</Label>
            <Select value={selectedTypeId} onValueChange={handleTypeChange}>
              <SelectTrigger>
                <SelectValue placeholder="Choose action type" />
              </SelectTrigger>
              <SelectContent>
                {actionTypes.map((item) => (
                  <SelectItem key={item.id} value={item.id}>
                    {item.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedType ? (
              <p className="text-xs text-muted-foreground">{selectedType.description}</p>
            ) : null}
          </div>

          {selectedType ? (
            <div className="space-y-3">
              {visibleFields.map((field) => (
                <div key={field.key} className="space-y-2">
                  <Label>
                    {field.label}
                    {field.required ? " *" : ""}
                  </Label>

                  {field.kind === "string" ? (
                    <Input
                      value={(params[field.key] as string) ?? ""}
                      onChange={(event) =>
                        setParams((current) => ({
                          ...current,
                          [field.key]: event.target.value
                        }))
                      }
                    />
                  ) : null}

                  {field.kind === "enum" ? (
                    <Select
                      value={String(params[field.key] ?? "")}
                      onValueChange={(value) =>
                        setParams((current) => ({
                          ...current,
                          [field.key]: value
                        }))
                      }
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
                        checked={Boolean(params[field.key])}
                        onCheckedChange={(checked) =>
                          setParams((current) => ({
                            ...current,
                            [field.key]: checked
                          }))
                        }
                      />
                      <span className="text-sm text-muted-foreground">
                        {params[field.key] ? "True" : "False"}
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
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={!selectedType}>
            Save action
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
