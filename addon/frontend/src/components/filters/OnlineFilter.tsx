import type { OnlineFilter as OnlineFilterValue } from "@/types/device";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

type Props = {
  value: OnlineFilterValue;
  onChange: (value: OnlineFilterValue) => void;
};

export function OnlineFilter({ value, onChange }: Props) {
  return (
    <ToggleGroup
      type="single"
      variant="outline"
      value={value}
      onValueChange={(next) => {
        if (next === "all" || next === "online" || next === "offline") {
          onChange(next);
        }
      }}
    >
      <ToggleGroupItem value="all" aria-label="Any status">
        Any
      </ToggleGroupItem>
      <ToggleGroupItem value="online" aria-label="Online only">
        Online
      </ToggleGroupItem>
      <ToggleGroupItem value="offline" aria-label="Offline only">
        Offline
      </ToggleGroupItem>
    </ToggleGroup>
  );
}
