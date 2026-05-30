// internal/aep/shape_stroke_linecap_test.go
//
// StrokeNode Line Cap / Line Join / Miter Limit: model defaults + setter
// validation + WriteAEP→FromReader roundtrip persistence. RE findings:
// incident-reports/stroke-line-cap-join-miter-re.md (OneD float64-BE cdat[0:8];
// Cap 1=Butt/2=Round/3=Proj, Join 1=Miter/2=Round/3=Bevel, Miter scalar def 4).
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestStrokeNode_LineCapJoinMiter_DefaultsAndSetters(t *testing.T) {
	s := aep.NewStrokeNode()
	if s.LineCap() != aep.StrokeLineCapButt {
		t.Errorf("default LineCap = %d, want Butt(1)", s.LineCap())
	}
	if s.LineJoin() != aep.StrokeLineJoinMiter {
		t.Errorf("default LineJoin = %d, want Miter(1)", s.LineJoin())
	}
	if s.MiterLimit() != 4 {
		t.Errorf("default MiterLimit = %g, want 4", s.MiterLimit())
	}

	if err := s.SetLineCap(aep.StrokeLineCapRound); err != nil {
		t.Fatalf("SetLineCap(Round): %v", err)
	}
	if err := s.SetLineJoin(aep.StrokeLineJoinBevel); err != nil {
		t.Fatalf("SetLineJoin(Bevel): %v", err)
	}
	if err := s.SetMiterLimit(10); err != nil {
		t.Fatalf("SetMiterLimit(10): %v", err)
	}
	if s.LineCap() != aep.StrokeLineCapRound || s.LineJoin() != aep.StrokeLineJoinBevel || s.MiterLimit() != 10 {
		t.Errorf("after sets: cap=%d join=%d miter=%g", s.LineCap(), s.LineJoin(), s.MiterLimit())
	}

	// Out-of-range rejections.
	if err := s.SetLineCap(0); err == nil {
		t.Error("SetLineCap(0) should error")
	}
	if err := s.SetLineCap(4); err == nil {
		t.Error("SetLineCap(4) should error")
	}
	if err := s.SetLineJoin(4); err == nil {
		t.Error("SetLineJoin(4) should error")
	}
	if err := s.SetMiterLimit(0.5); err == nil {
		t.Error("SetMiterLimit(0.5) should error")
	}
}

func TestV2_2_StrokeLineCapJoinMiter_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("StrokeCapLayer")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	_, _ = l.RootGroup().AddRect()
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.SetWidth(6)
	_ = stroke.SetLineCap(aep.StrokeLineCapProjecting)
	_ = stroke.SetLineJoin(aep.StrokeLineJoinRound)
	_ = stroke.SetMiterLimit(12)

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	rl := findLayerByName(re.Compositions[0], "StrokeCapLayer")
	if rl == nil {
		t.Fatal("missing StrokeCapLayer")
	}
	sl := aep.WrapShapeLayer(rl)
	var reStroke *aep.StrokeNode
	for _, ch := range sl.RootGroup().Children {
		if s, ok := ch.(*aep.StrokeNode); ok {
			reStroke = s
		}
	}
	if reStroke == nil {
		t.Fatal("stroke node not recovered post-roundtrip")
	}
	if reStroke.LineCap() != aep.StrokeLineCapProjecting {
		t.Errorf("LineCap post-roundtrip = %d, want Projecting(3)", reStroke.LineCap())
	}
	if reStroke.LineJoin() != aep.StrokeLineJoinRound {
		t.Errorf("LineJoin post-roundtrip = %d, want Round(2)", reStroke.LineJoin())
	}
	if reStroke.MiterLimit() != 12 {
		t.Errorf("MiterLimit post-roundtrip = %g, want 12", reStroke.MiterLimit())
	}
}
