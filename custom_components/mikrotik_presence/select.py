"""Select platform for MikroTik Presence capabilities."""

from __future__ import annotations

from homeassistant.components.select import SelectEntity
from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant, callback
from homeassistant.exceptions import HomeAssistantError
from homeassistant.helpers.entity_platform import AddEntitiesCallback

from .client import MikrotikPresenceClient, MikrotikApiError
from .const import DATA_CLIENT, DATA_COORDINATOR_DEVICES, DATA_COORDINATOR_GLOBAL, DOMAIN
from .coordinator import DevicesCoordinator, GlobalCoordinator
from .entity import MikrotikBaseEntity


async def async_setup_entry(
    hass: HomeAssistant,
    entry: ConfigEntry,
    async_add_entities: AddEntitiesCallback,
) -> None:
    """Set up select entities from a config entry."""
    data = hass.data[DOMAIN][entry.entry_id]
    client: MikrotikPresenceClient = data[DATA_CLIENT]
    devices_coordinator: DevicesCoordinator = data[DATA_COORDINATOR_DEVICES]
    global_coordinator: GlobalCoordinator = data[DATA_COORDINATOR_GLOBAL]

    known_unique_ids: set[str] = set()

    def _discover_new_entities() -> list[MikrotikCapabilitySelect]:
        entities: list[MikrotikCapabilitySelect] = []

        for device in (devices_coordinator.data or {}).values():
            for capability in device.capabilities:
                if capability.control.type != "select":
                    continue

                entity = MikrotikCapabilitySelect(
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
            if capability.control.type != "select":
                continue

            entity = MikrotikCapabilitySelect(
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


class MikrotikCapabilitySelect(MikrotikBaseEntity, SelectEntity):
    """Select entity for capability with select control."""

    _attr_should_poll = False

    @property
    def options(self) -> list[str]:
        """Return selectable labels."""
        capability = self._capability
        if capability is None:
            return []
        return [option.label for option in capability.control.options]

    @property
    def current_option(self) -> str | None:
        """Return currently selected label."""
        capability = self._capability
        if capability is None:
            return None

        for option in capability.control.options:
            if option.value == capability.state:
                return option.label
        return None

    async def async_select_option(self, option: str) -> None:
        """Set selected option through backend PATCH endpoint."""
        capability = self._capability
        if capability is None:
            raise HomeAssistantError("Capability data is unavailable")

        value = next(
            (
                item.value
                for item in capability.control.options
                if item.label == option or item.value == option
            ),
            None,
        )
        if value is None:
            raise HomeAssistantError(f"Unknown option: {option}")

        try:
            if self._scope == "device":
                if self._device_id is None:
                    raise HomeAssistantError("Device id is missing")
                await self._client.async_set_device_capability_state(
                    self._device_id,
                    self._capability_id,
                    value,
                )
            else:
                await self._client.async_set_global_capability_state(self._capability_id, value)
        except MikrotikApiError as err:
            raise HomeAssistantError(str(err)) from err

        await self.coordinator.async_request_refresh()
