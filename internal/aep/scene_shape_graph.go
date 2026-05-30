// internal/aep/shape_graph.go
package aep

import "fmt"

// ShapeNodeKind identifies a shape-graph node's runtime kind. AE match-name
// strings (`ADBE Vector Shape - Rect` etc.) are intentionally NOT exposed
// here — they belong to the serializer. Lowering maps Kind → match
// name via the `shapeMatchNames` table in `lower_shape_node.go`.
type ShapeNodeKind int

const (
	ShapeKindRect    ShapeNodeKind = iota // `ADBE Vector Shape - Rect`
	ShapeKindEllipse                      // `ADBE Vector Shape - Ellipse`
	ShapeKindPath                         // `ADBE Vector Shape - Group`
	ShapeKindFill                         // `ADBE Vector Graphic - Fill`
	ShapeKindStroke                       // `ADBE Vector Graphic - Stroke`
	ShapeKindGroup                        // `ADBE Vector Group` (V2.3+ user-created nested group)
	// V2.3+ candidates: PolyStar / GradientFill / GradientStroke / Trim / Merge /
	// Repeater / Transform.
)

// ShapeNode is the runtime-facing shape-graph node interface. All concrete
// node types (RectNode / EllipseNode / PathNode / FillNode / StrokeNode and
// the V2.3+ container VectorGroup) satisfy it.
type ShapeNode interface {
	Kind() ShapeNodeKind
	Properties() *PropertyGroup // escape hatch β; currently returns nil
}

// VectorGroup is the shape-graph container node. Every ShapeLayer carries
// one default RootGroup (constructed by WrapShapeLayer). Children render in
// order: Children[0] = bottom; Children[len-1] = top / most recently
// appended (render-order convention).
//
// `Transform` is the group-level Transform PropertyGroup placeholder. V2.2
// default-serialized form has no `ADBE Vector Transform Group`; the
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

// AddRect appends a default-valued RectNode and returns it. The new node is
// placed at the top of the render stack (Children[len-1]).
func (g *VectorGroup) AddRect() (*RectNode, error) {
	r := NewRectNode()
	g.Children = append(g.Children, r)
	return r, nil
}

// AddEllipse appends a default-valued EllipseNode and returns it.
func (g *VectorGroup) AddEllipse() (*EllipseNode, error) {
	e := NewEllipseNode()
	g.Children = append(g.Children, e)
	return e, nil
}

// AddPath appends an empty (closed) PathNode and returns it. Caller must
// call SetVertices to give it geometry (min 2 vertices).
func (g *VectorGroup) AddPath() (*PathNode, error) {
	p := NewPathNode()
	g.Children = append(g.Children, p)
	return p, nil
}

// AddFill appends a default-valued FillNode (white, 100% opacity) and
// returns it.
func (g *VectorGroup) AddFill() (*FillNode, error) {
	f := NewFillNode()
	g.Children = append(g.Children, f)
	return f, nil
}

// AddStroke appends a default-valued StrokeNode (black, width=2, 100%
// opacity) and returns it.
func (g *VectorGroup) AddStroke() (*StrokeNode, error) {
	s := NewStrokeNode()
	g.Children = append(g.Children, s)
	return s, nil
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
// Position=[0,0], Roundness=0 (AE elides all three at default).
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

// Properties returns the escape-hatch β view onto this RectNode's streams.
// Streams returned via PropertyGroup.Vec2Stream / Float64Stream
// are the same instances as the typed accessors (r.Size() etc.) — mutating
// one reflects through the other.
func (r *RectNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Rect",
		streams: map[string]any{
			"Size":      r.size,
			"Position":  r.position,
			"Roundness": r.roundness,
		},
	}
}

// EllipseNode — `ADBE Vector Shape - Ellipse`. Default Size=[100,100],
// Position=[0,0]. AE child[1] = `ADBE Vector Shape Direction` —
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

// Properties returns the escape-hatch β view.
func (e *EllipseNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Ellipse",
		streams: map[string]any{
			"Size":     e.size,
			"Position": e.position,
		},
	}
}

// PathNode — `ADBE Vector Shape - Group`. Default = empty Vertices,
// Closed=true. V2.2 SetVertices builds linear segments (tangents=0); min
// 2 vertices is enforced runtime-side (conservative invariant — AE itself
// was not directly probed for 0/1-vertex rejection).
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

// Properties returns the escape-hatch β view.
func (p *PathNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Path",
		streams: map[string]any{
			"Path": p.path,
		},
	}
}

