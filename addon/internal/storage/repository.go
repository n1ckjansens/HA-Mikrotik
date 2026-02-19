package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

var ErrNotFound = errors.New("not found")

func (r *Repository) LoadAllStates(ctx context.Context) (map[string]model.DeviceState, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			mac,
			online,
			last_seen_at,
			connected_since_at,
			last_ip,
			last_subnet,
			host_name,
			interface,
			bridge,
			ssid,
			dhcp_server,
			dhcp_status,
			dhcp_last_seen_sec,
			wifi_driver,
			wifi_interface,
			wifi_last_activity_sec,
			wifi_uptime_sec,
			wifi_auth_type,
			wifi_signal,
			arp_ip,
			arp_interface,
			arp_is_complete,
			bridge_host_port,
			bridge_host_vlan,
			connection_status,
			status_reason,
			last_sources_json,
			updated_at
		FROM devices_state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]model.DeviceState{}
	for rows.Next() {
		var (
			state model.DeviceState

			lastSeen, connectedSince sql.NullString
			lastIP, subnet           sql.NullString
			hostName, iface          sql.NullString
			bridge, ssid             sql.NullString
			dhcpServer, dhcpStatus   sql.NullString
			wifiDriver, wifiIface    sql.NullString
			wifiAuthType             sql.NullString
			arpIP, arpIface          sql.NullString
			bridgeHostPort           sql.NullString
			connectionStatus         sql.NullString
			statusReason             sql.NullString

			dhcpLastSeen, wifiLastAct, wifiUptime sql.NullInt64
			wifiSignal, bridgeHostVLAN            sql.NullInt64
			arpIsComplete                         sql.NullInt64

			updatedAt string
		)
		if err := rows.Scan(
			&state.MAC,
			&state.Online,
			&lastSeen,
			&connectedSince,
			&lastIP,
			&subnet,
			&hostName,
			&iface,
			&bridge,
			&ssid,
			&dhcpServer,
			&dhcpStatus,
			&dhcpLastSeen,
			&wifiDriver,
			&wifiIface,
			&wifiLastAct,
			&wifiUptime,
			&wifiAuthType,
			&wifiSignal,
			&arpIP,
			&arpIface,
			&arpIsComplete,
			&bridgeHostPort,
			&bridgeHostVLAN,
			&connectionStatus,
			&statusReason,
			&state.LastSourcesJSON,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		state.LastSeenAt = toTimePtr(lastSeen)
		state.ConnectedSinceAt = toTimePtr(connectedSince)
		state.LastIP = strPtr(lastIP)
		state.LastSubnet = strPtr(subnet)
		state.HostName = strPtr(hostName)
		state.Interface = strPtr(iface)
		state.Bridge = strPtr(bridge)
		state.SSID = strPtr(ssid)
		state.DHCPServer = strPtr(dhcpServer)
		state.DHCPStatus = strPtr(dhcpStatus)
		state.DHCPLastSeenSec = int64Ptr(dhcpLastSeen)
		state.WiFiDriver = strPtr(wifiDriver)
		state.WiFiInterface = strPtr(wifiIface)
		state.WiFiLastActSec = int64Ptr(wifiLastAct)
		state.WiFiUptimeSec = int64Ptr(wifiUptime)
		state.WiFiAuthType = strPtr(wifiAuthType)
		state.WiFiSignal = intPtr(wifiSignal)
		state.ARPIP = strPtr(arpIP)
		state.ARPInterface = strPtr(arpIface)
		state.ARPIsComplete = arpIsComplete.Valid && arpIsComplete.Int64 != 0
		state.BridgeHostPort = strPtr(bridgeHostPort)
		state.BridgeHostVLAN = intPtr(bridgeHostVLAN)
		if connectionStatus.Valid {
			state.ConnectionStatus = strings.TrimSpace(connectionStatus.String)
		}
		if statusReason.Valid {
			state.StatusReason = strings.TrimSpace(statusReason.String)
		}
		if state.ConnectionStatus == "" {
			state.ConnectionStatus = string(model.ConnectionStatusUnknown)
		}
		if state.StatusReason == "" {
			state.StatusReason = "no_signal"
		}
		if ts, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
			state.UpdatedAt = ts.UTC()
		}
		result[state.MAC] = state
	}
	return result, rows.Err()
}

