import { Input } from "@/components/ui/input";

type Props = {
  value: string;
  onChange: (value: string) => void;
};

export function DeviceSearch({ value, onChange }: Props) {
  return (
    <Input
      value={value}
      onChange={(event) => onChange(event.target.value)}
      placeholder="Search by name, MAC, vendor, IP"
    />
  );
}
