import { Layers, UserPlus, X } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

type Props = {
  selectedCount: number;
  selectableCount: number;
  onRegisterSelected: () => void;
  onClearSelection: () => void;
  isPending: boolean;
};

export function BulkActionsBar({
  selectedCount,
  selectableCount,
  onRegisterSelected,
  onClearSelection,
  isPending
}: Props) {
  if (selectedCount === 0) {
    return null;
  }

  return (
    <Alert>
      <Layers className="h-4 w-4" />
      <AlertTitle>{selectedCount} devices selected</AlertTitle>
      <AlertDescription className="flex flex-wrap items-center gap-2">
        <span>{selectableCount} can be bulk registered.</span>
        <Button
          size="sm"
          onClick={onRegisterSelected}
          disabled={selectableCount === 0 || isPending}
        >
          <UserPlus className="mr-2 h-4 w-4" /> Register selected
        </Button>
        <Button variant="ghost" size="sm" onClick={onClearSelection}>
          <X className="mr-2 h-4 w-4" /> Clear selection
        </Button>
      </AlertDescription>
    </Alert>
  );
}
