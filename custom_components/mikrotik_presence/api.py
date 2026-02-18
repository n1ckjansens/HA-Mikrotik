"""Internal API for MikroTik Presence Add-on."""

from __future__ import annotations

from homeassistant.components.http import HomeAssistantView
from homeassistant.core import HomeAssistant
from homeassistant.helpers.typing import ConfigType
from homeassistant.util.dt import utcnow

from .const import (
    API_CONFIG_URL,
    CONF_POLL_INTERVAL_SEC,
    CONF_ROLES,
    CONF_SSL,
    CONF_VERIFY_TLS,
    CONF_VERSION,
    DOMAIN,
)


class MikroTikPresenceConfigView(HomeAssistantView):
    """Return integration config for add-on."""

    url = API_CONFIG_URL
    name = "api:mikrotik_presence:config"
    requires_auth = True

    def __init__(self, hass: HomeAssistant) -> None:
        self.hass = hass

    async def get(self, request):
        user = request.get("hass_user")
        if user is None or not getattr(user, "is_system_generated", False):
            return self.json_message("Forbidden", status_code=403)
        if getattr(user, "name", "") != "Supervisor":
            return self.json_message("Forbidden", status_code=403)

        entries = self.hass.config_entries.async_entries(DOMAIN)
        if not entries:
            return self.json({"configured": False}, status_code=404)

        entry = entries[0]
        data = entry.data
        version = self.hass.data[DOMAIN].get(CONF_VERSION, 1)

        payload = {
            "configured": True,
            "version": version,
            "updated_at": utcnow().isoformat(),
            "host": data.get("host"),
            "username": data.get("username"),
            "password": data.get("password"),
            "ssl": data.get(CONF_SSL, True),
            "verify_tls": data.get(CONF_VERIFY_TLS, True),
            "poll_interval_sec": data.get(CONF_POLL_INTERVAL_SEC, 5),
            "roles": data.get(CONF_ROLES, []),
        }
        return self.json(payload)


async def async_register_view(hass: HomeAssistant, _config: ConfigType) -> None:
    hass.http.register_view(MikroTikPresenceConfigView(hass))
