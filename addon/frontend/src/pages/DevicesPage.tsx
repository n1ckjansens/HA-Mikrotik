import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  type SortingState,
  type VisibilityState,
  useReactTable
} from "@tanstack/react-table";
import { useQueryClient } from "@tanstack/react-query";
import { LayoutGrid, Table2 } from "lucide-react";
import { toast } from "sonner";

import { ApiError } from "@/api/client";
import { DeviceDrawerPanel } from "@/components/devices/DeviceDrawerPanel";
import { buildDeviceColumns } from "@/components/devices/device-table-columns";
import { DevicesTable } from "@/components/devices/DevicesTable";
import { KpiFilters } from "@/components/devices/KpiFilters";
import { OverviewHeader } from "@/components/devices/OverviewHeader";
import {
  DisconnectedState,
  IntegrationRequiredState,
  NoDevicesState,
  NoResultsState
} from "@/components/devices/SystemStates";
import { UnifiedFilterBar } from "@/components/devices/UnifiedFilterBar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { AppShell } from "@/components/layout/AppShell";
import { useDevice } from "@/hooks/useDevice";
import { useDevicesListQuery } from "@/hooks/useDevicesListQuery";
import { useIsMobile } from "@/hooks/useIsMobile";
import { useLiveUpdates } from "@/hooks/useLiveUpdates";
import { useRegisterDevice } from "@/hooks/useRegisterDevice";
import { useSavedViews } from "@/hooks/useSavedViews";
import {
  isDeviceOnline,
  isNewDevice,
  isUnregisteredDevice,
  matchFacet,
  matchOnlineScope,
  matchRegistrationScope,
  matchSearch,
  uniqueSorted,
  type OnlineScope,
  type RegistrationScope
} from "@/lib/device-semantics";
import { queryKeys } from "@/lib/query-keys";
import { formatExactTimestamp } from "@/lib/time";
import type { Device } from "@/types/device";

type Density = "comfortable" | "compact";
type FacetsState = {
  vendors: string[];
  sources: string[];
  subnets: string[];
};

const DEFAULT_FACETS: FacetsState = {
  vendors: [],
  sources: [],
  subnets: []
};

function toggleArrayValue(items: string[], value: string) {
  if (items.includes(value)) {
    return items.filter((item) => item !== value);
  }
  return [...items, value];
}

