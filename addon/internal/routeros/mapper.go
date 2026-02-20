package routeros

import (
	"fmt"
	"strconv"
	"strings"

	goros "github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

func mapReplyRows(reply *goros.Reply) []map[string]string {
	if reply == nil || len(reply.Re) == 0 {
		return []map[string]string{}
	}
	rows := make([]map[string]string, 0, len(reply.Re))
	for _, sentence := range reply.Re {
		rows = append(rows, mapSentence(sentence))
	}
	return rows
}

func mapSentence(sentence *proto.Sentence) map[string]string {
	mapped := make(map[string]string)
	if sentence == nil {
		return mapped
	}
	for key, value := range sentence.Map {
		mapped[key] = value
	}
	for _, pair := range sentence.List {
		mapped[pair.Key] = pair.Value
	}
	return mapped
}

func mapParams(params map[string]string) []string {
	if len(params) == 0 {
		return nil
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sortStrings(keys)

	words := make([]string, 0, len(keys))
	for _, key := range keys {
		value := params[key]
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "?"):
			words = append(words, trimmed+"="+value)
		case strings.HasPrefix(trimmed, "="):
			name := strings.TrimPrefix(trimmed, "=")
			words = append(words, "="+name+"="+value)
		default:
			words = append(words, "="+trimmed+"="+value)
		}
	}
	return words
}

func sortStrings(values []string) {
	if len(values) < 2 {
		return
	}
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

func boolFromWord(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "on", "enabled":
		return true
	default:
		return false
	}
}

func boolToWord(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func canonicalMAC(value string) string {
	clean := strings.TrimSpace(strings.ToUpper(value))
	clean = strings.ReplaceAll(clean, "-", ":")
	return clean
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func equalAddressTarget(a string, b string) bool {
	aNorm := normalizeAddressTarget(a)
	bNorm := normalizeAddressTarget(b)
	return aNorm == bNorm
}

func normalizeAddressTarget(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if strings.HasSuffix(trimmed, "/32") {
		trimmed = strings.TrimSuffix(trimmed, "/32")
	}
	return trimmed
}

func parseFloat64(value string) (float64, error) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float %q: %w", clean, err)
	}
	return parsed, nil
}
