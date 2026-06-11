package serializer

import (
	"github.com/example/aep-parser/internal/rifx"
)

// collectShapePaths walks a shape layer's property tree and decodes every
// shap LIST that is NOT inside an "ADBE Mask Parade" subtree. Shape-layer
// custom paths reuse the same om-s/omks/shap wrapper structure as masks;
// the only way to tell them apart is by the enclosing tdmn group name.
func collectShapePaths(layr *rifx.Chunk) []*ShapePath {
	var out []*ShapePath
	var walk func(c *rifx.Chunk, insideMaskParade bool)
	walk = func(c *rifx.Chunk, insideMaskParade bool) {
		kids := c.Children
		for i := 0; i < len(kids); i++ {
			ch := kids[i]
			// Detect entering an "ADBE Mask Parade" subtree by the tdmn +
			// payload pair pattern: when we see that tdmn, the next LIST is
			// the parade's body — recurse into it with insideMaskParade=true.
			if ch.ID == rifx.IDTdmn && trimNUL(ch.Data) == "ADBE Mask Parade" {
				if i+1 < len(kids) && kids[i+1].IsList() {
					walk(kids[i+1], true)
					i++
					continue
				}
			}
			if !ch.IsList() {
				continue
			}
			if !insideMaskParade && ch.FormType == rifx.IDShap {
				if sp := decodeShapePath(ch); sp != nil {
					out = append(out, sp)
				}
			}
			walk(ch, insideMaskParade)
		}
	}
	walk(layr, false)
	return out
}

func decodeShapePath(shap *rifx.Chunk) *ShapePath {
	sp := &ShapePath{}
	for _, ch := range shap.Children {
		switch {
		case ch.ID == rifx.IDShph:
			sp.ShphRaw = append([]byte(nil), ch.Data...)
			if len(ch.Data) >= 4 {
				// shph[3]: 0x01 closed, 0x09 open (bit3 = open). shph[0x14] is a
				// constant 0x01 on every AE-native path — NOT the closed flag.
				// The old [0x14] read reported every open shape path as closed
				// (it only coincided on closed paths). Now matches hydrate's
				// bezierFromShap; ground truth = v2_2_shape_path_re.aep open shaps.
				sp.Closed = ch.Data[3] == 0x01
			}
		case ch.IsList() && ch.FormType == rifx.IDkfl:
			sp.Vertices = decodeMaskVertices(ch)
		case ch.ID == rifx.IDOmtn:
			sp.Name = trimNUL(ch.Data)
		}
	}
	if len(sp.Vertices) == 0 && sp.ShphRaw == nil {
		return nil
	}
	return sp
}

// Match-names used by Shape Layer parametric primitives. Each primitive
// is a tdgp whose tdmn equals one of these "ADBE Vector Shape - …"
// strings; the tdgp's tdmn+tdbs children expose per-primitive params.
const (
	matchVectorShapeRect    = "ADBE Vector Shape - Rect"
	matchVectorShapeEllipse = "ADBE Vector Shape - Ellipse"
	matchVectorShapeStar    = "ADBE Vector Shape - Star"

	matchVectorGroup = "ADBE Vector Group"
)

// shapeFieldSlot maps each known sub-property match-name to the
// ShapePrimitive struct field it populates. Kind-checks happen in
// collectShapePrimitives — slots that don't apply to the current kind
// are skipped (e.g. a stray "Rect Size" inside a Star group).
var shapeFieldSlot = map[string]func(p *ShapePrimitive, prop *Property){
	"ADBE Vector Rect Size":         func(p *ShapePrimitive, v *Property) { p.Size = v },
	"ADBE Vector Rect Position":     func(p *ShapePrimitive, v *Property) { p.Position = v },
	"ADBE Vector Rect Roundness":    func(p *ShapePrimitive, v *Property) { p.Roundness = v },
	"ADBE Vector Ellipse Size":      func(p *ShapePrimitive, v *Property) { p.Size = v },
	"ADBE Vector Ellipse Position":  func(p *ShapePrimitive, v *Property) { p.Position = v },
	"ADBE Vector Star Type":         func(p *ShapePrimitive, v *Property) { p.StarType = v },
	"ADBE Vector Star Points":       func(p *ShapePrimitive, v *Property) { p.Points = v },
	"ADBE Vector Star Position":     func(p *ShapePrimitive, v *Property) { p.Position = v },
	"ADBE Vector Star Rotation":     func(p *ShapePrimitive, v *Property) { p.Rotation = v },
	"ADBE Vector Star Inner Radius": func(p *ShapePrimitive, v *Property) { p.InnerRadius = v },
	"ADBE Vector Star Outer Radius": func(p *ShapePrimitive, v *Property) { p.OuterRadius = v },
	// AE's internal name uses the typo "Roundess" — kept verbatim for matching.
	"ADBE Vector Star Inner Roundess": func(p *ShapePrimitive, v *Property) { p.InnerRoundness = v },
	"ADBE Vector Star Outer Roundess": func(p *ShapePrimitive, v *Property) { p.OuterRoundness = v },
}

