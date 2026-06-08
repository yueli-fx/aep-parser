// internal/aep/hydrate_shape.go
//
// chunk tree → runtime ShapeLayer state.
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
// PropertyStream.AddKeyframeLinear API, which also flips the stream into Animated
// mode. A roundtrip preserves runtime semantics.
package aep

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
)

// hydrateShapeNodes walks a parsed Layr LIST (descending into nested
// LIST(tdgp) wrappers — an outer property-group wrapper sits between Layr
// and its property tdmn siblings) and returns the runtime VectorGroup tree.
// Returns nil when no "ADBE Root Vectors Group" subtree exists; WrapShapeLayer
// then falls back to a fresh empty VectorGroup.
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

// hydrateVectorGroup turns the Root Vectors Group body into a runtime
// VectorGroup. The on-disk shape is a 5-level nesting:
//
//	Root Vectors Group body → tdmn(Vector Group) + tdgp →
//	  tdmn(Vectors Group) + tdgp → [shape kids]
//	  tdmn(Vector Transform Group) + tdgp(empty)
//	  tdmn(Vector Materials Group) + tdgp(empty)
//
// V2.2 collapses this into a flat shapeRootGroup.Children — the user
// API sees [Rect, Fill, ...] directly. Vector Group / Vectors Group /
// Vector Transform Group / Vector Materials Group are transparent on
// hydrate; lower deterministically reconstructs them.
//
// Children render in serialized order — Children[0] = first emitted =
// bottom of stack.
func hydrateVectorGroup(tdgp *rifx.Chunk, ctx *parseCtx) *VectorGroup {
	g := NewVectorGroup()
	collectShapeKids(tdgp, g, ctx)
	return g
}

// collectShapeKids walks a tdgp body, descending transparently through
// Vector Group / Vectors Group wrappers, and appends typed shape nodes
// to g.Children. Vector Transform Group / Vector Materials Group are
// ignored (V2.2 doesn't expose per-group transforms / materials).
func collectShapeKids(tdgp *rifx.Chunk, g *VectorGroup, ctx *parseCtx) {
	walkTdmnPairs(tdgp, func(matchName string, payload *rifx.Chunk) bool {
		if !payload.IsList() || payload.FormType != rifx.IDTdgp {
			return true
		}
		switch matchName {
		case "ADBE Vector Group", "ADBE Vectors Group":
			collectShapeKids(payload, g, ctx)
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
		case "ADBE Vector Graphic - G-Fill":
			if n := hydrateGradientFillNode(payload, ctx); n != nil {
				g.Children = append(g.Children, n)
			}
		}
		return true
	})
}

// hydrateLayerTransform populates layer.shapeTransform from the V1
// layer.Properties list (which parseProperties has already extracted from
// the Layr's Transform Group). Called from parseLayer's LayerTypeShape
// branch.
func hydrateLayerTransform(layer *Layer) {
	if layer.shapeTransform == nil {
		layer.shapeTransform = newLayerTransform()
	}
	// ShapeLayer canonical Transform splits Position into Position_0 (X) +
	// Position_1 (Y). We re-combine on hydrate so shapeTransform.position
	// remains the [2]float64 API surface.
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
// into a single codec.PropertyStream[[2]float64]. Static streams merge values
// directly; animated streams pair keyframes by index (AE-canonical split
// emits X and Y keyframes at the same times in the same order).
func combinePositionXY(ps *codec.PropertyStream[[2]float64], px, py *Property) {
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
	hydrateScalarStatic(props["ADBE Vector Shape Direction"], func(v float64) { r.direction = ShapeDirection(v) })
	return r
}

func hydrateEllipseNode(body *rifx.Chunk, ctx *parseCtx) *EllipseNode {
	e := NewEllipseNode()
	props := nodeStreamValues(body, ctx)
	hydrateVec2Stream(e.size, props["ADBE Vector Ellipse Size"])
	hydrateVec2Stream(e.position, props["ADBE Vector Ellipse Position"])
	hydrateScalarStatic(props["ADBE Vector Shape Direction"], func(v float64) { e.direction = ShapeDirection(v) })
	return e
}

func hydrateFillNode(body *rifx.Chunk, ctx *parseCtx) *FillNode {
	f := NewFillNode()
	props := nodeStreamValues(body, ctx)
	hydrateColor4Stream(f.color, props["ADBE Vector Fill Color"])
	hydrateFloat64Stream(f.opacity, props["ADBE Vector Fill Opacity"])
	hydrateScalarStatic(props["ADBE Vector Blend Mode"], func(v float64) { f.blendMode = ShapeBlendMode(v) })
	hydrateScalarStatic(props["ADBE Vector Composite Order"], func(v float64) { f.compositeOrder = ShapeCompositeOrder(v) })
	hydrateScalarStatic(props["ADBE Vector Fill Rule"], func(v float64) { f.fillRule = FillRule(v) })
	return f
}

// hydrateGradientFillNode reads the gradient-fill body back into a runtime
// GradientFillNode: descends the Grad Colors GCst→GCky→Utf8 and decodes the
// prop.map XML via codec.ParseGradientXML. Grad Type / Start Pt / End Pt are not
// modeled (elided in the serialized form). Returns a default-gradient node if
// the stops XML is absent (keeps the node visible rather than dropping it).
func hydrateGradientFillNode(body *rifx.Chunk, _ *parseCtx) *GradientFillNode {
	n := NewGradientFillNode()
	if xml := findGradientStopsXML(body, "ADBE Vector Grad Colors"); xml != "" {
		if g := codec.ParseGradientXML(xml); g != nil {
			n.gradient = g
		}
	}
	return n
}

// findGradientStopsXML returns the prop.map XML string in the GCst→GCky→Utf8
// leaf following the tdmn matching streamName, or "".
func findGradientStopsXML(body *rifx.Chunk, streamName string) string {
	kids := body.Children
	for i := 0; i+1 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || trimChunkNUL(kids[i].Data) != streamName {
			continue
		}
		gcst := kids[i+1]
		if !gcst.IsList() || gcst.FormType != rifx.IDGCst {
			return ""
		}
		for _, c := range gcst.Children {
			if c.IsList() && c.FormType == rifx.IDGCky {
				for _, u := range c.Children {
					if u.ID == rifx.IDUtf8 {
						return string(u.Data)
					}
				}
			}
		}
		return ""
	}
	return ""
}

