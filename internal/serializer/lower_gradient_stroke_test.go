package serializer

import (
	"testing"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/scene"
)

func TestLowerGradientStrokeNode_HasGradColors(t *testing.T) {
	n := scene.NewGradientStrokeNode()
	if err := n.SetColorStops([]scene.GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}); err != nil {
		t.Fatal(err)
	}
	body, err := LowerShapeNodeForTest(n)
	if err != nil {
		t.Fatalf("lower: %v", err)
	}
	xml := findGradientStopsXML(body, "ADBE Vector Grad Colors")
	if xml == "" {
		t.Fatal("lowered G-Stroke body has no Grad Colors XML")
	}
	g := codec.ParseGradientXML(xml)
	if g == nil {
		t.Fatal("ParseGradientXML returned nil")
	}
	if len(g.ColorStops) != 3 {
		t.Fatalf("lowered stops = %d, want 3", len(g.ColorStops))
	}
}