// collectShapePrimitives walks the layer's chunk tree and emits one
// ShapePrimitive per parametric primitive found. The traversal records
// the most-recent "ADBE Vector Group" name so the primitive can carry
// its owning group's display name.
func collectShapePrimitives(layr *rifx.Chunk, ctx *parseCtx) []*ShapePrimitive {
	var out []*ShapePrimitive
	var walk func(c *rifx.Chunk, currentGroup string)
	walk = func(c *rifx.Chunk, currentGroup string) {
		kids := c.Children
		for i := 0; i < len(kids); i++ {
			ch := kids[i]
			if ch.ID == rifx.IDTdmn && i+1 < len(kids) && kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdgp {
				name := trimNUL(ch.Data)
				switch name {
				case matchVectorShapeRect:
					if prim := decodeShapePrimitive(ShapePrimitiveRect, currentGroup, kids[i+1], ctx); prim != nil {
						out = append(out, prim)
					}
					i++
					continue
				case matchVectorShapeEllipse:
					if prim := decodeShapePrimitive(ShapePrimitiveEllipse, currentGroup, kids[i+1], ctx); prim != nil {
						out = append(out, prim)
					}
					i++
					continue
				case matchVectorShapeStar:
					if prim := decodeShapePrimitive(ShapePrimitiveStar, currentGroup, kids[i+1], ctx); prim != nil {
						out = append(out, prim)
					}
					i++
					continue
				case matchVectorGroup:
					// Vector Group tdgp — pick up its display name (sibling
					// Utf8 chunk inside) and recurse with the new context.
					subGroup := vectorGroupName(kids[i+1])
					if subGroup == "" {
						subGroup = currentGroup
					}
					walk(kids[i+1], subGroup)
					i++
					continue
				}
			}
			if ch.IsList() {
				walk(ch, currentGroup)
			}
		}
	}
	walk(layr, "")
	return out
}

// vectorGroupName returns the display name for a Vector Group tdgp,
// extracted from its `tdsn` child. AE writes the name as an embedded
// "Utf8"-prefixed sub-record inside tdsn:
//
//	tdsn payload bytes = [ "Utf8" (4) | size uint32 BE (4) | name bytes | optional NUL ]
//
// Returns "" when no tdsn is present or the embedded record is malformed.
func vectorGroupName(tdgp *rifx.Chunk) string {
	for _, ch := range tdgp.Children {
		if ch.ID != rifx.IDTdsn || len(ch.Data) < 8 {
			continue
		}
		if string(ch.Data[0:4]) != "Utf8" {
			continue
		}
		// ch.Data[4:8] is a big-endian uint32 size; we just take the
		// readable bytes after the 8-byte header until NUL or end.
		name := ch.Data[8:]
		if i := indexZero(name); i >= 0 {
			name = name[:i]
		}
		if len(name) > 0 {
			return string(name)
		}
	}
	return ""
}

func indexZero(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return -1
}

// decodeShapePrimitive walks one primitive's tdgp body, parsing each
// tdmn+tdbs leaf into a Property and assigning it to the matching
// ShapePrimitive field via shapeFieldSlot.
func decodeShapePrimitive(kind ShapePrimitiveKind, group string, tdgp *rifx.Chunk, ctx *parseCtx) *ShapePrimitive {
	prim := &ShapePrimitive{Kind: kind, GroupName: group}
	walkTdmnPairs(tdgp, func(name string, payload *rifx.Chunk) bool {
		if payload.FormType == rifx.IDTdbs {
			prop := parseLeafProperty(name, payload, ctx)
			if prop == nil {
				return true
			}
			if slot, ok := shapeFieldSlot[name]; ok {
				slot(prim, prop)
			}
		}
		return true
	})
	return prim
}
