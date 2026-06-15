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

	aep "github.com/example/aep-parser/internal/aep"
)

func TestSetMaskPath_RectToTriangle_RoundTrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
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

func TestSetMaskPath_RefuseFewVertices(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_property_struct_baseline.aep")
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
