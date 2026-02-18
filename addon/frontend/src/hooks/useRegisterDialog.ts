import { useState } from "react";

import { useRegisterDevice } from "@/hooks/useRegisterDevice";
import type { Device } from "@/types/device";

export function useRegisterDialog() {
  const register = useRegisterDevice();
  const [target, setTarget] = useState<Device | null>(null);
  const [name, setName] = useState("");

  const openForDevice = (device: Device) => {
    setTarget(device);
    setName(device.name);
  };

  const close = () => {
    if (register.registerMutation.isPending) {
      return;
    }
    setTarget(null);
    setName("");
  };

  const save = async () => {
    if (!target) {
      return;
    }
    try {
      await register.registerWithName(target.mac, name);
      setTarget(null);
      setName("");
    } catch {
      // Keep the form open so the user can retry or change input.
    }
  };

  return {
    target,
    name,
    setName,
    isOpen: target !== null,
    isSaving: register.registerMutation.isPending,
    openForDevice,
    close,
    save
  };
}
