import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import {
  useGlobalCapabilities,
  useUpdateGlobalCapability
} from "@/hooks/useGlobalCapabilities";
import type { CapabilityUIModel } from "@/types/automation";

function stateBadgeVariant(state: string) {
  if (state === "on" || state === "allow") {
    return "secondary" as const;
  }
  return "outline" as const;
}

function GlobalCapabilityRow({ capability }: { capability: CapabilityUIModel }) {
  const updateMutation = useUpdateGlobalCapability();

  const updateState = async (nextState: string) => {
    try {
      const result = await updateMutation.mutateAsync({
        capabilityId: capability.id,
        state: nextState
      });
      if ((result.warnings ?? []).length > 0) {
        toast.warning(result.warnings[0].message);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Capability update failed";
      toast.error(message);
    }
  };

  const updateEnabled = async (enabled: boolean) => {
    try {
      const result = await updateMutation.mutateAsync({
        capabilityId: capability.id,
        enabled
      });
      if ((result.warnings ?? []).length > 0) {
        toast.warning(result.warnings[0].message);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : "Capability update failed";
      toast.error(message);
    }
  };

  return (
    <div className="space-y-3 rounded-md border p-3">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-medium">{capability.label}</p>
          <p className="text-xs text-muted-foreground">{capability.id}</p>
          {capability.description ? (
            <p className="text-xs text-muted-foreground">{capability.description}</p>
          ) : null}
        </div>
        <div className="flex items-center gap-3">
          <Badge variant={stateBadgeVariant(capability.state)}>{capability.state}</Badge>
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">Enabled</span>
            <Switch
              checked={capability.enabled}
              disabled={updateMutation.isPending}
              onCheckedChange={(checked) => {
                void updateEnabled(checked);
              }}
            />
          </div>
        </div>
      </div>

      {capability.control.type === "switch" ? (
        <div className="flex items-center justify-between">
          <span className="text-sm">{capability.state === "on" ? "On" : "Off"}</span>
          <Switch
            checked={capability.state === "on"}
            disabled={!capability.enabled || updateMutation.isPending}
            onCheckedChange={(checked) => {
              void updateState(checked ? "on" : "off");
            }}
          />
        </div>
      ) : (
        <Select
          value={capability.state}
          disabled={!capability.enabled || updateMutation.isPending}
          onValueChange={(value) => {
            void updateState(value);
          }}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {capability.control.options.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </div>
  );
}

export function GlobalCapabilitiesPage() {
  const query = useGlobalCapabilities();

  return (
    <div className="space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Global</h1>
        <p className="text-sm text-muted-foreground">
          One-click global capabilities not bound to a specific device.
        </p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Global capabilities</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {query.isPending ? (
            <div className="space-y-2">
              <Skeleton className="h-20" />
              <Skeleton className="h-20" />
            </div>
          ) : null}

          {!query.isPending && (query.data ?? []).length === 0 ? (
            <p className="text-sm text-muted-foreground">No global capabilities configured.</p>
          ) : null}

          {(query.data ?? []).map((item) => (
            <GlobalCapabilityRow key={item.id} capability={item} />
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
