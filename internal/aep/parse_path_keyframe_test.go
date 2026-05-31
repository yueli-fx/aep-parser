// internal/aep/parse_path_keyframe_test.go
//
// Phase 1 hydrate: an animated shape path (N shap LISTs + tdbs time table,
// == animated mask structure) must hydrate into a single PathNode whose
// Path() stream carries N keyframes (time + denormalized geometry), not N
// separate static PathNodes.
//
// Fixture: test_data/re_path_anim.jsx (3 LINEAR keyframes, distinct vertex
// count + bbox per frame). See incident-reports/path-keyframe-write-re.md.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestPathKeyframe_HydrateAnimated(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_path_anim.aep")
	if err != nil {
		t.Skipf("re_path_anim.aep not present; run test_data/re_path_anim.jsx in AE")
	}
	if len(proj.Compositions) == 0 {
		t.Fatal("no compositions")
	}
	comp := proj.Compositions[0]
	var shapeLayer *aep.Layer
	for _, l := range comp.Layers {
		if l.Name == "PathAnim" {
			shapeLayer = l
			break
		}
	}
	if shapeLayer == nil {
		t.Fatalf("PathAnim layer not found (layers=%d)", len(comp.Layers))
	}

	sl := aep.WrapShapeLayer(shapeLayer)
	var pathNode *aep.PathNode
	for _, child := range sl.RootGroup().Children {
		if pn, ok := child.(*aep.PathNode); ok {
			pathNode = pn
			break
		}
	}
	if pathNode == nil {
		t.Fatal("no PathNode in shape root group")
	}

	path := pathNode.Path()
	kfs := path.Keyframes()
	if len(kfs) != 3 {
		t.Fatalf("path keyframes = %d, want 3 (animated, one per frame)", len(kfs))
	}

	wantTimes := []float64{0, 1, 2}
	wantNVerts := []int{4, 4, 3}
	for i, kf := range kfs {
		if kf.Time != wantTimes[i] {
			t.Errorf("kf[%d].Time = %g, want %g", i, kf.Time, wantTimes[i])
		}
		if len(kf.Value.Vertices) != wantNVerts[i] {
			t.Errorf("kf[%d] vertices = %d, want %d", i, len(kf.Value.Vertices), wantNVerts[i])
		}
		if !kf.Value.Closed {
			t.Errorf("kf[%d].Closed = false, want true", i)
		}
	}

	// kf2 = triangle, bbox 0..200, normalized (0,0)(1,0)(0.5,1) →
	// denormalized (0,0)(200,0)(100,200). Spot-check denormalization.
	if len(kfs) == 3 && len(kfs[2].Value.Vertices) == 3 {
		got := kfs[2].Value.Vertices
		want := [][2]float64{{0, 0}, {200, 0}, {100, 200}}
		for i := range want {
			if !approxXY(got[i], want[i]) {
				t.Errorf("kf2 vertex %d = %v, want ~%v", i, got[i], want[i])
			}
		}
	}
}

func approxXY(a, b [2]float64) bool {
	const eps = 0.01
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	return dx < eps && dx > -eps && dy < eps && dy > -eps
}

// TestPathKeyframe_WriteRoundTrip is the Phase 2 lower counterpart to the
// hydrate test: build a shape layer with an animated PathNode (3 linear
// keyframes, distinct vertex counts), WriteAEP → FromReader, and confirm the
// path comes back Animated with the same times + geometry. Vertices are chosen
// so each frame's bbox spans its own verts → bbox-normalize/denormalize is
// lossless. (findLayerByName lives in shape_graph_roundtrip_test.go.)
func TestPathKeyframe_WriteRoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	sl, err := comp.NewShapeLayer("PathAnim")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	pn, err := sl.RootGroup().AddPath()
	if err != nil {
		t.Fatalf("AddPath: %v", err)
	}
	frames := []struct {
		t     float64
		verts [][2]float64
	}{
		{0, [][2]float64{{0, 0}, {40, 0}, {40, 40}, {0, 40}}},
		{1, [][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}}},
		{2, [][2]float64{{0, 0}, {200, 0}, {100, 200}}},
	}
	for _, f := range frames {
		bp := aep.BezierPath{
			Vertices:    f.verts,
			InTangents:  make([][2]float64, len(f.verts)),
			OutTangents: make([][2]float64, len(f.verts)),
			Closed:      true,
		}
		if err := pn.Path().AddKeyframeLinear(f.t, bp); err != nil {
			t.Fatalf("AddKeyframeLinear(%g): %v", f.t, err)
		}
	}
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	got, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(got.Compositions) == 0 {
		t.Fatal("round-trip: no compositions")
	}
	gotLayer := findLayerByName(got.Compositions[0], "PathAnim")
	if gotLayer == nil {
		t.Fatal("round-trip: PathAnim layer not found")
	}
	gsl := aep.WrapShapeLayer(gotLayer)
	var gotPath *aep.PathNode
	for _, child := range gsl.RootGroup().Children {
		if p, ok := child.(*aep.PathNode); ok {
			gotPath = p
			break
		}
	}
	if gotPath == nil {
		t.Fatal("round-trip: no PathNode in shape root group")
	}

	kfs := gotPath.Path().Keyframes()
	if len(kfs) != len(frames) {
		t.Fatalf("round-trip path keyframes = %d, want %d", len(kfs), len(frames))
	}
	for i, f := range frames {
		if kfs[i].Time != f.t {
			t.Errorf("kf[%d].Time = %g, want %g", i, kfs[i].Time, f.t)
		}
		if !kfs[i].Value.Closed {
			t.Errorf("kf[%d].Closed = false, want true", i)
		}
		if len(kfs[i].Value.Vertices) != len(f.verts) {
			t.Errorf("kf[%d] vertices = %d, want %d", i, len(kfs[i].Value.Vertices), len(f.verts))
			continue
		}
		for j := range f.verts {
			if !approxXY(kfs[i].Value.Vertices[j], f.verts[j]) {
				t.Errorf("kf[%d] vertex %d = %v, want ~%v", i, j, kfs[i].Value.Vertices[j], f.verts[j])
			}
		}
	}
}
