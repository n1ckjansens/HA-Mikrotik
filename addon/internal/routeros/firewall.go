package routeros

import (
	"context"
	"fmt"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// FirewallRule represents a simplified RouterOS firewall rule.
type FirewallRule struct {
	ID       string
	Table    string
	Chain    string
	Action   string
	Comment  string
	Disabled bool
}

// ListFirewallRules returns rules from filter/nat/mangle/raw tables.
func (c *Client) ListFirewallRules(ctx context.Context) ([]FirewallRule, error) {
	tables := []string{"filter", "nat", "mangle", "raw"}
	all := make([]FirewallRule, 0)
	for _, table := range tables {
		rules, err := c.listFirewallRulesByTable(ctx, table)
		if err != nil {
			return nil, err
		}
		all = append(all, rules...)
	}
	return all, nil
}

// ListFirewallRules returns rules for selected pooled client.
func (m *Manager) ListFirewallRules(ctx context.Context, cfg model.RouterConfig) ([]FirewallRule, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return client.ListFirewallRules(ctx)
}

// EnableRule enables a firewall rule by .id, auto-detecting table.
func (c *Client) EnableRule(ctx context.Context, id string) error {
	return c.setRuleDisabledByID(ctx, id, false)
}

// EnableRule enables rule by id for selected pooled client.
func (m *Manager) EnableRule(ctx context.Context, cfg model.RouterConfig, id string) error {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return err
	}
	return client.EnableRule(ctx, id)
}

// DisableRule disables a firewall rule by .id, auto-detecting table.
func (c *Client) DisableRule(ctx context.Context, id string) error {
	return c.setRuleDisabledByID(ctx, id, true)
}

// DisableRule disables rule by id for selected pooled client.
func (m *Manager) DisableRule(ctx context.Context, cfg model.RouterConfig, id string) error {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return err
	}
	return client.DisableRule(ctx, id)
}

func (c *Client) setRuleDisabledByID(ctx context.Context, id string, disabled bool) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return &ValidationError{Field: "id", Reason: "is required"}
	}

	matches, err := c.findFirewallRulesByID(ctx, id)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return &RuleNotFoundError{ID: id}
	}
	if len(matches) > 1 {
		return fmt.Errorf("firewall rule id %q is ambiguous across tables", id)
	}

	rule := matches[0]
	if rule.Disabled == disabled {
		return nil
	}
	return c.setFirewallRuleDisabledInTable(ctx, rule.Table, rule.ID, disabled)
}

func (c *Client) findFirewallRulesByID(ctx context.Context, id string) ([]FirewallRule, error) {
	tables := []string{"filter", "nat", "mangle", "raw"}
	matches := make([]FirewallRule, 0, 1)
	for _, table := range tables {
		rules, err := c.listFirewallRulesByTable(ctx, table)
		if err != nil {
			return nil, err
		}
		for _, rule := range rules {
			if rule.ID == id {
				matches = append(matches, rule)
			}
		}
	}
	return matches, nil
}

func (c *Client) listFirewallRulesByTable(ctx context.Context, table string) ([]FirewallRule, error) {
	table = strings.ToLower(strings.TrimSpace(table))
	if table == "" {
		return nil, &ValidationError{Field: "table", Reason: "is required"}
	}

	rows, err := c.RunCommand(ctx, "/ip/firewall/"+table+"/print", map[string]string{
		".proplist": ".id,chain,action,comment,disabled",
	})
	if err != nil {
		return nil, fmt.Errorf("list firewall table %q: %w", table, err)
	}

	rules := make([]FirewallRule, 0, len(rows))
	for _, row := range rows {
		id := strings.TrimSpace(row[".id"])
		if id == "" {
			continue
		}
		rules = append(rules, FirewallRule{
			ID:       id,
			Table:    table,
			Chain:    strings.TrimSpace(row["chain"]),
			Action:   strings.TrimSpace(row["action"]),
			Comment:  strings.TrimSpace(row["comment"]),
			Disabled: boolFromWord(row["disabled"]),
		})
	}
	return rules, nil
}

func (c *Client) setFirewallRuleDisabledInTable(ctx context.Context, table string, ruleID string, disabled bool) error {
	_, err := c.RunCommand(ctx, "/ip/firewall/"+table+"/set", map[string]string{
		".id":      strings.TrimSpace(ruleID),
		"disabled": boolToWord(disabled),
	})
	if err != nil {
		return fmt.Errorf("set firewall rule %s in %s: %w", ruleID, table, err)
	}
	return nil
}

func (c *Client) rulesByComment(ctx context.Context, table string, comment string) ([]FirewallRule, error) {
	rules, err := c.listFirewallRulesByTable(ctx, table)
	if err != nil {
		return nil, err
	}

	comment = strings.TrimSpace(comment)
	matches := make([]FirewallRule, 0)
	for _, rule := range rules {
		if strings.TrimSpace(rule.Comment) == comment {
			matches = append(matches, rule)
		}
	}
	return matches, nil
}

// SetFirewallRuleDisabled keeps compatibility with current automation actions.
func (m *Manager) SetFirewallRuleDisabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
	disabled bool,
) error {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return err
	}
	return client.setFirewallRuleDisabledInTable(ctx, strings.ToLower(strings.TrimSpace(table)), ruleID, disabled)
}

// GetFirewallRuleEnabled keeps compatibility with current state-sources.
func (m *Manager) GetFirewallRuleEnabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
) (bool, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return false, err
	}
	rules, err := client.listFirewallRulesByTable(ctx, table)
	if err != nil {
		return false, err
	}
	ruleID = strings.TrimSpace(ruleID)
	for _, rule := range rules {
		if rule.ID == ruleID {
			return !rule.Disabled, nil
		}
	}
	return false, &RuleNotFoundError{ID: ruleID}
}

// SetFirewallRulesDisabledByComment keeps compatibility with current automation actions.
func (m *Manager) SetFirewallRulesDisabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
	disabled bool,
) error {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return err
	}

	rules, err := client.rulesByComment(ctx, table, comment)
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return fmt.Errorf("firewall rules with comment %q not found", comment)
	}

	for _, rule := range rules {
		if err := client.setFirewallRuleDisabledInTable(ctx, rule.Table, rule.ID, disabled); err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}
	}
	return nil
}

// GetFirewallRulesEnabledByComment keeps compatibility with current state-sources.
func (m *Manager) GetFirewallRulesEnabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
) (bool, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return false, err
	}
	rules, err := client.rulesByComment(ctx, table, comment)
	if err != nil {
		return false, err
	}
	if len(rules) == 0 {
		return false, fmt.Errorf("firewall rules with comment %q not found", comment)
	}
	for _, rule := range rules {
		if rule.Disabled {
			return false, nil
		}
	}
	return true, nil
}
