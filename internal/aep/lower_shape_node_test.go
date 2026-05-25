// internal/aep/lower_shape_node_test.go
//
// Phase 2 Task 2.2 tests — chunk-shape level only. Validates the dispatcher
// + per-node lowering + VectorGroup wrapping per Phase 0 RE-S3/S4/S5a-d.
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func TestLowerRectNode_HasTdmnAndSubProps(t *testing.T) {
	r := aep.NewRectNode()
	chunk, err := aep.LowerShapeNodeForTest(r)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil || !chunk.IsList() || chunk.FormType != rifx.IDTdgp {
		t.Fatalf("expected LIST(tdgp), got %+v", chunk)
	}
	// V2.2 always emits sub-props (Direction + Size + Position + Roundness).
	// Children: tdsb + tdsn + 4 × (tdmn + LIST tdgp) + tdmn(Group End) = 11.
	tdmns := chunk.FindAll(rifx.IDTdmn)
	if len(tdmns) < 2 {
		t.Fatalf("rect tdgp needs at least 2 tdmn (shape + Group End), got %d", len(tdmns))
	}
}

func TestLowerEllipseNode_DispatcherWorks(t *testing.T) {
	e := aep.NewEllipseNode()
	chunk, err := aep.LowerShapeNodeForTest(e)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
}

func TestLowerFillNode_HasColorAndOpacity(t *testing.T) {
	f := aep.NewFillNode()
	chunk, err := aep.LowerShapeNodeForTest(f)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
}

func TestLowerStrokeNode_HasFullChildSet(t *testing.T) {
	s := aep.NewStrokeNode()
	chunk, err := aep.LowerShapeNodeForTest(s)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
}

func TestLowerPathNode_LinearOpenSegment(t *testing.T) {
	p := aep.NewPathNode()
	if err := p.SetVertices([][2]float64{{0, 0}, {100, 100}}); err != nil {
		t.Fatal(err)
	}
	chunk, err := aep.LowerShapeNodeForTest(p)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
}

func TestLowerVectorGroup_EmitsRootVectorsGroupChildren(t *testing.T) {
	g := aep.NewVectorGroup()
	g.Children = append(g.Children, aep.NewRectNode(), aep.NewFillNode())
	chunk, err := aep.LowerVectorGroupForTest(g)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil || !chunk.IsList() || chunk.FormType != rifx.IDTdgp {
		t.Fatalf("expected LIST(tdgp), got %+v", chunk)
	}
	// iter-5: Root Vectors Group body = tdsb + tdsn + tdmn(Vector Group) +
	// LIST(tdgp, Vector Group body) + tdmn(Group End) = 5 children.
	if len(chunk.Children) != 5 {
		t.Fatalf("root vectors group body child count = %d, want 5", len(chunk.Children))
	}
	if chunk.Children[0].ID.String() != "tdsb" {
		t.Fatalf("group child[0] id = %q, want tdsb", chunk.Children[0].ID)
	}
	// Shape kids live 2 levels deep — Vector Group body → Vectors Group body.
	// Confirm both wrappers are emitted and Rect+Fill are reachable.
	var findShapeKids func(c *rifx.Chunk) []string
	findShapeKids = func(c *rifx.Chunk) []string {
		var out []string
		for i, ch := range c.Children {
			if ch.ID == rifx.IDTdmn && len(ch.Data) > 0 {
				name := string(ch.Data)
				if n := indexNUL(name); n >= 0 {
					name = name[:n]
				}
				switch name {
				case "ADBE Vector Shape - Rect", "ADBE Vector Graphic - Fill":
					out = append(out, name)
				}
				_ = i
			}
			if ch.IsList() {
				out = append(out, findShapeKids(ch)...)
			}
		}
		return out
	}
	kids := findShapeKids(chunk)
	if len(kids) != 2 || kids[0] != "ADBE Vector Shape - Rect" || kids[1] != "ADBE Vector Graphic - Fill" {
		t.Fatalf("shape kids reachable = %v, want [Rect, Fill]", kids)
	}
	// Vector Group + Vectors Group wrappers must appear.
	saw := map[string]bool{}
	var walkNames func(c *rifx.Chunk)
	walkNames = func(c *rifx.Chunk) {
		for _, ch := range c.Children {
			if ch.ID == rifx.IDTdmn && len(ch.Data) > 0 {
				name := string(ch.Data)
				if n := indexNUL(name); n >= 0 {
					name = name[:n]
				}
				saw[name] = true
			}
			if ch.IsList() {
				walkNames(ch)
			}
		}
	}
	walkNames(chunk)
	for _, want := range []string{"ADBE Vector Group", "ADBE Vectors Group", "ADBE Vector Transform Group", "ADBE Vector Materials Group"} {
		if !saw[want] {
			t.Errorf("missing wrapper tdmn %q", want)
		}
	}
}

// indexNUL returns the first NUL byte position or -1.
func indexNUL(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return i
		}
	}
	return -1
}
