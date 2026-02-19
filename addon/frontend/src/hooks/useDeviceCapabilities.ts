import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  fetchDeviceCapabilities,
  patchDeviceCapability
} from "@/api/automation";
import { queryKeys } from "@/lib/query-keys";
import type { SetStateResult } from "@/types/automation";

export function useDeviceCapabilities(deviceId: string | null) {
  return useQuery({
    queryKey: deviceId
      ? queryKeys.deviceCapabilities(deviceId)
      : ["devices", "capabilities", "none"],
    queryFn: () => fetchDeviceCapabilities(deviceId ?? ""),
    enabled: Boolean(deviceId),
    staleTime: 5000,
    refetchInterval: 5000,
    refetchIntervalInBackground: true
  });
}

type UpdateInput = {
  deviceId: string;
  capabilityId: string;
  state?: string;
  enabled?: boolean;
};

export function useUpdateDeviceCapability() {
  const client = useQueryClient();

  return useMutation<SetStateResult, Error, UpdateInput>({
    mutationFn: ({ deviceId, capabilityId, state, enabled }) =>
      patchDeviceCapability(deviceId, capabilityId, { state, enabled }),
    onSuccess: async (_result, vars) => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.deviceCapabilities(vars.deviceId) }),
        client.invalidateQueries({ queryKey: queryKeys.deviceDetail(vars.deviceId) })
      ]);
    }
  });
}
