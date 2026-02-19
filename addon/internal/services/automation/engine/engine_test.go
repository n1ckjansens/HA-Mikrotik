package engine

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	"github.com/micro-ha/mikrotik-presence/addon/internal/services/automation/registry"
)

type memoryRepository struct {
	templates    map[string]automationdomain.CapabilityTemplate
	states       map[string]automationdomain.DeviceCapability
	globalStates map[string]automationdomain.GlobalCapability
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		templates:    map[string]automationdomain.CapabilityTemplate{},
		states:       map[string]automationdomain.DeviceCapability{},
		globalStates: map[string]automationdomain.GlobalCapability{},
	}
}

func (r *memoryRepository) ListTemplates(
	ctx context.Context,
	search string,
	category string,
) ([]automationdomain.CapabilityTemplate, error) {
	items := make([]automationdomain.CapabilityTemplate, 0, len(r.templates))
	for _, template := range r.templates {
		items = append(items, template)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (r *memoryRepository) GetTemplate(ctx context.Context, id string) (automationdomain.CapabilityTemplate, error) {
	item, ok := r.templates[id]
	if !ok {
		return automationdomain.CapabilityTemplate{}, automationdomain.ErrNotFound
	}
	return item, nil
}

func (r *memoryRepository) CreateTemplate(ctx context.Context, template automationdomain.CapabilityTemplate) error {
	if _, exists := r.templates[template.ID]; exists {
		return errors.New("unique constraint")
	}
	r.templates[template.ID] = template
	return nil
}

func (r *memoryRepository) UpdateTemplate(ctx context.Context, template automationdomain.CapabilityTemplate) error {
	if _, exists := r.templates[template.ID]; !exists {
		return automationdomain.ErrNotFound
	}
	r.templates[template.ID] = template
	return nil
}

func (r *memoryRepository) DeleteTemplate(ctx context.Context, id string) error {
	if _, exists := r.templates[id]; !exists {
		return automationdomain.ErrNotFound
	}
	delete(r.templates, id)
	return nil
}

func (r *memoryRepository) UpsertDeviceCapabilityState(
	ctx context.Context,
	state automationdomain.DeviceCapability,
) error {
	r.states[stateKey(state.DeviceID, state.CapabilityID)] = state
	return nil
}

func (r *memoryRepository) GetDeviceCapabilityState(
	ctx context.Context,
	deviceID string,
	capabilityID string,
) (automationdomain.DeviceCapability, bool, error) {
	item, ok := r.states[stateKey(deviceID, capabilityID)]
	return item, ok, nil
}

func (r *memoryRepository) ListDeviceCapabilityStates(
	ctx context.Context,
	deviceID string,
) (map[string]automationdomain.DeviceCapability, error) {
	out := map[string]automationdomain.DeviceCapability{}
	for _, item := range r.states {
		if strings.EqualFold(item.DeviceID, deviceID) {
			out[item.CapabilityID] = item
		}
	}
	return out, nil
}

func (r *memoryRepository) ListCapabilityDeviceStates(
	ctx context.Context,
	capabilityID string,
) (map[string]automationdomain.DeviceCapability, error) {
	out := map[string]automationdomain.DeviceCapability{}
	for _, item := range r.states {
		if item.CapabilityID == capabilityID {
			out[item.DeviceID] = item
		}
	}
	return out, nil
}

func (r *memoryRepository) GetGlobalCapability(
	ctx context.Context,
	capabilityID string,
) (*automationdomain.GlobalCapability, error) {
	item, ok := r.globalStates[capabilityID]
	if !ok {
		return nil, nil
	}
	out := item
	return &out, nil
}

func (r *memoryRepository) SaveGlobalCapability(
	ctx context.Context,
	capability *automationdomain.GlobalCapability,
) error {
	if capability == nil {
		return errors.New("global capability is nil")
	}
	r.globalStates[capability.CapabilityID] = *capability
	return nil
}

func (r *memoryRepository) ListGlobalCapabilities(
	ctx context.Context,
) ([]automationdomain.GlobalCapability, error) {
	out := make([]automationdomain.GlobalCapability, 0, len(r.globalStates))
	for _, item := range r.globalStates {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CapabilityID < out[j].CapabilityID })
	return out, nil
}

func stateKey(deviceID string, capabilityID string) string {
	return strings.ToUpper(deviceID) + "|" + capabilityID
}

type fakeDeviceService struct {
	devices map[string]devicedomain.Device
}

func (s *fakeDeviceService) GetDevice(ctx context.Context, mac string) (devicedomain.Device, error) {
	device, ok := s.devices[strings.ToUpper(mac)]
	if !ok {
		return devicedomain.Device{}, devicedomain.ErrDeviceNotFound
	}
	return device, nil
}

func (s *fakeDeviceService) ListDevices(
	ctx context.Context,
	filter devicedomain.ListFilter,
) ([]devicedomain.Device, error) {
	out := make([]devicedomain.Device, 0, len(s.devices))
	for _, item := range s.devices {
		out = append(out, item)
	}
	return out, nil
}

type fakeConfigProvider struct {
	cfg model.RouterConfig
	ok  bool
}

func (f fakeConfigProvider) Get() (model.RouterConfig, bool) {
	return f.cfg, f.ok
}

type fakeRouterClient struct {
	addCalls      int
	removeCalls   int
	membershipMap map[string]bool
}

func (f *fakeRouterClient) AddAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	f.addCalls++
	return nil
}

