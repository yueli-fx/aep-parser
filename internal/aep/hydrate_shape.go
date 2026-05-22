// internal/aep/hydrate_shape.go
//
// V2.2 Phase 4 Task 4.1 — chunk tree → runtime VectorGroup tree.
//
// Inverse of lower_layer.go/lowerShapeLayer + lower_shape_node.go. Walks
// the parsed Layr LIST, finds the "ADBE Root Vectors Group" subtree, and
// rebuilds the typed runtime ShapeNode graph. Reuses V1 parseLeafProperty
// (parse_properties.go) so per-property cdat / tdb4 decoding stays in one
// place (Inv-1: serializer artifacts ↔ runtime values are bijective).
//
// V2.2 scope: static values only. Keyframe hydration (animated streams)
// arrives with Task 4.2's canonical roundtrip — V1's prop.Keyframes is
// the source, translation lives there.
package aep

import (
	"github.com/example/aep-parser/internal/rifx"
)

// hydrateShapeNodes walks a parsed Layr LIST and returns the runtime
// VectorGroup tree (root of the shape graph). Returns nil when the Layr
// has no "ADBE Root Vectors Group" subtree — caller (WrapShapeLayer)
// falls back to a fresh empty VectorGroup so user-built ShapeLayers and
// parsed-but-empty ones share the same wrapper code path.
func hydrateShapeNodes(layr *rifx.Chunk, ctx *parseCtx) *VectorGroup {
	for i := 0; i+1 < len(layr.Children); i++ {
		ch := layr.Children[i]
		if ch.ID != rifx.IDTdmn || trimNUL(ch.Data) != "ADBE Root Vectors Group" {
			continue
		}
		next := layr.Children[i+1]
		if !next.IsList() || next.FormType != rifx.IDTdgp {
			continue
		}
		return hydrateVectorGroup(next, ctx)
	}
	return nil
}

// hydrateVectorGroup turns a vector-group tdgp (the body of "ADBE Root
// Vectors Group" or a nested "ADBE Vector Group") into a runtime
// VectorGroup. Children render in serialized order — Children[0] = first
// emitted = bottom of stack per spec §3.2.
func hydrateVectorGroup(tdgp *rifx.Chunk, ctx *parseCtx) *VectorGroup {
	g := NewVectorGroup()
	walkTdmnPairs(tdgp, func(matchName string, payload *rifx.Chunk) bool {
		if !payload.IsList() || payload.FormType != rifx.IDTdgp {
			return true
		}
		switch matchName {
		case "ADBE Vector Shape - Rect":
			if n := hydrateRectNode(payload, ctx); n != nil {
				g.Children = append(g.Children, n)
			}
		case "ADBE Vector Shape - Ellipse":
			if n := hydrateEllipseNode(payload, ctx); n != nil {
				g.Children = append(g.Children, n)
			}
		case "ADBE Vector Shape - Group":
			if n := hydratePathNode(payload, ctx); n != nil {
				g.Children = append(g.Children, n)
			}
		case "ADBE Vector Graphic - Fill":
			if n := hydrateFillNode(payload, ctx); n != nil {
				g.Children = append(g.Children, n)
			}
		case "ADBE Vector Graphic - Stroke":
			if n := hydrateStrokeNode(payload, ctx); n != nil {
				g.Children = append(g.Children, n)
			}
		}
		return true
	})
	return g
}

// nodeStreamValues walks a per-node body tdgp and returns the map
// matchName → V1 *Property for every tdbs leaf encountered. Sub-properties
// the V2.2 serializer emitted as empty placeholders (Direction / Blend Mode
// / etc.) yield nil from parseLeafProperty and are dropped silently.
func nodeStreamValues(body *rifx.Chunk, ctx *parseCtx) map[string]*Property {
	out := map[string]*Property{}
	walkTdmnPairs(body, func(name string, payload *rifx.Chunk) bool {
		if payload.IsList() && payload.FormType == rifx.IDTdbs {
			if p := parseLeafProperty(name, payload, ctx); p != nil {
				out[name] = p
			}
		}
		return true
	})
	return out
}

