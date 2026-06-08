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
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("Grad")
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
