// internal/aep/lower_property_stream.go
//
// Phase 2 Task 2.1 — 5 typed lowering primitives that turn PropertyStream[T]
// into a LIST(tdgp) chunk subtree (tdmn + LIST(tdbs)(tdsb + tdsn + tdb4 +
// cdat OR LIST(list)(lhd3 + ldat))).
//
// Per V2.2 strategy (see spec §3.6 / §4.4): the builder ALWAYS emits cdat /
// the full keyframe substructure even when values equal AE defaults. AE
// elides defaults in its own writer; we don't replicate that (V3 may revisit).
// AE accepts the non-elided form. Phase 4 roundtrip is the byte-exact gate.
//
// Byte layouts are sourced from Phase 0 RE findings recorded in
// `flightdeck/specs/2026-05-22-v2-2-layer-creation-design.md` §8 (RE-S2 / S5a / S6 / S7 /
// S8). Where a single byte was observed but its meaning is unverified, the
// constant carries an "observed" comment and the layout falls back to a
// canonical fixture value rather than a synthesized guess.
package aep

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// lowerCtx is the serializer-side lowering state (spec §4.2). Per Inv-2 +
// "Inv strengthening: lowerCtx carries lowering state only", it does NOT
// hold runtime-graph handles. tickRate converts seconds → ticks for keyframe
// time fields; capabilities is reserved for V3 (currently always empty —
// see capability_matrix.go); nextLayerID is plumbed for Phase 3.
type lowerCtx struct {
	tickRate     float64
	compDuration float64 // seconds — owning comp's Duration (for ldta out-point)
	capabilities AECapabilities
	nextLayerID  func() uint32
}

// NewLowerCtxForTest exports a default lowerCtx for unit tests. Production
// code should not call this — lowerCtx is constructed internally by Phase
// 3+ entry points (NewShapeLayer, WriteAEP).
func NewLowerCtxForTest() *lowerCtx {
	return &lowerCtx{
		tickRate:     30720, // AE 30 fps default per RE-S6 (cdta @0x08)
		capabilities: Capabilities(TargetAE2020),
	}
}

// --- Public typed lowering API -------------------------------------------

