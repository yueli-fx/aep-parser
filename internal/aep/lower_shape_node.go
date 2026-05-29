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
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/example/aep-parser/internal/rifx"
)

// iter-8 embed: tolerance shape body bytes used as boilerplate skeleton.
// Validator boundary RE'd via tmp_debug/swap_rect_body + swap_fill_body +
// swap_both_bodies: each shape primitive body in our from-scratch emit
// triggers AE silent drop independently. embed-tolerance-bytes approach
// (same pattern as iter-7 lowerLayerTransform) bypasses byte-level RE.
//
//go:embed templates/v2_2_shape_rect_body.bin
var v22ShapeRectBodyBytes []byte

//go:embed templates/v2_2_shape_fill_body.bin
var v22ShapeFillBodyBytes []byte

//go:embed templates/v2_2_shape_ellipse_body.bin
var v22ShapeEllipseBodyBytes []byte

var (
	v22ShapeRectOnce  sync.Once
	v22ShapeRectCache *rifx.Chunk
	v22ShapeRectErr   error

	v22ShapeFillOnce  sync.Once
	v22ShapeFillCache *rifx.Chunk
	v22ShapeFillErr   error

	v22ShapeEllipseOnce  sync.Once
	v22ShapeEllipseCache *rifx.Chunk
	v22ShapeEllipseErr   error
)

func cloneShapeRectBody() (*rifx.Chunk, error) {
	v22ShapeRectOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeRectBodyBytes))
		if err != nil {
			v22ShapeRectErr = fmt.Errorf("parse v22ShapeRectBodyBytes: %w", err)
			return
		}
		v22ShapeRectCache = ch
	})
	if v22ShapeRectErr != nil {
		return nil, v22ShapeRectErr
	}
	return cloneChunk(v22ShapeRectCache), nil
}

func cloneShapeFillBody() (*rifx.Chunk, error) {
	v22ShapeFillOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeFillBodyBytes))
		if err != nil {
			v22ShapeFillErr = fmt.Errorf("parse v22ShapeFillBodyBytes: %w", err)
			return
		}
		v22ShapeFillCache = ch
	})
	if v22ShapeFillErr != nil {
		return nil, v22ShapeFillErr
	}
	return cloneChunk(v22ShapeFillCache), nil
}

func cloneShapeEllipseBody() (*rifx.Chunk, error) {
	v22ShapeEllipseOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeEllipseBodyBytes))
		if err != nil {
			v22ShapeEllipseErr = fmt.Errorf("parse v22ShapeEllipseBodyBytes: %w", err)
			return
		}
		v22ShapeEllipseCache = ch
	})
	if v22ShapeEllipseErr != nil {
		return nil, v22ShapeEllipseErr
	}
	return cloneChunk(v22ShapeEllipseCache), nil
}

// overwriteShapeStreamCdat finds the tdmn matching `streamName` inside
// `body` (LIST tdgp), descends into the inner LIST(tdbs), and overwrites
// the first `valueBytes` of the cdat with `data`. Used to inject runtime
// Size / Color / etc values into the embedded tolerance template.
func overwriteShapeStreamCdat(body *rifx.Chunk, streamName string, data []byte) {
	kids := body.Children
	for i := 0; i < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == streamName && i+1 < len(kids) {
			tdbs := kids[i+1]
			if !tdbs.IsList() || tdbs.FormType != rifx.IDTdbs {
				return
			}
			for _, ch := range tdbs.Children {
				if ch.ID == rifx.IDCdat && len(ch.Data) >= len(data) {
					copy(ch.Data[:len(data)], data)
					return
				}
			}
			return
		}
	}
}

// encodeF64sBE returns the BE bytes for a slice of f64 values, packed
// without padding.
func encodeF64sBE(vs ...float64) []byte {
	out := make([]byte, len(vs)*8)
	for i, v := range vs {
		binary.BigEndian.PutUint64(out[i*8:(i+1)*8], math.Float64bits(v))
	}
	return out
}

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

// lowerRectNode emits a Rect shape body using iter-8 embedded tolerance
// bytes (templates/v2_2_shape_rect_body.bin). From-scratch construction
// triggered silent drop (transplant-isolated via swap_rect_body); embedding
// the canonical body + overwriting Size cdat with runtime user values is
// the validator-safe path.
//
// V2.2 alpha limitations (V2.2.1 work):
//   - Rect Position / Roundness: runtime-only, NOT persisted (tolerance
//     elides them; embedded body has no slot to overwrite).
//   - Rect Direction: AE default ("ToTheRight"), no runtime customization.
//   - Animated Size: first keyframe value used as static fallback.
func lowerRectNode(r *RectNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeRectBody()
	if err != nil {
		return nil, err
	}
	val := r.size.static
	if r.size.mode == StreamModeAnimated && len(r.size.keyframes) > 0 {
		val = r.size.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Rect Size", encodeF64sBE(val[0], val[1]))
	return body, nil
}

// lowerEllipseNode emits an Ellipse shape body using V2.2.1 embedded tolerance
// bytes (templates/v2_2_shape_ellipse_body.bin). Same rationale as
// lowerRectNode — from-scratch emit triggered AE silent-drop (the body's
// boilerplate, not the values, fails AE's semantic validation); embedding the
// canonical AE-saved body + overwriting Size/Position cdat with runtime values
// is the validator-safe path.
//
// V2.2.1 limitations:
//   - Direction: AE default (the AE-saved body elides the Direction sub-prop;
//     embedded body has no slot to overwrite).
//   - Animated Size / Position: first keyframe value used as static fallback.
func lowerEllipseNode(e *EllipseNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeEllipseBody()
	if err != nil {
		return nil, err
	}
	sz := e.size.static
	if e.size.mode == StreamModeAnimated && len(e.size.keyframes) > 0 {
		sz = e.size.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Size", encodeF64sBE(sz[0], sz[1]))
	ps := e.position.static
	if e.position.mode == StreamModeAnimated && len(e.position.keyframes) > 0 {
		ps = e.position.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Position", encodeF64sBE(ps[0], ps[1]))
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

// lowerFillNode emits a Fill graphic body using iter-8 embedded tolerance
// bytes (templates/v2_2_shape_fill_body.bin). Same rationale as
// lowerRectNode — transplant tests proved from-scratch Fill body triggers
// silent drop; embedded canonical body + cdat overwrite for Color values
// is the validator-safe path.
//
// V2.2 alpha limitations (V2.2.1 work):
//   - Fill Opacity / Blend Mode / Composite Order / Fill Rule: tolerance
//     elides; embedded body has no slot to overwrite. Runtime-only API.
//   - Color encoding: tolerance.aep stores Fill Color cdat in a non-obvious
//     scale (bytes don't match user 0-1 input as f64 BE; e.g. JSX 0.5 →
//     disk byte 0x406fe... ≈ 255). iter-8 writes user's [r,g,b,a] as f64
//     BE in cdat[0..32] regardless — if AE applies internal scaling, visible
//     color may not match user input. V2.2.1 will RE the encoding.
//   - Animated Color: first keyframe value used as static fallback.
func lowerFillNode(f *FillNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeFillBody()
	if err != nil {
		return nil, err
	}
	val := f.color.static
	if f.color.mode == StreamModeAnimated && len(f.color.keyframes) > 0 {
		val = f.color.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Fill Color", encodeF64sBE(val[0], val[1], val[2], val[3]))
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
