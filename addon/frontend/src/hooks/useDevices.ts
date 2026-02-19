import { useQuery } from "@tanstack/react-query";

import { fetchDevices, type DevicesFilterParams } from "@/api/devices";

export function useDevices(filters: DevicesFilterParams) {
  return useQuery({
    queryKey: ["devices", filters],
    queryFn: () => fetchDevices(filters),
    staleTime: 3000,
    refetchInterval: 5000
  });
}
