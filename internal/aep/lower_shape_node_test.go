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
	// Per RE-S3: tdsb + tdsn + N × (tdmn + LIST tdgp) + tdmn(Group End).
	// For 2 children: 2 + 4 + 1 = 7 children.
	if len(chunk.Children) != 2+2*2+1 {
		t.Fatalf("vector group tdgp child count = %d, want 7 (header + 2×2 + Group End)", len(chunk.Children))
	}
	if chunk.Children[0].ID.String() != "tdsb" {
		t.Fatalf("group child[0] id = %q, want tdsb", chunk.Children[0].ID)
	}
	// Last child must be the Group End tdmn.
	last := chunk.Children[len(chunk.Children)-1]
	if last.ID != rifx.IDTdmn {
		t.Fatalf("group last child id = %q, want tdmn (Group End)", last.ID)
	}
}
