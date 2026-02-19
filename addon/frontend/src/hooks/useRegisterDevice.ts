import { useMutation, useQueryClient, type QueryKey } from "@tanstack/react-query";

import {
  patchDevice,
  refreshDevices,
  registerDevice,
  type RegisterDeviceInput
} from "@/api/devices";
import { queryKeys } from "@/lib/query-keys";
import type { Device } from "@/types/device";

type DeviceMutationVars = {
  device: Device;
  input: RegisterDeviceInput;
};

type OptimisticContext = {
  devicesSnapshots: Array<[QueryKey, unknown]>;
  detailSnapshot: Device | undefined;
};

function normalizeInput(input: RegisterDeviceInput): RegisterDeviceInput {
  const next: RegisterDeviceInput = {};

  const name = input.name?.trim();
  if (name) {
    next.name = name;
  }

  const icon = input.icon?.trim();
  if (icon) {
    next.icon = icon;
  }

  if (typeof input.comment === "string") {
    next.comment = input.comment.trim();
  }

  return next;
}

function buildOptimisticDevice(
  device: Device,
  input: RegisterDeviceInput,
  forceRegistered: boolean
): Device {
  return {
    ...device,
    name: input.name ?? device.name,
    icon: input.icon ?? device.icon ?? null,
    comment: input.comment ?? device.comment ?? null,
    status: forceRegistered ? "registered" : device.status,
    updated_at: new Date().toISOString()
  };
}

function isDeviceList(value: unknown): value is Device[] {
  return Array.isArray(value);
}

function applyDeviceUpdate(
  client: ReturnType<typeof useQueryClient>,
  device: Device,
  input: RegisterDeviceInput,
  forceRegistered: boolean
): OptimisticContext {
  const devicesSnapshots = client.getQueriesData<unknown>({
    queryKey: ["devices"]
  });
  const detailSnapshot = client.getQueryData<Device>(queryKeys.deviceDetail(device.mac));
  const optimistic = buildOptimisticDevice(device, input, forceRegistered);

  for (const [queryKey] of devicesSnapshots) {
    client.setQueryData<unknown>(queryKey, (current: unknown) => {
      if (!isDeviceList(current)) {
        return current;
      }
      return current.map((item) =>
        item.mac === device.mac ? { ...item, ...optimistic } : item
      );
    });
  }

  client.setQueryData<Device>(queryKeys.deviceDetail(device.mac), (current) => ({
    ...(current ?? device),
    ...optimistic
  }));

  return { devicesSnapshots, detailSnapshot };
}

function rollbackDeviceUpdate(
  client: ReturnType<typeof useQueryClient>,
  context: OptimisticContext | undefined,
  mac: string
) {
  if (!context) {
    return;
  }

  for (const [queryKey, data] of context.devicesSnapshots) {
    client.setQueryData(queryKey, data);
  }
  client.setQueryData(queryKeys.deviceDetail(mac), context.detailSnapshot);
}

export function useRegisterDevice() {
  const client = useQueryClient();

  const refreshMutation = useMutation({
    mutationFn: () => refreshDevices(),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ["devices"] });
    }
  });

  const registerMutation = useMutation({
    mutationFn: ({ device, input }: DeviceMutationVars) => registerDevice(device.mac, input),
    onMutate: async ({ device, input }): Promise<OptimisticContext> => {
      await client.cancelQueries({ queryKey: ["devices"] });
      await client.cancelQueries({ queryKey: queryKeys.deviceDetail(device.mac) });
      return applyDeviceUpdate(client, device, input, true);
    },
    onError: (_error, variables, context) => {
      rollbackDeviceUpdate(client, context, variables.device.mac);
    },
    onSettled: async (_data, _error, variables) => {
      await client.invalidateQueries({ queryKey: ["devices"] });
      await client.invalidateQueries({ queryKey: queryKeys.deviceDetail(variables.device.mac) });
    }
  });

  const patchMutation = useMutation({
    mutationFn: ({ device, input }: DeviceMutationVars) => patchDevice(device.mac, input),
    onMutate: async ({ device, input }): Promise<OptimisticContext> => {
      await client.cancelQueries({ queryKey: ["devices"] });
      await client.cancelQueries({ queryKey: queryKeys.deviceDetail(device.mac) });
      return applyDeviceUpdate(client, device, input, false);
    },
    onError: (_error, variables, context) => {
      rollbackDeviceUpdate(client, context, variables.device.mac);
    },
    onSettled: async (_data, _error, variables) => {
      await client.invalidateQueries({ queryKey: ["devices"] });
      await client.invalidateQueries({ queryKey: queryKeys.deviceDetail(variables.device.mac) });
    }
  });

  const saveDevice = async (device: Device, input: RegisterDeviceInput) => {
    const payload = normalizeInput(input);

    if (device.status === "new") {
      await registerMutation.mutateAsync({ device, input: payload });
      return;
    }

    if (Object.keys(payload).length === 0) {
      return;
    }

    await patchMutation.mutateAsync({ device, input: payload });
  };

  return {
    refreshMutation,
    registerMutation,
    patchMutation,
    isSaving: registerMutation.isPending || patchMutation.isPending,
    saveDevice,
    async registerWithName(mac: string, name: string) {
      const trimmedName = name.trim();
      if (!trimmedName) {
        return;
      }

      const fallback: Device = {
        mac,
        name: trimmedName,
        vendor: "Unknown",
        status: "new",
        online: false,
        last_sources: [],
        updated_at: new Date().toISOString(),
        icon: null,
        comment: null,
        last_seen_at: null,
        connected_since_at: null,
        last_ip: null,
        last_subnet: null,
        first_seen_at: null,
        created_at: null
      };

      const device =
        client.getQueryData<Device>(queryKeys.deviceDetail(mac)) ??
        client
          .getQueriesData<unknown>({ queryKey: ["devices"] })
          .flatMap(([, items]) => (isDeviceList(items) ? items : []))
          .find((item) => item.mac === mac) ??
        fallback;

      await registerMutation.mutateAsync({
        device,
        input: { name: trimmedName }
      });
    }
  };
}
