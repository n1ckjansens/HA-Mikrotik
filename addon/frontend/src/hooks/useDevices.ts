import { useQuery } from "@tanstack/react-query";

import { fetchDevices, type DevicesFilterParams } from "@/api/devices";

export function useDevices(filters: DevicesFilterParams) {
  return useQuery({
    queryKey: ["devices", filters],
    queryFn: () => fetchDevices(filters),
    refetchInterval: 5000
  });
}
