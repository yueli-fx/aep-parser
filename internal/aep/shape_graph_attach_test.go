package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestVectorGroup_AddRect_Appends(t *testing.T) {
	g := aep.NewVectorGroup()
	r, err := g.AddRect()
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("nil RectNode")
	}
	if len(g.Children) != 1 || g.Children[0] != r {
		t.Fatalf("Children not updated: len=%d", len(g.Children))
	}
}

func TestVectorGroup_RenderOrder(t *testing.T) {
	g := aep.NewVectorGroup()
	rect, _ := g.AddRect()
	fill, _ := g.AddFill()
	if g.Children[0] != rect {
		t.Fatal("first add should be Children[0] (bottom)")
	}
	if g.Children[1] != fill {
		t.Fatal("second add should be Children[1] (top)")
	}
}

func TestVectorGroup_AllAddMethods(t *testing.T) {
	g := aep.NewVectorGroup()
	if _, err := g.AddEllipse(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddPath(); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AddStroke(); err != nil {
		t.Fatal(err)
	}
	if len(g.Children) != 3 {
		t.Fatalf("Children = %d, want 3", len(g.Children))
	}
	if g.Children[0].Kind() != aep.ShapeKindEllipse {
		t.Fatalf("Children[0] kind = %v, want Ellipse", g.Children[0].Kind())
	}
	if g.Children[1].Kind() != aep.ShapeKindPath {
		t.Fatalf("Children[1] kind = %v, want Path", g.Children[1].Kind())
	}
	if g.Children[2].Kind() != aep.ShapeKindStroke {
		t.Fatalf("Children[2] kind = %v, want Stroke", g.Children[2].Kind())
	}
}
