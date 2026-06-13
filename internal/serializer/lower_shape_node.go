// internal/aep/lower_shape_node.go
//
// 5 per-node lowering funcs (Rect / Ellipse / Path / Fill / Stroke) +
// lowerVectorGroup for the Root Vectors Group container.
//
// Match-names are from AE-saved fixture observations. V2.2 emits ALL sub-props
// even when default-valued (AE elides; we don't — see lower_property_stream.go
// preamble); AE accepts the non-elided form.
package serializer

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
)

// Tolerance shape body bytes used as boilerplate skeleton. Transplant tests
// proved each shape primitive body in our from-scratch emit triggers AE silent
// drop independently. The embed-tolerance-bytes approach (same pattern as
// lowerLayerTransform) bypasses byte-level RE.
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

//go:embed templates/v2_2_shape_stroke_dashed_body.bin
var v22ShapeStrokeDashedBodyBytes []byte

//go:embed templates/v2_2_shape_gradfill_body.bin
var v22ShapeGradFillBodyBytes []byte

//go:embed templates/v2_2_shape_gradstroke_body.bin
var v22ShapeGradStrokeBodyBytes []byte

//go:embed templates/v2_2_shape_trim_body.bin
var v22ShapeTrimBodyBytes []byte

//go:embed templates/v2_2_shape_repeater_body.bin
var v22ShapeRepeaterBodyBytes []byte

//go:embed templates/v2_2_shape_roundcorners_body.bin
var v22ShapeRoundCornersBodyBytes []byte

//go:embed templates/v2_2_shape_offset_body.bin
var v22ShapeOffsetBodyBytes []byte

//go:embed templates/v2_2_shape_merge_body.bin
var v22ShapeMergeBodyBytes []byte

//go:embed templates/v2_2_shape_zigzag_body.bin
var v22ShapeZigZagBodyBytes []byte

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

	v22ShapeStrokeDashedOnce  sync.Once
	v22ShapeStrokeDashedCache *rifx.Chunk
	v22ShapeStrokeDashedErr   error

	v22ShapeGradFillOnce  sync.Once
	v22ShapeGradFillCache *rifx.Chunk
	v22ShapeGradFillErr   error

	v22ShapeGradStrokeOnce  sync.Once
	v22ShapeGradStrokeCache *rifx.Chunk
	v22ShapeGradStrokeErr   error

	v22ShapeTrimOnce  sync.Once
	v22ShapeTrimCache *rifx.Chunk
	v22ShapeTrimErr   error

	v22ShapeRepeaterOnce  sync.Once
	v22ShapeRepeaterCache *rifx.Chunk
	v22ShapeRepeaterErr   error

	v22ShapeRoundCornersOnce  sync.Once
	v22ShapeRoundCornersCache *rifx.Chunk
	v22ShapeRoundCornersErr   error

	v22ShapeOffsetOnce  sync.Once
	v22ShapeOffsetCache *rifx.Chunk
	v22ShapeOffsetErr   error

	v22ShapeMergeOnce  sync.Once
	v22ShapeMergeCache *rifx.Chunk
	v22ShapeMergeErr   error

	v22ShapeZigZagOnce  sync.Once
	v22ShapeZigZagCache *rifx.Chunk
	v22ShapeZigZagErr   error
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

// cloneShapeStrokeDashedBody returns a clone of the dashed-stroke template — a
// superset of the solid stroke body that additionally carries the Dashes group's
// Dash 1 / Gap 1 slots. lowerStrokeNode selects it when dashes are enabled.
func cloneShapeStrokeDashedBody() (*rifx.Chunk, error) {
	v22ShapeStrokeDashedOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeStrokeDashedBodyBytes))
		if err != nil {
			v22ShapeStrokeDashedErr = fmt.Errorf("parse v22ShapeStrokeDashedBodyBytes: %w", err)
			return
		}
		v22ShapeStrokeDashedCache = ch
	})
	if v22ShapeStrokeDashedErr != nil {
		return nil, v22ShapeStrokeDashedErr
	}
	return cloneChunk(v22ShapeStrokeDashedCache), nil
}

// cloneShapeGradFillBody returns a clone of the gradient-fill template. The
// body carries only `ADBE Vector Grad Colors` (GCst→GCky→Utf8 stops XML); Grad
// Type / Start Pt / End Pt were default in the source fixture and AE elided
// them, so the ramp geometry is not overwritable (default linear on open).
func cloneShapeGradFillBody() (*rifx.Chunk, error) {
	v22ShapeGradFillOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeGradFillBodyBytes))
		if err != nil {
			v22ShapeGradFillErr = fmt.Errorf("parse v22ShapeGradFillBodyBytes: %w", err)
			return
		}
		v22ShapeGradFillCache = ch
	})
	if v22ShapeGradFillErr != nil {
		return nil, v22ShapeGradFillErr
	}
	return cloneChunk(v22ShapeGradFillCache), nil
}