// LowerFloat64Stream emits LIST(tdgp) for a 1D PropertyStream[float64].
func LowerFloat64Stream(ps *PropertyStream[float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream[float64](ps.mode, encode1D, valueLayout{dim: 1, headerByte: 0x00, spatial: false}, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerVec2Stream emits LIST(tdgp) for a 2D PropertyStream[[2]float64].
// Hot-path use is Layer Position (spatial 2D, per RE-S6 header07=0x07).
func LowerVec2Stream(ps *PropertyStream[[2]float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream[[2]float64](ps.mode, encode2D, valueLayout{dim: 2, headerByte: 0x07, spatial: true}, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerVec3Stream emits LIST(tdgp) for a 3D PropertyStream[[3]float64].
func LowerVec3Stream(ps *PropertyStream[[3]float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream[[3]float64](ps.mode, encode3D, valueLayout{dim: 3, headerByte: 0x07, spatial: true}, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerColorStream emits LIST(tdgp) for a 4D RGBA PropertyStream[[4]float64].
func LowerColorStream(ps *PropertyStream[[4]float64], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	return lowerStream[[4]float64](ps.mode, encode4D, valueLayout{dim: 4, headerByte: 0x01, spatial: false}, matchName, displayName, ps.static, ps.keyframes, ctx)
}

// LowerPathStream emits LIST(tdgp) for a BezierPath PropertyStream. Unlike
// the scalar/vector streams, AE encodes BezierPath via LIST(om-s) holding a
// shap/shph/lhd3/ldat (f32 BE) substructure — NOT cdat float64 (RE-S5b).
func LowerPathStream(ps *PropertyStream[BezierPath], matchName, displayName string, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, makeTdmn(matchName))

	oms := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOmS}

	// First child of om-s: a tdbs holding tdsb + tdsn + tdb4 + cdat(4B flag).
	// Per RE-S5b: tdb4 dim@0x03=1, cdat is 4 bytes 0x00000000 (enable/flag).
	innerTdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs}
	innerTdbs.Children = append(innerTdbs.Children,
		makeTdsb(),
		makeTdsn(displayName),
		makeTdb4(valueLayout{dim: 1, headerByte: 0x00, spatial: false}),
		&rifx.Chunk{ID: rifx.IDCdat, Data: make([]byte, 4)},
	)
	oms.Children = append(oms.Children, innerTdbs)

	// Second child: LIST(omks) holding LIST(shap) holding shph + LIST(list)(lhd3 + ldat) + omtn.
	omks := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOmks}
	shap := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDShap}

	switch ps.mode {
	case StreamModeStatic:
		shph, lhd3, ldat := encodeBezier(ps.static)
		shap.Children = append(shap.Children, shph)
		kfList := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
		kfList.Children = append(kfList.Children, lhd3, ldat)
		shap.Children = append(shap.Children, kfList)
		shap.Children = append(shap.Children, &rifx.Chunk{ID: rifx.IDOmtn})
	case StreamModeAnimated:
		// V2.2 emits the first keyframe's path as the static encoding; full
		// path-keyframe animation is a V2.3+ topic. The chunk shape stays
		// valid because AE accepts a single-shape encoding.
		var p BezierPath
		if len(ps.keyframes) > 0 {
			p = ps.keyframes[0].Value
		}
		shph, lhd3, ldat := encodeBezier(p)
		shap.Children = append(shap.Children, shph)
		kfList := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
		kfList.Children = append(kfList.Children, lhd3, ldat)
		shap.Children = append(shap.Children, kfList)
		shap.Children = append(shap.Children, &rifx.Chunk{ID: rifx.IDOmtn})
	}

	omks.Children = append(omks.Children, shap)
	oms.Children = append(oms.Children, omks)

	tdgp.Children = append(tdgp.Children, oms)
	return tdgp, nil
}

// --- Generic stream core --------------------------------------------------

// valueLayout captures the per-property byte-layout knobs needed by tdb4
// + keyframe encoding (dimension count + RE-S6 header07 byte).
type valueLayout struct {
	dim        int  // 1 / 2 / 3 / 4
	headerByte byte // header07 value: 0x07 spatial / 0x01 4D-style / 0x00 non-spatial
	spatial    bool // mirrors layoutFor (parse_keyframe.go) spatialStyle field
	// motionPath: true for true spatial-motion-path streams (layer/shape
	// Position) — the keyframe block carries a 0x00000001 marker at 0x08.
	// Color uses the spatial block shape but is NOT a motion path (0x08 = 0).
	motionPath bool
}

type encodeFunc[T any] func(T) []byte

func encode1D(v float64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(v))
	return b
}

func encode2D(v [2]float64) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[0:8], math.Float64bits(v[0]))
	binary.BigEndian.PutUint64(b[8:16], math.Float64bits(v[1]))
	return b
}

func encode3D(v [3]float64) []byte {
	b := make([]byte, 24)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint64(b[i*8:(i+1)*8], math.Float64bits(v[i]))
	}
	return b
}

func encode4D(v [4]float64) []byte {
	b := make([]byte, 32)
	for i := 0; i < 4; i++ {
		binary.BigEndian.PutUint64(b[i*8:(i+1)*8], math.Float64bits(v[i]))
	}
	return b
}

// lowerStream is the generic core for scalar/vector/color streams. It does
// NOT handle BezierPath — see LowerPathStream for the om-s/shap path.
func lowerStream[T any](
	mode StreamMode,
	enc encodeFunc[T],
	layout valueLayout,
	matchName, displayName string,
	staticVal T,
	keyframes []StreamKeyframe[T],
	ctx *lowerCtx,
) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, makeTdmn(matchName))

	tdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs}
	tdbs.Children = append(tdbs.Children,
		makeTdsb(),
		makeTdsn(displayName),
		makeTdb4(layout),
	)

	switch mode {
	case StreamModeStatic:
		tdbs.Children = append(tdbs.Children, makeCdat(enc(staticVal), layout))
	case StreamModeAnimated:
		kfList, err := encodeKeyframes[T](keyframes, layout, enc, ctx)
		if err != nil {
			return nil, err
		}
		tdbs.Children = append(tdbs.Children, kfList)
	default:
		return nil, fmt.Errorf("lowerStream: unknown StreamMode %v", mode)
	}

	tdgp.Children = append(tdgp.Children, tdbs)
	return tdgp, nil
}

