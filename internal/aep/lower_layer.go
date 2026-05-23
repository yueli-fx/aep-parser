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
	"encoding/binary"

	"github.com/example/aep-parser/internal/rifx"
)

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

	if s.shapeRootGroup != nil && len(s.shapeRootGroup.Children) > 0 {
		outer.Children = append(outer.Children, makeTdmn("ADBE Root Vectors Group"))
		rootGroupTdgp, err := lowerVectorGroup(s.shapeRootGroup, ctx)
		if err != nil {
			return nil, err
		}
		outer.Children = append(outer.Children, rootGroupTdgp)
	}

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

	return layr, nil
}

// emptyPropGroup returns an empty 3-child LIST(tdgp) placeholder:
// [tdsb, tdsn(""), tdmn("ADBE Group End")]. AE emits this shape for every
// layer-property group at default (Audio / Layer Sets / Extrsn / Material).
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
	d := make([]byte, ldtaSize2020)

	// @0x00 — layer-local ID.
	binary.BigEndian.PutUint32(d[ldtaLayerID:ldtaLayerID+4], s.ID)

	// @0x04 — Quality. 2 = Best (AE default for new layers).
	binary.BigEndian.PutUint16(d[ldtaQuality:ldtaQuality+2], 2)

	// @0x08 — StretchDividend. 100 (= 100% stretch).
	binary.BigEndian.PutUint32(d[ldtaStretchDivd:ldtaStretchDivd+4], 100)
	// @0x6C — StretchDivisor. 100 (paired with dividend for 1× speed).
	binary.BigEndian.PutUint32(d[ldtaStretchDivs:ldtaStretchDivs+4], 100)

	// In/Out points: span the full source / comp default. AE writes
	// dividend=0, divisor=1 for fresh layers; out-point dividend = sentinel
	// (-1 cast to u32). Keep both divisors = 1 to avoid div-by-zero in
	// parsers.
	binary.BigEndian.PutUint32(d[ldtaInPointDivd:ldtaInPointDivd+4], 0)
	binary.BigEndian.PutUint32(d[ldtaInPointDivs:ldtaInPointDivs+4], 1)
	binary.BigEndian.PutUint32(d[ldtaOutPointDivd:ldtaOutPointDivd+4], 0xFFFFFFFF)
	binary.BigEndian.PutUint32(d[ldtaOutPointDivs:ldtaOutPointDivs+4], 1)

	// StartTime: 0/1.
	binary.BigEndian.PutUint32(d[ldtaStartTimeDivd:ldtaStartTimeDivd+4], 0)
	binary.BigEndian.PutUint32(d[ldtaStartTimeDivs:ldtaStartTimeDivs+4], 1)

	// Attr bytes — visible bit at byte 0x27 bit0 (parse_layer.go decoder).
	// Default Visible = true for fresh ShapeLayer.
	d[ldtaAttrByte2] = 0x01

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

	_ = ctx // reserved for capability-driven branches (V3)
	return d
}