// encodeShapeColorBE returns the AE shape-color cdat bytes for an [r,g,b,a]
// (0..1) color: AE stores colors as [A,R,G,B] × 255 as f64 BE (RE'd from the
// stroke tolerance fixture — JSX [0,0,1,1] → disk [255,0,0,255]). Both Stroke
// and Fill use it.
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
// table is serializer-only; runtime API uses the Go enum. Strings match
// AE-saved fixture observations.
var shapeMatchNames = map[ShapeNodeKind]string{
	ShapeKindRect:         "ADBE Vector Shape - Rect",
	ShapeKindEllipse:      "ADBE Vector Shape - Ellipse",
	ShapeKindPath:         "ADBE Vector Shape - Group",
	ShapeKindFill:         "ADBE Vector Graphic - Fill",
	ShapeKindStroke:       "ADBE Vector Graphic - Stroke",
	ShapeKindGroup:          "ADBE Vector Group",
	ShapeKindGradientFill:   "ADBE Vector Graphic - G-Fill",
	ShapeKindGradientStroke: "ADBE Vector Graphic - G-Stroke",
	ShapeKindTrim:           "ADBE Vector Filter - Trim",
	ShapeKindRepeater:       "ADBE Vector Filter - Repeater",
	ShapeKindRoundCorners:   "ADBE Vector Filter - RC",
	ShapeKindOffsetPaths:    "ADBE Vector Filter - Offset",
	ShapeKindMergePaths:     "ADBE Vector Filter - Merge",
	ShapeKindZigZag:         "ADBE Vector Filter - Zigzag",
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
	case *GradientFillNode:
		return lowerGradientFillNode(node, ctx)
	case *GradientStrokeNode:
		return lowerGradientStrokeNode(node, ctx)
	case *TrimNode:
		return lowerTrimNode(node, ctx)
	case *RepeaterNode:
		return lowerRepeaterNode(node, ctx)
	case *RoundCornersNode:
		return lowerRoundCornersNode(node, ctx)
	case *OffsetPathsNode:
		return lowerOffsetPathsNode(node, ctx)
	case *MergePathsNode:
		return lowerMergePathsNode(node, ctx)
	case *ZigZagNode:
		return lowerZigZagNode(node, ctx)
	default:
		return nil, fmt.Errorf("lowerShapeNode: unsupported kind %v", n.Kind())
	}
}

