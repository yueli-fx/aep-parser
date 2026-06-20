// internal/aep/lower_layer.go
//
// lowerShapeLayer produces a LIST(Layr) chunk from a runtime *ShapeLayer.
// Children of a canonical empty ShapeLayer:
//
//	ldta (160 B AE 2020 canonical)
//	Utf8 (layer name, length-variable)
//	LIST(tdgp) — layer Transform Group (V2.2: user-facing 2D 5-stream)
//	[if root has shapes:] tdmn("ADBE Root Vectors Group") + LIST(tdgp, root)
//
// AE's canonical fixture observed a 6-axis 3D-compatible Transform schema
// (Position_0 / Position_1 / Orientation / RotateX / RotateY / Envir Appear).
// V2.2 instead emits the user-facing 2D form (Anchor / Position / Scale /
// Rotate Z / Opacity) per V1 parser convention — AE accepts both.
package serializer

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// v2_2 ShapeLayer Transform Group body — byte-exact extracted from
// tolerance.aep (1842 B LIST(tdgp) with 15 children: tdsb + tdsn + 6 stream
// tdmn-LIST pairs + Group End). Transplant tests proved constructing this
// byte-correctly from scratch is too fragile (silent-drop trigger). V2.3 may
// RE the full byte layout and replace this blob with constructor code.
//
//go:embed templates/layers/transform_group_body.bin
var transformGroupBodyBytes []byte

var (
	transformGroupOnce  sync.Once
	transformGroupCache *rifx.Chunk
	transformGroupErr   error
)

