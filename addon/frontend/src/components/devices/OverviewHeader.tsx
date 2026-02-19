import { useEffect, type RefObject } from "react";
import { Pause, Play, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useNow } from "@/hooks/useNow";
import { formatUpdatedAgo } from "@/lib/time";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from "@/components/ui/tooltip";

type Props = {
  connected: boolean;
  updatedAt: number | undefined;
  isPaused: boolean;
  onTogglePause: () => void;
  onRefresh: () => void;
  isRefreshing: boolean;
  lastSuccessfulLabel: string;
  searchInputRef: RefObject<HTMLInputElement>;
};

export function OverviewHeader({
  connected,
  updatedAt,
  isPaused,
  onTogglePause,
  onRefresh,
  isRefreshing,
  lastSuccessfulLabel,
  searchInputRef
}: Props) {
  const now = useNow(1000);
  const updatedLabel = formatUpdatedAgo(updatedAt, now);
  const status = isPaused ? "paused" : connected ? "connected" : "disconnected";

  useEffect(() => {
    const handleKeydown = (event: KeyboardEvent) => {
      if (!(event.metaKey || event.ctrlKey)) {
        return;
      }
      if (event.key.toLowerCase() !== "k") {
        return;
      }

      event.preventDefault();
      searchInputRef.current?.focus();
      searchInputRef.current?.select();
    };

    window.addEventListener("keydown", handleKeydown);
    return () => {
      window.removeEventListener("keydown", handleKeydown);
    };
  }, [searchInputRef]);

  return (
    <TooltipProvider>
      <header className="flex flex-col gap-3 pt-0 pb-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-wrap items-center gap-3 md:flex-nowrap">
          <div className="flex items-center gap-2 text-sm">
            <span className="relative flex h-2 w-2 items-center justify-center">
              <span
                className={`absolute inline-flex h-full w-full rounded-full ${
                  status === "connected"
                    ? "animate-[ping_2.6s_ease-in-out_infinite] bg-emerald-500/50"
                    : "bg-transparent"
                }`}
                aria-hidden
              />
              <span
                className={`relative inline-flex h-2 w-2 rounded-full ${
                  status === "connected"
                    ? "bg-emerald-500"
                    : status === "paused"
                      ? "bg-amber-500"
                      : "bg-destructive"
                }`}
                aria-hidden
              />
            </span>
            {status === "connected"
              ? "Connected"
              : status === "paused"
                ? "Paused"
                : "Disconnected"}
          </div>

          <Tooltip>
            <TooltipTrigger asChild>
              <span className="w-32 text-sm text-muted-foreground tabular-nums">
                {updatedLabel}
              </span>
            </TooltipTrigger>
            <TooltipContent>Last successful sync: {lastSuccessfulLabel}</TooltipContent>
          </Tooltip>

          <Separator orientation="vertical" className="hidden h-6 opacity-40 md:block" />

          <Button variant="outline" size="sm" onClick={onTogglePause} className="h-9">
            {isPaused ? <Play className="mr-2 h-4 w-4" /> : <Pause className="mr-2 h-4 w-4" />}
            {isPaused ? "Resume" : "Pause"}
          </Button>

          <Button
            variant="outline"
            size="sm"
            className="h-9"
            onClick={onRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        </div>
      </header>
    </TooltipProvider>
  );
}
