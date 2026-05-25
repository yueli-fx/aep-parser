// internal/aep/lower_layer.go
//
// Phase 2 Task 2.3 — lowerShapeLayer produces a LIST(Layr) chunk from a
// runtime *ShapeLayer. Children per RE-S1 (empty ShapeLayer canonical):
//
//	ldta (160 B AE 2020 canonical)
//	Utf8 (layer name, length-variable)
//	LIST(tdgp) — layer Transform Group (V2.2: user-facing 2D 5-stream)
//	[if root has shapes:] tdmn("ADBE Root Vectors Group") + LIST(tdgp, root)
//
// Per RE-S2 the Phase-0 fixture observed a 6-axis 3D-compatible Transform
// schema (Position_0 / Position_1 / Orientation / RotateX / RotateY /
// Envir Appear). V2.2 instead emits the user-facing 2D form (Anchor /
// Position / Scale / Rotate Z / Opacity) per spec §3.3a + V1 parser
// convention — AE accepts both per the prompt's hot-path 2D guidance.
// Phase 4 roundtrip is the AE accept gate.
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

// v2_2 ShapeLayer Transform Group body — byte-exact extracted from
// tolerance.aep (1842 B LIST(tdgp) with 15 children: tdsb + tdsn + 6 stream
// tdmn-LIST pairs + Group End). iter-7 ship-gate: transplant tests proved
// constructing this byte-correctly from scratch is too fragile (silent-drop
// trigger). V2.3 may RE the full byte layout and replace this blob with
// constructor code.
//
//go:embed templates/v2_2_transform_group_body.bin
var v22TransformGroupBodyBytes []byte

var (
	v22TransformGroupOnce  sync.Once
	v22TransformGroupCache *rifx.Chunk
	v22TransformGroupErr   error
)

// cloneShapeTransformGroupBody returns a deep clone of the cached tolerance
// Transform Group body. Caller may modify the returned tree freely (typically
// to overwrite Position_0/_1 cdat with runtime values).
func cloneShapeTransformGroupBody() (*rifx.Chunk, error) {
	v22TransformGroupOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(v22TransformGroupBodyBytes))
		if err != nil {
			v22TransformGroupErr = fmt.Errorf("parse v22TransformGroupBodyBytes: %w", err)
			return
		}
		v22TransformGroupCache = ch
	})
	if v22TransformGroupErr != nil {
		return nil, v22TransformGroupErr
	}
	return cloneChunk(v22TransformGroupCache), nil
}

// cloneChunk deep-copies a chunk tree. Caller modifications to clone don't
// affect the cached source.
func cloneChunk(c *rifx.Chunk) *rifx.Chunk {
	out := &rifx.Chunk{
		ID:       c.ID,
		Size:     c.Size,
		FormType: c.FormType,
		Trailing: append([]byte(nil), c.Trailing...),
	}
	if c.Data != nil {
		out.Data = append([]byte(nil), c.Data...)
	}
	if len(c.Children) > 0 {
		out.Children = make([]*rifx.Chunk, len(c.Children))
		for i, ch := range c.Children {
			out.Children[i] = cloneChunk(ch)
		}
	}
	return out
}

// trimChunkNUL returns the prefix of d up to the first NUL byte.
func trimChunkNUL(d []byte) string {
	for i := 0; i < len(d); i++ {
		if d[i] == 0 {
			return string(d[:i])
		}
	}
	return string(d)
}

// LowerShapeLayerForTest exports lowerShapeLayer for unit tests.
func LowerShapeLayerForTest(s *ShapeLayer) (*rifx.Chunk, error) {
	return lowerShapeLayer(s, NewLowerCtxForTest())
}

