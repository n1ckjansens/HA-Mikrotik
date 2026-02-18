package service

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/aggregator"
	"github.com/micro-ha/mikrotik-presence/addon/internal/configsync"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	"github.com/micro-ha/mikrotik-presence/addon/internal/routeros"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

var ErrIntegrationNotConfigured = errors.New("integration not configured")

type RouterClient interface {
	FetchSnapshot(ctx context.Context, cfg model.RouterConfig) (*routeros.Snapshot, error)
}

type Service struct {
	repo       *storage.Repository
	aggregator *aggregator.Aggregator
	routeros   RouterClient
	config     *configsync.Manager
	logger     *slog.Logger
}

func New(repo *storage.Repository, agg *aggregator.Aggregator, client RouterClient, cfg *configsync.Manager, logger *slog.Logger) *Service {
	return &Service{repo: repo, aggregator: agg, routeros: client, config: cfg, logger: logger}
}

type ListFilter struct {
	Status string
	Online *bool
	Query  string
}

func (s *Service) PollOnce(ctx context.Context) error {
	cfg, ok := s.config.Get()
	if !ok {
		return ErrIntegrationNotConfigured
	}

	snapshot, err := s.routeros.FetchSnapshot(ctx, cfg)
	if err != nil {
		return err
	}
	observed := s.aggregator.Aggregate(snapshot)
	if err := s.persistSnapshot(ctx, observed); err != nil {
		return err
	}
	return nil
}

func (s *Service) persistSnapshot(ctx context.Context, observed map[string]model.Observation) error {
	prevStates, err := s.repo.LoadAllStates(ctx)
	if err != nil {
		return err
	}
	newCache, err := s.repo.ListNewCache(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	allMACs := map[string]struct{}{}
	for mac := range prevStates {
		allMACs[mac] = struct{}{}
	}
	for mac := range observed {
		allMACs[mac] = struct{}{}
	}

	states := make([]model.DeviceState, 0, len(allMACs))
	cacheRows := make([]model.DeviceNewCache, 0, len(observed))

	for mac := range allMACs {
		prev, hadPrev := prevStates[mac]
		obs, hasObs := observed[mac]

		next := model.DeviceState{MAC: mac, UpdatedAt: now}
		if hadPrev {
			next = prev
			next.UpdatedAt = now
		}

		if hasObs {
			next.Online = obs.Online
			if obs.LastSeenAt != nil {
				next.LastSeenAt = obs.LastSeenAt
			}
			if obs.IP != "" {
				ip := obs.IP
				next.LastIP = &ip
			}
			if obs.LastSubnet != "" {
				subnet := obs.LastSubnet
				next.LastSubnet = &subnet
			}
			next.LastSourcesJSON = storage.EncodeSourcesJSON(obs.Sources)
			if obs.Online && (!hadPrev || !prev.Online) {
				started := now
				next.ConnectedSinceAt = &started
			}
			if !obs.Online {
				next.LastSourcesJSON = "[]"
			}

			cache, hasCache := newCache[mac]
			if !hasCache {
				cache = model.DeviceNewCache{MAC: mac, FirstSeenAt: now}
			}
			if cache.FirstSeenAt.IsZero() {
				cache.FirstSeenAt = now
			}
			cache.Vendor = obs.Vendor
			cache.GeneratedName = obs.Generated
			cacheRows = append(cacheRows, cache)
		} else {
			next.Online = false
			next.LastSourcesJSON = "[]"
		}
		states = append(states, next)
	}

	if err := s.repo.UpsertStates(ctx, states); err != nil {
		return err
	}
	if len(cacheRows) > 0 {
		if err := s.repo.UpsertNewCache(ctx, cacheRows); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListDevices(ctx context.Context, filter ListFilter) ([]model.DeviceView, error) {
	states, err := s.repo.LoadAllStates(ctx)
	if err != nil {
		return nil, err
	}
	registered, err := s.repo.ListRegistered(ctx)
	if err != nil {
		return nil, err
	}
	newCache, err := s.repo.ListNewCache(ctx)
	if err != nil {
		return nil, err
	}

	items := storage.MergeDeviceViews(states, registered, newCache)
	filtered := filterViews(items, filter)
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Online != filtered[j].Online {
			return filtered[i].Online
		}
		if filtered[i].Status != filtered[j].Status {
			return filtered[i].Status < filtered[j].Status
		}
		return filtered[i].Name < filtered[j].Name
	})
	return filtered, nil
}

func (s *Service) GetDevice(ctx context.Context, mac string) (model.DeviceView, error) {
	items, err := s.ListDevices(ctx, ListFilter{})
	if err != nil {
		return model.DeviceView{}, err
	}
	return storage.MustFindDevice(items, normalizeMAC(mac))
}

type RegisterInput struct {
	Name    *string `json:"name"`
	Icon    *string `json:"icon"`
	Comment *string `json:"comment"`
}

func (s *Service) RegisterDevice(ctx context.Context, mac string, in RegisterInput) error {
	return s.repo.UpsertRegistered(ctx, normalizeMAC(mac), in.Name, in.Icon, in.Comment)
}

func (s *Service) PatchDevice(ctx context.Context, mac string, in RegisterInput) error {
	return s.repo.PatchRegistered(ctx, normalizeMAC(mac), in.Name, in.Icon, in.Comment)
}

func filterViews(items []model.DeviceView, filter ListFilter) []model.DeviceView {
	status := strings.ToLower(strings.TrimSpace(filter.Status))
	query := strings.ToLower(strings.TrimSpace(filter.Query))

	result := make([]model.DeviceView, 0, len(items))
	for _, item := range items {
		if status == "new" && item.Status != "new" {
			continue
		}
		if status == "registered" && item.Status != "registered" {
			continue
		}
		if filter.Online != nil && item.Online != *filter.Online {
			continue
		}
		if query != "" && !matchesQuery(item, query) {
			continue
		}
		result = append(result, item)
	}
	return result
}

func matchesQuery(item model.DeviceView, query string) bool {
	if strings.Contains(strings.ToLower(item.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(item.MAC), query) {
		return true
	}
	if strings.Contains(strings.ToLower(item.Vendor), query) {
		return true
	}
	if item.LastIP != nil && strings.Contains(strings.ToLower(*item.LastIP), query) {
		return true
	}
	return false
}

func normalizeMAC(mac string) string {
	mac = strings.TrimSpace(strings.ToUpper(strings.ReplaceAll(mac, "-", ":")))
	return mac
}
