package serializer

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/scene"
)

func TestGradientStroke_RoundTrip(t *testing.T) {
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
		t.Fatal(err)
	}
	got := hydrateGradientStrokeNode(body, nil)
	if got == nil {
		t.Fatal("hydrateGradientStrokeNode returned nil")
	}
	if len(got.Gradient().ColorStops) != 3 {
		t.Fatalf("round-trip color stops = %d, want 3", len(got.Gradient().ColorStops))
	}
}
