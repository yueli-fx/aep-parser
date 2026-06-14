// internal/aep/shape_gradient_test.go
//
// Gradient fill: XML encoder↔decoder round-trip, model defaults/validation,
// and WriteAEP→FromReader roundtrip. The serializer overwrites the GCst→GCky→
// Utf8 prop.map stops XML (length-variable; rifx recomputes LIST sizes). V2.2
// models color + alpha stops; Grad Type / Start Pt / End Pt are elided in the
// template (default linear ramp). RE: incident-reports/gradient-fill-write-re.md.
package aep_test

import (
	"bytes"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/codec"
)

func TestGradientXML_EncodeDecodeRoundtrip(t *testing.T) {
	g := &codec.Gradient{
		Version: "4",
		ColorStops: []codec.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
		},
		AlphaStops: []codec.GradientAlphaStop{
			{Offset: 0, Midpoint: 0.5, Alpha: 1},
			{Offset: 1, Midpoint: 0.5, Alpha: 0.5},
		},
	}
	xml := codec.EncodeGradientXML(g)
	back := codec.ParseGradientXML(xml)
	if back == nil {
		t.Fatalf("codec.ParseGradientXML returned nil for:\n%s", xml)
	}
	if len(back.ColorStops) != 3 {
		t.Fatalf("color stops = %d, want 3", len(back.ColorStops))
	}
	if len(back.AlphaStops) != 2 {
		t.Fatalf("alpha stops = %d, want 2", len(back.AlphaStops))
	}
	for i, want := range g.ColorStops {
		got := back.ColorStops[i]
		if math.Abs(got.Offset-want.Offset) > 1e-9 || got.Color != want.Color {
			t.Errorf("color stop %d = %+v, want %+v", i, got, want)
		}
	}
	if math.Abs(back.AlphaStops[1].Alpha-0.5) > 1e-9 {
		t.Errorf("alpha stop 1 = %g, want 0.5", back.AlphaStops[1].Alpha)
	}
}

func TestGradientFill_Defaults(t *testing.T) {
	n := aep.NewGradientFillNode()
	if n.Kind() != aep.ShapeKindGradientFill {
		t.Fatalf("kind = %v, want ShapeKindGradientFill", n.Kind())
	}
	g := n.Gradient()
	if len(g.ColorStops) != 2 || len(g.AlphaStops) != 2 {
		t.Fatalf("default stops = %d color / %d alpha, want 2/2", len(g.ColorStops), len(g.AlphaStops))
	}
	if g.ColorStops[0].Color != [3]float64{0, 0, 0} || g.ColorStops[1].Color != [3]float64{1, 1, 1} {
		t.Error("default gradient must be black→white")
	}
	// Validation.
	if err := n.SetColorStops([]codec.GradientColorStop{{Offset: 0}}); err == nil {
		t.Error("SetColorStops with 1 stop must reject (need ≥2)")
	}
	if err := n.SetColorStops([]codec.GradientColorStop{
		{Offset: 0, Color: [3]float64{2, 0, 0}}, {Offset: 1},
	}); err == nil {
		t.Error("SetColorStops with out-of-range color must reject")
	}
	if err := n.SetAlphaStops([]codec.GradientAlphaStop{
		{Offset: 0, Alpha: 1}, {Offset: 1, Alpha: 1},
	}); err != nil {
		t.Errorf("valid SetAlphaStops rejected: %v", err)
	}
}

