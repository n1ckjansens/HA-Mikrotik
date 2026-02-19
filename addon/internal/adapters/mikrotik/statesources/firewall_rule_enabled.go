package statesources

import (
	"context"
	"fmt"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

const (
	// StateSourceIDFirewallRuleEnabled reads current enabled state of firewall rules.
	StateSourceIDFirewallRuleEnabled = "mikrotik.firewall.rule.enabled"
)

// FirewallRuleEnabledSource reads enabled/disabled state for firewall rules.
type FirewallRuleEnabledSource struct{}

// NewFirewallRuleEnabledSource creates state-source implementation.
func NewFirewallRuleEnabledSource() *FirewallRuleEnabledSource {
	return &FirewallRuleEnabledSource{}
}

// ID returns unique state-source identifier.
func (s *FirewallRuleEnabledSource) ID() string {
	return StateSourceIDFirewallRuleEnabled
}

// Metadata returns state-source descriptor for UI.
func (s *FirewallRuleEnabledSource) Metadata() automationdomain.StateSourceMetadata {
	return automationdomain.StateSourceMetadata{
		ID:          StateSourceIDFirewallRuleEnabled,
		Label:       "MikroTik: Firewall rule enabled",
		Description: "Checks whether firewall rule is currently enabled in filter/nat/mangle/raw tables",
		OutputType:  "boolean",
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
				Key:         "match_by",
				Label:       "Match by",
				Kind:        automationdomain.ParamEnum,
				Required:    true,
				Options:     []string{"id", "comment"},
				Description: "Choose whether to read one rule id or all rules by comment",
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

// Validate validates state-source params against schema.
func (s *FirewallRuleEnabledSource) Validate(
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

// Read returns true when rule/rules are enabled.
func (s *FirewallRuleEnabledSource) Read(
	ctx context.Context,
	sourceCtx automationdomain.StateSourceContext,
	params map[string]any,
) (any, error) {
	if err := s.Validate(sourceCtx.Target, params); err != nil {
		return nil, err
	}
	if sourceCtx.RouterClient == nil {
		return nil, fmt.Errorf("router client is not configured")
	}

	table, _ := stringParam(params, "table")
	matchBy, _ := stringParam(params, "match_by")

	switch matchBy {
	case "id":
		ruleID, _ := stringParam(params, "rule_id")
		return sourceCtx.RouterClient.GetFirewallRuleEnabled(
			ctx,
			sourceCtx.RouterConfig,
			table,
			ruleID,
		)
	case "comment":
		comment, _ := stringParam(params, "comment")
		return sourceCtx.RouterClient.GetFirewallRulesEnabledByComment(
			ctx,
			sourceCtx.RouterConfig,
			table,
			comment,
		)
	default:
		return nil, fmt.Errorf("unsupported match_by %q", matchBy)
	}
}
