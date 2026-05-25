package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func TestNewShapeLayer_BasicCreation(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.NewShapeLayer("S1")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("nil ShapeLayer")
	}
	if s.Name != "S1" {
		t.Fatalf("Name = %q, want S1", s.Name)
	}
	if s.Type != aep.LayerTypeShape {
		t.Fatalf("Type = %v, want LayerTypeShape", s.Type)
	}
	if len(c.Layers) != 1 || c.Layers[0] != s.Layer {
		t.Fatalf("comp.Layers not updated correctly: len=%d", len(c.Layers))
	}
}

// TestNewShapeLayer_EmitsEwstSibling — iter-5b: every Layr at Item level
// must be followed by an empty LIST(Ewst) sibling. AE 2025 silently drops
// user Layr from comp.layers when this is absent (variant #2 bisect proof).
func TestNewShapeLayer_EmitsEwstSibling(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.NewShapeLayer("S1"); err != nil {
		t.Fatal(err)
	}
	// Find the user Layr in itemList and assert the next sibling is Ewst(0).
	il := aep.CompItemListForTest(c)
	if il == nil {
		t.Fatal("comp itemList nil")
	}
	var layrIdx int = -1
	for i, ch := range il.Children {
		if ch.IsList() && ch.FormType == rifx.IDLayr {
			layrIdx = i
			break
		}
	}
	if layrIdx < 0 {
		t.Fatal("no user Layr in itemList")
	}
	if layrIdx+1 >= len(il.Children) {
		t.Fatal("Layr is last child — no room for Ewst sibling")
	}
	next := il.Children[layrIdx+1]
	if !next.IsList() || next.FormType != rifx.IDEwst {
		t.Fatalf("Layr's next sibling = %q (formType=%q), want LIST(Ewst)", next.ID, next.FormType)
	}
	if len(next.Children) != 0 {
		t.Errorf("Ewst children = %d, want 0", len(next.Children))
	}
}

func TestNewShapeLayer_EmptyName_Error(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.NewShapeLayer(""); err == nil {
		t.Fatal("empty name should error")
	}
	if len(c.Layers) != 0 {
		t.Fatalf("comp polluted on failure: %d layers", len(c.Layers))
	}
}