// lowerRectNode emits a Rect shape body using embedded tolerance bytes
// (templates/v2_2_shape_rect_body.bin). From-scratch construction triggered
// silent drop (transplant-isolated); embedding the canonical body +
// overwriting Size cdat with runtime user values is the validator-safe path.
//
// Limitations:
//   - Rect Position / Roundness: runtime-only, NOT persisted (tolerance
//     elides them; embedded body has no slot to overwrite).
//   - Rect Direction: AE default ("ToTheRight"), no runtime customization.
//   - Animated Size: persisted as keyframes (cdat→LIST(list) inject).
func lowerRectNode(r *RectNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeRectBody()
	if err != nil {
		return nil, err
	}
	// Size — non-spatial Vec2 (bpk-88). Position — spatial Vec2 motion-path
	// (bpk-104, value@0x38; identical layout to Ellipse Position). Roundness —
	// 1D non-spatial (bpk-48). Position/Roundness persisted via the richer rect
	// body template (Size/Position/Roundness all cdat slots).
	if err := lowerShapeVec2(body, "ADBE Vector Rect Size", r.Size(), ctx, valueLayout{dim: 2, headerByte: 0x00, spatial: false}); err != nil {
		return nil, err
	}
	if err := lowerShapeVec2(body, "ADBE Vector Rect Position", r.Position(), ctx, valueLayout{dim: 2, headerByte: 0x07, spatial: true, motionPath: true}); err != nil {
		return nil, err
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Shape Direction", encodeF64sBE(float64(r.Direction())))
	if err := lowerShapeScalar(body, "ADBE Vector Rect Roundness", r.Roundness(), ctx); err != nil {
		return nil, err
	}
	return body, nil
}

// lowerShapeVec2 persists a shape Vec2 stream into an embedded body: animated →
// inject the keyframe container (using the supplied layout); static → overwrite
// the cdat with the 2 × f64 value.
func lowerShapeVec2(body *rifx.Chunk, name string, ps *codec.PropertyStream[[2]float64], ctx *lowerCtx, layout valueLayout) error {
	if ps.Mode() == codec.StreamModeAnimated && ps.HasKeyframes() {
		return injectAnimatedVec2L(body, name, ps.Keyframes(), ctx, layout)
	}
	sv, _ := ps.StaticValue()
	overwriteShapeStreamCdat(body, name, encodeF64sBE(sv[0], sv[1]))
	return nil
}

// lowerShapeScalar persists a shape 1D stream (e.g. Rect Roundness): animated →
// inject a 1D non-spatial keyframe container (bpk-48); static → overwrite cdat.
func lowerShapeScalar(body *rifx.Chunk, name string, ps *codec.PropertyStream[float64], ctx *lowerCtx) error {
	if ps.Mode() == codec.StreamModeAnimated && ps.HasKeyframes() {
		kfList, err := encodeKeyframes(ps.Keyframes(), valueLayout{dim: 1, headerByte: 0x00, spatial: false}, encode1D, ctx)
		if err != nil {
			return err
		}
		return injectAnimatedStream(body, name, kfList)
	}
	sv, _ := ps.StaticValue()
	overwriteShapeStreamCdat(body, name, encode1D(sv))
	return nil
}

// injectAnimatedVec2 converts a static shape Vec2 stream in an embedded body
// into an animated one: it finds the tdmn matching streamName, descends into
// the following LIST(tdbs), and replaces the static cdat child with the
// animated LIST(list)(lhd3+ldat) keyframe container (AE keeps tdsb/tdsn/tdb4/
// tdum/tduM unchanged — only cdat ↔ LIST(list) flips). Non-spatial dim-2
// layout (header07=0x00) per the kf RE fixture (Rect/Ellipse Size).
func injectAnimatedVec2(body *rifx.Chunk, streamName string, kfs []codec.StreamKeyframe[[2]float64], ctx *lowerCtx) error {
	return injectAnimatedVec2L(body, streamName, kfs, ctx, valueLayout{dim: 2, headerByte: 0x00, spatial: false})
}

// injectAnimatedVec2L injects an animated Vec2 stream with an explicit layout
// (Size = non-spatial header07=0x00 bpk 88; shape Position = spatial header07=
// 0x07 bpk 104 value@0x38).
func injectAnimatedVec2L(body *rifx.Chunk, streamName string, kfs []codec.StreamKeyframe[[2]float64], ctx *lowerCtx, layout valueLayout) error {
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
func injectAnimatedColor(body *rifx.Chunk, streamName string, kfs []codec.StreamKeyframe[[4]float64], ctx *lowerCtx) error {
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
			// tdb4 static→animated flags (RE'd across Rect Size + Fill Color):
			// @0x05 clear bit0, @0x44 = 0x01, @0x4f clear bit0. Without this AE
			// expects a cdat per the static tdb4 and reports "file data missing".
			// NB: modern AE writes lowercase "tdb4"; rifx.IDTdb4 is the legacy
			// UPPERCASE "Tdb4" (chunk IDs are case-sensitive), so match the
			// lowercase literal.
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

// lowerEllipseNode emits an Ellipse shape body using embedded tolerance
// bytes (templates/v2_2_shape_ellipse_body.bin). Same rationale as
// lowerRectNode — from-scratch emit triggered AE silent-drop (the body's
// boilerplate, not the values, fails AE's semantic validation); embedding the
// canonical AE-saved body + overwriting Size/Position cdat with runtime values
// is the validator-safe path.
//
// Limitations:
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
	if e.Size().Mode() == codec.StreamModeAnimated && e.Size().HasKeyframes() {
		if err := injectAnimatedVec2(body, "ADBE Vector Ellipse Size", e.Size().Keyframes(), ctx); err != nil {
			return nil, err
		}
	} else {
		sv, _ := e.Size().StaticValue()
		overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Size", encodeF64sBE(sv[0], sv[1]))
	}
	if e.Position().Mode() == codec.StreamModeAnimated && e.Position().HasKeyframes() {
		if err := injectAnimatedVec2L(body, "ADBE Vector Ellipse Position", e.Position().Keyframes(), ctx,
			valueLayout{dim: 2, headerByte: 0x07, spatial: true, motionPath: true}); err != nil {
			return nil, err
		}
	} else {
		pv, _ := e.Position().StaticValue()
		overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Position", encodeF64sBE(pv[0], pv[1]))
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Shape Direction", encodeF64sBE(float64(e.Direction())))
	return body, nil
}

// lowerPathNode emits a Path shape body using embedded tolerance bytes
// (templates/v2_2_shape_path_body.bin). From-scratch emit CRASHED AE 2020
// ("After Effects 已崩溃 (0::42)") — the om-s/tdb4 scaffolding is too fragile
// to hand-build. We clone the AE-native body and splice in the user's geometry
// (shph/lhd3/ldat from encodeBezier, whose layout matches AE byte-for-byte),
// keeping AE's exact scaffolding (om-s header + omks + omtn).
//
// A path with ≥2 keyframes is lowered as an animated shape path (N shaps + a
// tdbs time table, == animated mask path — see spliceAnimatedPath). ≤1
// keyframe stays a single static snapshot. Linear interp only (SetVertices
// zeroes tangents; temporal ease is V2.3+).
func lowerPathNode(p *PathNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapePathBody()
	if err != nil {
		return nil, err
	}
	if p.Path().Mode() == codec.StreamModeAnimated && len(p.Path().Keyframes()) >= 2 {
		if err := spliceAnimatedPath(body, p.Path().Keyframes(), ctx); err != nil {
			return nil, err
		}
		return body, nil
	}
	bp, _ := p.Path().StaticValue()
	if p.Path().Mode() == codec.StreamModeAnimated && len(p.Path().Keyframes()) > 0 {
		bp = p.Path().Keyframes()[0].Value
	}
	if err := splicePathGeometry(body, bp); err != nil {
		return nil, err
	}
	return body, nil
}

// splicePathGeometry replaces the shph/lhd3/ldat geometry chunks inside the
// embedded path body's (single) LIST(shap) with freshly-encoded geometry for
// bp, leaving AE's scaffolding (om-s header, omks/shap/kfl wrappers, omtn)
// intact. Static path path.
func splicePathGeometry(body *rifx.Chunk, bp BezierPath) error {
	shap := findListByForm(body, rifx.IDShap)
	if shap == nil {
		return fmt.Errorf("splicePathGeometry: LIST(shap) not found in embed body")
	}
	return spliceShapGeometry(shap, bp)
}

// spliceShapGeometry replaces shph/lhd3/ldat inside one LIST(shap) with freshly
// encoded geometry for bp, preserving the shap's (shph, kfl, omtn) sibling
// order and (lhd3, ldat) within kfl.
func spliceShapGeometry(shap *rifx.Chunk, bp BezierPath) error {
	kfl := findListByForm(shap, rifx.IDkfl)
	if kfl == nil {
		return fmt.Errorf("spliceShapGeometry: LIST(kfl) not found in shap")
	}
	newShph, newLhd3, newLdat := encodeBezier(bp)
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

// spliceAnimatedPath converts the embedded static path body into an animated
// shape path, byte-matching AE's own output (re_path_anim.aep; identical to an
// animated MASK path). Two edits inside the cloned om-s, reusing all of AE's
// scaffolding (om-s wrapper + its omtn, the tdbs's tdsb/tdsn, each shap's
// kfl/omtn):
//
//  1. The value tdbs (tdsb + tdsn + tdb4 + cdat) becomes a TIME-table tdbs:
//     tdb4 is KEPT with its static→animated flags patched (@0x05/@0x44/@0x4f,
//     same as injectAnimatedStream), only the cdat is dropped and a
//     LIST(kfl){lhd3, ldat} time table appended (one 64B block per keyframe —
//     see encodePathTimeTable). Byte-verified against re_path_anim.aep: the
//     animated time-table tdbs DOES carry a tdb4 (an earlier RE note claiming
//     "no tdb4" was a misread).
//  2. The single shap in omks is replaced by one shap per keyframe, each
//     spliced with that frame's geometry (encodeBezier, bbox-normalized).
//
// See incident-reports/path-keyframe-write-re.md.
func spliceAnimatedPath(body *rifx.Chunk, kfs []codec.StreamKeyframe[BezierPath], ctx *lowerCtx) error {
	oms := findListByForm(body, rifx.IDOmS)
	if oms == nil {
		return fmt.Errorf("spliceAnimatedPath: LIST(om-s) not found in embed body")
	}
	var tdbs, omks *rifx.Chunk
	for _, ch := range oms.Children {
		if !ch.IsList() {
			continue
		}
		switch ch.FormType {
		case rifx.IDTdbs:
			tdbs = ch
		case rifx.IDOmks:
			omks = ch
		}
	}
	if tdbs == nil || omks == nil {
		return fmt.Errorf("spliceAnimatedPath: om-s missing tdbs/omks")
	}

	// 1. value tdbs → time-table tdbs: keep tdsb/tdsn/tdb4 (patch tdb4's
	// static→animated flags), drop only cdat, append the kfl time table.
	// (Match tdb4 by literal — chunk IDs are case-sensitive and modern AE
	// writes lowercase "tdb4".)
	kept := tdbs.Children[:0:0]
	for _, ch := range tdbs.Children {
		if ch.ID == rifx.IDCdat {
			continue
		}
		if id := string(ch.ID[:]); id == "tdb4" || id == "Tdb4" {
			if len(ch.Data) > 0x4f {
				ch.Data[0x05] &^= 0x01
				ch.Data[0x44] = 0x01
				ch.Data[0x4f] &^= 0x01
			}
		}
		kept = append(kept, ch)
	}
	tdbs.Children = append(kept, encodePathTimeTable(kfs, ctx))

	// 2. omks single shap → N shaps. The embed's lone shap is the structural
	// prototype (shph + kfl + omtn); clone it per keyframe and splice geometry.
	var proto *rifx.Chunk
	for _, ch := range omks.Children {
		if ch.IsList() && ch.FormType == rifx.IDShap {
			proto = ch
			break
		}
	}
	if proto == nil {
		return fmt.Errorf("spliceAnimatedPath: omks has no prototype shap")
	}
	shaps := make([]*rifx.Chunk, 0, len(kfs))
	for _, kf := range kfs {
		s := cloneChunk(proto)
		if err := spliceShapGeometry(s, kf.Value); err != nil {
			return err
		}
		shaps = append(shaps, s)
	}
	omks.Children = shaps
	return nil
}

// encodePathTimeTable builds the LIST(kfl){lhd3, ldat} keyframe TIME table for
// an animated shape path: one 64-byte block per keyframe (bpk = 64), inverse of
// readMaskPathTimes (parse_mask.go) and byte-matched to AE's re_path_anim.aep.
//
// Per block (byte-verified against re_path_anim.aep): time ticks @0x00
// (round(sec * tickRate)); linear interp @0x04/0x05; a constant 0x01 @0x07 and
// u32 0x00000002 @0x08 on EVERY block; a 1.0 f64 @0x10 on every keyframe except
// the last; an 8-byte runtime-pointer trailer @0x38 that AE recomputes on load
// (we leave it zero for deterministic output). lhd3 mirrors encodeKeyframes'
// header constants with count@0x08 = #kf and bpk@0x10 = 64. (An earlier RE note
// placing the 1.0 at @0x30 and 0x02 at @0x07-of-first-block was a misread.)
func encodePathTimeTable(kfs []codec.StreamKeyframe[BezierPath], ctx *lowerCtx) *rifx.Chunk {
	const bpk = 64
	n := len(kfs)

	lhd3 := make([]byte, 52)
	lhd3[0], lhd3[1], lhd3[2], lhd3[3] = 0x00, 0xd0, 0x0b, 0xee
	binary.BigEndian.PutUint32(lhd3[0x08:0x0C], uint32(n))
	binary.BigEndian.PutUint32(lhd3[0x0C:0x10], 1)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], bpk)
	binary.BigEndian.PutUint32(lhd3[0x14:0x18], 4)
	binary.BigEndian.PutUint32(lhd3[0x18:0x1C], 1)
	binary.BigEndian.PutUint32(lhd3[0x1C:0x20], 4)

	tickRate := ctx.tickRate
	if tickRate <= 0 {
		tickRate = 30720
	}
	ldat := make([]byte, n*bpk)
	for i, kf := range kfs {
		blk := ldat[i*bpk : (i+1)*bpk]
		binary.BigEndian.PutUint32(blk[0x00:0x04], uint32(math.Round(kf.Time*tickRate)))
		blk[0x04] = byte(InterpLinear)
		blk[0x05] = byte(InterpLinear)
		blk[0x07] = 0x01
		binary.BigEndian.PutUint32(blk[0x08:0x0C], 2)
		if i != n-1 {
			binary.BigEndian.PutUint64(blk[0x10:0x18], math.Float64bits(1.0))
		}
		// @0x38 trailer left zero (AE runtime cache pointer; rebuilt on load).
	}

	kfl := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
	kfl.Children = append(kfl.Children,
		&rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3},
		&rifx.Chunk{ID: rifx.IDLdat, Data: ldat},
	)
	return kfl
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

// lowerFillNode emits a Fill graphic body using embedded tolerance bytes
// (templates/v2_2_shape_fill_body.bin). Same rationale as lowerRectNode —
// transplant tests proved from-scratch Fill body triggers silent drop;
// embedded canonical body + cdat overwrite for Color values is the
// validator-safe path.
//
// Fill Color (static + keyframe) and Fill Opacity (static + keyframe, raw %,
// 1D non-spatial bpk-48, richer fill body template) persist. Blend Mode /
// Composite Order / Fill Rule stay at the embed defaults (no runtime setter).
//
// Color encoding (RE'd via the stroke tolerance fixture): AE stores shape
// colors as [A,R,G,B] × 255 f64 BE (encodeShapeColorBE), NOT raw
// [r,g,b,a] × 1.0. The raw encoding produced wrong visible colors.
func lowerFillNode(f *FillNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeFillBody()
	if err != nil {
		return nil, err
	}
	if f.Color().Mode() == codec.StreamModeAnimated && f.Color().HasKeyframes() {
		if err := injectAnimatedColor(body, "ADBE Vector Fill Color", f.Color().Keyframes(), ctx); err != nil {
			return nil, err
		}
	} else {
		cv, _ := f.Color().StaticValue()
		overwriteShapeStreamCdat(body, "ADBE Vector Fill Color", encodeShapeColorBE(cv))
	}
	// Opacity (raw %) — 1D non-spatial (bpk-48, value@0x08). The richer fill
	// body template carries an Opacity cdat slot (default 100 was elided), so
	// static/animated Opacity persists.
	if err := lowerShapeScalar(body, "ADBE Vector Fill Opacity", f.Opacity(), ctx); err != nil {
		return nil, err
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Blend Mode", encodeF64sBE(float64(f.BlendMode())))
	overwriteShapeStreamCdat(body, "ADBE Vector Composite Order", encodeF64sBE(float64(f.CompositeOrder())))
	overwriteShapeStreamCdat(body, "ADBE Vector Fill Rule", encodeF64sBE(float64(f.FillRule())))
	return body, nil
}

// lowerGradientStops overwrites the Grad Colors stops XML in a gradient body
// (G-Fill or G-Stroke) with the runtime gradient. No-op if gradient is nil.
// Shared by lowerGradientFillNode / lowerGradientStrokeNode.
func lowerGradientStops(body *rifx.Chunk, gradient *codec.Gradient) *rifx.Chunk {
	if gradient != nil {
		overwriteGradientStopsXML(body, "ADBE Vector Grad Colors", codec.EncodeGradientXML(gradient))
	}
	return body
}

// lowerGradientFillNode emits a gradient-fill graphic body from the embedded
// template (templates/v2_2_shape_gradfill_body.bin), overwriting the Grad
// Start/End Pt cdats (the linear ramp direction) + the Grad Colors stops XML
// with the runtime gradient. The XML length changes per stop count →
// length-variable; the Utf8 chunk's Data is swapped and rifx.Chunk.Write
// recomputes the enclosing GCky / GCst / tdgp LIST sizes on serialization.
//
// Start/End Pt are Vec2 (2 × f64 BE at cdat[0:16], same layout as the Repeater
// Transform points), overwritten with the node's StartPoint/EndPoint (default
// [0,0]→[100,0] = AE's horizontal ramp, so a gradient that doesn't set direction
// reproduces the pre-direction behavior). Grad Type / HiLite stay at the embed
// default (linear).
func lowerGradientFillNode(n *GradientFillNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeGradFillBody()
	if err != nil {
		return nil, err
	}
	sp, ep := n.StartPoint(), n.EndPoint()
	overwriteShapeStreamCdat(body, "ADBE Vector Grad Start Pt", encodeF64sBE(sp[0], sp[1]))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad End Pt", encodeF64sBE(ep[0], ep[1]))
	return lowerGradientStops(body, n.Gradient()), nil
}

// cloneShapeGradStrokeBody returns a clone of the gradient-stroke template
// (templates/v2_2_shape_gradstroke_body.bin). Like gradfill it carries only
// `ADBE Vector Grad Colors`; stroke geometry stays at the extracted values.
func cloneShapeGradStrokeBody() (*rifx.Chunk, error) {
	v22ShapeGradStrokeOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeGradStrokeBodyBytes))
		if err != nil {
			v22ShapeGradStrokeErr = fmt.Errorf("parse v22ShapeGradStrokeBodyBytes: %w", err)
			return
		}
		v22ShapeGradStrokeCache = ch
	})
	if v22ShapeGradStrokeErr != nil {
		return nil, v22ShapeGradStrokeErr
	}
	return cloneChunk(v22ShapeGradStrokeCache), nil
}

