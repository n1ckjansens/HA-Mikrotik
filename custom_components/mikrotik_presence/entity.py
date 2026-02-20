"""Base entity classes for MikroTik Presence integration."""

from __future__ import annotations

from typing import Any

from homeassistant.helpers.device_registry import DeviceInfo
from homeassistant.helpers.update_coordinator import CoordinatorEntity, DataUpdateCoordinator

from .client import CapabilityDTO, DeviceDTO, MikrotikPresenceClient
from .const import DOMAIN


class MikrotikBaseEntity(CoordinatorEntity[Any]):
    """Shared base entity bound to coordinator-backed capability data."""

    def __init__(
        self,
        coordinator: DataUpdateCoordinator[Any],
        client: MikrotikPresenceClient,
        capability: CapabilityDTO,
        device: DeviceDTO | None,
        scope: str,
    ) -> None:
        super().__init__(coordinator)
        self._client = client
        self._scope = scope

        self._device_id = device.id if device is not None else None
        self._capability_id = capability.id

        self._fallback_device_name = device.name if device is not None else None
        self._fallback_device_vendor = device.vendor if device is not None else None
        self._fallback_capability_label = capability.label

    @property
    def name(self) -> str:
        """Return display name for entity."""
        capability_label = (
            self._capability.label if self._capability is not None else self._fallback_capability_label
        )

        if self._scope == "device":
            device_name = (
                self._device.name
                if self._device is not None
                else self._fallback_device_name or self._device_id or "Unknown device"
            )
            return f"{device_name} {capability_label}"

        return capability_label

    @property
    def unique_id(self) -> str:
        """Return stable unique ID."""
        if self._scope == "device" and self._device_id is not None:
            return f"mikrotik_presence_{self._device_id}_{self._capability_id}"
        return f"mikrotik_presence_global_{self._capability_id}"

    @property
    def device_info(self) -> DeviceInfo:
        """Return device registry info."""
        if self._scope == "device" and self._device_id is not None:
            device = self._device
            return DeviceInfo(
                identifiers={(DOMAIN, self._device_id)},
                name=(device.name if device is not None else self._fallback_device_name or self._device_id),
                manufacturer=(
                    device.vendor
                    if device is not None and device.vendor
                    else self._fallback_device_vendor or "MikroTik Client"
                ),
                model="MikroTik Client",
            )

        return DeviceInfo(
            identifiers={(DOMAIN, "hub")},
            name="MikroTik Presence",
            manufacturer="n1ckjansens",
            model="Presence Engine",
        )

    @property
    def available(self) -> bool:
        """Return availability derived from coordinator data."""
        if not self.coordinator.last_update_success:
            return False

        capability = self._capability
        if capability is None:
            return False

        if self._scope == "device":
            device = self._device
            return device is not None and device.online

        return True

    @property
    def _device(self) -> DeviceDTO | None:
        """Return current device from coordinator data."""
        if self._scope != "device" or self._device_id is None:
            return None

        data = self.coordinator.data
        if not isinstance(data, dict):
            return None

        return data.get(self._device_id)

    @property
    def _capability(self) -> CapabilityDTO | None:
        """Return current capability from coordinator data."""
        if self._scope == "device":
            device = self._device
            if device is None:
                return None

            for capability in device.capabilities:
                if capability.id == self._capability_id:
                    return capability
            return None

        data = self.coordinator.data
        if not isinstance(data, list):
            return None

        for capability in data:
            if capability.id == self._capability_id:
                return capability

        return None
