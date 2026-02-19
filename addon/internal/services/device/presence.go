package device

import (
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

func applyObservationToState(state *model.DeviceState, obs model.Observation) {
	state.HostName = strPtrOrNil(obs.HostName)
	state.Interface = strPtrOrNil(obs.Interface)
	state.Bridge = strPtrOrNil(obs.Bridge)
	state.SSID = strPtrOrNil(obs.SSID)

	state.DHCPServer = strPtrOrNil(obs.DHCPServer)
	state.DHCPStatus = strPtrOrNil(obs.DHCPStatus)
	state.DHCPLastSeenSec = durationToSeconds(obs.DHCPLastSeen)

	state.WiFiDriver = strPtrOrNil(obs.WiFiDriver)
	state.WiFiInterface = strPtrOrNil(obs.WiFiInterface)
	state.WiFiLastActSec = durationToSeconds(obs.WiFiLastActivity)
	state.WiFiUptimeSec = durationToSeconds(obs.WiFiUptime)
	state.WiFiAuthType = strPtrOrNil(obs.WiFiAuthType)
	state.WiFiSignal = copyIntPtr(obs.WiFiSignal)

	state.ARPIP = strPtrOrNil(obs.ARPIP)
	state.ARPInterface = strPtrOrNil(obs.ARPInterface)
	state.ARPIsComplete = obs.ARPIsComplete

	state.BridgeHostPort = strPtrOrNil(obs.BridgeHostPort)
	state.BridgeHostVLAN = copyIntPtr(obs.BridgeHostVLAN)
}

func deriveStatusWithoutObservation(
	now time.Time,
	state model.DeviceState,
	thresholds model.PresenceThresholds,
) (model.ConnectionStatus, string) {
	if state.LastSeenAt != nil && now.Sub(state.LastSeenAt.UTC()) > thresholds.OfflineHardThreshold {
		return model.ConnectionStatusOffline, "no_signal;offline_hard_threshold_exceeded"
	}
	if hasAnyHistoricalTrace(state) {
		return model.ConnectionStatusIdleRecent, "no_current_signal;historical_trace_present"
	}
	return model.ConnectionStatusUnknown, "no_signal"
}

func hasAnyHistoricalTrace(state model.DeviceState) bool {
	return state.LastSeenAt != nil ||
		state.LastIP != nil ||
		state.HostName != nil ||
		state.DHCPStatus != nil ||
		state.WiFiInterface != nil ||
		state.ARPIP != nil ||
		state.BridgeHostPort != nil ||
		state.Bridge != nil
}

func durationToSeconds(value *time.Duration) *int64 {
	if value == nil {
		return nil
	}
	seconds := int64(value.Round(time.Second) / time.Second)
	if seconds < 0 {
		seconds = -seconds
	}
	return &seconds
}

func strPtrOrNil(value string) *string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func copyIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