// iter-7: splitVec2Stream removed — lowerLayerTransform no longer
// constructs Position_0/_1 from scratch; tolerance-boilerplate approach
// embeds them ready-formed and lowerLayerTransform only overwrites cdat
// scalar values via overwriteScalarCdat. V2.3 may re-introduce when
// proper Layr Transform keyframe persistence is built.

// --- Chunk builders -------------------------------------------------------

// padMatchName returns a 40-byte NUL-padded ASCII buffer carrying the
// match-name. AE writes tdmn names as a fixed 40-byte slot (RE observation
// across every probed fixture).
func padMatchName(s string) []byte {
	b := make([]byte, 40)
	copy(b, []byte(s))
	return b
}

func makeTdmn(matchName string) *rifx.Chunk {
	return &rifx.Chunk{ID: rifx.IDTdmn, Data: padMatchName(matchName)}
}

// makeTdsb returns a tdsb chunk carrying the standard subprop-flag bits.
// RE-S2/S4/S5a all observed hex `00000001` on user-facing leaf properties.
// (Container groups use other flag values, e.g. `00000401` for Root Vectors
// Group per RE-S3 — handled separately at the call site if needed.)
func makeTdsb() *rifx.Chunk {
	return &rifx.Chunk{ID: rifx.ChunkID{'t', 'd', 's', 'b'}, Data: []byte{0x00, 0x00, 0x00, 0x01}}
}

// makeTdsbContainer returns the `0x00000401` variant observed at user-extensible
// shape-container levels per tolerance.aep iter-5 RE: the Root Vectors Group
// body, the Vectors Group body. AE appears to set the 0x0400 bit to mark
// "this group accepts addProperty()" — non-extensible structural bodies
// (Vector Group routing body, empty placeholders, leaf tdbs) keep 0x00000001.
func makeTdsbContainer() *rifx.Chunk {
	return &rifx.Chunk{ID: rifx.ChunkID{'t', 'd', 's', 'b'}, Data: []byte{0x00, 0x00, 0x04, 0x01}}
}

// makeTdsn returns a tdsn chunk carrying an embedded Utf8 record for the
// display name. tdsn payload format (per parse_shape.go vectorGroupName):
//
//	[ "Utf8" (4) | size uint32 BE (4) | name bytes | optional NUL ]
//
// V2.2 emits the bare format with no trailing NUL — AE accepts.
func makeTdsn(displayName string) *rifx.Chunk {
	nameBytes := []byte(displayName)
	data := make([]byte, 8+len(nameBytes))
	copy(data[0:4], []byte("Utf8"))
	binary.BigEndian.PutUint32(data[4:8], uint32(len(nameBytes)))
	copy(data[8:], nameBytes)
	return &rifx.Chunk{ID: rifx.IDTdsn, Data: data}
}

