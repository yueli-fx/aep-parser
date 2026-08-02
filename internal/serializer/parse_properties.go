package serializer

import (
	"encoding/binary"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// parseProperties extracts the property tree from a Layr. The property tree
// is a hierarchy of tdgp LISTs; each named property is either:
//   - a tdbs LIST (scalar/vector property with optional keyframes), or
//   - a nested tdgp LIST (group like "Transform", terminated by an
//     "ADBE Group End" sentinel match-name).
//
// Only properties with content we know how to read (a cdat static value or
// an lhd3+ldat keyframe stream) are returned. Groups themselves are not
// returned as properties — they're traversed transparently.
//
// Effects ("ADBE Effect Parade" group) are diverted out of props into a
// separate Effects slice, where each effect is one tdmn+sspc-wrapper pair
// holding its own parameter tdbs leaves. "ADBE Marker" (mrst wrapper) is
// diverted into Markers in the same way.
func parseProperties(layr *rifx.Chunk, ctx *parseCtx) ([]*Property, []*Effect, []*Marker) {
	var props []*Property
	var effects []*Effect
	var markers []*Marker
	for _, ch := range layr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			collectFromGroup(ch, &props, &effects, &markers, ctx)
		}
	}
	return props, effects, markers
}

// collectFromGroup walks one tdgp LIST. Each property is announced by a
// tdmn chunk followed by either a tdbs (leaf) or a nested tdgp (subgroup).
// "ADBE Group End" tdmn terminates the group at this level.
//
// Special-cases:
//   - "ADBE Effect Parade" subgroup → divert to collectEffects
//   - "ADBE Marker" + mrst wrapper → divert to parseMarkers
func collectFromGroup(group *rifx.Chunk, out *[]*Property, effects *[]*Effect, markers *[]*Marker, ctx *parseCtx) {
	walkTdmnPairs(group, func(name string, payload *rifx.Chunk) bool {
		switch payload.FormType {
		case rifx.IDTdbs:
			if p := parseLeafProperty(name, payload, ctx); p != nil {
				*out = append(*out, p)
			}
		case rifx.IDTdgp:
			if name == "ADBE Effect Parade" {
				collectEffects(payload, effects, ctx)
			} else {
				collectFromGroup(payload, out, effects, markers, ctx)
			}
		case rifx.IDMrst:
			// Markers — divert; don't append a bogus Property.
			if m := parseMarkers(payload, ctx); len(m) > 0 {
				*markers = append(*markers, m...)
			}
		case rifx.IDGCst:
			// Gradient color stops wrapper — inner tdbs holds base metadata
			// and GCky carries Utf8 chunks with the prop.map XML (one per
			// keyframe; first one used as the static value).
			if p := parseGradientStopsProperty(name, payload, ctx); p != nil {
				*out = append(*out, p)
			}
		case rifx.IDOtst:
			// 3D Orientation wrapper — cdat is little-endian and 3-component.
			if p := parseOrientationProperty(name, payload, ctx); p != nil {
				*out = append(*out, p)
			}
		default:
			// otst (orientation), parT (effect param), etc. — descend so we
			// catch anything tdbs-shaped inside.
			descend(payload, name, out, ctx)
		}
		return true
	})
}

// parseGradientStopsProperty reads an "ADBE Vector Grad Colors" GCst LIST.
// Structure (per the reference parser's parsers/specialized_properties.py::parse_gradient):
//
//	[LIST GCst]
//	  [LIST tdbs]  — base property metadata (tdb4 + small placeholder cdat)
//	  [LIST GCky]  — gradient keyframe container
//	    Utf8     — prop.map XML (one per keyframe; first = static value)
//
// We expose the first decoded codec.Gradient as Property.codec.Gradient (the static or
// first-keyframe value). Per-keyframe gradients are deferred until a
// fixture demonstrates animated gradients.
func parseGradientStopsProperty(matchName string, gcst *rifx.Chunk, ctx *parseCtx) *Property {
	innerTdbs := gcst.FindFirstList(rifx.IDTdbs)
	if innerTdbs == nil {
		return nil
	}
	prop := parseLeafProperty(matchName, innerTdbs, ctx)
	if prop == nil {
		// Even when the inner tdbs has no decodable cdat (the placeholder
		// in real fixtures is only 4 bytes), we still want to surface the
		// gradient. Fabricate a minimal Property carrying the XML.
		prop = &Property{MatchName: matchName, Name: matchName, NameSource: "fallback", Components: 1}
		scene.SetPropertyBack(prop, &propertyBackrefs{tdbs: innerTdbs, decodeStatus: "partially-decoded"})
	}
	gcky := gcst.FindFirstList(rifx.IDGCky)
	if gcky == nil {
		return prop
	}
	for _, ch := range gcky.Children {
		if ch.IsList() || ch.ID != rifx.IDUtf8 {
			continue
		}
		if g := codec.ParseGradientXML(string(ch.Data)); g != nil {
			prop.Gradient = g
			break // first one wins (static / first-keyframe value)
		}
	}
	return prop
}

