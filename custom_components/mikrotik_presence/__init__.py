"""MikroTik Presence integration setup."""

from __future__ import annotations

import logging

from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant
from homeassistant.helpers.aiohttp_client import async_get_clientsession

from .client import MikrotikPresenceClient
from .const import (
    CONF_API_KEY,
    CONF_BASE_URL,
    DATA_CLIENT,
    DATA_COORDINATOR_DEVICES,
    DATA_COORDINATOR_GLOBAL,
    DOMAIN,
    PLATFORMS,
)
from .coordinator import DevicesCoordinator, GlobalCoordinator

_LOGGER = logging.getLogger(__name__)


async def async_setup_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    """Set up MikroTik Presence from a config entry."""
    hass.data.setdefault(DOMAIN, {})

    base_url = str(entry.options.get(CONF_BASE_URL) or entry.data.get(CONF_BASE_URL) or "").strip()
    api_key = str(entry.options.get(CONF_API_KEY) or entry.data.get(CONF_API_KEY) or "").strip()

    session = async_get_clientsession(hass)
    client = MikrotikPresenceClient(session, base_url, api_key or None)

    devices_coordinator = DevicesCoordinator(hass, client, _LOGGER)
    global_coordinator = GlobalCoordinator(hass, client, _LOGGER)

    await devices_coordinator.async_config_entry_first_refresh()
    await global_coordinator.async_config_entry_first_refresh()

    hass.data[DOMAIN][entry.entry_id] = {
        DATA_CLIENT: client,
        DATA_COORDINATOR_DEVICES: devices_coordinator,
        DATA_COORDINATOR_GLOBAL: global_coordinator,
    }

    entry.async_on_unload(entry.add_update_listener(async_reload_entry))
    await hass.config_entries.async_forward_entry_setups(entry, PLATFORMS)
    return True


async def async_unload_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    """Unload a config entry."""
    unload_ok = await hass.config_entries.async_unload_platforms(entry, PLATFORMS)
    if unload_ok:
        hass.data[DOMAIN].pop(entry.entry_id, None)
        if not hass.data[DOMAIN]:
            hass.data.pop(DOMAIN, None)
    return unload_ok


async def async_reload_entry(hass: HomeAssistant, entry: ConfigEntry) -> None:
    """Reload config entry when options are updated."""
    await hass.config_entries.async_reload(entry.entry_id)
