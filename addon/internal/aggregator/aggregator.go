package aggregator

import (
	"slices"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	"github.com/micro-ha/mikrotik-presence/addon/internal/routeros"
	"github.com/micro-ha/mikrotik-presence/addon/internal/subnet"
)

type OUILookup interface {
	Lookup(mac string) string
}

type Aggregator struct {
	subnetMatcher *subnet.Matcher
	ouiLookup     OUILookup
}

func New(matcher *subnet.Matcher, ouiLookup OUILookup) *Aggregator {
	return &Aggregator{subnetMatcher: matcher, ouiLookup: ouiLookup}
}

func (a *Aggregator) Aggregate(snapshot *routeros.Snapshot) map[string]model.Observation {
	now := snapshot.FetchedAt.UTC()
	result := make(map[string]model.Observation)
	priorities := make(map[string]int)
	matcher := a.subnetMatcher.WithAddresses(snapshot.Addresses)

	for _, lease := range snapshot.DHCP {
		obs := getObservation(result, lease.MAC, now)
		active := strings.EqualFold(lease.Status, "bound") || lease.Status == ""
		if active {
			obs.Online = true
			appendSource(obs, model.SourceDHCP)
		}
		if lease.Address != "" {
			obs.IP = lease.Address
			subnetName := matcher.Match(lease.Address)
			if subnetName != "" {
				obs.LastSubnet = subnetName
			}
		}
		if lease.HostName != "" {
			obs.HostName = lease.HostName
		}
		if ts := parseRouterOSTimestamp(lease.LastSeen, now); ts != nil {
			setLastSeen(obs, priorities, ts, 1)
		}
		putRawSource(obs, model.SourceDHCP, lease)
		result[lease.MAC] = *obs
	}

	for _, reg := range snapshot.WiFi {
		obs := getObservation(result, reg.MAC, now)
		obs.Online = true
		appendSource(obs, model.SourceWiFi)
		if ts := parseRouterOSTimestamp(firstNonEmpty(reg.LastActivity, reg.Uptime), now); ts != nil {
			setLastSeen(obs, priorities, ts, 2)
		}
		putRawSource(obs, model.SourceWiFi, reg)
		result[reg.MAC] = *obs
	}

	for _, host := range snapshot.Bridge {
		obs := getObservation(result, host.MAC, now)
		obs.Online = true
		appendSource(obs, model.SourceBridge)
		putRawSource(obs, model.SourceBridge, host)
		result[host.MAC] = *obs
	}

	for _, arp := range snapshot.ARP {
		obs := getObservation(result, arp.MAC, now)
		obs.Online = true
		appendSource(obs, model.SourceARP)
		if obs.IP == "" && arp.Address != "" {
			obs.IP = arp.Address
			subnetName := matcher.Match(arp.Address)
			if subnetName != "" {
				obs.LastSubnet = subnetName
			}
		}
		putRawSource(obs, model.SourceARP, arp)
		result[arp.MAC] = *obs
	}

	for mac, obs := range result {
		if obs.LastSeenAt == nil && obs.Online {
			ts := now
			obs.LastSeenAt = &ts
		}
		obs.Vendor = a.ouiLookup.Lookup(mac)
		obs.Generated = generatedName(mac, obs.Vendor)
		result[mac] = obs
	}

	return result
}

func getObservation(items map[string]model.Observation, mac string, now time.Time) *model.Observation {
	obs, ok := items[mac]
	if !ok {
		obs = model.Observation{
			MAC:        mac,
			ObservedAt: now,
			RawSources: make(map[string]any),
		}
	}
	if obs.RawSources == nil {
		obs.RawSources = make(map[string]any)
	}
	return &obs
}

func setLastSeen(obs *model.Observation, priorities map[string]int, ts *time.Time, priority int) {
	if ts == nil {
		return
	}
	current, ok := priorities[obs.MAC]
	if !ok || priority < current || obs.LastSeenAt == nil {
		v := ts.UTC()
		obs.LastSeenAt = &v
		priorities[obs.MAC] = priority
	}
}

func appendSource(obs *model.Observation, source string) {
	if !slices.Contains(obs.Sources, source) {
		obs.Sources = append(obs.Sources, source)
	}
}

func putRawSource(obs *model.Observation, source string, payload any) {
	obs.RawSources[source] = payload
}

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
}

func generatedName(mac, vendor string) string {
	suffix := strings.ReplaceAll(mac, ":", "")
	if len(suffix) >= 4 {
		suffix = suffix[len(suffix)-4:]
	}
	if vendor == "Unknown" {
		return "Device-" + suffix
	}
	return vendor + "-" + suffix
}
