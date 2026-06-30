// Code moved from lower_shape_node.go; keep behavior-only edits out of split commits.
package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

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
