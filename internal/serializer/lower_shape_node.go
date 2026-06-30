// internal/aep/lower_shape_node.go
//
// 5 per-node lowering funcs (Rect / Ellipse / Path / Fill / Stroke) +
// lowerVectorGroup for the Root Vectors Group container.
//
// Match-names are from AE-saved fixture observations. This implementation
// emits ALL sub-props even when default-valued (AE elides; we don't — see
// lower_property_stream.go preamble); AE accepts the non-elided form.
package serializer

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// Tolerance shape body bytes used as boilerplate skeleton. Transplant tests
// proved each shape primitive body in our from-scratch emit triggers AE silent
// drop independently. The embed-tolerance-bytes approach (same pattern as
// lowerLayerTransform) bypasses byte-level RE.
//
//go:embed templates/shapes/rect_body.bin
var shapeRectBodyBytes []byte

//go:embed templates/shapes/fill_body.bin
var shapeFillBodyBytes []byte

//go:embed templates/shapes/ellipse_body.bin
var shapeEllipseBodyBytes []byte

//go:embed templates/shapes/path_body.bin
var shapePathBodyBytes []byte

//go:embed templates/shapes/stroke_body.bin
var shapeStrokeBodyBytes []byte

//go:embed templates/shapes/stroke_dashed_body.bin
var shapeStrokeDashedBodyBytes []byte

//go:embed templates/shapes/gradfill_body.bin
var shapeGradFillBodyBytes []byte

//go:embed templates/shapes/gradstroke_body.bin
var shapeGradStrokeBodyBytes []byte

//go:embed templates/shapes/trim_body.bin
var shapeTrimBodyBytes []byte

//go:embed templates/shapes/trim_type_leaf.bin
var shapeTrimTypeLeafBytes []byte

//go:embed templates/shapes/repeater_body.bin
var shapeRepeaterBodyBytes []byte

//go:embed templates/shapes/repeater_order_leaf.bin
var shapeRepeaterOrderLeafBytes []byte

//go:embed templates/shapes/roundcorners_body.bin
var shapeRoundCornersBodyBytes []byte

//go:embed templates/shapes/offset_body.bin
var shapeOffsetBodyBytes []byte

//go:embed templates/shapes/offset_copies_leaf.bin
var shapeOffsetCopiesLeafBytes []byte

//go:embed templates/shapes/offset_linejoin_leaf.bin
var shapeOffsetLineJoinLeafBytes []byte

//go:embed templates/shapes/offset_miter_leaf.bin
var shapeOffsetMiterLeafBytes []byte

//go:embed templates/shapes/offset_copyoffset_leaf.bin
var shapeOffsetCopyOffsetLeafBytes []byte

//go:embed templates/shapes/merge_body.bin
var shapeMergeBodyBytes []byte

//go:embed templates/shapes/zigzag_body.bin
var shapeZigZagBodyBytes []byte

//go:embed templates/shapes/zigzag_points_leaf.bin
var shapeZigZagPointsLeafBytes []byte

//go:embed templates/shapes/star_body.bin
var shapeStarBodyBytes []byte

//go:embed templates/shapes/starpolygon_body.bin
var shapeStarPolygonBodyBytes []byte

//go:embed templates/shapes/puckerbloat_body.bin
var shapePuckerBloatBodyBytes []byte

//go:embed templates/shapes/twist_body.bin
var shapeTwistBodyBytes []byte

//go:embed templates/shapes/twist_center_leaf.bin
var shapeTwistCenterLeafBytes []byte

//go:embed templates/shapes/wiggle_body.bin
var shapeWiggleBodyBytes []byte

//go:embed templates/shapes/wiggletransform_body.bin
var shapeWiggleTransformBodyBytes []byte

//go:embed templates/shapes/roughen_points_leaf.bin
var shapeRoughenPointsLeafBytes []byte

//go:embed templates/shapes/correlation_leaf.bin
var shapeCorrelationLeafBytes []byte

//go:embed templates/shapes/temporal_phase_leaf.bin
var shapeTemporalPhaseLeafBytes []byte

