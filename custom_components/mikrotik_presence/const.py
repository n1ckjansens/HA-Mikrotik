"""Constants for MikroTik Presence integration."""

DOMAIN = "mikrotik_presence"
PLATFORMS: list[str] = ["switch", "select"]

CONF_BASE_URL = "base_url"
CONF_API_KEY = "api_key"

DATA_CLIENT = "client"
DATA_COORDINATOR_DEVICES = "devices_coordinator"
DATA_COORDINATOR_GLOBAL = "global_coordinator"

UPDATE_INTERVAL_SECONDS = 10

ADDON_SLUG = "mikrotik_presence"
ADDON_NAME = "MikroTik Presence"
