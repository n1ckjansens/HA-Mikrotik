import { useQuery } from "@tanstack/react-query";

import { fetchCapabilities } from "@/api/automation";
import { queryKeys } from "@/lib/query-keys";

type Params = {
  search: string;
  category: string;
};

export function useCapabilities({ search, category }: Params) {
  const normalizedSearch = search.trim();
  const normalizedCategory = category.trim();

  return useQuery({
    queryKey: queryKeys.automationCapabilities(normalizedSearch, normalizedCategory),
    queryFn: () =>
      fetchCapabilities({
        search: normalizedSearch,
        category: normalizedCategory
      }),
    staleTime: 15_000
  });
}
