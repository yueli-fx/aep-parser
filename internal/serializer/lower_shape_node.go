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
func lowerGradientStops(body *rifx.Chunk, gradient *codec.Gradient) *rifx.Chunk {
	if gradient != nil {
		overwriteGradientStopsXML(body, "ADBE Vector Grad Colors", codec.EncodeGradientXML(gradient))
	}
	return body
}

// lowerGradientFillNode emits a gradient-fill graphic body from the embedded
// template (templates/shapes/gradfill_body.bin), overwriting the Grad
// Start/End Pt cdats (the linear ramp direction) + the Grad Colors stops XML
// with the runtime gradient. The XML length changes per stop count →
// length-variable; the Utf8 chunk's Data is swapped and rifx.Chunk.Write
// recomputes the enclosing GCky / GCst / tdgp LIST sizes on serialization.
//
// Start/End Pt are Vec2 (2 × f64 BE at cdat[0:16], same layout as the Repeater
// Transform points), overwritten with the node's StartPoint/EndPoint (default
// [0,0]→[100,0] = AE's horizontal ramp, so a gradient that doesn't set direction
// reproduces the pre-direction behavior). Grad Type (1D f64 BE enum: 1=Linear /
// 2=Radial) is overwritten with the node's GradientType — the template bakes
// Radial(2) so the slot exists, lowered back to the node's value (default
// Linear=1 reproduces the pre-type behavior). HiLite Length / Angle (both 1D f64
// BE at cdat[0:8]) offset a radial gradient's bright centre; default 0/0
// overwrites the baked slots with no visible change (no regression).
func lowerGradientFillNode(n *GradientFillNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeGradFillBody()
	if err != nil {
		return nil, err
	}
	sp, ep := n.StartPoint(), n.EndPoint()
	overwriteShapeStreamCdat(body, "ADBE Vector Grad Type", encodeF64sBE(float64(n.GradientType())))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad Start Pt", encodeF64sBE(sp[0], sp[1]))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad End Pt", encodeF64sBE(ep[0], ep[1]))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad HiLite Length", encodeF64sBE(n.HighlightLength()))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad HiLite Angle", encodeF64sBE(n.HighlightAngle()))
	if kfs := n.GradientKeyframes(); len(kfs) > 0 {
		if err := animateGradientStops(body, "ADBE Vector Grad Colors", kfs, ctx); err != nil {
			return nil, err
		}
		return body, nil
	}
	return lowerGradientStops(body, n.Gradient()), nil
}

// cloneShapeGradStrokeBody returns a clone of the gradient-stroke template
// (templates/shapes/gradstroke_body.bin). Like gradfill it carries the five
// gradient-geometry slots (Grad Type / Start Pt / End Pt / HiLite Length·Angle)
// + `ADBE Vector Grad Colors`; the stroke geometry (width/cap/join/…) stays at
// the extracted values.
func cloneShapeGradStrokeBody() (*rifx.Chunk, error) {
	shapeGradStrokeOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeGradStrokeBodyBytes))
		if err != nil {
			shapeGradStrokeErr = fmt.Errorf("parse shapeGradStrokeBodyBytes: %w", err)
			return
		}
		shapeGradStrokeCache = ch
	})
	if shapeGradStrokeErr != nil {
		return nil, shapeGradStrokeErr
	}
	return cloneChunk(shapeGradStrokeCache), nil
}

// lowerGradientStrokeNode emits a gradient-stroke body from the embedded
// template, overwriting the Grad Type / Start Pt / End Pt / HiLite Length·Angle
// cdats + the Grad Colors stops XML. Geometry cdat layout is identical to
// lowerGradientFillNode (Type/HiLite = 1D f64 BE @cdat[0:8], Start/End = Vec2 @
// cdat[0:16]); only the cloned template differs. The template bakes Radial(2) /
// [-120,-120]→[120,120] so the slots exist; defaults (Linear=1, [0,0]→[100,0],
// HiLite 0/0) reproduce AE's pre-geometry stroke behavior (no regression).
func lowerGradientStrokeNode(n *GradientStrokeNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeGradStrokeBody()
	if err != nil {
		return nil, err
	}
	sp, ep := n.StartPoint(), n.EndPoint()
	overwriteShapeStreamCdat(body, "ADBE Vector Grad Type", encodeF64sBE(float64(n.GradientType())))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad Start Pt", encodeF64sBE(sp[0], sp[1]))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad End Pt", encodeF64sBE(ep[0], ep[1]))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad HiLite Length", encodeF64sBE(n.HighlightLength()))
	overwriteShapeStreamCdat(body, "ADBE Vector Grad HiLite Angle", encodeF64sBE(n.HighlightAngle()))
	// Stroke geometry (1D f64 BE @cdat[0:8]; enums 1-based). Defaults mirror the
	// template's baked values (18 / round / round / 4), so a node that doesn't
	// override them re-emits byte-identically (no regression on the existing
	// gradstroke gates).
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Width", encodeF64sBE(n.StrokeWidth()))
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Line Cap", encodeF64sBE(float64(n.LineCap())))
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Line Join", encodeF64sBE(float64(n.LineJoin())))
	overwriteShapeStreamCdat(body, "ADBE Vector Stroke Miter Limit", encodeF64sBE(n.MiterLimit()))
	// Animated color stops — same `ADBE Vector Grad Colors` stream + helper as the
	// gradient fill (the stream is identical on fill and stroke).
	if kfs := n.GradientKeyframes(); len(kfs) > 0 {
		if err := animateGradientStops(body, "ADBE Vector Grad Colors", kfs, ctx); err != nil {
			return nil, err
		}
		return body, nil
	}
	return lowerGradientStops(body, n.Gradient()), nil
}