// tdb4CanonicalHead is the 16-byte tdb4 head observed across every probed
// scalar/vector fixture in RE-S2 / RE-S5a / RE-S5c / RE-S5d:
//
//	db99 [dim_u16_BE] [headerByte byte] [flags 0x00] ...
//
// Bytes 0x04..0x05 = 0x0001 (constant); byte 0x06 = headerByte (RE-S2
// observed 0x07 for Orientation; 0x00 for scalars; etc.); byte 0x07 = 0x00.
// The remaining 108 bytes are AE-internal padding + a trailing constant
// `00 00 78 00` at 0x0C..0x0F (observed everywhere). V2.2 emits the
// canonical 124-byte tdb4 with the head + zero padding to total 124.
func makeTdb4(layout valueLayout) *rifx.Chunk {
	d := make([]byte, 124)
	d[0] = 0xdb
	d[1] = 0x99
	// @0x02..0x03: dim count u16 BE. RE observed 0x0001/0x0002/0x0004 at @0x02-0x03.
	binary.BigEndian.PutUint16(d[2:4], uint16(layout.dim))
	d[4] = 0x00
	d[5] = 0x01 // constant per RE
	d[6] = layout.headerByte
	d[7] = 0x00
	// @0x08..0x0B: observed 0x00000000 (RE-S2 Position_0) or 0xffffffff (RE-S5a Ellipse Size).
	// AE accepts 0; that's what we emit. The 4-byte trailing constant @0x0C..0x0F = `00 00 78 00`
	// (RE-S2 / S5a observed).
	d[0x0C] = 0x00
	d[0x0D] = 0x00
	d[0x0E] = 0x78
	d[0x0F] = 0x00
	// Remaining @0x10..0x7B: zero padding (124 total). @0x78 = expression-disabled byte
	// (V1 parse_properties.go: `tdb4 @0x78 inverted = ExpressionEnabled`). 0 = enabled.
	return &rifx.Chunk{ID: rifx.IDtdb4, Data: d}
}

// makeCdat returns a cdat chunk holding `valueBytes` in the first
// `dim * 8` bytes, padded with zeros up to AE's canonical footprint.
//
// AE's cdat padding (RE-S2 / S5a observed): 40B for dim=1, 48B for dim=2
// non-spatial (RE-S5a Position), 80B for dim=2 with hint-range / spatial
// (RE-S5a Size), 96B for dim=4 (RE-S5d Stroke Color). V2.2 picks a single
// canonical size per dim — value bytes only depend on `dim*8`. The exact
// padding chosen is the AE-observed minimum that's been seen to round-trip:
//
//	dim=1 → 40B   dim=2 → 48B   dim=3 → 56B   dim=4 → 96B
//
// Phase 4 roundtrip will validate; if AE rejects, adjust per-dim padding.
func makeCdat(valueBytes []byte, layout valueLayout) *rifx.Chunk {
	size := canonicalCdatSize(layout.dim)
	d := make([]byte, size)
	copy(d, valueBytes)
	return &rifx.Chunk{ID: rifx.IDCdat, Data: d}
}

func canonicalCdatSize(dim int) int {
	switch dim {
	case 1:
		return 40 // RE-S2 (Position_0), RE-S5c (Opacity), RE-S5d (Width / Opacity)
	case 2:
		return 48 // RE-S5a (Ellipse Position; 6 × f64)
	case 3:
		return 56 // 7 × f64 (extrapolated; Phase 4 may adjust)
	case 4:
		return 96 // RE-S5d (Stroke Color: 12 × f64)
	default:
		return dim * 8
	}
}

// --- Keyframe encoding ----------------------------------------------------

