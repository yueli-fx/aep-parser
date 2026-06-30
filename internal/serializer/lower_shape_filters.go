// Code moved from lower_shape_node.go; keep behavior-only edits out of split commits.
package serializer

import (
	"bytes"
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

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
