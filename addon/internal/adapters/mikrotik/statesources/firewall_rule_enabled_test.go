package statesources

import (
	"context"
	"testing"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type fakeFirewallRuleStateClient struct {
	enabledByID      bool
	enabledByComment bool
	idCalls          int
	commentCalls     int
	lastTable        string
	lastRuleID       string
	lastComment      string
}

func (f *fakeFirewallRuleStateClient) AddressListContains(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) (bool, error) {
	return false, nil
}

func (f *fakeFirewallRuleStateClient) GetFirewallRuleEnabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
) (bool, error) {
	f.idCalls++
	f.lastTable = table
	f.lastRuleID = ruleID
	return f.enabledByID, nil
}

func (f *fakeFirewallRuleStateClient) GetFirewallRulesEnabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
) (bool, error) {
	f.commentCalls++
	f.lastTable = table
	f.lastComment = comment
	return f.enabledByComment, nil
}

func TestFirewallRuleEnabledSourceReadByID(t *testing.T) {
	source := NewFirewallRuleEnabledSource()
	client := &fakeFirewallRuleStateClient{enabledByID: true}

	result, err := source.Read(context.Background(), automationdomain.StateSourceContext{
		Target:       automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal},
		RouterClient: client,
		RouterConfig: model.RouterConfig{Host: "router.local"},
	}, map[string]any{
		"table":    "filter",
		"match_by": "id",
		"rule_id":  "*A",
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	value, ok := result.(bool)
	if !ok || !value {
		t.Fatalf("expected true bool result, got %T (%v)", result, result)
	}
	if client.idCalls != 1 || client.commentCalls != 0 {
		t.Fatalf("unexpected calls: id=%d comment=%d", client.idCalls, client.commentCalls)
	}
	if client.lastTable != "filter" || client.lastRuleID != "*A" {
		t.Fatalf("unexpected args: table=%q rule_id=%q", client.lastTable, client.lastRuleID)
	}
}

func TestFirewallRuleEnabledSourceReadByComment(t *testing.T) {
	source := NewFirewallRuleEnabledSource()
	client := &fakeFirewallRuleStateClient{enabledByComment: false}

	result, err := source.Read(context.Background(), automationdomain.StateSourceContext{
		Target:       automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice},
		RouterClient: client,
		RouterConfig: model.RouterConfig{Host: "router.local"},
	}, map[string]any{
		"table":    "filter",
		"match_by": "comment",
		"comment":  "VPN_PROFILE",
	})
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	value, ok := result.(bool)
	if !ok || value {
		t.Fatalf("expected false bool result, got %T (%v)", result, result)
	}
	if client.commentCalls != 1 || client.idCalls != 0 {
		t.Fatalf("unexpected calls: id=%d comment=%d", client.idCalls, client.commentCalls)
	}
	if client.lastTable != "filter" || client.lastComment != "VPN_PROFILE" {
		t.Fatalf("unexpected args: table=%q comment=%q", client.lastTable, client.lastComment)
	}
}

func TestFirewallRuleEnabledSourceValidateRejectsGlobalDevicePlaceholder(t *testing.T) {
	source := NewFirewallRuleEnabledSource()
	err := source.Validate(automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal}, map[string]any{
		"table":    "filter",
		"match_by": "comment",
		"comment":  "{{device.name}}",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
