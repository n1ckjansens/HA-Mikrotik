import { useEffect, useMemo, useState } from "react";

import type { OnlineFilter } from "@/types/device";

export type StatusFilter = "all" | "new" | "registered";

export function useFilters() {
  const [status, setStatus] = useState<StatusFilter>("all");
  const [online, setOnline] = useState<OnlineFilter>("all");
  const [queryInput, setQueryInput] = useState("");
  const [query, setQuery] = useState("");

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setQuery(queryInput);
    }, 300);

    return () => {
      window.clearTimeout(timer);
    };
  }, [queryInput]);

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
    queryInput,
    setQueryInput,
    params
  };
}
