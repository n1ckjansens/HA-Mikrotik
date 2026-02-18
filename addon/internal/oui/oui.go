package oui

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed data/oui.json
var embeddedDB []byte

type DB struct {
	vendors map[string]string
}

func LoadEmbedded() (*DB, error) {
	return Load(embeddedDB)
}

func Load(data []byte) (*DB, error) {
	m := map[string]string{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	normalized := make(map[string]string, len(m))
	for k, v := range m {
		normalized[normalizePrefix(k)] = strings.TrimSpace(v)
	}
	return &DB{vendors: normalized}, nil
}

func (db *DB) Lookup(mac string) string {
	if db == nil {
		return "Unknown"
	}
	prefix := normalizePrefix(mac)
	if len(prefix) > 6 {
		prefix = prefix[:6]
	}
	if vendor, ok := db.vendors[prefix]; ok && vendor != "" {
		return vendor
	}
	return "Unknown"
}

func normalizePrefix(v string) string {
	replacer := strings.NewReplacer(":", "", "-", "", ".", "")
	v = strings.ToUpper(strings.TrimSpace(replacer.Replace(v)))
	if len(v) >= 6 {
		return v[:6]
	}
	return v
}
