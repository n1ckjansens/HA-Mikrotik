import { cn } from "@/lib/utils";
import { formatLastSeenLabel } from "@/lib/time";

type Props = {
  online: boolean;
  lastSeenAt: string | null | undefined;
  now: number;
};

export function DeviceLastSeen({ online, lastSeenAt, now }: Props) {
  const label = formatLastSeenLabel(online, lastSeenAt, now);

  return (
    <p
      className={cn(
        "text-xs",
        online ? "text-muted-foreground" : "font-medium text-foreground"
      )}
    >
      {online ? label : `Last seen ${label}`}
    </p>
  );
}
