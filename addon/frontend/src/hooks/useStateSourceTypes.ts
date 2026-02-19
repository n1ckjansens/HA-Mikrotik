import { useQuery } from "@tanstack/react-query";

import { fetchStateSourceTypes } from "@/api/automation";
import { queryKeys } from "@/lib/query-keys";

export function useStateSourceTypes() {
  return useQuery({
    queryKey: queryKeys.automationStateSourceTypes,
    queryFn: () => fetchStateSourceTypes(),
    staleTime: 60_000
  });
}