//go:embed templates/shapes/spatial_phase_leaf.bin
var shapeSpatialPhaseLeafBytes []byte

var (
	shapeRectOnce  sync.Once
	shapeRectCache *rifx.Chunk
	shapeRectErr   error

	shapeFillOnce  sync.Once
	shapeFillCache *rifx.Chunk
	shapeFillErr   error

	shapeEllipseOnce  sync.Once
	shapeEllipseCache *rifx.Chunk
	shapeEllipseErr   error

	shapePathOnce  sync.Once
	shapePathCache *rifx.Chunk
	shapePathErr   error

	shapeStrokeOnce  sync.Once
	shapeStrokeCache *rifx.Chunk
	shapeStrokeErr   error

	shapeStrokeDashedOnce  sync.Once
	shapeStrokeDashedCache *rifx.Chunk
	shapeStrokeDashedErr   error

	shapeGradFillOnce  sync.Once
	shapeGradFillCache *rifx.Chunk
	shapeGradFillErr   error

	shapeGradStrokeOnce  sync.Once
	shapeGradStrokeCache *rifx.Chunk
	shapeGradStrokeErr   error

	shapeTrimOnce  sync.Once
	shapeTrimCache *rifx.Chunk
	shapeTrimErr   error

	shapeTrimTypeLeafOnce  sync.Once
	shapeTrimTypeLeafCache *rifx.Chunk
	shapeTrimTypeLeafErr   error

	shapeRepeaterOnce  sync.Once
	shapeRepeaterCache *rifx.Chunk
	shapeRepeaterErr   error

	shapeRepeaterOrderLeafOnce  sync.Once
	shapeRepeaterOrderLeafCache *rifx.Chunk
	shapeRepeaterOrderLeafErr   error

	shapeRoundCornersOnce  sync.Once
	shapeRoundCornersCache *rifx.Chunk
	shapeRoundCornersErr   error

	shapeOffsetOnce  sync.Once
	shapeOffsetCache *rifx.Chunk
	shapeOffsetErr   error

	shapeOffsetCopiesLeafOnce  sync.Once
	shapeOffsetCopiesLeafCache *rifx.Chunk
	shapeOffsetCopiesLeafErr   error

	shapeOffsetLineJoinLeafOnce  sync.Once
	shapeOffsetLineJoinLeafCache *rifx.Chunk
	shapeOffsetLineJoinLeafErr   error

	shapeOffsetMiterLeafOnce  sync.Once
	shapeOffsetMiterLeafCache *rifx.Chunk
	shapeOffsetMiterLeafErr   error

	shapeOffsetCopyOffsetLeafOnce  sync.Once
	shapeOffsetCopyOffsetLeafCache *rifx.Chunk
	shapeOffsetCopyOffsetLeafErr   error

	shapeMergeOnce  sync.Once
	shapeMergeCache *rifx.Chunk
	shapeMergeErr   error

	shapeZigZagOnce  sync.Once
	shapeZigZagCache *rifx.Chunk
	shapeZigZagErr   error

	shapeZigZagPointsLeafOnce  sync.Once
	shapeZigZagPointsLeafCache *rifx.Chunk
	shapeZigZagPointsLeafErr   error

	shapeStarOnce  sync.Once
	shapeStarCache *rifx.Chunk
	shapeStarErr   error

	shapeStarPolygonOnce  sync.Once
	shapeStarPolygonCache *rifx.Chunk
	shapeStarPolygonErr   error

	shapePuckerBloatOnce  sync.Once
	shapePuckerBloatCache *rifx.Chunk
	shapePuckerBloatErr   error

	shapeTwistOnce  sync.Once
	shapeTwistCache *rifx.Chunk
	shapeTwistErr   error

	shapeTwistCenterLeafOnce  sync.Once
	shapeTwistCenterLeafCache *rifx.Chunk
	shapeTwistCenterLeafErr   error

	shapeWiggleOnce  sync.Once
	shapeWiggleCache *rifx.Chunk
	shapeWiggleErr   error

	shapeWiggleTransformOnce  sync.Once
	shapeWiggleTransformCache *rifx.Chunk
	shapeWiggleTransformErr   error

	shapeRoughenPointsLeafOnce  sync.Once
	shapeRoughenPointsLeafCache *rifx.Chunk
	shapeRoughenPointsLeafErr   error

	shapeCorrelationLeafOnce  sync.Once
	shapeCorrelationLeafCache *rifx.Chunk
	shapeCorrelationLeafErr   error

	shapeTemporalPhaseLeafOnce  sync.Once
	shapeTemporalPhaseLeafCache *rifx.Chunk
	shapeTemporalPhaseLeafErr   error

	shapeSpatialPhaseLeafOnce  sync.Once
	shapeSpatialPhaseLeafCache *rifx.Chunk
	shapeSpatialPhaseLeafErr   error
)

