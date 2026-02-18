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
		SELECT mac, online, last_seen_at, connected_since_at, last_ip, last_subnet, last_sources_json, updated_at
		FROM devices_state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]model.DeviceState{}
	for rows.Next() {
		var (
			state                                    model.DeviceState
			lastSeen, connectedSince, lastIP, subnet sql.NullString
			updatedAt                                string
		)
		if err := rows.Scan(&state.MAC, &state.Online, &lastSeen, &connectedSince, &lastIP, &subnet, &state.LastSourcesJSON, &updatedAt); err != nil {
			return nil, err
		}
		state.LastSeenAt = toTimePtr(lastSeen)
		state.ConnectedSinceAt = toTimePtr(connectedSince)
		state.LastIP = strPtr(lastIP)
		state.LastSubnet = strPtr(subnet)
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
		INSERT INTO devices_state (mac, online, last_seen_at, connected_since_at, last_ip, last_subnet, last_sources_json, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET
			online=excluded.online,
			last_seen_at=excluded.last_seen_at,
			connected_since_at=excluded.connected_since_at,
			last_ip=excluded.last_ip,
			last_subnet=excluded.last_subnet,
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
		if hasState {
			updated = state.UpdatedAt
			sources = ParseSourcesJSON(state.LastSourcesJSON)
			lastIP = state.LastIP
			subnet = state.LastSubnet
			lastSeen = state.LastSeenAt
			connectedSince = state.ConnectedSinceAt
			online = state.Online
		}

		firstSeen := cache.FirstSeenAt
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
			LastSources:      sources,
			CreatedAt:        createdAt,
			UpdatedAt:        updated,
			FirstSeenAt:      &firstSeen,
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
