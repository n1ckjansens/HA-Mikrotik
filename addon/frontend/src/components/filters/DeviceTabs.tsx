import type { StatusFilter } from "@/hooks/useFilters";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

type Props = {
  value: StatusFilter;
  onChange: (value: StatusFilter) => void;
};

export function DeviceTabs({ value, onChange }: Props) {
  return (
    <Tabs value={value} onValueChange={(next) => onChange(next as StatusFilter)}>
      <TabsList>
        <TabsTrigger value="all">All</TabsTrigger>
        <TabsTrigger value="new">New</TabsTrigger>
        <TabsTrigger value="registered">Registered</TabsTrigger>
      </TabsList>
    </Tabs>
  );
}
