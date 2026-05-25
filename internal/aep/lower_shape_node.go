// internal/aep/lower_shape_node.go
//
// Phase 2 Task 2.2 — 5 per-node lowering funcs (Rect / Ellipse / Path / Fill
// / Stroke) + lowerVectorGroup for the Root Vectors Group container.
//
// Match-names per RE-S5a-d Phase 0 observations. V2.2 emits ALL sub-props
// even when default-valued (AE elides; we don't — see lower_property_stream.go
// preamble); Phase 4 roundtrip validates AE accepts the non-elided form.
package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// shapeMatchNames maps runtime ShapeNodeKind → AE match-name string. The
// table is serializer-only; runtime API uses the Go enum (Inv-2). Strings
// match Phase 0 RE fixture observations (RE-S4 / RE-S5a-d).
var shapeMatchNames = map[ShapeNodeKind]string{
	ShapeKindRect:    "ADBE Vector Shape - Rect",
	ShapeKindEllipse: "ADBE Vector Shape - Ellipse",
	ShapeKindPath:    "ADBE Vector Shape - Group",
	ShapeKindFill:    "ADBE Vector Graphic - Fill",
	ShapeKindStroke:  "ADBE Vector Graphic - Stroke",
	ShapeKindGroup:   "ADBE Vector Group",
}

// LowerShapeNodeForTest exports lowerShapeNode for unit tests.
func LowerShapeNodeForTest(n ShapeNode) (*rifx.Chunk, error) {
	return lowerShapeNode(n, NewLowerCtxForTest())
}

// LowerVectorGroupForTest exports lowerVectorGroup for unit tests.
func LowerVectorGroupForTest(g *VectorGroup) (*rifx.Chunk, error) {
	return lowerVectorGroup(g, NewLowerCtxForTest())
}

// lowerShapeNode dispatches to the per-kind lowering function. Returns a
// LIST(tdgp) chunk wrapping the node body.
func lowerShapeNode(n ShapeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	switch node := n.(type) {
	case *RectNode:
		return lowerRectNode(node, ctx)
	case *EllipseNode:
		return lowerEllipseNode(node, ctx)
	case *PathNode:
		return lowerPathNode(node, ctx)
	case *FillNode:
		return lowerFillNode(node, ctx)
	case *StrokeNode:
		return lowerStrokeNode(node, ctx)
	default:
		return nil, fmt.Errorf("lowerShapeNode: unsupported kind %v", n.Kind())
	}
}

// nodeBodyTdgp builds the inner LIST(tdgp) carried after each shape's tdmn.
// Standard shape: tdsb + tdsn + N × sub-property tdgp + tdmn(Group End).
func nodeBodyTdgp(displayName string, subProps []*rifx.Chunk) *rifx.Chunk {
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	body.Children = append(body.Children, makeTdsb(), makeTdsn(displayName))
	// Each sub-property is a LIST(tdgp) preceded by its own tdmn; the
	// `LowerXxxStream` funcs return the LIST(tdgp) with tdmn already as
	// child[0]. Inline their children into the parent body so the on-disk
	// tdmn + LIST(tdbs) pattern appears flat (matching RE observations).
	for _, sp := range subProps {
		body.Children = append(body.Children, sp.Children...)
	}
	body.Children = append(body.Children, makeTdmn("ADBE Group End"))
	return body
}

func lowerRectNode(r *RectNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	size, err := LowerVec2Stream(r.size, "ADBE Vector Rect Size", "Size", ctx)
	if err != nil {
		return nil, err
	}
	pos, err := LowerVec2Stream(r.position, "ADBE Vector Rect Position", "Position", ctx)
	if err != nil {
		return nil, err
	}
	rnd, err := LowerFloat64Stream(r.roundness, "ADBE Vector Rect Roundness", "Roundness", ctx)
	if err != nil {
		return nil, err
	}
	direction := emptySubPropPlaceholder("ADBE Vector Shape Direction", "Direction")
	body := nodeBodyTdgp("Rectangle Path", []*rifx.Chunk{direction, size, pos, rnd})
	return body, nil
}

