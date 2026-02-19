import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  createCapability,
  deleteCapability,
  fetchCapability,
  updateCapability
} from "@/api/automation";
import { queryKeys } from "@/lib/query-keys";
import type { CapabilityTemplate } from "@/types/automation";

export function useCapabilityEditor(capabilityId: string | null) {
  const client = useQueryClient();
  const isNew = !capabilityId || capabilityId === "new";

  const capabilityQuery = useQuery({
    queryKey: capabilityId
      ? queryKeys.automationCapabilityDetail(capabilityId)
      : ["automation", "capability", "new"],
    queryFn: () => fetchCapability(capabilityId ?? ""),
    enabled: !isNew
  });

  const createMutation = useMutation({
    mutationFn: (template: CapabilityTemplate) => createCapability(template),
    onSuccess: async (_created, template) => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ["automation", "capabilities"] }),
        client.invalidateQueries({ queryKey: queryKeys.automationCapabilityDetail(template.id) }),
        client.invalidateQueries({ queryKey: queryKeys.automationGlobalCapabilities })
      ]);
    }
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, template }: { id: string; template: CapabilityTemplate }) =>
      updateCapability(id, template),
    onSuccess: async (_updated, vars) => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ["automation", "capabilities"] }),
        client.invalidateQueries({ queryKey: queryKeys.automationCapabilityDetail(vars.id) }),
        client.invalidateQueries({ queryKey: queryKeys.automationGlobalCapabilities })
      ]);
    }
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteCapability(id),
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ["automation", "capabilities"] }),
        client.invalidateQueries({ queryKey: queryKeys.automationGlobalCapabilities })
      ]);
    }
  });

  const saveCapability = async (template: CapabilityTemplate) => {
    if (isNew) {
      await createMutation.mutateAsync(template);
      return;
    }
    await updateMutation.mutateAsync({ id: capabilityId ?? template.id, template });
  };

  return {
    isNew,
    capabilityQuery,
    createMutation,
    updateMutation,
    deleteMutation,
    saveCapability,
    isSaving: createMutation.isPending || updateMutation.isPending
  };
}