// lowerGradientStrokeNode emits a gradient-stroke body from the embedded
// template, overwriting the Grad Colors stops XML. Body logic is identical to
// lowerGradientFillNode; only the cloned template differs.
func lowerGradientStrokeNode(n *GradientStrokeNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeGradStrokeBody()
	if err != nil {
		return nil, err
	}
	return lowerGradientStops(body, n.Gradient()), nil
}

// cloneShapeTrimBody returns a clone of the Trim Paths template
// (templates/v2_2_shape_trim_body.bin). The body carries the Start / End /
// Offset cdat slots (Trim Type was AE-default and elided — no slot).
func cloneShapeTrimBody() (*rifx.Chunk, error) {
	v22ShapeTrimOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeTrimBodyBytes))
		if err != nil {
			v22ShapeTrimErr = fmt.Errorf("parse v22ShapeTrimBodyBytes: %w", err)
			return
		}
		v22ShapeTrimCache = ch
	})
	if v22ShapeTrimErr != nil {
		return nil, v22ShapeTrimErr
	}
	return cloneChunk(v22ShapeTrimCache), nil
}

// lowerTrimNode emits a Trim Paths filter body from the embedded template,
// overwriting the Start / End / Offset cdat slots with runtime values. Start /
// End are raw percentages (0..100), Offset is raw degrees — all float64 BE at
// cdat[0:8], identical to the other shape scalars (RE'd from v2_2_trim.aep).
//
// Static → cdat overwrite; animated → the cdat flips to a 1D non-spatial
// keyframe container (same injectAnimatedStream path as Rect Roundness / Fill
// Opacity), so the line-draw reveal (keyframed End 0→100) persists.
//
// Trim Type (Simultaneously/Individually) stays at the embed default — AE
// elided it in the source fixture so there is no slot to overwrite.
func lowerTrimNode(n *TrimNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeTrimBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Trim Start", n.Start(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Trim End", n.End(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Trim Offset", n.Offset(), ctx); err != nil {
		return nil, err
	}
	return body, nil
}

