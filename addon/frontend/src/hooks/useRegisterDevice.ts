import { useMutation, useQueryClient } from "@tanstack/react-query";

import { patchDevice, refreshDevices, registerDevice, type RegisterDeviceInput } from "@/api/devices";

export function useRegisterDevice() {
  const client = useQueryClient();
  const refreshMutation = useMutation({
    mutationFn: () => refreshDevices()
  });

  const registerMutation = useMutation({
    mutationFn: ({ mac, input }: { mac: string; input: RegisterDeviceInput }) =>
      registerDevice(mac, input),
    onSuccess: async () => {
      await refreshMutation.mutateAsync();
      await client.invalidateQueries({ queryKey: ["devices"] });
    }
  });

  const patchMutation = useMutation({
    mutationFn: ({ mac, input }: { mac: string; input: RegisterDeviceInput }) =>
      patchDevice(mac, input),
    onSuccess: async (_data, variables) => {
      await client.invalidateQueries({ queryKey: ["devices"] });
      await client.invalidateQueries({ queryKey: ["device", variables.mac] });
    }
  });

  return {
    refreshMutation,
    registerMutation,
    patchMutation,
    async registerWithName(mac: string, name: string) {
      const trimmedName = name.trim();
      if (!trimmedName) {
        return;
      }
      await registerMutation.mutateAsync({
        mac,
        input: { name: trimmedName }
      });
    }
  };
}