// lowerShapeLayer → LIST(Layr).
func lowerShapeLayer(s *ShapeLayer, ctx *lowerCtx) (*rifx.Chunk, error) {
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr}

	// ldta (160 B AE 2020 canonical per RE-S1; capability matrix doesn't
	// branch — AE 2025's +4 zero-pad tail isn't ours to emit, the higher
	// AE just reads our 160 B fine).
	ldta := &rifx.Chunk{ID: rifx.IDLdta, Data: buildLdtaBytes(s, ctx)}
	layr.Children = append(layr.Children, ldta)

	// Utf8 (layer name). Length-variable; written verbatim.
	layr.Children = append(layr.Children, &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(s.Name),
	})

	// Layer-level property groups live INSIDE an outer LIST(tdgp) — AE
	// rejects the flat-Layr-children form (Phase 5 ship gate FAIL repro;
	// scars/v2-2-aelayer-structure.md fix A). Outer body shape per
	// tolerance.aep:
	//   tdsb + tdsn("") + (tdmn + LIST(tdgp))* + tdmn("ADBE Group End")
	//
	// V2.2 minimum outer body emits: Root Vectors Group (when shapes
	// present) + Transform Group. Additional layer-property groups AE
	// emits at default (Layer Styles / Extrsn Options / Material Options
	// / Audio Group / Layer Sets) get added if AE 2020/25 still rejects.
	outer := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	outer.Children = append(outer.Children, makeTdsb(), makeTdsn(""))

	// Root Vectors Group is ALWAYS emitted on ShapeLayer (per iter 4 bisection
	// #2 finding: AE 2025 rejects ShapeLayer even with zero shapes when Root
	// Vectors Group is absent; tolerance.aep dumps confirm AE always emits it).
	// lowerVectorGroup handles an empty VectorGroup (3-child LIST(tdgp): tdsb +
	// tdsn("Contents") + Group End).
	if s.shapeRootGroup == nil {
		s.shapeRootGroup = NewVectorGroup()
	}
	outer.Children = append(outer.Children, makeTdmn("ADBE Root Vectors Group"))
	rootGroupTdgp, err := lowerVectorGroup(s.shapeRootGroup, ctx)
	if err != nil {
		return nil, err
	}
	outer.Children = append(outer.Children, rootGroupTdgp)

	transformWrapper, err := lowerLayerTransform(s.shapeTransform, ctx)
	if err != nil {
		return nil, err
	}
	// transformWrapper.Children[0] = tdmn("ADBE Transform Group");
	// transformWrapper.Children[1..] = the tdgp body chunks. Append both:
	// the tdmn becomes the property-group name marker, the second is the
	// body LIST(tdgp).
	outer.Children = append(outer.Children, transformWrapper.Children[0])
	transformBody := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: transformWrapper.Children[1:]}
	outer.Children = append(outer.Children, transformBody)

	// Layer-property-group placeholders AE 2020/25 emit on every ShapeLayer
	// (tolerance.aep dump line 145-361). Ship-gate iter 1 (fix A only) still
	// rejected — iter 2 adds these placeholders. Layer Styles has the
	// canonical Blend Options + 10 fx/enabled nested-empty structure; the
	// other four are emitted as empty 3-child LIST(tdgp).
	appendLayerStylesPlaceholder(outer)
	outer.Children = append(outer.Children,
		makeTdmn("ADBE Extrsn Options Group"), emptyPropGroup(),
		makeTdmn("ADBE Material Options Group"), emptyPropGroup(),
		makeTdmn("ADBE Audio Group"), emptyPropGroup(),
		makeTdmn("ADBE Layer Sets"), emptyPropGroup(),
	)

	outer.Children = append(outer.Children, makeTdmn("ADBE Group End"))
	layr.Children = append(layr.Children, outer)

	// 4th Layr child: Gide boilerplate. Every AE-saved Layr (user shape +
	// template service layers DLay/SLay/CLay/SecL) carries an identical
	// LIST(Gide) at this position. iter-5 bisect proof: AE 2025 silently
	// drops user Layr from comp.layers when this chunk is absent — even
	// for variant #2 (empty ShapeLayer with no shape kids). AE never
	// reaches shape-content validation; drop happens at layer-instantiation
	// stage.
	//
	// Content (byte-identical across all 10+ Layrs observed in
	// tolerance.aep + minfail_v2.aep template service layers):
	//
	//	[LIST Gide]
	//	  chunk gdta (8 B all zero)
	//	  [LIST list]
	//	    chunk lhd3 (52 B, observed constant)
	//
	// "Gide" likely stands for layer-side guide/handle; lhd3 here is NOT
	// the keyframe-list header form (despite sharing chunk ID). Treated as
	// opaque AE-internal boilerplate.
	layr.Children = append(layr.Children, makeGideBoilerplate())

	return layr, nil
}

