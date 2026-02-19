package actions

import (
	"context"
	"testing"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type fakeFirewallRuleClient struct {
	setByIDCalls      int
	setByCommentCalls int
	lastTable         string
	lastRuleID        string
	lastComment       string
	lastDisabled      bool
}

func (f *fakeFirewallRuleClient) AddAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	return nil
}

func (f *fakeFirewallRuleClient) RemoveAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	return nil
}

func (f *fakeFirewallRuleClient) SetFirewallRuleDisabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
	disabled bool,
) error {
	f.setByIDCalls++
	f.lastTable = table
	f.lastRuleID = ruleID
	f.lastDisabled = disabled
	return nil
}

func (f *fakeFirewallRuleClient) SetFirewallRulesDisabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
	disabled bool,
) error {
	f.setByCommentCalls++
	f.lastTable = table
	f.lastComment = comment
	f.lastDisabled = disabled
	return nil
}

func TestFirewallRuleToggleActionExecuteDisableByID(t *testing.T) {
	action := NewFirewallRuleToggleAction()
	client := &fakeFirewallRuleClient{}

	err := action.Execute(context.Background(), automationdomain.ActionExecutionContext{
		Target:       automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice},
		RouterClient: client,
		RouterConfig: model.RouterConfig{Host: "router.local"},
	}, map[string]any{
		"table":    "filter",
		"mode":     "disable",
		"match_by": "id",
		"rule_id":  "*A",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if client.setByIDCalls != 1 || client.setByCommentCalls != 0 {
		t.Fatalf("unexpected calls: by_id=%d by_comment=%d", client.setByIDCalls, client.setByCommentCalls)
	}
	if client.lastTable != "filter" || client.lastRuleID != "*A" || !client.lastDisabled {
		t.Fatalf("unexpected args: table=%q id=%q disabled=%v", client.lastTable, client.lastRuleID, client.lastDisabled)
	}
}

func TestFirewallRuleToggleActionExecuteEnableByComment(t *testing.T) {
	action := NewFirewallRuleToggleAction()
	client := &fakeFirewallRuleClient{}

	err := action.Execute(context.Background(), automationdomain.ActionExecutionContext{
		Target:       automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal},
		RouterClient: client,
		RouterConfig: model.RouterConfig{Host: "router.local"},
	}, map[string]any{
		"table":    "filter",
		"mode":     "enable",
		"match_by": "comment",
		"comment":  "VPN_PROFILE",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if client.setByCommentCalls != 1 || client.setByIDCalls != 0 {
		t.Fatalf("unexpected calls: by_id=%d by_comment=%d", client.setByIDCalls, client.setByCommentCalls)
	}
	if client.lastTable != "filter" || client.lastComment != "VPN_PROFILE" || client.lastDisabled {
		t.Fatalf("unexpected args: table=%q comment=%q disabled=%v", client.lastTable, client.lastComment, client.lastDisabled)
	}
}

func TestFirewallRuleToggleActionValidateRejectsInvalidMode(t *testing.T) {
	action := NewFirewallRuleToggleAction()
	err := action.Validate(automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice}, map[string]any{
		"table":    "filter",
		"mode":     "toggle",
		"match_by": "id",
		"rule_id":  "*A",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestFirewallRuleToggleActionValidateRejectsDevicePlaceholderForGlobalScope(t *testing.T) {
	action := NewFirewallRuleToggleAction()
	err := action.Validate(automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal}, map[string]any{
		"table":    "filter",
		"mode":     "disable",
		"match_by": "comment",
		"comment":  "{{device.name}}",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