func (f *fakeRouterClient) RemoveAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	f.removeCalls++
	return nil
}

func (f *fakeRouterClient) AddressListContains(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) (bool, error) {
	key := list + "|" + address
	return f.membershipMap[key], nil
}

type fakeAction struct {
	id         string
	execCalled int
}

func (a *fakeAction) ID() string { return a.id }

func (a *fakeAction) Metadata() automationdomain.ActionMetadata {
	return automationdomain.ActionMetadata{
		ID:          a.id,
		Label:       a.id,
		Description: a.id,
	}
}

func (a *fakeAction) Validate(target automationdomain.AutomationTarget, params map[string]any) error {
	return nil
}

func (a *fakeAction) Execute(
	ctx context.Context,
	execCtx automationdomain.ActionExecutionContext,
	params map[string]any,
) error {
	a.execCalled++
	return nil
}

type fakeStateSource struct {
	id    string
	value bool
}

func (s *fakeStateSource) ID() string { return s.id }

func (s *fakeStateSource) Metadata() automationdomain.StateSourceMetadata {
	return automationdomain.StateSourceMetadata{
		ID:         s.id,
		Label:      s.id,
		OutputType: "boolean",
	}
}

func (s *fakeStateSource) Validate(target automationdomain.AutomationTarget, params map[string]any) error {
	return nil
}

func (s *fakeStateSource) Read(
	ctx context.Context,
	sourceCtx automationdomain.StateSourceContext,
	params map[string]any,
) (any, error) {
	return s.value, nil
}

func TestEngineSetCapabilityStateExecutesActionsAndPersistsState(t *testing.T) {
	repo := newMemoryRepository()
	deviceService := &fakeDeviceService{
		devices: map[string]devicedomain.Device{
			"AA:BB:CC:DD:EE:01": {
				MAC:    "AA:BB:CC:DD:EE:01",
				Name:   "Test device",
				Online: true,
			},
		},
	}
	action := &fakeAction{id: "test.action"}
	reg := registry.New()
	reg.RegisterAction(action)

	repo.templates["routing.vpn"] = automationdomain.CapabilityTemplate{
		ID:           "routing.vpn",
		Label:        "VPN",
		Control:      automationdomain.CapabilityControl{Type: automationdomain.ControlSwitch, Options: []automationdomain.CapabilityControlOption{{Value: "on", Label: "On"}, {Value: "off", Label: "Off"}}},
		DefaultState: "off",
		States: map[string]automationdomain.CapabilityStateConfig{
			"on": {
				Label:          "On",
				ActionsOnEnter: []automationdomain.ActionInstance{{ID: "a1", TypeID: "test.action", Params: map[string]any{}}},
			},
			"off": {Label: "Off"},
		},
	}

	engine := New(
		repo,
		deviceService,
		reg,
		fakeConfigProvider{ok: true, cfg: model.RouterConfig{Host: "router.local"}},
		&fakeRouterClient{membershipMap: map[string]bool{}},
		nil,
	)

	result, err := engine.SetCapabilityState(context.Background(), automationdomain.CapabilityTargetRef{
		Scope:    automationdomain.ScopeDevice,
		DeviceID: "AA:BB:CC:DD:EE:01",
	}, "routing.vpn", "on")
	if err != nil {
		t.Fatalf("SetCapabilityState returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK result")
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(result.Warnings))
	}
	if action.execCalled != 1 {
		t.Fatalf("expected action execution once, got %d", action.execCalled)
	}

	stored, ok, err := repo.GetDeviceCapabilityState(context.Background(), "AA:BB:CC:DD:EE:01", "routing.vpn")
	if err != nil {
		t.Fatalf("GetDeviceCapabilityState returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected persisted state")
	}
	if stored.State != "on" || !stored.Enabled {
		t.Fatalf("unexpected stored state: %+v", stored)
	}
}

