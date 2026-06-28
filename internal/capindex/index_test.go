package capindex

import "testing"

func TestQueryMatchesSymbolDomainSummaryAndAlias(t *testing.T) {
	idx := New([]Entry{
		{Symbol: "NewSolidLayer", Summary: "Adds a solid layer.", HasCap: true,
			Cap: Cap{Domain: "layer-create", Tier: "stable", Verify: "ae-accept", Alias: []string{"纯色", "solid"}}},
		{Symbol: "AddEffect", Summary: "Appends an effect.", HasCap: true,
			Cap: Cap{Domain: "effect", Tier: "stable", Verify: "render-pixel", Alias: []string{"特效"}}},
		{Symbol: "Untagged", HasCap: false},
	})

	if got := idx.Query("solid"); len(got) != 1 || got[0].Symbol != "NewSolidLayer" {
		t.Errorf("solid: %v", got)
	}
	if got := idx.Query("SOLID"); len(got) != 1 {
		t.Errorf("case-insensitive failed: %v", got)
	}
	if got := idx.Query("特效"); len(got) != 1 || got[0].Symbol != "AddEffect" {
		t.Errorf("CJK alias: %v", got)
	}
	if got := idx.Query("effect"); len(got) != 1 || got[0].Symbol != "AddEffect" {
		t.Errorf("domain match: %v", got)
	}
	if got := idx.Query(""); len(got) != 2 {
		t.Errorf("empty term should return all tagged: %d", len(got))
	}
}

func TestLookupReportsStatusFromTier(t *testing.T) {
	idx := New([]Entry{
		{Symbol: "Ready", HasCap: true, Cap: Cap{Domain: "x", Tier: "stable"}},
		{Symbol: "Experimental", HasCap: true, Cap: Cap{Domain: "x", Tier: "alpha"}},
		{Symbol: "Absent", HasCap: true, Cap: Cap{Domain: "x", Tier: "missing"}},
	})

	if got := idx.Lookup("Ready"); got.Status != StatusSupported || got.Entry.Symbol != "Ready" {
		t.Fatalf("Ready lookup = %+v", got)
	}
	if got := idx.Lookup("Experimental"); got.Status != StatusSupported || got.Entry.Cap.Tier != "alpha" {
		t.Fatalf("Experimental lookup = %+v", got)
	}
	if got := idx.Lookup("Absent"); got.Status != StatusUnsupported {
		t.Fatalf("Absent lookup = %+v", got)
	}
	if got := idx.Lookup("nope"); got.Status != StatusUnknown {
		t.Fatalf("nope lookup = %+v", got)
	}
}

func TestLookupPrefersExactReceiverSymbol(t *testing.T) {
	idx := New([]Entry{
		{Symbol: "SetSize", Recv: "*EllipseNode", Summary: "ellipse size", HasCap: true, Cap: Cap{Domain: "shape", Tier: "stable"}},
		{Symbol: "SetSize", Recv: "*RectNode", Summary: "rect size", HasCap: true, Cap: Cap{Domain: "shape", Tier: "stable"}},
	})

	got := idx.Lookup("RectNode.SetSize")
	if got.Entry.Recv != "*RectNode" {
		t.Fatalf("Lookup RectNode.SetSize = %+v", got.Entry)
	}
}