// cloneShapeRepeaterBody returns a clone of the Repeater template
// (templates/v2_2_shape_repeater_body.bin): top-level Copies/Offset cdat slots
// + a nested `ADBE Vector Repeater Transform` group (Anchor/Position/Scale/
// Rotation/Opacity 1·2). Order (Composite) was AE-default and elided.
func cloneShapeRepeaterBody() (*rifx.Chunk, error) {
	v22ShapeRepeaterOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeRepeaterBodyBytes))
		if err != nil {
			v22ShapeRepeaterErr = fmt.Errorf("parse v22ShapeRepeaterBodyBytes: %w", err)
			return
		}
		v22ShapeRepeaterCache = ch
	})
	if v22ShapeRepeaterErr != nil {
		return nil, v22ShapeRepeaterErr
	}
	return cloneChunk(v22ShapeRepeaterCache), nil
}

// lowerRepeaterNode emits a Repeater filter body from the embedded template.
// Top-level Copies/Offset are 1D scalars (static cdat overwrite / animated flip
// via lowerShapeScalar). The per-copy Transform sub-streams live in the nested
// `ADBE Vector Repeater Transform` group (descend via findGroupBody) and are
// static cdat overwrites (Anchor/Position/Scale are Vec2 at cdat[0:16],
// Rotation/Opacity 1·2 are 1D at cdat[0:8]) — same mechanism as Stroke
// Taper/Wave. RE'd from v2_2_repeater.aep.
func lowerRepeaterNode(n *RepeaterNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeRepeaterBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Repeater Copies", n.Copies(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Repeater Offset", n.Offset(), ctx); err != nil {
		return nil, err
	}
	if xf := findGroupBody(body, "ADBE Vector Repeater Transform"); xf != nil {
		t := n.Transform()
		a, p, s := t.Anchor(), t.Position(), t.Scale()
		overwriteShapeStreamCdat(xf, "ADBE Vector Repeater Anchor", encodeF64sBE(a[0], a[1]))
		overwriteShapeStreamCdat(xf, "ADBE Vector Repeater Position", encodeF64sBE(p[0], p[1]))
		overwriteShapeStreamCdat(xf, "ADBE Vector Repeater Scale", encodeF64sBE(s[0], s[1]))
		overwriteShapeStreamCdat(xf, "ADBE Vector Repeater Rotation", encodeF64sBE(t.Rotation()))
		overwriteShapeStreamCdat(xf, "ADBE Vector Repeater Opacity 1", encodeF64sBE(t.StartOpacity()))
		overwriteShapeStreamCdat(xf, "ADBE Vector Repeater Opacity 2", encodeF64sBE(t.EndOpacity()))
	}
	return body, nil
}

