package routeros

import "testing"

func TestMapARPRowsCompleteness(t *testing.T) {
	t.Helper()

	rows := []map[string]string{
		{
			"mac-address": "AA:BB:CC:DD:EE:01",
			"address":     "192.168.88.10",
			"interface":   "bridge",
			"complete":    "yes",
		},
		{
			"mac-address": "AA:BB:CC:DD:EE:02",
			"address":     "192.168.88.11",
			"interface":   "bridge",
			"flags":       "DC",
		},
		{
			"mac-address": "AA:BB:CC:DD:EE:03",
			"address":     "192.168.88.12",
			"interface":   "bridge",
			"status":      "complete",
		},
		{
			"mac-address": "AA:BB:CC:DD:EE:04",
			"address":     "192.168.88.13",
			"interface":   "bridge",
			"flags":       "D",
			"status":      "stale",
		},
	}

	items := mapARPRows(rows)
	if len(items) != 4 {
		t.Fatalf("unexpected mapped items count: %d", len(items))
	}

	if !items[0].Complete {
		t.Fatalf("expected complete=yes row to be complete")
	}
	if !items[1].Complete {
		t.Fatalf("expected flags containing C row to be complete")
	}
	if !items[2].Complete {
		t.Fatalf("expected status=complete row to be complete")
	}
	if items[3].Complete {
		t.Fatalf("expected incomplete row to stay incomplete")
	}
}
