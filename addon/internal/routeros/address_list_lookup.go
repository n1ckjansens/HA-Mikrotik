package routeros

import (
	"context"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// AddressListContains checks whether list contains given address.
func (c *Client) AddressListContains(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) (bool, error) {
	list = strings.TrimSpace(list)
	address = strings.TrimSpace(address)
	if list == "" || address == "" {
		return false, nil
	}
	httpClient := c.httpClientForConfig(cfg)
	base := strings.TrimSuffix(cfg.BaseURL(), "/")
	ids, err := c.findAddressListEntryIDs(ctx, httpClient, cfg, base, list, address)
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}