// cloneShapeRoundCornersBody returns a clone of the Round Corners template
// (templates/v2_2_shape_roundcorners_body.bin): a single `ADBE Vector
// RoundCorner Radius` cdat slot (Radius was set non-default in the fixture so
// AE emitted it).
func cloneShapeRoundCornersBody() (*rifx.Chunk, error) {
	v22ShapeRoundCornersOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeRoundCornersBodyBytes))
		if err != nil {
			v22ShapeRoundCornersErr = fmt.Errorf("parse v22ShapeRoundCornersBodyBytes: %w", err)
			return
		}
		v22ShapeRoundCornersCache = ch
	})
	if v22ShapeRoundCornersErr != nil {
		return nil, v22ShapeRoundCornersErr
	}
	return cloneChunk(v22ShapeRoundCornersCache), nil
}

// lowerRoundCornersNode emits a Round Corners filter body from the embedded
// template, overwriting the single `ADBE Vector RoundCorner Radius` cdat (1D
// f64 BE at cdat[0:8], raw pixels — same scalar layout as Trim Start/End).
// Static → cdat overwrite; animated → the cdat flips to a 1D non-spatial
// keyframe container via the shared injectAnimatedStream path.
func lowerRoundCornersNode(n *RoundCornersNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeRoundCornersBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector RoundCorner Radius", n.Radius(), ctx); err != nil {
		return nil, err
	}
	return body, nil
}

// cloneShapeOffsetBody returns a clone of the Offset Paths template
// (templates/v2_2_shape_offset_body.bin): a single `ADBE Vector Offset Amount`
// cdat slot (Amount set non-default in the fixture so AE emitted it; Line Join /
// Miter Limit / Copies / Copy Offset stayed default and are elided).
func cloneShapeOffsetBody() (*rifx.Chunk, error) {
	v22ShapeOffsetOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeOffsetBodyBytes))
		if err != nil {
			v22ShapeOffsetErr = fmt.Errorf("parse v22ShapeOffsetBodyBytes: %w", err)
			return
		}
		v22ShapeOffsetCache = ch
	})
	if v22ShapeOffsetErr != nil {
		return nil, v22ShapeOffsetErr
	}
	return cloneChunk(v22ShapeOffsetCache), nil
}

