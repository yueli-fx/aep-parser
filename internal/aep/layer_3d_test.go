package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestGeometryOptionsTypedSettersRoundtrip exercises Layer.Geometry*
// typed getters/setters against re_geometry_options.aep. AE 24+
// Advanced 3D renderer enabled in fixture; Plane Curvature / Plane
// Subdivision / Bevel Direction present.
func TestGeometryOptionsTypedSettersRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_geometry_options.aep")
	if err != nil {
		t.Skipf("re_geometry_options.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "GEO_OPT" {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("GEO_OPT comp missing")
	}
	var solid *aep.Layer
	for _, l := range comp.Layers {
		if l.GeometryPlaneCurvature() != nil {
			solid = l
			break
		}
	}
	if solid == nil {
		t.Fatal("3D solid with geometry options not found")
	}
	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
	mustNoErr("SetGeometryPlaneCurvature", solid.SetGeometryPlaneCurvature(0.4))
	mustNoErr("SetGeometryPlaneSubdivision", solid.SetGeometryPlaneSubdivision(8))
	mustNoErr("SetGeometryBevelDirection", solid.SetGeometryBevelDirection(2))

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var reSolid *aep.Layer
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			if p := l.GeometryPlaneCurvature(); p != nil {
				if v, ok := p.StaticValue.(float64); ok && math.Abs(v-0.4) < 1e-3 {
					reSolid = l
				}
			}
		}
	}
	if reSolid == nil {
		t.Fatal("post-roundtrip solid missing")
	}
	if v := reSolid.GeometryPlaneSubdivision().StaticValue.(float64); v != 8 {
		t.Errorf("post-roundtrip Subdivision = %g, want 8", v)
	}
	if v := reSolid.GeometryBevelDirection().StaticValue.(float64); v != 2 {
		t.Errorf("post-roundtrip BevelDirection = %g, want 2", v)
	}
}

// TestMaterialOptionsTypedSettersRoundtrip exercises Layer.Material* typed
// getters/setters against re_material_options.aep, which has 3D-enabled
// solids carrying the full 17-property materialOption tree.
func TestMaterialOptionsTypedSettersRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_material_options.aep")
	if err != nil {
		t.Skipf("re_material_options.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "MAT_OPT" {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("MAT_OPT comp missing")
	}
	// Find the 3D solid with custom material (has Specular = 0.6).
	var solid *aep.Layer
	for _, l := range comp.Layers {
		if sp := l.MaterialSpecular(); sp != nil {
			if v, ok := sp.StaticValue.(float64); ok && v > 0.5 && v < 0.7 {
				solid = l
			}
		}
	}
	if solid == nil {
		t.Fatal("3D solid with custom Specular not found")
	}

	// Verify getters surface fixture values.
	if v := solid.MaterialCastsShadows().StaticValue.(float64); v != 2 {
		t.Errorf("CastsShadows = %g, want 2 (Only)", v)
	}
	if v := solid.MaterialAmbient().StaticValue.(float64); math.Abs(v-0.5) > 1e-3 {
		t.Errorf("Ambient = %g, want ~0.5", v)
	}

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
	// Exercise material setters. NOTE: AE prunes properties from
	// serialization when they sit at default — AcceptsShadows on this
	// custom layer is missing because the JSX wrote it = 1 (default).
	// Shininess Coefficient is also pruned on the custom layer (JSX
	// didn't touch it). Find the sister default-valued layer to test
	// those two.
	mustNoErr("SetMaterialCastsShadows", solid.SetMaterialCastsShadows(aep.MaterialCastsOn))
	mustNoErr("SetMaterialLightTransmission", solid.SetMaterialLightTransmission(0.3))
	mustNoErr("SetMaterialAcceptsLights", solid.SetMaterialAcceptsLights(true))
	mustNoErr("SetMaterialAppearsInReflections", solid.SetMaterialAppearsInReflections(false))
	mustNoErr("SetMaterialAmbient", solid.SetMaterialAmbient(0.75))
	mustNoErr("SetMaterialDiffuse", solid.SetMaterialDiffuse(0.65))
	mustNoErr("SetMaterialSpecular", solid.SetMaterialSpecular(0.55))
	mustNoErr("SetMaterialMetal", solid.SetMaterialMetal(0.1))
	mustNoErr("SetMaterialReflection", solid.SetMaterialReflection(0.4))
	mustNoErr("SetMaterialFresnel", solid.SetMaterialFresnel(0.2))
	mustNoErr("SetMaterialTransparency", solid.SetMaterialTransparency(0.0))
	mustNoErr("SetMaterialTranspRolloff", solid.SetMaterialTranspRolloff(0.0))
	mustNoErr("SetMaterialIndexOfRefraction", solid.SetMaterialIndexOfRefraction(1.5))
	mustNoErr("SetMaterialShadowColor", solid.SetMaterialShadowColor([]float64{0.5, 0, 0, 1}))
	if g := solid.MaterialGlossiness(); g == nil {
		t.Error("Glossiness getter returned nil")
	}

	// Find the default-valued 3D solid that still carries AcceptsShadows
	// and Shininess (since they're at default, weren't pruned).
	var defSolid *aep.Layer
	for _, l := range comp.Layers {
		if l == solid {
			continue
		}
		if l.MaterialAcceptsShadows() != nil && l.MaterialShininess() != nil {
			defSolid = l
		}
	}
	if defSolid != nil {
		mustNoErr("SetMaterialAcceptsShadows", defSolid.SetMaterialAcceptsShadows(false))
		mustNoErr("SetMaterialShininess", defSolid.SetMaterialShininess(45))
	} else {
		t.Log("default-valued 3D solid not found; skipping AcceptsShadows/Shininess setter test")
	}

	// Roundtrip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var reSolid *aep.Layer
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			if sp := l.MaterialSpecular(); sp != nil {
				if v, ok := sp.StaticValue.(float64); ok && math.Abs(v-0.55) < 1e-3 {
					reSolid = l
				}
			}
		}
	}
	if reSolid == nil {
		t.Fatal("post-roundtrip solid missing")
	}
	if v := reSolid.MaterialCastsShadows().StaticValue.(float64); v != 1 {
		t.Errorf("post-roundtrip CastsShadows = %g, want 1", v)
	}
	if v := reSolid.MaterialIndexOfRefraction().StaticValue.(float64); math.Abs(v-1.5) > 1e-3 {
		t.Errorf("post-roundtrip IOR = %g, want 1.5", v)
	}

	// Negative: 2D layer (no materialOption) should error.
	var solid2D *aep.Layer
	for _, l := range comp.Layers {
		if l.MaterialSpecular() == nil {
			solid2D = l
			break
		}
	}
	if solid2D != nil {
		if err := solid2D.SetMaterialSpecular(0.5); err == nil {
			t.Error("SetMaterialSpecular on 2D layer should error")
		}
	}
}

