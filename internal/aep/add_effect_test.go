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

// TestAddEffect_AllTemplates_RoundTrip adds every supported effect to a fresh
// copy of the baseline layer and confirms it splices + survives WriteAEP →
// re-parse as the last effect in the parade. Cheap structural coverage for the
// whole template library (AE acceptance for a representative sample is the
// ship-gate's job).
func TestAddEffect_AllTemplates_RoundTrip(t *testing.T) {
	for _, name := range aep.SupportedEffects() {
		t.Run(name, func(t *testing.T) {
			proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
			if err != nil {
				t.Skipf("baseline not present: %v", err)
			}
			l := layerWithEffects(proj)
			if l == nil {
				t.Fatal("no layer with effects")
			}
			before := len(l.Effects)
			fx, err := aep.AddEffect(l, name)
			if err != nil {
				t.Fatalf("AddEffect(%q): %v", name, err)
			}
			if fx.MatchName != name {
				t.Errorf("MatchName = %q, want %q", fx.MatchName, name)
			}
			if len(l.Effects) != before+1 {
				t.Fatalf("Effects = %d, want %d", len(l.Effects), before+1)
			}

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
			if len(names) != before+1 {
				t.Fatalf("re-parsed parade len = %d, want %d (%v)", len(names), before+1, names)
			}
			if names[len(names)-1] != name {
				t.Errorf("re-parsed last effect = %q, want %q", names[len(names)-1], name)
			}
		})
	}
}

func TestRemoveEffect_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects")
	}
	// baseline: [Gaussian Blur, Tint, Fill]; remove middle (index 1, Tint).
	if err := aep.RemoveEffect(l, 1); err != nil {
		t.Fatalf("RemoveEffect: %v", err)
	}
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
	if got, want := paradeChildNames(rl), []string{"ADBE Gaussian Blur 2", "ADBE Fill"}; !eq(got, want) {
		t.Errorf("after RemoveEffect(1): parade = %v, want %v", got, want)
	}
}

func TestRemoveEffect_Refuse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects")
	}
	if err := aep.RemoveEffect(l, 99); err == nil {
		t.Error("RemoveEffect out-of-range: want error, got nil")
	}
	// Parade-less from-scratch shape layer.
	comp, _ := aep.NewComposition(aep.NewProject(aep.TargetAE2025), "M", 1920, 1080, 30, 5)
	sl, _ := aep.NewShapeLayer(comp, "S")
	if err := aep.RemoveEffect(sl.Layer, 0); err == nil {
		t.Error("RemoveEffect on parade-less layer: want error, got nil")
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

// TestEffectConstants_MatchRegistry guards against the constants and the
// template registry diverging — every exported Effect* constant must be a
// supported (addable) match-name.
func TestEffectConstants_MatchRegistry(t *testing.T) {
	consts := []string{
		aep.EffectGaussianBlur, aep.EffectFill, aep.EffectTint,
		aep.EffectBrightnessContrast, aep.EffectTritone, aep.EffectLevels,
		aep.EffectLevelsIndividual, aep.EffectHueSaturation, aep.EffectBoxBlur,
		aep.EffectGlow, aep.EffectInvert, aep.EffectExposure,
	}
	supported := map[string]bool{}
	for _, n := range aep.SupportedEffects() {
		supported[n] = true
	}
	if len(consts) != len(aep.SupportedEffects()) {
		t.Errorf("constant count %d != SupportedEffects count %d", len(consts), len(aep.SupportedEffects()))
	}
	for _, c := range consts {
		if !supported[c] {
			t.Errorf("constant %q not in SupportedEffects()", c)
		}
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