func TestEngineSyncOnceMapsExternalStateToInternalState(t *testing.T) {
	repo := newMemoryRepository()
	deviceService := &fakeDeviceService{
		devices: map[string]devicedomain.Device{
			"AA:BB:CC:DD:EE:02": {
				MAC:    "AA:BB:CC:DD:EE:02",
				Name:   "Sync device",
				Online: true,
			},
		},
	}

	source := &fakeStateSource{id: "test.source", value: true}
	reg := registry.New()
	reg.RegisterStateSource(source)

	repo.templates["routing.sync"] = automationdomain.CapabilityTemplate{
		ID:           "routing.sync",
		Label:        "Sync capability",
		Control:      automationdomain.CapabilityControl{Type: automationdomain.ControlSwitch, Options: []automationdomain.CapabilityControlOption{{Value: "on", Label: "On"}, {Value: "off", Label: "Off"}}},
		DefaultState: "off",
		States: map[string]automationdomain.CapabilityStateConfig{
			"on":  {Label: "On"},
			"off": {Label: "Off"},
		},
		Sync: &automationdomain.CapabilitySyncConfig{
			Enabled: true,
			Source: automationdomain.CapabilitySyncSource{
				TypeID: "test.source",
				Params: map[string]any{},
			},
			Mapping: automationdomain.CapabilitySyncMapping{
				WhenTrue:  "on",
				WhenFalse: "off",
			},
			Mode:                 "external_truth",
			TriggerActionsOnSync: false,
		},
	}
	_ = repo.UpsertDeviceCapabilityState(context.Background(), automationdomain.DeviceCapability{
		DeviceID:     "AA:BB:CC:DD:EE:02",
		CapabilityID: "routing.sync",
		Enabled:      true,
		State:        "off",
		UpdatedAt:    time.Now().UTC(),
	})

	engine := New(
		repo,
		deviceService,
		reg,
		fakeConfigProvider{ok: true, cfg: model.RouterConfig{Host: "router.local"}},
		&fakeRouterClient{membershipMap: map[string]bool{}},
		nil,
	)

	if err := engine.SyncOnce(context.Background()); err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}

	stored, ok, err := repo.GetDeviceCapabilityState(context.Background(), "AA:BB:CC:DD:EE:02", "routing.sync")
	if err != nil {
		t.Fatalf("GetDeviceCapabilityState returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected persisted state")
	}
	if stored.State != "on" {
		t.Fatalf("expected synced state 'on', got %q", stored.State)
	}
}

func TestEngineSyncOnceCanTriggerActions(t *testing.T) {
	repo := newMemoryRepository()
	deviceService := &fakeDeviceService{
		devices: map[string]devicedomain.Device{
			"AA:BB:CC:DD:EE:03": {
				MAC:    "AA:BB:CC:DD:EE:03",
				Name:   "Action sync device",
				Online: true,
			},
		},
	}

	source := &fakeStateSource{id: "test.source", value: true}
	action := &fakeAction{id: "test.action"}
	reg := registry.New()
	reg.RegisterStateSource(source)
	reg.RegisterAction(action)

	repo.templates["routing.sync.actions"] = automationdomain.CapabilityTemplate{
		ID:           "routing.sync.actions",
		Label:        "Sync action capability",
		Control:      automationdomain.CapabilityControl{Type: automationdomain.ControlSwitch, Options: []automationdomain.CapabilityControlOption{{Value: "on", Label: "On"}, {Value: "off", Label: "Off"}}},
		DefaultState: "off",
		States: map[string]automationdomain.CapabilityStateConfig{
			"on": {
				Label: "On",
				ActionsOnEnter: []automationdomain.ActionInstance{
					{ID: "a1", TypeID: "test.action", Params: map[string]any{}},
				},
			},
			"off": {Label: "Off"},
		},
		Sync: &automationdomain.CapabilitySyncConfig{
			Enabled: true,
			Source: automationdomain.CapabilitySyncSource{
				TypeID: "test.source",
				Params: map[string]any{},
			},
			Mapping: automationdomain.CapabilitySyncMapping{
				WhenTrue:  "on",
				WhenFalse: "off",
			},
			Mode:                 "external_truth",
			TriggerActionsOnSync: true,
		},
	}
	_ = repo.UpsertDeviceCapabilityState(context.Background(), automationdomain.DeviceCapability{
		DeviceID:     "AA:BB:CC:DD:EE:03",
		CapabilityID: "routing.sync.actions",
		Enabled:      true,
		State:        "off",
		UpdatedAt:    time.Now().UTC(),
	})

	engine := New(
		repo,
		deviceService,
		reg,
		fakeConfigProvider{ok: true, cfg: model.RouterConfig{Host: "router.local"}},
		&fakeRouterClient{membershipMap: map[string]bool{}},
		nil,
	)

	if err := engine.SyncOnce(context.Background()); err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}

	if action.execCalled != 1 {
		t.Fatalf("expected sync-triggered action execution once, got %d", action.execCalled)
	}
	stored, ok, err := repo.GetDeviceCapabilityState(context.Background(), "AA:BB:CC:DD:EE:03", "routing.sync.actions")
	if err != nil {
		t.Fatalf("GetDeviceCapabilityState returned error: %v", err)
	}
	if !ok || stored.State != "on" {
		t.Fatalf("unexpected stored sync state: %+v", stored)
	}
}

