import { useQuery } from "@tanstack/react-query";

import { fetchDevices } from "@/api/devices";
import { queryKeys } from "@/lib/query-keys";
import type { RegistrationScope } from "@/lib/device-semantics";
import type { Device } from "@/types/device";

function sameSources(a: string[], b: string[]) {
  if (a.length !== b.length) {
    return false;
  }
  for (let index = 0; index < a.length; index += 1) {
    if (a[index] !== b[index]) {
      return false;
    }
  }
  return true;
}

function sameDeviceForList(a: Device, b: Device) {
  return (
    a.mac === b.mac &&
    a.name === b.name &&
    a.vendor === b.vendor &&
    (a.icon ?? null) === (b.icon ?? null) &&
    (a.comment ?? null) === (b.comment ?? null) &&
    a.status === b.status &&
    a.online === b.online &&
    (a.last_seen_at ?? null) === (b.last_seen_at ?? null) &&
    (a.connected_since_at ?? null) === (b.connected_since_at ?? null) &&
    (a.last_ip ?? null) === (b.last_ip ?? null) &&
    (a.last_subnet ?? null) === (b.last_subnet ?? null) &&
    (a.first_seen_at ?? null) === (b.first_seen_at ?? null) &&
    (a.created_at ?? null) === (b.created_at ?? null) &&
    sameSources(a.last_sources, b.last_sources)
  );
}

function mergeDevicesForList(
  previous: Device[] | undefined,
  next: Device[]
) {
  if (!previous || previous.length === 0) {
    return next;
  }

  const previousByMac = new Map(previous.map((item) => [item.mac, item]));
  const merged: Device[] = new Array(next.length);
  let hasAnyChange = previous.length !== next.length;

  for (let index = 0; index < next.length; index += 1) {
    const current = next[index];
    const old = previousByMac.get(current.mac);

    if (old && sameDeviceForList(old, current)) {
      merged[index] = old;
      if (!hasAnyChange && previous[index] !== old) {
        hasAnyChange = true;
      }
      continue;
    }

    merged[index] = current;
    hasAnyChange = true;
  }

  return hasAnyChange ? merged : previous;
}

type UseDevicesListQueryParams = {
  paused: boolean;
  query: string;
  segmentation: RegistrationScope;
  vendors: string[];
  sources: string[];
  subnets: string[];
  pageIndex: number;
  pageSize: number;
};

export function useDevicesListQuery({
  paused,
  query,
  segmentation,
  vendors,
  sources,
  subnets,
  pageIndex,
  pageSize
}: UseDevicesListQueryParams) {
  const status = segmentation === "new" || segmentation === "registered" ? segmentation : "all";
  const normalizedQuery = query.trim();

  return useQuery<Device[]>({
    queryKey: queryKeys.devicesList({
      paused,
      query: normalizedQuery,
      segmentation,
      vendors: [...vendors].sort((a, b) => a.localeCompare(b)),
      sources: [...sources].sort((a, b) => a.localeCompare(b)),
      subnets: [...subnets].sort((a, b) => a.localeCompare(b)),
      pageIndex,
      pageSize
    }),
    queryFn: () =>
      fetchDevices({
        status,
        online: "all",
        query: normalizedQuery
      }),
    placeholderData: (previous) =>
      Array.isArray(previous) ? (previous as Device[]) : undefined,
    staleTime: 3000,
    refetchInterval: paused ? false : 5000,
    structuralSharing: (previous, next) =>
      mergeDevicesForList(
        Array.isArray(previous) ? (previous as Device[]) : undefined,
        next as Device[]
      ),
    notifyOnChangeProps: ["data", "error", "isPending"]
  });
}
