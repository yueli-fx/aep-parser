package main

import "testing"

func TestQuery(t *testing.T) {
	entries := []Entry{
		{Symbol: "NewSolidLayer", Summary: "Adds a solid layer.", HasCap: true,
			Cap: Cap{Domain: "layer-create", Tier: "stable", Verify: "ae-accept", Alias: []string{"纯色", "solid"}}},
		{Symbol: "AddEffect", Summary: "Appends an effect.", HasCap: true,
			Cap: Cap{Domain: "effect", Tier: "stable", Verify: "render-pixel", Alias: []string{"特效"}}},
		{Symbol: "Untagged", HasCap: false},
	}

	if got := query(entries, "solid"); len(got) != 1 || got[0].Symbol != "NewSolidLayer" {
		t.Errorf("solid: %v", got)
	}
	if got := query(entries, "SOLID"); len(got) != 1 {
		t.Errorf("case-insensitive failed: %v", got)
	}
	if got := query(entries, "特效"); len(got) != 1 || got[0].Symbol != "AddEffect" {
		t.Errorf("CJK alias: %v", got)
	}
	if got := query(entries, "effect"); len(got) != 1 || got[0].Symbol != "AddEffect" {
		t.Errorf("domain match: %v", got)
	}
	if got := query(entries, ""); len(got) != 2 {
		t.Errorf("empty term should return all tagged: %d", len(got))
	}
}
