// internal/aep/hydrate_shape.go
//
// V2.2 Phase 4 — chunk tree → runtime ShapeLayer state.
//
// Inverse of lower_layer.go/lowerShapeLayer + lower_shape_node.go +
// lower_property_stream.go. Walks the parsed Layr LIST and populates two
// runtime trees on the Layer:
//
//   - layer.shapeRootGroup — the VectorGroup tree (Rect/Ellipse/Path/Fill/
//     Stroke), via hydrateShapeNodes.
//   - layer.shapeTransform — the Layer-level Transform streams (Anchor /
//     Position / Scale / Rotate Z / Opacity), via hydrateLayerTransform.
//
// Both static and animated streams are handled. The static path uses
// V1 parseLeafProperty's StaticValue (any); the animated path translates
// V1 prop.Keyframes ([]*Keyframe) into typed StreamKeyframe[T] via the
// PropertyStream.AddKeyframeLinear API, which also flips the stream into
// Animated mode. Per Inv-1: a roundtrip preserves runtime semantics.
package aep

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/rifx"
)

// hydrateShapeNodes walks a parsed Layr LIST (descending into nested
// LIST(tdgp) wrappers — V2.2 fix A introduced an outer property-group
// wrapper between Layr and its property tdmn siblings) and returns the
// runtime VectorGroup tree. Returns nil when no "ADBE Root Vectors
// Group" subtree exists; WrapShapeLayer then falls back to a fresh
// empty VectorGroup.
func hydrateShapeNodes(layr *rifx.Chunk, ctx *parseCtx) *VectorGroup {
	var found *VectorGroup
	var visit func(c *rifx.Chunk)
	visit = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		kids := c.Children
		for i := 0; i < len(kids); i++ {
			ch := kids[i]
			if ch.ID == rifx.IDTdmn && trimNUL(ch.Data) == "ADBE Root Vectors Group" && i+1 < len(kids) {
				next := kids[i+1]
				if next.IsList() && next.FormType == rifx.IDTdgp {
					found = hydrateVectorGroup(next, ctx)
					return
				}
			}
			if ch.IsList() {
				visit(ch)
				if found != nil {
					return
				}
			}
		}
	}
	visit(layr)
	return found
}

// hydrateVectorGroup turns a vector-group tdgp into a runtime VectorGroup.
// Children render in serialized order — Children[0] = first emitted =
// bottom of stack per spec §3.2.
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

// hydrateLayerTransform populates layer.shapeTransform from the V1
// layer.Properties list (which parseProperties has already extracted from
// the Layr's Transform Group). Called from parseLayer's LayerTypeShape
// branch.
func hydrateLayerTransform(layer *Layer) {
	if layer.shapeTransform == nil {
		layer.shapeTransform = newLayerTransform()
	}
	// Phase 5 fix B: ShapeLayer canonical Transform splits Position into
	// Position_0 (X) + Position_1 (Y). We re-combine on hydrate so
	// shapeTransform.position remains the [2]float64 API surface.
	var posX, posY *Property
	for _, p := range layer.Properties {
		switch p.MatchName {
		case MatchNameAnchorPoint:
			hydrateVec2Stream(layer.shapeTransform.anchorPoint, p)
		case MatchNamePosition:
			hydrateVec2Stream(layer.shapeTransform.position, p)
		case MatchNamePosition0:
			posX = p
		case MatchNamePosition1:
			posY = p
		case MatchNameScale:
			hydrateVec2Stream(layer.shapeTransform.scale, p)
		case MatchNameRotateZ:
			hydrateFloat64Stream(layer.shapeTransform.rotation, p)
		case MatchNameOpacity:
			hydrateFloat64Stream(layer.shapeTransform.opacity, p)
		}
	}
	if posX != nil || posY != nil {
		combinePositionXY(layer.shapeTransform.position, posX, posY)
	}
}

// combinePositionXY re-combines split Position_0 + Position_1 V1 properties
// into a single PropertyStream[[2]float64]. Static streams merge values
// directly; animated streams pair keyframes by index (AE-canonical split
// emits X and Y keyframes at the same times in the same order).
func combinePositionXY(ps *PropertyStream[[2]float64], px, py *Property) {
	getXY := func(p *Property) (float64, []*Keyframe) {
		if p == nil {
			return 0, nil
		}
		var v float64
		switch x := p.StaticValue.(type) {
		case float64:
			v = x
		case []float64:
			if len(x) > 0 {
				v = x[0]
			}
		}
		return v, p.Keyframes
	}
	xVal, xKfs := getXY(px)
	yVal, yKfs := getXY(py)
	if len(xKfs) == 0 && len(yKfs) == 0 {
		_ = ps.SetStaticValue([2]float64{xVal, yVal})
		return
	}
	n := len(xKfs)
	if len(yKfs) > n {
		n = len(yKfs)
	}
	kfVal := func(kfs []*Keyframe, i int) float64 {
		if i >= len(kfs) {
			return 0
		}
		switch x := kfs[i].Value.(type) {
		case float64:
			return x
		case []float64:
			if len(x) > 0 {
				return x[0]
			}
		}
		return 0
	}
	kfTime := func(kfs []*Keyframe, i int) float64 {
		if i < len(kfs) {
			return kfs[i].Time
		}
		return 0
	}
	for i := 0; i < n; i++ {
		t := kfTime(xKfs, i)
		if i >= len(xKfs) {
			t = kfTime(yKfs, i)
		}
		_ = ps.AddKeyframeLinear(t, [2]float64{kfVal(xKfs, i), kfVal(yKfs, i)})
	}
}