// TestTransformTypedSettersRoundtrip exercises Layer transform-group
// typed setters against re_cameralight.aep. MyCamera is 3D-enabled in
// the fixture, so RotateX/Y/Orientation are present too.
func TestTransformTypedSettersRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CL" {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("RE_CL comp missing")
	}
	var cam *aep.Layer
	for _, l := range comp.Layers {
		if l.Name == "MyCamera" {
			cam = l
		}
	}
	if cam == nil {
		t.Fatal("MyCamera missing")
	}

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
	mustNoErr("SetAnchorPoint", cam.SetAnchorPoint([]float64{100, 200, 300}))
	mustNoErr("SetPosition", cam.SetPosition([]float64{500, 600, -700}))
	mustNoErr("SetScale", cam.SetScale([]float64{2, 2, 2}))
	mustNoErr("SetRotation", cam.SetRotation(45))
	mustNoErr("SetOpacity", cam.SetOpacity(0.5))
	// Camera in this fixture has RotateZ-only (no Is3D rotation axes in the
	// property dump); RotateX / RotateY / Orientation may or may not be
	// present. Probe and skip if absent (typed setter surfaces this).
	if err := cam.SetRotateX(30); err != nil {
		t.Logf("SetRotateX: %v (expected if not 3D)", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var reCam *aep.Layer
	for _, c := range re.Compositions {
		for _, l := range c.Layers {
			if l.Name == "MyCamera" {
				reCam = l
			}
		}
	}
	if reCam == nil {
		t.Fatal("post-roundtrip MyCamera missing")
	}
	got, _ := reCam.Position().StaticValue.([]float64)
	want := []float64{500, 600, -700}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("post-roundtrip Position = %v, want %v", got, want)
	}
	if v, _ := reCam.Rotation().StaticValue.(float64); v != 45 {
		t.Errorf("post-roundtrip Rotation = %g, want 45", v)
	}
	if v, _ := reCam.Opacity().StaticValue.(float64); v != 0.5 {
		t.Errorf("post-roundtrip Opacity = %g, want 0.5", v)
	}
	gotS, _ := reCam.Scale().StaticValue.([]float64)
	if len(gotS) != 3 || gotS[0] != 2 {
		t.Errorf("post-roundtrip Scale = %v, want [2 2 2]", gotS)
	}
}
