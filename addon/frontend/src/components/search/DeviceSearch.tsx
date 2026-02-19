import { Search } from "lucide-react";

import { Input } from "@/components/ui/input";

type Props = {
  value: string;
  count: number;
  onChange: (value: string) => void;
};

export function DeviceSearch({ value, count, onChange }: Props) {
  return (
    <div className="relative">
      <Search
        className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
        aria-hidden
      />
      <Input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder="Search devices by name, MAC, vendor, IP"
        className="pr-28 pl-9"
      />
      <p className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
        {count} devices
      </p>
    </div>
  );
}
