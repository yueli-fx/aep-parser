// internal/aep/animate_effect_param_test.go
//
// Go round-trip tests for AnimateEffectParam — keyframing a 1D-scalar effect
// parameter from scratch (the case InsertKeyframe refuses). The on-disk
// container is byte-identical to an AE-saved animated Gaussian-Blur-Blurriness
// fixture (verified during RE); AE acceptance + render are covered by the
// ship-gate (animate_effect_param_shipgate_test.go).
package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestAnimateEffectParam_RoundTrip(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(p, "AB", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	sq, err := aep.NewShapeLayer(comp, "SQ")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, _ := sq.RootGroup().AddRect()
	r.SetSize([2]float64{400, 400})
	f, _ := sq.RootGroup().AddFill()
	f.SetColor([4]float64{1, 1, 1, 1})

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	l := rp.Compositions[0].LayerByName("SQ")
	fx, err := aep.AddEffect(l, "ADBE Gaussian Blur 2")
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}
	prop, err := aep.AnimateEffectParam(l, fx, "ADBE Gaussian Blur 2-0001",
		[]aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}})
	if err != nil {
		t.Fatalf("AnimateEffectParam: %v", err)
	}
	if len(prop.Keyframes) != 2 {
		t.Fatalf("scene keyframes = %d, want 2", len(prop.Keyframes))
	}

	var buf bytes.Buffer
	if err := rp.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := re.Compositions[0].LayerByName("SQ")
	if rl == nil || len(rl.Effects) != 1 {
		t.Fatalf("re-parsed effects = %v", rl)
	}
	var blur *aep.Property
	for _, pp := range rl.Effects[0].Parameters {
		if pp.MatchName == "ADBE Gaussian Blur 2-0001" {
			blur = pp
		}
	}
	if blur == nil {
		t.Fatal("re-parsed Blurriness param missing")
	}
	if len(blur.Keyframes) != 2 {
		t.Fatalf("re-parsed keyframes = %d, want 2", len(blur.Keyframes))
	}
	if v, ok := blur.Keyframes[0].Value.(float64); !ok || !floatEq(v, 0.0) {
		t.Errorf("kf0 value = %v, want 0", blur.Keyframes[0].Value)
	}
	if v, ok := blur.Keyframes[1].Value.(float64); !ok || !floatEq(v, 100.0) {
		t.Errorf("kf1 value = %v, want 100", blur.Keyframes[1].Value)
	}
	if tm := blur.Keyframes[1].Time; tm < 1.99 || tm > 2.01 {
		t.Errorf("kf1 time = %v, want ~2.0s", tm)
	}
}

func TestAnimateEffectParam_RefuseTooFew(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2025)
	comp, _ := aep.NewComposition(p, "AB", 1920, 1080, 30, 5)
	sq, _ := aep.NewShapeLayer(comp, "SQ")
	r, _ := sq.RootGroup().AddRect()
	r.SetSize([2]float64{400, 400})
	rp, _ := aep.Reopen(p)
	l := rp.Compositions[0].LayerByName("SQ")
	fx, _ := aep.AddEffect(l, "ADBE Gaussian Blur 2")
	if _, err := aep.AnimateEffectParam(l, fx, "ADBE Gaussian Blur 2-0001",
		[]aep.ScalarKeyframe{{Time: 0, Value: 0}}); err == nil {
		t.Error("AnimateEffectParam with 1 keyframe: want error, got nil")
	}
}

func floatEq(a, b float64) bool {
	d := a - b
	return d < 1e-6 && d > -1e-6
}
