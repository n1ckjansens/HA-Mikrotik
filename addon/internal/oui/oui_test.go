package oui

import "testing"

func TestLoadAndLookup(t *testing.T) {
	data := []byte(`{"000C42":"MikroTik","AABBCC":"VendorX"}`)
	db, err := Load(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := db.Lookup("00:0c:42:11:22:33"); got != "MikroTik" {
		t.Fatalf("expected MikroTik, got %s", got)
	}
	if got := db.Lookup("AA-BB-CC-01-02-03"); got != "VendorX" {
		t.Fatalf("expected VendorX, got %s", got)
	}
	if got := db.Lookup("11:22:33:44:55:66"); got != "Unknown" {
		t.Fatalf("expected Unknown, got %s", got)
	}
}
