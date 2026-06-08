// internal/aep/shape_enums_test.go
//
// Shape-node enum setters: Rect/Ellipse Direction, Fill Blend Mode / Composite
// Order / Fill Rule, Stroke Blend Mode / Composite Order. Model defaults +
// setter validation + WriteAEP->FromReader roundtrip. All OneD float64-BE at
// cdat[0:8]; RE: incident-reports/stroke-line-cap-join-miter-re.md (same run).
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestShapeEnums_DefaultsAndValidation(t *testing.T) {
	r := aep.NewRectNode()
	if r.Direction() != aep.ShapeDirectionNormal {
		t.Errorf("rect default Direction = %d, want Normal(1)", r.Direction())
	}
	if err := r.SetDirection(aep.ShapeDirectionReversed); err != nil {
		t.Fatalf("SetDirection(Reversed): %v", err)
	}
	if err := r.SetDirection(2); err == nil {
		t.Error("SetDirection(2) should error (only 1/3 valid)")
	}

	f := aep.NewFillNode()
	if f.BlendMode() != aep.ShapeBlendModeNormal || f.CompositeOrder() != aep.ShapeCompositeOrderAbovePrevious || f.FillRule() != aep.FillRuleNonzeroWinding {
		t.Errorf("fill defaults wrong: bm=%d co=%d fr=%d", f.BlendMode(), f.CompositeOrder(), f.FillRule())
	}
	if err := f.SetBlendMode(0); err == nil {
		t.Error("SetBlendMode(0) should error")
	}
	if err := f.SetCompositeOrder(3); err == nil {
		t.Error("SetCompositeOrder(3) should error")
	}
	if err := f.SetFillRule(3); err == nil {
		t.Error("SetFillRule(3) should error")
	}
}

func TestV2_2_ShapeEnums_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Enums")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetDirection(aep.ShapeDirectionReversed)
	ell, _ := l.RootGroup().AddEllipse()
	_ = ell.SetDirection(aep.ShapeDirectionReversed)
	fill, _ := l.RootGroup().AddFill()
	_ = fill.SetBlendMode(3)
	_ = fill.SetCompositeOrder(aep.ShapeCompositeOrderBelowPrevious)
	_ = fill.SetFillRule(aep.FillRuleEvenOdd)
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.SetBlendMode(4)
	_ = stroke.SetCompositeOrder(aep.ShapeCompositeOrderBelowPrevious)

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	sl := aep.WrapShapeLayer(findLayerByName(re.Compositions[0], "Enums"))

	var rr *aep.RectNode
	var ee *aep.EllipseNode
	var ff *aep.FillNode
	var ss *aep.StrokeNode
	for _, ch := range sl.RootGroup().Children {
		switch n := ch.(type) {
		case *aep.RectNode:
			rr = n
		case *aep.EllipseNode:
			ee = n
		case *aep.FillNode:
			ff = n
		case *aep.StrokeNode:
			ss = n
		}
	}
	if rr == nil || ee == nil || ff == nil || ss == nil {
		t.Fatalf("nodes not all recovered: rect=%v ell=%v fill=%v stroke=%v", rr != nil, ee != nil, ff != nil, ss != nil)
	}
	if rr.Direction() != aep.ShapeDirectionReversed {
		t.Errorf("rect Direction = %d, want Reversed(3)", rr.Direction())
	}
	if ee.Direction() != aep.ShapeDirectionReversed {
		t.Errorf("ellipse Direction = %d, want Reversed(3)", ee.Direction())
	}
	if ff.BlendMode() != 3 {
		t.Errorf("fill BlendMode = %d, want 3", ff.BlendMode())
	}
	if ff.CompositeOrder() != aep.ShapeCompositeOrderBelowPrevious {
		t.Errorf("fill CompositeOrder = %d, want BelowPrevious(2)", ff.CompositeOrder())
	}
	if ff.FillRule() != aep.FillRuleEvenOdd {
		t.Errorf("fill FillRule = %d, want EvenOdd(2)", ff.FillRule())
	}
	if ss.BlendMode() != 4 {
		t.Errorf("stroke BlendMode = %d, want 4", ss.BlendMode())
	}
	if ss.CompositeOrder() != aep.ShapeCompositeOrderBelowPrevious {
		t.Errorf("stroke CompositeOrder = %d, want BelowPrevious(2)", ss.CompositeOrder())
	}
}