func cloneShapeRectBody() (*rifx.Chunk, error) {
	shapeRectOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeRectBodyBytes))
		if err != nil {
			shapeRectErr = fmt.Errorf("parse shapeRectBodyBytes: %w", err)
			return
		}
		shapeRectCache = ch
	})
	if shapeRectErr != nil {
		return nil, shapeRectErr
	}
	return cloneChunk(shapeRectCache), nil
}

func cloneShapeFillBody() (*rifx.Chunk, error) {
	shapeFillOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeFillBodyBytes))
		if err != nil {
			shapeFillErr = fmt.Errorf("parse shapeFillBodyBytes: %w", err)
			return
		}
		shapeFillCache = ch
	})
	if shapeFillErr != nil {
		return nil, shapeFillErr
	}
	return cloneChunk(shapeFillCache), nil
}

func cloneShapeEllipseBody() (*rifx.Chunk, error) {
	shapeEllipseOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeEllipseBodyBytes))
		if err != nil {
			shapeEllipseErr = fmt.Errorf("parse shapeEllipseBodyBytes: %w", err)
			return
		}
		shapeEllipseCache = ch
	})
	if shapeEllipseErr != nil {
		return nil, shapeEllipseErr
	}
	return cloneChunk(shapeEllipseCache), nil
}

func cloneShapePathBody() (*rifx.Chunk, error) {
	shapePathOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapePathBodyBytes))
		if err != nil {
			shapePathErr = fmt.Errorf("parse shapePathBodyBytes: %w", err)
			return
		}
		shapePathCache = ch
	})
	if shapePathErr != nil {
		return nil, shapePathErr
	}
	return cloneChunk(shapePathCache), nil
}

func cloneShapeStrokeBody() (*rifx.Chunk, error) {
	shapeStrokeOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeStrokeBodyBytes))
		if err != nil {
			shapeStrokeErr = fmt.Errorf("parse shapeStrokeBodyBytes: %w", err)
			return
		}
		shapeStrokeCache = ch
	})
	if shapeStrokeErr != nil {
		return nil, shapeStrokeErr
	}
	return cloneChunk(shapeStrokeCache), nil
}

// cloneShapeStrokeDashedBody returns a clone of the dashed-stroke template — a
// superset of the solid stroke body that additionally carries the Dashes group's
// Dash 1 / Gap 1 slots. lowerStrokeNode selects it when dashes are enabled.
func cloneShapeStrokeDashedBody() (*rifx.Chunk, error) {
	shapeStrokeDashedOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeStrokeDashedBodyBytes))
		if err != nil {
			shapeStrokeDashedErr = fmt.Errorf("parse shapeStrokeDashedBodyBytes: %w", err)
			return
		}
		shapeStrokeDashedCache = ch
	})
	if shapeStrokeDashedErr != nil {
		return nil, shapeStrokeDashedErr
	}
	return cloneChunk(shapeStrokeDashedCache), nil
}