// lowerOffsetPathsNode emits an Offset Paths filter body from the embedded
// template, overwriting the single `ADBE Vector Offset Amount` cdat (1D f64 BE
// at cdat[0:8], raw pixels — same scalar layout as Round Corners Radius).
// Static → cdat overwrite; animated → the cdat flips to a 1D non-spatial
// keyframe container via the shared injectAnimatedStream path.
func lowerOffsetPathsNode(n *OffsetPathsNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeOffsetBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Offset Amount", n.Amount(), ctx); err != nil {
		return nil, err
	}
	return body, nil
}

// cloneShapeMergeBody returns a clone of the Merge Paths template
// (templates/v2_2_shape_merge_body.bin): a single `ADBE Vector Merge Type` enum
// cdat slot (Type set non-default in the fixture so AE emitted it).
func cloneShapeMergeBody() (*rifx.Chunk, error) {
	v22ShapeMergeOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeMergeBodyBytes))
		if err != nil {
			v22ShapeMergeErr = fmt.Errorf("parse v22ShapeMergeBodyBytes: %w", err)
			return
		}
		v22ShapeMergeCache = ch
	})
	if v22ShapeMergeErr != nil {
		return nil, v22ShapeMergeErr
	}
	return cloneChunk(v22ShapeMergeCache), nil
}

// lowerMergePathsNode emits a Merge Paths filter body from the embedded
// template, overwriting the single `ADBE Vector Merge Type` cdat (1D f64 BE enum
// at cdat[0:8], 1-based index — same layout as the Fill/Stroke enums). Static
// only (the merge mode is not animated).
func lowerMergePathsNode(n *MergePathsNode, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeMergeBody()
	if err != nil {
		return nil, err
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Merge Type", encodeF64sBE(float64(n.Type())))
	return body, nil
}

// cloneShapeZigZagBody returns a clone of the ZigZag template
// (templates/v2_2_shape_zigzag_body.bin): `ADBE Vector Zigzag Size` +
// `ADBE Vector Zigzag Detail` cdat slots (both set non-default in the fixture so
// AE emitted them; the Points enum stayed default and is elided).
func cloneShapeZigZagBody() (*rifx.Chunk, error) {
	v22ShapeZigZagOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22ShapeZigZagBodyBytes))
		if err != nil {
			v22ShapeZigZagErr = fmt.Errorf("parse v22ShapeZigZagBodyBytes: %w", err)
			return
		}
		v22ShapeZigZagCache = ch
	})
	if v22ShapeZigZagErr != nil {
		return nil, v22ShapeZigZagErr
	}
	return cloneChunk(v22ShapeZigZagCache), nil
}

// lowerZigZagNode emits a ZigZag filter body from the embedded template,
// overwriting the `ADBE Vector Zigzag Size` (amplitude) and `ADBE Vector Zigzag
// Detail` (ridges per segment) cdats — both 1D f64 BE at cdat[0:8], same scalar
// layout as the other shape scalars. Static → cdat overwrite; animated → the
// cdat flips to a 1D non-spatial keyframe container via injectAnimatedStream.
func lowerZigZagNode(n *ZigZagNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeZigZagBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Zigzag Size", n.Size(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Zigzag Detail", n.Detail(), ctx); err != nil {
		return nil, err
	}
	return body, nil
}

// overwriteGradientStopsXML finds the tdmn matching streamName inside body,
// descends into the following LIST(GCst) → LIST(GCky), and replaces the first
// Utf8 leaf's Data with xml. No-op if the structure is absent.
func overwriteGradientStopsXML(body *rifx.Chunk, streamName, xml string) {
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || trimChunkNUL(kids[i].Data) != streamName {
			continue
		}
		gcst := kids[i+1]
		if !gcst.IsList() || gcst.FormType != rifx.IDGCst {
			return
		}
		for _, c := range gcst.Children {
			if c.IsList() && c.FormType == rifx.IDGCky {
				for _, u := range c.Children {
					if u.ID == rifx.IDUtf8 {
						u.Data = []byte(xml)
						return
					}
				}
			}
		}
		return
	}
}

