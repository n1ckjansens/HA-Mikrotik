import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  fetchCapabilityAssignments,
  patchCapabilityDevice
} from "@/api/automation";
import { queryKeys } from "@/lib/query-keys";
import type { SetStateResult } from "@/types/automation";

type UpdateAssignmentInput = {
  capabilityId: string;
  deviceId: string;
  enabled?: boolean;
  state?: string;
};

export function useAssignments(capabilityId: string) {
  const client = useQueryClient();

  const assignmentsQuery = useQuery({
    queryKey: queryKeys.automationAssignments(capabilityId),
    queryFn: () => fetchCapabilityAssignments(capabilityId),
    enabled: capabilityId.trim().length > 0,
    staleTime: 5000
  });

  const updateAssignment = useMutation<SetStateResult, Error, UpdateAssignmentInput>({
    mutationFn: ({ capabilityId, deviceId, enabled, state }) =>
      patchCapabilityDevice(capabilityId, deviceId, { enabled, state }),
    onSuccess: async (_result, vars) => {
      await Promise.all([
        client.invalidateQueries({ queryKey: queryKeys.automationAssignments(vars.capabilityId) }),
        client.invalidateQueries({ queryKey: queryKeys.deviceCapabilities(vars.deviceId) })
      ]);
    }
  });

  return {
    assignmentsQuery,
    updateAssignment
  };
}
