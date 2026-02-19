import { useQuery } from "@tanstack/react-query";

import { fetchActionTypes } from "@/api/automation";
import { queryKeys } from "@/lib/query-keys";

export function useActionTypes() {
  return useQuery({
    queryKey: queryKeys.automationActionTypes,
    queryFn: () => fetchActionTypes(),
    staleTime: 60_000
  });
}
