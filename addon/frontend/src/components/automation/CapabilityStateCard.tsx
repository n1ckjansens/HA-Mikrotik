import { Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import type { ActionInstance, ActionType, CapabilityStateConfig } from "@/types/automation";

type Props = {
  stateId: string;
  stateConfig: CapabilityStateConfig;
  actionTypes: ActionType[];
  onStateLabelChange: (stateId: string, label: string) => void;
  onAddAction: (stateId: string) => void;
  onRemoveAction: (stateId: string, actionId: string) => void;
};

function actionLabel(actionTypes: ActionType[], typeId: string) {
  return actionTypes.find((item) => item.id === typeId)?.label ?? typeId;
}

function paramsSummary(action: ActionInstance) {
  const entries = Object.entries(action.params ?? {});
  if (entries.length === 0) {
    return "No params";
  }
  return entries
    .slice(0, 3)
    .map(([key, value]) => `${key}: ${String(value)}`)
    .join(" · ");
}

export function CapabilityStateCard({
  stateId,
  stateConfig,
  actionTypes,
  onStateLabelChange,
  onAddAction,
  onRemoveAction
}: Props) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-3">
          <CardTitle className="text-base">State: {stateId}</CardTitle>
          <Badge variant="outline">{stateConfig.actions_on_enter.length} actions</Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">State label</label>
          <Input
            value={stateConfig.label}
            onChange={(event) => onStateLabelChange(stateId, event.target.value)}
            placeholder="State label"
          />
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between gap-2">
            <p className="text-sm font-medium">Actions on enter</p>
            <Button variant="outline" size="sm" onClick={() => onAddAction(stateId)}>
              Add action
            </Button>
          </div>

          {stateConfig.actions_on_enter.length === 0 ? (
            <p className="text-sm text-muted-foreground">No actions configured for this state.</p>
          ) : (
            <div className="space-y-2">
              {stateConfig.actions_on_enter.map((action) => (
                <div
                  key={action.id}
                  className="flex items-start justify-between gap-3 rounded-md border p-3"
                >
                  <div>
                    <p className="text-sm font-medium">{actionLabel(actionTypes, action.type_id)}</p>
                    <p className="text-xs text-muted-foreground">{action.type_id}</p>
                    <p className="text-xs text-muted-foreground">{paramsSummary(action)}</p>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => onRemoveAction(stateId, action.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
