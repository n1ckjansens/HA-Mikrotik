import { RefreshStatus } from "@/components/layout/RefreshStatus";
import { RouterStatusIndicator } from "@/components/layout/RouterStatusIndicator";

type Props = {
  connected: boolean;
  updatedAt: number | undefined;
  now: number;
};

export function Header({ connected, updatedAt, now }: Props) {
  return (
    <header className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
      <div>
        <h1 className="text-2xl font-semibold">MikroTik Presence</h1>
        <p className="text-sm text-muted-foreground">RouterOS v7 device monitoring</p>
      </div>
      <div className="flex flex-wrap items-center gap-3">
        <RouterStatusIndicator connected={connected} />
        <RefreshStatus updatedAt={updatedAt} now={now} />
      </div>
    </header>
  );
}