func hydrateRectNode(body *rifx.Chunk, ctx *parseCtx) *RectNode {
	r := NewRectNode()
	props := nodeStreamValues(body, ctx)
	if p, ok := props["ADBE Vector Rect Size"]; ok {
		if v, ok := vec2Of(p.StaticValue); ok {
			_ = r.SetSize(v)
		}
	}
	if p, ok := props["ADBE Vector Rect Position"]; ok {
		if v, ok := vec2Of(p.StaticValue); ok {
			_ = r.SetPosition(v)
		}
	}
	if p, ok := props["ADBE Vector Rect Roundness"]; ok {
		if v, ok := scalarOf(p.StaticValue); ok {
			_ = r.SetRoundness(v)
		}
	}
	return r
}

func hydrateEllipseNode(body *rifx.Chunk, ctx *parseCtx) *EllipseNode {
	e := NewEllipseNode()
	props := nodeStreamValues(body, ctx)
	if p, ok := props["ADBE Vector Ellipse Size"]; ok {
		if v, ok := vec2Of(p.StaticValue); ok {
			_ = e.SetSize(v)
		}
	}
	if p, ok := props["ADBE Vector Ellipse Position"]; ok {
		if v, ok := vec2Of(p.StaticValue); ok {
			_ = e.SetPosition(v)
		}
	}
	return e
}

// hydratePathNode reads the om-s / omks / shap subtree the serializer
// emits for a PathNode (RE-S5b). V2.2 hydration is a placeholder — full
// bbox-normalized BezierPath decoding lands when Task 4.2 demands it.
func hydratePathNode(body *rifx.Chunk, _ *parseCtx) *PathNode {
	return NewPathNode()
}

func hydrateFillNode(body *rifx.Chunk, ctx *parseCtx) *FillNode {
	f := NewFillNode()
	props := nodeStreamValues(body, ctx)
	if p, ok := props["ADBE Vector Fill Color"]; ok {
		if v, ok := color4Of(p.StaticValue); ok {
			_ = f.SetColor(v)
		}
	}
	if p, ok := props["ADBE Vector Fill Opacity"]; ok {
		if v, ok := scalarOf(p.StaticValue); ok {
			_ = f.SetOpacity(v)
		}
	}
	return f
}

func hydrateStrokeNode(body *rifx.Chunk, ctx *parseCtx) *StrokeNode {
	s := NewStrokeNode()
	props := nodeStreamValues(body, ctx)
	if p, ok := props["ADBE Vector Stroke Color"]; ok {
		if v, ok := color4Of(p.StaticValue); ok {
			_ = s.SetColor(v)
		}
	}
	if p, ok := props["ADBE Vector Stroke Opacity"]; ok {
		if v, ok := scalarOf(p.StaticValue); ok {
			_ = s.SetOpacity(v)
		}
	}
	if p, ok := props["ADBE Vector Stroke Width"]; ok {
		if v, ok := scalarOf(p.StaticValue); ok {
			_ = s.SetWidth(v)
		}
	}
	return s
}

// scalarOf extracts a 1D float64 from a V1 Property.StaticValue. The
// 1-component code path in decodeCdatValue returns a bare float64 (not a
// []float64).
func scalarOf(v any) (float64, bool) {
	if f, ok := v.(float64); ok {
		return f, true
	}
	if s, ok := v.([]float64); ok && len(s) >= 1 {
		return s[0], true
	}
	return 0, false
}

// vec2Of extracts [2]float64 from a V1 Property.StaticValue ([]float64
// len>=2).
func vec2Of(v any) ([2]float64, bool) {
	s, ok := v.([]float64)
	if !ok || len(s) < 2 {
		return [2]float64{}, false
	}
	return [2]float64{s[0], s[1]}, true
}

// color4Of extracts [4]float64 RGBA from a V1 Property.StaticValue.
func color4Of(v any) ([4]float64, bool) {
	s, ok := v.([]float64)
	if !ok || len(s) < 4 {
		return [4]float64{}, false
	}
	return [4]float64{s[0], s[1], s[2], s[3]}, true
}
