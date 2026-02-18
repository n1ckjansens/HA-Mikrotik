import { Badge } from "@/components/ui/badge";
import type { DeviceStatus } from "@/types/device";

type Props = {
  status: DeviceStatus;
};

export function DeviceStatusBadge({ status }: Props) {
  if (status === "registered") {
    return <Badge variant="default">Registered</Badge>;
  }
  return <Badge variant="secondary">New</Badge>;
}
