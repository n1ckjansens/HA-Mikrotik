package mikrotik

import (
	"context"
	"log/slog"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	"github.com/micro-ha/mikrotik-presence/addon/internal/routeros"
)

// RestClient is DI-friendly wrapper around RouterOS REST client.
type RestClient struct {
	client *routeros.Client
}

// NewRestClient creates adapter wrapper for RouterOS REST calls.
func NewRestClient(logger *slog.Logger) *RestClient {
	inner := routeros.NewClient()
	if logger != nil {
		inner = inner.WithLogger(logger.With("component", "routeros"))
	}
	return &RestClient{client: inner}
}

// FetchSnapshot returns current RouterOS signal snapshot.
func (c *RestClient) FetchSnapshot(ctx context.Context, cfg model.RouterConfig) (*routeros.Snapshot, error) {
	return c.client.FetchSnapshot(ctx, cfg)
}

// AddAddressListEntry adds one entry to firewall address-list.
func (c *RestClient) AddAddressListEntry(ctx context.Context, cfg model.RouterConfig, list, address string) error {
	return c.client.AddAddressListEntry(ctx, cfg, list, address)
}

// RemoveAddressListEntry removes all matching entries from address-list.
func (c *RestClient) RemoveAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	return c.client.RemoveAddressListEntry(ctx, cfg, list, address)
}

// AddressListContains checks whether address is already listed.
func (c *RestClient) AddressListContains(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) (bool, error) {
	return c.client.AddressListContains(ctx, cfg, list, address)
}

// GetFirewallRuleEnabled returns true when rule is enabled.
func (c *RestClient) GetFirewallRuleEnabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
) (bool, error) {
	return c.client.GetFirewallRuleEnabled(ctx, cfg, table, ruleID)
}

// GetFirewallRulesEnabledByComment returns true when all matched rules are enabled.
func (c *RestClient) GetFirewallRulesEnabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
) (bool, error) {
	return c.client.GetFirewallRulesEnabledByComment(ctx, cfg, table, comment)
}

// SetFirewallRuleDisabled toggles one firewall rule by id.
func (c *RestClient) SetFirewallRuleDisabled(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	ruleID string,
	disabled bool,
) error {
	return c.client.SetFirewallRuleDisabled(ctx, cfg, table, ruleID, disabled)
}

// SetFirewallRulesDisabledByComment toggles all rules with exact comment.
func (c *RestClient) SetFirewallRulesDisabledByComment(
	ctx context.Context,
	cfg model.RouterConfig,
	table string,
	comment string,
	disabled bool,
) error {
	return c.client.SetFirewallRulesDisabledByComment(ctx, cfg, table, comment, disabled)
}