export function DevicesPage() {
  const live = useLiveUpdates();
  const register = useRegisterDevice();
  const savedViews = useSavedViews();
  const isMobile = useIsMobile();
  const queryClient = useQueryClient();

  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [segmentation, setSegmentation] = useState<RegistrationScope>("all");
  const [onlineScope, setOnlineScope] = useState<OnlineScope>("any");
  const [facets, setFacets] = useState<FacetsState>(DEFAULT_FACETS);
  const [sorting, setSorting] = useState<SortingState>([
    { id: "name", desc: false }
  ]);
  const [pagination, setPagination] = useState({ pageIndex: 0, pageSize: 20 });
  const [density, setDensity] = useState<Density>("comfortable");
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [selectedMac, setSelectedMac] = useState<string | null>(null);
  const [focusRestoreEl, setFocusRestoreEl] = useState<HTMLElement | null>(null);
  const [isManualRefreshing, setIsManualRefreshing] = useState(false);
  const refreshTimerRef = useRef<number | null>(null);

  const searchInputRef = useRef<HTMLInputElement>(null);

  const devicesQuery = useDevicesListQuery({
    paused: live.isPaused,
    query: debouncedQuery,
    segmentation,
    vendors: facets.vendors,
    sources: facets.sources,
    subnets: facets.subnets,
    pageIndex: pagination.pageIndex,
    pageSize: pagination.pageSize
  });

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebouncedQuery(query);
    }, 300);

    return () => {
      window.clearTimeout(timer);
    };
  }, [query]);

  useEffect(() => {
    return () => {
      if (refreshTimerRef.current !== null) {
        window.clearTimeout(refreshTimerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    setPagination((current) => ({ ...current, pageIndex: 0 }));
  }, [
    debouncedQuery,
    segmentation,
    onlineScope,
    facets.vendors,
    facets.sources,
    facets.subnets
  ]);

  useEffect(() => {
    setColumnVisibility(
      isMobile
        ? {
            last_subnet: false,
            vendor: false,
            source: false,
            registration: false
          }
        : {}
    );
  }, [isMobile]);

  const integrationNotConfigured =
    devicesQuery.error instanceof ApiError &&
    devicesQuery.error.code === "integration_not_configured";
  const isInitialLoading = devicesQuery.isPending && !devicesQuery.data;
  const lastSuccessfulAt =
    devicesQuery.dataUpdatedAt > 0 ? devicesQuery.dataUpdatedAt : undefined;

  const routerDisconnected = Boolean(devicesQuery.error) && !integrationNotConfigured;
  const routerMessage =
    devicesQuery.error instanceof Error
      ? devicesQuery.error.message
      : "Unable to reach router";

  const devices = useMemo(() => devicesQuery.data ?? [], [devicesQuery.data]);

  const summary = useMemo(() => {
    const nowMs = Date.now();
    const online = devices.filter((device) => isDeviceOnline(device, nowMs)).length;
    const newCount = devices.filter((device) => isNewDevice(device)).length;
    const unregistered = devices.filter((device) => isUnregisteredDevice(device)).length;
    const registered = devices.length - unregistered;
    const offline = devices.filter((device) => !isDeviceOnline(device, nowMs)).length;

    return {
      online,
      offline,
      newCount,
      unregistered,
      registered
    };
  }, [devices]);

  const options = useMemo(
    () => ({
      vendors: uniqueSorted(devices.map((item) => item.vendor)),
      sources: uniqueSorted(devices.flatMap((item) => item.last_sources)),
      subnets: uniqueSorted(devices.map((item) => item.last_subnet))
    }),
    [devices]
  );

  const filteredDevices = useMemo(
    () => {
      const nowMs = Date.now();
      return devices.filter((device) => {
        if (!matchRegistrationScope(device, segmentation)) {
          return false;
        }
        if (!matchOnlineScope(device, onlineScope, nowMs)) {
          return false;
        }
        if (!matchSearch(device, debouncedQuery)) {
          return false;
        }
        if (!matchFacet(device.vendor, facets.vendors)) {
          return false;
        }
        if (!matchFacet(device.last_subnet, facets.subnets)) {
          return false;
        }
        if (
          facets.sources.length > 0 &&
          !device.last_sources.some((source) => facets.sources.includes(source))
        ) {
          return false;
        }
        return true;
      });
    },
    [devices, segmentation, onlineScope, debouncedQuery, facets]
  );

  const selectedFromList = useMemo(
    () => devices.find((device) => device.mac === selectedMac) ?? null,
    [devices, selectedMac]
  );
  const detailQuery = useDevice(selectedMac);
  const selectedDevice = detailQuery.data ?? selectedFromList;

  const openDevice = useCallback(
    (device: Device, trigger: HTMLElement | null) => {
      setSelectedMac(device.mac);
      setFocusRestoreEl(trigger);
    },
    []
  );

  const closeDevice = () => {
    setSelectedMac(null);
    const toFocus = focusRestoreEl;
    setFocusRestoreEl(null);
    window.setTimeout(() => {
      toFocus?.focus();
    }, 0);
  };

  const columns = useMemo(
    () => buildDeviceColumns({ isMobile, onOpenDevice: openDevice }),
    [isMobile, openDevice]
  );

  const table = useReactTable({
    data: filteredDevices,
    columns,
    state: {
      sorting,
      columnVisibility,
      pagination
    },
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    onPaginationChange: setPagination,
    getRowId: (row) => row.mac,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel()
  });

  const refetchDevices = devicesQuery.refetch;

  const handleRefresh = useCallback(() => {
    if (refreshTimerRef.current !== null) {
      window.clearTimeout(refreshTimerRef.current);
    }

    setIsManualRefreshing(true);
    void refetchDevices();
    refreshTimerRef.current = window.setTimeout(() => {
      setIsManualRefreshing(false);
      refreshTimerRef.current = null;
    }, 1000);
  }, [refetchDevices]);

  const handleSaveDevice = async (payload: {
    name: string;
    icon: string | undefined;
    comment: string;
  }) => {
    if (!selectedDevice) {
      return;
    }

    const previous = selectedDevice;
    await register.saveDevice(selectedDevice, payload);

    toast.success(
      isUnregisteredDevice(previous) ? "Device registered" : "Device updated",
      {
        action: {
          label: "Undo",
          onClick: () => {
            const current =
              queryClient.getQueryData<Device>(
                queryKeys.deviceDetail(previous.mac)
              ) ?? previous;
            void register.saveDevice(
              current,
              {
                name: previous.name,
                icon: previous.icon ?? undefined,
                comment: previous.comment ?? ""
              }
            );
          }
        }
      }
    );
  };

  const lastSuccessfulLabel = formatExactTimestamp(
    lastSuccessfulAt ? new Date(lastSuccessfulAt).toISOString() : undefined
  );

  const clearAllFilters = () => {
    setQuery("");
    setDebouncedQuery("");
    setSegmentation("all");
    setOnlineScope("any");
    setFacets(DEFAULT_FACETS);
  };

  const applySavedView = (id: string) => {
    const view = savedViews.getView(id);
    if (!view) {
      return;
    }

    setQuery(view.search);
    setDebouncedQuery(view.search);
    setSegmentation(view.registrationScope);
    setOnlineScope(view.onlineScope);
    setFacets({
      vendors: view.vendors,
      sources: view.sources,
      subnets: view.subnets
    });
  };

  const toggleFacet = (kind: keyof FacetsState, value: string) => {
    setFacets((current) => ({
      ...current,
      [kind]: toggleArrayValue(current[kind], value)
    }));
  };

  if (integrationNotConfigured) {
    return (
      <AppShell
        header={
          <OverviewHeader
            connected={false}
            updatedAt={lastSuccessfulAt}
            isPaused={live.isPaused}
            onTogglePause={live.toggle}
            onRefresh={() => {
              void handleRefresh();
            }}
            isRefreshing={isManualRefreshing}
            lastSuccessfulLabel={lastSuccessfulLabel}
            searchInputRef={searchInputRef}
          />
        }
      >
        <IntegrationRequiredState />
      </AppShell>
    );
  }

  return (
    <AppShell
      header={
        <OverviewHeader
          connected={!routerDisconnected}
          updatedAt={lastSuccessfulAt}
          isPaused={live.isPaused}
          onTogglePause={live.toggle}
          onRefresh={() => {
            void handleRefresh();
          }}
          isRefreshing={isManualRefreshing}
          lastSuccessfulLabel={lastSuccessfulLabel}
          searchInputRef={searchInputRef}
        />
      }
    >
      {routerDisconnected ? (
        <DisconnectedState
          message={routerMessage}
          onRetry={() => {
            void handleRefresh();
          }}
          lastSuccessfulLabel={lastSuccessfulLabel}
        />
      ) : null}

      {isInitialLoading ? (
        <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
          {Array.from({ length: 5 }).map((_, index) => (
            <Skeleton key={index} className="h-20" />
          ))}
        </div>
      ) : (
        <KpiFilters
          online={summary.online}
          offline={summary.offline}
          newCount={summary.newCount}
          registered={summary.registered}
          unregistered={summary.unregistered}
          activeScope={segmentation}
          onScopeChange={setSegmentation}
        />
      )}

      <UnifiedFilterBar
        isMobile={isMobile}
        searchInputRef={searchInputRef}
        query={query}
        segmentation={segmentation}
        onlineScope={onlineScope}
        facets={facets}
        options={options}
        savedViews={savedViews.savedViews}
        onQueryChange={setQuery}
        onClearQuery={() => {
          setQuery("");
          setDebouncedQuery("");
        }}
        onSegmentationChange={setSegmentation}
        onOnlineScopeChange={setOnlineScope}
        onToggleFacet={toggleFacet}
        onClearAll={clearAllFilters}
        onApplyView={applySavedView}
        onSaveCurrentView={() => {
          savedViews.saveView({
            search: query,
            registrationScope: segmentation,
            onlineScope,
            vendors: facets.vendors,
            sources: facets.sources,
            subnets: facets.subnets
          });
          toast.success("View saved");
        }}
        onDeleteView={(id) => {
          savedViews.removeView(id);
        }}
      />

      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm text-muted-foreground">
          {filteredDevices.length} / {devices.length} devices
        </p>

        <div className="flex items-center gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <LayoutGrid className="mr-2 h-4 w-4" />
                Density
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Density</DropdownMenuLabel>
              <DropdownMenuRadioGroup
                value={density}
                onValueChange={(value) => {
                  if (value === "comfortable" || value === "compact") {
                    setDensity(value);
                  }
                }}
              >
                <DropdownMenuRadioItem value="comfortable">
                  Comfortable
                </DropdownMenuRadioItem>
                <DropdownMenuRadioItem value="compact">Compact</DropdownMenuRadioItem>
              </DropdownMenuRadioGroup>
            </DropdownMenuContent>
          </DropdownMenu>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <Table2 className="mr-2 h-4 w-4" />
                Columns
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Toggle columns</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {table
                .getAllLeafColumns()
                .filter((column) => column.getCanHide())
                .map((column) => (
                  <DropdownMenuCheckboxItem
                    key={column.id}
                    checked={column.getIsVisible()}
                    onCheckedChange={(value) => column.toggleVisibility(!!value)}
                  >
                    {column.id}
                  </DropdownMenuCheckboxItem>
                ))}
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => {
                  setColumnVisibility(
                    isMobile
                      ? {
                          last_subnet: false,
                          vendor: false,
                          source: false,
                          registration: false
                        }
                      : {}
                  );
                }}
              >
                Reset
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {devices.length === 0 && !isInitialLoading ? <NoDevicesState /> : null}
      {devices.length > 0 && filteredDevices.length === 0 ? (
        <NoResultsState onClear={clearAllFilters} />
      ) : null}
      {filteredDevices.length > 0 ? (
        <DevicesTable
          table={table}
          isLoading={isInitialLoading}
          density={density}
        />
      ) : null}

      <DeviceDrawerPanel
        open={selectedMac !== null}
        onOpenChange={(open) => {
          if (!open) {
            closeDevice();
          }
        }}
        isMobile={isMobile}
        device={selectedDevice}
        now={Date.now()}
        isLoading={detailQuery.isPending && !detailQuery.data}
        isSaving={register.isSaving}
        onSave={handleSaveDevice}
      />
    </AppShell>
  );
}
