"""Config flow for MikroTik Presence."""

from __future__ import annotations

import aiohttp
import voluptuous as vol

from homeassistant import config_entries
from homeassistant.const import CONF_HOST, CONF_PASSWORD, CONF_USERNAME, CONF_SSL
from homeassistant.core import callback

from .const import CONF_POLL_INTERVAL_SEC, CONF_VERIFY_TLS, DOMAIN

DEFAULT_POLL_INTERVAL = 5

STEP_USER_DATA_SCHEMA = vol.Schema(
    {
        vol.Required(CONF_HOST): str,
        vol.Required(CONF_USERNAME): str,
        vol.Required(CONF_PASSWORD): str,
        vol.Required(CONF_SSL, default=True): bool,
        vol.Required(CONF_VERIFY_TLS, default=True): bool,
        vol.Required(CONF_POLL_INTERVAL_SEC, default=DEFAULT_POLL_INTERVAL): vol.All(
            int, vol.Range(min=5, max=3600)
        ),
    }
)


async def _validate_connection(data: dict) -> None:
    scheme = "https" if data[CONF_SSL] else "http"
    url = f"{scheme}://{data[CONF_HOST]}/rest/system/resource"
    timeout = aiohttp.ClientTimeout(total=8)

    async with aiohttp.ClientSession(timeout=timeout) as session:
        async with session.get(
            url,
            auth=aiohttp.BasicAuth(data[CONF_USERNAME], data[CONF_PASSWORD]),
            ssl=data[CONF_VERIFY_TLS],
        ) as response:
            if response.status >= 400:
                raise aiohttp.ClientError(f"status={response.status}")


class ConfigFlow(config_entries.ConfigFlow, domain=DOMAIN):
    """Handle a config flow for MikroTik Presence."""

    VERSION = 1

    async def async_step_user(self, user_input: dict | None = None):
        """Handle the initial step."""
        errors: dict[str, str] = {}

        if self._async_current_entries():
            return self.async_abort(reason="single_instance_allowed")

        if user_input is not None:
            try:
                await _validate_connection(user_input)
            except (aiohttp.ClientError, TimeoutError):
                errors["base"] = "cannot_connect"
            else:
                return self.async_create_entry(title=user_input[CONF_HOST], data=user_input)

        return self.async_show_form(
            step_id="user", data_schema=STEP_USER_DATA_SCHEMA, errors=errors
        )

    @staticmethod
    @callback
    def async_get_options_flow(config_entry: config_entries.ConfigEntry):
        return OptionsFlow(config_entry)


class OptionsFlow(config_entries.OptionsFlow):
    """Options flow for MikroTik Presence."""

    def __init__(self, config_entry: config_entries.ConfigEntry) -> None:
        self._config_entry = config_entry

    async def async_step_init(self, user_input: dict | None = None):
        errors: dict[str, str] = {}
        existing = self._config_entry.data

        schema = vol.Schema(
            {
                vol.Required(CONF_HOST, default=existing.get(CONF_HOST)): str,
                vol.Required(CONF_USERNAME, default=existing.get(CONF_USERNAME)): str,
                vol.Required(CONF_PASSWORD, default=existing.get(CONF_PASSWORD)): str,
                vol.Required(CONF_SSL, default=existing.get(CONF_SSL, True)): bool,
                vol.Required(
                    CONF_VERIFY_TLS, default=existing.get(CONF_VERIFY_TLS, True)
                ): bool,
                vol.Required(
                    CONF_POLL_INTERVAL_SEC,
                    default=existing.get(CONF_POLL_INTERVAL_SEC, DEFAULT_POLL_INTERVAL),
                ): vol.All(int, vol.Range(min=5, max=3600)),
            }
        )

        if user_input is not None:
            try:
                await _validate_connection(user_input)
            except (aiohttp.ClientError, TimeoutError):
                errors["base"] = "cannot_connect"
            else:
                self.hass.config_entries.async_update_entry(
                    self._config_entry, data={**self._config_entry.data, **user_input}
                )
                return self.async_create_entry(title="", data={})

        return self.async_show_form(step_id="init", data_schema=schema, errors=errors)