// cloneShapeTransformGroupBody returns a deep clone of the cached tolerance
// Transform Group body. Caller may modify the returned tree freely (typically
// to overwrite Position_0/_1 cdat with runtime values).
func cloneShapeTransformGroupBody() (*rifx.Chunk, error) {
	transformGroupOnce.Do(func() {
		ch, err := rifx.ReadChunk(bytes.NewReader(transformGroupBodyBytes))
		if err != nil {
			transformGroupErr = fmt.Errorf("parse transformGroupBodyBytes: %w", err)
			return
		}
		transformGroupCache = ch
	})
	if transformGroupErr != nil {
		return nil, transformGroupErr
	}
	return cloneChunk(transformGroupCache), nil
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

// LowerShapeLayerForTargetForTest lowers a ShapeLayer with the capability set
// of the given AE target — used to assert target-conditional serialization
// (e.g. ldta size: 160 B for AE 2020/2022, 164 B for AE 2025).
func LowerShapeLayerForTargetForTest(s *ShapeLayer, target AETarget) (*rifx.Chunk, error) {
	ctx := &lowerCtx{
		tickRate:     30720,
		capabilities: Capabilities(target),
	}
	return lowerShapeLayer(s, ctx)
}

// lowerShapeLayer → LIST(Layr).
func lowerShapeLayer(s *ShapeLayer, ctx *lowerCtx) (*rifx.Chunk, error) {
	layr := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDLayr}

	// ldta — target-conditional size (capability matrix LdtaSize): 160 B for
	// AE 2020/22, 164 B for AE 2025. AE 2020 rejects a 164-B ldta as corrupt.
	ldta := &rifx.Chunk{ID: rifx.IDLdta, Data: buildLdtaBytes(s, ctx)}
	layr.Children = append(layr.Children, ldta)

	// Utf8 (layer name). Length-variable; written verbatim.
	layr.Children = append(layr.Children, &rifx.Chunk{
		ID:   rifx.IDUtf8,
		Data: []byte(s.Name),
	})

	// Layer-level property groups live INSIDE an outer LIST(tdgp) — AE
	// rejects the flat-Layr-children form. Outer body shape per
	// tolerance.aep:
	//   tdsb + tdsn("") + (tdmn + LIST(tdgp))* + tdmn("ADBE Group End")
	//
	// V2.2 minimum outer body emits: Root Vectors Group (when shapes
	// present) + Transform Group. Additional layer-property groups AE
	// emits at default (Layer Styles / Extrsn Options / Material Options
	// / Audio Group / Layer Sets) get added if AE 2020/25 still rejects.
	outer := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	outer.Children = append(outer.Children, makeTdsb(), makeTdsn(""))

	// Root Vectors Group is ALWAYS emitted on ShapeLayer: AE 2025 rejects
	// ShapeLayer even with zero shapes when Root Vectors Group is absent;
	// tolerance.aep dumps confirm AE always emits it. lowerVectorGroup handles
	// an empty VectorGroup (3-child LIST(tdgp): tdsb + tdsn("Contents") +
	// Group End).
	if s.RootGroup() == nil {
		scene.SetLayerShapeRootGroup(s.Layer, NewVectorGroup())
	}
	outer.Children = append(outer.Children, makeTdmn("ADBE Root Vectors Group"))
	rootGroupTdgp, err := lowerVectorGroup(s.RootGroup(), ctx)
	if err != nil {
		return nil, err
	}
	outer.Children = append(outer.Children, rootGroupTdgp)

	transformWrapper, err := lowerLayerTransform(s.Transform(), ctx)
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
	// (tolerance.aep dump line 145-361). AE rejects the layer without these
	// placeholders. Layer Styles has the canonical Blend Options + 10
	// fx/enabled nested-empty structure; the other four are emitted as empty
	// 3-child LIST(tdgp).
	appendLayerStylesPlaceholder(outer)
	outer.Children = append(outer.Children,
		makeTdmn("ADBE Extrsn Options Group"), emptyPropGroupFlags(0x03),
		makeTdmn("ADBE Material Options Group"), emptyPropGroupFlags(0x03),
		makeTdmn("ADBE Audio Group"), emptyPropGroupFlags(0x03),
		makeTdmn("ADBE Layer Sets"), emptyPropGroupFlags(0x03),
	)

	outer.Children = append(outer.Children, makeTdmn("ADBE Group End"))
	layr.Children = append(layr.Children, outer)

	// 4th Layr child: Gide boilerplate. Every AE-saved Layr (user shape +
	// template service layers DLay/SLay/CLay/SecL) carries an identical
	// LIST(Gide) at this position. AE 2025 silently drops user Layr from
	// comp.layers when this chunk is absent — even for an empty ShapeLayer
	// with no shape kids. AE never reaches shape-content validation; drop
	// happens at layer-instantiation stage.
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
// service layers. No AE doc; treated as opaque constant.
var gideLhd3Boilerplate = []byte{
	0x00, 0xd0, 0x0b, 0xee, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
}

// makeGideBoilerplate returns the constant LIST(Gide) every Layr must carry
// as 4th child (see lowerShapeLayer caller comment).
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
	return emptyPropGroupFlags(0x01)
}

// emptyPropGroupFlags is emptyPropGroup with an explicit tdsb flag word —
// see makeTdsbFlags for why the Layer Styles family must not use 0x01.
func emptyPropGroupFlags(flags uint32) *rifx.Chunk {
	g := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	g.Children = append(g.Children,
		makeTdsbFlags(flags),
		makeTdsn(""),
		makeTdmn("ADBE Group End"),
	)
	return g
}

// appendLayerStylesPlaceholder appends `tdmn(ADBE Layer Styles) +
// LIST(tdgp, canonical nested structure)` to outer. The canonical body
// per tolerance.aep dump line 145-209 holds:
//   - tdsb(0x03) + tdsn
//   - tdmn(ADBE Blend Options Group) + LIST(tdgp){ tdsb(0x03) + tdsn +
//     tdmn(ADBE Adv Blend Group) + emptyPropGroup() + Group End }
//   - 10 × (tdmn(fxName/enabled) + emptyPropGroupFlags(0x02))
//   - tdmn(ADBE Group End)
//
// The tdsb words are load-bearing: bit0 = enabled. fx/enabled groups must be
// 0x02 (present, OFF) — 0x01 makes AE render all 10 styles (red solid-fill +
// bevel collapse) while every readable DOM value still looks correct.
func appendLayerStylesPlaceholder(outer *rifx.Chunk) {
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	body.Children = append(body.Children, makeTdsbFlags(0x03), makeTdsn(""))

	blendOpts := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	blendOpts.Children = append(blendOpts.Children,
		makeTdsbFlags(0x03),
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
		body.Children = append(body.Children, makeTdmn(n), emptyPropGroupFlags(0x02))
	}
	body.Children = append(body.Children, makeTdmn("ADBE Group End"))

	outer.Children = append(outer.Children, makeTdmn("ADBE Layer Styles"), body)
}