// cloneShapeTrimBody returns a clone of the Trim Paths template
// (templates/shapes/trim_body.bin). The body carries the Start / End /
// Offset cdat slots (Trim Type was AE-default and elided — no slot).
func cloneShapeTrimBody() (*rifx.Chunk, error) {
	shapeTrimOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeTrimBodyBytes))
		if err != nil {
			shapeTrimErr = fmt.Errorf("parse shapeTrimBodyBytes: %w", err)
			return
		}
		shapeTrimCache = ch
	})
	if shapeTrimErr != nil {
		return nil, shapeTrimErr
	}
	return cloneChunk(shapeTrimCache), nil
}

// lowerTrimNode emits a Trim Paths filter body from the embedded template,
// overwriting the Start / End / Offset cdat slots with runtime values. Start /
// End are raw percentages (0..100), Offset is raw degrees — all float64 BE at
// cdat[0:8], identical to the other shape scalars (RE'd from v2_2_trim.aep). //nolint:jargon
//
// Static → cdat overwrite; animated → the cdat flips to a 1D non-spatial
// keyframe container (same injectAnimatedStream path as Rect Roundness / Fill
// Opacity), so the line-draw reveal (keyframed End 0→100) persists.
//
// Trim Type (Simultaneously/Individually) is AE-default-elided; when set to
// Individually the leaf is spliced into the body in canonical order (after
// Offset, before Group End) and its enum cdat overwritten — synthesis-insert,
// mirroring Offset Copies / SetMaterialOption.
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
	if n.Type() != TrimTypeSimultaneously {
		tdmn, tdbs, err := cloneShapeTrimTypeLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Trim Type", encodeF64sBE(float64(n.Type())))
	}
	return body, nil
}

// cloneShapeTrimTypeLeaf returns a fresh (tdmn, LIST:tdbs) clone of the
// `ADBE Vector Trim Type` enum leaf from its embedded template, spliced into the
// trim body when Trim Type is set non-default (Individually). Mirrors
// cloneShapeOffsetCopiesLeaf.
func cloneShapeTrimTypeLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeTrimTypeLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeTrimTypeLeafBytes))
		if e != nil {
			shapeTrimTypeLeafErr = fmt.Errorf("parse shapeTrimTypeLeafBytes: %w", e)
			return
		}
		shapeTrimTypeLeafCache = ch
	})
	if shapeTrimTypeLeafErr != nil {
		return nil, nil, shapeTrimTypeLeafErr
	}
	kids := shapeTrimTypeLeafCache.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Vector Trim Type" &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return cloneChunk(kids[i]), cloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("trim type leaf missing from template")
}

// cloneShapeRepeaterBody returns a clone of the Repeater template
// (templates/shapes/repeater_body.bin): top-level Copies/Offset cdat slots
// + a nested `ADBE Vector Repeater Transform` group (Anchor/Position/Scale/
// Rotation/Opacity 1·2). Order (Composite) was AE-default and elided.
func cloneShapeRepeaterBody() (*rifx.Chunk, error) {
	shapeRepeaterOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeRepeaterBodyBytes))
		if err != nil {
			shapeRepeaterErr = fmt.Errorf("parse shapeRepeaterBodyBytes: %w", err)
			return
		}
		shapeRepeaterCache = ch
	})
	if shapeRepeaterErr != nil {
		return nil, shapeRepeaterErr
	}
	return cloneChunk(shapeRepeaterCache), nil
}

// lowerRepeaterNode emits a Repeater filter body from the embedded template.
// Top-level Copies/Offset are 1D scalars (static cdat overwrite / animated flip
// via lowerShapeScalar). The per-copy Transform sub-streams live in the nested
// `ADBE Vector Repeater Transform` group (descend via findGroupBody) and are
// static cdat overwrites (Anchor/Position/Scale are Vec2 at cdat[0:16],
// Rotation/Opacity 1·2 are 1D at cdat[0:8]) — same mechanism as Stroke
// Taper/Wave. RE'd from v2_2_repeater.aep. //nolint:jargon
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
	// Order (Composite) is canonically between Offset and the Transform group, so
	// splice it BEFORE the Transform group (not before Group End) — synthesis-insert.
	if n.OrderSet() && n.Order() != RepeaterOrderBelow {
		tdmn, tdbs, err := cloneShapeRepeaterOrderLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Repeater Transform", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Repeater Order", encodeF64sBE(float64(n.Order())))
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

// spliceShapeLeafBefore inserts a (tdmn, tdbs) leaf pair into a shape filter body
// immediately before the child tdmn matching beforeMatchName (e.g. before a
// nested group, for a leaf whose canonical position precedes it). Falls back to
// spliceShapeLeafBeforeGroupEnd if the target child is absent.
func spliceShapeLeafBefore(body *rifx.Chunk, beforeMatchName string, tdmn, tdbs *rifx.Chunk) {
	kids := body.Children
	insertIdx := -1
	for i := 0; i < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == beforeMatchName {
			insertIdx = i
			break
		}
	}
	if insertIdx < 0 {
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		return
	}
	spliced := make([]*rifx.Chunk, 0, len(kids)+2)
	spliced = append(spliced, kids[:insertIdx]...)
	spliced = append(spliced, tdmn, tdbs)
	spliced = append(spliced, kids[insertIdx:]...)
	body.Children = spliced
}

// cloneShapeRepeaterOrderLeaf returns a fresh (tdmn, LIST:tdbs) clone of the
// `ADBE Vector Repeater Order` enum leaf from its embedded template, spliced
// before the Repeater Transform group when Order is set Above. Mirrors
// cloneShapeTrimTypeLeaf.
func cloneShapeRepeaterOrderLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeRepeaterOrderLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeRepeaterOrderLeafBytes))
		if e != nil {
			shapeRepeaterOrderLeafErr = fmt.Errorf("parse shapeRepeaterOrderLeafBytes: %w", e)
			return
		}
		shapeRepeaterOrderLeafCache = ch
	})
	if shapeRepeaterOrderLeafErr != nil {
		return nil, nil, shapeRepeaterOrderLeafErr
	}
	kids := shapeRepeaterOrderLeafCache.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Vector Repeater Order" &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return cloneChunk(kids[i]), cloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("repeater order leaf missing from template")
}