// gideLhd3Boilerplate is the 52-byte lhd3 content observed identical across
// every Layr's LIST(Gide → list → lhd3) in tolerance.aep + the template
// service layers. iter-5 RE — no AE doc; treated as opaque constant.
var gideLhd3Boilerplate = []byte{
	0x00, 0xd0, 0x0b, 0xee, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
}

// makeGideBoilerplate returns the constant LIST(Gide) every Layr must carry
// as 4th child. iter-5 finding (see lowerShapeLayer caller comment).
func makeGideBoilerplate() *rifx.Chunk {
	innerList := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDkfl}
	innerList.Children = append(innerList.Children, &rifx.Chunk{
		ID:   rifx.IDLhd3,
		Data: append([]byte(nil), gideLhd3Boilerplate...),
	})
	g := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDGide}
	g.Children = append(g.Children, &rifx.Chunk{ID: rifx.IDGdta, Data: make([]byte, 8)}, innerList)
	return g
}

// emptyPropGroup returns an empty 3-child LIST(tdgp) placeholder:
// [tdsb(0x01), tdsn(""), tdmn("ADBE Group End")]. Used for nested shape
// sub-property placeholders (e.g. Vector Transform Group inside Vector Group;
// Adv Blend Group inside Layer Styles).
func emptyPropGroup() *rifx.Chunk {
	g := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	g.Children = append(g.Children,
		makeTdsb(),
		makeTdsn(""),
		makeTdmn("ADBE Group End"),
	)
	return g
}

// appendLayerStylesPlaceholder appends `tdmn(ADBE Layer Styles) +
// LIST(tdgp, canonical nested structure)` to outer. The canonical body
// per tolerance.aep dump line 145-209 holds:
//   - tdsb + tdsn
//   - tdmn(ADBE Blend Options Group) + LIST(tdgp){ tdsb + tdsn +
//     tdmn(ADBE Adv Blend Group) + emptyPropGroup() + Group End }
//   - 10 × (tdmn(fxName/enabled) + emptyPropGroup())
//   - tdmn(ADBE Group End)
func appendLayerStylesPlaceholder(outer *rifx.Chunk) {
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	body.Children = append(body.Children, makeTdsb(), makeTdsn(""))

	blendOpts := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	blendOpts.Children = append(blendOpts.Children,
		makeTdsb(),
		makeTdsn(""),
		makeTdmn("ADBE Adv Blend Group"), emptyPropGroup(),
		makeTdmn("ADBE Group End"),
	)
	body.Children = append(body.Children,
		makeTdmn("ADBE Blend Options Group"), blendOpts,
	)

	fxNames := []string{
		"dropShadow/enabled",
		"innerShadow/enabled",
		"outerGlow/enabled",
		"innerGlow/enabled",
		"bevelEmboss/enabled",
		"chromeFX/enabled",
		"solidFill/enabled",
		"gradientFill/enabled",
		"patternFill/enabled",
		"frameFX/enabled",
	}
	for _, n := range fxNames {
		body.Children = append(body.Children, makeTdmn(n), emptyPropGroup())
	}
	body.Children = append(body.Children, makeTdmn("ADBE Group End"))

	outer.Children = append(outer.Children, makeTdmn("ADBE Layer Styles"), body)
}