// buildLdtaBytes returns the ldta payload for a ShapeLayer, sized per the
// target's capability matrix (160 B AE 2020/22, 164 B AE 2025). Fills the
// well-known offsets via ldta_layout.go constants; everything else is zero
// (AE-friendly default).
//
// Layer subtype byte (@0x80) = 4 (Shape) per ldta_layout.go codec.LdtaLayerSubtype
// comment. Quality (@0x04) = 2 (Best) — AE's typical default. Visible bit
// is at byte 0x27 bit0 (per parse_layer.go decoder); default Visible = true
// → 0x01.
func buildLdtaBytes(s *ShapeLayer, ctx *lowerCtx) []byte {
	// ldta size is target-conditional (capability matrix LdtaSize): AE 2020/22
	// accept 160 B, AE 2025 accepts its native 164 B. Emitting 164 B for an
	// AE 2020 target makes AE 2020 reject the layer as corrupt and skip it.
	// All written fields fit in the first 0x88 bytes, so the size choice only
	// varies the trailing zero-pad. Fall back to 164 if a caller left LdtaSize
	// unset.
	size := codec.LdtaSize2025
	if ctx != nil && ctx.capabilities.LdtaSize > 0 {
		size = ctx.capabilities.LdtaSize
	}
	d := make([]byte, size)

	tickRate := uint32(0)
	if ctx != nil && ctx.tickRate > 0 {
		tickRate = uint32(ctx.tickRate)
	}
	if tickRate == 0 {
		tickRate = 30720 // AE default 30 fps tick rate
	}
	// Layer timeline span. A fresh layer fills the whole comp; a caller can trim
	// it by setting the scene layer's StartTime/Duration (e.g. replication copying
	// the original's in/out). Honored opt-in — callers that leave them 0 keep the
	// old full-comp behavior, so no existing gate regresses.
	startTime := s.StartTime
	duration := 1.0
	if ctx != nil && ctx.compDuration > 0 {
		duration = ctx.compDuration
	}
	if s.Duration > 0 {
		duration = s.Duration
	}
	startTicks := uint32(startTime * float64(tickRate))
	outTicks := uint32((startTime + duration) * float64(tickRate))

	// @0x00 — layer-local ID.
	binary.BigEndian.PutUint32(d[codec.LdtaLayerID:codec.LdtaLayerID+4], s.ID)

	// @0x04 — Quality. 2 = Best (AE default for new layers).
	binary.BigEndian.PutUint16(d[codec.LdtaQuality:codec.LdtaQuality+2], 2)

	// @0x08 — StretchDividend = 1 (per tolerance.aep).
	// @0x6C — StretchDivisor = 1 (1/1 = 1× speed).
	binary.BigEndian.PutUint32(d[codec.LdtaStretchDivd:codec.LdtaStretchDivd+4], 1)
	binary.BigEndian.PutUint32(d[codec.LdtaStretchDivs:codec.LdtaStretchDivs+4], 1)

	// Time fields are encoded as (ticks_dividend, ticks/sec_divisor). Per
	// tolerance.aep: divisor = TickRate (30720 for 30fps), NOT 1. A 0/1
	// encoding makes AE compute zero-duration layers and silently drop them
	// from comp.layers.
	binary.BigEndian.PutUint32(d[codec.LdtaStartTimeDivd:codec.LdtaStartTimeDivd+4], startTicks)
	binary.BigEndian.PutUint32(d[codec.LdtaStartTimeDivs:codec.LdtaStartTimeDivs+4], tickRate)
	binary.BigEndian.PutUint32(d[codec.LdtaInPointDivd:codec.LdtaInPointDivd+4], startTicks)
	binary.BigEndian.PutUint32(d[codec.LdtaInPointDivs:codec.LdtaInPointDivs+4], tickRate)
	binary.BigEndian.PutUint32(d[codec.LdtaOutPointDivd:codec.LdtaOutPointDivd+4], outTicks)
	binary.BigEndian.PutUint32(d[codec.LdtaOutPointDivs:codec.LdtaOutPointDivs+4], tickRate)

	// Attr bytes — tolerance.aep ShapeLayer @0x27 = 0x87 (visible + audio +
	// effects + collapse-transform). Per write_layer.go flag map:
	//   bit0 0x01 visible / bit1 0x02 audio-enabled / bit2 0x04 effects-enabled
	//   bit3 0x08 motion-blur / bit7 0x80 collapse-transform
	d[codec.LdtaAttrByte2] = 0x87

	// AttrByte0 @0x25: tolerance has bit0 set (0x01). Not in the documented
	// bit map (parse_layer.go doc covers bit1/2/4/6). Empirically required —
	// AE 2025 silently drops ShapeLayer from comp.layers without it.
	// Speculated as "layer-real" / "ready" / "validated" flag.
	d[0x25] = 0x01

	// @0x3B: tolerance sets to 0x01 (unknown semantics).
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
	copy(d[codec.LdtaLegacyName:codec.LdtaLegacyName+maxName], nameBytes[:maxName])

	// @0x80 — LayerSubtype = Shape (4).
	binary.BigEndian.PutUint32(d[codec.LdtaLayerSubtype:codec.LdtaLayerSubtype+4], 4)

	// @0x84 — ParentID. 0 = no parent.
	binary.BigEndian.PutUint32(d[codec.LdtaParentID:codec.LdtaParentID+4], s.ParentID)

	return d
}

