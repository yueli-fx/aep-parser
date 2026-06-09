// internal/aep/add_effect_test.go
//
// Go round-trip tests for AddEffect (no AE required). Proves the spliced effect
// pair survives WriteAEP → re-parse with the new effect present, ordered last,
// and with settable parameters. AE acceptance is covered separately by the
// ship-gate (add_effect_shipgate_test.go).
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// layerWithEffects is defined in property_structural_shipgate_test.go.

func TestAddEffect_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	before := len(l.Effects)

	fx, err := aep.AddEffect(l, "ADBE Gaussian Blur 2")
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}
	if fx == nil {
		t.Fatal("AddEffect returned nil effect")
	}
	if got := len(l.Effects); got != before+1 {
		t.Fatalf("layer.Effects = %d, want %d", got, before+1)
	}
	if l.Effects[len(l.Effects)-1] != fx {
		t.Errorf("added effect is not last in layer.Effects")
	}
	if fx.MatchName != "ADBE Gaussian Blur 2" {
		t.Errorf("effect MatchName = %q, want ADBE Gaussian Blur 2", fx.MatchName)
	}

	// Tune the new effect's first surfaced param before writing (the default
	// GBlur instance surfaces "-0000"; the others are default-elided).
	const blurParam = "ADBE Gaussian Blur 2-0000"
	var blur *aep.Property
	for _, p := range fx.Parameters {
		if p.MatchName == blurParam {
			blur = p
		}
	}
	if blur == nil {
		t.Fatalf("%s param not found among %d params", blurParam, len(fx.Parameters))
	}
	if err := blur.SetStaticValue(33.0); err != nil {
		t.Fatalf("SetStaticValue: %v", err)
	}

	// WriteAEP → re-parse and confirm AE-independent structural survival.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("re-parsed: no layer with effects")
	}
	names := paradeChildNames(rl)
	want := []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill", "ADBE Gaussian Blur 2"}
	if !eq(names, want) {
		t.Errorf("re-parsed parade = %v, want %v", names, want)
	}

	// The Blurriness value we set round-trips on the last (added) effect.
	rfx := rl.Effects[len(rl.Effects)-1]
	var rblur *aep.Property
	for _, p := range rfx.Parameters {
		if p.MatchName == blurParam {
			rblur = p
		}
	}
	if rblur == nil {
		t.Fatal("re-parsed param not found")
	}
	if v, ok := rblur.StaticValue.(float64); !ok || v != 33.0 {
		t.Errorf("re-parsed Blurriness = %v (ok=%v), want 33", rblur.StaticValue, ok)
	}
}

func TestAddEffect_Unsupported(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects")
	}
	if _, err := aep.AddEffect(l, "ADBE No Such Effect"); err == nil {
		t.Error("AddEffect with unsupported name: want error, got nil")
	}
}

func TestAddEffect_NoParade(t *testing.T) {
	// A from-scratch shape layer has no Effect Parade yet → AddEffect refuses.
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	sl, err := aep.NewShapeLayer(comp, "S")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddEffect(sl.Layer, "ADBE Gaussian Blur 2"); err == nil {
		t.Error("AddEffect on parade-less layer: want error, got nil")
	}
}

func TestSupportedEffects(t *testing.T) {
	got := aep.SupportedEffects()
	if len(got) == 0 {
		t.Fatal("SupportedEffects returned empty")
	}
	found := false
	for _, n := range got {
		if n == "ADBE Gaussian Blur 2" {
			found = true
		}
	}
	if !found {
		t.Errorf("SupportedEffects %v missing ADBE Gaussian Blur 2", got)
	}
}