// buildLdtaBytes returns the 160-byte canonical ldta payload for a
// ShapeLayer. Fills the well-known offsets via ldta_layout.go constants;
// everything else is zero (AE-friendly default per RE-S1).
//
// Layer subtype byte (@0x80) = 4 (Shape) per ldta_layout.go ldtaLayerSubtype
// comment. Quality (@0x04) = 2 (Best) — AE's typical default. Visible bit
// is at byte 0x27 bit0 (per parse_layer.go decoder); default Visible = true
// → 0x01.
func buildLdtaBytes(s *ShapeLayer, ctx *lowerCtx) []byte {
	// 164 B (AE 2025 canonical) — trailing 4 B zero. AE 2020 has been observed
	// to accept 164 B too (template's DLay is 160 B, but our user Layr matches
	// AE-saved ShapeLayer fixtures = 164 B).
	d := make([]byte, ldtaSize2025)

	tickRate := uint32(0)
	if ctx != nil && ctx.tickRate > 0 {
		tickRate = uint32(ctx.tickRate)
	}
	if tickRate == 0 {
		tickRate = 30720 // AE default 30 fps tick rate
	}
	duration := 1.0
	if ctx != nil && ctx.compDuration > 0 {
		duration = ctx.compDuration
	}
	outTicks := uint32(duration * float64(tickRate))

	// @0x00 — layer-local ID.
	binary.BigEndian.PutUint32(d[ldtaLayerID:ldtaLayerID+4], s.ID)

	// @0x04 — Quality. 2 = Best (AE default for new layers).
	binary.BigEndian.PutUint16(d[ldtaQuality:ldtaQuality+2], 2)

	// @0x08 — StretchDividend = 1 (per tolerance.aep iter 4 RE).
	// @0x6C — StretchDivisor = 1 (1/1 = 1× speed).
	binary.BigEndian.PutUint32(d[ldtaStretchDivd:ldtaStretchDivd+4], 1)
	binary.BigEndian.PutUint32(d[ldtaStretchDivs:ldtaStretchDivs+4], 1)

	// Time fields are encoded as (ticks_dividend, ticks/sec_divisor). Per
	// iter 4 RE of tolerance.aep: divisor = TickRate (30720 for 30fps), NOT
	// 1. Our previous 0/1 encoding made AE compute zero-duration layers and
	// silently drop them from comp.layers.
	binary.BigEndian.PutUint32(d[ldtaStartTimeDivd:ldtaStartTimeDivd+4], 0)
	binary.BigEndian.PutUint32(d[ldtaStartTimeDivs:ldtaStartTimeDivs+4], tickRate)
	binary.BigEndian.PutUint32(d[ldtaInPointDivd:ldtaInPointDivd+4], 0)
	binary.BigEndian.PutUint32(d[ldtaInPointDivs:ldtaInPointDivs+4], tickRate)
	binary.BigEndian.PutUint32(d[ldtaOutPointDivd:ldtaOutPointDivd+4], outTicks)
	binary.BigEndian.PutUint32(d[ldtaOutPointDivs:ldtaOutPointDivs+4], tickRate)

	// Attr bytes — tolerance.aep ShapeLayer @0x27 = 0x87 (visible + audio +
	// effects + collapse-transform). Per write_layer.go flag map:
	//   bit0 0x01 visible / bit1 0x02 audio-enabled / bit2 0x04 effects-enabled
	//   bit3 0x08 motion-blur / bit7 0x80 collapse-transform
	d[ldtaAttrByte2] = 0x87

	// AttrByte0 @0x25: tolerance has bit0 set (0x01). Not in the documented
	// bit map (parse_layer.go doc covers bit1/2/4/6). Empirically required —
	// AE 2025 silently drops ShapeLayer from comp.layers without it (iter 4
	// RE finding). Speculated as "layer-real" / "ready" / "validated" flag.
	d[0x25] = 0x01

	// @0x3B: tolerance sets to 0x01 (unknown semantics; iter 4 RE match).
	d[0x3B] = 0x01
	// @0x3D: label color index. AE default Shape = 0x08 (per tolerance.aep).
	d[0x3D] = 0x08
	// @0x63: blending mode (parse_layer.go doc). Tolerance ShapeLayer = 0x02.
	d[0x63] = 0x02

	// @0x40 — legacy 32-byte name slot. Mirror up to 31 bytes of name +
	// NUL terminator. Parser ignores this in favor of the Utf8 chunk, but
	// AE's saved fixtures populate it.
	nameBytes := []byte(s.Name)
	maxName := 31
	if len(nameBytes) < maxName {
		maxName = len(nameBytes)
	}
	copy(d[ldtaLegacyName:ldtaLegacyName+maxName], nameBytes[:maxName])

	// @0x80 — LayerSubtype = Shape (4).
	binary.BigEndian.PutUint32(d[ldtaLayerSubtype:ldtaLayerSubtype+4], 4)

	// @0x84 — ParentID. 0 = no parent.
	binary.BigEndian.PutUint32(d[ldtaParentID:ldtaParentID+4], s.ParentID)

	return d
}

