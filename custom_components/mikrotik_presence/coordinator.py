"""Data coordinators for MikroTik Presence integration."""

from __future__ import annotations

import asyncio
from datetime import timedelta
import logging

from homeassistant.core import HomeAssistant
from homeassistant.helpers.update_coordinator import DataUpdateCoordinator, UpdateFailed

from .client import CapabilityDTO, DeviceDTO, MikrotikApiError, MikrotikPresenceClient
from .const import DOMAIN, UPDATE_INTERVAL_SECONDS

_LOGGER = logging.getLogger(__name__)


class DevicesCoordinator(DataUpdateCoordinator[dict[str, DeviceDTO]]):
    """Coordinator for registered devices and their capabilities."""

    def __init__(
        self,
        hass: HomeAssistant,
        client: MikrotikPresenceClient,
    ) -> None:
        super().__init__(
            hass,
            _LOGGER,
            name=f"{DOMAIN}_devices",
            update_interval=timedelta(seconds=UPDATE_INTERVAL_SECONDS),
        )
        self._client = client

    async def _async_update_data(self) -> dict[str, DeviceDTO]:
        """Fetch registered devices with capabilities from backend."""
        try:
            devices = await self._client.async_fetch_devices()
            registered_devices = [device for device in devices if device.registered]

            capabilities_batches = await asyncio.gather(
                *(
                    self._client.async_fetch_device_capabilities(device.id)
                    for device in registered_devices
                )
            )

            data: dict[str, DeviceDTO] = {}
            for device, capabilities in zip(registered_devices, capabilities_batches):
                device.capabilities = capabilities
                data[device.id] = device

            return data
        except MikrotikApiError as err:
            raise UpdateFailed(str(err)) from err


class GlobalCoordinator(DataUpdateCoordinator[list[CapabilityDTO]]):
    """Coordinator for global capabilities."""

    def __init__(
        self,
        hass: HomeAssistant,
        client: MikrotikPresenceClient,
    ) -> None:
        super().__init__(
            hass,
            _LOGGER,
            name=f"{DOMAIN}_global",
            update_interval=timedelta(seconds=UPDATE_INTERVAL_SECONDS),
        )
        self._client = client

    async def _async_update_data(self) -> list[CapabilityDTO]:
        """Fetch global capabilities from backend."""
        try:
            return await self._client.async_fetch_global_capabilities()
        except MikrotikApiError as err:
            raise UpdateFailed(str(err)) from err
