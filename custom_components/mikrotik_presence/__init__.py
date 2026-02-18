"""MikroTik Presence integration."""

from __future__ import annotations

from homeassistant.config_entries import ConfigEntry
from homeassistant.core import HomeAssistant

from .api import async_register_view
from .const import CONF_VERSION, DOMAIN, EVENT_CONFIG_UPDATED


def _bump_version(hass: HomeAssistant) -> int:
    domain_data = hass.data.setdefault(DOMAIN, {})
    current = domain_data.get(CONF_VERSION, 1)
    current += 1
    domain_data[CONF_VERSION] = current
    return current


async def async_setup(hass: HomeAssistant, config: dict) -> bool:
    """Set up the integration from yaml (unused)."""
    domain_data = hass.data.setdefault(DOMAIN, {})
    domain_data.setdefault(CONF_VERSION, 1)
    await async_register_view(hass, config)
    return True


async def async_setup_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    """Set up integration from a config entry."""

    async def _on_options_update(_: HomeAssistant, updated_entry: ConfigEntry) -> None:
        version = _bump_version(hass)
        hass.bus.async_fire(EVENT_CONFIG_UPDATED, {"version": version, "entry_id": updated_entry.entry_id})

    unsub = entry.add_update_listener(_on_options_update)
    hass.data[DOMAIN][entry.entry_id] = unsub

    version = _bump_version(hass)
    hass.bus.async_fire(EVENT_CONFIG_UPDATED, {"version": version, "entry_id": entry.entry_id})
    return True


async def async_unload_entry(hass: HomeAssistant, entry: ConfigEntry) -> bool:
    """Unload config entry."""
    unsub = hass.data.get(DOMAIN, {}).pop(entry.entry_id, None)
    if unsub is not None:
        unsub()
    version = _bump_version(hass)
    hass.bus.async_fire(EVENT_CONFIG_UPDATED, {"version": version, "entry_id": entry.entry_id})
    return True