// --- per-node hydrators ---------------------------------------------------

func hydrateRectNode(body *rifx.Chunk, ctx *parseCtx) *RectNode {
	r := NewRectNode()
	props := nodeStreamValues(body, ctx)
	hydrateVec2Stream(r.size, props["ADBE Vector Rect Size"])
	hydrateVec2Stream(r.position, props["ADBE Vector Rect Position"])
	hydrateFloat64Stream(r.roundness, props["ADBE Vector Rect Roundness"])
	return r
}

func hydrateEllipseNode(body *rifx.Chunk, ctx *parseCtx) *EllipseNode {
	e := NewEllipseNode()
	props := nodeStreamValues(body, ctx)
	hydrateVec2Stream(e.size, props["ADBE Vector Ellipse Size"])
	hydrateVec2Stream(e.position, props["ADBE Vector Ellipse Position"])
	return e
}

func hydrateFillNode(body *rifx.Chunk, ctx *parseCtx) *FillNode {
	f := NewFillNode()
	props := nodeStreamValues(body, ctx)
	hydrateColor4Stream(f.color, props["ADBE Vector Fill Color"])
	hydrateFloat64Stream(f.opacity, props["ADBE Vector Fill Opacity"])
	return f
}

func hydrateStrokeNode(body *rifx.Chunk, ctx *parseCtx) *StrokeNode {
	s := NewStrokeNode()
	props := nodeStreamValues(body, ctx)
	hydrateColor4Stream(s.color, props["ADBE Vector Stroke Color"])
	hydrateFloat64Stream(s.opacity, props["ADBE Vector Stroke Opacity"])
	hydrateFloat64Stream(s.width, props["ADBE Vector Stroke Width"])
	return s
}

// hydratePathNode reads the om-s/omks/shap subtree the serializer emits
// for a PathNode (RE-S5b). Recovers vertex count + Closed flag; vertex
// positions are bbox-normalized f32 in the on-disk form so byte-exact
// vertex roundtrip isn't free — V2.2 hydration recovers the structural
// shape (n vertices, closed/open) which is what callers see through
// PathNode.Path().StaticValue().Vertices.
func hydratePathNode(body *rifx.Chunk, _ *parseCtx) *PathNode {
	p := NewPathNode()
	// Find the om-s LIST under the node body's tdmn pairs.
	var oms *rifx.Chunk
	walkTdmnPairs(body, func(name string, payload *rifx.Chunk) bool {
		if name == "ADBE Vector Shape" && payload.IsList() && payload.FormType == rifx.IDOmS {
			oms = payload
			return false
		}
		return true
	})
	if oms == nil {
		return p
	}
	// om-s → LIST(omks) → LIST(shap) → shph + LIST(kfl) → lhd3 + ldat
	var shap *rifx.Chunk
	for _, ch := range oms.Children {
		if ch.IsList() && ch.FormType == rifx.IDOmks {
			for _, sub := range ch.Children {
				if sub.IsList() && sub.FormType == rifx.IDShap {
					shap = sub
					break
				}
			}
		}
	}
	if shap == nil {
		return p
	}
	var shph, lhd3, ldat *rifx.Chunk
	for _, ch := range shap.Children {
		switch {
		case ch.ID == rifx.IDShph:
			shph = ch
		case ch.IsList() && ch.FormType == rifx.IDkfl:
			for _, sub := range ch.Children {
				switch sub.ID {
				case rifx.IDLhd3:
					lhd3 = sub
				case rifx.IDLdat:
					ldat = sub
				}
			}
		}
	}
	closed := true
	if shph != nil && len(shph.Data) >= 4 {
		// Per lower_property_stream.go encodeBezier: flags bytes [2..3] =
		// 0x0201 closed, 0x0200 open.
		closed = shph.Data[3] == 0x01
	}
	verts := decodeBezierVertices(shph, lhd3, ldat)
	bp := BezierPath{
		Vertices:    verts,
		InTangents:  make([][2]float64, len(verts)),
		OutTangents: make([][2]float64, len(verts)),
		Closed:      closed,
	}
	_ = p.path.SetStaticValue(bp)
	return p
}