// encodeKeyframes builds the LIST(list)(lhd3 + ldat) container per RE-S6/S7.
// lhd3 is 52 bytes (magic + numKeyframes + bpk + a handful of canonical
// constants). ldat carries N × bpk bytes; bpk depends on the layout style:
//
//	spatial-style (RE-S6): bpk = 0x38 + 3*dim*8 (128 for dim=2)
//	non-spatial:           bpk = 0x08 + 5*dim*8 (48 for dim=1, 88 for dim=2)
//
// Time field within each block = seconds * ctx.tickRate (RE-S6/S7 confirmed).
func encodeKeyframes[T any](kfs []StreamKeyframe[T], layout valueLayout, enc encodeFunc[T], ctx *lowerCtx) (*rifx.Chunk, error) {
	if len(kfs) == 0 {
		return nil, fmt.Errorf("encodeKeyframes: empty keyframe slice (caller should pick StreamModeStatic)")
	}
	if ctx == nil || ctx.tickRate <= 0 {
		return nil, fmt.Errorf("encodeKeyframes: lowerCtx.tickRate not set")
	}

	bpk := bytesPerKeyframe(layout)

	// lhd3 — 52 bytes per RE-S6. Magic + numKeyframes @0x08 + bpk @0x10.
	// Bytes @0x0C/0x14/0x18/0x1C have RE-observed constants; we replicate
	// the most-common values (Phase 0 RE-S6 dump of Layer Position 2D).
	lhd3Data := make([]byte, 52)
	lhd3Data[0] = 0x00
	lhd3Data[1] = 0xd0
	lhd3Data[2] = 0x0b
	lhd3Data[3] = 0xee
	binary.BigEndian.PutUint32(lhd3Data[0x08:0x0C], uint32(len(kfs)))
	binary.BigEndian.PutUint32(lhd3Data[0x0C:0x10], 1) // RE-S6 constant
	binary.BigEndian.PutUint32(lhd3Data[0x10:0x14], uint32(bpk))
	// @0x14..0x1F: RE-S6 observed `00000004 00000001 00000004`. Semantics
	// not pinned; replicating verbatim. TODO: confirm via Phase 4 roundtrip.
	binary.BigEndian.PutUint32(lhd3Data[0x14:0x18], 4)
	binary.BigEndian.PutUint32(lhd3Data[0x18:0x1C], 1)
	binary.BigEndian.PutUint32(lhd3Data[0x1C:0x20], 4)

	// ldat — N × bpk bytes.
	ldatData := make([]byte, len(kfs)*bpk)
	for i, kf := range kfs {
		blk := ldatData[i*bpk : (i+1)*bpk]
		writeKeyframeBlock(blk, kf, layout, enc, ctx)
	}

	kfList := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
	kfList.Children = append(kfList.Children,
		&rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3Data},
		&rifx.Chunk{ID: rifx.IDLdat, Data: ldatData},
	)
	return kfList, nil
}

func bytesPerKeyframe(layout valueLayout) int {
	if layout.spatial {
		return 0x38 + 3*layout.dim*8 // RE-S6: 128 for dim=2
	}
	return 0x08 + 5*layout.dim*8 // mirrors layoutFor / decodeEasing (non-spatial)
}

// writeKeyframeBlock fills a single bpk-byte block per parse_keyframe.go
// layout (kfLayout / decodeEasing). For V2.2 hot path we emit linear-interp
// keyframes with the InEase/OutEase fields carried verbatim from the
// StreamKeyframe — zero ease is the default and matches AE-linear.
func writeKeyframeBlock[T any](blk []byte, kf StreamKeyframe[T], layout valueLayout, enc encodeFunc[T], ctx *lowerCtx) {
	// Time @0x00..0x03 = round(seconds * tickRate).
	binary.BigEndian.PutUint32(blk[0x00:0x04], uint32(math.Round(kf.Time*ctx.tickRate)))
	// In/Out interp bytes (1 = linear by default; AddKeyframeLinear path).
	blk[0x04] = byte(InterpLinear)
	blk[0x05] = byte(InterpLinear)
	blk[0x06] = 0x00
	blk[0x07] = layout.headerByte

	valueBytes := enc(kf.Value)

	if layout.spatial {
		// Spatial-style: ease at 0x18/0x20/0x28/0x30; value at 0x38;
		// spatial tangents follow (RE-S6 / parse_keyframe.go kfLayout).
		// Motion-path streams (Position) carry a 0x00000001 marker at 0x08
		// (RE'd from layer/shape Position kf fixtures); Color does not.
		if layout.motionPath {
			binary.BigEndian.PutUint32(blk[0x08:0x0C], 1)
		}
		binary.BigEndian.PutUint64(blk[0x18:0x20], math.Float64bits(kf.InEase.Speed))
		binary.BigEndian.PutUint64(blk[0x20:0x28], math.Float64bits(kf.InEase.Influence))
		binary.BigEndian.PutUint64(blk[0x28:0x30], math.Float64bits(kf.OutEase.Speed))
		binary.BigEndian.PutUint64(blk[0x30:0x38], math.Float64bits(kf.OutEase.Influence))
		copy(blk[0x38:0x38+len(valueBytes)], valueBytes)
		// In/Out spatial tangents follow at 0x38 + dim*8 / + 2*dim*8 (zeroed = linear).
		return
	}

	// Non-spatial: value at 0x08; per-component temporal ease at
	// 0x08+(dim+i)*8 / +(2*dim+i)*8 / +(3*dim+i)*8 / +(4*dim+i)*8.
	copy(blk[0x08:0x08+len(valueBytes)], valueBytes)
	for i := 0; i < layout.dim; i++ {
		binary.BigEndian.PutUint64(blk[0x08+(layout.dim+i)*8:0x08+(layout.dim+i+1)*8], math.Float64bits(kf.InEase.Speed))
		binary.BigEndian.PutUint64(blk[0x08+(2*layout.dim+i)*8:0x08+(2*layout.dim+i+1)*8], math.Float64bits(kf.InEase.Influence))
		binary.BigEndian.PutUint64(blk[0x08+(3*layout.dim+i)*8:0x08+(3*layout.dim+i+1)*8], math.Float64bits(kf.OutEase.Speed))
		binary.BigEndian.PutUint64(blk[0x08+(4*layout.dim+i)*8:0x08+(4*layout.dim+i+1)*8], math.Float64bits(kf.OutEase.Influence))
	}
}

