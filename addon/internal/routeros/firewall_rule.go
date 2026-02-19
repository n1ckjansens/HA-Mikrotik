package routeros

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type firewallRuleMatch struct {
	ID       string
	Comment  string
	Disabled bool
}

// SetFirewallRuleDisabled toggles disabled flag for one firewall rule by RouterOS object ID.
func (c *Client) SetFirewallRuleDisabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
	disabled bool,
) error {
	tablePath, err := firewallRuleTablePath(table)
	if err != nil {
		return err
	}

	ruleID = normalizeRouterObjectID(ruleID)
	if ruleID == "" {
		return fmt.Errorf("firewall rule id is required")
	}

	httpClient := c.httpClientForConfig(cfg)
	base := strings.TrimSuffix(cfg.BaseURL(), "/")
	return c.setFirewallRuleDisabledByID(ctx, httpClient, cfg, base, tablePath, ruleID, disabled)
}

// GetFirewallRuleEnabled returns true when selected firewall rule is enabled.
func (c *Client) GetFirewallRuleEnabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
) (bool, error) {
	tablePath, err := firewallRuleTablePath(table)
	if err != nil {
		return false, err
	}

	ruleID = normalizeRouterObjectID(ruleID)
	if ruleID == "" {
		return false, fmt.Errorf("firewall rule id is required")
	}

	httpClient := c.httpClientForConfig(cfg)
	base := strings.TrimSuffix(cfg.BaseURL(), "/")
	rules, err := c.findFirewallRules(ctx, httpClient, cfg, base, tablePath, nil)
	if err != nil {
		return false, err
	}
	for _, rule := range rules {
		if rule.ID == ruleID {
			return !rule.Disabled, nil
		}
	}
	return false, fmt.Errorf("firewall rule %q not found", ruleID)
}

// SetFirewallRulesDisabledByComment toggles disabled flag for all rules with exact comment match.
func (c *Client) SetFirewallRulesDisabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
	disabled bool,
) error {
	tablePath, err := firewallRuleTablePath(table)
	if err != nil {
		return err
	}

	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fmt.Errorf("firewall rule comment is required")
	}

	httpClient := c.httpClientForConfig(cfg)
	base := strings.TrimSuffix(cfg.BaseURL(), "/")
	rules, err := c.findFirewallRules(ctx, httpClient, cfg, base, tablePath, []string{"comment=" + comment})
	if err != nil {
		return err
	}

	matched := 0
	for _, rule := range rules {
		if strings.TrimSpace(rule.Comment) != comment {
			continue
		}
		matched++
		if err := c.setFirewallRuleDisabledByID(ctx, httpClient, cfg, base, tablePath, rule.ID, disabled); err != nil {
			var statusErr *HTTPStatusError
			if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
				continue
			}
			return err
		}
	}
	if matched == 0 {
		return fmt.Errorf("firewall rules with comment %q not found", comment)
	}
	return nil
}

// GetFirewallRulesEnabledByComment returns true when all matched rules are enabled.
func (c *Client) GetFirewallRulesEnabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
) (bool, error) {
	tablePath, err := firewallRuleTablePath(table)
	if err != nil {
		return false, err
	}

	comment = strings.TrimSpace(comment)
	if comment == "" {
		return false, fmt.Errorf("firewall rule comment is required")
	}

	httpClient := c.httpClientForConfig(cfg)
	base := strings.TrimSuffix(cfg.BaseURL(), "/")
	rules, err := c.findFirewallRules(ctx, httpClient, cfg, base, tablePath, []string{"comment=" + comment})
	if err != nil {
		return false, err
	}

	matched := 0
	for _, rule := range rules {
		if strings.TrimSpace(rule.Comment) != comment {
			continue
		}
		matched++
		if rule.Disabled {
			return false, nil
		}
	}
	if matched == 0 {
		return false, fmt.Errorf("firewall rules with comment %q not found", comment)
	}
	return true, nil
}

func (c *Client) findFirewallRules(
	ctx context.Context,
	httpClient *http.Client,
	cfg model.RouterConfig,
	base string,
	tablePath string,
	query []string,
) ([]firewallRuleMatch, error) {
	endpoint := base + tablePath + "/print"
	payload := map[string]any{
		".proplist": []string{".id", "comment", "disabled"},
	}
	if len(query) > 0 {
		payload[".query"] = query
	}

	c.logRouterRequest("firewall-rule.lookup", http.MethodPost, endpoint)
	rows, err := fetchRowsByPost(ctx, httpClient, endpoint, cfg, payload)
	if err != nil {
		return nil, err
	}

	matches := make([]firewallRuleMatch, 0, len(rows))
	for _, row := range rows {
		id := normalizeRouterObjectID(str(row[".id"]))
		if id == "" {
			continue
		}
		matches = append(matches, firewallRuleMatch{
			ID:       id,
			Comment:  strings.TrimSpace(str(row["comment"])),
			Disabled: boolValue(row["disabled"]),
		})
	}
	return matches, nil
}

func (c *Client) setFirewallRuleDisabledByID(
	ctx context.Context,
	httpClient *http.Client,
	cfg model.RouterConfig,
	base string,
	tablePath string,
	ruleID string,
	disabled bool,
) error {
	endpoint := base + tablePath + "/" + ruleID
	c.logRouterRequest("firewall-rule.set-disabled", http.MethodPatch, endpoint)

	// RouterOS accepts both bool and string payload types across versions.
	err := executeRouterRequestWithRetry(
		ctx,
		httpClient,
		http.MethodPatch,
		endpoint,
		cfg,
		map[string]any{"disabled": disabled},
	)
	if err == nil {
		return nil
	}

	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusBadRequest {
		return err
	}

	return executeRouterRequestWithRetry(
		ctx,
		httpClient,
		http.MethodPatch,
		endpoint,
		cfg,
		map[string]string{"disabled": strconv.FormatBool(disabled)},
	)
}

func firewallRuleTablePath(table string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(table)) {
	case "filter":
		return "/ip/firewall/filter", nil
	case "nat":
		return "/ip/firewall/nat", nil
	case "mangle":
		return "/ip/firewall/mangle", nil
	case "raw":
		return "/ip/firewall/raw", nil
	default:
		return "", fmt.Errorf("unsupported firewall table %q", table)
	}
}
