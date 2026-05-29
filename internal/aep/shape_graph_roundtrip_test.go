// internal/aep/shape_graph_roundtrip_test.go
//
// V2.2 Phase 4 Task 4.2 — canonical 3-layer roundtrip. Drives every
// node kind (Rect/Ellipse/Path/Fill/Stroke), both stream modes
// (static + animated), and the Layer-level Transform keyframe path
// through WriteAEP → FromReader.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestV2_2_CanonicalShapeGraph_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// ShapeLayer A: animated streams
	a, err := comp.NewShapeLayer("A_RectFill_Animated")
	if err != nil {
		t.Fatalf("NewShapeLayer A: %v", err)
	}
	rectA, _ := a.RootGroup().AddRect()
	if err := rectA.Size().AddKeyframeLinear(0, [2]float64{50, 50}); err != nil {
		t.Fatalf("rectA kf0: %v", err)
	}
	if err := rectA.Size().AddKeyframeLinear(2, [2]float64{300, 200}); err != nil {
		t.Fatalf("rectA kf1: %v", err)
	}
	fillA, _ := a.RootGroup().AddFill()
	if err := fillA.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1}); err != nil {
		t.Fatalf("fillA kf0: %v", err)
	}
	if err := fillA.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1}); err != nil {
		t.Fatalf("fillA kf1: %v", err)
	}
	if err := a.Position().AddKeyframeLinear(0, [2]float64{0, 0}); err != nil {
		t.Fatalf("aPos kf0: %v", err)
	}
	if err := a.Position().AddKeyframeLinear(2, [2]float64{500, 300}); err != nil {
		t.Fatalf("aPos kf1: %v", err)
	}

	// ShapeLayer B: static
	b, _ := comp.NewShapeLayer("B_EllipseStroke_Static")
	ellB, _ := b.RootGroup().AddEllipse()
	_ = ellB.SetSize([2]float64{150, 150})
	strokeB, _ := b.RootGroup().AddStroke()
	_ = strokeB.SetColor([4]float64{0, 0, 1, 1})
	_ = strokeB.SetWidth(5)

	// ShapeLayer C: static path + fill + stroke
	cl, _ := comp.NewShapeLayer("C_PathFillStroke_Static")
	pathC, _ := cl.RootGroup().AddPath()
	_ = pathC.SetVertices([][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	_ = pathC.SetClosed(true)
	fillC, _ := cl.RootGroup().AddFill()
	_ = fillC.SetColor([4]float64{0, 1, 0, 1})
	strokeC, _ := cl.RootGroup().AddStroke()
	_ = strokeC.SetWidth(2)

	// Roundtrip
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	reMain := re.Compositions[0]
	if len(reMain.Layers) != 3 {
		t.Fatalf("layers = %d, want 3", len(reMain.Layers))
	}

	// Layer A (animated)
	la := findLayerByName(reMain, "A_RectFill_Animated")
	if la == nil {
		t.Fatal("missing A_RectFill_Animated")
	}
	sa := aep.WrapShapeLayer(la)
	if len(sa.RootGroup().Children) != 2 {
		t.Fatalf("A contents = %d, want 2", len(sa.RootGroup().Children))
	}
	reRectA, ok := sa.RootGroup().Children[0].(*aep.RectNode)
	if !ok {
		t.Fatalf("A[0] type = %T, want *RectNode", sa.RootGroup().Children[0])
	}
	// V2.2.1: Rect Size keyframes ARE persisted (cdat→LIST(list) inject;
	// non-spatial Vec2 layout byte-identical to AE). Roundtrip recovers the
	// 2 linear keyframes.
	_, sizeIsStatic := reRectA.Size().StaticValue()
	if sizeIsStatic {
		t.Errorf("A Rect Size should be animated post-roundtrip (V2.2.1 keyframe persistence)")
	}
	kfs := reRectA.Size().Keyframes()
	if len(kfs) != 2 {
		t.Fatalf("A Rect Size keyframes = %d, want 2", len(kfs))
	}
	if kfs[0].Value != [2]float64{50, 50} || kfs[1].Value != [2]float64{300, 200} {
		t.Errorf("A Rect Size kf values = %v / %v, want [50,50] / [300,200]", kfs[0].Value, kfs[1].Value)
	}

	// V2.2.1 Path B: Layr-level Position keyframes ARE persisted to disk —
	// injected into the combined "ADBE Position" stream (bpk-128 spatial dim-3,
	// Z=0) in the embedded transform-group template. Roundtrip recovers the 2
	// linear keyframes.
	if _, posIsStatic := sa.Position().StaticValue(); posIsStatic {
		t.Errorf("A Position should be animated post-roundtrip (V2.2.1 Layr Position keyframe persistence)")
	}
	posKfs := sa.Position().Keyframes()
	if len(posKfs) != 2 {
		t.Fatalf("A Position keyframes = %d, want 2", len(posKfs))
	}
	if posKfs[0].Value != [2]float64{0, 0} || posKfs[1].Value != [2]float64{500, 300} {
		t.Errorf("A Position kf values = %v / %v, want [0,0] / [500,300]", posKfs[0].Value, posKfs[1].Value)
	}

	// Layer B (static)
	lb := findLayerByName(reMain, "B_EllipseStroke_Static")
	if lb == nil {
		t.Fatal("missing B_EllipseStroke_Static")
	}
	sb := aep.WrapShapeLayer(lb)
	if len(sb.RootGroup().Children) != 2 {
		t.Fatalf("B contents = %d, want 2", len(sb.RootGroup().Children))
	}
	reEllB, ok := sb.RootGroup().Children[0].(*aep.EllipseNode)
	if !ok {
		t.Fatalf("B[0] type = %T, want *EllipseNode", sb.RootGroup().Children[0])
	}
	reEllSize, isStatic := reEllB.Size().StaticValue()
	if !isStatic {
		t.Fatal("B Ellipse Size not static")
	}
	if reEllSize != [2]float64{150, 150} {
		t.Fatalf("B Ellipse Size = %v, want [150,150]", reEllSize)
	}

	// Layer C (static)
	lc := findLayerByName(reMain, "C_PathFillStroke_Static")
	if lc == nil {
		t.Fatal("missing C_PathFillStroke_Static")
	}
	sc := aep.WrapShapeLayer(lc)
	rePathC, ok := sc.RootGroup().Children[0].(*aep.PathNode)
	if !ok {
		t.Fatalf("C[0] type = %T, want *PathNode", sc.RootGroup().Children[0])
	}
	rePath, _ := rePathC.Path().StaticValue()
	if len(rePath.Vertices) != 4 {
		t.Fatalf("C Path vertices = %d, want 4", len(rePath.Vertices))
	}
	if !rePath.Closed {
		t.Fatal("C Path Closed = false, want true")
	}
}

func findLayerByName(c *aep.Composition, name string) *aep.Layer {
	for _, l := range c.Layers {
		if l.Name == name {
			return l
		}
	}
	return nil
}
