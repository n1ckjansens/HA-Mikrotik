import { formatUpdatedAgo } from "@/lib/time";

type Props = {
  updatedAt: number | undefined;
  now: number;
};

export function RefreshStatus({ updatedAt, now }: Props) {
  return (
    <p className="text-sm text-muted-foreground" aria-live="polite">
      {formatUpdatedAgo(updatedAt, now)}
    </p>
  );
}