func (r *Repository) UpsertStates(ctx context.Context, states []model.DeviceState) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO devices_state (
			mac,
			online,
			last_seen_at,
			connected_since_at,
			last_ip,
			last_subnet,
			host_name,
			interface,
			bridge,
			ssid,
			dhcp_server,
			dhcp_status,
			dhcp_last_seen_sec,
			wifi_driver,
			wifi_interface,
			wifi_last_activity_sec,
			wifi_uptime_sec,
			wifi_auth_type,
			wifi_signal,
			arp_ip,
			arp_interface,
			arp_is_complete,
			bridge_host_port,
			bridge_host_vlan,
			connection_status,
			status_reason,
			last_sources_json,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET
			online=excluded.online,
			last_seen_at=excluded.last_seen_at,
			connected_since_at=excluded.connected_since_at,
			last_ip=excluded.last_ip,
			last_subnet=excluded.last_subnet,
			host_name=excluded.host_name,
			interface=excluded.interface,
			bridge=excluded.bridge,
			ssid=excluded.ssid,
			dhcp_server=excluded.dhcp_server,
			dhcp_status=excluded.dhcp_status,
			dhcp_last_seen_sec=excluded.dhcp_last_seen_sec,
			wifi_driver=excluded.wifi_driver,
			wifi_interface=excluded.wifi_interface,
			wifi_last_activity_sec=excluded.wifi_last_activity_sec,
			wifi_uptime_sec=excluded.wifi_uptime_sec,
			wifi_auth_type=excluded.wifi_auth_type,
			wifi_signal=excluded.wifi_signal,
			arp_ip=excluded.arp_ip,
			arp_interface=excluded.arp_interface,
			arp_is_complete=excluded.arp_is_complete,
			bridge_host_port=excluded.bridge_host_port,
			bridge_host_vlan=excluded.bridge_host_vlan,
			connection_status=excluded.connection_status,
			status_reason=excluded.status_reason,
			last_sources_json=excluded.last_sources_json,
			updated_at=excluded.updated_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, state := range states {
		if _, err := stmt.ExecContext(
			ctx,
			state.MAC,
			state.Online,
			fromTimePtr(state.LastSeenAt),
			fromTimePtr(state.ConnectedSinceAt),
			fromStringPtr(state.LastIP),
			fromStringPtr(state.LastSubnet),
			fromStringPtr(state.HostName),
			fromStringPtr(state.Interface),
			fromStringPtr(state.Bridge),
			fromStringPtr(state.SSID),
			fromStringPtr(state.DHCPServer),
			fromStringPtr(state.DHCPStatus),
			fromInt64Ptr(state.DHCPLastSeenSec),
			fromStringPtr(state.WiFiDriver),
			fromStringPtr(state.WiFiInterface),
			fromInt64Ptr(state.WiFiLastActSec),
			fromInt64Ptr(state.WiFiUptimeSec),
			fromStringPtr(state.WiFiAuthType),
			fromIntPtr(state.WiFiSignal),
			fromStringPtr(state.ARPIP),
			fromStringPtr(state.ARPInterface),
			state.ARPIsComplete,
			fromStringPtr(state.BridgeHostPort),
			fromIntPtr(state.BridgeHostVLAN),
			defaultConnectionStatus(state.ConnectionStatus),
			defaultStatusReason(state.StatusReason),
			state.LastSourcesJSON,
			state.UpdatedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) UpsertNewCache(ctx context.Context, rows []model.DeviceNewCache) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO devices_new_cache(mac, first_seen_at, vendor, generated_name)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET
			vendor=excluded.vendor,
			generated_name=excluded.generated_name`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.ExecContext(
			ctx,
			row.MAC,
			row.FirstSeenAt.UTC().Format(time.RFC3339Nano),
			row.Vendor,
			row.GeneratedName,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) DeleteStates(ctx context.Context, macs []string) error {
	return r.deleteByMACs(ctx, "devices_state", macs)
}

func (r *Repository) DeleteNewCache(ctx context.Context, macs []string) error {
	return r.deleteByMACs(ctx, "devices_new_cache", macs)
}

func (r *Repository) deleteByMACs(ctx context.Context, table string, macs []string) error {
	if len(macs) == 0 {
		return nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(macs)), ",")
	args := make([]any, 0, len(macs))
	for _, mac := range macs {
		args = append(args, mac)
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE mac IN (%s)", table, placeholders)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) ListRegistered(ctx context.Context) (map[string]model.DeviceRegistered, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT mac, name, icon, comment, created_at, updated_at FROM devices_registered`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]model.DeviceRegistered{}
	for rows.Next() {
		var (
			item                 model.DeviceRegistered
			name, icon, comment  sql.NullString
			createdAt, updatedAt string
		)
		if err := rows.Scan(&item.MAC, &name, &icon, &comment, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.Name = strPtr(name)
		item.Icon = strPtr(icon)
		item.Comment = strPtr(comment)
		if ts, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
			item.CreatedAt = ts.UTC()
		}
		if ts, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
			item.UpdatedAt = ts.UTC()
		}
		result[item.MAC] = item
	}
	return result, rows.Err()
}

