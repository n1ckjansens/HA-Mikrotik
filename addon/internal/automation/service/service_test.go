package service

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/repository"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	presence "github.com/micro-ha/mikrotik-presence/addon/internal/service"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

type fakeDeviceService struct {
	device model.DeviceView
}

func (f *fakeDeviceService) GetDevice(ctx context.Context, mac string) (model.DeviceView, error) {
	return f.device, nil
}

func (f *fakeDeviceService) ListDevices(ctx context.Context, filter presence.ListFilter) ([]model.DeviceView, error) {
	return []model.DeviceView{f.device}, nil
}

type fakeConfigProvider struct {
	cfg model.RouterConfig
	ok  bool
}

func (f *fakeConfigProvider) Get() (model.RouterConfig, bool) {
	return f.cfg, f.ok
}

type fakeRouterClient struct {
	addCalls    int
	removeCalls int
}

func (f *fakeRouterClient) AddAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	f.addCalls += 1
	return nil
}

func (f *fakeRouterClient) RemoveAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	f.removeCalls += 1
	return nil
}

func newServiceUnderTest(t *testing.T) (*Service, *repository.Repository, *fakeRouterClient) {
	t.Helper()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	baseRepo, err := storage.New(ctx, filepath.Join(t.TempDir(), "service.db"), logger)
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	t.Cleanup(func() {
		_ = baseRepo.Close()
	})

	repo := repository.New(baseRepo.SQLDB(), logger)
	router := &fakeRouterClient{}

	device := model.DeviceView{
		MAC:    "AA:BB:CC:DD:EE:AA",
		Name:   "Alice phone",
		Online: true,
		LastIP: stringPtr("192.168.88.101"),
	}

	svc := New(
		repo,
		&fakeDeviceService{device: device},
		&fakeConfigProvider{
			cfg: model.RouterConfig{Host: "router.local", Username: "u", Password: "p", SSL: false},
			ok:  true,
		},
		router,
		logger,
	)

	template := domain.CapabilityTemplate{
		ID:          "routing.vpn",
		Label:       "VPN routing",
		Description: "Route this device via VPN",
		Category:    "Routing",
		Control: domain.CapabilityControl{
			Type:    domain.ControlSwitch,
			Options: []domain.CapabilityControlOption{{Value: "on", Label: "On"}, {Value: "off", Label: "Off"}},
		},
		States: map[string]domain.CapabilityStateConfig{
			"on": {
				Label: "On",
				ActionsOnEnter: []domain.ActionInstance{
					{
						ID:     "a1",
						TypeID: "mikrotik.address_list.set_membership",
						Params: map[string]any{
							"list":   "VPN_CLIENTS",
							"mode":   "add",
							"target": "device.ip",
						},
					},
				},
			},
			"off": {Label: "Off", ActionsOnEnter: []domain.ActionInstance{}},
		},
		DefaultState: "off",
		HAExpose: domain.HAExposeConfig{
			Enabled: false,
		},
	}
	if err := repo.CreateTemplate(ctx, template); err != nil {
		t.Fatalf("create template: %v", err)
	}

	return svc, repo, router
}

func TestSetCapabilityState_ExecutesActionsAndPersistsState(t *testing.T) {
	ctx := context.Background()
	svc, repo, router := newServiceUnderTest(t)

	result, err := svc.SetCapabilityState(ctx, "AA:BB:CC:DD:EE:AA", "routing.vpn", "on")
	if err != nil {
		t.Fatalf("set state: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected ok result")
	}
	if len(result.Warnings) > 0 {
		t.Fatalf("unexpected warnings: %+v", result.Warnings)
	}
	if router.addCalls != 1 {
		t.Fatalf("expected 1 add call, got %d", router.addCalls)
	}

	stored, ok, err := repo.GetDeviceCapabilityState(ctx, "AA:BB:CC:DD:EE:AA", "routing.vpn")
	if err != nil {
		t.Fatalf("get stored state: %v", err)
	}
	if !ok {
		t.Fatalf("state not stored")
	}
	if stored.State != "on" || !stored.Enabled {
		t.Fatalf("unexpected stored state: %+v", stored)
	}
}

func TestSetCapabilityState_ReturnsWarningsWhenActionCannotResolvePlaceholder(t *testing.T) {
	ctx := context.Background()
	svc, repo, _ := newServiceUnderTest(t)

	template, err := repo.GetTemplate(ctx, "routing.vpn")
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	template.States["on"] = domain.CapabilityStateConfig{
		Label: "On",
		ActionsOnEnter: []domain.ActionInstance{
			{
				ID:     "a1",
				TypeID: "mikrotik.address_list.set_membership",
				Params: map[string]any{
					"list":       "VPN_CLIENTS",
					"mode":       "add",
					"target":     "literal_ip",
					"literal_ip": "{{device.unknown}}",
				},
			},
		},
	}
	if err := repo.UpdateTemplate(ctx, template); err != nil {
		t.Fatalf("update template: %v", err)
	}

	result, err := svc.SetCapabilityState(ctx, "AA:BB:CC:DD:EE:AA", "routing.vpn", "on")
	if err != nil {
		t.Fatalf("set state: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected ok result")
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected warnings")
	}
}

func stringPtr(value string) *string {
	return &value
}
