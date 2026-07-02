package aepmigrate

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestVersionCapabilityRuleBlocksSourceTarget(t *testing.T) {
	rule := VersionCapabilityRule{
		SourceVersions: []VersionLabel{VersionAE2025},
		TargetSupport: map[VersionLabel]CapabilitySupport{
			VersionAE2024: CapabilityBlocked,
			VersionAE2025: CapabilityPreserved,
		},
	}
	if !rule.Blocks(VersionAE2025, VersionAE2024) {
		t.Fatal("Blocks(AE2025, AE2024) = false, want true")
	}
	if rule.Blocks(VersionAE2024, VersionAE2024) {
		t.Fatal("Blocks(AE2024, AE2024) = true, want false for source outside rule")
	}
	if rule.Blocks(VersionAE2025, VersionAE2025) {
		t.Fatal("Blocks(AE2025, AE2025) = true, want false for preserved target")
	}
}

func TestCapabilityBlockerEntriesUseProfileFindings(t *testing.T) {
	ledger := VersionCapabilityLedger{
		Rules: []VersionCapabilityRule{
			{
				ID:            explicitMatteSourceRuleID,
				CapabilityKey: "layer.set_track_matte_source",
				SourceVersions: []VersionLabel{
					VersionAE2025,
				},
				TargetSupport: map[VersionLabel]CapabilitySupport{
					VersionAE2024: CapabilityBlocked,
					VersionAE2025: CapabilityPreserved,
				},
				BlockedReason: "explicit matte blocked",
			},
		},
	}
	prof := &profile.Profile{
		Comps: []profile.Composition{
			{
				Name: "Comp 1",
				Layers: []profile.Layer{
					{Name: "Fill", MatteRef: &profile.LayerRef{Name: "Matte"}, MatteRefKind: "explicit"},
					{Name: "Matte", MatteRefKind: ""},
				},
			},
		},
	}

	entries := capabilityBlockerEntries(VersionAE2025, VersionAE2024, prof, ledger)
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1: %+v", len(entries), entries)
	}
	entry := entries[0]
	if entry.Path != `comps["Comp 1"].layers["Fill"].matte_ref` {
		t.Fatalf("Path = %q", entry.Path)
	}
	if entry.Class != ClassBlocked || entry.CapabilityKey != "layer.set_track_matte_source" || entry.Reason != "explicit matte blocked" {
		t.Fatalf("entry = %+v, want blocked explicit matte capability", entry)
	}
}
