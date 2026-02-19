import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useCapabilities } from "@/hooks/useCapabilities";
import { useAssignments } from "@/hooks/useAssignments";
import { useDeviceCapabilities, useUpdateDeviceCapability } from "@/hooks/useDeviceCapabilities";
import { useDevices } from "@/hooks/useDevices";

function stateBadgeVariant(state: string) {
  if (state === "on" || state === "allow") {
    return "secondary" as const;
  }
  return "outline" as const;
}

export function AssignmentsPage() {
  const capabilitiesQuery = useCapabilities({ search: "", category: "" });
  const devicesQuery = useDevices({ status: "all", online: "all", query: "" });

  const [activeTab, setActiveTab] = useState("capability");
  const [selectedCapabilityId, setSelectedCapabilityId] = useState("");
  const [selectedDeviceId, setSelectedDeviceId] = useState("");

  const assignments = useAssignments(selectedCapabilityId);
  const deviceCapabilitiesQuery = useDeviceCapabilities(selectedDeviceId || null);
  const updateDeviceCapability = useUpdateDeviceCapability();

  const capabilities = useMemo(
    () => (capabilitiesQuery.data ?? []).filter((item) => item.scope === "device"),
    [capabilitiesQuery.data]
  );
  const devices = useMemo(() => devicesQuery.data ?? [], [devicesQuery.data]);

  useEffect(() => {
    if (capabilities.length === 0) {
      if (selectedCapabilityId) {
        setSelectedCapabilityId("");
      }
      return;
    }
    if (!capabilities.some((item) => item.id === selectedCapabilityId)) {
      setSelectedCapabilityId(capabilities[0].id);
    }
  }, [capabilities, selectedCapabilityId]);

  useEffect(() => {
    if (!selectedDeviceId && devices.length > 0) {
      setSelectedDeviceId(devices[0].mac);
    }
  }, [devices, selectedDeviceId]);

  const currentCapability = useMemo(
    () => capabilities.find((item) => item.id === selectedCapabilityId) ?? null,
    [capabilities, selectedCapabilityId]
  );

  return (
    <div className="space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Assignments</h1>
        <p className="text-sm text-muted-foreground">
          Manage capability bindings and per-device enabled state.
        </p>
      </header>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList>
          <TabsTrigger value="capability">By capability</TabsTrigger>
          <TabsTrigger value="device">By device</TabsTrigger>
        </TabsList>

        <TabsContent value="capability" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Capability</CardTitle>
            </CardHeader>
            <CardContent>
              <Select value={selectedCapabilityId} onValueChange={setSelectedCapabilityId}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose capability" />
                </SelectTrigger>
                <SelectContent>
                  {capabilities.map((capability) => (
                    <SelectItem key={capability.id} value={capability.id}>
                      {capability.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Device</TableHead>
                    <TableHead>IP</TableHead>
                    <TableHead>Enabled</TableHead>
                    <TableHead>State</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(assignments.assignmentsQuery.data ?? []).map((item) => (
                    <TableRow key={item.device_id}>
                      <TableCell>
                        <p className="font-medium">{item.device_name}</p>
                        <p className="text-xs text-muted-foreground">{item.device_id}</p>
                      </TableCell>
                      <TableCell>{item.device_ip || "-"}</TableCell>
                      <TableCell>
                        <Switch
                          checked={item.enabled}
                          onCheckedChange={(checked) => {
                            void assignments.updateAssignment
                              .mutateAsync({
                                capabilityId: selectedCapabilityId,
                                deviceId: item.device_id,
                                enabled: checked
                              })
                              .then((result) => {
                                if ((result.warnings ?? []).length > 0) {
                                  toast.warning(result.warnings[0].message);
                                }
                              })
                              .catch((error: Error) => toast.error(error.message));
                          }}
                        />
                      </TableCell>
                      <TableCell>
                        <Badge variant={stateBadgeVariant(item.state)}>{item.state}</Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {currentCapability ? (
            <p className="text-xs text-muted-foreground">
              Capability control: <span className="capitalize">{currentCapability.control.type}</span>
            </p>
          ) : null}
        </TabsContent>

        <TabsContent value="device" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Device</CardTitle>
            </CardHeader>
            <CardContent>
              <Select value={selectedDeviceId} onValueChange={setSelectedDeviceId}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose device" />
                </SelectTrigger>
                <SelectContent>
                  {devices.map((device) => (
                    <SelectItem key={device.mac} value={device.mac}>
                      {device.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="space-y-3 p-4">
              {(deviceCapabilitiesQuery.data ?? []).map((item) => (
                <div
                  key={item.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-md border p-3"
                >
                  <div>
                    <p className="font-medium">{item.label}</p>
                    <p className="text-xs text-muted-foreground">{item.id}</p>
                  </div>

                  <div className="flex items-center gap-3">
                    <Badge variant={stateBadgeVariant(item.state)}>{item.state}</Badge>
                    <Switch
                      checked={item.enabled}
                      onCheckedChange={(checked) => {
                        if (!selectedDeviceId) {
                          return;
                        }
                        void updateDeviceCapability
                          .mutateAsync({
                            deviceId: selectedDeviceId,
                            capabilityId: item.id,
                            enabled: checked
                          })
                          .then((result) => {
                            if ((result.warnings ?? []).length > 0) {
                              toast.warning(result.warnings[0].message);
                            }
                          })
                          .catch((error: Error) => toast.error(error.message));
                      }}
                    />
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