func TestEngineSetCapabilityStateGlobalPersistsState(t *testing.T) {
	repo := newMemoryRepository()
	action := &fakeAction{id: "test.action"}
	reg := registry.New()
	reg.RegisterAction(action)

	repo.templates["global.vpn_profile"] = automationdomain.CapabilityTemplate{
		ID:           "global.vpn_profile",
		Label:        "Global VPN profile",
		Scope:        automationdomain.ScopeGlobal,
		Control:      automationdomain.CapabilityControl{Type: automationdomain.ControlSwitch, Options: []automationdomain.CapabilityControlOption{{Value: "on", Label: "On"}, {Value: "off", Label: "Off"}}},
		DefaultState: "off",
		States: map[string]automationdomain.CapabilityStateConfig{
			"on": {
				Label:          "On",
				ActionsOnEnter: []automationdomain.ActionInstance{{ID: "a1", TypeID: "test.action", Params: map[string]any{}}},
			},
			"off": {Label: "Off"},
		},
	}

	engine := New(
		repo,
		&fakeDeviceService{devices: map[string]devicedomain.Device{}},
		reg,
		fakeConfigProvider{ok: true, cfg: model.RouterConfig{Host: "router.local"}},
		&fakeRouterClient{membershipMap: map[string]bool{}},
		nil,
	)

	result, err := engine.SetCapabilityState(context.Background(), automationdomain.CapabilityTargetRef{
		Scope: automationdomain.ScopeGlobal,
	}, "global.vpn_profile", "on")
	if err != nil {
		t.Fatalf("SetCapabilityState returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected OK result")
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(result.Warnings))
	}
	if action.execCalled != 1 {
		t.Fatalf("expected action execution once, got %d", action.execCalled)
	}

	stored, err := repo.GetGlobalCapability(context.Background(), "global.vpn_profile")
	if err != nil {
		t.Fatalf("GetGlobalCapability returned error: %v", err)
	}
	if stored == nil {
		t.Fatalf("expected persisted global state")
	}
	if stored.State != "on" || !stored.Enabled {
		t.Fatalf("unexpected stored state: %+v", stored)
	}
}

func TestEngineSyncOnceGlobalUpdatesState(t *testing.T) {
	repo := newMemoryRepository()
	source := &fakeStateSource{id: "test.source", value: true}
	reg := registry.New()
	reg.RegisterStateSource(source)

	repo.templates["global.vpn_profile"] = automationdomain.CapabilityTemplate{
		ID:           "global.vpn_profile",
		Label:        "Global VPN profile",
		Scope:        automationdomain.ScopeGlobal,
		Control:      automationdomain.CapabilityControl{Type: automationdomain.ControlSwitch, Options: []automationdomain.CapabilityControlOption{{Value: "on", Label: "On"}, {Value: "off", Label: "Off"}}},
		DefaultState: "off",
		States: map[string]automationdomain.CapabilityStateConfig{
			"on":  {Label: "On"},
			"off": {Label: "Off"},
		},
		Sync: &automationdomain.CapabilitySyncConfig{
			Enabled: true,
			Source: automationdomain.CapabilitySyncSource{
				TypeID: "test.source",
				Params: map[string]any{},
			},
			Mapping: automationdomain.CapabilitySyncMapping{
				WhenTrue:  "on",
				WhenFalse: "off",
			},
			Mode:                 "external_truth",
			TriggerActionsOnSync: false,
		},
	}
	_ = repo.SaveGlobalCapability(context.Background(), &automationdomain.GlobalCapability{
		CapabilityID: "global.vpn_profile",
		Enabled:      true,
		State:        "off",
	})

	engine := New(
		repo,
		&fakeDeviceService{devices: map[string]devicedomain.Device{}},
		reg,
		fakeConfigProvider{ok: true, cfg: model.RouterConfig{Host: "router.local"}},
		&fakeRouterClient{membershipMap: map[string]bool{}},
		nil,
	)

	if err := engine.SyncOnce(context.Background()); err != nil {
		t.Fatalf("SyncOnce returned error: %v", err)
	}

	stored, err := repo.GetGlobalCapability(context.Background(), "global.vpn_profile")
	if err != nil {
		t.Fatalf("GetGlobalCapability returned error: %v", err)
	}
	if stored == nil {
		t.Fatalf("expected persisted global state")
	}
	if stored.State != "on" {
		t.Fatalf("expected synced state 'on', got %q", stored.State)
	}
}
