package scene

import "testing"

func TestGradientStrokeNode_Defaults(t *testing.T) {
	n := NewGradientStrokeNode()
	if n.Kind() != ShapeKindGradientStroke {
		t.Fatalf("Kind = %v, want ShapeKindGradientStroke", n.Kind())
	}
	if got := len(n.Gradient().ColorStops); got != 2 {
		t.Fatalf("default color stops = %d, want 2", got)
	}
}

func TestGradientStrokeNode_SetColorStops_Validation(t *testing.T) {
	n := NewGradientStrokeNode()
	if err := n.SetColorStops([]GradientColorStop{{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}}}); err == nil {
		t.Fatal("want error for <2 stops, got nil")
	}
	ok := []GradientColorStop{
		{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
		{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
	}
	if err := n.SetColorStops(ok); err != nil {
		t.Fatalf("SetColorStops valid: %v", err)
	}
	if n.Gradient().ColorStops[1].Color != [3]float64{0, 0, 1} {
		t.Fatal("color stop not applied")
	}
}

func TestAddGradientStroke(t *testing.T) {
	g := NewVectorGroup()
	n, err := g.AddGradientStroke()
	if err != nil {
		t.Fatal(err)
	}
	if g.Children[len(g.Children)-1] != n {
		t.Fatal("AddGradientStroke did not append node")
	}
}
