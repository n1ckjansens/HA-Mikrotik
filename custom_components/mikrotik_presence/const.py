"""Constants for MikroTik Presence integration."""

DOMAIN = "mikrotik_presence"
PLATFORMS: list[str] = ["switch", "select"]

CONF_BASE_URL = "base_url"
CONF_API_KEY = "api_key"

DATA_CLIENT = "client"
DATA_COORDINATOR_DEVICES = "devices_coordinator"
DATA_COORDINATOR_GLOBAL = "global_coordinator"

DEFAULT_BACKEND_PORT = 8099
SUPERVISOR_ADDON_SLUG = "mikrotik_presence"
AUTO_DISCOVERY_CANDIDATES: tuple[str, ...] = (
    f"http://homeassistant:{DEFAULT_BACKEND_PORT}",
    f"http://127.0.0.1:{DEFAULT_BACKEND_PORT}",
    f"http://host.docker.internal:{DEFAULT_BACKEND_PORT}",
)

UPDATE_INTERVAL_SECONDS = 10
