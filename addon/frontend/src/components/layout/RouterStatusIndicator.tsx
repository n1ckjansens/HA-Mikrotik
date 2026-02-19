import { cn } from "@/lib/utils";

type Props = {
  connected: boolean;
};

export function RouterStatusIndicator({ connected }: Props) {
  return (
    <div
      className="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm"
      aria-label={connected ? "Router connected" : "Router disconnected"}
    >
      <span
        className={cn(
          "h-2.5 w-2.5 rounded-full",
          connected ? "bg-emerald-500" : "bg-destructive"
        )}
        aria-hidden
      />
      <span>{connected ? "Connected" : "Disconnected"}</span>
    </div>
  );
}
