package actions

import (
	"context"
	"fmt"
	"strings"
)

func handleAddressListMembership(ctx context.Context, execCtx ExecutionContext, params map[string]any) error {
	if execCtx.RouterClient == nil {
		return fmt.Errorf("router client is not configured")
	}

	listName, err := stringParam(params, "list")
	if err != nil {
		return err
	}
	mode, err := stringParam(params, "mode")
	if err != nil {
		return err
	}
	target, err := stringParam(params, "target")
	if err != nil {
		return err
	}

	address, err := resolveTargetAddress(target, params, execCtx)
	if err != nil {
		return err
	}

	if execCtx.Logger != nil {
		execCtx.Logger.Info(
			"executing address-list membership action",
			"device_mac", execCtx.Device.MAC,
			"list", listName,
			"mode", mode,
			"target", target,
			"address", address,
		)
	}

	switch mode {
	case "add":
		return execCtx.RouterClient.AddAddressListEntry(ctx, execCtx.RouterConfig, listName, address)
	case "remove":
		return execCtx.RouterClient.RemoveAddressListEntry(ctx, execCtx.RouterConfig, listName, address)
	default:
		return fmt.Errorf("unsupported mode %q", mode)
	}
}

func resolveTargetAddress(target string, params map[string]any, execCtx ExecutionContext) (string, error) {
	switch target {
	case "device.ip":
		if execCtx.Device.LastIP == nil || strings.TrimSpace(*execCtx.Device.LastIP) == "" {
			return "", fmt.Errorf("device IP is empty")
		}
		return strings.TrimSpace(*execCtx.Device.LastIP), nil
	case "device.mac":
		if strings.TrimSpace(execCtx.Device.MAC) == "" {
			return "", fmt.Errorf("device MAC is empty")
		}
		return execCtx.Device.MAC, nil
	case "literal_ip":
		value, err := stringParam(params, "literal_ip")
		if err != nil {
			return "", err
		}
		return value, nil
	default:
		return "", fmt.Errorf("unsupported target %q", target)
	}
}

func stringParam(params map[string]any, key string) (string, error) {
	raw, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing param %q", key)
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("param %q must be string", key)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("param %q is empty", key)
	}
	return value, nil
}
