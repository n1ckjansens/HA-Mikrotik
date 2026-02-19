package routeros

import "time"

type DHCPLease struct {
	MAC      string
	Address  string
	HostName string
	Server   string
	Status   string
	LastSeen string
	Dynamic  bool
	Blocked  bool
	Disabled bool
}

type WiFiRegistration struct {
	MAC          string
	Interface    string
	SSID         string
	Uptime       string
	LastActivity string
	Signal       string
	AuthType     string
	Band         string
	Driver       string
}

type BridgeHost struct {
	MAC       string
	Bridge    string
	Interface string
	VID       string
}

type ARPEntry struct {
	MAC       string
	Address   string
	Interface string
	Flags     string
}

type IPAddress struct {
	Address   string
	Interface string
}

type Snapshot struct {
	DHCP      []DHCPLease
	WiFi      []WiFiRegistration
	Bridge    []BridgeHost
	ARP       []ARPEntry
	Addresses []IPAddress
	FetchedAt time.Time
}
