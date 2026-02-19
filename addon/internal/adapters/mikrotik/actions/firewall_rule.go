package actions

import (
	"context"
	"fmt"
	"strings"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

const (
	// ActionIDFirewallRuleToggle toggles disabled flag for firewall rule.
	ActionIDFirewallRuleToggle = "mikrotik.firewall.rule.set_enabled"
)

// FirewallRuleToggleAction enables/disables RouterOS firewall rules.
type FirewallRuleToggleAction struct{}

// NewFirewallRuleToggleAction creates MikroTik firewall rule action.
func NewFirewallRuleToggleAction() *FirewallRuleToggleAction {
	return &FirewallRuleToggleAction{}
}

// ID returns unique action identifier.
func (a *FirewallRuleToggleAction) ID() string {
	return ActionIDFirewallRuleToggle
}

// Metadata returns action descriptor for UI.
func (a *FirewallRuleToggleAction) Metadata() automationdomain.ActionMetadata {
	return automationdomain.ActionMetadata{
		ID:          ActionIDFirewallRuleToggle,
		Label:       "MikroTik: Firewall rule toggle",
		Description: "Enable or disable firewall rule in filter/nat/mangle/raw tables",
		ParamSchema: []automationdomain.ParamField{
			{
				Key:         "table",
				Label:       "Firewall table",
				Kind:        automationdomain.ParamEnum,
				Required:    true,
				Options:     []string{"filter", "nat", "mangle", "raw"},
				Description: "RouterOS firewall table where the rule exists",
			},
			{
				Key:         "mode",
				Label:       "Mode",
				Kind:        automationdomain.ParamEnum,
				Required:    true,
				Options:     []string{"enable", "disable"},
				Description: "Enable or disable selected rule(s)",
			},
			{
				Key:         "match_by",
				Label:       "Match by",
				Kind:        automationdomain.ParamEnum,
				Required:    true,
				Options:     []string{"id", "comment"},
				Description: "Choose whether to target one rule id or all rules by comment",
			},
			{
				Key:       "rule_id",
				Label:     "Rule id",
				Kind:      automationdomain.ParamString,
				Required:  true,
				VisibleIf: &automationdomain.VisibleIfCondition{Key: "match_by", Equals: "id"},
			},
			{
				Key:       "comment",
				Label:     "Rule comment",
				Kind:      automationdomain.ParamString,
				Required:  true,
				VisibleIf: &automationdomain.VisibleIfCondition{Key: "match_by", Equals: "comment"},
			},
		},
	}
}

// Validate validates action params against metadata schema.
func (a *FirewallRuleToggleAction) Validate(
	target automationdomain.AutomationTarget,
	params map[string]any,
) error {
	table, err := stringParam(params, "table")
	if err != nil {
		return err
	}
	switch table {
	case "filter", "nat", "mangle", "raw":
	default:
		return fmt.Errorf("unsupported table %q", table)
	}

	mode, err := stringParam(params, "mode")
	if err != nil {
		return err
	}
	if mode != "enable" && mode != "disable" {
		return fmt.Errorf("unsupported mode %q", mode)
	}

	matchBy, err := stringParam(params, "match_by")
	if err != nil {
		return err
	}

	scope := automationdomain.NormalizeCapabilityScope(target.Scope)
	switch matchBy {
	case "id":
		ruleID, err := stringParam(params, "rule_id")
		if err != nil {
			return err
		}
		if scope == automationdomain.ScopeGlobal && containsDevicePlaceholder(ruleID) {
			return fmt.Errorf("global scope does not support device placeholders")
		}
	case "comment":
		comment, err := stringParam(params, "comment")
		if err != nil {
			return err
		}
		if scope == automationdomain.ScopeGlobal && containsDevicePlaceholder(comment) {
			return fmt.Errorf("global scope does not support device placeholders")
		}
	default:
		return fmt.Errorf("unsupported match_by %q", matchBy)
	}

	return nil
}

// Execute applies firewall rule toggle via RouterOS API.
func (a *FirewallRuleToggleAction) Execute(
	ctx context.Context,
	execCtx automationdomain.ActionExecutionContext,
	params map[string]any,
) error {
	if err := a.Validate(execCtx.Target, params); err != nil {
		return err
	}
	if execCtx.RouterClient == nil {
		return fmt.Errorf("router client is not configured")
	}

	table, _ := stringParam(params, "table")
	mode, _ := stringParam(params, "mode")
	matchBy, _ := stringParam(params, "match_by")
	disabled := strings.EqualFold(mode, "disable")

	switch matchBy {
	case "id":
		ruleID, _ := stringParam(params, "rule_id")
		return execCtx.RouterClient.SetFirewallRuleDisabled(
			ctx,
			execCtx.RouterConfig,
			table,
			ruleID,
			disabled,
		)
	case "comment":
		comment, _ := stringParam(params, "comment")
		return execCtx.RouterClient.SetFirewallRulesDisabledByComment(
			ctx,
			execCtx.RouterConfig,
			table,
			comment,
			disabled,
		)
	default:
		return fmt.Errorf("unsupported match_by %q", matchBy)
	}
}
