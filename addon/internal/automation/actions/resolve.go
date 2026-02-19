package actions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

var placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9._-]+)\s*\}\}`)

func ResolveParams(params map[string]any, device model.DeviceView) (map[string]any, error) {
	resolved := make(map[string]any, len(params))
	placeholders := placeholdersForDevice(device)
	for key, value := range params {
		next, err := resolveValue(value, placeholders)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", key, err)
		}
		resolved[key] = next
	}
	return resolved, nil
}

func resolveValue(value any, placeholders map[string]string) (any, error) {
	switch typed := value.(type) {
	case string:
		return resolveString(typed, placeholders)
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, child := range typed {
			next, err := resolveValue(child, placeholders)
			if err != nil {
				return nil, fmt.Errorf("resolve nested key %q: %w", key, err)
			}
			cloned[key] = next
		}
		return cloned, nil
	case []any:
		cloned := make([]any, 0, len(typed))
		for index, child := range typed {
			next, err := resolveValue(child, placeholders)
			if err != nil {
				return nil, fmt.Errorf("resolve array item %d: %w", index, err)
			}
			cloned = append(cloned, next)
		}
		return cloned, nil
	default:
		return value, nil
	}
}

func resolveString(input string, placeholders map[string]string) (string, error) {
	matches := placeholderPattern.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return input, nil
	}

	resolved := input
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		key := strings.TrimSpace(match[1])
		value, ok := placeholders[key]
		if !ok {
			return "", fmt.Errorf("unknown placeholder %q", key)
		}
		if strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("placeholder %q resolved to empty value", key)
		}
		resolved = strings.ReplaceAll(resolved, match[0], value)
	}
	return resolved, nil
}

func placeholdersForDevice(device model.DeviceView) map[string]string {
	placeholders := map[string]string{
		"device.mac":  device.MAC,
		"device.name": device.Name,
	}
	if device.LastIP != nil {
		placeholders["device.ip"] = *device.LastIP
	}
	if device.HostName != nil {
		placeholders["device.host_name"] = *device.HostName
	}
	return placeholders
}
