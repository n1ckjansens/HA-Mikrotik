"""HTTP client for MikroTik Presence backend."""

from __future__ import annotations

import asyncio
import json
from dataclasses import dataclass, field
from typing import Any

import aiohttp


class MikrotikApiError(Exception):
    """Base backend API error."""


class MikrotikAuthError(MikrotikApiError):
    """Authentication/authorization backend API error."""


@dataclass
class CapabilityControlOption:
    """Capability option for select/switch control."""

    value: str
    label: str


@dataclass
class CapabilityControl:
    """Capability UI control model."""

    type: str
    options: list[CapabilityControlOption]


@dataclass
class CapabilityDTO:
    """Capability payload returned by backend."""

    id: str
    label: str
    description: str | None
    scope: str
    control: CapabilityControl
    enabled: bool
    state: str


@dataclass
class DeviceDTO:
    """Device payload normalized for HA integration."""

    id: str
    name: str
    mac: str | None
    ip: str | None
    vendor: str | None
    subnet: str | None
    online: bool
    registered: bool
    capabilities: list[CapabilityDTO] = field(default_factory=list)


class MikrotikPresenceClient:
    """Async HTTP client for backend API."""

    def __init__(
        self,
        session: aiohttp.ClientSession,
        base_url: str,
        api_key: str | None = None,
    ) -> None:
        self._session = session
        self._base_url = base_url.rstrip("/")
        normalized_key = (api_key or "").strip()
        self._api_key = normalized_key or None

    async def async_fetch_devices(self) -> list[DeviceDTO]:
        """Fetch all devices from backend."""
        payload = await self._async_request("GET", "/api/devices")

        if isinstance(payload, dict):
            raw_items = payload.get("items", [])
        elif isinstance(payload, list):
            raw_items = payload
        else:
            raise MikrotikApiError("Invalid devices response format")

        if not isinstance(raw_items, list):
            raise MikrotikApiError("Invalid devices response payload")

        return [self._parse_device(item) for item in raw_items]

    async def async_fetch_device_capabilities(self, device_id: str) -> list[CapabilityDTO]:
        """Fetch capabilities for one device."""
        payload = await self._async_request("GET", f"/api/devices/{device_id}/capabilities")
        return self._parse_capabilities(payload, default_scope="device")

    async def async_fetch_global_capabilities(self) -> list[CapabilityDTO]:
        """Fetch global capabilities."""
        payload = await self._async_request("GET", "/api/global/capabilities")
        return self._parse_capabilities(payload, default_scope="global")

    async def async_set_device_capability_state(
        self,
        device_id: str,
        capability_id: str,
        state: str,
    ) -> None:
        """Set state for one device capability."""
        await self._async_request(
            "PATCH",
            f"/api/devices/{device_id}/capabilities/{capability_id}",
            payload={"state": state},
        )

    async def async_set_global_capability_state(self, capability_id: str, state: str) -> None:
        """Set state for one global capability."""
        await self._async_request(
            "PATCH",
            f"/api/global/capabilities/{capability_id}",
            payload={"state": state},
        )

    async def _async_request(
        self,
        method: str,
        path: str,
        payload: dict[str, Any] | None = None,
    ) -> Any:
        """Perform HTTP request and decode JSON response."""
        url = f"{self._base_url}{path}"

        try:
            async with self._session.request(
                method,
                url,
                headers=self._build_headers(),
                json=payload,
                timeout=aiohttp.ClientTimeout(total=15),
            ) as response:
                body_text = await response.text()
                response_payload = None
                if body_text.strip():
                    try:
                        response_payload = json.loads(body_text)
                    except json.JSONDecodeError:
                        response_payload = None

                if response.status in (401, 403):
                    message = self._extract_error_message(response_payload, body_text, "Invalid authentication")
                    raise MikrotikAuthError(message)

                if response.status >= 400:
                    message = self._extract_error_message(
                        response_payload,
                        body_text,
                        f"HTTP {response.status}",
                    )
                    raise MikrotikApiError(message)

                if not body_text.strip():
                    return None
                if response_payload is None:
                    raise MikrotikApiError("Invalid JSON in backend response")
                return response_payload
        except MikrotikApiError:
            raise
        except asyncio.TimeoutError as err:
            raise MikrotikApiError("Timeout while connecting to backend") from err
        except aiohttp.ClientError as err:
            raise MikrotikApiError("Cannot connect to backend") from err

    def _build_headers(self) -> dict[str, str]:
        """Build HTTP headers for request."""
        headers: dict[str, str] = {"Accept": "application/json"}
        if self._api_key:
            headers["Authorization"] = f"Bearer {self._api_key}"
        return headers

    def _parse_device(self, raw_device: Any) -> DeviceDTO:
        """Normalize one raw device payload."""
        if not isinstance(raw_device, dict):
            raise MikrotikApiError("Invalid device item in response")

        mac = self._as_optional_string(raw_device.get("mac"))
        device_id = mac or self._as_optional_string(raw_device.get("id"))
        if device_id is None:
            raise MikrotikApiError("Device ID is missing in backend response")

        status = (self._as_optional_string(raw_device.get("status")) or "").lower()
        registered_raw = raw_device.get("registered")
        if isinstance(registered_raw, bool):
            registered = registered_raw
        else:
            registered = status == "registered"

        return DeviceDTO(
            id=device_id,
            name=self._as_optional_string(raw_device.get("name")) or device_id,
            mac=mac,
            ip=(
                self._as_optional_string(raw_device.get("ip"))
                or self._as_optional_string(raw_device.get("last_ip"))
            ),
            vendor=self._as_optional_string(raw_device.get("vendor")),
            subnet=(
                self._as_optional_string(raw_device.get("subnet"))
                or self._as_optional_string(raw_device.get("last_subnet"))
            ),
            online=self._as_bool(raw_device.get("online"), default=False),
            registered=registered,
        )

    def _parse_capabilities(self, payload: Any, default_scope: str) -> list[CapabilityDTO]:
        """Normalize list of capabilities from backend payload."""
        if isinstance(payload, dict):
            raw_items = payload.get("items", [])
        else:
            raw_items = payload

        if not isinstance(raw_items, list):
            raise MikrotikApiError("Invalid capabilities response payload")

        capabilities: list[CapabilityDTO] = []
        for raw_capability in raw_items:
            if not isinstance(raw_capability, dict):
                continue

            capability_id = self._as_optional_string(raw_capability.get("id")) or self._as_optional_string(
                raw_capability.get("capability_id")
            )
            if capability_id is None:
                continue

            capabilities.append(
                CapabilityDTO(
                    id=capability_id,
                    label=self._as_optional_string(raw_capability.get("label")) or capability_id,
                    description=self._as_optional_string(raw_capability.get("description")),
                    scope=self._as_optional_string(raw_capability.get("scope")) or default_scope,
                    control=self._parse_control(raw_capability.get("control")),
                    enabled=self._as_bool(raw_capability.get("enabled"), default=True),
                    state=self._as_optional_string(raw_capability.get("state")) or "",
                )
            )

        return capabilities

    def _parse_control(self, raw_control: Any) -> CapabilityControl:
        """Normalize capability control section."""
        if not isinstance(raw_control, dict):
            return CapabilityControl(type="switch", options=[])

        control_type = self._as_optional_string(raw_control.get("type")) or "switch"
        raw_options = raw_control.get("options")

        options: list[CapabilityControlOption] = []
        if isinstance(raw_options, list):
            for raw_option in raw_options:
                if isinstance(raw_option, dict):
                    value = self._as_optional_string(raw_option.get("value"))
                    if value is None:
                        continue
                    label = self._as_optional_string(raw_option.get("label")) or value
                    options.append(CapabilityControlOption(value=value, label=label))
                elif isinstance(raw_option, str):
                    options.append(CapabilityControlOption(value=raw_option, label=raw_option))

        return CapabilityControl(type=control_type, options=options)

    def _extract_error_message(
        self,
        payload: Any,
        body_text: str,
        fallback: str,
    ) -> str:
        """Extract API error message from JSON/text body."""
        if isinstance(payload, dict):
            error_data = payload.get("error")
            if isinstance(error_data, dict):
                message = self._as_optional_string(error_data.get("message"))
                if message:
                    return message
            message = self._as_optional_string(payload.get("message"))
            if message:
                return message

        body = body_text.strip()
        if body:
            return body

        return fallback

    def _as_optional_string(self, value: Any) -> str | None:
        """Convert value to normalized string or None."""
        if value is None:
            return None
        if isinstance(value, str):
            normalized = value.strip()
            return normalized or None
        return str(value)

    def _as_bool(self, value: Any, default: bool) -> bool:
        """Convert value to bool with fallback default."""
        if isinstance(value, bool):
            return value
        if isinstance(value, str):
            normalized = value.strip().lower()
            if normalized in {"true", "1", "yes", "on"}:
                return True
            if normalized in {"false", "0", "no", "off"}:
                return False
        if isinstance(value, (int, float)):
            return bool(value)
        return default
