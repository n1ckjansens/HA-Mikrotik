import { Badge } from "@/components/ui/badge";

type Props = {
  online: boolean;
};

export function DeviceOnlineBadge({ online }: Props) {
  return online ? <Badge variant="default">Online</Badge> : <Badge variant="outline">Offline</Badge>;
}
