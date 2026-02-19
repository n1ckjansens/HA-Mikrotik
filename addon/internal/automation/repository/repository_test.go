package repository

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

func TestRepository_TemplateAndDeviceStateCRUD(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	baseRepo, err := storage.New(ctx, filepath.Join(t.TempDir(), "automation.db"), logger)
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	t.Cleanup(func() {
		_ = baseRepo.Close()
	})

	repo := New(baseRepo.SQLDB(), logger)
	template := domain.CapabilityTemplate{
		ID:          "routing.vpn",
		Label:       "VPN routing",
		Description: "Route this device via VPN",
		Category:    "Routing",
		Control: domain.CapabilityControl{
			Type: domain.ControlSwitch,
			Options: []domain.CapabilityControlOption{
				{Value: "on", Label: "On"},
				{Value: "off", Label: "Off"},
			},
		},
		States: map[string]domain.CapabilityStateConfig{
			"on":  {Label: "On", ActionsOnEnter: []domain.ActionInstance{}},
			"off": {Label: "Off", ActionsOnEnter: []domain.ActionInstance{}},
		},
		DefaultState: "off",
		HAExpose: domain.HAExposeConfig{
			Enabled:      true,
			EntityType:   "switch",
			EntitySuffix: "vpn",
			NameTemplate: "{{device.name}} VPN",
		},
	}

	if err := repo.CreateTemplate(ctx, template); err != nil {
		t.Fatalf("create template: %v", err)
	}

	fetched, err := repo.GetTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	if fetched.ID != template.ID {
		t.Fatalf("unexpected template id: %s", fetched.ID)
	}

	templates, err := repo.ListTemplates(ctx, "vpn", "routing")
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}

	state := domain.DeviceCapabilityState{
		DeviceID:     "AA:BB:CC:DD:EE:01",
		CapabilityID: template.ID,
		Enabled:      true,
		State:        "on",
	}
	if err := repo.UpsertDeviceCapabilityState(ctx, state); err != nil {
		t.Fatalf("upsert state: %v", err)
	}

	storedState, ok, err := repo.GetDeviceCapabilityState(ctx, state.DeviceID, state.CapabilityID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if !ok {
		t.Fatalf("expected state to exist")
	}
	if storedState.State != "on" || !storedState.Enabled {
		t.Fatalf("unexpected state: %+v", storedState)
	}

	byDevice, err := repo.ListDeviceCapabilityStates(ctx, state.DeviceID)
	if err != nil {
		t.Fatalf("list states by device: %v", err)
	}
	if len(byDevice) != 1 {
		t.Fatalf("expected 1 state by device, got %d", len(byDevice))
	}

	byCapability, err := repo.ListCapabilityDeviceStates(ctx, state.CapabilityID)
	if err != nil {
		t.Fatalf("list states by capability: %v", err)
	}
	if len(byCapability) != 1 {
		t.Fatalf("expected 1 state by capability, got %d", len(byCapability))
	}
}
