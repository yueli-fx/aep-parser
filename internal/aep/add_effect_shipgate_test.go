// internal/aep/add_effect_shipgate_test.go
//
// AE ship gate for AddEffect on the Effect Parade. Reuses the generic
// runPropStructShipGate harness (verify_property_struct.jsx): loads the
// AE-2020-native baseline (3 effects), Go-adds a 4th effect from an embedded
// template, WriteAEP, and has AE open the mutated file — proving AE ACCEPTS the
// Go-spliced effect (no data-loss / corrupt) and reads back the expected 4
// effects in order. Resaves so the Go side confirms AE kept the addition.
//
// Gated by AE_SHIP_GATE. Baseline built by test_data/re_property_struct.jsx.
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func addGBlurEffect(l *aep.Layer) error {
	_, err := aep.AddEffect(l, "ADBE Gaussian Blur 2")
	return err
}

// Baseline parade is [Gaussian Blur, Tint, Fill]; AddEffect appends a second
// Gaussian Blur at the end.
var addEffectExpect = []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill", "ADBE Gaussian Blur 2"}

func TestAddEffect_AEShipGate_AE2020(t *testing.T) {
	runPropStructShipGate(t, ae2020(), "addeffect-AE2020", addGBlurEffect, addEffectExpect)
}

func TestAddEffect_AEShipGate_AE2025(t *testing.T) {
	runPropStructShipGate(t, ae2025(), "addeffect-AE2025", addGBlurEffect, addEffectExpect)
}
