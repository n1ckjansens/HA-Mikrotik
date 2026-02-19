import { useMemo } from "react";

import { useDevices } from "@/hooks/useDevices";

const defaultSummaryFilters = {
  status: "all" as const,
  online: "all" as const,
  query: ""
};

export function useSummary() {
  const devicesQuery = useDevices(defaultSummaryFilters);

  const summary = useMemo(() => {
    const items = devicesQuery.data ?? [];
    return {
      online: items.filter((item) => item.online).length,
      offline: items.filter((item) => !item.online).length,
      newDevices: items.filter((item) => item.status === "new").length,
      registered: items.filter((item) => item.status === "registered").length
    };
  }, [devicesQuery.data]);

  return {
    ...devicesQuery,
    summary
  };
}