// lowerLayerTransform emits the Layer Transform Group LIST(tdgp) using the
// ShapeLayer-canonical 6-axis schema (per iter 2 RE of tolerance.aep —
// scars/v2-2-aelayer-structure.md "iter 2 新 RE 发现"):
//
//   - ADBE Anchor Point         (2-vec, always emit, runtime t.anchorPoint)
//   - ADBE Position_0           (1-d, t.position[0] / X-axis projection)
//   - ADBE Position_1           (1-d, t.position[1] / Y-axis projection)
//   - ADBE Scale                (2-vec, always emit, runtime t.scale)
//   - ADBE Rotate Z             (1-d, runtime t.rotation)
//   - ADBE Opacity              (1-d, runtime t.opacity)
//   - ADBE Orientation          (3-vec via otst wrapper, default [0,0,0])
//   - ADBE Rotate X             (1-d, default 0)
//   - ADBE Rotate Y             (1-d, default 0)
//   - ADBE Envir Appear in Reflect (1-d, default 100)
//
// V2.2 over-emit strategy: even default-valued streams get emitted. AE's
// own elide-default convention is more compact, but AE accepts non-elided
// form. If AE rejects, iter 4 may need selective emit (PropertyStream Mode
// Unset state). Position is split into Position_0/Position_1 because
// tolerance.aep does NOT contain a combined "ADBE Position" on ShapeLayer.
func lowerLayerTransform(t *LayerTransform, ctx *lowerCtx) (*rifx.Chunk, error) {
	tdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdgp.Children = append(tdgp.Children, makeTdsb(), makeTdsn("Transform"))

	posX, posY := splitVec2Stream(t.position)

	// Step 1: Anchor Point (2-vec).
	if c, err := LowerVec2Stream(t.anchorPoint, MatchNameAnchorPoint, "Anchor Point", ctx); err == nil {
		tdgp.Children = append(tdgp.Children, c.Children...)
	} else {
		return nil, err
	}

	// Step 2: Position split into Position_0 (X) + Position_1 (Y).
	if c, err := LowerFloat64Stream(posX, MatchNamePosition0, "X Position", ctx); err == nil {
		tdgp.Children = append(tdgp.Children, c.Children...)
	} else {
		return nil, err
	}
	if c, err := LowerFloat64Stream(posY, MatchNamePosition1, "Y Position", ctx); err == nil {
		tdgp.Children = append(tdgp.Children, c.Children...)
	} else {
		return nil, err
	}

	// Step 3: Scale (2-vec).
	if c, err := LowerVec2Stream(t.scale, MatchNameScale, "Scale", ctx); err == nil {
		tdgp.Children = append(tdgp.Children, c.Children...)
	} else {
		return nil, err
	}

	// Step 4: Rotate Z (1-d).
	if c, err := LowerFloat64Stream(t.rotation, MatchNameRotateZ, "Rotation", ctx); err == nil {
		tdgp.Children = append(tdgp.Children, c.Children...)
	} else {
		return nil, err
	}

	// Step 5: Opacity (1-d).
	if c, err := LowerFloat64Stream(t.opacity, MatchNameOpacity, "Opacity", ctx); err == nil {
		tdgp.Children = append(tdgp.Children, c.Children...)
	} else {
		return nil, err
	}

	// Step 6: Orientation (3-vec, default [0,0,0]) — emitted via otst wrapper.
	tdgp.Children = append(tdgp.Children, makeTdmn(MatchNameOrientation), lowerOrientationDefault())

	// Steps 7-9: Rotate X / Rotate Y / Envir Appear in Reflect — default emits.
	for _, axis := range []struct {
		matchName string
		display   string
		defaultV  float64
	}{
		{MatchNameRotateX, "X Rotation", 0},
		{MatchNameRotateY, "Y Rotation", 0},
		{MatchNameEnvirAppear, "Envir Appear", 100},
	} {
		ps := &PropertyStream[float64]{mode: StreamModeStatic, static: axis.defaultV}
		c, err := LowerFloat64Stream(ps, axis.matchName, axis.display, ctx)
		if err != nil {
			return nil, err
		}
		tdgp.Children = append(tdgp.Children, c.Children...)
	}

	tdgp.Children = append(tdgp.Children, makeTdmn("ADBE Group End"))

	wrapper := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	wrapper.Children = append(wrapper.Children, makeTdmn("ADBE Transform Group"))
	wrapper.Children = append(wrapper.Children, tdgp.Children...)
	return wrapper, nil
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
// Orientation is a 3D quaternion-style stream — AE uses a unique chunk
// structure (otst / otky / otda) distinct from regular tdbs cdat / keyframe
// LIST(list). V2.2 only emits the default form; user-driven orientation
// keyframes are V2.3+.
func lowerOrientationDefault() *rifx.Chunk {
	otst := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOtst}

	innerTdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs}
	innerTdbs.Children = append(innerTdbs.Children,
		makeTdsb(),
		makeTdsn("Orientation"),
		makeTdb4(valueLayout{dim: 3, headerByte: 0x07, spatial: true}),
		&rifx.Chunk{ID: rifx.IDCdat, Data: make([]byte, 24)}, // 3 × f64 = 0,0,0
	)

	otky := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDOtky}
	otky.Children = append(otky.Children, &rifx.Chunk{ID: rifx.IDOtda, Data: make([]byte, 24)})

	otst.Children = append(otst.Children, innerTdbs, otky)
	return otst
}