// cloneShapeGradFillBody returns a clone of the gradient-fill template. The
// body carries only `ADBE Vector Grad Colors` (GCst→GCky→Utf8 stops XML); Grad
// Type / Start Pt / End Pt were default in the source fixture and AE elided
// them, so the ramp geometry is not overwritable (default linear on open).
func cloneShapeGradFillBody() (*rifx.Chunk, error) {
	shapeGradFillOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeGradFillBodyBytes))
		if err != nil {
			shapeGradFillErr = fmt.Errorf("parse shapeGradFillBodyBytes: %w", err)
			return
		}
		shapeGradFillCache = ch
	})
	if shapeGradFillErr != nil {
		return nil, shapeGradFillErr
	}
	return cloneChunk(shapeGradFillCache), nil
}

// encodeShapeColorBE returns the AE shape-color cdat bytes for an [r,g,b,a]
// (0..1) color: AE stores colors as [A,R,G,B] × 255 as f64 BE (RE'd from the //nolint:jargon
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
	ShapeKindRect:            "ADBE Vector Shape - Rect",
	ShapeKindEllipse:         "ADBE Vector Shape - Ellipse",
	ShapeKindPath:            "ADBE Vector Shape - Group",
	ShapeKindFill:            "ADBE Vector Graphic - Fill",
	ShapeKindStroke:          "ADBE Vector Graphic - Stroke",
	ShapeKindGroup:           "ADBE Vector Group",
	ShapeKindGradientFill:    "ADBE Vector Graphic - G-Fill",
	ShapeKindGradientStroke:  "ADBE Vector Graphic - G-Stroke",
	ShapeKindTrim:            "ADBE Vector Filter - Trim",
	ShapeKindRepeater:        "ADBE Vector Filter - Repeater",
	ShapeKindRoundCorners:    "ADBE Vector Filter - RC",
	ShapeKindOffsetPaths:     "ADBE Vector Filter - Offset",
	ShapeKindMergePaths:      "ADBE Vector Filter - Merge",
	ShapeKindZigZag:          "ADBE Vector Filter - Zigzag",
	ShapeKindStar:            "ADBE Vector Shape - Star",
	ShapeKindPuckerBloat:     "ADBE Vector Filter - PB",
	ShapeKindTwist:           "ADBE Vector Filter - Twist",
	ShapeKindWigglePaths:     "ADBE Vector Filter - Roughen",
	ShapeKindWiggleTransform: "ADBE Vector Filter - Wiggler",
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
	case *StarNode:
		return lowerStarNode(node, ctx)
	case *PuckerBloatNode:
		return lowerPuckerBloatNode(node, ctx)
	case *TwistNode:
		return lowerTwistNode(node, ctx)
	case *WigglePathsNode:
		return lowerWigglePathsNode(node, ctx)
	case *WiggleTransformNode:
		return lowerWiggleTransformNode(node, ctx)
	default:
		return nil, fmt.Errorf("lowerShapeNode: unsupported kind %v", n.Kind())
	}
}

// lowerRectNode emits a Rect shape body using embedded tolerance bytes
// (templates/shapes/rect_body.bin). From-scratch construction triggered
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
		return injectAnimatedStream(body, name, kfList, ctx.tickRate)
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
	return injectAnimatedStream(body, streamName, kfList, ctx.tickRate)
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
	return injectAnimatedStream(body, streamName, kfList, ctx.tickRate)
}