// lowerLayerTransform emits the Layer Transform Group LIST(tdgp) for a
// ShapeLayer. iter-7 approach (post-bisect): use the byte-exact Transform
// Group body extracted from tolerance.aep as boilerplate, then overwrite
// the Position_0/_1 inner cdat with runtime t.position values.
//
// Background: iter-5b..iter-6f attempted to construct the Transform Group
// from scratch (Anchor / Position_0/_1 / Scale / RotateZ / Opacity + 6-axis
// 3D defaults). AE 2025 accepted those files but silently dropped the
// ShapeLayer from comp.layers. Transplant tests (tmp_debug/swap_propgroup)
// isolated the silent-drop trigger to the Transform Group body alone —
// 4 other property groups (Root Vectors, Layer Styles, Extrsn/Material/
// Audio/Layer Sets placeholders) emit byte-identically to tolerance and
// pass AE acceptance; only Transform Group construction had subtle
// byte errors (tdsb 0x03 vs 0x01, tdb4 head bytes, missing tdum/tduM,
// over-emit of Anchor/Scale/RotateZ/Opacity as Vec2 instead of 3D).
//
// V2.2 ship gate uses verbatim tolerance bytes; runtime user-set values
// for non-Position streams (Anchor / Scale / Rotation / Opacity) are
// runtime-only — they don't persist to disk in V2.2. V2.3 will RE the
// proper byte layout for full Transform persistence.
func lowerLayerTransform(t *LayerTransform, _ *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeTransformGroupBody()
	if err != nil {
		return nil, err
	}
	// Overwrite Position_0 / Position_1 cdat values with runtime user input.
	// Tolerance's Position_0/_1 cdat are 40B with the f64 value at bytes 0..7.
	overwriteScalarCdat(body, MatchNamePosition0, t.position.static[0])
	overwriteScalarCdat(body, MatchNamePosition1, t.position.static[1])
	if t.position.mode == StreamModeAnimated && len(t.position.keyframes) > 0 {
		// Fall back to first keyframe value as static slot. Full animated
		// persistence on Layr Transform is V2.3 work.
		overwriteScalarCdat(body, MatchNamePosition0, t.position.keyframes[0].Value[0])
		overwriteScalarCdat(body, MatchNamePosition1, t.position.keyframes[0].Value[1])
	}

	wrapper := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	wrapper.Children = append(wrapper.Children, makeTdmn("ADBE Transform Group"))
	wrapper.Children = append(wrapper.Children, body.Children...)
	return wrapper, nil
}

// overwriteScalarCdat finds the tdmn `name` inside `body` and overwrites the
// first 8 bytes of the inner cdat (scalar value) with the f64 BE encoding of v.
// Used by lowerLayerTransform's iter-7 Position post-process.
func overwriteScalarCdat(body *rifx.Chunk, name string, v float64) {
	kids := body.Children
	for i := 0; i < len(kids); i++ {
		if kids[i].ID == rifx.IDTdmn && trimChunkNUL(kids[i].Data) == name && i+1 < len(kids) {
			tdbs := kids[i+1]
			if !tdbs.IsList() || tdbs.FormType != rifx.IDTdbs {
				return
			}
			for _, ch := range tdbs.Children {
				if ch.ID == rifx.IDCdat && len(ch.Data) >= 8 {
					binary.BigEndian.PutUint64(ch.Data[0:8], math.Float64bits(v))
					return
				}
			}
			return
		}
	}
}

// lowerOrientationDefault emits the LIST(otst) wrapper holding a default
// Orientation stream — tolerance.aep dump line 117-125 canonical shape:
//
//	[LIST otst]
//	  [LIST tdbs]
//	    tdsb (4 B) + tdsn (14 B) + tdb4 (124 B) + cdat (24 B = 3 × f64 = 0,0,0)
//	  [LIST otky]
//	    otda (24 B = 3 × f64 = 0,0,0)
//
// iter-7: lowerOrientationDefault no longer called — Transform Group body is
// now embedded as tolerance bytes (templates/v2_2_transform_group_body.bin)
// and includes its own Orientation otst wrapper. Retired here; keep the
// chunk-IDs (IDOtst/IDOtky/IDOtda) in rifx.go for parser-side use.