// cloneShapeRoundCornersBody returns a clone of the Round Corners template
// (templates/shapes/roundcorners_body.bin): a single `ADBE Vector
// RoundCorner Radius` cdat slot (Radius was set non-default in the fixture so
// AE emitted it).
func cloneShapeRoundCornersBody() (*rifx.Chunk, error) {
	shapeRoundCornersOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeRoundCornersBodyBytes))
		if err != nil {
			shapeRoundCornersErr = fmt.Errorf("parse shapeRoundCornersBodyBytes: %w", err)
			return
		}
		shapeRoundCornersCache = ch
	})
	if shapeRoundCornersErr != nil {
		return nil, shapeRoundCornersErr
	}
	return cloneChunk(shapeRoundCornersCache), nil
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
// (templates/shapes/offset_body.bin): a single `ADBE Vector Offset Amount`
// cdat slot (Amount set non-default in the fixture so AE emitted it; Line Join /
// Miter Limit / Copies / Copy Offset stayed default and are elided).
func cloneShapeOffsetBody() (*rifx.Chunk, error) {
	shapeOffsetOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeOffsetBodyBytes))
		if err != nil {
			shapeOffsetErr = fmt.Errorf("parse shapeOffsetBodyBytes: %w", err)
			return
		}
		shapeOffsetCache = ch
	})
	if shapeOffsetErr != nil {
		return nil, shapeOffsetErr
	}
	return cloneChunk(shapeOffsetCache), nil
}

// cloneShapeOffsetCopiesLeaf returns a fresh (tdmn, LIST:tdbs) clone of the
// `ADBE Vector Offset Copies` leaf from its embedded template (a LIST(tdgp)
// wrapper holding the single pair, extracted from an AE-saved offset whose Copies
// was authored non-default). spliced into the Amount-only offset body when Copies
// is set — mirrors SetMaterialOption's per-leaf splice for default-elided props.
func cloneShapeOffsetCopiesLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeOffsetCopiesLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeOffsetCopiesLeafBytes))
		if e != nil {
			shapeOffsetCopiesLeafErr = fmt.Errorf("parse shapeOffsetCopiesLeafBytes: %w", e)
			return
		}
		shapeOffsetCopiesLeafCache = ch
	})
	if shapeOffsetCopiesLeafErr != nil {
		return nil, nil, shapeOffsetCopiesLeafErr
	}
	kids := shapeOffsetCopiesLeafCache.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Vector Offset Copies" &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return cloneChunk(kids[i]), cloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("offset copies leaf missing from template")
}

// spliceShapeLeafBeforeGroupEnd inserts a (tdmn, LIST:tdbs) leaf pair into a
// shape filter body's LIST(tdgp) immediately before the trailing
// `ADBE Group End` sentinel — the canonical tail position for a leaf that sorts
// after the body's existing slots (Offset Copies follows Amount). No-op if no
// Group End is found (returns the pair appended at the end).
func spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs *rifx.Chunk) {
	kids := body.Children
	insertIdx := len(kids)
	for i := 0; i < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Group End" {
			insertIdx = i
			break
		}
	}
	spliced := make([]*rifx.Chunk, 0, len(kids)+2)
	spliced = append(spliced, kids[:insertIdx]...)
	spliced = append(spliced, tdmn, tdbs)
	spliced = append(spliced, kids[insertIdx:]...)
	body.Children = spliced
}

// lowerOffsetPathsNode emits an Offset Paths filter body from the embedded
// template, overwriting the single `ADBE Vector Offset Amount` cdat (1D f64 BE
// at cdat[0:8], raw pixels — same scalar layout as Round Corners Radius).
// Static → cdat overwrite; animated → the cdat flips to a 1D non-spatial
// keyframe container via the shared injectAnimatedStream path.
//
// The four AE-default-elided sub-streams (Line Join / Miter Limit / Copies /
// Copy Offset) are each spliced in on demand when their setter is used. They are
// spliced in canonical order (Line Join → Miter Limit → Copies → Copy Offset),
// each inserted just before Group End, so call order == stored order —
// synthesis-insert, mirroring SetMaterialOption.
func lowerOffsetPathsNode(n *OffsetPathsNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeOffsetBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Offset Amount", n.Amount(), ctx); err != nil {
		return nil, err
	}
	if n.LineJoinSet() && n.LineJoin() != StrokeLineJoinMiter {
		tdmn, tdbs, err := cloneShapeOffsetLineJoinLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Offset Line Join", encodeF64sBE(float64(n.LineJoin())))
	}
	if n.MiterLimitSet() && n.MiterLimit() != 4 {
		tdmn, tdbs, err := cloneShapeOffsetMiterLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Offset Miter Limit", encodeF64sBE(n.MiterLimit()))
	}
	if n.CopiesSet() && n.Copies() != 1 {
		tdmn, tdbs, err := cloneShapeOffsetCopiesLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Offset Copies", encodeF64sBE(n.Copies()))
	}
	if n.CopyOffsetSet() && n.CopyOffset() != 1 {
		tdmn, tdbs, err := cloneShapeOffsetCopyOffsetLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Offset Copy Offset", encodeF64sBE(n.CopyOffset()))
	}
	return body, nil
}

