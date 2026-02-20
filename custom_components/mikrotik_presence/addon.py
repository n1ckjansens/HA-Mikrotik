"""Supervisor add-on helpers for MikroTik Presence integration."""

from __future__ import annotations

import logging

from homeassistant.components.hassio import AddonManager
from homeassistant.core import HomeAssistant

from .const import ADDON_NAME, ADDON_SLUG

LOGGER = logging.getLogger(__name__)


def get_addon_manager(hass: HomeAssistant) -> AddonManager:
    """Return AddonManager for MikroTik Presence add-on."""
    return AddonManager(hass, LOGGER, ADDON_NAME, ADDON_SLUG)
