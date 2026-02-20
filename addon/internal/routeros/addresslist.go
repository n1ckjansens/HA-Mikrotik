package routeros

import (
	"context"
	"fmt"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// AddAddressToList adds an entry into firewall address-list (idempotent).
func (c *Client) AddAddressToList(ctx context.Context, list string, address string) error {
	list = strings.TrimSpace(list)
	address = strings.TrimSpace(address)
	if list == "" {
		return &ValidationError{Field: "list", Reason: "is required"}
	}
	if address == "" {
		return &ValidationError{Field: "address", Reason: "is required"}
	}

	c.addressList.Lock()
	defer c.addressList.Unlock()

	exists, err := c.AddressExists(ctx, list, address)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = c.RunCommand(ctx, "/ip/firewall/address-list/add", map[string]string{
		"list":    list,
		"address": address,
	})
	if err != nil {
		if isAlreadyExistsError(err) {
			return nil
		}
		return fmt.Errorf("add address to list %q: %w", list, err)
	}
	return nil
}

// AddAddressToList executes add operation on pooled client selected by cfg.
func (m *Manager) AddAddressToList(ctx context.Context, cfg model.RouterConfig, list string, address string) error {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return err
	}
	return client.AddAddressToList(ctx, list, address)
}

// RemoveAddressFromList removes all matching entries from firewall address-list (idempotent).
func (c *Client) RemoveAddressFromList(ctx context.Context, list string, address string) error {
	list = strings.TrimSpace(list)
	address = strings.TrimSpace(address)
	if list == "" {
		return &ValidationError{Field: "list", Reason: "is required"}
	}
	if address == "" {
		return &ValidationError{Field: "address", Reason: "is required"}
	}

	c.addressList.Lock()
	defer c.addressList.Unlock()

	ids, err := c.findAddressIDs(ctx, list, address)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		_, err := c.RunCommand(ctx, "/ip/firewall/address-list/remove", map[string]string{
			".id": id,
		})
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return fmt.Errorf("remove address-list entry %s: %w", id, err)
		}
	}
	return nil
}

// RemoveAddressFromList executes remove operation on pooled client selected by cfg.
func (m *Manager) RemoveAddressFromList(ctx context.Context, cfg model.RouterConfig, list string, address string) error {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return err
	}
	return client.RemoveAddressFromList(ctx, list, address)
}

// AddressExists returns true when firewall address-list contains value.
func (c *Client) AddressExists(ctx context.Context, list string, address string) (bool, error) {
	list = strings.TrimSpace(list)
	address = strings.TrimSpace(address)
	if list == "" || address == "" {
		return false, nil
	}

	ids, err := c.findAddressIDs(ctx, list, address)
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

// AddressExists checks address-list on pooled client selected by cfg.
func (m *Manager) AddressExists(ctx context.Context, cfg model.RouterConfig, list string, address string) (bool, error) {
	client, err := m.getClient(ctx, cfg)
	if err != nil {
		return false, err
	}
	return client.AddressExists(ctx, list, address)
}

func (c *Client) findAddressIDs(ctx context.Context, list string, address string) ([]string, error) {
	rows, err := c.RunCommand(ctx, "/ip/firewall/address-list/print", map[string]string{
		"?list":     list,
		".proplist": ".id,list,address",
	})
	if err != nil {
		return nil, fmt.Errorf("lookup address-list %q: %w", list, err)
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row["list"]) != list {
			continue
		}
		if !equalAddressTarget(row["address"], address) {
			continue
		}
		id := strings.TrimSpace(row[".id"])
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// AddAddressListEntry keeps compatibility with current automation contracts.
func (m *Manager) AddAddressListEntry(ctx context.Context, cfg model.RouterConfig, list string, address string) error {
	return m.AddAddressToList(ctx, cfg, list, address)
}

// RemoveAddressListEntry keeps compatibility with current automation contracts.
func (m *Manager) RemoveAddressListEntry(ctx context.Context, cfg model.RouterConfig, list string, address string) error {
	return m.RemoveAddressFromList(ctx, cfg, list, address)
}

// AddressListContains keeps compatibility with current automation contracts.
func (m *Manager) AddressListContains(ctx context.Context, cfg model.RouterConfig, list string, address string) (bool, error) {
	return m.AddressExists(ctx, cfg, list, address)
}
