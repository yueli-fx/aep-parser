// internal/aep/mask_path_test.go
//
// Go round-trip tests for SetMaskPath — reshaping an existing mask's outline,
// including changing the vertex count (rectangle → triangle). Proves the rebuilt
// om-s survives WriteAEP → re-parse with the new geometry. AE acceptance +
// render are covered by the ship-gate (mg_mask_path_shipgate_test.go).
package aep_test

import (
	"bytes"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestSetMaskPath_RectToTriangle_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}

	m, err := aep.AddMask(l, "Reshape", rectPath()) // 4-vertex rectangle
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if len(m.Vertices) != 4 {
		t.Fatalf("initial mask vertices = %d, want 4", len(m.Vertices))
	}

	tri := aep.BezierPath{
		Vertices: [][2]float64{{100, 10}, {190, 190}, {10, 190}},
		Closed:   true,
	}
	if err := aep.SetMaskPath(l, m, tri); err != nil {
		t.Fatalf("SetMaskPath: %v", err)
	}
	if len(m.Vertices) != 3 {
		t.Errorf("after reshape, scene vertices = %d, want 3", len(m.Vertices))
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil || len(rl.Masks) != 1 {
		t.Fatalf("re-parsed masks = %v, want 1", rl)
	}
	rm := rl.Masks[0]
	if len(rm.Vertices) != 3 {
		t.Fatalf("re-parsed vertices = %d, want 3 (triangle)", len(rm.Vertices))
	}
	if !rm.Closed {
		t.Errorf("re-parsed triangle should be closed")
	}
	// Vertices are bbox-normalized; the apex (input 100,10 — top-centre) must
	// normalize to x=0.5, y=0 within the triangle's bounding box.
	apex := rm.Vertices[0].Anchor
	if apex[1] > 0.01 {
		t.Errorf("apex y = %v, want ~0 (top of bbox)", apex[1])
	}
	if apex[0] < 0.4 || apex[0] > 0.6 {
		t.Errorf("apex x = %v, want ~0.5 (horizontal centre)", apex[0])
	}
}

func TestSetMaskPathKeyframes_RectToTriangle_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}

	m, err := aep.AddMask(l, "Morph", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}

	tri := aep.BezierPath{
		Vertices: [][2]float64{{100, 10}, {190, 190}, {10, 190}},
		Closed:   true,
	}
	keys := []aep.MaskPathKey{
		{Time: 0, Path: rectPath()}, // 4-vertex rectangle
		{Time: 1, Path: tri},        // 3-vertex triangle
	}
	if err := aep.SetMaskPathKeyframes(l, m, keys); err != nil {
		t.Fatalf("SetMaskPathKeyframes: %v", err)
	}
	if len(m.PathKeyframes) != 2 {
		t.Fatalf("scene PathKeyframes = %d, want 2", len(m.PathKeyframes))
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	rl := maskParadeLayer(re, l.ID)
	if rl == nil || len(rl.Masks) != 1 {
		t.Fatalf("re-parsed masks = %v, want 1", rl)
	}
	rm := rl.Masks[0]
	if len(rm.PathKeyframes) != 2 {
		t.Fatalf("re-parsed PathKeyframes = %d, want 2", len(rm.PathKeyframes))
	}
	if n := len(rm.PathKeyframes[0].Vertices); n != 4 {
		t.Errorf("kf0 vertices = %d, want 4 (rect)", n)
	}
	if n := len(rm.PathKeyframes[1].Vertices); n != 3 {
		t.Errorf("kf1 vertices = %d, want 3 (triangle)", n)
	}
	// Second keyframe at t=1s.
	if got := rm.PathKeyframes[1].Time; got < 0.99 || got > 1.01 {
		t.Errorf("kf1 time = %v, want ~1.0s", got)
	}
	// Triangle apex (input 100,10) normalizes to top-centre of its bbox.
	apex := rm.PathKeyframes[1].Vertices[0].Anchor
	if apex[1] > 0.01 || apex[0] < 0.4 || apex[0] > 0.6 {
		t.Errorf("kf1 apex = %v, want ~(0.5, 0)", apex)
	}
}

func TestSetMaskPathKeyframes_RefuseTooFew(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	m, err := aep.AddMask(l, "M", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if err := aep.SetMaskPathKeyframes(l, m, []aep.MaskPathKey{{Time: 0, Path: rectPath()}}); err == nil {
		t.Error("SetMaskPathKeyframes with 1 keyframe: want error, got nil")
	}
}

func TestSetMaskPath_RefuseFewVertices(t *testing.T) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("baseline not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	m, err := aep.AddMask(l, "M", rectPath())
	if err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	if err := aep.SetMaskPath(l, m, aep.BezierPath{Vertices: [][2]float64{{1, 1}}, Closed: true}); err == nil {
		t.Error("SetMaskPath with 1 vertex: want error, got nil")
	}
}
