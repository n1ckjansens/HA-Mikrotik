package routeros

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

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
	Complete  bool
	Status    string
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

// FetchSnapshot collects RouterOS signals for device presence aggregation.
func (c *Client) FetchSnapshot(ctx context.Context) (*Snapshot, error) {
	snapshot := &Snapshot{FetchedAt: time.Now().UTC()}

	dhcpRows, err := c.RunCommand(ctx, "/ip/dhcp-server/lease/print", map[string]string{
		".proplist": "mac-address,address,host-name,server,status,last-seen,dynamic,blocked,disabled",
	})
	if err != nil {
		return nil, fmt.Errorf("fetch dhcp leases: %w", err)
	}
	snapshot.DHCP = mapDHCPRows(dhcpRows)

	wifiRows, err := c.fetchWiFiRows(ctx)
	if err != nil {
		return nil, err
	}
	snapshot.WiFi = mapWiFiRows(wifiRows)

	bridgeRows, err := c.RunCommand(ctx, "/interface/bridge/host/print", map[string]string{
		".proplist": "mac-address,bridge,interface,on-interface,vid",
	})
	if err != nil {
		return nil, fmt.Errorf("fetch bridge hosts: %w", err)
	}
	snapshot.Bridge = mapBridgeRows(bridgeRows)

	arpRows, err := c.RunCommand(ctx, "/ip/arp/print", map[string]string{
		".proplist": "mac-address,address,interface,complete,status,flags",
	})
	if err != nil {
		return nil, fmt.Errorf("fetch arp: %w", err)
	}
	snapshot.ARP = mapARPRows(arpRows)

	ipRows, err := c.RunCommand(ctx, "/ip/address/print", map[string]string{
		".proplist": "address,interface",
	})
	if err != nil {
		return nil, fmt.Errorf("fetch ip addresses: %w", err)
	}
	snapshot.Addresses = mapAddressRows(ipRows)

	return snapshot, nil
}

// FetchSnapshot keeps compatibility with current device service contract.
func (m *Manager) FetchSnapshot(ctx context.Context, cfg model.RouterConfig) (*Snapshot, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client.FetchSnapshot(ctx)
}

func (c *Client) fetchWiFiRows(ctx context.Context) ([]wifiRow, error) {
	targets := []struct {
		Path   string
		Driver string
	}{
		{Path: "/interface/wifi/registration-table/print", Driver: "wifi"},
		{Path: "/interface/wifiwave2/registration-table/print", Driver: "wifiwave2"},
		{Path: "/interface/wireless/registration-table/print", Driver: "wireless"},
	}

	rows := make([]wifiRow, 0)
	for _, target := range targets {
		current, err := c.RunCommand(ctx, target.Path, map[string]string{
			".proplist": "mac-address,interface,ssid,uptime,last-activity,signal,tx-signal,auth-type,authentication-types,band",
		})
		if err != nil {
			if isMissingCommandError(err) {
				continue
			}
			return nil, fmt.Errorf("fetch wifi registrations (%s): %w", target.Driver, err)
		}

		for _, row := range current {
			rows = append(rows, wifiRow{Driver: target.Driver, Row: row})
		}
	}
	return rows, nil
}

type wifiRow struct {
	Driver string
	Row    map[string]string
}

func mapDHCPRows(rows []map[string]string) []DHCPLease {
	items := make([]DHCPLease, 0, len(rows))
	for _, row := range rows {
		mac := canonicalMAC(row["mac-address"])
		if mac == "" {
			continue
		}
		items = append(items, DHCPLease{
			MAC:      mac,
			Address:  strings.TrimSpace(row["address"]),
			HostName: strings.TrimSpace(row["host-name"]),
			Server:   strings.TrimSpace(row["server"]),
			Status:   strings.TrimSpace(row["status"]),
			LastSeen: strings.TrimSpace(row["last-seen"]),
			Dynamic:  boolFromWord(row["dynamic"]),
			Blocked:  boolFromWord(row["blocked"]),
			Disabled: boolFromWord(row["disabled"]),
		})
	}
	return items
}

func mapWiFiRows(rows []wifiRow) []WiFiRegistration {
	items := make([]WiFiRegistration, 0, len(rows))
	for _, item := range rows {
		mac := canonicalMAC(item.Row["mac-address"])
		if mac == "" {
			continue
		}
		items = append(items, WiFiRegistration{
			MAC:          mac,
			Interface:    strings.TrimSpace(item.Row["interface"]),
			SSID:         strings.TrimSpace(item.Row["ssid"]),
			Uptime:       strings.TrimSpace(item.Row["uptime"]),
			LastActivity: strings.TrimSpace(item.Row["last-activity"]),
			Signal:       firstNonEmpty(strings.TrimSpace(item.Row["signal"]), strings.TrimSpace(item.Row["tx-signal"])),
			AuthType:     firstNonEmpty(strings.TrimSpace(item.Row["auth-type"]), strings.TrimSpace(item.Row["authentication-types"])),
			Band:         strings.TrimSpace(item.Row["band"]),
			Driver:       item.Driver,
		})
	}
	return items
}

func mapBridgeRows(rows []map[string]string) []BridgeHost {
	items := make([]BridgeHost, 0, len(rows))
	for _, row := range rows {
		mac := canonicalMAC(row["mac-address"])
		if mac == "" {
			continue
		}
		items = append(items, BridgeHost{
			MAC:       mac,
			Bridge:    strings.TrimSpace(row["bridge"]),
			Interface: firstNonEmpty(strings.TrimSpace(row["interface"]), strings.TrimSpace(row["on-interface"])),
			VID:       strings.TrimSpace(row["vid"]),
		})
	}
	return items
}

func mapARPRows(rows []map[string]string) []ARPEntry {
	items := make([]ARPEntry, 0, len(rows))
	for _, row := range rows {
		mac := canonicalMAC(row["mac-address"])
		if mac == "" {
			continue
		}
		items = append(items, ARPEntry{
			MAC:       mac,
			Address:   strings.TrimSpace(row["address"]),
			Interface: strings.TrimSpace(row["interface"]),
			Complete: boolFromWord(row["complete"]) ||
				strings.EqualFold(strings.TrimSpace(row["status"]), "complete") ||
				strings.Contains(strings.ToUpper(strings.TrimSpace(row["flags"])), "C"),
			Status: strings.TrimSpace(row["status"]),
			Flags:  strings.TrimSpace(row["flags"]),
		})
	}
	return items
}

func mapAddressRows(rows []map[string]string) []IPAddress {
	items := make([]IPAddress, 0, len(rows))
	for _, row := range rows {
		address := strings.TrimSpace(row["address"])
		if address == "" {
			continue
		}
		items = append(items, IPAddress{
			Address:   address,
			Interface: strings.TrimSpace(row["interface"]),
		})
	}
	return items
}
