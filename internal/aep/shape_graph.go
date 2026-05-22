// internal/aep/shape_graph.go
package aep

import "fmt"

// ShapeNodeKind identifies a shape-graph node's runtime kind. AE match-name
// strings (`ADBE Vector Shape - Rect` etc.) are intentionally NOT exposed
// here — they belong to the serializer (Inv-2). Lowering maps Kind → match
// name via the `shapeMatchNames` table in `lower_shape_node.go` (Phase 2).
type ShapeNodeKind int

const (
	ShapeKindRect    ShapeNodeKind = iota // `ADBE Vector Shape - Rect` (RE-S4)
	ShapeKindEllipse                      // `ADBE Vector Shape - Ellipse` (RE-S5a)
	ShapeKindPath                         // `ADBE Vector Shape - Group` (RE-S5b)
	ShapeKindFill                         // `ADBE Vector Graphic - Fill` (RE-S5c)
	ShapeKindStroke                       // `ADBE Vector Graphic - Stroke` (RE-S5d)
	ShapeKindGroup                        // `ADBE Vector Group` (V2.3+ user-created nested group)
	// V2.3+ candidates: PolyStar / GradientFill / GradientStroke / Trim / Merge /
	// Repeater / Transform.
)

// ShapeNode is the runtime-facing shape-graph node interface. All concrete
// node types (RectNode / EllipseNode / PathNode / FillNode / StrokeNode and
// the V2.3+ container VectorGroup) satisfy it.
type ShapeNode interface {
	Kind() ShapeNodeKind
	Properties() *PropertyGroup // escape hatch β (spec §3.5); Phase 1 returns nil
}

// VectorGroup is the shape-graph container node. Every ShapeLayer carries
// one default RootGroup (constructed by WrapShapeLayer). Children render in
// order: Children[0] = bottom; Children[len-1] = top / most recently
// appended (spec §3.2 render-order convention).
//
// `Transform` is the group-level Transform PropertyGroup placeholder. V2.2
// default-serialized form has no `ADBE Vector Transform Group` (RE-S3); the
// field exists for the escape-hatch / hydration path on user-created nested
// Vector Groups (V2.3+).
type VectorGroup struct {
	Children  []ShapeNode
	Transform *PropertyGroup
}

// NewVectorGroup returns an empty group with an identity Transform placeholder.
func NewVectorGroup() *VectorGroup {
	return &VectorGroup{Transform: newGroupTransform()}
}

// BezierPath is the runtime geometry object — NOT a serializer encoding
// mirror. `Vertices` are the path control points; `InTangents` /
// `OutTangents` are the per-vertex bezier tangent offsets (zero = linear
// segment); `Closed` distinguishes closed shapes from open polylines.
type BezierPath struct {
	Vertices    [][2]float64
	InTangents  [][2]float64
	OutTangents [][2]float64
	Closed      bool
}

// RectNode — `ADBE Vector Shape - Rect`. Default Size=[100,100],
// Position=[0,0], Roundness=0 (RE-S4; AE elides all three at default).
type RectNode struct {
	size      *PropertyStream[[2]float64]
	position  *PropertyStream[[2]float64]
	roundness *PropertyStream[float64]
}

// NewRectNode constructs a default-valued RectNode.
func NewRectNode() *RectNode {
	r := &RectNode{
		size:      NewPropertyStream[[2]float64](),
		position:  NewPropertyStream[[2]float64](),
		roundness: NewPropertyStream[float64](),
	}
	_ = r.size.SetStaticValue([2]float64{100, 100})
	_ = r.position.SetStaticValue([2]float64{0, 0})
	_ = r.roundness.SetStaticValue(0)
	return r
}

func (r *RectNode) Kind() ShapeNodeKind                   { return ShapeKindRect }
func (r *RectNode) Size() *PropertyStream[[2]float64]     { return r.size }
func (r *RectNode) Position() *PropertyStream[[2]float64] { return r.position }
func (r *RectNode) Roundness() *PropertyStream[float64]   { return r.roundness }
func (r *RectNode) SetSize(v [2]float64) error            { return r.size.SetStaticValue(v) }
func (r *RectNode) SetPosition(v [2]float64) error        { return r.position.SetStaticValue(v) }
func (r *RectNode) SetRoundness(v float64) error          { return r.roundness.SetStaticValue(v) }
func (r *RectNode) Properties() *PropertyGroup            { return nil /* Phase 3 escape hatch */ }

// EllipseNode — `ADBE Vector Shape - Ellipse`. Default Size=[100,100],
// Position=[0,0] (RE-S5a). AE child[1] = `ADBE Vector Shape Direction` —
// runtime-default CCW; not exposed as a typed setter in V2.2.
type EllipseNode struct {
	size, position *PropertyStream[[2]float64]
}

// NewEllipseNode constructs a default-valued EllipseNode.
func NewEllipseNode() *EllipseNode {
	e := &EllipseNode{
		size:     NewPropertyStream[[2]float64](),
		position: NewPropertyStream[[2]float64](),
	}
	_ = e.size.SetStaticValue([2]float64{100, 100})
	_ = e.position.SetStaticValue([2]float64{0, 0})
	return e
}

func (e *EllipseNode) Kind() ShapeNodeKind                   { return ShapeKindEllipse }
func (e *EllipseNode) Size() *PropertyStream[[2]float64]     { return e.size }
func (e *EllipseNode) Position() *PropertyStream[[2]float64] { return e.position }
func (e *EllipseNode) SetSize(v [2]float64) error            { return e.size.SetStaticValue(v) }
func (e *EllipseNode) SetPosition(v [2]float64) error        { return e.position.SetStaticValue(v) }
func (e *EllipseNode) Properties() *PropertyGroup            { return nil }