// cloneShapeOffsetLineJoinLeaf / cloneShapeOffsetMiterLeaf /
// cloneShapeOffsetCopyOffsetLeaf return fresh clones of the three remaining
// AE-default-elided Offset sub-stream leaves (enum / scalar / scalar) from their
// embedded templates, spliced into the Amount-only offset body on demand. All
// mirror cloneShapeOffsetCopiesLeaf.
func cloneShapeOffsetLineJoinLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeOffsetLineJoinLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeOffsetLineJoinLeafBytes))
		if e != nil {
			shapeOffsetLineJoinLeafErr = fmt.Errorf("parse shapeOffsetLineJoinLeafBytes: %w", e)
			return
		}
		shapeOffsetLineJoinLeafCache = ch
	})
	if shapeOffsetLineJoinLeafErr != nil {
		return nil, nil, shapeOffsetLineJoinLeafErr
	}
	return offsetLeafPair(shapeOffsetLineJoinLeafCache, "ADBE Vector Offset Line Join")
}

func cloneShapeOffsetMiterLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeOffsetMiterLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeOffsetMiterLeafBytes))
		if e != nil {
			shapeOffsetMiterLeafErr = fmt.Errorf("parse shapeOffsetMiterLeafBytes: %w", e)
			return
		}
		shapeOffsetMiterLeafCache = ch
	})
	if shapeOffsetMiterLeafErr != nil {
		return nil, nil, shapeOffsetMiterLeafErr
	}
	return offsetLeafPair(shapeOffsetMiterLeafCache, "ADBE Vector Offset Miter Limit")
}

func cloneShapeOffsetCopyOffsetLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeOffsetCopyOffsetLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeOffsetCopyOffsetLeafBytes))
		if e != nil {
			shapeOffsetCopyOffsetLeafErr = fmt.Errorf("parse shapeOffsetCopyOffsetLeafBytes: %w", e)
			return
		}
		shapeOffsetCopyOffsetLeafCache = ch
	})
	if shapeOffsetCopyOffsetLeafErr != nil {
		return nil, nil, shapeOffsetCopyOffsetLeafErr
	}
	return offsetLeafPair(shapeOffsetCopyOffsetLeafCache, "ADBE Vector Offset Copy Offset")
}

// offsetLeafPair finds the (tdmn, LIST:tdbs) pair named `name` in a parsed leaf
// template wrapper and returns fresh clones.
func offsetLeafPair(cache *rifx.Chunk, name string) (tdmn, tdbs *rifx.Chunk, err error) {
	kids := cache.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == name &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return cloneChunk(kids[i]), cloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("offset leaf %q missing from template", name)
}

// cloneShapeMergeBody returns a clone of the Merge Paths template
// (templates/shapes/merge_body.bin): a single `ADBE Vector Merge Type` enum
// cdat slot (Type set non-default in the fixture so AE emitted it).
func cloneShapeMergeBody() (*rifx.Chunk, error) {
	shapeMergeOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeMergeBodyBytes))
		if err != nil {
			shapeMergeErr = fmt.Errorf("parse shapeMergeBodyBytes: %w", err)
			return
		}
		shapeMergeCache = ch
	})
	if shapeMergeErr != nil {
		return nil, shapeMergeErr
	}
	return cloneChunk(shapeMergeCache), nil
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
// (templates/shapes/zigzag_body.bin): `ADBE Vector Zigzag Size` +
// `ADBE Vector Zigzag Detail` cdat slots (both set non-default in the fixture so
// AE emitted them; the Points enum stayed default and is elided).
func cloneShapeZigZagBody() (*rifx.Chunk, error) {
	shapeZigZagOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeZigZagBodyBytes))
		if err != nil {
			shapeZigZagErr = fmt.Errorf("parse shapeZigZagBodyBytes: %w", err)
			return
		}
		shapeZigZagCache = ch
	})
	if shapeZigZagErr != nil {
		return nil, shapeZigZagErr
	}
	return cloneChunk(shapeZigZagCache), nil
}

// lowerZigZagNode emits a ZigZag filter body from the embedded template,
// overwriting the `ADBE Vector Zigzag Size` (amplitude) and `ADBE Vector Zigzag
// Detail` (ridges per segment) cdats — both 1D f64 BE at cdat[0:8], same scalar
// layout as the other shape scalars. Static → cdat overwrite; animated → the
// cdat flips to a 1D non-spatial keyframe container via injectAnimatedStream.
//
// When Points is set non-default (Smooth), the AE-default-elided `ADBE Vector
// Zigzag Points` enum leaf is spliced into the body in canonical order (after
// Detail, before Group End) and its cdat overwritten — synthesis-insert,
// mirroring Trim Type / Offset Copies.
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
	if n.Points() != ZigZagPointsCorner {
		tdmn, tdbs, err := cloneShapeZigZagPointsLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Zigzag Points", encodeF64sBE(float64(n.Points())))
	}
	return body, nil
}

// cloneShapeZigZagPointsLeaf returns a fresh (tdmn, LIST:tdbs) clone of the
// `ADBE Vector Zigzag Points` enum leaf from its embedded template, spliced into
// the zigzag body when Points is set non-default (Smooth). Mirrors
// cloneShapeTrimTypeLeaf.
func cloneShapeZigZagPointsLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeZigZagPointsLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeZigZagPointsLeafBytes))
		if e != nil {
			shapeZigZagPointsLeafErr = fmt.Errorf("parse shapeZigZagPointsLeafBytes: %w", e)
			return
		}
		shapeZigZagPointsLeafCache = ch
	})
	if shapeZigZagPointsLeafErr != nil {
		return nil, nil, shapeZigZagPointsLeafErr
	}
	kids := shapeZigZagPointsLeafCache.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Vector Zigzag Points" &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return cloneChunk(kids[i]), cloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("zigzag points leaf missing from template")
}

// cloneShapeStarBody returns a clone of the Star template
// (templates/shapes/star_body.bin): Points / Position / Rotation / Inner
// Radius / Outer Radius / Inner·Outer Roundess cdat slots (all set non-default in
// the fixture so AE emitted them). Star Type / Shape Direction stayed default
// and are elided (Star type only).
func cloneShapeStarBody() (*rifx.Chunk, error) {
	shapeStarOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeStarBodyBytes))
		if err != nil {
			shapeStarErr = fmt.Errorf("parse shapeStarBodyBytes: %w", err)
			return
		}
		shapeStarCache = ch
	})
	if shapeStarErr != nil {
		return nil, shapeStarErr
	}
	return cloneChunk(shapeStarCache), nil
}

