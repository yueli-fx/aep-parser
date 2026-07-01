package aepmigrate

import "testing"

func TestDefaultCapabilityLedgerDescribesExplicitMatteBoundary(t *testing.T) {
	ledger := DefaultCapabilityLedger()
	rule, ok := ledger.RuleByID("layer-explicit-matte-source")
	if !ok {
		t.Fatal("explicit matte rule missing from default capability ledger")
	}
	if rule.CapabilityKey != "layer.set_track_matte_source" {
		t.Fatalf("CapabilityKey = %q, want layer.set_track_matte_source", rule.CapabilityKey)
	}
	for _, target := range []VersionLabel{VersionAE2020, VersionAE2021, VersionAE2022, VersionAE2023, VersionAE2024} {
		if got := rule.TargetSupport[target]; got != CapabilityBlocked {
			t.Fatalf("TargetSupport[%s] = %q, want %q", target, got, CapabilityBlocked)
		}
	}
	if got := rule.TargetSupport[VersionAE2025]; got != CapabilityPreserved {
		t.Fatalf("TargetSupport[AE2025] = %q, want %q", got, CapabilityPreserved)
	}
	if rule.BlockedReason == "" {
		t.Fatal("BlockedReason is empty")
	}
}

func TestAssessExplicitMatteBoundaryUsesCapabilityLedger(t *testing.T) {
	source := writeTempProjectWithExplicitMatte(t)
	ledger := DefaultCapabilityLedger()
	rule, ok := ledger.RuleByID("layer-explicit-matte-source")
	if !ok {
		t.Fatal("explicit matte rule missing from default capability ledger")
	}

	report, err := Assess(Options{InputPath: source, Target: VersionAE2020})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}

	for _, entry := range report.Entries {
		if entry.Class != ClassBlocked {
			continue
		}
		if entry.CapabilityKey != rule.CapabilityKey {
			t.Fatalf("CapabilityKey = %q, want %q", entry.CapabilityKey, rule.CapabilityKey)
		}
		if entry.Reason != rule.BlockedReason {
			t.Fatalf("Reason = %q, want %q", entry.Reason, rule.BlockedReason)
		}
		return
	}
	t.Fatalf("no blocked entry found: %+v", report.Entries)
}