func (r *Repository) ListNewCache(ctx context.Context) (map[string]model.DeviceNewCache, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT mac, first_seen_at, vendor, generated_name FROM devices_new_cache`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]model.DeviceNewCache{}
	for rows.Next() {
		var row model.DeviceNewCache
		var firstSeenAt string
		if err := rows.Scan(&row.MAC, &firstSeenAt, &row.Vendor, &row.GeneratedName); err != nil {
			return nil, err
		}
		if ts, err := time.Parse(time.RFC3339Nano, firstSeenAt); err == nil {
			row.FirstSeenAt = ts.UTC()
		}
		result[row.MAC] = row
	}
	return result, rows.Err()
}

func (r *Repository) UpsertRegistered(ctx context.Context, mac string, name, icon, comment *string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO devices_registered(mac, name, icon, comment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET
			name=COALESCE(excluded.name, devices_registered.name),
			icon=COALESCE(excluded.icon, devices_registered.icon),
			comment=COALESCE(excluded.comment, devices_registered.comment),
			updated_at=excluded.updated_at`,
		mac, nullable(name), nullable(icon), nullable(comment), now, now,
	)
	return err
}

func (r *Repository) PatchRegistered(ctx context.Context, mac string, name, icon, comment *string) error {
	sets := []string{}
	args := []any{}
	if name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *name)
	}
	if icon != nil {
		sets = append(sets, "icon = ?")
		args = append(args, *icon)
	}
	if comment != nil {
		sets = append(sets, "comment = ?")
		args = append(args, *comment)
	}
	if len(sets) == 0 {
		return nil
	}
	sets = append(sets, "updated_at = ?")
	args = append(args, time.Now().UTC().Format(time.RFC3339Nano), mac)
	query := `UPDATE devices_registered SET ` + strings.Join(sets, ", ") + ` WHERE mac = ?`
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func nullable(v *string) any {
	if v == nil {
		return nil
	}
	value := strings.TrimSpace(*v)
	if value == "" {
		return nil
	}
	return value
}

func defaultConnectionStatus(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return string(model.ConnectionStatusUnknown)
	}
	return normalized
}

func defaultStatusReason(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "no_signal"
	}
	return normalized
}

func ParseSourcesJSON(v string) []string {
	if strings.TrimSpace(v) == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(v), &out); err != nil {
		return []string{}
	}
	return out
}

func EncodeSourcesJSON(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	body, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(body)
}

func EncodeRawSourcesJSON(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(body)
}

