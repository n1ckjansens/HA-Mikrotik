import { useMemo, useState } from "react";

import type { OnlineFilter } from "@/types/device";

export type StatusFilter = "all" | "new" | "registered";

export function useFilters() {
  const [status, setStatus] = useState<StatusFilter>("all");
  const [online, setOnline] = useState<OnlineFilter>("all");
  const [query, setQuery] = useState("");

  const params = useMemo(
    () => ({
      status,
      online,
      query
    }),
    [status, online, query]
  );

  return {
    status,
    setStatus,
    online,
    setOnline,
    query,
    setQuery,
    params
  };
}
