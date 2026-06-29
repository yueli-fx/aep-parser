package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// from-scratch 3D shape → synthesize material leaves → round-trip persists.
func newReopened3DCaster(t *testing.T) (*aep.Project, *aep.Layer) {
	t.Helper()
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "C", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	sh, _ := aep.NewShapeLayer(comp, "CASTER")
	r, _ := sh.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 200})
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	l := rp.Compositions[0].LayerByName("CASTER")
	if err := l.SetIs3D(true); err != nil {
		t.Fatalf("SetIs3D: %v", err)
	}
	return rp, l
}

func TestSetMaterialOptionSynthesisRoundtrip(t *testing.T) {
	rp, caster := newReopened3DCaster(t)

	if caster.MaterialCastsShadows() != nil {
		t.Fatal("from-scratch caster should have NO Casts Shadows leaf before synthesis")
	}
	prop, err := aep.SetMaterialOption(caster, "ADBE Casts Shadows", float64(aep.MaterialCastsOn))
	if err != nil {
		t.Fatalf("SetMaterialOption(Casts Shadows): %v", err)
	}
	if prop == nil || prop.StaticValue != float64(aep.MaterialCastsOn) {
		t.Fatalf("returned prop value = %v, want %d", prop, aep.MaterialCastsOn)
	}
	if caster.MaterialCastsShadows() == nil {
		t.Fatal("MaterialCastsShadows() still nil after synthesis")
	}

	// A second material leaf (higher ordinal — exercises canonical-order insert).
	if _, err := aep.SetMaterialOption(caster, "ADBE Diffuse Coefficient", 0.7); err != nil {
		t.Fatalf("SetMaterialOption(Diffuse): %v", err)
	}

	var buf bytes.Buffer
	if err := rp.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rc := re.Compositions[0].LayerByName("CASTER")
	if !rc.Is3D {
		t.Error("Is3D lost on round-trip")
	}
	if p := rc.MaterialCastsShadows(); p == nil || p.StaticValue != float64(aep.MaterialCastsOn) {
		t.Errorf("post-roundtrip MaterialCastsShadows = %v, want %d", staticValueOf(p), aep.MaterialCastsOn)
	}
	if p := rc.MaterialDiffuse(); p == nil || p.StaticValue != 0.7 {
		t.Errorf("post-roundtrip MaterialDiffuse = %v, want 0.7", staticValueOf(p))
	}
}

func TestSetMaterialOptionAlreadyPresent(t *testing.T) {
	_, caster := newReopened3DCaster(t)
	if _, err := aep.SetMaterialOption(caster, "ADBE Casts Shadows", float64(aep.MaterialCastsOff)); err != nil {
		t.Fatalf("first set: %v", err)
	}
	// Second call hits the already-present path (plain SetStaticValue).
	p, err := aep.SetMaterialOption(caster, "ADBE Casts Shadows", float64(aep.MaterialCastsOnly))
	if err != nil {
		t.Fatalf("second set: %v", err)
	}
	if p.StaticValue != float64(aep.MaterialCastsOnly) {
		t.Errorf("value = %v, want %d", p.StaticValue, aep.MaterialCastsOnly)
	}
}

func TestSetMaterialOptionErrors(t *testing.T) {
	_, caster := newReopened3DCaster(t)
	if _, err := aep.SetMaterialOption(nil, "ADBE Casts Shadows", 1.0); err == nil {
		t.Error("nil layer: expected error")
	}
	if _, err := aep.SetMaterialOption(caster, "ADBE Not A Material Prop", 1.0); err == nil {
		t.Error("unknown match-name: expected error")
	}
}
