import { toast } from "sonner";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { useDeviceCapabilities, useUpdateDeviceCapability } from "@/hooks/useDeviceCapabilities";
import type { CapabilityUIModel } from "@/types/automation";

type RowProps = {
  deviceId: string;
  capability: CapabilityUIModel;
};

function CapabilityControlRow({ deviceId, capability }: RowProps) {
  const updateMutation = useUpdateDeviceCapability();

  const updateState = async (nextState: string) => {
    try {
      const result = await updateMutation.mutateAsync({
        deviceId,
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

  return (
    <div className="rounded-md border p-3">
      <div className="mb-2">
        <p className="text-sm font-medium">{capability.label}</p>
        <p className="text-xs text-muted-foreground">{capability.description || capability.id}</p>
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

export function DeviceControlsSection({ deviceId }: { deviceId: string }) {
  const capabilitiesQuery = useDeviceCapabilities(deviceId);

  if (capabilitiesQuery.isPending) {
    return (
      <div className="space-y-2">
        <Skeleton className="h-14" />
        <Skeleton className="h-14" />
      </div>
    );
  }

  if (!capabilitiesQuery.data || capabilitiesQuery.data.length === 0) {
    return <p className="text-sm text-muted-foreground">No controls available for this device.</p>;
  }

  return (
    <div className="space-y-2">
      {capabilitiesQuery.data.map((capability) => (
        <CapabilityControlRow
          key={capability.id}
          deviceId={deviceId}
          capability={capability}
        />
      ))}
    </div>
  );
}
