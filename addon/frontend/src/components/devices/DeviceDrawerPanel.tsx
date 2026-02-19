import { useEffect, useMemo, useRef, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Cable,
  Expand,
  Monitor,
  Shrink,
  Wifi
} from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { CopyValue } from "@/components/devices/CopyValue";
import { DeviceStatusBadge } from "@/components/device/DeviceStatusBadge";
import { DeviceTypeIcon } from "@/components/device/DeviceTypeIcon";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from "@/components/ui/tooltip";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle
} from "@/components/ui/drawer";
import {
  getPrimaryInterface,
  getSourceBreakdown,
  inferDeviceType,
  parseStoredIcon,
  toStoredIcon,
  type DeviceType
} from "@/lib/device";
import {
  isDeviceOnline,
  isNewDevice,
  isUnregisteredDevice
} from "@/lib/device-semantics";
import { formatExactTimestamp, formatLastSeenLabel } from "@/lib/time";
import type { Device } from "@/types/device";

const registrationSchema = z.object({
  name: z.string().trim().min(1, "Display name is required"),
  icon: z.enum(["wifi", "wired", "unknown"]),
  comment: z.string().max(200, "Comment is too long")
});

type RegistrationValues = z.infer<typeof registrationSchema>;

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  isMobile: boolean;
  device: Device | null;
  now: number;
  isLoading: boolean;
  isSaving: boolean;
  onSave: (payload: { name: string; icon: string | undefined; comment: string }) => Promise<void>;
};

function SourcePill({ active, label }: { active: boolean; label: string }) {
  return (
    <Badge
      variant={active ? "secondary" : "outline"}
      className={active ? "font-medium" : "text-muted-foreground"}
      aria-label={`${label} source ${active ? "active" : "inactive"}`}
    >
      {label}
    </Badge>
  );
}