// cloneShapeStarPolygonBody returns a clone of the Polygon-type polystar template
// (templates/shapes/starpolygon_body.bin): same sub-streams as the Star body
// PLUS the `ADBE Vector Star Type` slot baked to 2 (Polygon). AE saves Inner
// Radius/Roundness too (hidden, no visual effect) so the slot set is a superset
// of the Star body — the shared lower overwrites apply unchanged.
func cloneShapeStarPolygonBody() (*rifx.Chunk, error) {
	shapeStarPolygonOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeStarPolygonBodyBytes))
		if err != nil {
			shapeStarPolygonErr = fmt.Errorf("parse shapeStarPolygonBodyBytes: %w", err)
			return
		}
		shapeStarPolygonCache = ch
	})
	if shapeStarPolygonErr != nil {
		return nil, shapeStarPolygonErr
	}
	return cloneChunk(shapeStarPolygonCache), nil
}

// lowerStarNode emits a Star shape body from the embedded template, overwriting
// the Points / Rotation / Inner·Outer Radius / Inner·Outer Roundess 1D scalars
// (lowerShapeScalar, static cdat / animated flip) and Position (Vec2, spatial
// motion-path layout, same as Rect Position). Star Type stays at the embed
// default (Star). RE'd from v2_2_star.aep. //nolint:jargon
func lowerStarNode(n *StarNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	// Polygon uses a separate template (carries the Star Type=2 slot, baked); Star
	// uses the original (Type elided at default). The remaining sub-stream
	// overwrites are identical (Inner Radius/Roundness are present in both but
	// have no visual effect on a Polygon).
	clone := cloneShapeStarBody
	if n.IsPolygon() {
		clone = cloneShapeStarPolygonBody
	}
	body, err := clone()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Star Points", n.Points(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeVec2(body, "ADBE Vector Star Position", n.Position(), ctx,
		valueLayout{dim: 2, headerByte: 0x07, spatial: true, motionPath: true}); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Star Rotation", n.Rotation(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Star Inner Radius", n.InnerRadius(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Star Outer Radius", n.OuterRadius(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Star Inner Roundess", n.InnerRoundness(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Star Outer Roundess", n.OuterRoundness(), ctx); err != nil {
		return nil, err
	}
	return body, nil
}

// cloneShapePuckerBloatBody returns a clone of the Pucker & Bloat template
// (templates/shapes/puckerbloat_body.bin): a single `ADBE Vector
// PuckerBloat Amount` cdat slot (Amount was set non-default in the fixture so
// AE emitted it).
func cloneShapePuckerBloatBody() (*rifx.Chunk, error) {
	shapePuckerBloatOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapePuckerBloatBodyBytes))
		if err != nil {
			shapePuckerBloatErr = fmt.Errorf("parse shapePuckerBloatBodyBytes: %w", err)
			return
		}
		shapePuckerBloatCache = ch
	})
	if shapePuckerBloatErr != nil {
		return nil, shapePuckerBloatErr
	}
	return cloneChunk(shapePuckerBloatCache), nil
}

// lowerPuckerBloatNode emits a Pucker & Bloat filter body from the embedded
// template, overwriting the single `ADBE Vector PuckerBloat Amount` cdat (1D
// f64 BE at cdat[0:8], raw percent — same scalar layout as Round Corners
// Radius). Static → cdat overwrite; animated → the cdat flips to a 1D
// non-spatial keyframe container via the shared injectAnimatedStream path.
func lowerPuckerBloatNode(n *PuckerBloatNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapePuckerBloatBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector PuckerBloat Amount", n.Amount(), ctx); err != nil {
		return nil, err
	}
	return body, nil
}

// cloneShapeTwistBody returns a clone of the Twist template
// (templates/shapes/twist_body.bin): a single `ADBE Vector Twist Angle`
// cdat slot (Angle was set non-default in the fixture so AE emitted it; the
// Vec2 Twist Center stayed default and is elided).
func cloneShapeTwistBody() (*rifx.Chunk, error) {
	shapeTwistOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeTwistBodyBytes))
		if err != nil {
			shapeTwistErr = fmt.Errorf("parse shapeTwistBodyBytes: %w", err)
			return
		}
		shapeTwistCache = ch
	})
	if shapeTwistErr != nil {
		return nil, shapeTwistErr
	}
	return cloneChunk(shapeTwistCache), nil
}

// lowerTwistNode emits a Twist filter body from the embedded template,
// overwriting the headline `ADBE Vector Twist Angle` cdat (1D f64 BE at
// cdat[0:8], degrees — same scalar layout as Round Corners Radius). Static →
// cdat overwrite; animated → the cdat flips to a 1D non-spatial keyframe
// container via the shared injectAnimatedStream path.
//
// When Center is set non-zero, the AE-default-elided `ADBE Vector Twist Center`
// Vec2 leaf is spliced into the body in canonical order (after Angle, before
// Group End) and its cdat overwritten as two f64 BE at cdat[0:16] —
// synthesis-insert, the first Vec2 leaf (after the scalar Offset Copies and the
// Trim Type / ZigZag Points enums).
func lowerTwistNode(n *TwistNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeTwistBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Twist Angle", n.Angle(), ctx); err != nil {
		return nil, err
	}
	if c := n.Center(); n.CenterSet() && (c[0] != 0 || c[1] != 0) {
		tdmn, tdbs, err := cloneShapeTwistCenterLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBeforeGroupEnd(body, tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Twist Center", encodeF64sBE(c[0], c[1]))
	}
	return body, nil
}

