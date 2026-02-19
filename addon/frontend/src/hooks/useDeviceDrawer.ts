import { useEffect, useState } from "react";

import { useDevice } from "@/hooks/useDevice";
import { useRegisterDevice } from "@/hooks/useRegisterDevice";
import {
  inferDeviceType,
  parseStoredIcon,
  toStoredIcon,
  type DeviceType
} from "@/lib/device";

export function useDeviceDrawer() {
  const [selectedMac, setSelectedMac] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [icon, setIcon] = useState<DeviceType>("unknown");
  const [comment, setComment] = useState("");

  const detailQuery = useDevice(selectedMac);
  const register = useRegisterDevice();

  useEffect(() => {
    if (!detailQuery.data || detailQuery.data.mac !== selectedMac) {
      return;
    }

    setName(detailQuery.data.name);
    setComment(detailQuery.data.comment ?? "");
    setIcon(parseStoredIcon(detailQuery.data.icon, inferDeviceType(detailQuery.data)));
  }, [detailQuery.data, selectedMac]);

  const open = (mac: string) => {
    setSelectedMac(mac);
    setName("");
    setComment("");
    setIcon("unknown");
  };

  const close = () => {
    if (register.isSaving) {
      return;
    }
    setSelectedMac(null);
  };

  const save = async () => {
    if (!detailQuery.data || name.trim() === "") {
      return;
    }

    await register.saveDevice(detailQuery.data, {
      name,
      icon: toStoredIcon(icon),
      comment
    });
  };

  return {
    open,
    close,
    save,
    device: detailQuery.data ?? null,
    isOpen: selectedMac !== null,
    isLoading: detailQuery.isPending,
    isSaving: register.isSaving,
    name,
    setName,
    icon,
    setIcon,
    comment,
    setComment
  };
}
