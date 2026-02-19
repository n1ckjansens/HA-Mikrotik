package statesources

import (
	"context"
	"testing"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type fakeStateClient struct {
	contains bool
	calls    int
	list     string
	address  string
}

func (f *fakeStateClient) AddressListContains(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) (bool, error) {
	f.calls++
	f.list = list
	f.address = address
	return f.contains, nil
}

func (f *fakeStateClient) GetFirewallRuleEnabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
) (bool, error) {
	return false, nil
}

func (f *fakeStateClient) GetFirewallRulesEnabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
) (bool, error) {
	return false, nil
}

func TestAddressListMembershipSourceRead(t *testing.T) {
	source := NewAddressListMembershipSource()
	ip := "192.168.88.15"
	client := &fakeStateClient{contains: true}
	device := model.DeviceView{MAC: "AA:BB:CC:DD:EE:02", LastIP: &ip}

	result, err := source.Read(context.Background(), automationdomain.StateSourceContext{
		Target:       automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice, Device: &device},
		RouterClient: client,
		RouterConfig: model.RouterConfig{Host: "router.local"},
	}, map[string]any{
		"list":   "VPN_CLIENTS",
		"target": "device.ip",
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	value, ok := result.(bool)
	if !ok {
		t.Fatalf("expected bool result, got %T", result)
	}
	if !value {
		t.Fatalf("expected true membership")
	}
	if client.calls != 1 || client.list != "VPN_CLIENTS" || client.address != ip {
		t.Fatalf("unexpected client calls: calls=%d list=%q address=%q", client.calls, client.list, client.address)
	}
}

func TestAddressListMembershipSourceValidate(t *testing.T) {
	source := NewAddressListMembershipSource()
	err := source.Validate(automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice}, map[string]any{
		"list":   "VPN_CLIENTS",
		"target": "literal_ip",
	})
	if err == nil {
		t.Fatalf("expected validation error for missing literal_ip")
	}
}

func TestAddressListMembershipSourceValidateRejectsDeviceTargetForGlobalScope(t *testing.T) {
	source := NewAddressListMembershipSource()
	err := source.Validate(automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal}, map[string]any{
		"list":   "VPN_CLIENTS",
		"target": "device.mac",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
