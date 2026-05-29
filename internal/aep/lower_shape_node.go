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

//go:embed templates/v2_2_shape_path_body.bin
var v22ShapePathBodyBytes []byte

//go:embed templates/v2_2_shape_stroke_body.bin
var v22ShapeStrokeBodyBytes []byte

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

	v22ShapePathOnce  sync.Once
	v22ShapePathCache *rifx.Chunk
	v22ShapePathErr   error

	v22ShapeStrokeOnce  sync.Once
	v22ShapeStrokeCache *rifx.Chunk
	v22ShapeStrokeErr   error
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

func cloneShapePathBody() (*rifx.Chunk, error) {
	v22ShapePathOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapePathBodyBytes))
		if err != nil {
			v22ShapePathErr = fmt.Errorf("parse v22ShapePathBodyBytes: %w", err)
			return
		}
		v22ShapePathCache = ch
	})
	if v22ShapePathErr != nil {
		return nil, v22ShapePathErr
	}
	return cloneChunk(v22ShapePathCache), nil
}

func cloneShapeStrokeBody() (*rifx.Chunk, error) {
	v22ShapeStrokeOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeStrokeBodyBytes))
		if err != nil {
			v22ShapeStrokeErr = fmt.Errorf("parse v22ShapeStrokeBodyBytes: %w", err)
			return
		}
		v22ShapeStrokeCache = ch
	})
	if v22ShapeStrokeErr != nil {
		return nil, v22ShapeStrokeErr
	}
	return cloneChunk(v22ShapeStrokeCache), nil
}

