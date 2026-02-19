package service

import (
	"sort"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
)

func normalizeDeviceID(raw string) string {
	value := strings.TrimSpace(strings.ToUpper(raw))
	value = strings.ReplaceAll(value, "%3A", ":")
	value = strings.ReplaceAll(value, "%3a", ":")
	value = strings.ReplaceAll(value, "-", ":")
	return value
}

func sortCapabilityUIModels(items []domain.CapabilityUIModel) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Label != items[j].Label {
			return strings.ToLower(items[i].Label) < strings.ToLower(items[j].Label)
		}
		return items[i].ID < items[j].ID
	})
}

func sortCapabilityAssignments(items []domain.CapabilityDeviceAssignment) {
	sort.SliceStable(items, func(i, j int) bool {
		nameI := strings.ToLower(items[i].DeviceName)
		nameJ := strings.ToLower(items[j].DeviceName)
		if nameI != nameJ {
			return nameI < nameJ
		}
		return items[i].DeviceID < items[j].DeviceID
	})
}

func isUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint")
}
