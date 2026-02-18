import { Fragment, useMemo, useState } from "react";

import { DeviceCard } from "@/components/device/DeviceCard";
import { DeviceGroupDivider } from "@/components/device/DeviceGroupDivider";
import { RegisterDevicePrompt } from "@/components/device/RegisterDevicePrompt";
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
import { useIsMobile } from "@/hooks/useIsMobile";
import { useRegisterDialog } from "@/hooks/useRegisterDialog";

export function DevicesPage() {
  const filters = useFilters();
  const devicesQuery = useDevices(filters.params);
  const registerDialog = useRegisterDialog();
  const isMobile = useIsMobile();

  const [selectedMac, setSelectedMac] = useState<string | null>(null);
  const detailQuery = useDevice(selectedMac);

  const isBusy = registerDialog.isSaving;

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
        {devicesQuery.data?.map((device, index, items) => {
          const prev = index > 0 ? items[index - 1] : null;
          const showRegisteredDivider =
            filters.status === "all" && prev?.status === "new" && device.status === "registered";

          return (
            <Fragment key={device.mac}>
              {showRegisteredDivider ? <DeviceGroupDivider label="Registered" /> : null}
              <DeviceRow
                device={device}
                busy={isBusy}
                onOpenDetails={(mac) => setSelectedMac(mac)}
                onRegister={registerDialog.openForDevice}
              />
            </Fragment>
          );
        })}
      </div>

      <DeviceCard
        open={selectedMac !== null}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedMac(null);
          }
        }}
        device={detailQuery.data ?? null}
        isMobile={isMobile}
      />

      <RegisterDevicePrompt
        device={registerDialog.target}
        isMobile={isMobile}
        open={registerDialog.isOpen}
        name={registerDialog.name}
        isSaving={registerDialog.isSaving}
        onOpenChange={(open) => {
          if (!open) {
            registerDialog.close();
          }
        }}
        onNameChange={registerDialog.setName}
        onSave={() => {
          void registerDialog.save();
        }}
      />
    </AppShell>
  );
}