// encodeShapeColorBE returns the AE shape-color cdat bytes for an [r,g,b,a]
// (0..1) color: AE stores colors as [A,R,G,B] × 255 as f64 BE (RE'd from the
// stroke tolerance fixture — JSX [0,0,1,1] → disk [255,0,0,255]). This is the
// long-deferred "Fill Color encoding" too; both Stroke and Fill use it.
func encodeShapeColorBE(c [4]float64) []byte {
	return encodeF64sBE(c[3]*255, c[0]*255, c[1]*255, c[2]*255)
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
//   - Animated Size: persisted as keyframes (V2.2.1 — cdat→LIST(list) inject).
func lowerRectNode(r *RectNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeRectBody()
	if err != nil {
		return nil, err
	}
	if r.size.mode == StreamModeAnimated && len(r.size.keyframes) > 0 {
		if err := injectAnimatedVec2(body, "ADBE Vector Rect Size", r.size.keyframes, ctx); err != nil {
			return nil, err
		}
		return body, nil
	}
	val := r.size.static
	overwriteShapeStreamCdat(body, "ADBE Vector Rect Size", encodeF64sBE(val[0], val[1]))
	return body, nil
}

// injectAnimatedVec2 converts a static shape Vec2 stream in an embedded body
// into an animated one: it finds the tdmn matching streamName, descends into
// the following LIST(tdbs), and replaces the static cdat child with the
// animated LIST(list)(lhd3+ldat) keyframe container (AE keeps tdsb/tdsn/tdb4/
// tdum/tduM unchanged — only cdat ↔ LIST(list) flips). Non-spatial dim-2
// layout (header07=0x00) per the kf RE fixture (Rect/Ellipse Size).
func injectAnimatedVec2(body *rifx.Chunk, streamName string, kfs []StreamKeyframe[[2]float64], ctx *lowerCtx) error {
	return injectAnimatedVec2L(body, streamName, kfs, ctx, valueLayout{dim: 2, headerByte: 0x00, spatial: false})
}

// injectAnimatedVec2L injects an animated Vec2 stream with an explicit layout
// (Size = non-spatial header07=0x00 bpk 88; shape Position = spatial header07=
// 0x07 bpk 104 value@0x38).
func injectAnimatedVec2L(body *rifx.Chunk, streamName string, kfs []StreamKeyframe[[2]float64], ctx *lowerCtx, layout valueLayout) error {
	kfList, err := encodeKeyframes(kfs, layout, encode2D, ctx)
	if err != nil {
		return err
	}
	return injectAnimatedStream(body, streamName, kfList)
}

// injectAnimatedColor converts a static shape Color stream into an animated
// one. Color keyframes use the spatial-style block (value at 0x38, bpk
// 0x38+3*dim*8 = 152 for dim=4) with the [A,R,G,B]×255 value encoding — RE'd
// from the kf fixture (Fill/Stroke Color).
func injectAnimatedColor(body *rifx.Chunk, streamName string, kfs []StreamKeyframe[[4]float64], ctx *lowerCtx) error {
	encColor := func(c [4]float64) []byte { return encodeShapeColorBE(c) }
	kfList, err := encodeKeyframes(kfs, valueLayout{dim: 4, headerByte: 0x01, spatial: true}, encColor, ctx)
	if err != nil {
		return err
	}
	return injectAnimatedStream(body, streamName, kfList)
}

// injectAnimatedStream flips a static shape stream in an embedded body to
// animated: it finds the tdmn matching streamName, descends into the following
// LIST(tdbs), patches the tdb4 static→animated flags, and replaces the static
// cdat with the supplied LIST(list)(lhd3+ldat) keyframe container. AE keeps
// tdsb/tdsn/tdb4/tdum/tduM otherwise unchanged.
func injectAnimatedStream(body *rifx.Chunk, streamName string, kfList *rifx.Chunk) error {
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == streamName {
			tdbs := kids[i+1]
			if !tdbs.IsList() || tdbs.FormType != rifx.IDTdbs {
				return fmt.Errorf("injectAnimatedStream: %s next chunk not LIST(tdbs)", streamName)
			}
			// tdb4 static→animated flags (RE across Rect Size + Fill Color):
			// @0x05 clear bit0, @0x44 = 0x01, @0x4f clear bit0. Without this AE
			// expects a cdat per the static tdb4 and reports "file data missing".
			// NB: modern AE writes lowercase "tdb4"; rifx.IDTdb4 is the legacy
			// UPPERCASE "Tdb4" (chunk IDs are case-sensitive — incident report
			// chunk-id-case-tdb4.md), so match the lowercase literal.
			if tdb4 := findChildID(tdbs, rifx.ChunkID{'t', 'd', 'b', '4'}); tdb4 != nil && len(tdb4.Data) > 0x4f {
				tdb4.Data[0x05] &^= 0x01
				tdb4.Data[0x44] = 0x01
				tdb4.Data[0x4f] &^= 0x01
			}
			for j, ch := range tdbs.Children {
				if ch.ID == rifx.IDCdat {
					tdbs.Children[j] = kfList
					return nil
				}
			}
			return fmt.Errorf("injectAnimatedStream: %s no cdat to replace", streamName)
		}
	}
	return fmt.Errorf("injectAnimatedStream: %s tdmn not found", streamName)
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
//   - Animated Size: keyframes persisted (non-spatial Vec2). Animated Position:
//     keyframes persisted (spatial Vec2, bpk 104 value@0x38 — RE'd from the
//     ellipse-kf fixture; AE recomputes spatial tangents on load).
func lowerEllipseNode(e *EllipseNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeEllipseBody()
	if err != nil {
		return nil, err
	}
	if e.size.mode == StreamModeAnimated && len(e.size.keyframes) > 0 {
		if err := injectAnimatedVec2(body, "ADBE Vector Ellipse Size", e.size.keyframes, ctx); err != nil {
			return nil, err
		}
	} else {
		overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Size", encodeF64sBE(e.size.static[0], e.size.static[1]))
	}
	if e.position.mode == StreamModeAnimated && len(e.position.keyframes) > 0 {
		if err := injectAnimatedVec2L(body, "ADBE Vector Ellipse Position", e.position.keyframes, ctx,
			valueLayout{dim: 2, headerByte: 0x07, spatial: true, motionPath: true}); err != nil {
			return nil, err
		}
	} else {
		overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Position", encodeF64sBE(e.position.static[0], e.position.static[1]))
	}
	return body, nil
}

// lowerPathNode emits a Path shape body using V2.2.1 embedded tolerance bytes
// (templates/v2_2_shape_path_body.bin). From-scratch emit CRASHED AE 2020
// ("After Effects 已崩溃 (0::42)") — the om-s/tdb4 scaffolding is too fragile
// to hand-build. We clone the AE-native body and splice in the user's geometry
// (shph/lhd3/ldat from encodeBezier, whose layout matches AE byte-for-byte per
// V2.2.1 ldat RE), keeping AE's exact scaffolding (om-s header + omks + omtn).
//
// V2.2.1 limitations: linear segments only (SetVertices zeroes tangents);
// animated paths use the first keyframe as a static fallback.
func lowerPathNode(p *PathNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapePathBody()
	if err != nil {
		return nil, err
	}
	bp := p.path.static
	if p.path.mode == StreamModeAnimated && len(p.path.keyframes) > 0 {
		bp = p.path.keyframes[0].Value
	}
	if err := splicePathGeometry(body, bp); err != nil {
		return nil, err
	}
	return body, nil
}

// splicePathGeometry replaces the shph/lhd3/ldat geometry chunks inside the
// embedded path body's LIST(shap) with freshly-encoded geometry for bp,
// leaving AE's scaffolding (om-s header, omks/shap/kfl wrappers, omtn) intact.
func splicePathGeometry(body *rifx.Chunk, bp BezierPath) error {
	shap := findListByForm(body, rifx.IDShap)
	if shap == nil {
		return fmt.Errorf("splicePathGeometry: LIST(shap) not found in embed body")
	}
	newShph, newLhd3, newLdat := encodeBezier(bp)
	kfl := findListByForm(shap, rifx.IDkfl)
	if kfl == nil {
		return fmt.Errorf("splicePathGeometry: LIST(kfl) not found in shap")
	}
	// Replace shph (direct child of shap) and lhd3/ldat (children of kfl) in
	// place, preserving sibling order (shph, kfl, omtn) and (lhd3, ldat).
	for i, ch := range shap.Children {
		if ch.ID == rifx.IDShph {
			shap.Children[i] = newShph
		}
	}
	for i, ch := range kfl.Children {
		switch ch.ID {
		case rifx.IDLhd3:
			kfl.Children[i] = newLhd3
		case rifx.IDLdat:
			kfl.Children[i] = newLdat
		}
	}
	return nil
}

// findChildID returns the first direct child with the given chunk ID, or nil.
func findChildID(c *rifx.Chunk, id rifx.ChunkID) *rifx.Chunk {
	for _, ch := range c.Children {
		if ch.ID == id {
			return ch
		}
	}
	return nil
}

// findListByForm returns the first descendant LIST chunk with the given
// FormType (depth-first), or nil.
func findListByForm(c *rifx.Chunk, form rifx.ChunkID) *rifx.Chunk {
	for _, ch := range c.Children {
		if ch.IsList() && ch.FormType == form {
			return ch
		}
		if ch.IsList() {
			if g := findListByForm(ch, form); g != nil {
				return g
			}
		}
	}
	return nil
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
//   - Animated Color: first keyframe value used as static fallback.
//
// Color encoding (V2.2.1 RE, via the stroke tolerance fixture): AE stores
// shape colors as [A,R,G,B] × 255 f64 BE (encodeShapeColorBE), NOT raw
// [r,g,b,a] × 1.0. The pre-V2.2.1 raw encoding produced wrong visible colors.
func lowerFillNode(f *FillNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeFillBody()
	if err != nil {
		return nil, err
	}
	if f.color.mode == StreamModeAnimated && len(f.color.keyframes) > 0 {
		if err := injectAnimatedColor(body, "ADBE Vector Fill Color", f.color.keyframes, ctx); err != nil {
			return nil, err
		}
		return body, nil
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Fill Color", encodeShapeColorBE(f.color.static))
	return body, nil
}

// lowerStrokeNode emits a Stroke graphic body using V2.2.1 embedded tolerance
// bytes (templates/v2_2_shape_stroke_body.bin). Same rationale as the other
// shape kinds — from-scratch emit triggers AE silent-drop; the embedded
// AE-native body carries the full child set (Blend Mode / Composite Order /
// Line Cap / Line Join / Miter Limit + Dashes/Taper/Wave nested groups), and
// we overwrite only the Color/Opacity/Width cdat with runtime values.
//
// V2.2.1 limitations: Blend Mode / Composite Order / Line Cap / Line Join /
// Miter Limit / Dashes / Taper / Wave stay at the embed's defaults; animated
// Color/Opacity/Width use the first keyframe as a static fallback.
func lowerStrokeNode(s *StrokeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeStrokeBody()
	if err != nil {
		return nil, err
	}
	if s.color.mode == StreamModeAnimated && len(s.color.keyframes) > 0 {
		if err := injectAnimatedColor(body, "ADBE Vector Stroke Color", s.color.keyframes, ctx); err != nil {
			return nil, err
		}
	} else {
		overwriteShapeStreamCdat(body, "ADBE Vector Stroke Color", encodeShapeColorBE(s.color.static))
	}

	op := s.opacity.static
	if s.opacity.mode == StreamModeAnimated && len(s.opacity.keyframes) > 0 {
		op = s.opacity.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Opacity", encodeF64sBE(op))

	w := s.width.static
	if s.width.mode == StreamModeAnimated && len(s.width.keyframes) > 0 {
		w = s.width.keyframes[0].Value
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Width", encodeF64sBE(w))
	return body, nil
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
