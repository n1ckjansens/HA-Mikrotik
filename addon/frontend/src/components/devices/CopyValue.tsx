import type { MouseEvent } from "react";
import { Copy } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { copyTextToClipboard } from "@/lib/clipboard";
import { cn } from "@/lib/utils";

type CopyValueProps = {
  label: string;
  value: string | null | undefined;
  placeholder?: string;
  mono?: boolean;
  className?: string;
  onClick?: (event: MouseEvent<HTMLButtonElement>) => void;
};

export function CopyValue({
  label,
  value,
  placeholder = "-",
  mono = false,
  className,
  onClick
}: CopyValueProps) {
  const normalized = value?.trim() ? value : null;

  const handleClick = async (event: MouseEvent<HTMLButtonElement>) => {
    onClick?.(event);
    if (!normalized) {
      return;
    }

    const copied = await copyTextToClipboard(normalized);
    if (copied) {
      toast.success(`${label} copied`);
    } else {
      toast.error(`Unable to copy ${label.toLowerCase()}`);
    }
  };

  return (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      className={cn(
        "h-auto max-w-full justify-between gap-2 px-2 py-1 text-left font-normal",
        className,
        !normalized && "cursor-default text-muted-foreground hover:bg-transparent"
      )}
      onClick={(event) => {
        void handleClick(event);
      }}
      aria-label={normalized ? `Copy ${label}` : `${label} unavailable`}
    >
      <span className={cn("truncate", mono && "font-mono")}>{normalized ?? placeholder}</span>
      <Copy className="h-3.5 w-3.5 shrink-0 opacity-60 pointer-events-none" aria-hidden="true" />
    </Button>
  );
}
