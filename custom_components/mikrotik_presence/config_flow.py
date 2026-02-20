"""Config flow for MikroTik Presence integration."""

from __future__ import annotations

import logging
from typing import Any

import aiohttp
import voluptuous as vol

from homeassistant import config_entries
from homeassistant.components.hassio import AddonError
from homeassistant.data_entry_flow import FlowResult
from homeassistant.helpers.aiohttp_client import async_get_clientsession
from homeassistant.helpers.hassio import is_hassio

from .addon import get_addon_manager
from .const import (
    ADDON_BASE_SLUG,
    ADDON_NAME,
    CONF_API_KEY,
    CONF_BASE_URL,
    CONF_USE_ADDON,
    DOMAIN,
)

_LOGGER = logging.getLogger(__name__)

DEFAULT_BASE_URL = "http://localhost:8080"


class MikrotikPresenceConfigFlow(config_entries.ConfigFlow, domain=DOMAIN):
    """Handle config flow for MikroTik Presence."""

    VERSION = 1

    async def async_step_hassio(self, discovery_info: Any) -> FlowResult:
        """Handle Supervisor discovery."""
        discovery_slug = str(getattr(discovery_info, "slug", "") or "")
        if not _slug_matches(discovery_slug):
            return self.async_abort(reason="not_our_addon")

        return await self.async_step_addon()

    async def async_step_user(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Handle user initiated flow."""
        del user_input

        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        if is_hassio(self.hass):
            return self.async_show_menu(step_id="user", menu_options=["addon", "manual"])

        return await self.async_step_manual()

    async def async_step_addon(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Create an entry in add-on mode.

        Add-on installation/start is executed during integration startup.
        """
        del user_input

        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        if not is_hassio(self.hass):
            return self.async_abort(reason="no_supervisor")

        addon_manager = get_addon_manager(self.hass)
        try:
            await addon_manager.async_resolve_slug()
        except AddonError as err:
            _LOGGER.warning("Cannot resolve add-on in Supervisor Store: %s", err)
            return self.async_abort(reason="addon_not_available")

        await self.async_set_unique_id(DOMAIN)
        self._abort_if_unique_id_configured()

        return self.async_create_entry(
            title=ADDON_NAME,
            data={
                CONF_USE_ADDON: True,
                CONF_BASE_URL: "",
                CONF_API_KEY: None,
            },
        )

    async def async_step_manual(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Handle manual backend configuration."""
        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        errors: dict[str, str] = {}

        if user_input is not None:
            try:
                base_url = _normalize_base_url(user_input.get(CONF_BASE_URL))
                api_key = _normalize_optional_string(user_input.get(CONF_API_KEY))
            except ValueError:
                errors["base"] = "cannot_connect"
            else:
                if await self._async_can_connect(base_url, api_key):
                    await self.async_set_unique_id(DOMAIN)
                    self._abort_if_unique_id_configured()

                    return self.async_create_entry(
                        title=ADDON_NAME,
                        data={
                            CONF_USE_ADDON: False,
                            CONF_BASE_URL: base_url,
                            CONF_API_KEY: api_key,
                        },
                    )
                errors["base"] = "cannot_connect"

        return self.async_show_form(
            step_id="manual",
            data_schema=vol.Schema(
                {
                    vol.Required(
                        CONF_BASE_URL,
                        default=(user_input or {}).get(CONF_BASE_URL, DEFAULT_BASE_URL),
                    ): str,
                    vol.Optional(
                        CONF_API_KEY,
                        default=(user_input or {}).get(CONF_API_KEY, ""),
                    ): str,
                }
            ),
            errors=errors,
        )

    async def _async_can_connect(self, base_url: str, api_key: str | None) -> bool:
        """Try connecting to backend and validating API response format."""
        session = async_get_clientsession(self.hass)
        url = f"{base_url.rstrip('/')}/api/devices"

        headers: dict[str, str] = {}
        if api_key:
            headers["Authorization"] = f"Bearer {api_key}"

        try:
            async with session.get(
                url,
                headers=headers,
                timeout=aiohttp.ClientTimeout(total=10),
            ) as response:
                if response.status != 200:
                    return False

                payload = await response.json(content_type=None)
                return isinstance(payload, (list, dict))
        except (aiohttp.ClientError, TimeoutError, ValueError):
            return False


def _slug_matches(slug: str) -> bool:
    """Return true when slug is base slug or repository-prefixed slug."""
    normalized = str(slug).strip()
    return normalized == ADDON_BASE_SLUG or normalized.endswith(f"_{ADDON_BASE_SLUG}")


def _normalize_base_url(value: Any) -> str:
    """Normalize and validate backend base URL input."""
    base_url = str(value or "").strip().rstrip("/")
    if not base_url:
        raise ValueError("Base URL is empty")
    return base_url


def _normalize_optional_string(value: Any) -> str | None:
    """Normalize optional string field."""
    if value is None:
        return None

    normalized = str(value).strip()
    return normalized or None