// lowerLayerTransform emits the Layer Transform Group LIST(tdgp) for a
// ShapeLayer. Uses the byte-exact Transform Group body extracted from
// tolerance.aep as boilerplate, then overwrites the Position_0/_1 inner cdat
// with runtime t.Position() values.
//
// Background: constructing the Transform Group from scratch (Anchor /
// Position_0/_1 / Scale / RotateZ / Opacity + 6-axis 3D defaults) made AE 2025
// accept the file but silently drop the ShapeLayer from comp.layers.
// Transplant tests isolated the silent-drop trigger to the Transform Group
// body alone — 4 other property groups (Root Vectors, Layer Styles,
// Extrsn/Material/Audio/Layer Sets placeholders) emit byte-identically to
// tolerance and pass AE acceptance; only Transform Group construction had
// subtle byte errors (tdsb 0x03 vs 0x01, tdb4 head bytes, missing tdum/tduM,
// over-emit of Anchor/Scale/RotateZ/Opacity as Vec2 instead of 3D).
//
// V2.2 ship gate uses verbatim tolerance bytes; runtime user-set values
// for non-Position streams (Anchor / Scale / Rotation / Opacity) are
// runtime-only — they don't persist to disk in V2.2. V2.3 will RE the
// proper byte layout for full Transform persistence.
func lowerLayerTransform(t *LayerTransform, ctx *lowerCtx) (*rifx.Chunk, error) {
	body, err := cloneShapeTransformGroupBody()
	if err != nil {
		return nil, err
	}
	// Each transform channel is persisted to disk in its RE'd on-disk encoding
	// (test_data/v2_2_transform_kf_re.aep): Anchor/Position are 3D spatial
	// motion-path ([x,y,0]); Scale is 3D non-spatial (÷100, Z=1.0); Rotation is
	// 1D non-spatial (degrees); Opacity is 1D non-spatial (÷100). Animated →
	// inject keyframes; otherwise overwrite the embedded template's static cdat.
	if err := lowerTransformVec2Spatial(body, MatchNameAnchorPoint, t.AnchorPoint(), ctx); err != nil {
		return nil, err
	}
	if err := lowerTransformVec2Spatial(body, MatchNamePosition, t.Position(), ctx); err != nil {
		return nil, err
	}
	if err := lowerTransformScale(body, t.Scale(), ctx); err != nil {
		return nil, err
	}
	if err := lowerTransformScalar(body, MatchNameRotateZ, t.Rotation(), ctx, 1); err != nil {
		return nil, err
	}
	if err := lowerTransformScalar(body, MatchNameOpacity, t.Opacity(), ctx, 0.01); err != nil {
		return nil, err
	}

	wrapper := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	wrapper.Children = append(wrapper.Children, makeTdmn("ADBE Transform Group"))
	wrapper.Children = append(wrapper.Children, body.Children...)
	return wrapper, nil
}

