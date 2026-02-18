import { Badge } from "@/components/ui/badge";
import type { DeviceStatus } from "@/types/device";

type Props = {
  status: DeviceStatus;
};

export function DeviceStatusBadge({ status }: Props) {
  if (status !== "new") {
    return null;
  }
  return <Badge variant="secondary">New</Badge>;
}