// parseOrientationProperty reads the otst wrapper that holds a 3D layer's
// Orientation. The wrapper nests a tdbs whose cdat stores the X/Y/Z value
// LITTLE-ENDIAN (the reference parser: cdat is_le when its grandparent LIST is otst) plus
// an otky keyframe container. Orientation is always 3-component even though
// its tdb4 dimension byte reads 0x01, so Components/StaticValue are fixed up
// after the shared parseLeafProperty pass.
//
// Keyframe VALUES come from the otky/otda chunks (each otda = one keyframe's
// X/Y/Z, big-endian), NOT from the kfl ldat (whose value slot is zero for
// orientation). Keyframe timing is taken from the shared parseLeafProperty
// pass; easing/tangents on animated orientation are not yet validated
// (see incidents/transform-group-default-omission.md § otst fidelity).
func parseOrientationProperty(matchName string, otst *rifx.Chunk, ctx *parseCtx) *Property {
	tdbs := otst.FindFirstList(rifx.IDTdbs)
	if tdbs == nil {
		return nil
	}
	prop := parseLeafProperty(matchName, tdbs, ctx)
	if prop == nil {
		prop = &Property{MatchName: matchName, Name: matchName, NameSource: "fallback"}
		scene.SetPropertyBack(prop, &propertyBackrefs{tdbs: tdbs, decodeStatus: "partially-decoded"})
	}
	prop.Components = 3

	if len(prop.Keyframes) > 0 {
		// Animated: replace the (zero) ldat values with the real per-keyframe
		// X/Y/Z from otda, in order.
		if pb := propertyBack(prop); pb != nil {
			pb.keyframeValuesExternal = true
		}
		decodedValues := 0
		if otky := otst.FindFirstList(rifx.IDOtky); otky != nil {
			for _, ch := range otky.Children {
				if ch.ID != rifx.IDOtda || len(ch.Data) < 24 {
					continue
				}
				if decodedValues >= len(prop.Keyframes) {
					break
				}
				prop.Keyframes[decodedValues].Value = decodeCdatValue(ch.Data, 3) // otda is big-endian
				decodedValues++
			}
		}
		if decodedValues < len(prop.Keyframes) {
			setPropertyDecodeEvidence(prop, "partially-decoded", "")
		}
	} else if cdat := tdbs.FindFirst(rifx.IDCdat); cdat != nil {
		// Static: the cdat value is little-endian inside an otst.
		if pb := propertyBack(prop); pb != nil {
			pb.cdat = cdat
			pb.cdatLE = true // SetStaticValue must write LE to match (orientation quirk)
			// AE reads the static orientation from otda (BE), not cdat — capture
			// it so SetStaticValue can mirror the value there too.
			if otky := otst.FindFirstList(rifx.IDOtky); otky != nil {
				if otda := otky.FindFirst(rifx.IDOtda); otda != nil && len(otda.Data) >= 24 {
					pb.otda = otda
				}
			}
		}
		if len(cdat.Data) >= 8 {
			prop.StaticValue = decodeCdatValueLE(cdat.Data, 3)
		}
		if len(cdat.Data) < 24 {
			setPropertyDecodeEvidence(prop, "partially-decoded", "")
		}
	}
	return prop
}

// decodeCdatValueLE reads `components` little-endian float64 from a cdat
// chunk. Always returns []float64 (orientation is multi-component); the
// big-endian counterpart is decodeCdatValue.
func decodeCdatValueLE(d []byte, components int) any {
	vals := make([]float64, 0, components)
	for i := 0; i < components; i++ {
		v, ok := readFloat64LE(d, i*8)
		if !ok {
			break
		}
		vals = append(vals, v)
	}
	return vals
}

