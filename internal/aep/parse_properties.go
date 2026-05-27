package aep

import (
	"github.com/example/aep-parser/internal/rifx"
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
		default:
			// otst (orientation), parT (effect param), etc. — descend so we
			// catch anything tdbs-shaped inside.
			descend(payload, name, out, ctx)
		}
		return true
	})
}

// parseGradientStopsProperty reads an "ADBE Vector Grad Colors" GCst LIST.
// Structure (per py-aep parsers/specialized_properties.py::parse_gradient):
//
//	[LIST GCst]
//	  [LIST tdbs]  — base property metadata (tdb4 + small placeholder cdat)
//	  [LIST GCky]  — gradient keyframe container
//	    Utf8     — prop.map XML (one per keyframe; first = static value)
//
// We expose the first decoded Gradient as Property.Gradient (the static or
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
		prop = &Property{MatchName: matchName, Name: matchName, Components: 1, tdbs: innerTdbs}
	}
	gcky := gcst.FindFirstList(rifx.IDGCky)
	if gcky == nil {
		return prop
	}
	for _, ch := range gcky.Children {
		if ch.IsList() || ch.ID != rifx.IDUtf8 {
			continue
		}
		if g := ParseGradientXML(string(ch.Data)); g != nil {
			prop.Gradient = g
			break // first one wins (static / first-keyframe value)
		}
	}
	return prop
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
			// Store on property for ControlType() to use.
			// We reuse the tdb4-derived value as fallback; pard is authoritative.
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
	prop := &Property{MatchName: matchName, Name: matchName, Components: 1}

	if tdb4 := tdbs.FindFirst(rifx.IDtdb4); tdb4 != nil {
		prop.Components = decodeTdb4Components(tdb4.Data)
		prop.tdb4 = tdb4
	}
	// Some shape primitive paths use uppercase IDTdb4 — record either.
	if prop.tdb4 == nil {
		if tdb4 := tdbs.FindFirst(rifx.IDTdb4); tdb4 != nil {
			prop.tdb4 = tdb4
		}
	}

	// Parse tdsb subprop flags chunk (4 bytes) if present.
	if tdsb := tdbs.FindFirst(rifx.IDTdsb); tdsb != nil {
		prop.tdsb = tdsb
	}

	// Parse tdum/tduM min/max value chunks if present.
	if tdum := tdbs.FindFirst(rifx.IDtdum); tdum != nil {
		prop.tdum = tdum
	}
	if tduM := tdbs.FindFirst(rifx.IDtduM); tduM != nil {
		prop.tduM = tduM
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
	prop.tdbs = tdbs
	if utf8 := tdbs.FindFirst(rifx.IDUtf8); utf8 != nil {
		prop.Expression = utf8.Text()
		prop.exprChunk = utf8
	}

	// ExpressionEnabled: tdb4 @0x78 is an INVERTED "disabled" byte —
	// 0 = enabled (AE evaluates), 1 = disabled. We expose the
	// non-inverted form. Default true when the byte is absent so properties
	// without an expression don't read as "disabled".
	prop.ExpressionEnabled = true
	tdb4 := tdbs.FindFirst(rifx.IDtdb4)
	if tdb4 == nil {
		tdb4 = tdbs.FindFirst(rifx.IDTdb4)
	}
	if tdb4 != nil && len(tdb4.Data) > 0x78 {
		prop.ExpressionEnabled = tdb4.Data[0x78] == 0
	}

	if kfList != nil {
		lhd3 := kfList.FindFirst(rifx.IDLhd3)
		ldat := kfList.FindFirst(rifx.IDLdat)
		if lhd3 != nil && ldat != nil {
			parseKeyframes(prop, lhd3, ldat, ctx)
		}
	} else if cdat != nil && len(cdat.Data) >= 8 {
		prop.cdat = cdat
		prop.StaticValue = decodeCdatValue(cdat.Data, prop.Components)
	} else if prop.Expression == "" {
		// Nothing useful in this tdbs.
		return nil
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
		v, _ := readFloat64BE(d, 0)
		return v
	}
	vals := make([]float64, components)
	for i := 0; i < components; i++ {
		v, ok := readFloat64BE(d, i*8)
		if !ok {
			break
		}
		vals[i] = v
	}
	return vals
}
