import { useQuery } from "@tanstack/react-query";

import { fetchDevice } from "@/api/devices";
import { queryKeys } from "@/lib/query-keys";

export function useDevice(mac: string | null) {
  return useQuery({
    queryKey: mac ? queryKeys.deviceDetail(mac) : ["devices", "detail", "none"],
    queryFn: () => fetchDevice(mac ?? ""),
    enabled: Boolean(mac),
    staleTime: 3000
  });
}
