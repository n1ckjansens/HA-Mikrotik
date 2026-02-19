package actions

import (
	"context"
	"testing"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type fakeAddressListClient struct {
	addCalls    int
	removeCalls int
	lastList    string
	lastAddress string
}

func (f *fakeAddressListClient) AddAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	f.addCalls++
	f.lastList = list
	f.lastAddress = address
	return nil
}

func (f *fakeAddressListClient) RemoveAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	f.removeCalls++
	f.lastList = list
	f.lastAddress = address
	return nil
}

func TestAddressListMembershipActionExecuteAdd(t *testing.T) {
	action := NewAddressListMembershipAction()
	ip := "192.168.88.10"
	client := &fakeAddressListClient{}
	device := model.DeviceView{MAC: "AA:BB:CC:DD:EE:01", LastIP: &ip}

	err := action.Execute(context.Background(), automationdomain.ActionExecutionContext{
		Target:       automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice, Device: &device},
		RouterClient: client,
		RouterConfig: model.RouterConfig{Host: "router.local"},
	}, map[string]any{
		"list":   "VPN_CLIENTS",
		"mode":   "add",
		"target": "device.ip",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if client.addCalls != 1 || client.removeCalls != 0 {
		t.Fatalf("unexpected calls: add=%d remove=%d", client.addCalls, client.removeCalls)
	}
	if client.lastList != "VPN_CLIENTS" || client.lastAddress != ip {
		t.Fatalf("unexpected args: list=%q address=%q", client.lastList, client.lastAddress)
	}
}

func TestAddressListMembershipActionValidateRejectsInvalidMode(t *testing.T) {
	action := NewAddressListMembershipAction()
	err := action.Validate(automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice}, map[string]any{
		"list":   "VPN_CLIENTS",
		"mode":   "toggle",
		"target": "device.ip",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestAddressListMembershipActionValidateRejectsDeviceTargetForGlobalScope(t *testing.T) {
	action := NewAddressListMembershipAction()
	err := action.Validate(automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal}, map[string]any{
		"list":   "VPN_CLIENTS",
		"mode":   "add",
		"target": "device.ip",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
