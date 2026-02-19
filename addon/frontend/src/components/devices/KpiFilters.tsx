import { CheckCircle2, CircleDashed, Wifi, WifiOff } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { RegistrationScope } from "@/lib/device-semantics";

type Props = {
  online: number;
  offline: number;
  newCount: number;
  registered: number;
  unregistered: number;
  activeScope: RegistrationScope;
  onScopeChange: (scope: RegistrationScope) => void;
};

const items = [
  {
    key: "online",
    label: "Online",
    icon: Wifi,
    valueKey: "online",
    scope: "all" as const,
    badgeClass: "bg-emerald-500/15 text-emerald-700"
  },
  {
    key: "offline",
    label: "Offline",
    icon: WifiOff,
    valueKey: "offline",
    scope: "all" as const,
    badgeClass: "bg-muted text-muted-foreground"
  },
  {
    key: "new",
    label: "New",
    icon: CircleDashed,
    valueKey: "newCount",
    scope: "new" as const,
    badgeClass: "bg-blue-500/15 text-blue-700"
  },
  {
    key: "registered",
    label: "Registered",
    icon: CheckCircle2,
    valueKey: "registered",
    scope: "registered" as const,
    badgeClass: "bg-primary/15 text-primary"
  },
  {
    key: "unregistered",
    label: "Unregistered",
    icon: CircleDashed,
    valueKey: "unregistered",
    scope: "unregistered" as const,
    badgeClass: "bg-amber-500/15 text-amber-700"
  }
] as const;

export function KpiFilters({
  online,
  offline,
  newCount,
  registered,
  unregistered,
  activeScope,
  onScopeChange
}: Props) {
  const values = { online, offline, newCount, registered, unregistered };

  return (
    <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-5">
      {items.map((item) => {
        const Icon = item.icon;
        const active = item.scope !== "all" && activeScope === item.scope;
        return (
          <Button
            key={item.key}
            variant={active ? "default" : "outline"}
            className="h-auto justify-start px-3 py-3"
            onClick={() => {
              if (item.scope === "all") {
                onScopeChange("all");
              } else {
                onScopeChange(active ? "all" : item.scope);
              }
            }}
            aria-pressed={active}
          >
            <div className="flex w-full items-start justify-between">
              <div className="space-y-1 text-left">
                <div className="flex items-center gap-2">
                  <Icon className="h-4 w-4" />
                  <span className="text-xs uppercase tracking-wide">{item.label}</span>
                </div>
                <p className="text-2xl font-semibold leading-none">{values[item.valueKey]}</p>
              </div>
              <Badge variant="secondary" className={item.badgeClass}>
                KPI
              </Badge>
            </div>
          </Button>
        );
      })}
    </div>
  );
}
