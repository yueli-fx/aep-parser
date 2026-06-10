// internal/aep/set_effect_param_test.go
//
// Go-side coverage for SetEffectParam (effect-parameter materialization,
// synthesis-lite). AE acceptance is the ship-gate's job
// (set_effect_param_shipgate_test.go).
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// gbDefaultInstanceLayer returns a layer carrying a default Gaussian Blur
// instance (params -0001..-0003 elided) on a 100% Go-built project after one
// Reopen, plus the effect itself.
func gbDefaultInstanceLayer(t *testing.T) (*aep.Project, *aep.Layer, *aep.Effect) {
	t.Helper()
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatal(err)
	}
	var l *aep.Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			if cl.Name == "S" {
				l = cl
			}
		}
	}
	if l == nil {
		t.Fatal("layer S not found after Reopen")
	}
	fx, err := aep.AddEffect(l, aep.EffectGaussianBlur)
	if err != nil {
		t.Fatal(err)
	}
	return rp, l, fx
}

func TestSetEffectParam_MaterializesElidedParams(t *testing.T) {
	rp, l, fx := gbDefaultInstanceLayer(t)
	if len(fx.Parameters) != 1 {
		t.Fatalf("precondition: default GB instance should expose 1 param, got %d", len(fx.Parameters))
	}

	// Materialize out of definition order on purpose (order must still land
	// sorted in both chunk and mirror).
	if _, err := aep.SetEffectParam(l, fx, "ADBE Gaussian Blur 2-0003", 1.0); err != nil {
		t.Fatalf("SetEffectParam(-0003): %v", err)
	}
	if _, err := aep.SetEffectParam(l, fx, "ADBE Gaussian Blur 2-0001", 25.0); err != nil {
		t.Fatalf("SetEffectParam(-0001): %v", err)
	}
	if _, err := aep.SetEffectParam(l, fx, "ADBE Gaussian Blur 2-0002", 2.0); err != nil {
		t.Fatalf("SetEffectParam(-0002): %v", err)
	}

	wantOrder := []string{
		"ADBE Gaussian Blur 2-0000",
		"ADBE Gaussian Blur 2-0001",
		"ADBE Gaussian Blur 2-0002",
		"ADBE Gaussian Blur 2-0003",
	}
	var got []string
	for _, p := range fx.Parameters {
		got = append(got, p.MatchName)
	}
	if !eq(got, wantOrder) {
		t.Fatalf("Parameters order = %v, want %v", got, wantOrder)
	}

	// Round-trip: values survive WriteAEP -> re-parse.
	var buf bytes.Buffer
	if err := rp.WriteAEP(&buf); err != nil {
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
	rfx := rl.Effects[0]
	var reOrder []string
	vals := map[string]any{}
	for _, p := range rfx.Parameters {
		reOrder = append(reOrder, p.MatchName)
		vals[p.MatchName] = p.StaticValue
	}
	if !eq(reOrder, wantOrder) {
		t.Fatalf("re-parsed order = %v, want %v", reOrder, wantOrder)
	}
	checkF64 := func(mn string, want float64) {
		t.Helper()
		f, ok := vals[mn].(float64)
		if !ok || f != want {
			t.Errorf("%s = %v, want %v", mn, vals[mn], want)
		}
	}
	checkF64("ADBE Gaussian Blur 2-0001", 25)
	checkF64("ADBE Gaussian Blur 2-0002", 2)
	checkF64("ADBE Gaussian Blur 2-0003", 1)
}

func TestSetEffectParam_ExistingParamFastPath(t *testing.T) {
	_, l, fx := gbDefaultInstanceLayer(t)
	if _, err := aep.SetEffectParam(l, fx, "ADBE Gaussian Blur 2-0001", 10.0); err != nil {
		t.Fatal(err)
	}
	n := len(fx.Parameters)
	p, err := aep.SetEffectParam(l, fx, "ADBE Gaussian Blur 2-0001", 30.0)
	if err != nil {
		t.Fatal(err)
	}
	if len(fx.Parameters) != n {
		t.Fatalf("fast path duplicated the param: %d -> %d", n, len(fx.Parameters))
	}
	if f, ok := p.StaticValue.(float64); !ok || f != 30 {
		t.Fatalf("StaticValue = %v, want 30", p.StaticValue)
	}
}

func TestSetEffectParam_Refusals(t *testing.T) {
	_, l, fx := gbDefaultInstanceLayer(t)

	if _, err := aep.SetEffectParam(l, fx, "ADBE Tint-0001", 1.0); err == nil {
		t.Error("foreign-effect param: want error, got nil")
	}
	if _, err := aep.SetEffectParam(l, fx, "ADBE Gaussian Blur 2-0042", 1.0); err == nil {
		t.Error("unknown param without template: want error, got nil")
	}
	if _, err := aep.SetEffectParam(nil, fx, "ADBE Gaussian Blur 2-0001", 1.0); err == nil {
		t.Error("nil layer: want error, got nil")
	}
	if _, err := aep.SetEffectParam(l, nil, "ADBE Gaussian Blur 2-0001", 1.0); err == nil {
		t.Error("nil effect: want error, got nil")
	}
}

func TestSetEffectParam_ParsedFixtureLayer(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects")
	}
	var gb *aep.Effect
	for _, e := range l.Effects {
		if e.MatchName == aep.EffectGaussianBlur {
			gb = e
		}
	}
	if gb == nil {
		t.Skip("baseline has no Gaussian Blur instance")
	}
	p, err := aep.SetEffectParam(l, gb, "ADBE Gaussian Blur 2-0001", 12.5)
	if err != nil {
		t.Fatalf("SetEffectParam on parsed fixture layer: %v", err)
	}
	if f, ok := p.StaticValue.(float64); !ok || f != 12.5 {
		t.Fatalf("StaticValue = %v, want 12.5", p.StaticValue)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	if _, err := aep.FromReader(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
}

func TestSupportedEffectParams(t *testing.T) {
	got := aep.SupportedEffectParams()
	if len(got) != 3 {
		t.Fatalf("SupportedEffectParams = %v, want 3 GB params", got)
	}
}