// decodeBezierVertices reads the bbox-normalized f32 ldat and denormalizes
// using the shph bbox. Returns a slice of len = lhd3 vertex count (zero if
// any chunk is missing or sizes don't add up).
func decodeBezierVertices(shph, lhd3, ldat *rifx.Chunk) [][2]float64 {
	if lhd3 == nil || ldat == nil || len(lhd3.Data) < 0x10 {
		return nil
	}
	n := int(binary.BigEndian.Uint32(lhd3.Data[0x0C:0x10]))
	if n <= 0 || len(ldat.Data) < n*24 {
		return nil
	}
	var minX, minY, maxX, maxY float64
	if shph != nil && len(shph.Data) >= 20 {
		minX = float64(math.Float32frombits(binary.BigEndian.Uint32(shph.Data[4:8])))
		minY = float64(math.Float32frombits(binary.BigEndian.Uint32(shph.Data[8:12])))
		maxX = float64(math.Float32frombits(binary.BigEndian.Uint32(shph.Data[12:16])))
		maxY = float64(math.Float32frombits(binary.BigEndian.Uint32(shph.Data[16:20])))
	}
	rangeX := maxX - minX
	rangeY := maxY - minY
	out := make([][2]float64, n)
	for i := 0; i < n; i++ {
		base := i * 24
		nx := float64(math.Float32frombits(binary.BigEndian.Uint32(ldat.Data[base : base+4])))
		ny := float64(math.Float32frombits(binary.BigEndian.Uint32(ldat.Data[base+4 : base+8])))
		out[i] = [2]float64{minX + nx*rangeX, minY + ny*rangeY}
	}
	return out
}

// --- helpers ---------------------------------------------------------------

// nodeStreamValues walks a per-node body tdgp and returns the map
// matchName → V1 *Property for every tdbs leaf encountered.
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

// hydrateFloat64Stream populates dst from a V1 *Property. Animated when
// p.Keyframes is non-empty; otherwise static (or no-op if p is nil).
func hydrateFloat64Stream(dst *PropertyStream[float64], p *Property) {
	if dst == nil || p == nil {
		return
	}
	if len(p.Keyframes) > 0 {
		for _, kf := range p.Keyframes {
			if v, ok := scalarOf(kf.Value); ok {
				_ = dst.AddKeyframeLinear(kf.Time, v)
			}
		}
		return
	}
	if v, ok := scalarOf(p.StaticValue); ok {
		_ = dst.SetStaticValue(v)
	}
}

// hydrateVec2Stream populates dst from a V1 *Property.
func hydrateVec2Stream(dst *PropertyStream[[2]float64], p *Property) {
	if dst == nil || p == nil {
		return
	}
	if len(p.Keyframes) > 0 {
		for _, kf := range p.Keyframes {
			if v, ok := vec2Of(kf.Value); ok {
				_ = dst.AddKeyframeLinear(kf.Time, v)
			}
		}
		return
	}
	if v, ok := vec2Of(p.StaticValue); ok {
		_ = dst.SetStaticValue(v)
	}
}

// hydrateColor4Stream populates dst from a V1 *Property (RGBA).
func hydrateColor4Stream(dst *PropertyStream[[4]float64], p *Property) {
	if dst == nil || p == nil {
		return
	}
	if len(p.Keyframes) > 0 {
		for _, kf := range p.Keyframes {
			if v, ok := color4Of(kf.Value); ok {
				_ = dst.AddKeyframeLinear(kf.Time, v)
			}
		}
		return
	}
	if v, ok := color4Of(p.StaticValue); ok {
		_ = dst.SetStaticValue(v)
	}
}

// scalarOf extracts 1D float64 from V1 Property.StaticValue / Keyframe.Value.
// decodeCdatValue returns float64 for 1D, []float64 for ND.
func scalarOf(v any) (float64, bool) {
	if f, ok := v.(float64); ok {
		return f, true
	}
	if s, ok := v.([]float64); ok && len(s) >= 1 {
		return s[0], true
	}
	return 0, false
}

// vec2Of extracts [2]float64.
func vec2Of(v any) ([2]float64, bool) {
	s, ok := v.([]float64)
	if !ok || len(s) < 2 {
		return [2]float64{}, false
	}
	return [2]float64{s[0], s[1]}, true
}

// color4Of extracts [4]float64 RGBA.
func color4Of(v any) ([4]float64, bool) {
	s, ok := v.([]float64)
	if !ok || len(s) < 4 {
		return [4]float64{}, false
	}
	return [4]float64{s[0], s[1], s[2], s[3]}, true
}
