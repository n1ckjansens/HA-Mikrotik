import { Badge } from "@/components/ui/badge";

type Props = {
  online: boolean;
};

export function DeviceStatusBadge({ online }: Props) {
  if (online) {
    return (
      <Badge
        variant="outline"
        className="border-emerald-500/30 bg-emerald-500/10 text-emerald-700"
        aria-label="Device online"
      >
        Online
      </Badge>
    );
  }

  return (
    <Badge variant="outline" className="text-muted-foreground" aria-label="Device offline">
      Offline
    </Badge>
  );
}