// SetVertices replaces the path's vertex list with linear segments
// (tangents zeroed). Preserves the current `Closed` flag. Requires
// len(verts) >= 2.
func (p *PathNode) SetVertices(verts [][2]float64) error {
	if len(verts) < 2 {
		return fmt.Errorf("PathNode.SetVertices: need >= 2 vertices, got %d", len(verts))
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
// Opacity=100 (runtime defaults; AE elides at default).
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

// Properties returns the escape-hatch β view.
func (f *FillNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Fill",
		streams: map[string]any{
			"Color":   f.color,
			"Opacity": f.opacity,
		},
	}
}

// StrokeLineCap is the stroke end-cap style (`ADBE Vector Stroke Line Cap`).
// Stored on disk as a 1-based float64 enum index.
type StrokeLineCap int

const (
	StrokeLineCapButt       StrokeLineCap = 1 // default
	StrokeLineCapRound      StrokeLineCap = 2
	StrokeLineCapProjecting StrokeLineCap = 3
)

// StrokeLineJoin is the stroke corner-join style (`ADBE Vector Stroke Line
// Join`). Stored on disk as a 1-based float64 enum index.
type StrokeLineJoin int

const (
	StrokeLineJoinMiter StrokeLineJoin = 1 // default
	StrokeLineJoinRound StrokeLineJoin = 2
	StrokeLineJoinBevel StrokeLineJoin = 3
)

// StrokeNode — `ADBE Vector Graphic - Stroke`. Default Color=[0,0,0,1]
// black, Width=2, Opacity=100, Line Cap=Butt, Line Join=Miter, Miter Limit=4.
type StrokeNode struct {
	color   *PropertyStream[[4]float64]
	opacity *PropertyStream[float64]
	width   *PropertyStream[float64]

	// Line Cap / Line Join are enums; Miter Limit is a scalar. AE does not
	// animate them, so they are plain values rather than PropertyStreams.
	lineCap    StrokeLineCap
	lineJoin   StrokeLineJoin
	miterLimit float64
}

// NewStrokeNode constructs a default-valued StrokeNode.
func NewStrokeNode() *StrokeNode {
	s := &StrokeNode{
		color:      NewPropertyStream[[4]float64](),
		opacity:    NewPropertyStream[float64](),
		width:      NewPropertyStream[float64](),
		lineCap:    StrokeLineCapButt,
		lineJoin:   StrokeLineJoinMiter,
		miterLimit: 4,
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

func (s *StrokeNode) LineCap() StrokeLineCap   { return s.lineCap }
func (s *StrokeNode) LineJoin() StrokeLineJoin { return s.lineJoin }
func (s *StrokeNode) MiterLimit() float64      { return s.miterLimit }

// SetLineCap sets the end-cap style. Rejects values outside {Butt,Round,Projecting}.
func (s *StrokeNode) SetLineCap(v StrokeLineCap) error {
	if v < StrokeLineCapButt || v > StrokeLineCapProjecting {
		return fmt.Errorf("StrokeNode.SetLineCap: invalid value %d (want 1..3)", v)
	}
	s.lineCap = v
	return nil
}

// SetLineJoin sets the corner-join style. Rejects values outside {Miter,Round,Bevel}.
func (s *StrokeNode) SetLineJoin(v StrokeLineJoin) error {
	if v < StrokeLineJoinMiter || v > StrokeLineJoinBevel {
		return fmt.Errorf("StrokeNode.SetLineJoin: invalid value %d (want 1..3)", v)
	}
	s.lineJoin = v
	return nil
}

// SetMiterLimit sets the miter limit. AE only applies it when Line Join =
// Miter, but the value is stored regardless. Rejects values < 1.
func (s *StrokeNode) SetMiterLimit(v float64) error {
	if v < 1 {
		return fmt.Errorf("StrokeNode.SetMiterLimit: %g out of range (want >= 1)", v)
	}
	s.miterLimit = v
	return nil
}

// Properties returns the escape-hatch β view.
func (s *StrokeNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Stroke",
		streams: map[string]any{
			"Color":   s.color,
			"Opacity": s.opacity,
			"Width":   s.width,
		},
	}
}

// PropertyGroup is the escape-hatch β surface. Currently the minimal
// struct — `Name` and the empty `Children` / `streams` maps —
// so the field exists on VectorGroup.Transform / node Properties() but
// none of the typed lookup methods (`Float64Stream`, `Vec2Stream`, ...)
// are wired yet.
type PropertyGroup struct {
	Name      string
	Children  map[string]*PropertyGroup
	streams   map[string]any
	Separated bool // false in V2.2; V2.3+ exposes separated dimension setters
}

// newGroupTransform constructs the identity Transform PropertyGroup placeholder
// for a fresh VectorGroup. RootGroup default-serialized form has no
// `ADBE Vector Transform Group` — this placeholder stays nil-children
// until the escape hatch wires it.
func newGroupTransform() *PropertyGroup {
	return &PropertyGroup{Name: "Transform"}
}

// Child returns the nested PropertyGroup by name, or nil if not present.
// Use for walking deeper-than-leaf escape-hatch trees (V2.3+ nested groups).
func (pg *PropertyGroup) Child(name string) *PropertyGroup {
	if pg == nil || pg.Children == nil {
		return nil
	}
	return pg.Children[name]
}

// Float64Stream returns the PropertyStream[float64] under the given name, or
// an error if no stream by that name exists or it isn't the expected type.
// Mutations on the returned stream are visible through the typed accessor.
func (pg *PropertyGroup) Float64Stream(name string) (*PropertyStream[float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not float64", pg.Name, name)
	}
	return ps, nil
}

// Vec2Stream returns the PropertyStream[[2]float64] under the given name.
func (pg *PropertyGroup) Vec2Stream(name string) (*PropertyStream[[2]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[[2]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [2]float64", pg.Name, name)
	}
	return ps, nil
}

// Vec3Stream returns the PropertyStream[[3]float64] under the given name.
func (pg *PropertyGroup) Vec3Stream(name string) (*PropertyStream[[3]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[[3]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [3]float64", pg.Name, name)
	}
	return ps, nil
}

// ColorStream returns the PropertyStream[[4]float64] (RGBA) under the given name.
func (pg *PropertyGroup) ColorStream(name string) (*PropertyStream[[4]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[[4]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [4]float64", pg.Name, name)
	}
	return ps, nil
}

// PathStream returns the PropertyStream[BezierPath] under the given name.
func (pg *PropertyGroup) PathStream(name string) (*PropertyStream[BezierPath], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*PropertyStream[BezierPath])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not BezierPath", pg.Name, name)
	}
	return ps, nil
}
