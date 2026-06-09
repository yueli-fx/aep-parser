package serializer

import (
	"testing"

	"github.com/example/aep-parser/internal/rifx"
)

func TestNewShapeLayer_BasicCreation(t *testing.T) {
	p := NewProject()
	c, err := NewComposition(p, "Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewShapeLayer(c, "S1")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("nil ShapeLayer")
	}
	if s.Name != "S1" {
		t.Fatalf("Name = %q, want S1", s.Name)
	}
	if s.Type != LayerTypeShape {
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
	p := NewProject()
	c, err := NewComposition(p, "Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewShapeLayer(c, "S1"); err != nil {
		t.Fatal(err)
	}
	// Find the user Layr in itemList and assert the next sibling is Ewst(0).
	il := CompItemListForTest(c)
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
	p := NewProject()
	c, err := NewComposition(p, "Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewShapeLayer(c, ""); err == nil {
		t.Fatal("empty name should error")
	}
	if len(c.Layers) != 0 {
		t.Fatalf("comp polluted on failure: %d layers", len(c.Layers))
	}
}

// TestNewShapeLayer_MultiLayer_EmitsLayerSiblings — every user Layr's
// serialized unit is Layr LIST + empty Ewst + two fvdv/fiop/ftts/foac/fiac/
// fipc/fifl groups. The Ewst alone passes a single user layer, but AE
// silent-drops every layer past the first without the fvdv… siblings
// (RE: multi-layer-silent-drop). Two NewShapeLayer calls must each emit them.
func TestNewShapeLayer_MultiLayer_EmitsLayerSiblings(t *testing.T) {
	p := NewProject(TargetAE2025)
	c, err := NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewShapeLayer(c, "L1"); err != nil {
		t.Fatal(err)
	}
	if _, err := NewShapeLayer(c, "L2"); err != nil {
		t.Fatal(err)
	}

	fv := rifx.ChunkID{'f', 'v', 'd', 'v'}
	fi := rifx.ChunkID{'f', 'i', 'o', 'p'}
	ft := rifx.ChunkID{'f', 't', 't', 's'}
	fo := rifx.ChunkID{'f', 'o', 'a', 'c'}
	fa := rifx.ChunkID{'f', 'i', 'a', 'c'}
	fp := rifx.ChunkID{'f', 'i', 'p', 'c'}
	ff := rifx.ChunkID{'f', 'i', 'f', 'l'}
	want := []rifx.ChunkID{rifx.IDEwst, fv, fi, ft, fo, fa, fp, ff, fv, fi, ft, fo, fa, fp, ff}

	il := CompItemListForTest(c)
	userLayrs := 0
	for i, ch := range il.Children {
		if !ch.IsList() || ch.FormType != rifx.IDLayr {
			continue
		}
		userLayrs++
		for j, w := range want {
			if i+1+j >= len(il.Children) {
				t.Fatalf("user Layr #%d: ran out of children expecting sibling %d", userLayrs, j)
			}
			sib := il.Children[i+1+j]
			got := sib.ID
			if sib.IsList() {
				got = sib.FormType
			}
			if got != w {
				t.Errorf("user Layr #%d sibling[%d] = %q, want %q", userLayrs, j, got, w)
			}
		}
	}
	if userLayrs != 2 {
		t.Fatalf("user Layr count = %d, want 2", userLayrs)
	}
}