// lowerTransformVec2Spatial persists a 2D spatial transform channel (Anchor /
// Position) into the embedded template: 3D spatial motion-path on disk (bpk-128,
// value@0x38 X/Y/Z with Z=0). The runtime API is 2D ([2]float64); Z is pinned 0.
func lowerTransformVec2Spatial(body *rifx.Chunk, name string, ps *codec.PropertyStream[[2]float64], ctx *lowerCtx) error {
	encXYZ := func(v [2]float64) []byte { return encode3D([3]float64{v[0], v[1], 0}) }
	if ps.Mode() == codec.StreamModeAnimated && ps.HasKeyframes() {
		kfList, err := encodeKeyframes(ps.Keyframes(), valueLayout{dim: 3, headerByte: 0x07, spatial: true, motionPath: true}, encXYZ, ctx)
		if err != nil {
			return err
		}
		return injectAnimatedStream(body, name, kfList)
	}
	sv, _ := ps.StaticValue()
	overwriteShapeStreamCdat(body, name, encXYZ(sv))
	return nil
}

// lowerTransformScale persists Scale: 3D non-spatial on disk (bpk-128,
// value@0x08), values are percent÷100 with the Z (depth) component pinned to
// 1.0 (= 100%). Runtime API is 2D percent ([2]float64).
func lowerTransformScale(body *rifx.Chunk, ps *codec.PropertyStream[[2]float64], ctx *lowerCtx) error {
	encScale := func(v [2]float64) []byte { return encode3D([3]float64{v[0] / 100, v[1] / 100, 1}) }
	if ps.Mode() == codec.StreamModeAnimated && ps.HasKeyframes() {
		kfList, err := encodeKeyframes(ps.Keyframes(), valueLayout{dim: 3, headerByte: 0x00, spatial: false}, encScale, ctx)
		if err != nil {
			return err
		}
		return injectAnimatedStream(body, MatchNameScale, kfList)
	}
	sv, _ := ps.StaticValue()
	overwriteShapeStreamCdat(body, MatchNameScale, encScale(sv))
	return nil
}

// lowerTransformScalar persists a 1D non-spatial transform channel (Rotation /
// Opacity): bpk-48, value@0x08. `scale` converts the runtime value to its
// on-disk form (Rotation 1.0 = degrees as-is; Opacity 0.01 = percent÷100).
func lowerTransformScalar(body *rifx.Chunk, name string, ps *codec.PropertyStream[float64], ctx *lowerCtx, scale float64) error {
	enc := func(v float64) []byte { return encode1D(v * scale) }
	if ps.Mode() == codec.StreamModeAnimated && ps.HasKeyframes() {
		kfList, err := encodeKeyframes(ps.Keyframes(), valueLayout{dim: 1, headerByte: 0x00, spatial: false}, enc, ctx)
		if err != nil {
			return err
		}
		return injectAnimatedStream(body, name, kfList)
	}
	sv, _ := ps.StaticValue()
	overwriteShapeStreamCdat(body, name, enc(sv))
	return nil
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
// lowerOrientationDefault no longer called — Transform Group body is now
// embedded as tolerance bytes (templates/layers/transform_group_body.bin) and
// includes its own Orientation otst wrapper. Retired here; keep the chunk-IDs
// (IDOtst/IDOtky/IDOtda) in rifx.go for parser-side use.
