package routeros

import "time"

type DHCPLease struct {
	MAC      string
	Address  string
	HostName string
	Status   string
	LastSeen string
}

type WiFiRegistration struct {
	MAC          string
	Interface    string
	Uptime       string
	LastActivity string
}

type BridgeHost struct {
	MAC       string
	Interface string
}

type ARPEntry struct {
	MAC       string
	Address   string
	Interface string
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
