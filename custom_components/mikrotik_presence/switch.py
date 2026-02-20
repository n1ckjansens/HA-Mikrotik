"""Switch platform for MikroTik Presence capabilities."""

from __future__ import annotations

from typing import Any

from homeassistant.components.switch import SwitchEntity
from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant, callback
from homeassistant.exceptions import HomeAssistantError
from homeassistant.helpers.entity_platform import AddEntitiesCallback

from .client import MikrotikApiError, MikrotikPresenceClient
from .const import DATA_CLIENT, DATA_COORDINATOR_DEVICES, DATA_COORDINATOR_GLOBAL, DOMAIN
from .coordinator import DevicesCoordinator, GlobalCoordinator
from .entity import MikrotikBaseEntity


async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up switch entities from a config entry."""
    data = hass.data[DOMAIN][entry.entry_id]
    client: MikrotikPresenceClient = data[DATA_CLIENT]
    devices_coordinator: DevicesCoordinator = data[DATA_COORDINATOR_DEVICES]
    global_coordinator: GlobalCoordinator = data[DATA_COORDINATOR_GLOBAL]

    known_unique_ids: set[str] = set()

    def _discover_new_entities() -> list[MikrotikCapabilitySwitch]:
        entities: list[MikrotikCapabilitySwitch] = []

        for device in (devices_coordinator.data or {}).values():
            for capability in device.capabilities:
                if capability.control.type != "switch":
                    continue

                entity = MikrotikCapabilitySwitch(
                    devices_coordinator,
                    client,
                    capability=capability,
                    device=device,
                    scope="device",
                )
                if entity.unique_id in known_unique_ids:
                    continue

                known_unique_ids.add(entity.unique_id)
                entities.append(entity)

        for capability in global_coordinator.data or []:
            if capability.control.type != "switch":
                continue

            entity = MikrotikCapabilitySwitch(
                global_coordinator,
                client,
                capability=capability,
                device=None,
                scope="global",
            )
            if entity.unique_id in known_unique_ids:
                continue

            known_unique_ids.add(entity.unique_id)
            entities.append(entity)

        return entities

    initial_entities = _discover_new_entities()
    if initial_entities:
        async_add_entities(initial_entities, update_before_add=True)

    @callback
    def _handle_coordinator_update() -> None:
        new_entities = _discover_new_entities()
        if new_entities:
            async_add_entities(new_entities, update_before_add=True)

    remove_device_listener = devices_coordinator.async_add_listener(_handle_coordinator_update)
    remove_global_listener = global_coordinator.async_add_listener(_handle_coordinator_update)
    entry.async_on_unload(remove_device_listener)
    entry.async_on_unload(remove_global_listener)


class MikrotikCapabilitySwitch(MikrotikBaseEntity, SwitchEntity):
    """Switch entity for capability with switch control."""

    _attr_should_poll = False

    @property
    def is_on(self) -> bool:
        """Return true if capability is currently active."""
        capability = self._capability
        if capability is None:
            return False

        return capability.state.strip().lower() in {"on", "enabled", "allow", "true", "1"}

    async def async_turn_on(self, **kwargs: Any) -> None:
        """Turn switch on."""
        del kwargs
        await self._async_set_state(self._resolve_target_state(turn_on=True))

    async def async_turn_off(self, **kwargs: Any) -> None:
        """Turn switch off."""
        del kwargs
        await self._async_set_state(self._resolve_target_state(turn_on=False))

    async def _async_set_state(self, state: str) -> None:
        """Update capability state in backend and refresh coordinator."""
        try:
            if self._scope == "device":
                if self._device_id is None:
                    raise HomeAssistantError("Device id is missing")
                await self._client.async_set_device_capability_state(
                    self._device_id,
                    self._capability_id,
                    state,
                )
            else:
                await self._client.async_set_global_capability_state(self._capability_id, state)
        except MikrotikApiError as err:
            raise HomeAssistantError(str(err)) from err

        await self.coordinator.async_request_refresh()

    def _resolve_target_state(self, turn_on: bool) -> str:
        """Pick backend state value for on/off based on capability options."""
        capability = self._capability
        if capability is None or not capability.control.options:
            return "on" if turn_on else "off"

        normalized_options = {
            option.value.strip().lower(): option.value for option in capability.control.options
        }

        if turn_on:
            for candidate in ("on", "enabled", "allow", "true", "1"):
                if candidate in normalized_options:
                    return normalized_options[candidate]
            return capability.control.options[0].value

        for candidate in ("off", "disabled", "deny", "false", "0"):
            if candidate in normalized_options:
                return normalized_options[candidate]
        return capability.control.options[-1].value