// cloneShapeTwistCenterLeaf returns a fresh (tdmn, LIST:tdbs) clone of the
// `ADBE Vector Twist Center` Vec2 leaf from its embedded template, spliced into
// the twist body when Center is offset from [0,0]. Mirrors
// cloneShapeZigZagPointsLeaf (the cdat holds two f64 BE rather than one).
func cloneShapeTwistCenterLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeTwistCenterLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeTwistCenterLeafBytes))
		if e != nil {
			shapeTwistCenterLeafErr = fmt.Errorf("parse shapeTwistCenterLeafBytes: %w", e)
			return
		}
		shapeTwistCenterLeafCache = ch
	})
	if shapeTwistCenterLeafErr != nil {
		return nil, nil, shapeTwistCenterLeafErr
	}
	kids := shapeTwistCenterLeafCache.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == "ADBE Vector Twist Center" &&
			kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdbs {
			return cloneChunk(kids[i]), cloneChunk(kids[i+1]), nil
		}
	}
	return nil, nil, fmt.Errorf("twist center leaf missing from template")
}

// cloneShapeWiggleBody returns a clone of the Wiggle Paths template
// (templates/shapes/wiggle_body.bin): four cdat slots — `ADBE Vector Roughen
// Size` / `Roughen Detail` / `Temporal Freq` / `Random Seed` — all set
// non-default in the fixture so AE emitted them. Points / Correlation /
// Temporal·Spatial Phase stayed default and are elided.
func cloneShapeWiggleBody() (*rifx.Chunk, error) {
	shapeWiggleOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeWiggleBodyBytes))
		if err != nil {
			shapeWiggleErr = fmt.Errorf("parse shapeWiggleBodyBytes: %w", err)
			return
		}
		shapeWiggleCache = ch
	})
	if shapeWiggleErr != nil {
		return nil, shapeWiggleErr
	}
	return cloneChunk(shapeWiggleCache), nil
}

// lowerWigglePathsNode emits a Wiggle Paths filter body from the embedded
// template, overwriting the four modeled scalar cdats (Size / Detail /
// WigglesPerSecond=Temporal Freq / RandomSeed; each 1D f64 BE at cdat[0:8]).
// Static → cdat overwrite; animated → each cdat flips to a 1D non-spatial
// keyframe container via the shared injectAnimatedStream path.
func lowerWigglePathsNode(n *WigglePathsNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeWiggleBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Roughen Size", n.Size(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Roughen Detail", n.Detail(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Temporal Freq", n.WigglesPerSecond(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Random Seed", n.RandomSeed(), ctx); err != nil {
		return nil, err
	}
	// Synthesis-insert the AE-default-elided modulation leaves, each in canonical
	// order. Points sits before Temporal Freq; Correlation / Temporal Phase /
	// Spatial Phase sit before Random Seed (canonical Roughen order is Size /
	// Detail / Points / Temporal Freq / Correlation / Temporal Phase / Spatial
	// Phase / Random Seed).
	if n.Points() != RoughenPointsCorner {
		tdmn, tdbs, err := cloneShapeRoughenPointsLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Temporal Freq", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Roughen Points", encodeF64sBE(float64(n.Points())))
	}
	if n.CorrelationSet() && n.Correlation() != 50 {
		tdmn, tdbs, err := cloneShapeCorrelationLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Random Seed", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Correlation", encodeF64sBE(n.Correlation()))
	}
	if n.TemporalPhaseSet() && n.TemporalPhase() != 0 {
		tdmn, tdbs, err := cloneShapeTemporalPhaseLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Random Seed", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Temporal Phase", encodeF64sBE(n.TemporalPhase()))
	}
	if n.SpatialPhaseSet() && n.SpatialPhase() != 0 {
		tdmn, tdbs, err := cloneShapeSpatialPhaseLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Random Seed", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Spatial Phase", encodeF64sBE(n.SpatialPhase()))
	}
	return body, nil
}

// cloneShapeRoughenPointsLeaf / cloneShapeCorrelationLeaf /
// cloneShapeTemporalPhaseLeaf / cloneShapeSpatialPhaseLeaf return fresh
// (tdmn, LIST:tdbs) clones of the four AE-default-elided Wiggle modulation
// sub-stream leaves from their embedded templates. The Correlation / Temporal
// Phase / Spatial Phase leaves are shared by both Wiggle Paths (Roughen) and
// Wiggle Transform (Wiggler) — the on-disk match-names and tdbs layouts are
// identical (the embedded value is overwritten on splice).
func cloneShapeRoughenPointsLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeRoughenPointsLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeRoughenPointsLeafBytes))
		if e != nil {
			shapeRoughenPointsLeafErr = fmt.Errorf("parse shapeRoughenPointsLeafBytes: %w", e)
			return
		}
		shapeRoughenPointsLeafCache = ch
	})
	if shapeRoughenPointsLeafErr != nil {
		return nil, nil, shapeRoughenPointsLeafErr
	}
	return offsetLeafPair(shapeRoughenPointsLeafCache, "ADBE Vector Roughen Points")
}

func cloneShapeCorrelationLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeCorrelationLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeCorrelationLeafBytes))
		if e != nil {
			shapeCorrelationLeafErr = fmt.Errorf("parse shapeCorrelationLeafBytes: %w", e)
			return
		}
		shapeCorrelationLeafCache = ch
	})
	if shapeCorrelationLeafErr != nil {
		return nil, nil, shapeCorrelationLeafErr
	}
	return offsetLeafPair(shapeCorrelationLeafCache, "ADBE Vector Correlation")
}

func cloneShapeTemporalPhaseLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeTemporalPhaseLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeTemporalPhaseLeafBytes))
		if e != nil {
			shapeTemporalPhaseLeafErr = fmt.Errorf("parse shapeTemporalPhaseLeafBytes: %w", e)
			return
		}
		shapeTemporalPhaseLeafCache = ch
	})
	if shapeTemporalPhaseLeafErr != nil {
		return nil, nil, shapeTemporalPhaseLeafErr
	}
	return offsetLeafPair(shapeTemporalPhaseLeafCache, "ADBE Vector Temporal Phase")
}