func TestV2_2_GradientFill_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Grad")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	gf, _ := l.RootGroup().AddGradientFill()
	// Distinct 3-color stops so the roundtrip proves the XML overwrite, not
	// template passthrough (the template carries py-aep's 2 stops).
	if err := gf.SetColorStops([]codec.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}); err != nil {
		t.Fatalf("SetColorStops: %v", err)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rg := recoverGradientFillNode(t, re.Compositions[0], "Grad")
	g := rg.Gradient()
	if len(g.ColorStops) != 3 {
		t.Fatalf("roundtrip color stops = %d, want 3", len(g.ColorStops))
	}
	want := [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	for i, w := range want {
		if g.ColorStops[i].Color != w {
			t.Errorf("color stop %d = %v, want %v", i, g.ColorStops[i].Color, w)
		}
	}
}

func recoverGradientFillNode(t *testing.T, comp *aep.Composition, layerName string) *aep.GradientFillNode {
	t.Helper()
	sl := aep.WrapShapeLayer(findLayerByName(comp, layerName))
	for _, ch := range sl.RootGroup().Children {
		if n, ok := ch.(*aep.GradientFillNode); ok {
			return n
		}
	}
	t.Fatalf("gradient fill node not recovered after roundtrip (layer %q)", layerName)
	return nil
}

func recoverGradientStrokeNode(t *testing.T, comp *aep.Composition, layerName string) *aep.GradientStrokeNode {
	t.Helper()
	sl := aep.WrapShapeLayer(findLayerByName(comp, layerName))
	for _, ch := range sl.RootGroup().Children {
		if n, ok := ch.(*aep.GradientStrokeNode); ok {
			return n
		}
	}
	t.Fatalf("gradient stroke node not recovered after roundtrip (layer %q)", layerName)
	return nil
}

// TestV2_2_GradientGeometry_Roundtrip proves the hydrate path reads the ramp
// geometry (Grad Type / Start Pt / End Pt / HiLite Length·Angle) back into the
// typed getters — not just the color stops. Builds a radial G-Fill and G-Stroke
// with distinct non-default geometry, writes, re-parses, and asserts every
// geometry getter reflects the written value.
func TestV2_2_GradientGeometry_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Grad")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})

	gf, _ := l.RootGroup().AddGradientFill()
	must2(t, gf.SetGradientType(aep.GradientRadial))
	must2(t, gf.SetStartPoint([2]float64{-30, 40}))
	must2(t, gf.SetEndPoint([2]float64{90, -10}))
	must2(t, gf.SetHighlightLength(55))
	must2(t, gf.SetHighlightAngle(25))

	gs, _ := l.RootGroup().AddGradientStroke()
	must2(t, gs.SetGradientType(aep.GradientRadial))
	must2(t, gs.SetStartPoint([2]float64{12, -8}))
	must2(t, gs.SetEndPoint([2]float64{70, 60}))
	must2(t, gs.SetHighlightLength(-40))
	must2(t, gs.SetHighlightAngle(110))

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	rf := recoverGradientFillNode(t, re.Compositions[0], "Grad")
	if rf.GradientType() != aep.GradientRadial {
		t.Errorf("G-Fill type = %d, want radial(2)", rf.GradientType())
	}
	if rf.StartPoint() != [2]float64{-30, 40} {
		t.Errorf("G-Fill StartPoint = %v, want [-30 40]", rf.StartPoint())
	}
	if rf.EndPoint() != [2]float64{90, -10} {
		t.Errorf("G-Fill EndPoint = %v, want [90 -10]", rf.EndPoint())
	}
	if rf.HighlightLength() != 55 || rf.HighlightAngle() != 25 {
		t.Errorf("G-Fill HiLite = (%g,%g), want (55,25)", rf.HighlightLength(), rf.HighlightAngle())
	}

	rs := recoverGradientStrokeNode(t, re.Compositions[0], "Grad")
	if rs.GradientType() != aep.GradientRadial {
		t.Errorf("G-Stroke type = %d, want radial(2)", rs.GradientType())
	}
	if rs.StartPoint() != [2]float64{12, -8} {
		t.Errorf("G-Stroke StartPoint = %v, want [12 -8]", rs.StartPoint())
	}
	if rs.EndPoint() != [2]float64{70, 60} {
		t.Errorf("G-Stroke EndPoint = %v, want [70 60]", rs.EndPoint())
	}
	if rs.HighlightLength() != -40 || rs.HighlightAngle() != 110 {
		t.Errorf("G-Stroke HiLite = (%g,%g), want (-40,110)", rs.HighlightLength(), rs.HighlightAngle())
	}
}

func must2(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