// --- BezierPath encoding (RE-S5b + RE-S8) ---------------------------------

// encodeBezier emits (shph, lhd3, ldat) for one BezierPath. The encoding is
// the single canonical format per RE-S8: tangents always occupy 24B/vertex
// (6 × f32 BE per vertex; zero tangents are NOT omitted), and coordinates
// are bbox-normalized to 0..1 over the union of (verts, verts+inTan,
// verts+outTan).
func encodeBezier(p BezierPath) (shph, lhd3, ldat *rifx.Chunk) {
	n := len(p.Vertices)
	// Generalize per-vertex tangent slices; treat missing as zero.
	getIn := func(i int) [2]float64 {
		if i < len(p.InTangents) {
			return p.InTangents[i]
		}
		return [2]float64{0, 0}
	}
	getOut := func(i int) [2]float64 {
		if i < len(p.OutTangents) {
			return p.OutTangents[i]
		}
		return [2]float64{0, 0}
	}

	// bbox = min/max over (verts ∪ verts+inTan ∪ verts+outTan).
	var minX, minY, maxX, maxY float64
	first := true
	upd := func(x, y float64) {
		if first {
			minX, maxX, minY, maxY = x, x, y, y
			first = false
			return
		}
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	for i := 0; i < n; i++ {
		v := p.Vertices[i]
		in := getIn(i)
		out := getOut(i)
		upd(v[0], v[1])
		upd(v[0]+in[0], v[1]+in[1])
		upd(v[0]+out[0], v[1]+out[1])
	}
	// Empty path → identity bbox 0..0; avoid division by zero below.
	if first {
		minX, maxX, minY, maxY = 0, 0, 0, 0
	}
	rangeX := maxX - minX
	rangeY := maxY - minY
	norm := func(v, lo, r float64) float32 {
		if r == 0 {
			return 0
		}
		return float32((v - lo) / r)
	}

	// shph (24 B). Layout per RE-S5b/S8:
	//   [0..1] magic 0xb3de; [2..3] flags (0x0201 closed-with-trailer);
	//   [4..7] bboxMinX f32; [8..11] bboxMinY f32;
	//   [12..15] bboxMaxX f32; [16..19] bboxMaxY f32; [20..23] trailer 0x01000000.
	shphData := make([]byte, 24)
	shphData[0] = 0xb3
	shphData[1] = 0xde
	// flags: RE-S5b observed 0x0201 (closed); open shapes likely 0x0200.
	// V2.2 emits 0x0201 when Closed, else 0x0200 — Phase 4 to confirm.
	if p.Closed {
		shphData[2] = 0x02
		shphData[3] = 0x01
	} else {
		shphData[2] = 0x02
		shphData[3] = 0x00
	}
	binary.BigEndian.PutUint32(shphData[4:8], math.Float32bits(float32(minX)))
	binary.BigEndian.PutUint32(shphData[8:12], math.Float32bits(float32(minY)))
	binary.BigEndian.PutUint32(shphData[12:16], math.Float32bits(float32(maxX)))
	binary.BigEndian.PutUint32(shphData[16:20], math.Float32bits(float32(maxY)))
	shphData[20] = 0x01

	// lhd3 (52 B) — RE-S5b observed hex for 4-vertex linear:
	//   00d00bee 00000000 0000000c 00000004 00000008 00000004 00000001 00000010 00...
	// The @0x08 u32 = 12 (= n*3, "stride bytes"?); @0x0C u32 = n (vertex count);
	// @0x10 u32 = 8; @0x14 u32 = 4 (= n? or per-vertex word count?);
	// @0x18 u32 = 1 (closed flag); @0x1C u32 = 16. We replicate verbatim.
	lhd3Data := make([]byte, 52)
	lhd3Data[0] = 0x00
	lhd3Data[1] = 0xd0
	lhd3Data[2] = 0x0b
	lhd3Data[3] = 0xee
	binary.BigEndian.PutUint32(lhd3Data[0x08:0x0C], uint32(n*3))
	binary.BigEndian.PutUint32(lhd3Data[0x0C:0x10], uint32(n))
	binary.BigEndian.PutUint32(lhd3Data[0x10:0x14], 8)
	binary.BigEndian.PutUint32(lhd3Data[0x14:0x18], uint32(n))
	closedFlag := uint32(0)
	if p.Closed {
		closedFlag = 1
	}
	binary.BigEndian.PutUint32(lhd3Data[0x18:0x1C], closedFlag)
	binary.BigEndian.PutUint32(lhd3Data[0x1C:0x20], 16)

	// ldat — 24 B/vertex (6 × f32 BE), bbox-normalized. Per-vertex layout RE'd
	// from AE-native fixtures (decode_path_ldat on v2_2_shape_path_re.aep):
	//   [ anchor_i , anchor_i+outTangent_i , anchor_{(i+1)%n}+inTangent_{(i+1)%n} ]
	// i.e. anchor, THIS vertex's out-control, then the NEXT vertex's in-control
	// (wraps mod n). NOT this vertex's own in/out — that mis-encoding rendered
	// the wrong shape in AE (V2.2.1 RE; path was never ship-gated before).
	ldatData := make([]byte, n*24)
	for i := 0; i < n; i++ {
		v := p.Vertices[i]
		out := getOut(i)
		ni := (i + 1) % n
		nv := p.Vertices[ni]
		nin := getIn(ni)
		base := i * 24
		binary.BigEndian.PutUint32(ldatData[base+0:base+4], math.Float32bits(norm(v[0], minX, rangeX)))
		binary.BigEndian.PutUint32(ldatData[base+4:base+8], math.Float32bits(norm(v[1], minY, rangeY)))
		binary.BigEndian.PutUint32(ldatData[base+8:base+12], math.Float32bits(norm(v[0]+out[0], minX, rangeX)))
		binary.BigEndian.PutUint32(ldatData[base+12:base+16], math.Float32bits(norm(v[1]+out[1], minY, rangeY)))
		binary.BigEndian.PutUint32(ldatData[base+16:base+20], math.Float32bits(norm(nv[0]+nin[0], minX, rangeX)))
		binary.BigEndian.PutUint32(ldatData[base+20:base+24], math.Float32bits(norm(nv[1]+nin[1], minY, rangeY)))
	}

	shph = &rifx.Chunk{ID: rifx.IDShph, Data: shphData}
	lhd3 = &rifx.Chunk{ID: rifx.IDLhd3, Data: lhd3Data}
	ldat = &rifx.Chunk{ID: rifx.IDLdat, Data: ldatData}
	return shph, lhd3, ldat
}