// collectEffects walks the "ADBE Effect Parade" tdgp. Each effect is a
// tdmn (effect match-name like "ADBE Gaussian Blur 2") followed by a
// wrapper LIST whose formType is the effect's type tag (sspc in observed
// files). The wrapper's children are tdmn + tdbs (or tdgp) pairs for each
// parameter.
func collectEffects(parade *rifx.Chunk, effects *[]*Effect, ctx *parseCtx) {
	walkTdmnPairs(parade, func(name string, wrapper *rifx.Chunk) bool {
		effect := &Effect{MatchName: name, Name: name}
		// The wrapper LIST is usually "sspc" (or similar) holding a nested
		// tdgp that owns the parameter tdmn+tdbs pairs. Some effects nest
		// the parameters directly as a tdgp; handle both.
		var unusedMarkers []*Marker
		if wrapper.FormType == rifx.IDTdgp {
			collectFromGroup(wrapper, &effect.Parameters, effects, &unusedMarkers, ctx)
		} else {
			for _, w := range wrapper.Children {
				if w.IsList() && w.FormType == rifx.IDTdgp {
					collectFromGroup(w, &effect.Parameters, effects, &unusedMarkers, ctx)
				}
			}
		}
		// Extract pard metadata and apply to parameters.
		if pardDefs := parsePardParams(wrapper); pardDefs != nil {
			applyPardDefs(effect.Parameters, pardDefs)
		}
		*effects = append(*effects, effect)
		return true
	})
}

// applyPardDefs applies pard parameter definition metadata to the
// parsed effect parameters. Matches by match-name.
func applyPardDefs(params []*Property, defs map[string]*pardParamDef) {
	for _, p := range params {
		def, ok := defs[p.MatchName]
		if !ok {
			continue
		}
		p.LastValue = def.lastValue
		p.NbOptions = def.nbOptions
		if def.defaultVal != nil {
			p.DefaultValue = def.defaultVal
		}
		// Override control type when pard provides a more precise value.
		if def.controlType != PCTLUnknown {
			p.DeclaredControlType = def.controlType
			p.HasDeclaredControlType = true
		}
	}
}

// descend walks an unknown wrapper LIST (otst, parT, ...) looking for tdbs
// children so we still surface their values.
func descend(c *rifx.Chunk, parentName string, out *[]*Property, ctx *parseCtx) {
	for _, ch := range c.Children {
		if !ch.IsList() {
			continue
		}
		if ch.FormType == rifx.IDTdbs {
			if p := parseLeafProperty(parentName, ch, ctx); p != nil {
				*out = append(*out, p)
			}
		} else {
			descend(ch, parentName, out, ctx)
		}
	}
}