// injectAnimatedStream flips a static shape stream in an embedded body to
// animated: it finds the tdmn matching streamName, descends into the following
// LIST(tdbs), patches the tdb4 static→animated flags, and replaces the static
// cdat with the supplied LIST(list)(lhd3+ldat) keyframe container. AE keeps
// tdsb/tdsn/tdb4/tdum/tduM otherwise unchanged.
func injectAnimatedStream(body *rifx.Chunk, streamName string, kfList *rifx.Chunk, tickRate float64) error {
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
				// @0x0C..0x0F = keyframe time base (= comp TickRate). The embedded
				// template carries its source fixture's 30 fps value (30720); AE
				// divides each keyframe tick by this, so leaving 30720 on a 29.97
				// comp evaluates keyframes 30720/23976× too fast. Stamp the comp's
				// real TickRate. See ntsc-tickrate-derive-3x-off § tdb4 finding.
				if tickRate > 0 {
					binary.BigEndian.PutUint32(tdb4.Data[0x0C:0x10], uint32(math.Round(tickRate)))
				}
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
// bytes (templates/shapes/ellipse_body.bin). Same rationale as
// lowerRectNode — from-scratch emit triggered AE silent-drop (the body's
// boilerplate, not the values, fails AE's semantic validation); embedding the
// canonical AE-saved body + overwriting Size/Position cdat with runtime values
// is the validator-safe path.
//
// Limitations:
//   - Direction: AE default (the AE-saved body elides the Direction sub-prop;
//     embedded body has no slot to overwrite).
//   - Animated Size: keyframes persisted (non-spatial Vec2). Animated Position:
//     keyframes persisted (spatial Vec2, bpk 104 value@0x38 — RE'd from the //nolint:jargon
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
// (templates/shapes/path_body.bin). From-scratch emit CRASHED AE 2020
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

	// @0x0C / @0x1C encode list capacity in pages of 4 keyframes (same as
	// encodeKeyframes): @0x0C = ceil(n/4) pages, @0x1C = 4 × pages. Hardcoding
	// 1 / 4 made any path with >4 keyframes claim capacity 4 → AE 2025 rejects
	// the file as corrupt (AE 2020 recomputes loosely). See
	// incidents/lhd3-keyframe-capacity-pages.md.
	pages := uint32((n + 3) / 4)
	if pages == 0 {
		pages = 1
	}
	lhd3 := make([]byte, 52)
	lhd3[0], lhd3[1], lhd3[2], lhd3[3] = 0x00, 0xd0, 0x0b, 0xee
	binary.BigEndian.PutUint32(lhd3[0x08:0x0C], uint32(n))
	binary.BigEndian.PutUint32(lhd3[0x0C:0x10], pages)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], bpk)
	binary.BigEndian.PutUint32(lhd3[0x14:0x18], 4)
	binary.BigEndian.PutUint32(lhd3[0x18:0x1C], 1)
	binary.BigEndian.PutUint32(lhd3[0x1C:0x20], 4*pages)

	tickRate := ctx.tickRate
	if tickRate <= 0 {
		tickRate = 30720
	}
	// The path time-table block is the spatial-style 64B ease block (== scalar
	// spatial keyframe): inInterp@0x04 / outInterp@0x05, then in/out
	// speed·influence f64 at 0x18/0x20/0x28/0x30. Honor each keyframe's ease
	// (was hardcoded linear, silently dropping AddKeyframeWithEase on paths). A
	// side with non-zero ease emits Bezier(2); AE ignores the speed/influence
	// table on a side whose interp byte says Linear.
	interpFor := func(e codec.TemporalEase) byte {
		if e.Speed != 0 || e.Influence != 0 {
			return byte(InterpBezier)
		}
		return byte(InterpLinear)
	}
	ldat := make([]byte, n*bpk)
	for i, kf := range kfs {
		blk := ldat[i*bpk : (i+1)*bpk]
		binary.BigEndian.PutUint32(blk[0x00:0x04], uint32(math.Round(kf.Time*tickRate)))
		blk[0x04] = interpFor(kf.InEase)
		blk[0x05] = interpFor(kf.OutEase)
		blk[0x07] = 0x01
		binary.BigEndian.PutUint32(blk[0x08:0x0C], 2)
		if i != n-1 {
			binary.BigEndian.PutUint64(blk[0x10:0x18], math.Float64bits(1.0))
		}
		binary.BigEndian.PutUint64(blk[0x18:0x20], math.Float64bits(kf.InEase.Speed))
		binary.BigEndian.PutUint64(blk[0x20:0x28], math.Float64bits(kf.InEase.Influence))
		binary.BigEndian.PutUint64(blk[0x28:0x30], math.Float64bits(kf.OutEase.Speed))
		binary.BigEndian.PutUint64(blk[0x30:0x38], math.Float64bits(kf.OutEase.Influence))
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
// (templates/shapes/fill_body.bin). Same rationale as lowerRectNode —
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
