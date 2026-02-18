import { useMemo, useState } from "react";

import { DeviceCard } from "@/components/device/DeviceCard";
import { DeviceRow } from "@/components/device/DeviceRow";
import { AppShell } from "@/components/layout/AppShell";
import { Header } from "@/components/layout/Header";
import { DeviceSearch } from "@/components/search/DeviceSearch";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ApiError } from "@/api/client";
import { useDevice } from "@/hooks/useDevice";
import { useDevices } from "@/hooks/useDevices";
import { useFilters } from "@/hooks/useFilters";
import { useRegisterDevice } from "@/hooks/useRegisterDevice";

export function DevicesPage() {
  const filters = useFilters();
  const devicesQuery = useDevices(filters.params);
  const register = useRegisterDevice();

  const [selectedMac, setSelectedMac] = useState<string | null>(null);
  const detailQuery = useDevice(selectedMac);

  const isBusy = register.registerMutation.isPending || register.patchMutation.isPending;

  const integrationNotConfigured = useMemo(() => {
    if (!(devicesQuery.error instanceof ApiError)) {
      return false;
    }
    return devicesQuery.error.code === "integration_not_configured";
  }, [devicesQuery.error]);

  if (integrationNotConfigured) {
    return (
      <AppShell header={<Header />}>
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm">Integration not configured.</p>
            <p className="text-sm text-muted-foreground">
              Add and configure MikroTik Presence integration in Home Assistant Devices &amp; Services.
            </p>
          </CardContent>
        </Card>
      </AppShell>
    );
  }

  return (
    <AppShell header={<Header />}>
      <DeviceSearch value={filters.query} onChange={filters.setQuery} />

      <div className="flex flex-wrap items-center gap-3">
        <Tabs value={filters.status} onValueChange={(value) => filters.setStatus(value as "all" | "new" | "registered") }>
          <TabsList>
            <TabsTrigger value="all">All</TabsTrigger>
            <TabsTrigger value="new">New</TabsTrigger>
            <TabsTrigger value="registered">Registered</TabsTrigger>
          </TabsList>
        </Tabs>
        <div className="flex items-center gap-2">
          <Button
            variant={filters.online === "online" ? "default" : "outline"}
            size="sm"
            onClick={() => filters.setOnline("online")}
          >
            Online
          </Button>
          <Button
            variant={filters.online === "offline" ? "default" : "outline"}
            size="sm"
            onClick={() => filters.setOnline("offline")}
          >
            Offline
          </Button>
          <Button
            variant={filters.online === "all" ? "default" : "outline"}
            size="sm"
            onClick={() => filters.setOnline("all")}
          >
            Any
          </Button>
        </div>
      </div>

      <div className="grid gap-3">
        {devicesQuery.data?.map((device) => (
          <DeviceRow
            key={device.mac}
            device={device}
            busy={isBusy}
            onOpenDetails={(mac) => setSelectedMac(mac)}
            onRegister={register.registerByMac}
          />
        ))}
      </div>

      <DeviceCard
        open={selectedMac !== null}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedMac(null);
          }
        }}
        device={detailQuery.data ?? null}
      />
    </AppShell>
  );
}
