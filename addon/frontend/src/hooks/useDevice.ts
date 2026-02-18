import { useQuery } from "@tanstack/react-query";

import { fetchDevice } from "@/api/devices";

export function useDevice(mac: string | null) {
  return useQuery({
    queryKey: ["device", mac],
    queryFn: () => fetchDevice(mac ?? ""),
    enabled: Boolean(mac)
  });
}