func lowerEllipseNode(e *EllipseNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	size, err := LowerVec2Stream(e.size, "ADBE Vector Ellipse Size", "Size", ctx)
	if err != nil {
		return nil, err
	}
	pos, err := LowerVec2Stream(e.position, "ADBE Vector Ellipse Position", "Position", ctx)
	if err != nil {
		return nil, err
	}
	direction := emptySubPropPlaceholder("ADBE Vector Shape Direction", "Direction")
	body := nodeBodyTdgp("Ellipse Path", []*rifx.Chunk{direction, size, pos})
	return body, nil
}

func lowerPathNode(p *PathNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	path, err := LowerPathStream(p.path, "ADBE Vector Shape", "Path", ctx)
	if err != nil {
		return nil, err
	}
	body := nodeBodyTdgp("Path", []*rifx.Chunk{path})
	return body, nil
}

func lowerFillNode(f *FillNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	// Per RE-S5c: Fill children are Blend Mode / Composite Order / Fill Rule /
	// Color / Opacity (5 total). V2.2 hot path emits Color + Opacity typed;
	// the 3 enum-style props get empty-placeholder tdgp groups.
	blendMode := emptySubPropPlaceholder("ADBE Vector Blend Mode", "Blend Mode")
	compOrder := emptySubPropPlaceholder("ADBE Vector Composite Order", "Composite Order")
	fillRule := emptySubPropPlaceholder("ADBE Vector Fill Rule", "Fill Rule")
	color, err := LowerColorStream(f.color, "ADBE Vector Fill Color", "Color", ctx)
	if err != nil {
		return nil, err
	}
	op, err := LowerFloat64Stream(f.opacity, "ADBE Vector Fill Opacity", "Opacity", ctx)
	if err != nil {
		return nil, err
	}
	body := nodeBodyTdgp("Fill", []*rifx.Chunk{blendMode, compOrder, fillRule, color, op})
	return body, nil
}

func lowerStrokeNode(s *StrokeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	// Per RE-S5d: Stroke children (11) =
	//   Blend Mode / Composite Order / Stroke Color / Stroke Opacity /
	//   Stroke Width / Line Cap / Line Join / Miter Limit /
	//   Dashes (nested group) / Taper (nested group) / Wave (nested group).
	// V2.2 hot path emits Color + Opacity + Width typed; the 8 remaining
	// children get 3-child empty placeholders (RE-S5d showed Dashes / Taper /
	// Wave persist 3-child header-only group even at default).
	blendMode := emptySubPropPlaceholder("ADBE Vector Blend Mode", "Blend Mode")
	compOrder := emptySubPropPlaceholder("ADBE Vector Composite Order", "Composite Order")
	color, err := LowerColorStream(s.color, "ADBE Vector Stroke Color", "Color", ctx)
	if err != nil {
		return nil, err
	}
	op, err := LowerFloat64Stream(s.opacity, "ADBE Vector Stroke Opacity", "Opacity", ctx)
	if err != nil {
		return nil, err
	}
	width, err := LowerFloat64Stream(s.width, "ADBE Vector Stroke Width", "Width", ctx)
	if err != nil {
		return nil, err
	}
	lineCap := emptySubPropPlaceholder("ADBE Vector Stroke Line Cap", "Line Cap")
	lineJoin := emptySubPropPlaceholder("ADBE Vector Stroke Line Join", "Line Join")
	miter := emptySubPropPlaceholder("ADBE Vector Stroke Miter Limit", "Miter Limit")
	dashes := emptySubPropPlaceholder("ADBE Vector Stroke Dashes", "Dashes")
	taper := emptySubPropPlaceholder("ADBE Vector Stroke Taper", "Taper")
	wave := emptySubPropPlaceholder("ADBE Vector Stroke Wave", "Wave")
	body := nodeBodyTdgp("Stroke", []*rifx.Chunk{
		blendMode, compOrder, color, op, width,
		lineCap, lineJoin, miter, dashes, taper, wave,
	})
	return body, nil
}

// emptySubPropPlaceholder returns a `tdmn + LIST(tdgp)(tdsb + tdsn +
// tdmn(Group End))` pair for sub-properties V2.2 doesn't expose as typed
// setters. The pair matches RE-S5d's "Dashes/Taper/Wave 3-child empty
// header-only" placeholder pattern. Returned chunk is shaped as a tdgp
// container holding two children (tdmn + LIST tdgp) so callers can inline
// via nodeBodyTdgp's expansion logic.
func emptySubPropPlaceholder(matchName, displayName string) *rifx.Chunk {
	holder := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	holder.Children = append(holder.Children, makeTdmn(matchName))
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	body.Children = append(body.Children, makeTdsb(), makeTdsn(displayName), makeTdmn("ADBE Group End"))
	holder.Children = append(holder.Children, body)
	return holder
}

