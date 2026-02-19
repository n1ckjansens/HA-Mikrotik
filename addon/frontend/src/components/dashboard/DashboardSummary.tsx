import { Card, CardContent } from "@/components/ui/card";

type Props = {
  online: number;
  offline: number;
  newDevices: number;
  registered: number;
};

const items = [
  { key: "online", label: "Online" },
  { key: "offline", label: "Offline" },
  { key: "newDevices", label: "New" },
  { key: "registered", label: "Registered" }
] as const;

export function DashboardSummary({ online, offline, newDevices, registered }: Props) {
  const values = { online, offline, newDevices, registered };

  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {items.map((item) => (
        <Card key={item.key}>
          <CardContent className="flex items-center justify-between p-4">
            <p className="text-sm text-muted-foreground">{item.label}</p>
            <p className="text-2xl font-semibold">{values[item.key]}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