func cloneShapeSpatialPhaseLeaf() (tdmn, tdbs *rifx.Chunk, err error) {
	shapeSpatialPhaseLeafOnce.Do(func() {
		ch, e := rifx.ReadChunk(bytes.NewReader(shapeSpatialPhaseLeafBytes))
		if e != nil {
			shapeSpatialPhaseLeafErr = fmt.Errorf("parse shapeSpatialPhaseLeafBytes: %w", e)
			return
		}
		shapeSpatialPhaseLeafCache = ch
	})
	if shapeSpatialPhaseLeafErr != nil {
		return nil, nil, shapeSpatialPhaseLeafErr
	}
	return offsetLeafPair(shapeSpatialPhaseLeafCache, "ADBE Vector Spatial Phase")
}

// cloneShapeWiggleTransformBody returns a clone of the Wiggle Transform template
// (templates/shapes/wiggletransform_body.bin): top-level `ADBE Vector Xform
// Temporal Freq` / `Random Seed` cdats plus the nested `ADBE Vector Wiggler
// Transform` group's Anchor/Position/Scale/Rotation cdats — all set non-default
// in the fixture so AE emitted them. Correlation / Temporal·Spatial Phase stayed
// default and are elided.
func cloneShapeWiggleTransformBody() (*rifx.Chunk, error) {
	shapeWiggleTransformOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(shapeWiggleTransformBodyBytes))
		if err != nil {
			shapeWiggleTransformErr = fmt.Errorf("parse shapeWiggleTransformBodyBytes: %w", err)
			return
		}
		shapeWiggleTransformCache = ch
	})
	if shapeWiggleTransformErr != nil {
		return nil, shapeWiggleTransformErr
	}
	return cloneChunk(shapeWiggleTransformCache), nil
}

// lowerWiggleTransformNode emits a Wiggle Transform filter body from the embedded
// template. Top-level WigglesPerSecond (Temporal Freq) / RandomSeed are 1D
// scalars (static cdat overwrite / animated flip via lowerShapeScalar). The
// per-channel wiggle amplitudes live in the nested `ADBE Vector Wiggler
// Transform` group (descend via findGroupBody) as static cdat overwrites —
// Anchor/Position/Scale are Vec2 at cdat[0:16], Rotation is 1D at cdat[0:8] —
// same mechanism as the Repeater Transform. RE'd from v2_2_wiggletransform.aep. //nolint:jargon
func lowerWiggleTransformNode(n *WiggleTransformNode, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeWiggleTransformBody()
	if err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Xform Temporal Freq", n.WigglesPerSecond(), ctx); err != nil {
		return nil, err
	}
	if err := lowerShapeScalar(body, "ADBE Vector Random Seed", n.RandomSeed(), ctx); err != nil {
		return nil, err
	}
	if xf := findGroupBody(body, "ADBE Vector Wiggler Transform"); xf != nil {
		t := n.Transform()
		a, p, s := t.Anchor(), t.Position(), t.Scale()
		overwriteShapeStreamCdat(xf, "ADBE Vector Wiggler Anchor", encodeF64sBE(a[0], a[1]))
		overwriteShapeStreamCdat(xf, "ADBE Vector Wiggler Position", encodeF64sBE(p[0], p[1]))
		overwriteShapeStreamCdat(xf, "ADBE Vector Wiggler Scale", encodeF64sBE(s[0], s[1]))
		overwriteShapeStreamCdat(xf, "ADBE Vector Wiggler Rotation", encodeF64sBE(t.Rotation()))
	}
	// Synthesis-insert the AE-default-elided modulation leaves, each before
	// Random Seed (canonical Wiggler order is Xform Temporal Freq / Correlation /
	// Temporal Phase / Spatial Phase / Random Seed / Wiggler Transform). The
	// leaves are shared with Wiggle Paths (same match-names + layouts).
	if n.CorrelationSet() && n.Correlation() != 50 {
		tdmn, tdbs, err := cloneShapeCorrelationLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Random Seed", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Correlation", encodeF64sBE(n.Correlation()))
	}
	if n.TemporalPhaseSet() && n.TemporalPhase() != 0 {
		tdmn, tdbs, err := cloneShapeTemporalPhaseLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Random Seed", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Temporal Phase", encodeF64sBE(n.TemporalPhase()))
	}
	if n.SpatialPhaseSet() && n.SpatialPhase() != 0 {
		tdmn, tdbs, err := cloneShapeSpatialPhaseLeaf()
		if err != nil {
			return nil, err
		}
		spliceShapeLeafBefore(body, "ADBE Vector Random Seed", tdmn, tdbs)
		overwriteShapeStreamCdat(body, "ADBE Vector Spatial Phase", encodeF64sBE(n.SpatialPhase()))
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

// animateGradientStops converts a static gradient-colors stream in an embedded
// body to animated. It finds the tdmn matching streamName, descends into the
// following LIST(GCst), and (a) patches the tdb4 static→animated flags (same
// three bits as injectAnimatedStream — @0x05 clear bit0, @0x44=1, @0x4f clear
// bit0), (b) replaces the static cdat in the inner LIST(tdbs) with a keyframe
// time-table LIST(list)(lhd3+ldat), and (c) replaces the single GCky/Utf8 with
// one Utf8 leaf per keyframe (that keyframe's gradient as prop.map XML). The
// per-keyframe value lives in the parallel GCky/Utf8 list, not in the ldat —
// structurally the same split as path keyframes (om-s time-table + shap leaves).
// RE: incidents/gradient-fill-write-re.md § animated color stops.
func animateGradientStops(body *rifx.Chunk, streamName string, kfs []GradientKeyframe, ctx *lowerCtx) error {
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || trimChunkNUL(kids[i].Data) != streamName {
			continue
		}
		gcst := kids[i+1]
		if !gcst.IsList() || gcst.FormType != rifx.IDGCst {
			return fmt.Errorf("animateGradientStops: %s next chunk not LIST(GCst)", streamName)
		}
		times := make([]float64, len(kfs))
		for j, kf := range kfs {
			times[j] = kf.Time
		}
		timeTable, err := encodeGradientColorTimeTable(times, ctx)
		if err != nil {
			return err
		}
		var tdbs, gcky *rifx.Chunk
		for _, c := range gcst.Children {
			if c.IsList() && c.FormType == rifx.IDTdbs {
				tdbs = c
			}
			if c.IsList() && c.FormType == rifx.IDGCky {
				gcky = c
			}
		}
		if tdbs == nil || gcky == nil {
			return fmt.Errorf("animateGradientStops: %s missing tdbs/GCky", streamName)
		}
		// (a) tdb4 static→animated flags.
		if tdb4 := findChildID(tdbs, rifx.ChunkID{'t', 'd', 'b', '4'}); tdb4 != nil && len(tdb4.Data) > 0x4f {
			tdb4.Data[0x05] &^= 0x01
			tdb4.Data[0x44] = 0x01
			tdb4.Data[0x4f] &^= 0x01
		}
		// (b) replace the static cdat with the keyframe time-table.
		replaced := false
		for j, c := range tdbs.Children {
			if c.ID == rifx.IDCdat {
				tdbs.Children[j] = timeTable
				replaced = true
				break
			}
		}
		if !replaced {
			return fmt.Errorf("animateGradientStops: %s no cdat to replace in tdbs", streamName)
		}
		// (c) replace the GCky's single Utf8 with one per keyframe.
		gcky.Children = gcky.Children[:0]
		for _, kf := range kfs {
			gcky.Children = append(gcky.Children, &rifx.Chunk{
				ID:   rifx.IDUtf8,
				Data: []byte(codec.EncodeGradientXML(kf.Gradient)),
			})
		}
		return nil
	}
	return fmt.Errorf("animateGradientStops: %s tdmn not found", streamName)
}

