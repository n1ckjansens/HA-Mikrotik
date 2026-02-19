PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS devices_registered (
	mac TEXT PRIMARY KEY,
	name TEXT,
	icon TEXT,
	comment TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS devices_state (
	mac TEXT PRIMARY KEY,
	online INTEGER NOT NULL,
	last_seen_at TEXT,
	connected_since_at TEXT,
	last_ip TEXT,
	last_subnet TEXT,
	host_name TEXT,
	interface TEXT,
	bridge TEXT,
	ssid TEXT,
	dhcp_server TEXT,
	dhcp_status TEXT,
	dhcp_last_seen_sec INTEGER,
	wifi_driver TEXT,
	wifi_interface TEXT,
	wifi_last_activity_sec INTEGER,
	wifi_uptime_sec INTEGER,
	wifi_auth_type TEXT,
	wifi_signal INTEGER,
	arp_ip TEXT,
	arp_interface TEXT,
	arp_is_complete INTEGER NOT NULL DEFAULT 0,
	bridge_host_port TEXT,
	bridge_host_vlan INTEGER,
	connection_status TEXT NOT NULL DEFAULT 'UNKNOWN',
	status_reason TEXT NOT NULL DEFAULT '',
	last_sources_json TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS devices_new_cache (
	mac TEXT PRIMARY KEY,
	first_seen_at TEXT NOT NULL,
	vendor TEXT NOT NULL,
	generated_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS capability_templates (
	id TEXT PRIMARY KEY,
	data TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS device_capabilities_state (
	device_id TEXT NOT NULL,
	capability_id TEXT NOT NULL,
	enabled INTEGER NOT NULL,
	state TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY (device_id, capability_id)
);

CREATE INDEX IF NOT EXISTS idx_devices_state_online ON devices_state(online);
CREATE INDEX IF NOT EXISTS idx_capabilities_updated_at ON capability_templates(updated_at);
CREATE INDEX IF NOT EXISTS idx_device_cap_state_device ON device_capabilities_state(device_id);
CREATE INDEX IF NOT EXISTS idx_device_cap_state_capability ON device_capabilities_state(capability_id);