// lowerStrokeNode emits a Stroke graphic body using embedded tolerance
// bytes (templates/v2_2_shape_stroke_body.bin). Same rationale as the other
// shape kinds — from-scratch emit triggers AE silent-drop; the embedded
// AE-native body carries the child set (Color / Opacity / Width / Line Cap /
// Line Join / Miter Limit + Dashes/Taper/Wave nested groups), and we overwrite
// only the cdat slots we model with runtime values.
//
// Line Cap / Line Join / Miter Limit are 1D scalars at cdat[0:8] (float64 BE;
// enums store a 1-based index). The template was re-extracted from a fixture
// with all three set non-default so the slots exist to overwrite (AE elides
// defaults). RE: incident-reports/stroke-line-cap-join-miter-re.md.
//
// Limitations: Blend Mode / Composite Order / Dashes / Taper / Wave stay at
// the embed's defaults; animated Color/Opacity/Width use the first keyframe as
// a static fallback. Line Cap / Join / Miter are static-only.
func lowerStrokeNode(s *StrokeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	// Dashes enabled → swap to the dashed template (carries Dash 1 / Gap 1
	// slots); else the solid template (Dashes group is an empty placeholder).
	clone := cloneShapeStrokeBody
	if s.Dashes() != nil && s.Dashes().Enabled() {
		clone = cloneShapeStrokeDashedBody
	}
	body, err := clone()
	if err != nil {
		return nil, err
	}
	if s.Color().Mode() == codec.StreamModeAnimated && s.Color().HasKeyframes() {
		if err := injectAnimatedColor(body, "ADBE Vector Stroke Color", s.Color().Keyframes(), ctx); err != nil {
			return nil, err
		}
	} else {
		cv, _ := s.Color().StaticValue()
		overwriteShapeStreamCdat(body, "ADBE Vector Stroke Color", encodeShapeColorBE(cv))
	}

	// Opacity (raw %) + Width (raw px) — 1D non-spatial scalars (bpk-48,
	// value@0x08, no normalization; RE'd from v2_2_stroke_kf_re.aep). The stroke
	// body template already carries both cdat slots, so animated streams flip in
	// place (previously collapsed to the first keyframe value).
	if err := lowerShapeScalar(body, "ADBE Vector Stroke Opacity", s.Opacity(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Stroke Width", s.Width(), ctx); err != nil {
		return nil, err
	}
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Line Cap", encodeF64sBE(float64(s.LineCap())))
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Line Join", encodeF64sBE(float64(s.LineJoin())))
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Miter Limit", encodeF64sBE(s.MiterLimit()))
	overwriteShapeStreamCdat(body, "ADBE Vector Blend Mode", encodeF64sBE(float64(s.BlendMode())))
	overwriteShapeStreamCdat(body, "ADBE Vector Composite Order", encodeF64sBE(float64(s.CompositeOrder())))
	lowerStrokeTaper(body, s.Taper())
	lowerStrokeWave(body, s.Wave())
	lowerStrokeDashes(body, s.Dashes())
	return body, nil
}

// lowerStrokeDashes overwrites the Dashes group's Dash 1 / Gap 1 cdats inside
// the embedded dashed-stroke body. No-op when dashes are disabled (the solid
// template was cloned and has no Dash/Gap slots to overwrite). Offset is not
// modeled (AE keeps it hidden / script-ungettable — no template slot exists).
func lowerStrokeDashes(strokeBody *rifx.Chunk, d *StrokeDashes) {
	if d == nil || !d.Enabled() {
		return
	}
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Dashes")
	if g == nil {
		return
	}
	overwriteShapeStreamCdat(g, "ADBE Vector Stroke Dash 1", encodeF64sBE(d.Dash()))
	overwriteShapeStreamCdat(g, "ADBE Vector Stroke Gap 1", encodeF64sBE(d.Gap()))
}

// lowerStrokeTaper overwrites the Taper group's %-mode scalar cdats inside the
// embedded stroke body. The body template (re-extracted from a fixture with the
// Taper group's % controls set non-default) carries the 6 always-active slots:
// Start/End Length, Start/End Width, Start/End Ease — each a float64 BE at
// cdat[0:8]. Length Units / StartWidthPx / EndWidthPx are AE-elided in % mode
// and absent from the template (not modeled in V2.2).
func lowerStrokeTaper(strokeBody *rifx.Chunk, t *StrokeTaper) {
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Taper")
	if g == nil || t == nil {
		return
	}
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Start Length", encodeF64sBE(t.StartLength()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper End Length", encodeF64sBE(t.EndLength()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Start Width", encodeF64sBE(t.StartWidth()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper End Width", encodeF64sBE(t.EndWidth()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Start Ease", encodeF64sBE(t.StartEase()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper End Ease", encodeF64sBE(t.EndEase()))
}

// lowerStrokeWave overwrites the Wave group's Wavelength-mode scalar cdats
// (Amount / Wavelength / Phase). Units / Cycles are AE-elided in Wavelength mode
// and absent from the template (not modeled in V2.2).
func lowerStrokeWave(strokeBody *rifx.Chunk, w *StrokeWave) {
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Wave")
	if g == nil || w == nil {
		return
	}
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Wave Amount", encodeF64sBE(w.Amount()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Wavelength", encodeF64sBE(w.Wavelength()))
	overwriteShapeStreamCdat(g, "ADBE Vector Taper Wave Phase", encodeF64sBE(w.Phase()))
}

// findGroupBody returns the LIST(tdgp) group body following the tdmn matching
// groupName among body.Children (one level), or nil. Used to descend into a
// nested shape group (Stroke Taper / Wave) before overwriting its sub-stream
// cdats with overwriteShapeStreamCdat.
func findGroupBody(body *rifx.Chunk, groupName string) *rifx.Chunk {
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == groupName {
			if next := kids[i+1]; next.IsList() && next.FormType == rifx.IDTdgp {
				return next
			}
		}
	}
	return nil
}

// lowerVectorGroup wraps shape-node children into the Root Vectors Group's
// inner LIST(tdgp). The on-disk shape (every AE-saved fixture observed —
// tolerance.aep + re_shapes.aep) is a 5-level nesting:
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
// Flattening everything into the Root Vectors Group body directly made AE 2025
// parse the file without exception but silently drop the layer from
// comp.layers. The wrappers are structural: AE Shape Layer's Contents always
// holds one or more "ADBE Vector Group" entries (each is what UI shows as
// "Group N"), and each Vector Group always carries the 3-child fixed routing
// (Vectors Group for shape kids + Transform + Materials).
//
// V2.2 maps the runtime `shapeRootGroup.Children = [Rect, Fill, ...]` to a
// SINGLE Vector Group wrapper (semantic = AE's auto-created "Group 1"). V2.3+
// may expose multiple user-named groups.
//
// Children render in order: Children[0] = bottom, Children[len-1] = top.
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
