import { useMutation, useQueryClient } from "@tanstack/react-query";

import { patchDevice, refreshDevices, registerDevice, type RegisterDeviceInput } from "@/api/devices";

export function useRegisterDevice() {
  const client = useQueryClient();

  const registerMutation = useMutation({
    mutationFn: ({ mac, input }: { mac: string; input: RegisterDeviceInput }) =>
      registerDevice(mac, input),
    onSuccess: async () => {
      await refreshDevices();
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
    registerMutation,
    patchMutation,
    registerByMac(mac: string) {
      registerMutation.mutate({
        mac,
        input: { name: `Device ${mac.slice(-5)}` }
      });
    }
  };
}
