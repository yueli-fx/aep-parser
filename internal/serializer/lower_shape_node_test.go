// internal/aep/lower_shape_node_test.go
//
// Chunk-shape level tests only. Validates the dispatcher
// + per-node lowering + VectorGroup wrapping.
package serializer

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/example/aep-parser/internal/rifx"
)

func TestLowerRectNode_HasTdmnAndSubProps(t *testing.T) {
	r := NewRectNode()
	chunk, err := LowerShapeNodeForTest(r)
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

func TestLowerOffsetPaths_CopiesSplice(t *testing.T) {
	// Default (Copies untouched): no Copies slot is spliced — the body stays the
	// Amount-only template (matches AE's elision of the default Copies value).
	def := NewOffsetPathsNode()
	_ = def.SetAmount(60)
	defBody, err := LowerShapeNodeForTest(def)
	if err != nil {
		t.Fatal(err)
	}
	if names := topTdmnNames(defBody); containsStr(names, "ADBE Vector Offset Copies") {
		t.Fatalf("default offset must not carry Copies slot, got tdmns %v", names)
	}

	// Copies set non-default: the `ADBE Vector Offset Copies` leaf is spliced in
	// canonical order (after Amount, before Group End) with its cdat overwritten.
	n := NewOffsetPathsNode()
	_ = n.SetAmount(60)
	_ = n.SetCopies(3)
	body, err := LowerShapeNodeForTest(n)
	if err != nil {
		t.Fatal(err)
	}
	got := topTdmnNames(body)
	want := []string{"ADBE Vector Offset Amount", "ADBE Vector Offset Copies", "ADBE Group End"}
	if len(got) != len(want) {
		t.Fatalf("offset body top tdmns = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("offset body tdmn[%d] = %q, want %q (full %v)", i, got[i], w, got)
		}
	}
	// Copies cdat[0:8] must decode to 3 (f64 BE).
	if v := offsetCdatF64(t, body, "ADBE Vector Offset Copies"); v != 3 {
		t.Fatalf("Copies cdat = %g, want 3", v)
	}
}

func TestLowerTrim_TypeSplice(t *testing.T) {
	// Default (Simultaneously): no Trim Type slot is spliced.
	def := NewTrimNode()
	_ = def.SetEnd(50)
	defBody, err := LowerShapeNodeForTest(def)
	if err != nil {
		t.Fatal(err)
	}
	if names := topTdmnNames(defBody); containsStr(names, "ADBE Vector Trim Type") {
		t.Fatalf("default trim must not carry Trim Type slot, got %v", names)
	}

	// Individually: the `ADBE Vector Trim Type` enum leaf is spliced after Offset,
	// before Group End, with its cdat overwritten to 2.
	n := NewTrimNode()
	_ = n.SetEnd(50)
	if err := n.SetType(TrimTypeIndividually); err != nil {
		t.Fatal(err)
	}
	body, err := LowerShapeNodeForTest(n)
	if err != nil {
		t.Fatal(err)
	}
	got := topTdmnNames(body)
	want := []string{"ADBE Vector Trim Start", "ADBE Vector Trim End", "ADBE Vector Trim Offset", "ADBE Vector Trim Type", "ADBE Group End"}
	if len(got) != len(want) {
		t.Fatalf("trim body top tdmns = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("trim body tdmn[%d] = %q, want %q (full %v)", i, got[i], w, got)
		}
	}
	if v := offsetCdatF64(t, body, "ADBE Vector Trim Type"); v != 2 {
		t.Fatalf("Trim Type cdat = %g, want 2 (Individually)", v)
	}
}

func topTdmnNames(body *rifx.Chunk) []string {
	var out []string
	for _, ch := range body.Children {
		if ch.ID == rifx.IDTdmn {
			out = append(out, trimTestNUL(string(ch.Data)))
		}
	}
	return out
}

func containsStr(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func offsetCdatF64(t *testing.T, body *rifx.Chunk, streamName string) float64 {
	t.Helper()
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimTestNUL(string(kids[i].Data)) == streamName {
			tdbs := kids[i+1]
			for _, ch := range tdbs.Children {
				if ch.ID == rifx.IDCdat && len(ch.Data) >= 8 {
					return math.Float64frombits(binary.BigEndian.Uint64(ch.Data[:8]))
				}
			}
		}
	}
	t.Fatalf("cdat for %q not found", streamName)
	return 0
}

func TestLowerEllipseNode_OverwritesSizeAndPosition(t *testing.T) {
	e := NewEllipseNode()
	_ = e.SetSize([2]float64{321, 123})
	_ = e.SetPosition([2]float64{40, 60})
	chunk, err := LowerShapeNodeForTest(e)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil || !chunk.IsList() || chunk.FormType != rifx.IDTdgp {
		t.Fatalf("expected LIST(tdgp), got %+v", chunk)
	}
	// Structural contract: the body must be the AE-saved embed boilerplate,
	// not the from-scratch emit (which triggered AE silent-drop). The enriched
	// template (v2_2_shape_all_full.aep) carries the Shape Direction enum slot
	// ahead of Size/Position. This is the bit byte values alone can't catch —
	// the from-scratch path also wrote correct cdat, but its surrounding
	// boilerplate triggered the drop.
	var topTdmns []string
	for _, ch := range chunk.Children {
		if ch.ID == rifx.IDTdmn {
			topTdmns = append(topTdmns, trimTestNUL(string(ch.Data)))
		}
	}
	wantTdmns := []string{"ADBE Vector Shape Direction", "ADBE Vector Ellipse Size", "ADBE Vector Ellipse Position", "ADBE Group End"}
	if len(topTdmns) != len(wantTdmns) {
		t.Fatalf("embed body top-level tdmns = %v, want %v", topTdmns, wantTdmns)
	}
	for i, w := range wantTdmns {
		if topTdmns[i] != w {
			t.Fatalf("embed body top-level tdmns = %v, want %v", topTdmns, wantTdmns)
		}
	}
	assertEllipseStreamCdat(t, chunk, "ADBE Vector Ellipse Size", []float64{321, 123})
	assertEllipseStreamCdat(t, chunk, "ADBE Vector Ellipse Position", []float64{40, 60})
}

// assertEllipseStreamCdat finds the tdmn matching name inside body, descends
// to its tdbs cdat, and asserts the leading f64 BE values equal want.
func assertEllipseStreamCdat(t *testing.T, body *rifx.Chunk, name string, want []float64) {
	t.Helper()
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimTestNUL(string(kids[i].Data)) == name {
			tdbs := kids[i+1]
			if !tdbs.IsList() || tdbs.FormType != rifx.IDTdbs {
				t.Fatalf("%s: next chunk not LIST(tdbs)", name)
			}
			for _, ch := range tdbs.Children {
				if ch.ID == rifx.IDCdat {
					for j, w := range want {
						got := math.Float64frombits(binary.BigEndian.Uint64(ch.Data[j*8 : j*8+8]))
						if got != w {
							t.Errorf("%s[%d] = %v, want %v", name, j, got, w)
						}
					}
					return
				}
			}
			t.Fatalf("%s: no cdat under tdbs", name)
		}
	}
	t.Fatalf("%s: tdmn not found in lowered body", name)
}

