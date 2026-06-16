package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func sampleEntries() []Entry {
	return []Entry{
		{Symbol: "NewSolidLayer", Kind: "func", Signature: "func NewSolidLayer(...)", Summary: "Adds a solid.", HasCap: true,
			Cap: Cap{Domain: "layer-create", Tier: "stable", Verify: "ae-accept", MinVer: "2020", Gate: []string{"TestSolid"}}},
		{Symbol: "SetSomething", Kind: "method", Recv: "*Project", Summary: "Sets x.", HasCap: true,
			Cap: Cap{Domain: "project", Tier: "alpha", Verify: "roundtrip", MinVer: "2024"}},
		{Symbol: "Untagged", Kind: "func", Summary: "no tag", HasCap: false},
	}
}

func TestRenderJSON(t *testing.T) {
	out, err := renderJSON(sampleEntries())
	if err != nil {
		t.Fatal(err)
	}
	var got []Entry
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tagged entries, got %d", len(got))
	}
	// sorted by domain: layer-create before project
	if got[0].Symbol != "NewSolidLayer" || got[1].Symbol != "SetSomething" {
		t.Errorf("order: %s,%s", got[0].Symbol, got[1].Symbol)
	}
	if got[0].Cap.Gate[0] != "TestSolid" {
		t.Errorf("gate lost: %+v", got[0].Cap)
	}
}

func TestRenderMarkdown(t *testing.T) {
	md := string(renderMarkdown(sampleEntries()))
	for _, want := range []string{"DO NOT EDIT", "## layer-create", "## project", "🟢stable", "🟡alpha", "需 AE ≥", "2024", "`NewSolidLayer`"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
	if strings.Contains(md, "Untagged") {
		t.Error("untagged symbol must not appear in capability table")
	}
}