// parseLeafProperty reads a tdbs LIST — the property block for a single
// named property. Returns nil if neither a cdat nor a keyframe list is
// found.
//
// An optional sibling Utf8 chunk inside the tdbs holds the property's
// expression source (JavaScript). If present, it's surfaced as
// Property.Expression.
func parseLeafProperty(matchName string, tdbs *rifx.Chunk, ctx *parseCtx) *Property {
	prop := &Property{MatchName: matchName, Name: matchName, NameSource: "fallback", Components: 1}
	scene.SetPropertyBack(prop, &propertyBackrefs{decodeStatus: "decoded"})
	pb := propertyBack(prop)

	if tdb4 := tdbs.FindFirst(rifx.IDtdb4); tdb4 != nil {
		prop.Components = decodeTdb4Components(tdb4.Data)
		pb.tdb4 = tdb4
	}
	// Some shape primitive paths use uppercase IDTdb4 — record either.
	if pb.tdb4 == nil {
		if tdb4 := tdbs.FindFirst(rifx.IDTdb4); tdb4 != nil {
			pb.tdb4 = tdb4
		}
	}

	// Parse tdsb subprop flags chunk (4 bytes) if present.
	if tdsb := tdbs.FindFirst(rifx.IDTdsb); tdsb != nil {
		pb.tdsb = tdsb
	}

	// Parse tdum/tduM min/max value chunks if present.
	if tdum := tdbs.FindFirst(rifx.IDtdum); tdum != nil {
		pb.tdum = tdum
	}
	if tduM := tdbs.FindFirst(rifx.IDtduM); tduM != nil {
		pb.tduM = tduM
	}
	for _, ch := range tdbs.Children {
		if ch.ID == rifx.IDTdpi && len(ch.Data) >= 4 {
			prop.LayerRefID = binary.BigEndian.Uint32(ch.Data[:4])
			prop.LayerRefPresent = true
			break
		}
	}

	cdat := tdbs.FindFirst(rifx.IDCdat)
	var kfList *rifx.Chunk
	for _, ch := range tdbs.Children {
		if ch.IsList() && ch.FormType == rifx.IDkfl {
			kfList = ch
			break
		}
	}

	// Expression: a Utf8 chunk inside the tdbs (sibling to tdb4/cdat).
	pb.tdbs = tdbs
	if utf8 := tdbs.FindFirst(rifx.IDUtf8); utf8 != nil {
		prop.Expression = utf8.Text()
		pb.exprChunk = utf8
	}

	// ExpressionEnabled: tdb4 @0x77 is the disabled byte (0 = AE evaluates,
	// 1 = expression kept but off); @0x78 is the has-expression marker, NOT
	// the enabled flag (RE'd against an AE-2025-native enabled/disabled
	// fixture pair, expr_re 2026-06-12: enabled = 00 01, disabled = 01 01
	// at @0x77/@0x78 — the historic reading of @0x78 as an inverted enabled
	// byte conflated the two and made every SetExpression render-dead).
	// Without an expression we report AE's scripting default true.
	tdb4 := tdbs.FindFirst(rifx.IDtdb4)
	if tdb4 == nil {
		tdb4 = tdbs.FindFirst(rifx.IDTdb4)
	}
	if prop.Expression != "" && tdb4 != nil && len(tdb4.Data) > 0x77 {
		prop.ExpressionEnabled = tdb4.Data[0x77] == 0
	} else {
		prop.ExpressionEnabled = true
	}

	if kfList != nil {
		lhd3 := kfList.FindFirst(rifx.IDLhd3)
		ldat := kfList.FindFirst(rifx.IDLdat)
		if lhd3 != nil && ldat != nil {
			parseKeyframes(prop, lhd3, ldat, ctx)
		} else {
			setPropertyDecodeEvidence(prop, "partially-decoded", "invalid-preserved")
		}
	} else if cdat != nil && len(cdat.Data) >= 8 {
		pb.cdat = cdat
		prop.StaticValue = decodeCdatValue(cdat.Data, prop.Components)
		if prop.Components <= 0 || len(cdat.Data) < prop.Components*8 {
			setPropertyDecodeEvidence(prop, "partially-decoded", "")
		}
	} else if prop.Expression == "" {
		// Nothing useful in this tdbs.
		return nil
	} else {
		setPropertyDecodeEvidence(prop, "partially-decoded", "")
	}
	return prop
}

// decodeTdb4Components peeks tdb4 to decide how many float64 components the
// property has. Byte 0x03 encodes the dimension:
//
//	0x01 → 1D (sliders, opacity, single-axis rotation)
//	0x02 → 2D non-spatial (Mask Feather X/Y, 2D point sliders)
//	0x03 → 3D non-spatial (Scale, Orientation when stored as 3 doubles)
//	0x07 → 3D spatial (Position, Anchor Point — motion path enabled)
//	0x04 → 4D (color properties like Tritone Highlights/Midtones/Shadows
//	       — encoded as RGBA / 4-channel in 0..255 byte range stored as
//	       doubles for 8bpc projects)
//
// AE stores all multi-component transform values as 3D even when the layer
// is 2D — Z is just always zero.
func decodeTdb4Components(d []byte) int {
	if len(d) < 4 {
		return 1
	}
	switch d[3] {
	case 0x01:
		return 1
	case 0x02:
		return 2
	case 0x03, 0x07:
		return 3
	case 0x04:
		return 4
	default:
		return 1
	}
}

// decodeCdatValue reads 1..n float64 BE values from a cdat chunk. The cdat
// is padded to a fixed footprint (40 bytes for 1D, 72 for 3D, etc.) but the
// actual values live at the very start.
func decodeCdatValue(d []byte, components int) any {
	if components <= 1 {
		v, ok := readFloat64BE(d, 0)
		if !ok {
			return nil
		}
		return v
	}
	vals := make([]float64, 0, components)
	for i := 0; i < components; i++ {
		v, ok := readFloat64BE(d, i*8)
		if !ok {
			break
		}
		vals = append(vals, v)
	}
	return vals
}

// decodeTdumValue reads a tdum/tduM chunk's payload. Layout depends on
// tdb4 type flags: color → 4×float32 BE, integer → 1×uint32 BE,
// otherwise N×float64 BE (N = size/8).
