package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestVectorGroup_Empty(t *testing.T) {
	g := aep.NewVectorGroup()
	if got := len(g.Children); got != 0 {
		t.Fatalf("new VectorGroup children = %d, want 0", got)
	}
	if g.Transform == nil {
		t.Fatal("Transform PropertyGroup nil, want non-nil (group-level Transform placeholder)")
	}
}

func TestRectNode_Construct_Defaults(t *testing.T) {
	r := aep.NewRectNode()
	if r.Kind() != aep.ShapeKindRect {
		t.Fatalf("kind = %v, want ShapeKindRect", r.Kind())
	}
	// Defaults per RE-S4 (Phase 0 校准):
	v, _ := r.Size().StaticValue()
	if v != [2]float64{100, 100} {
		t.Fatalf("Rect default Size = %v, want [100,100] (RE-S4)", v)
	}
	pos, _ := r.Position().StaticValue()
	if pos != [2]float64{0, 0} {
		t.Fatalf("Rect default Position = %v, want [0,0]", pos)
	}
	rnd, _ := r.Roundness().StaticValue()
	if rnd != 0 {
		t.Fatalf("Rect default Roundness = %v, want 0", rnd)
	}
}

func TestRectNode_SetSize(t *testing.T) {
	r := aep.NewRectNode()
	if err := r.SetSize([2]float64{200, 100}); err != nil {
		t.Fatal(err)
	}
	v, _ := r.Size().StaticValue()
	if v != [2]float64{200, 100} {
		t.Fatalf("after SetSize: %v, want [200,100]", v)
	}
}

func TestEllipseNode_Construct_Defaults(t *testing.T) {
	e := aep.NewEllipseNode()
	if e.Kind() != aep.ShapeKindEllipse {
		t.Fatalf("kind = %v, want ShapeKindEllipse", e.Kind())
	}
	v, _ := e.Size().StaticValue()
	if v != [2]float64{100, 100} {
		t.Fatalf("Ellipse default Size = %v, want [100,100]", v)
	}
	pos, _ := e.Position().StaticValue()
	if pos != [2]float64{0, 0} {
		t.Fatalf("Ellipse default Position = %v, want [0,0]", pos)
	}
}

func TestFillNode_Defaults(t *testing.T) {
	f := aep.NewFillNode()
	if f.Kind() != aep.ShapeKindFill {
		t.Fatalf("kind = %v, want ShapeKindFill", f.Kind())
	}
	c, _ := f.Color().StaticValue()
	if c != [4]float64{1, 1, 1, 1} {
		t.Fatalf("Fill default Color = %v, want [1,1,1,1] white", c)
	}
	op, _ := f.Opacity().StaticValue()
	if op != 100 {
		t.Fatalf("Fill default Opacity = %v, want 100", op)
	}
}

func TestStrokeNode_Defaults(t *testing.T) {
	s := aep.NewStrokeNode()
	if s.Kind() != aep.ShapeKindStroke {
		t.Fatalf("kind = %v, want ShapeKindStroke", s.Kind())
	}
	c, _ := s.Color().StaticValue()
	if c != [4]float64{0, 0, 0, 1} {
		t.Fatalf("Stroke default Color = %v, want [0,0,0,1] black", c)
	}
	w, _ := s.Width().StaticValue()
	if w != 2 {
		t.Fatalf("Stroke default Width = %v, want 2", w)
	}
	op, _ := s.Opacity().StaticValue()
	if op != 100 {
		t.Fatalf("Stroke default Opacity = %v, want 100", op)
	}
}

func TestPathNode_SetVertices(t *testing.T) {
	p := aep.NewPathNode()
	if err := p.SetVertices([][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}); err != nil {
		t.Fatal(err)
	}
	bz, _ := p.Path().StaticValue()
	if len(bz.Vertices) != 4 {
		t.Fatalf("Vertices count = %d, want 4", len(bz.Vertices))
	}
	if !bz.Closed {
		t.Fatal("PathNode default Closed = false, want true (RE-S5b)")
	}
}

func TestPathNode_SetVertices_RejectsEmpty(t *testing.T) {
	p := aep.NewPathNode()
	if err := p.SetVertices(nil); err == nil {
		t.Fatal("nil vertices should error (per RE-S8 min check)")
	}
	if err := p.SetVertices([][2]float64{{0, 0}}); err == nil {
		t.Fatal("single vertex should error (per RE-S8 min check)")
	}
}

func TestVectorGroup_RenderOrder_AppendsTop(t *testing.T) {
	// Children[0] = bottom; len-1 = top. (Spec §3.2)
	g := aep.NewVectorGroup()
	g.Children = append(g.Children, aep.NewRectNode())
	g.Children = append(g.Children, aep.NewFillNode())
	if len(g.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(g.Children))
	}
	if g.Children[0].Kind() != aep.ShapeKindRect {
		t.Fatalf("Children[0] kind = %v, want Rect (bottom)", g.Children[0].Kind())
	}
	if g.Children[1].Kind() != aep.ShapeKindFill {
		t.Fatalf("Children[1] kind = %v, want Fill (top)", g.Children[1].Kind())
	}
}
