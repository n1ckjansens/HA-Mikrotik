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
  deviceCapabilities: (mac: string) => ["devices", "capabilities", mac] as const,
  devicesSummary: ["devices", "summary"] as const,
  savedViews: ["devices", "savedViews"] as const,
  automationActionTypes: ["automation", "action-types"] as const,
  automationStateSourceTypes: ["automation", "state-source-types"] as const,
  automationCapabilities: (search: string, category: string) =>
    ["automation", "capabilities", { search, category }] as const,
  automationCapabilityDetail: (capabilityId: string) =>
    ["automation", "capability", capabilityId] as const,
  automationAssignments: (capabilityId: string) =>
    ["automation", "assignments", capabilityId] as const
};
