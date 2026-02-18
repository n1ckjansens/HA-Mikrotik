package model

import "time"

const (
	SourceDHCP   = "dhcp"
	SourceWiFi   = "wifi"
	SourceBridge = "bridge"
	SourceARP    = "arp"
)

// Observation is a merged snapshot for one MAC at a given poll cycle.
type Observation struct {
	MAC        string
	IP         string
	HostName   string
	Online     bool
	LastSeenAt *time.Time
	Sources    []string
	RawSources map[string]any
	LastSubnet string
	Vendor     string
	Generated  string
	ObservedAt time.Time
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
	LastSources      []string   `json:"last_sources"`
	RawSources       any        `json:"raw_sources,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
	FirstSeenAt      *time.Time `json:"first_seen_at,omitempty"`
}