func hydrateStrokeNode(body *rifx.Chunk, ctx *parseCtx) *StrokeNode {
	s := NewStrokeNode()
	props := nodeStreamValues(body, ctx)
	hydrateColor4Stream(s.color, props["ADBE Vector Stroke Color"])
	hydrateFloat64Stream(s.opacity, props["ADBE Vector Stroke Opacity"])
	hydrateFloat64Stream(s.width, props["ADBE Vector Stroke Width"])
	hydrateScalarStatic(props["ADBE Vector Stroke Line Cap"], func(v float64) { s.lineCap = StrokeLineCap(v) })
	hydrateScalarStatic(props["ADBE Vector Stroke Line Join"], func(v float64) { s.lineJoin = StrokeLineJoin(v) })
	hydrateScalarStatic(props["ADBE Vector Stroke Miter Limit"], func(v float64) { s.miterLimit = v })
	hydrateScalarStatic(props["ADBE Vector Blend Mode"], func(v float64) { s.blendMode = ShapeBlendMode(v) })
	hydrateScalarStatic(props["ADBE Vector Composite Order"], func(v float64) { s.compositeOrder = ShapeCompositeOrder(v) })
	hydrateStrokeTaper(body, s.taper, ctx)
	hydrateStrokeWave(body, s.wave, ctx)
	hydrateStrokeDashes(body, s.dashes, ctx)
	return s
}

// hydrateStrokeDashes reads the Dashes group's Dash 1 / Gap 1 sub-streams back
// into the runtime StrokeDashes and flags it enabled. The presence of a Dash 1
// or Gap 1 leaf is the enable signal: a solid stroke serializes the Dashes group
// as an empty placeholder (no Dash/Gap leaves), so enabled stays false. Offset
// is not modeled (no template slot).
func hydrateStrokeDashes(strokeBody *rifx.Chunk, d *StrokeDashes, ctx *parseCtx) {
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Dashes")
	if g == nil || d == nil {
		return
	}
	p := nodeStreamValues(g, ctx)
	dash, gap := p["ADBE Vector Stroke Dash 1"], p["ADBE Vector Stroke Gap 1"]
	if dash == nil && gap == nil {
		return
	}
	d.enabled = true
	hydrateScalarStatic(dash, func(v float64) { d.dash = v })
	hydrateScalarStatic(gap, func(v float64) { d.gap = v })
}

