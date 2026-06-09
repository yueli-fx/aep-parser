// internal/aep/add_effect_shipgate_test.go
//
// AE ship gate for AddEffect on the Effect Parade. Reuses the generic
// runPropStructShipGate harness (verify_property_struct.jsx): loads the
// AE-2020-native baseline (3 effects), Go-adds a 4th effect from an embedded
// template, WriteAEP, and has AE open the mutated file — proving AE ACCEPTS the
// Go-spliced effect (no data-loss / corrupt) and reads back the expected 4
// effects in order. Resaves so the Go side confirms AE kept the addition.
//
// Sampled across the template library by sspc-payload size — a small effect
// (Gaussian Blur 1.7KB), mid (Fill 2.5KB, Tint 2.2KB / 2 params) and the two
// largest (Easy Levels2 11.7KB, Pro Levels2 20.6KB) — to prove the splice +
// bottom-up size recompute holds regardless of payload size. The remaining
// templates ride the identical mechanism (Go round-trip covered by
// TestAddEffect_AllTemplates_RoundTrip).
//
// Gated by AE_SHIP_GATE. Baseline built by test_data/re_property_struct.jsx.
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// addEffectSample is the ship-gated subset (match-name → resulting parade order).
// Baseline parade is [Gaussian Blur, Tint, Fill]; the added effect appends last.
var addEffectSample = []string{
	"ADBE Gaussian Blur 2",
	"ADBE Fill",
	"ADBE Tint",
	"ADBE Easy Levels2",
	"ADBE Pro Levels2",
}

func runAddEffectGate(t *testing.T, aeExe, ver string) {
	for _, fxName := range addEffectSample {
		expect := []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill", fxName}
		mutate := func(l *aep.Layer) error {
			_, err := aep.AddEffect(l, fxName)
			return err
		}
		t.Run(fxName, func(t *testing.T) {
			runPropStructShipGate(t, aeExe, "addeffect-"+ver+"-"+fxName, mutate, expect)
		})
	}
}

func TestAddEffect_AEShipGate_AE2020(t *testing.T) { runAddEffectGate(t, ae2020(), "AE2020") }
func TestAddEffect_AEShipGate_AE2025(t *testing.T) { runAddEffectGate(t, ae2025(), "AE2025") }
