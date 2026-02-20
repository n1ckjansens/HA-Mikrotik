package routeros

import (
	"context"
	"fmt"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// InterfaceInfo describes one RouterOS network interface.
type InterfaceInfo struct {
	ID         string
	Name       string
	Type       string
	MACAddress string
	Comment    string
	Running    bool
	Disabled   bool
}

// InterfaceTrafficStats is one-shot monitor-traffic sample.
type InterfaceTrafficStats struct {
	Name               string
	RxBitsPerSecond    float64
	TxBitsPerSecond    float64
	RxPacketsPerSecond float64
	TxPacketsPerSecond float64
}

// ListInterfaces returns /interface/print mapped output.
func (c *Client) ListInterfaces(ctx context.Context) ([]InterfaceInfo, error) {
	rows, err := c.RunCommand(ctx, "/interface/print", map[string]string{
		".proplist": ".id,name,type,mac-address,comment,running,disabled",
	})
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}

	items := make([]InterfaceInfo, 0, len(rows))
	for _, row := range rows {
		id := strings.TrimSpace(row[".id"])
		if id == "" {
			continue
		}
		items = append(items, InterfaceInfo{
			ID:         id,
			Name:       strings.TrimSpace(row["name"]),
			Type:       strings.TrimSpace(row["type"]),
			MACAddress: canonicalMAC(row["mac-address"]),
			Comment:    strings.TrimSpace(row["comment"]),
			Running:    boolFromWord(row["running"]),
			Disabled:   boolFromWord(row["disabled"]),
		})
	}
	return items, nil
}

// ListInterfaces returns interfaces for selected pooled client.
func (m *Manager) ListInterfaces(ctx context.Context, cfg model.RouterConfig) ([]InterfaceInfo, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client.ListInterfaces(ctx)
}

// InterfaceTraffic returns one sample from /interface/monitor-traffic.
func (c *Client) InterfaceTraffic(ctx context.Context, name string) (InterfaceTrafficStats, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return InterfaceTrafficStats{}, &ValidationError{Field: "name", Reason: "is required"}
	}

	rows, err := c.RunCommand(ctx, "/interface/monitor-traffic", map[string]string{
		"interface": name,
		"once":      "",
		".proplist": "name,rx-bits-per-second,tx-bits-per-second,rx-packets-per-second,tx-packets-per-second",
	})
	if err != nil {
		return InterfaceTrafficStats{}, fmt.Errorf("monitor interface %q: %w", name, err)
	}
	if len(rows) == 0 {
		return InterfaceTrafficStats{}, fmt.Errorf("monitor interface %q: empty response", name)
	}

	row := rows[0]
	rxBPS, err := parseFloat64(row["rx-bits-per-second"])
	if err != nil {
		return InterfaceTrafficStats{}, err
	}
	txBPS, err := parseFloat64(row["tx-bits-per-second"])
	if err != nil {
		return InterfaceTrafficStats{}, err
	}
	rxPPS, err := parseFloat64(row["rx-packets-per-second"])
	if err != nil {
		return InterfaceTrafficStats{}, err
	}
	txPPS, err := parseFloat64(row["tx-packets-per-second"])
	if err != nil {
		return InterfaceTrafficStats{}, err
	}

	return InterfaceTrafficStats{
		Name:               firstNonEmpty(strings.TrimSpace(row["name"]), name),
		RxBitsPerSecond:    rxBPS,
		TxBitsPerSecond:    txBPS,
		RxPacketsPerSecond: rxPPS,
		TxPacketsPerSecond: txPPS,
	}, nil
}

// InterfaceTraffic returns one monitor sample for selected pooled client.
func (m *Manager) InterfaceTraffic(
	ctx context.Context,
	cfg model.RouterConfig,
	name string,
) (InterfaceTrafficStats, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return InterfaceTrafficStats{}, err
	}
	return client.InterfaceTraffic(ctx, name)
}
