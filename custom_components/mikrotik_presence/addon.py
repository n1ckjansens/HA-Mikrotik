"""Supervisor add-on helpers for MikroTik Presence integration."""

from __future__ import annotations

import logging
from typing import Any

from aiohasupervisor import SupervisorError

from homeassistant.components.hassio import (
    AddonError,
    AddonManager,
    AddonState,
    hostname_from_addon_slug,
)
from homeassistant.components.hassio.handler import HassioAPIError, get_supervisor_client
from homeassistant.core import HomeAssistant

from .const import ADDON_BASE_SLUG, ADDON_DEFAULT_PORT, ADDON_NAME

LOGGER = logging.getLogger(__name__)


class MikrotikAddonManager(AddonManager):
    """Addon manager that supports custom repository slug resolution."""

    async def async_resolve_slug(self) -> str:
        """Resolve full add-on slug from Supervisor Store.

        Custom repositories use hashed prefixes in store slugs
        (for example: ``<repo_hash>_mikrotik_presence``).
        """
        if self.addon_slug != ADDON_BASE_SLUG:
            return self.addon_slug

        supervisor = get_supervisor_client(self._hass)

        try:
            candidates = self._matching_candidates(await supervisor.store.addons_list())
            if not candidates:
                await supervisor.store.reload()
                candidates = self._matching_candidates(await supervisor.store.addons_list())
        except (HassioAPIError, SupervisorError) as err:
            raise AddonError(f"Failed to read Supervisor Store: {err}") from err

        if not candidates:
            raise AddonError(f"{ADDON_NAME} add-on is not available in Store")

        candidates.sort(key=self._candidate_sort_key)
        resolved_slug = candidates[0].slug

        if resolved_slug != self.addon_slug:
            LOGGER.debug("Resolved add-on slug %s -> %s", self.addon_slug, resolved_slug)
            self.addon_slug = resolved_slug

        return self.addon_slug

    async def async_ensure_installed_and_running(self) -> None:
        """Ensure add-on is installed and running."""
        await self.async_resolve_slug()

        addon_info = await self.async_get_addon_info()

        if addon_info.state is AddonState.NOT_INSTALLED:
            await self.async_schedule_install_addon()
            addon_info = await self.async_get_addon_info()

        if addon_info.state is not AddonState.RUNNING:
            await self.async_schedule_start_addon()

    async def async_get_backend_base_url(self) -> str:
        """Return backend URL from discovery, with hostname fallback."""
        await self.async_resolve_slug()

        try:
            discovery = await self.async_get_addon_discovery_info()
        except AddonError:
            discovery = None

        if isinstance(discovery, dict):
            host = _normalize_host(discovery.get("host"))
            port = _normalize_port(discovery.get("port"))
            if host and port:
                return f"http://{host}:{port}"

        host = hostname_from_addon_slug(self.addon_slug)
        return f"http://{host}:{ADDON_DEFAULT_PORT}"

    def _matching_candidates(self, addons: list[Any]) -> list[Any]:
        """Return store add-ons matching this integration add-on slug."""
        return [addon for addon in addons if _slug_matches(getattr(addon, "slug", ""))]

    def _candidate_sort_key(self, addon: Any) -> tuple[int, int, int, str]:
        """Sort key for candidate slugs.

        Priority:
        1. Exact slug match.
        2. Installed add-ons.
        3. Shorter slug.
        4. Lexicographical fallback.
        """
        slug = str(getattr(addon, "slug", ""))
        return (
            0 if slug == ADDON_BASE_SLUG else 1,
            0 if bool(getattr(addon, "installed", False)) else 1,
            len(slug),
            slug,
        )


def get_addon_manager(hass: HomeAssistant) -> MikrotikAddonManager:
    """Return add-on manager for MikroTik Presence."""
    return MikrotikAddonManager(hass, LOGGER, ADDON_NAME, ADDON_BASE_SLUG)


def _slug_matches(slug: str) -> bool:
    """Return true when slug is the base slug or repository-prefixed slug."""
    normalized = str(slug).strip()
    return normalized == ADDON_BASE_SLUG or normalized.endswith(f"_{ADDON_BASE_SLUG}")


def _normalize_host(value: Any) -> str | None:
    """Normalize discovery host value."""
    if value is None:
        return None

    host = str(value).strip()
    return host or None


def _normalize_port(value: Any) -> int | None:
    """Normalize discovery port value."""
    if isinstance(value, int):
        port = value
    elif isinstance(value, str) and value.strip().isdigit():
        port = int(value.strip())
    else:
        return None

    if port <= 0:
        return None

    return port
