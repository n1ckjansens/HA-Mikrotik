import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  fetchGlobalCapabilities,
  patchGlobalCapability
} from "@/api/automation";
import { queryKeys } from "@/lib/query-keys";
import type { SetStateResult } from "@/types/automation";

export function useGlobalCapabilities() {
  return useQuery({
    queryKey: queryKeys.automationGlobalCapabilities,
    queryFn: () => fetchGlobalCapabilities(),
    staleTime: 5000,
    refetchInterval: 5000,
    refetchIntervalInBackground: true
  });
}

type UpdateInput = {
  capabilityId: string;
  state?: string;
  enabled?: boolean;
};

export function useUpdateGlobalCapability() {
  const client = useQueryClient();

  return useMutation<SetStateResult, Error, UpdateInput>({
    mutationFn: ({ capabilityId, state, enabled }) =>
      patchGlobalCapability(capabilityId, { state, enabled }),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.automationGlobalCapabilities });
    }
  });
}