// lowerVectorGroup wraps shape-node children into the Root Vectors Group's
// inner LIST(tdgp). The on-disk shape per iter-5 RE of tolerance.aep +
// re_shapes.aep (every AE-saved fixture observed) is a 5-level nesting:
//
//	[LIST tdgp]                      ← Root Vectors Group body (this return)
//	  tdsb(0x00000401) + tdsn
//	  tdmn("ADBE Vector Group")
//	  [LIST tdgp]                    ← Vector Group body (3-child fixed routing)
//	    tdsb(0x00000001) + tdsn
//	    tdmn("ADBE Vectors Group")
//	    [LIST tdgp]                  ← Vectors Group body (holds user shapes)
//	      tdsb(0x00000401) + tdsn
//	      N × (tdmn(<shape>) + LIST(tdgp, shape body))
//	      tdmn("ADBE Group End")
//	    tdmn("ADBE Vector Transform Group") + empty LIST(tdgp)
//	    tdmn("ADBE Vector Materials Group") + empty LIST(tdgp)
//	    tdmn("ADBE Group End")
//	  tdmn("ADBE Group End")
//
// iter-4 had us flatten everything into Root Vectors Group body directly —
// AE 2025 parsed the file without exception but silently dropped the layer
// from comp.layers (iter-4 scar "bug 8 candidate"). The wrappers are
// structural: AE Shape Layer's Contents always holds one or more
// "ADBE Vector Group" entries (each is what UI shows as "Group N"), and
// each Vector Group always carries the 3-child fixed routing (Vectors Group
// for shape kids + Transform + Materials).
//
// V2.2 maps the runtime `shapeRootGroup.Children = [Rect, Fill, ...]` to a
// SINGLE Vector Group wrapper (semantic = AE's auto-created "Group 1"). V2.3+
// may expose multiple user-named groups.
//
// Children render in order: Children[0] = bottom, Children[len-1] = top
// (spec §3.2).
func lowerVectorGroup(g *VectorGroup, ctx *lowerCtx) (*rifx.Chunk, error) {
	// Innermost: Vectors Group body — holds the actual shape kids.
	vectorsGroupBody := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	vectorsGroupBody.Children = append(vectorsGroupBody.Children, makeTdsbContainer(), makeTdsn(""))
	for _, child := range g.Children {
		mn := shapeMatchNames[child.Kind()]
		body, err := lowerShapeNode(child, ctx)
		if err != nil {
			return nil, err
		}
		vectorsGroupBody.Children = append(vectorsGroupBody.Children, makeTdmn(mn), body)
	}
	vectorsGroupBody.Children = append(vectorsGroupBody.Children, makeTdmn("ADBE Group End"))

	// Middle: Vector Group body — fixed 3-child routing (Vectors Group +
	// Vector Transform Group + Vector Materials Group). The latter two are
	// per-group transform / materials property groups that AE always emits
	// even when default; tolerance.aep dumps them as 3-child empty
	// placeholders (tdsb + tdsn + Group End).
	vectorGroupBody := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	vectorGroupBody.Children = append(vectorGroupBody.Children, makeTdsb(), makeTdsn(""))
	vectorGroupBody.Children = append(vectorGroupBody.Children,
		makeTdmn("ADBE Vectors Group"), vectorsGroupBody,
		makeTdmn("ADBE Vector Transform Group"), emptyPropGroup(),
		makeTdmn("ADBE Vector Materials Group"), emptyPropGroup(),
		makeTdmn("ADBE Group End"),
	)

	// Outermost: Root Vectors Group body — holds one Vector Group wrapper.
	root := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	root.Children = append(root.Children, makeTdsbContainer(), makeTdsn(""))
	root.Children = append(root.Children,
		makeTdmn("ADBE Vector Group"), vectorGroupBody,
		makeTdmn("ADBE Group End"),
	)
	return root, nil
}