// hydrateStrokeTaper reads the Taper group's %-mode scalar sub-streams back
// into the runtime StrokeTaper. Descends into the nested LIST(tdgp) following
// the "ADBE Vector Stroke Taper" tdmn (walkTdmnPairs stops at Group End and is
// not recursive, so the top-level nodeStreamValues skips it).
func hydrateStrokeTaper(strokeBody *rifx.Chunk, t *StrokeTaper, ctx *parseCtx) {
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Taper")
	if g == nil || t == nil {
		return
	}
	p := nodeStreamValues(g, ctx)
	hydrateScalarStatic(p["ADBE Vector Taper Start Length"], func(v float64) { t.startLength = v })
	hydrateScalarStatic(p["ADBE Vector Taper End Length"], func(v float64) { t.endLength = v })
	hydrateScalarStatic(p["ADBE Vector Taper Start Width"], func(v float64) { t.startWidth = v })
	hydrateScalarStatic(p["ADBE Vector Taper End Width"], func(v float64) { t.endWidth = v })
	hydrateScalarStatic(p["ADBE Vector Taper Start Ease"], func(v float64) { t.startEase = v })
	hydrateScalarStatic(p["ADBE Vector Taper End Ease"], func(v float64) { t.endEase = v })
}

// hydrateStrokeWave reads the Wave group's Wavelength-mode scalars (Amount /
// Wavelength / Phase) back into the runtime StrokeWave.
func hydrateStrokeWave(strokeBody *rifx.Chunk, w *StrokeWave, ctx *parseCtx) {
	g := findGroupBody(strokeBody, "ADBE Vector Stroke Wave")
	if g == nil || w == nil {
		return
	}
	p := nodeStreamValues(g, ctx)
	hydrateScalarStatic(p["ADBE Vector Taper Wave Amount"], func(v float64) { w.amount = v })
	hydrateScalarStatic(p["ADBE Vector Taper Wavelength"], func(v float64) { w.wavelength = v })
	hydrateScalarStatic(p["ADBE Vector Taper Wave Phase"], func(v float64) { w.phase = v })
}

// hydrateScalarStatic applies the static 1D value of p (if present) via set.
// For non-animated enum/scalar properties that are not modeled as streams.
func hydrateScalarStatic(p *Property, set func(float64)) {
	if p == nil {
		return
	}
	if v, ok := scalarOf(p.StaticValue); ok {
		set(v)
	}
}

// hydratePathNode reads the om-s/omks/shap subtree the serializer emits
// for a PathNode. Recovers vertex count + Closed flag; vertex positions are
// bbox-normalized f32 in the on-disk form so byte-exact vertex roundtrip
// isn't free — hydration recovers the structural shape (n vertices,
// closed/open) which is what callers see through PathNode.Path().
//
// A static path has ONE shap inside omks → Path() stays Static. An animated
// path has N shaps (one geometry per keyframe, == animated mask) plus a
// sibling tdbs time table → Path() becomes Animated with N linear keyframes
// (times read from tdbs.kfl via readMaskPathTimes). Temporal ease is V2.3+;
// V2.2 hydrates animated paths as linear. See
// incident-reports/path-keyframe-write-re.md.
func hydratePathNode(body *rifx.Chunk, ctx *parseCtx) *PathNode {
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
	// om-s → { tdbs (time table), LIST(omks) → N × LIST(shap) }
	var tdbs *rifx.Chunk
	var shaps []*rifx.Chunk
	for _, ch := range oms.Children {
		if !ch.IsList() {
			continue
		}
		switch ch.FormType {
		case rifx.IDTdbs:
			tdbs = ch
		case rifx.IDOmks:
			for _, sub := range ch.Children {
				if sub.IsList() && sub.FormType == rifx.IDShap {
					shaps = append(shaps, sub)
				}
			}
		}
	}
	if len(shaps) == 0 {
		return p
	}
	if len(shaps) == 1 {
		// Static path — single snapshot, stream stays Static.
		_ = p.path.SetStaticValue(bezierFromShap(shaps[0]))
		return p
	}
	// Animated: pair each shap's geometry with its tdbs time entry.
	times := readMaskPathTimes(tdbs, ctx)
	for i, s := range shaps {
		t := 0.0
		if i < len(times) {
			t = times[i].time
		}
		_ = p.path.AddKeyframeLinear(t, bezierFromShap(s))
	}
	return p
}

// bezierFromShap decodes one shap LIST (shph + kfl{lhd3,ldat}) into a
// BezierPath with denormalized vertices. Tangents are zeroed (V2.2 linear
// scope; on-disk tangent fidelity is deferred — see
// incident-reports/path-keyframe-write-re.md).
func bezierFromShap(shap *rifx.Chunk) BezierPath {
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
	return BezierPath{
		Vertices:    verts,
		InTangents:  make([][2]float64, len(verts)),
		OutTangents: make([][2]float64, len(verts)),
		Closed:      closed,
	}
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
func hydrateFloat64Stream(dst *codec.PropertyStream[float64], p *Property) {
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
func hydrateVec2Stream(dst *codec.PropertyStream[[2]float64], p *Property) {
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
func hydrateColor4Stream(dst *codec.PropertyStream[[4]float64], p *Property) {
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
