package aggregator

import (
	"slices"
	"strconv"
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
	thresholds    model.PresenceThresholds
}

func New(matcher *subnet.Matcher, ouiLookup OUILookup) *Aggregator {
	return NewWithThresholds(matcher, ouiLookup, model.DefaultPresenceThresholds())
}

func NewWithThresholds(
	matcher *subnet.Matcher,
	ouiLookup OUILookup,
	thresholds model.PresenceThresholds,
) *Aggregator {
	return &Aggregator{
		subnetMatcher: matcher,
		ouiLookup:     ouiLookup,
		thresholds:    thresholds.Normalize(),
	}
}

func (a *Aggregator) Aggregate(snapshot *routeros.Snapshot) map[string]model.Observation {
	now := snapshot.FetchedAt.UTC()
	result := make(map[string]model.Observation)
	priorities := make(map[string]int)
	matcher := a.subnetMatcher.WithAddresses(snapshot.Addresses)

	// 1) DHCP pass
	for _, lease := range snapshot.DHCP {
		obs := getObservation(result, lease.MAC, now)
		appendSource(obs, model.SourceDHCP)

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
		obs.DHCPServer = lease.Server
		obs.DHCPStatus = strings.ToLower(strings.TrimSpace(lease.Status))

		if age := parseRouterOSAge(lease.LastSeen, now); age != nil {
			obs.DHCPLastSeen = age
			ts := now.Add(-*age)
			setLastSeen(obs, priorities, &ts, 2)
		}

		putRawSource(obs, model.SourceDHCP, lease)
		result[lease.MAC] = *obs
	}

	// 2) Wi-Fi pass
	for _, reg := range snapshot.WiFi {
		obs := getObservation(result, reg.MAC, now)
		appendSource(obs, model.SourceWiFi)

		obs.WiFiDriver = reg.Driver
		obs.WiFiInterface = reg.Interface
		obs.Interface = firstNonEmpty(obs.Interface, reg.Interface)
		obs.SSID = reg.SSID
		obs.WiFiAuthType = reg.AuthType

		if age := parseRouterOSAge(reg.LastActivity, now); age != nil {
			obs.WiFiLastActivity = age
			ts := now.Add(-*age)
			setLastSeen(obs, priorities, &ts, 1)
		} else {
			setLastSeen(obs, priorities, &now, 1)
		}
		if age := parseRouterOSAge(reg.Uptime, now); age != nil {
			obs.WiFiUptime = age
		}
		if signal, ok := parseSignal(reg.Signal); ok {
			obs.WiFiSignal = &signal
		}

		putRawSource(obs, model.SourceWiFi, reg)
		result[reg.MAC] = *obs
	}

	// 3) ARP pass
	for _, arp := range snapshot.ARP {
		obs := getObservation(result, arp.MAC, now)
		appendSource(obs, model.SourceARP)

		obs.ARPIP = arp.Address
		obs.ARPInterface = arp.Interface
		if obs.Interface == "" {
			obs.Interface = arp.Interface
		}
		obs.ARPIsComplete = strings.Contains(strings.ToUpper(arp.Flags), "C")
		if obs.IP == "" && arp.Address != "" {
			obs.IP = arp.Address
			subnetName := matcher.Match(arp.Address)
			if subnetName != "" {
				obs.LastSubnet = subnetName
			}
		}
		if obs.ARPIsComplete {
			setLastSeen(obs, priorities, &now, 3)
		}

		putRawSource(obs, model.SourceARP, arp)
		result[arp.MAC] = *obs
	}

	// 4) Bridge host pass
	for _, host := range snapshot.Bridge {
		obs := getObservation(result, host.MAC, now)
		appendSource(obs, model.SourceBridge)

		obs.Bridge = host.Bridge
		obs.BridgeHostPort = host.Interface
		if obs.Interface == "" {
			obs.Interface = host.Interface
		}
		if vlan, ok := parseVLAN(host.VID); ok {
			obs.BridgeHostVLAN = &vlan
		}
		setLastSeen(obs, priorities, &now, 4)

		putRawSource(obs, model.SourceBridge, host)
		result[host.MAC] = *obs
	}

	for mac, obs := range result {
		obs.Vendor = a.ouiLookup.Lookup(mac)
		obs.Generated = generatedName(mac, obs.Vendor)
		a.evaluateObservedStatus(&obs)
		result[mac] = obs
	}

	return result
}

func (a *Aggregator) evaluateObservedStatus(obs *model.Observation) {
	wifiPresent := strings.TrimSpace(obs.WiFiInterface) != ""
	wifiActive := wifiPresent && (obs.WiFiLastActivity == nil || *obs.WiFiLastActivity <= a.thresholds.WiFiIdleThreshold)

	dhcpBound := strings.EqualFold(strings.TrimSpace(obs.DHCPStatus), "bound")
	dhcpRecent := dhcpBound && obs.DHCPLastSeen != nil && *obs.DHCPLastSeen <= a.thresholds.DHCPRecentThreshold
	dhcpIdle := dhcpBound && !dhcpRecent

	arpValid := obs.ARPIsComplete && strings.TrimSpace(obs.ARPIP) != ""

	onlineByWifi := wifiActive
	onlineByDhcp := dhcpRecent
	onlineByArp := arpValid

	reasons := make([]string, 0, 5)
	if onlineByWifi {
		reasons = append(reasons, "wifi_active")
	}
	if onlineByDhcp {
		reasons = append(reasons, "dhcp_bound_recent")
	}
	if onlineByArp {
		reasons = append(reasons, "arp_complete")
	}
	if dhcpIdle {
		reasons = append(reasons, "dhcp_bound_stale")
	}

	switch {
	case onlineByWifi || onlineByDhcp || onlineByArp:
		obs.ConnectionStatus = model.ConnectionStatusOnline
	case dhcpIdle:
		obs.ConnectionStatus = model.ConnectionStatusIdleRecent
	case hasAnyTrace(obs):
		obs.ConnectionStatus = model.ConnectionStatusIdleRecent
	default:
		obs.ConnectionStatus = model.ConnectionStatusUnknown
	}

	obs.Online = obs.ConnectionStatus == model.ConnectionStatusOnline
	if len(reasons) == 0 {
		if hasAnyTrace(obs) {
			reasons = append(reasons, "trace_present")
		} else {
			reasons = append(reasons, "no_signal")
		}
	}
	obs.StatusReason = strings.Join(reasons, ";")
}

func hasAnyTrace(obs *model.Observation) bool {
	return len(obs.Sources) > 0 ||
		strings.TrimSpace(obs.IP) != "" ||
		strings.TrimSpace(obs.HostName) != "" ||
		strings.TrimSpace(obs.DHCPStatus) != "" ||
		strings.TrimSpace(obs.WiFiInterface) != "" ||
		strings.TrimSpace(obs.ARPIP) != "" ||
		strings.TrimSpace(obs.BridgeHostPort) != ""
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseSignal(value string) (int, bool) {
	normalized := strings.TrimSpace(strings.TrimSuffix(strings.ToLower(value), "dbm"))
	if normalized == "" {
		return 0, false
	}
	signal, err := strconv.Atoi(normalized)
	if err != nil {
		return 0, false
	}
	return signal, true
}

func parseVLAN(value string) (int, bool) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return 0, false
	}
	vlan, err := strconv.Atoi(normalized)
	if err != nil {
		return 0, false
	}
	return vlan, true
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
