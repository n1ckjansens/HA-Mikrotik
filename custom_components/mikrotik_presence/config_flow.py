"""Config flow for MikroTik Presence integration."""

from __future__ import annotations

from collections.abc import Iterable
from typing import Any
import logging
import os

import aiohttp
import voluptuous as vol

from homeassistant import config_entries
from homeassistant.core import HomeAssistant, callback
from homeassistant.data_entry_flow import AbortFlow, FlowResult
from homeassistant.helpers.aiohttp_client import async_get_clientsession

from .client import MikrotikApiError, MikrotikAuthError, MikrotikPresenceClient
from .const import (
    AUTO_DISCOVERY_CANDIDATES,
    CONF_API_KEY,
    CONF_BASE_URL,
    DEFAULT_BACKEND_PORT,
    DOMAIN,
    SUPERVISOR_ADDON_SLUG,
)

_LOGGER = logging.getLogger(__name__)


class MikrotikPresenceConfigFlow(config_entries.ConfigFlow, domain=DOMAIN):
    """Handle a config flow for MikroTik Presence."""

    VERSION = 1

    def __init__(self) -> None:
        """Initialize config flow state."""
        self._autodiscovered_url: str | None = None

    async def async_step_user(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Handle the initial step."""
        errors: dict[str, str] = {}

        if user_input is not None:
            raw_base_url = user_input.get(CONF_BASE_URL, "")
            api_key = _normalize_optional_string(user_input.get(CONF_API_KEY))
            try:
                if str(raw_base_url).strip():
                    base_url = _normalize_base_url(raw_base_url)
                else:
                    discovered_url = await _async_discover_backend_url(self.hass, api_key)
                    if discovered_url is None:
                        raise CannotConnect("Cannot auto-discover backend URL")
                    base_url = discovered_url

                await self.async_set_unique_id(base_url.lower())
                self._abort_if_unique_id_configured()

                await _async_validate_connection(self.hass, base_url, api_key)
            except InvalidAuth:
                errors["base"] = "invalid_auth"
            except CannotConnect:
                errors["base"] = "cannot_connect"
            except AbortFlow:
                raise
            except Exception:
                _LOGGER.exception("Unexpected error while validating MikroTik Presence connection")
                errors["base"] = "unknown"
            else:
                return self.async_create_entry(
                    title="MikroTik Presence",
                    data={
                        CONF_BASE_URL: base_url,
                        CONF_API_KEY: api_key or "",
                    },
                )

        if self._autodiscovered_url is None:
            self._autodiscovered_url = await _async_discover_backend_url(self.hass)

        data_schema = vol.Schema(
            {
                vol.Required(CONF_BASE_URL): str,
                vol.Optional(CONF_API_KEY, default=""): str,
            }
        )
        if self._autodiscovered_url:
            data_schema = self.add_suggested_values_to_schema(
                data_schema,
                {CONF_BASE_URL: self._autodiscovered_url},
            )

        return self.async_show_form(
            step_id="user",
            data_schema=data_schema,
            errors=errors,
        )

    @staticmethod
    @callback
    def async_get_options_flow(config_entry: config_entries.ConfigEntry) -> config_entries.OptionsFlow:
        """Create options flow handler."""
        return MikrotikPresenceOptionsFlow(config_entry)


class MikrotikPresenceOptionsFlow(config_entries.OptionsFlow):
    """Handle options flow for MikroTik Presence."""

    def __init__(self, config_entry: config_entries.ConfigEntry) -> None:
        self._config_entry = config_entry

    async def async_step_init(self, user_input: dict[str, Any] | None = None) -> FlowResult:
        """Manage integration options."""
        errors: dict[str, str] = {}

        current_base_url = str(
            self._config_entry.options.get(CONF_BASE_URL)
            or self._config_entry.data.get(CONF_BASE_URL)
            or ""
        )
        current_api_key = str(
            self._config_entry.options.get(CONF_API_KEY)
            or self._config_entry.data.get(CONF_API_KEY)
            or ""
        )

        if user_input is not None:
            raw_base_url = user_input.get(CONF_BASE_URL, "")
            raw_api_key = user_input.get(CONF_API_KEY)
            api_key = _normalize_optional_string(raw_api_key)
            try:
                if str(raw_base_url).strip():
                    base_url = _normalize_base_url(raw_base_url)
                else:
                    discovered_url = await _async_discover_backend_url(self.hass, api_key)
                    if discovered_url is None:
                        raise CannotConnect("Cannot auto-discover backend URL")
                    base_url = discovered_url
                await _async_validate_connection(self.hass, base_url, api_key)
            except InvalidAuth:
                errors["base"] = "invalid_auth"
            except CannotConnect:
                errors["base"] = "cannot_connect"
            except Exception:
                _LOGGER.exception("Unexpected error while validating MikroTik Presence options")
                errors["base"] = "unknown"
            else:
                return self.async_create_entry(
                    title="",
                    data={
                        CONF_BASE_URL: base_url,
                        CONF_API_KEY: api_key or "",
                    },
                )

            current_base_url = str(raw_base_url).strip() or current_base_url
            current_api_key = str(raw_api_key or "").strip()

        return self.async_show_form(
            step_id="init",
            data_schema=vol.Schema(
                {
                    vol.Required(CONF_BASE_URL, default=current_base_url): str,
                    vol.Optional(CONF_API_KEY, default=current_api_key): str,
                }
            ),
            errors=errors,
        )


class CannotConnect(Exception):
    """Error to indicate we cannot connect."""


class InvalidAuth(Exception):
    """Error to indicate there is invalid auth."""


async def _async_validate_connection(
    hass: HomeAssistant,
    base_url: str,
    api_key: str | None,
) -> None:
    """Validate backend connection by calling devices endpoint."""
    session = async_get_clientsession(hass)
    client = MikrotikPresenceClient(session, base_url, api_key)

    try:
        await client.async_fetch_devices()
    except MikrotikAuthError as err:
        raise InvalidAuth from err
    except MikrotikApiError as err:
        raise CannotConnect from err


async def _async_discover_backend_url(
    hass: HomeAssistant,
    api_key: str | None = None,
) -> str | None:
    """Auto-discover backend URL with supervisor hint and fallback candidates."""
    supervisor_candidates = await _async_try_supervisor_discovery(hass)
    if supervisor_candidates:
        discovered = await _async_try_candidate_urls(
            hass,
            supervisor_candidates,
            api_key,
        )
        if discovered:
            return discovered

    return await _async_try_candidate_urls(hass, AUTO_DISCOVERY_CANDIDATES, api_key)


async def _async_try_supervisor_discovery(hass: HomeAssistant) -> tuple[str, ...]:
    """Discover likely backend URL candidates via Supervisor API."""
    supervisor_token = os.environ.get("SUPERVISOR_TOKEN")
    if not supervisor_token:
        return ()

    session = async_get_clientsession(hass)
    try:
        async with session.get(
            "http://supervisor/addons",
            headers={"Authorization": f"Bearer {supervisor_token}"},
            timeout=aiohttp.ClientTimeout(total=10),
        ) as response:
            if response.status >= 400:
                _LOGGER.debug(
                    "Supervisor add-ons discovery failed with status %s",
                    response.status,
                )
                return ()
            payload = await response.json(content_type=None)
    except (aiohttp.ClientError, TimeoutError, ValueError) as err:
        _LOGGER.debug("Supervisor add-ons discovery failed: %s", err)
        return ()

    add_ons = _extract_supervisor_addons(payload)
    matched_add_on = _match_target_add_on(add_ons)
    if matched_add_on is None:
        _LOGGER.debug("Supervisor add-ons list does not contain %s", SUPERVISOR_ADDON_SLUG)
        return ()

    port = _extract_supervisor_add_on_port(matched_add_on)
    return (f"http://homeassistant:{port}",)


async def _async_try_candidate_urls(
    hass: HomeAssistant,
    candidates: Iterable[str],
    api_key: str | None,
) -> str | None:
    """Try backend candidates and return the first valid URL."""
    session = async_get_clientsession(hass)
    checked: set[str] = set()
    for raw_candidate in candidates:
        try:
            candidate = _normalize_base_url(raw_candidate)
        except CannotConnect:
            continue
        if candidate in checked:
            continue
        checked.add(candidate)

        client = MikrotikPresenceClient(session, candidate, api_key)
        try:
            await client.async_fetch_devices()
            return candidate
        except MikrotikAuthError:
            _LOGGER.debug("Backend candidate %s responded with auth error", candidate)
            return candidate
        except MikrotikApiError as err:
            _LOGGER.debug("Backend candidate %s is not available: %s", candidate, err)

    return None


def _extract_supervisor_addons(payload: Any) -> list[dict[str, Any]]:
    """Extract add-ons list from Supervisor API response payload."""
    if not isinstance(payload, dict):
        return []

    data = payload.get("data")
    if isinstance(data, dict):
        add_ons = data.get("addons")
        if isinstance(add_ons, list):
            return [item for item in add_ons if isinstance(item, dict)]

    add_ons = payload.get("addons")
    if isinstance(add_ons, list):
        return [item for item in add_ons if isinstance(item, dict)]

    return []


def _match_target_add_on(add_ons: list[dict[str, Any]]) -> dict[str, Any] | None:
    """Find target add-on in Supervisor list by slug or name."""
    slug_suffix = f"_{SUPERVISOR_ADDON_SLUG}"
    for add_on in add_ons:
        slug = (
            _normalize_optional_string(add_on.get("slug"))
            or _normalize_optional_string(add_on.get("addon"))
            or _normalize_optional_string(add_on.get("id"))
            or ""
        ).lower()
        name = (_normalize_optional_string(add_on.get("name")) or "").lower()

        if slug == SUPERVISOR_ADDON_SLUG:
            return add_on
        if slug.endswith(slug_suffix):
            return add_on
        if "mikrotik presence" in name:
            return add_on

    return None


def _extract_supervisor_add_on_port(add_on: dict[str, Any]) -> int:
    """Extract backend port from add-on payload."""
    for key in ("ingress_port", "port"):
        value = add_on.get(key)
        if isinstance(value, int) and value > 0:
            return value
        if isinstance(value, str) and value.isdigit():
            parsed = int(value)
            if parsed > 0:
                return parsed

    return DEFAULT_BACKEND_PORT


def _normalize_base_url(raw_base_url: Any) -> str:
    """Normalize backend URL from user input."""
    base_url = str(raw_base_url or "").strip().rstrip("/")
    if not base_url:
        raise CannotConnect("Base URL is empty")
    return base_url


def _normalize_optional_string(value: Any) -> str | None:
    """Normalize optional string field."""
    if value is None:
        return None
    normalized = str(value).strip()
    return normalized or None
