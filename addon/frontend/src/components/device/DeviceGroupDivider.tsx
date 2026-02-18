import { Card, CardContent } from "@/components/ui/card";

type Props = {
  label: string;
};

export function DeviceGroupDivider({ label }: Props) {
  return (
    <Card>
      <CardContent className="py-3">
        <p className="text-center text-xs text-muted-foreground">{label}</p>
      </CardContent>
    </Card>
  );
}
