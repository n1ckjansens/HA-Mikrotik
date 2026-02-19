package model

import "time"

const (
	SourceDHCP   = "dhcp"
	SourceWiFi   = "wifi"
	SourceBridge = "bridge"
	SourceARP    = "arp"
)

type ConnectionStatus string

const (
	ConnectionStatusOnline     ConnectionStatus = "ONLINE"
	ConnectionStatusIdleRecent ConnectionStatus = "IDLE_RECENT"
	ConnectionStatusOffline    ConnectionStatus = "OFFLINE"
	ConnectionStatusUnknown    ConnectionStatus = "UNKNOWN"
)

type PresenceThresholds struct {
	WiFiIdleThreshold    time.Duration
	DHCPRecentThreshold  time.Duration
	OfflineHardThreshold time.Duration
}

func DefaultPresenceThresholds() PresenceThresholds {
	return PresenceThresholds{
		WiFiIdleThreshold:    5 * time.Minute,
		DHCPRecentThreshold:  30 * time.Minute,
		OfflineHardThreshold: 24 * time.Hour,
	}
}

func (p PresenceThresholds) Normalize() PresenceThresholds {
	defaults := DefaultPresenceThresholds()
	if p.WiFiIdleThreshold <= 0 {
		p.WiFiIdleThreshold = defaults.WiFiIdleThreshold
	}
	if p.DHCPRecentThreshold <= 0 {
		p.DHCPRecentThreshold = defaults.DHCPRecentThreshold
	}
	if p.OfflineHardThreshold <= 0 {
		p.OfflineHardThreshold = defaults.OfflineHardThreshold
	}
	return p
}

// Observation is a merged snapshot for one MAC at a given poll cycle.
type Observation struct {
	MAC        string
	IP         string
	HostName   string
	Interface  string
	Bridge     string
	SSID       string
	Online     bool
	LastSeenAt *time.Time
	Sources    []string
	RawSources map[string]any
	LastSubnet string
	Vendor     string
	Generated  string
	ObservedAt time.Time

	DHCPServer   string
	DHCPStatus   string
	DHCPLastSeen *time.Duration

	WiFiDriver       string
	WiFiInterface    string
	WiFiLastActivity *time.Duration
	WiFiUptime       *time.Duration
	WiFiAuthType     string
	WiFiSignal       *int

	ARPIP         string
	ARPInterface  string
	ARPIsComplete bool

	BridgeHostPort string
	BridgeHostVLAN *int

	ConnectionStatus ConnectionStatus
	StatusReason     string
}

type DeviceRegistered struct {
	MAC       string    `json:"mac"`
	Name      *string   `json:"name,omitempty"`
	Icon      *string   `json:"icon,omitempty"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeviceState struct {
	MAC              string     `json:"mac"`
	Online           bool       `json:"online"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	ConnectedSinceAt *time.Time `json:"connected_since_at,omitempty"`
	LastIP           *string    `json:"last_ip,omitempty"`
	LastSubnet       *string    `json:"last_subnet,omitempty"`
	HostName         *string    `json:"host_name,omitempty"`
	Interface        *string    `json:"interface,omitempty"`
	Bridge           *string    `json:"bridge,omitempty"`
	SSID             *string    `json:"ssid,omitempty"`
	DHCPServer       *string    `json:"dhcp_server,omitempty"`
	DHCPStatus       *string    `json:"dhcp_status,omitempty"`
	DHCPLastSeenSec  *int64     `json:"dhcp_last_seen_sec,omitempty"`
	WiFiDriver       *string    `json:"wifi_driver,omitempty"`
	WiFiInterface    *string    `json:"wifi_interface,omitempty"`
	WiFiLastActSec   *int64     `json:"wifi_last_activity_sec,omitempty"`
	WiFiUptimeSec    *int64     `json:"wifi_uptime_sec,omitempty"`
	WiFiAuthType     *string    `json:"wifi_auth_type,omitempty"`
	WiFiSignal       *int       `json:"wifi_signal,omitempty"`
	ARPIP            *string    `json:"arp_ip,omitempty"`
	ARPInterface     *string    `json:"arp_interface,omitempty"`
	ARPIsComplete    bool       `json:"arp_is_complete"`
	BridgeHostPort   *string    `json:"bridge_host_port,omitempty"`
	BridgeHostVLAN   *int       `json:"bridge_host_vlan,omitempty"`
	ConnectionStatus string     `json:"connection_status"`
	StatusReason     string     `json:"status_reason"`
	LastSourcesJSON  string     `json:"last_sources_json"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type DeviceNewCache struct {
	MAC           string    `json:"mac"`
	FirstSeenAt   time.Time `json:"first_seen_at"`
	Vendor        string    `json:"vendor"`
	GeneratedName string    `json:"generated_name"`
}

type DeviceView struct {
	MAC              string     `json:"mac"`
	Name             string     `json:"name"`
	Vendor           string     `json:"vendor"`
	Icon             *string    `json:"icon,omitempty"`
	Comment          *string    `json:"comment,omitempty"`
	Status           string     `json:"status"`
	Online           bool       `json:"online"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	ConnectedSinceAt *time.Time `json:"connected_since_at,omitempty"`
	LastIP           *string    `json:"last_ip,omitempty"`
	LastSubnet       *string    `json:"last_subnet,omitempty"`
	HostName         *string    `json:"host_name,omitempty"`
	Interface        *string    `json:"interface,omitempty"`
	Bridge           *string    `json:"bridge,omitempty"`
	SSID             *string    `json:"ssid,omitempty"`
	DHCPServer       *string    `json:"dhcp_server,omitempty"`
	DHCPStatus       *string    `json:"dhcp_status,omitempty"`
	DHCPLastSeenSec  *int64     `json:"dhcp_last_seen_sec,omitempty"`
	WiFiDriver       *string    `json:"wifi_driver,omitempty"`
	WiFiInterface    *string    `json:"wifi_interface,omitempty"`
	WiFiLastActSec   *int64     `json:"wifi_last_activity_sec,omitempty"`
	WiFiUptimeSec    *int64     `json:"wifi_uptime_sec,omitempty"`
	WiFiAuthType     *string    `json:"wifi_auth_type,omitempty"`
	WiFiSignal       *int       `json:"wifi_signal,omitempty"`
	ARPIP            *string    `json:"arp_ip,omitempty"`
	ARPInterface     *string    `json:"arp_interface,omitempty"`
	ARPIsComplete    bool       `json:"arp_is_complete"`
	BridgeHostPort   *string    `json:"bridge_host_port,omitempty"`
	BridgeHostVLAN   *int       `json:"bridge_host_vlan,omitempty"`
	ConnectionStatus string     `json:"connection_status"`
	StatusReason     string     `json:"status_reason"`
	LastSources      []string   `json:"last_sources"`
	RawSources       any        `json:"raw_sources,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
	FirstSeenAt      *time.Time `json:"first_seen_at,omitempty"`
}