export function DeviceDrawerPanel({
  open,
  onOpenChange,
  isMobile,
  device,
  now,
  isLoading,
  isSaving,
  onSave
}: Props) {
  const [wide, setWide] = useState(false);
  const inputRef = useRef<HTMLInputElement | null>(null);

  const form = useForm<RegistrationValues>({
    resolver: zodResolver(registrationSchema),
    defaultValues: {
      name: "",
      icon: "unknown",
      comment: ""
    }
  });

  useEffect(() => {
    if (!open) {
      return;
    }
    const timer = window.setTimeout(() => {
      inputRef.current?.focus();
      inputRef.current?.select();
    }, 50);

    return () => {
      window.clearTimeout(timer);
    };
  }, [open]);

  useEffect(() => {
    if (!device) {
      return;
    }

    form.reset({
      name: device.name,
      icon: parseStoredIcon(device.icon, inferDeviceType(device)),
      comment: device.comment ?? ""
    });
  }, [device, form]);

  const breakdown = useMemo(
    () =>
      device
        ? getSourceBreakdown(device)
        : { dhcp: false, wifi: false, arp: false, bridge: false },
    [device]
  );

  const saveLabel = device && isUnregisteredDevice(device) ? "Register" : "Save";
  const isOnline = device ? isDeviceOnline(device) : false;

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSave({
      name: values.name,
      icon: toStoredIcon(values.icon as DeviceType),
      comment: values.comment
    });
  });

  const content = (
    <>
      {!isMobile ? (
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setWide((current) => !current)}>
            {wide ? <Shrink className="mr-1 h-4 w-4" /> : <Expand className="mr-1 h-4 w-4" />}
            {wide ? "Default width" : "Wide mode"}
          </Button>
        </div>
      ) : null}

      <Tabs defaultValue="identity" className="w-full">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="identity">Identity</TabsTrigger>
          <TabsTrigger value="network">Network</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
          <TabsTrigger value="notes">Notes</TabsTrigger>
          <TabsTrigger value="registration">Registration</TabsTrigger>
        </TabsList>

        <TabsContent value="identity" className="space-y-4 pt-4">
          <div className="flex items-start gap-3">
            <DeviceTypeIcon
              type={device ? parseStoredIcon(device.icon, inferDeviceType(device)) : "unknown"}
            />
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <p className="text-base font-semibold">{device?.name ?? "-"}</p>
                <DeviceStatusBadge online={isOnline} />
                {device && isNewDevice(device) ? (
                  <Badge variant="outline" aria-label="New device">
                    NEW
                  </Badge>
                ) : null}
              </div>
              <p className="text-sm text-muted-foreground">{device?.vendor || "Unknown vendor"}</p>
              <CopyValue
                label="MAC"
                value={device?.mac}
                mono
                className="h-7 w-fit max-w-xs px-2 text-xs text-muted-foreground"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">Registration</p>
              <p>{device && isUnregisteredDevice(device) ? "Unregistered" : "Registered"}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">First seen</p>
              <p>{formatExactTimestamp(device?.first_seen_at)}</p>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="network" className="space-y-4 pt-4">
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">IP</p>
              <CopyValue
                label="IP"
                value={device?.last_ip}
                mono
                className="h-8 w-full px-2 text-sm text-foreground"
              />
            </div>
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">Subnet</p>
              <CopyValue
                label="Subnet"
                value={device?.last_subnet}
                mono
                className="h-8 w-full px-2 text-sm text-foreground"
              />
            </div>
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">Interface</p>
              <CopyValue
                label="Interface"
                value={device ? getPrimaryInterface(device) : null}
                className="h-8 w-full px-2 text-sm text-foreground"
              />
            </div>
          </div>

          <div className="space-y-2">
            <p className="text-xs uppercase tracking-wide text-muted-foreground">Source Breakdown</p>
            <div className="flex flex-wrap gap-2">
              <SourcePill label="DHCP" active={breakdown.dhcp} />
              <SourcePill label="WiFi" active={breakdown.wifi} />
              <SourcePill label="ARP" active={breakdown.arp} />
              <SourcePill label="Bridge" active={breakdown.bridge} />
            </div>
          </div>
        </TabsContent>

        <TabsContent value="activity" className="space-y-4 pt-4">
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">Last seen</p>
              <Tooltip>
                <TooltipTrigger asChild>
                  <p className="cursor-help">
                    {formatLastSeenLabel(isOnline, device?.last_seen_at, now)}
                  </p>
                </TooltipTrigger>
                <TooltipContent>{formatExactTimestamp(device?.last_seen_at)}</TooltipContent>
              </Tooltip>
            </div>
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">Connected since</p>
              <Tooltip>
                <TooltipTrigger asChild>
                  <p className="cursor-help">
                    {device?.connected_since_at
                      ? formatLastSeenLabel(false, device.connected_since_at, now)
                      : "-"}
                  </p>
                </TooltipTrigger>
                <TooltipContent>
                  {formatExactTimestamp(device?.connected_since_at)}
                </TooltipContent>
              </Tooltip>
            </div>
            <div className="space-y-1">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">Updated at</p>
              <p>{formatExactTimestamp(device?.updated_at)}</p>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="notes" className="space-y-4 pt-4">
          <div className="space-y-1 text-sm">
            <p className="text-xs uppercase tracking-wide text-muted-foreground">Current comment</p>
            <p>{device?.comment?.trim() || "No comment"}</p>
          </div>
        </TabsContent>

        <TabsContent value="registration" className="space-y-4 pt-4">
          <Form {...form}>
            <form className="space-y-4" onSubmit={handleSubmit}>
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Display name</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        ref={(element) => {
                          field.ref(element);
                          inputRef.current = element;
                        }}
                        placeholder="Living room iPhone"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="icon"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Icon</FormLabel>
                    <FormControl>
                      <ToggleGroup
                        type="single"
                        value={field.value}
                        onValueChange={(value) => {
                          if (value === "wifi" || value === "wired" || value === "unknown") {
                            field.onChange(value);
                          }
                        }}
                        variant="outline"
                      >
                        <ToggleGroupItem value="wifi" aria-label="WiFi icon">
                          <Wifi className="mr-1 h-4 w-4" /> WiFi
                        </ToggleGroupItem>
                        <ToggleGroupItem value="wired" aria-label="Wired icon">
                          <Cable className="mr-1 h-4 w-4" /> Wired
                        </ToggleGroupItem>
                        <ToggleGroupItem value="unknown" aria-label="Unknown icon">
                          <Monitor className="mr-1 h-4 w-4" /> Unknown
                        </ToggleGroupItem>
                      </ToggleGroup>
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="comment"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Comment</FormLabel>
                    <FormControl>
                      <Textarea
                        {...field}
                        className="min-h-20"
                        placeholder="Optional note"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </form>
          </Form>
        </TabsContent>
      </Tabs>
    </>
  );

  if (isMobile) {
    return (
      <TooltipProvider>
        <Drawer open={open} onOpenChange={onOpenChange}>
          <DrawerContent className="flex h-[100dvh] max-h-[100dvh] flex-col p-0">
            <DrawerHeader className="sticky top-0 z-10 border-b bg-background px-4 pt-[max(env(safe-area-inset-top),0.75rem)] text-left">
              <DrawerTitle>{device?.name ?? "Device details"}</DrawerTitle>
              <DrawerDescription>{device?.mac ?? ""}</DrawerDescription>
            </DrawerHeader>

            <ScrollArea className="flex-1">
              <div className="space-y-4 p-4 pb-[calc(1rem+env(safe-area-inset-bottom))]">
                {isLoading ? <p className="text-sm text-muted-foreground">Loading...</p> : null}
                {content}
              </div>
            </ScrollArea>

            <DrawerFooter className="sticky bottom-0 z-10 border-t bg-background px-4 pb-[max(env(safe-area-inset-bottom),0.75rem)] pt-3">
              <div className="flex w-full items-center justify-between gap-2">
                <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={isSaving}>
                  Cancel
                </Button>
                <Button onClick={() => void handleSubmit()} disabled={isSaving || !device}>
                  {saveLabel}
                </Button>
              </div>
            </DrawerFooter>
          </DrawerContent>
        </Drawer>
      </TooltipProvider>
    );
  }

  return (
    <TooltipProvider>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent
          side="right"
          className={`flex h-[100dvh] max-h-[100dvh] flex-col p-0 ${
            wide ? "w-[min(94vw,920px)] sm:max-w-[920px]" : "w-[min(94vw,640px)] sm:max-w-[640px]"
          }`}
        >
          <SheetHeader className="sticky top-0 z-10 border-b bg-background px-6 pb-4 pt-[max(env(safe-area-inset-top),1rem)] text-left">
            <SheetTitle>{device?.name ?? "Device details"}</SheetTitle>
            <SheetDescription>{device?.mac ?? ""}</SheetDescription>
          </SheetHeader>

          <ScrollArea className="flex-1">
            <div className="space-y-4 p-6 pb-[calc(1rem+env(safe-area-inset-bottom))]">
              {isLoading ? <p className="text-sm text-muted-foreground">Loading...</p> : null}
              {content}
            </div>
          </ScrollArea>

          <div className="sticky bottom-0 z-10 flex items-center justify-between gap-2 border-t bg-background px-6 pb-[max(env(safe-area-inset-bottom),1rem)] pt-4">
            <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={isSaving}>
              Cancel
            </Button>
            <Button onClick={() => void handleSubmit()} disabled={isSaving || !device}>
              {saveLabel}
            </Button>
          </div>
        </SheetContent>
      </Sheet>
    </TooltipProvider>
  );
}
