"""MikroTik Presence integration setup."""

from __future__ import annotations

from homeassistant.components.hassio import AddonError
from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant
from homeassistant.exceptions import ConfigEntryNotReady
from homeassistant.helpers.aiohttp_client import async_get_clientsession
from homeassistant.helpers.hassio import is_hassio

from .addon import get_addon_manager
from .client import MikrotikPresenceClient
from .const import (
    CONF_API_KEY,
    CONF_BASE_URL,
    CONF_USE_ADDON,
    DATA_CLIENT,
    DATA_COORDINATOR_DEVICES,
    DATA_COORDINATOR_GLOBAL,
    DOMAIN,
    PLATFORMS,
)
from .coordinator import DevicesCoordinator, GlobalCoordinator


async def async_setup_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    """Set up MikroTik Presence from a config entry."""
    hass.data.setdefault(DOMAIN, {})

    use_addon = bool(entry.data.get(CONF_USE_ADDON, is_hassio(hass)))
    base_url: str

    if use_addon:
        if not is_hassio(hass):
            raise ConfigEntryNotReady("Home Assistant Supervisor is required for add-on mode")

        addon_manager = get_addon_manager(hass)
        try:
            await addon_manager.async_ensure_installed_and_running()
            base_url = await addon_manager.async_get_backend_base_url()
        except AddonError as err:
            raise ConfigEntryNotReady(f"Failed to prepare add-on: {err}") from err
    else:
        raw_base_url = entry.data.get(CONF_BASE_URL)
        if not isinstance(raw_base_url, str) or not raw_base_url.strip():
            raise ConfigEntryNotReady("Backend URL is missing")
        base_url = raw_base_url.strip().rstrip("/")

    raw_api_key = entry.data.get(CONF_API_KEY)
    api_key = raw_api_key.strip() if isinstance(raw_api_key, str) and raw_api_key.strip() else None

    session = async_get_clientsession(hass)
    client = MikrotikPresenceClient(session, base_url, api_key)

    devices_coordinator = DevicesCoordinator(hass, client)
    global_coordinator = GlobalCoordinator(hass, client)

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
