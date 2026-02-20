"""Config flow for MikroTik Presence integration."""

from __future__ import annotations

import asyncio
import logging
from typing import Any

import aiohttp
import voluptuous as vol

from homeassistant import config_entries
from homeassistant.components.hassio import AddonError, is_hassio
from homeassistant.core import HomeAssistant
from homeassistant.data_entry_flow import FlowResult
from homeassistant.helpers.aiohttp_client import async_get_clientsession

from .addon import get_addon_manager
from .const import ADDON_NAME, ADDON_SLUG, CONF_API_KEY, CONF_BASE_URL, DOMAIN

_LOGGER = logging.getLogger(__name__)

DEFAULT_BASE_URL = "http://homeassistant:8080"
ADDON_ENTRY_TITLE = f"{ADDON_NAME} (addon)"


class MikrotikPresenceConfigFlow(config_entries.ConfigFlow, domain=DOMAIN):
    """Handle config flow for MikroTik Presence."""

    VERSION = 1

    async def async_step_hassio(self, discovery_info: Any) -> FlowResult:
        """Handle Supervisor discovery."""
        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        discovery_slug = getattr(discovery_info, "slug", None)
        if discovery_slug != ADDON_SLUG:
            return self.async_abort(reason="not_our_addon")

        discovery_config = getattr(discovery_info, "config", None)
        base_url = _build_base_url(
            discovery_config.get("host") if isinstance(discovery_config, dict) else None,
            discovery_config.get("port") if isinstance(discovery_config, dict) else None,
        )
        if base_url is None:
            return self.async_abort(reason="invalid_discovery")

        if not await self._async_can_connect(base_url=base_url, api_key=None):
            return self.async_abort(reason="cannot_connect")

        return await self._async_create_config_entry(
            title=ADDON_ENTRY_TITLE,
            base_url=base_url,
            api_key=None,
        )

    async def async_step_user(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Handle user initiated flow."""
        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        if user_input is not None:
            return await self.async_step_manual(user_input)

        if not is_hassio(self.hass):
            return await self.async_step_manual()

        addon_manager = get_addon_manager(self.hass)
        is_installed = await self._async_is_addon_installed(addon_manager)
        if is_installed:
            base_url = await self._async_get_addon_base_url(addon_manager)
            if base_url is not None and await self._async_can_connect(base_url=base_url, api_key=None):
                return await self._async_create_config_entry(
                    title=ADDON_ENTRY_TITLE,
                    base_url=base_url,
                    api_key=None,
                )

        if not is_installed:
            return self.async_show_menu(step_id="user", menu_options=["addon_install", "manual"])

        return self.async_show_menu(step_id="user", menu_options=["addon", "manual"])

    async def async_step_addon(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Start installed add-on and connect automatically."""
        del user_input

        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        if not is_hassio(self.hass):
            return self.async_abort(reason="no_supervisor")

        addon_manager = get_addon_manager(self.hass)
        if not await self._async_is_addon_installed(addon_manager):
            return self.async_abort(reason="addon_not_installed")

        try:
            await addon_manager.async_schedule_start_addon()
        except AddonError as err:
            _LOGGER.warning("Failed to start addon %s: %s", ADDON_SLUG, err)
            return self.async_abort(reason="addon_operation_failed")

        base_url = await self._async_wait_for_addon_base_url(addon_manager)
        if base_url is None:
            return self.async_abort(reason="cannot_connect")

        if not await self._async_can_connect(base_url=base_url, api_key=None):
            return self.async_abort(reason="cannot_connect")

        return await self._async_create_config_entry(
            title=ADDON_ENTRY_TITLE,
            base_url=base_url,
            api_key=None,
        )

    async def async_step_addon_install(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Install and start add-on, then connect automatically."""
        del user_input

        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        if not is_hassio(self.hass):
            return self.async_abort(reason="no_supervisor")

        addon_manager = get_addon_manager(self.hass)
        try:
            if not await self._async_is_addon_installed(addon_manager):
                await addon_manager.async_schedule_install_addon()
            await addon_manager.async_schedule_start_addon()
        except AddonError as err:
            _LOGGER.warning("Failed to install/start addon %s: %s", ADDON_SLUG, err)
            return self.async_abort(reason="addon_operation_failed")

        base_url = await self._async_wait_for_addon_base_url(addon_manager)
        if base_url is None:
            return self.async_abort(reason="cannot_connect")

        if not await self._async_can_connect(base_url=base_url, api_key=None):
            return self.async_abort(reason="cannot_connect")

        return await self._async_create_config_entry(
            title=ADDON_ENTRY_TITLE,
            base_url=base_url,
            api_key=None,
        )

    async def async_step_manual(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Manual backend configuration."""
        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        errors: dict[str, str] = {}

        if user_input is not None:
            try:
                base_url = _normalize_base_url(user_input.get(CONF_BASE_URL, ""))
                api_key = _normalize_optional_string(user_input.get(CONF_API_KEY))
            except ValueError:
                errors["base"] = "cannot_connect"
            else:
                if await self._async_can_connect(base_url=base_url, api_key=api_key):
                    return await self._async_create_config_entry(
                        title=ADDON_NAME,
                        base_url=base_url,
                        api_key=api_key,
                    )
                errors["base"] = "cannot_connect"

        return self.async_show_form(
            step_id="manual",
            data_schema=vol.Schema(
                {
                    vol.Required(CONF_BASE_URL, default=DEFAULT_BASE_URL): str,
                    vol.Optional(CONF_API_KEY, default=""): str,
                }
            ),
            errors=errors,
        )

    async def _async_create_config_entry(
        self,
        title: str,
        base_url: str,
        api_key: str | None,
    ) -> FlowResult:
        """Create config entry with uniqueness checks."""
        if self._async_current_entries():
            return self.async_abort(reason="already_configured")

        await self.async_set_unique_id(DOMAIN)
        self._abort_if_unique_id_configured()

        return self.async_create_entry(
            title=title,
            data={
                CONF_BASE_URL: base_url,
                CONF_API_KEY: api_key,
            },
        )

    async def _async_is_addon_installed(self, addon_manager: Any) -> bool:
        """Return whether addon is installed."""
        try:
            return bool(await addon_manager.async_is_installed())
        except AddonError as err:
            _LOGGER.debug("Cannot determine add-on installation state for %s: %s", ADDON_SLUG, err)
            return False

    async def _async_get_addon_info(self, addon_manager: Any) -> dict[str, Any] | None:
        """Return add-on info dict or None when unavailable."""
        try:
            info = await addon_manager.async_get_addon_info()
        except AddonError as err:
            _LOGGER.debug("Cannot load addon info for %s: %s", ADDON_SLUG, err)
            return None
        if isinstance(info, dict):
            return info
        return None

    async def _async_get_addon_base_url(self, addon_manager: Any) -> str | None:
        """Return backend base URL from add-on discovery info."""
        addon_info = await self._async_get_addon_info(addon_manager)
        if addon_info is not None and not addon_info.get("started", True):
            return None

        try:
            discovery = await addon_manager.async_get_addon_discovery_info()
        except AddonError as err:
            _LOGGER.debug("Cannot load addon discovery info for %s: %s", ADDON_SLUG, err)
            return None

        if not isinstance(discovery, dict):
            return None

        return _build_base_url(discovery.get("host"), discovery.get("port"))

    async def _async_wait_for_addon_base_url(
        self,
        addon_manager: Any,
        attempts: int = 15,
        delay_seconds: float = 2.0,
    ) -> str | None:
        """Wait until add-on exposes discovery info and backend responds."""
        for _ in range(attempts):
            base_url = await self._async_get_addon_base_url(addon_manager)
            if base_url is not None:
                return base_url
            await asyncio.sleep(delay_seconds)

        return None

    async def _async_can_connect(self, base_url: str, api_key: str | None) -> bool:
        """Try to connect to backend and validate API response."""
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
        except (aiohttp.ClientError, asyncio.TimeoutError, ValueError):
            return False


def _build_base_url(host: Any, port: Any) -> str | None:
    """Build backend URL from discovery host/port values."""
    host_value = _normalize_optional_string(host)
    if not host_value:
        return None

    port_value: int | None = None
    if isinstance(port, int):
        port_value = port
    elif isinstance(port, str):
        stripped = port.strip()
        if stripped.isdigit():
            port_value = int(stripped)

    if port_value is None or port_value <= 0:
        return None

    return f"http://{host_value}:{port_value}"


def _normalize_base_url(raw_base_url: Any) -> str:
    """Normalize backend URL from user input."""
    base_url = str(raw_base_url or "").strip().rstrip("/")
    if not base_url:
        raise ValueError("Base URL is empty")
    return base_url


def _normalize_optional_string(value: Any) -> str | None:
    """Normalize optional string field."""
    if value is None:
        return None
    normalized = str(value).strip()
    return normalized or None