// encodeGradientColorTimeTable builds the keyframe time-table LIST(list)(lhd3+
// ldat) for animated gradient color stops. The per-keyframe block is a
// gradient-specific bpk-64 record (NOT the scalar/spatial layouts): time @0x00
// (round(sec*tickRate)), linear interp bytes @0x04/0x05, headerByte 0x01 @0x07,
// a constant 0x00000002 @0x08, and an f64 1.0 @0x10 + tangent-scratch @0x38 —
// the latter two replicated verbatim from the AE-saved oracle fixture
// (v2_2_gradient_anim_src.aep); the gradient VALUES live in the parallel //nolint:jargon
// GCky/Utf8 leaves, so this table carries only timing/interp. lhd3 mirrors
// encodeKeyframes (magic / numKf @0x08 / pages @0x0C / bpk @0x10 / 4×pages @0x1C).
func encodeGradientColorTimeTable(times []float64, ctx *lowerCtx) (*rifx.Chunk, error) {
	if len(times) == 0 {
		return nil, fmt.Errorf("encodeGradientColorTimeTable: no keyframes")
	}
	if ctx == nil || ctx.tickRate <= 0 {
		return nil, fmt.Errorf("encodeGradientColorTimeTable: lowerCtx.tickRate not set")
	}
	const bpk = 64
	n := len(times)
	pages := uint32((n + 3) / 4)

	lhd3 := make([]byte, 52)
	lhd3[0], lhd3[1], lhd3[2], lhd3[3] = 0x00, 0xd0, 0x0b, 0xee
	binary.BigEndian.PutUint32(lhd3[0x08:0x0C], uint32(n))
	binary.BigEndian.PutUint32(lhd3[0x0C:0x10], pages)
	binary.BigEndian.PutUint32(lhd3[0x10:0x14], bpk)
	binary.BigEndian.PutUint32(lhd3[0x14:0x18], 4)
	binary.BigEndian.PutUint32(lhd3[0x18:0x1C], 1)
	binary.BigEndian.PutUint32(lhd3[0x1C:0x20], 4*pages)

	ldat := make([]byte, n*bpk)
	for i, t := range times {
		blk := ldat[i*bpk : (i+1)*bpk]
		binary.BigEndian.PutUint32(blk[0x00:0x04], uint32(math.Round(t*ctx.tickRate)))
		blk[0x04] = byte(InterpLinear)
		blk[0x05] = byte(InterpLinear)
		blk[0x07] = 0x01
		binary.BigEndian.PutUint32(blk[0x08:0x0C], 2)
		binary.BigEndian.PutUint64(blk[0x10:0x18], math.Float64bits(1.0))
		// @0x38 tangent-scratch (verbatim from the oracle's first record).
		blk[0x38], blk[0x39], blk[0x3A], blk[0x3B] = 0x80, 0x80, 0x9f, 0xbe
	}

	kfList := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
	kfList.Children = append(kfList.Children,
		&rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3},
		&rifx.Chunk{ID: rifx.IDLdat, Data: ldat},
	)
	return kfList, nil
}

// lowerStrokeNode emits a Stroke graphic body using embedded tolerance
// bytes (templates/shapes/stroke_body.bin). Same rationale as the other
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
	// value@0x08, no normalization; RE'd from v2_2_stroke_kf_re.aep). The stroke //nolint:jargon
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
// and absent from the template (not currently modeled).
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
// and absent from the template (not currently modeled).
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
// This implementation maps the runtime `shapeRootGroup.Children = [Rect,
// Fill, ...]` to a SINGLE Vector Group wrapper (semantic = AE's
// auto-created "Group 1"). A future version may expose multiple
// user-named groups.
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
