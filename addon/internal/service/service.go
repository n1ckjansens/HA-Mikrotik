package service

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
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
	thresholds model.PresenceThresholds
	logger     *slog.Logger
}

func New(repo *storage.Repository, agg *aggregator.Aggregator, client RouterClient, cfg *configsync.Manager, logger *slog.Logger) *Service {
	return NewWithThresholds(repo, agg, client, cfg, logger, model.DefaultPresenceThresholds())
}

func NewWithThresholds(
	repo *storage.Repository,
	agg *aggregator.Aggregator,
	client RouterClient,
	cfg *configsync.Manager,
	logger *slog.Logger,
	thresholds model.PresenceThresholds,
) *Service {
	return &Service{
		repo:       repo,
		aggregator: agg,
		routeros:   client,
		config:     cfg,
		thresholds: thresholds.Normalize(),
		logger:     logger,
	}
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
	registered, err := s.repo.ListRegistered(ctx)
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
	deleteMACs := make([]string, 0)

	for mac := range allMACs {
		prev, hadPrev := prevStates[mac]
		obs, hasObs := observed[mac]
		_, isRegistered := registered[mac]

		if !isRegistered && (!hasObs || obs.ConnectionStatus != model.ConnectionStatusOnline) {
			deleteMACs = append(deleteMACs, mac)
			continue
		}

		next := model.DeviceState{
			MAC:              mac,
			UpdatedAt:        now,
			LastSourcesJSON:  "[]",
			ConnectionStatus: string(model.ConnectionStatusUnknown),
			StatusReason:     "no_signal",
		}
		if hadPrev {
			next = prev
			next.UpdatedAt = now
		}

		if hasObs {
			applyObservationToState(&next, obs)

			next.Online = obs.ConnectionStatus == model.ConnectionStatusOnline
			next.ConnectionStatus = string(obs.ConnectionStatus)
			next.StatusReason = obs.StatusReason
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
			if next.Online && (!hadPrev || !prev.Online) {
				started := now
				next.ConnectedSinceAt = &started
			}

			cache, hasCache := newCache[mac]
			if !hasCache {
				cache = model.DeviceNewCache{MAC: mac, FirstSeenAt: now}
			}
			if cache.FirstSeenAt.IsZero() {
				cache.FirstSeenAt = now
			}
			if obs.Vendor != "" && obs.Vendor != "Unknown" {
				cache.Vendor = obs.Vendor
			} else if strings.TrimSpace(cache.Vendor) == "" {
				cache.Vendor = obs.Vendor
			}
			if hostName := strings.TrimSpace(obs.HostName); hostName != "" {
				cache.GeneratedName = hostName
			} else if strings.TrimSpace(cache.GeneratedName) == "" || strings.HasPrefix(cache.GeneratedName, "Device-") {
				cache.GeneratedName = obs.Generated
			}
			cacheRows = append(cacheRows, cache)
		} else {
			next.Online = false
			next.LastSourcesJSON = "[]"
			status, reason := deriveStatusWithoutObservation(now, next, s.thresholds)
			next.ConnectionStatus = string(status)
			next.StatusReason = reason
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
	if len(deleteMACs) > 0 {
		if err := s.repo.DeleteStates(ctx, deleteMACs); err != nil {
			return err
		}
		if err := s.repo.DeleteNewCache(ctx, deleteMACs); err != nil {
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
	sort.SliceStable(filtered, func(i, j int) bool {
		a := filtered[i]
		b := filtered[j]

		aRank := statusRank(a.Status)
		bRank := statusRank(b.Status)
		if aRank != bRank {
			return aRank < bRank
		}

		aTime := primarySortTime(a)
		bTime := primarySortTime(b)
		if !aTime.Equal(bTime) {
			return aTime.After(bTime)
		}

		aName := strings.ToLower(a.Name)
		bName := strings.ToLower(b.Name)
		if aName != bName {
			return aName < bName
		}

		return a.MAC < b.MAC
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
	mac = strings.TrimSpace(mac)
	if decoded, err := url.PathUnescape(mac); err == nil {
		mac = decoded
	}
	mac = strings.ReplaceAll(mac, " ", "")
	mac = strings.ToUpper(strings.ReplaceAll(mac, "-", ":"))
	return mac
}

func statusRank(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "new":
		return 0
	case "registered":
		return 1
	default:
		return 2
	}
}

func primarySortTime(item model.DeviceView) time.Time {
	if item.Status == "registered" && item.CreatedAt != nil {
		return item.CreatedAt.UTC()
	}
	if item.Status == "new" && item.FirstSeenAt != nil {
		return item.FirstSeenAt.UTC()
	}
	return item.UpdatedAt.UTC()
}