func MergeDeviceViews(
	states map[string]model.DeviceState,
	registered map[string]model.DeviceRegistered,
	newCache map[string]model.DeviceNewCache,
) []model.DeviceView {
	all := map[string]struct{}{}
	for mac := range states {
		all[mac] = struct{}{}
	}
	for mac := range registered {
		all[mac] = struct{}{}
	}

	result := make([]model.DeviceView, 0, len(all))
	for mac := range all {
		state, hasState := states[mac]
		reg, hasReg := registered[mac]
		cache := newCache[mac]

		name := cache.GeneratedName
		status := "new"
		var createdAt *time.Time
		var icon, comment *string
		if hasReg {
			status = "registered"
			if reg.Name != nil && strings.TrimSpace(*reg.Name) != "" {
				name = *reg.Name
			}
			createdAt = &reg.CreatedAt
			icon = reg.Icon
			comment = reg.Comment
		}

		vendor := cache.Vendor
		if vendor == "" {
			vendor = "Unknown"
		}

		updated := time.Now().UTC()
		sources := []string{}
		var lastIP, subnet *string
		var lastSeen, connectedSince *time.Time
		online := false
		var hostName, iface, bridge, ssid *string
		var dhcpServer, dhcpStatus *string
		var dhcpLastSeenSec *int64
		var wifiDriver, wifiIface *string
		var wifiLastActSec, wifiUptimeSec *int64
		var wifiAuthType *string
		var wifiSignal *int
		var arpIP, arpInterface *string
		arpIsComplete := false
		var bridgeHostPort *string
		var bridgeHostVLAN *int
		connectionStatus := string(model.ConnectionStatusUnknown)
		statusReason := "no_signal"
		if hasState {
			updated = state.UpdatedAt
			sources = ParseSourcesJSON(state.LastSourcesJSON)
			lastIP = state.LastIP
			subnet = state.LastSubnet
			lastSeen = state.LastSeenAt
			connectedSince = state.ConnectedSinceAt
			online = state.Online
			hostName = state.HostName
			iface = state.Interface
			bridge = state.Bridge
			ssid = state.SSID
			dhcpServer = state.DHCPServer
			dhcpStatus = state.DHCPStatus
			dhcpLastSeenSec = state.DHCPLastSeenSec
			wifiDriver = state.WiFiDriver
			wifiIface = state.WiFiInterface
			wifiLastActSec = state.WiFiLastActSec
			wifiUptimeSec = state.WiFiUptimeSec
			wifiAuthType = state.WiFiAuthType
			wifiSignal = state.WiFiSignal
			arpIP = state.ARPIP
			arpInterface = state.ARPInterface
			arpIsComplete = state.ARPIsComplete
			bridgeHostPort = state.BridgeHostPort
			bridgeHostVLAN = state.BridgeHostVLAN
			connectionStatus = defaultConnectionStatus(state.ConnectionStatus)
			statusReason = defaultStatusReason(state.StatusReason)
		}

		var firstSeenAt *time.Time
		if !cache.FirstSeenAt.IsZero() {
			firstSeen := cache.FirstSeenAt
			firstSeenAt = &firstSeen
		}
		view := model.DeviceView{
			MAC:              mac,
			Name:             name,
			Vendor:           vendor,
			Icon:             icon,
			Comment:          comment,
			Status:           status,
			Online:           online,
			LastSeenAt:       lastSeen,
			ConnectedSinceAt: connectedSince,
			LastIP:           lastIP,
			LastSubnet:       subnet,
			HostName:         hostName,
			Interface:        iface,
			Bridge:           bridge,
			SSID:             ssid,
			DHCPServer:       dhcpServer,
			DHCPStatus:       dhcpStatus,
			DHCPLastSeenSec:  dhcpLastSeenSec,
			WiFiDriver:       wifiDriver,
			WiFiInterface:    wifiIface,
			WiFiLastActSec:   wifiLastActSec,
			WiFiUptimeSec:    wifiUptimeSec,
			WiFiAuthType:     wifiAuthType,
			WiFiSignal:       wifiSignal,
			ARPIP:            arpIP,
			ARPInterface:     arpInterface,
			ARPIsComplete:    arpIsComplete,
			BridgeHostPort:   bridgeHostPort,
			BridgeHostVLAN:   bridgeHostVLAN,
			ConnectionStatus: connectionStatus,
			StatusReason:     statusReason,
			LastSources:      sources,
			CreatedAt:        createdAt,
			UpdatedAt:        updated,
			FirstSeenAt:      firstSeenAt,
		}
		result = append(result, view)
	}
	return result
}

func MustFindDevice(items []model.DeviceView, mac string) (model.DeviceView, error) {
	for _, item := range items {
		if strings.EqualFold(item.MAC, mac) {
			return item, nil
		}
	}
	return model.DeviceView{}, fmt.Errorf("%w: device %s", ErrNotFound, mac)
}
