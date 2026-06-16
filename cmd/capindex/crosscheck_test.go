package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "known-incident.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gates := map[string]bool{"TestGood": false, "TestSkipped": true}

	mk := func(c Cap) Entry { return Entry{Symbol: "Sym", HasCap: true, Cap: c} }

	cases := []struct {
		name    string
		entry   Entry
		wantErr bool
	}{
		{"valid", mk(Cap{Domain: "layer-create", Tier: "stable", Verify: "ae-accept", Gate: []string{"TestGood"}, Incident: []string{"known-incident"}}), false},
		{"missing gate", mk(Cap{Domain: "x", Tier: "stable", Verify: "ae-accept", Gate: []string{"TestNope"}}), true},
		{"skipped gate", mk(Cap{Domain: "x", Tier: "stable", Verify: "ae-accept", Gate: []string{"TestSkipped"}}), true},
		{"dangling incident", mk(Cap{Domain: "x", Tier: "stable", Verify: "ae-accept", Gate: []string{"TestGood"}, Incident: []string{"ghost"}}), true},
		{"invalid cap", mk(Cap{Domain: "x", Tier: "stable", Verify: "roundtrip", Gate: []string{"TestGood"}}), true},
	}
	for _, tc := range cases {
		errs := validateEntries([]Entry{tc.entry}, gates, dir)
		if tc.wantErr && len(errs) == 0 {
			t.Errorf("%s: expected error, got none", tc.name)
		}
		if !tc.wantErr && len(errs) != 0 {
			t.Errorf("%s: expected no error, got %v", tc.name, errs)
		}
	}
}

func TestValidateEntries_ParseErr(t *testing.T) {
	e := Entry{Symbol: "Bad", HasCap: true, parseErr: os.ErrInvalid}
	if errs := validateEntries([]Entry{e}, nil, t.TempDir()); len(errs) == 0 {
		t.Error("expected parseErr to surface")
	}
}
