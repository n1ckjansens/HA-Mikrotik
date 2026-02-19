import type { RegistrationScope } from "@/lib/device-semantics";

type DevicesListQueryParams = {
  paused: boolean;
  query: string;
  segmentation: RegistrationScope;
  vendors: string[];
  sources: string[];
  subnets: string[];
  pageIndex: number;
  pageSize: number;
};

export const queryKeys = {
  devicesList: (params: DevicesListQueryParams) =>
    ["devices", "list", params] as const,
  deviceDetail: (mac: string) => ["devices", "detail", mac] as const,
  devicesSummary: ["devices", "summary"] as const,
  savedViews: ["devices", "savedViews"] as const
};
