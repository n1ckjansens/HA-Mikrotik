import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ResponsiveOverlay } from "@/components/ui/responsive-overlay";

import type { Device } from "@/types/device";

type Props = {
  device: Device | null;
  isMobile: boolean;
  open: boolean;
  name: string;
  isSaving: boolean;
  onOpenChange: (open: boolean) => void;
  onNameChange: (value: string) => void;
  onSave: () => void;
};

function PromptBody({
  device,
  name,
  isSaving,
  onNameChange,
  onSave,
  onCancel
}: {
  device: Device | null;
  name: string;
  isSaving: boolean;
  onNameChange: (value: string) => void;
  onSave: () => void;
  onCancel: () => void;
}) {
  return (
    <>
      <div className="grid gap-2">
        <p className="text-sm text-muted-foreground">MAC: {device?.mac ?? ""}</p>
        <Input value={name} onChange={(event) => onNameChange(event.target.value)} />
      </div>
      <div className="flex items-center justify-end gap-2">
        <Button variant="outline" onClick={onCancel} disabled={isSaving}>
          Cancel
        </Button>
        <Button onClick={onSave} disabled={isSaving || name.trim() === ""}>
          Save
        </Button>
      </div>
    </>
  );
}

export function RegisterDevicePrompt({
  device,
  isMobile,
  open,
  name,
  isSaving,
  onOpenChange,
  onNameChange,
  onSave
}: Props) {
  return (
    <ResponsiveOverlay
      open={open}
      isMobile={isMobile}
      onOpenChange={onOpenChange}
      title="Register device"
      description="Set a display name for this device."
    >
      <PromptBody
        device={device}
        name={name}
        isSaving={isSaving}
        onNameChange={onNameChange}
        onSave={onSave}
        onCancel={() => onOpenChange(false)}
      />
    </ResponsiveOverlay>
  );
}