// PathNode — `ADBE Vector Shape - Group`. Default = empty Vertices,
// Closed=true. V2.2 SetVertices builds linear segments (tangents=0); per
// RE-S8, min 2 vertices is enforced runtime-side (conservative invariant —
// AE itself was not directly probed for 0/1-vertex rejection, see §6.6).
type PathNode struct {
	path *PropertyStream[BezierPath]
}

// NewPathNode constructs an empty PathNode (Closed=true).
func NewPathNode() *PathNode {
	p := &PathNode{path: NewPropertyStream[BezierPath]()}
	_ = p.path.SetStaticValue(BezierPath{Closed: true})
	return p
}

func (p *PathNode) Kind() ShapeNodeKind                { return ShapeKindPath }
func (p *PathNode) Path() *PropertyStream[BezierPath]  { return p.path }
func (p *PathNode) Properties() *PropertyGroup         { return nil }

// SetVertices replaces the path's vertex list with linear segments
// (tangents zeroed). Preserves the current `Closed` flag. Requires
// len(verts) >= 2 (RE-S8).
func (p *PathNode) SetVertices(verts [][2]float64) error {
	if len(verts) < 2 {
		return fmt.Errorf("PathNode.SetVertices: need >= 2 vertices, got %d (RE-S8)", len(verts))
	}
	current, _ := p.path.StaticValue()
	zeroTangents := make([][2]float64, len(verts))
	return p.path.SetStaticValue(BezierPath{
		Vertices:    append([][2]float64(nil), verts...),
		InTangents:  zeroTangents,
		OutTangents: zeroTangents,
		Closed:      current.Closed,
	})
}

// SetClosed toggles the `Closed` flag without disturbing vertices/tangents.
func (p *PathNode) SetClosed(closed bool) error {
	current, _ := p.path.StaticValue()
	current.Closed = closed
	return p.path.SetStaticValue(current)
}

// FillNode — `ADBE Vector Graphic - Fill`. Default Color=[1,1,1,1] white,
// Opacity=100 (per spec §3.6 runtime defaults; AE elides at default).
type FillNode struct {
	color   *PropertyStream[[4]float64]
	opacity *PropertyStream[float64]
}

// NewFillNode constructs a default-valued FillNode.
func NewFillNode() *FillNode {
	f := &FillNode{
		color:   NewPropertyStream[[4]float64](),
		opacity: NewPropertyStream[float64](),
	}
	_ = f.color.SetStaticValue([4]float64{1, 1, 1, 1}) // white
	_ = f.opacity.SetStaticValue(100)
	return f
}

func (f *FillNode) Kind() ShapeNodeKind                { return ShapeKindFill }
func (f *FillNode) Color() *PropertyStream[[4]float64] { return f.color }
func (f *FillNode) Opacity() *PropertyStream[float64]  { return f.opacity }
func (f *FillNode) SetColor(v [4]float64) error        { return f.color.SetStaticValue(v) }
func (f *FillNode) SetOpacity(v float64) error         { return f.opacity.SetStaticValue(v) }
func (f *FillNode) Properties() *PropertyGroup         { return nil }

// StrokeNode — `ADBE Vector Graphic - Stroke`. Default Color=[0,0,0,1]
// black, Width=2, Opacity=100 (per spec §3.6).
type StrokeNode struct {
	color   *PropertyStream[[4]float64]
	opacity *PropertyStream[float64]
	width   *PropertyStream[float64]
}

// NewStrokeNode constructs a default-valued StrokeNode.
func NewStrokeNode() *StrokeNode {
	s := &StrokeNode{
		color:   NewPropertyStream[[4]float64](),
		opacity: NewPropertyStream[float64](),
		width:   NewPropertyStream[float64](),
	}
	_ = s.color.SetStaticValue([4]float64{0, 0, 0, 1}) // black
	_ = s.opacity.SetStaticValue(100)
	_ = s.width.SetStaticValue(2)
	return s
}

func (s *StrokeNode) Kind() ShapeNodeKind                { return ShapeKindStroke }
func (s *StrokeNode) Color() *PropertyStream[[4]float64] { return s.color }
func (s *StrokeNode) Opacity() *PropertyStream[float64]  { return s.opacity }
func (s *StrokeNode) Width() *PropertyStream[float64]    { return s.width }
func (s *StrokeNode) SetColor(v [4]float64) error        { return s.color.SetStaticValue(v) }
func (s *StrokeNode) SetOpacity(v float64) error         { return s.opacity.SetStaticValue(v) }
func (s *StrokeNode) SetWidth(v float64) error           { return s.width.SetStaticValue(v) }
func (s *StrokeNode) Properties() *PropertyGroup         { return nil }

// PropertyGroup is the escape-hatch β surface (spec §3.5). Phase 1 ships
// the minimal struct — `Name` and the empty `Children` / `streams` maps —
// so the field exists on VectorGroup.Transform / node Properties() but
// none of the typed lookup methods (`Float64Stream`, `Vec2Stream`, ...)
// are wired yet. Phase 3 fills the escape-hatch methods.
type PropertyGroup struct {
	Name      string
	Children  map[string]*PropertyGroup
	streams   map[string]any
	Separated bool // false in V2.2; V2.3+ exposes separated dimension setters
}

// newGroupTransform constructs the identity Transform PropertyGroup placeholder
// for a fresh VectorGroup. RootGroup default-serialized form has no
// `ADBE Vector Transform Group` (RE-S3) — this placeholder stays nil-children
// until the escape hatch wires it (Phase 3+).
func newGroupTransform() *PropertyGroup {
	return &PropertyGroup{Name: "Transform"}
}