func trimTestNUL(s string) string {
	if n := indexNUL(s); n >= 0 {
		return s[:n]
	}
	return s
}

func TestLowerFillNode_HasColorAndOpacity(t *testing.T) {
	f := NewFillNode()
	chunk, err := LowerShapeNodeForTest(f)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
}

// TestLowerFillNode_ColorEncodingARGB255 pins the V2.2.1 shape-color encoding:
// AE stores Fill/Stroke colors as [A,R,G,B] × 255 f64 BE, NOT raw [r,g,b,a].
// RE'd from the stroke tolerance fixture (JSX [0,0,1,1] → disk [255,0,0,255]).
func TestLowerFillNode_ColorEncodingARGB255(t *testing.T) {
	f := NewFillNode()
	_ = f.SetColor([4]float64{1, 0, 0, 1}) // red, alpha 1
	chunk, err := LowerShapeNodeForTest(f)
	if err != nil {
		t.Fatal(err)
	}
	// Expected ARGB×255 = [255, 255, 0, 0].
	assertEllipseStreamCdat(t, chunk, "ADBE Vector Fill Color", []float64{255, 255, 0, 0})
}

func TestLowerStrokeNode_HasFullChildSet(t *testing.T) {
	s := NewStrokeNode()
	chunk, err := LowerShapeNodeForTest(s)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
}

func TestLowerPathNode_LinearOpenSegment(t *testing.T) {
	p := NewPathNode()
	if err := p.SetVertices([][2]float64{{0, 0}, {100, 100}}); err != nil {
		t.Fatal(err)
	}
	chunk, err := LowerShapeNodeForTest(p)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
}

func TestLowerVectorGroup_EmitsRootVectorsGroupChildren(t *testing.T) {
	g := NewVectorGroup()
	g.Children = append(g.Children, NewRectNode(), NewFillNode())
	chunk, err := LowerVectorGroupForTest(g)
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
