// internal/aep/shape_graph.go
package scene

import (
	"fmt"

	"github.com/example/aep-parser/internal/codec"
)

// ShapeNodeKind identifies a shape-graph node's runtime kind. AE match-name
// strings (`ADBE Vector Shape - Rect` etc.) are intentionally NOT exposed
// here — they belong to the serializer. Lowering maps Kind → match
// name via the `shapeMatchNames` table in `lower_shape_node.go`.
type ShapeNodeKind int

const (
	ShapeKindRect            ShapeNodeKind = iota // `ADBE Vector Shape - Rect`
	ShapeKindEllipse                              // `ADBE Vector Shape - Ellipse`
	ShapeKindPath                                 // `ADBE Vector Shape - Group`
	ShapeKindFill                                 // `ADBE Vector Graphic - Fill`
	ShapeKindStroke                               // `ADBE Vector Graphic - Stroke`
	ShapeKindGroup                                // `ADBE Vector Group` (V2.3+ user-created nested group)
	ShapeKindGradientFill                         // `ADBE Vector Graphic - G-Fill`
	ShapeKindGradientStroke                       // `ADBE Vector Graphic - G-Stroke`
	ShapeKindTrim                                 // `ADBE Vector Filter - Trim`
	ShapeKindRepeater                             // `ADBE Vector Filter - Repeater`
	ShapeKindRoundCorners                         // `ADBE Vector Filter - RC`
	ShapeKindOffsetPaths                          // `ADBE Vector Filter - Offset`
	ShapeKindMergePaths                           // `ADBE Vector Filter - Merge`
	ShapeKindZigZag                               // `ADBE Vector Filter - Zigzag`
	ShapeKindStar                                 // `ADBE Vector Shape - Star`
	ShapeKindPuckerBloat                          // `ADBE Vector Filter - PB`
	ShapeKindTwist                                // `ADBE Vector Filter - Twist`
	ShapeKindWigglePaths                          // `ADBE Vector Filter - Roughen`
	ShapeKindWiggleTransform                      // `ADBE Vector Filter - Wiggler`
	// V2.3+ candidates: Transform / nested user groups.
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

// AddStar appends a default-valued StarNode (5-point star, OuterRadius=100,
// InnerRadius=50) and returns it — the canonical star / sparkle / badge MG
// primitive. V2.2 models the Star type only (not Polygon).
func (g *VectorGroup) AddStar() (*StarNode, error) {
	n := NewStarNode()
	g.Children = append(g.Children, n)
	return n, nil
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

// AddGradientFill appends a default-valued GradientFillNode (2-stop black→white
// linear gradient, fully opaque) and returns it. Set the stops via
// SetColorStops / SetAlphaStops.
func (g *VectorGroup) AddGradientFill() (*GradientFillNode, error) {
	n := NewGradientFillNode()
	g.Children = append(g.Children, n)
	return n, nil
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
// ShapeDirection is the parametric-shape path direction (`ADBE Vector Shape
// Direction`), shared by Rect/Ellipse. Stored as a float64 enum index.
type ShapeDirection int

const (
	ShapeDirectionNormal   ShapeDirection = 1 // default
	ShapeDirectionReversed ShapeDirection = 3
)

// ShapeBlendMode is a Fill/Stroke graphic blend mode (`ADBE Vector Blend
// Mode`). Stored as AE's 1-based blend-mode index (Normal=1); the full mode
// list is large, so pass the AE index directly.
type ShapeBlendMode int

const ShapeBlendModeNormal ShapeBlendMode = 1 // default

// ShapeCompositeOrder controls how a Fill/Stroke composites within its group
// (`ADBE Vector Composite Order`).
type ShapeCompositeOrder int

const (
	ShapeCompositeOrderAbovePrevious ShapeCompositeOrder = 1 // default
	ShapeCompositeOrderBelowPrevious ShapeCompositeOrder = 2
)

// FillRule is the Fill winding rule (`ADBE Vector Fill Rule`).
type FillRule int

const (
	FillRuleNonzeroWinding FillRule = 1 // default
	FillRuleEvenOdd        FillRule = 2
)

type RectNode struct {
	size      *codec.PropertyStream[[2]float64]
	position  *codec.PropertyStream[[2]float64]
	roundness *codec.PropertyStream[float64]
	direction ShapeDirection
}

// NewRectNode constructs a default-valued RectNode.
func NewRectNode() *RectNode {
	r := &RectNode{
		size:      codec.NewPropertyStream[[2]float64](),
		position:  codec.NewPropertyStream[[2]float64](),
		roundness: codec.NewPropertyStream[float64](),
		direction: ShapeDirectionNormal,
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
func (r *RectNode) Direction() ShapeDirection             { return r.direction }
func (r *RectNode) SetSize(v [2]float64) error            { return r.size.SetStaticValue(v) }
func (r *RectNode) SetPosition(v [2]float64) error        { return r.position.SetStaticValue(v) }
func (r *RectNode) SetRoundness(v float64) error          { return r.roundness.SetStaticValue(v) }
func (r *RectNode) SetDirection(v ShapeDirection) error   { return setShapeDirection(&r.direction, v) }

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
	size, position *codec.PropertyStream[[2]float64]
	direction      ShapeDirection
}

// NewEllipseNode constructs a default-valued EllipseNode.
func NewEllipseNode() *EllipseNode {
	e := &EllipseNode{
		size:      codec.NewPropertyStream[[2]float64](),
		position:  codec.NewPropertyStream[[2]float64](),
		direction: ShapeDirectionNormal,
	}
	_ = e.size.SetStaticValue([2]float64{100, 100})
	_ = e.position.SetStaticValue([2]float64{0, 0})
	return e
}

func (e *EllipseNode) Kind() ShapeNodeKind                   { return ShapeKindEllipse }
func (e *EllipseNode) Size() *PropertyStream[[2]float64]     { return e.size }
func (e *EllipseNode) Position() *PropertyStream[[2]float64] { return e.position }
func (e *EllipseNode) Direction() ShapeDirection             { return e.direction }
func (e *EllipseNode) SetSize(v [2]float64) error            { return e.size.SetStaticValue(v) }
func (e *EllipseNode) SetPosition(v [2]float64) error        { return e.position.SetStaticValue(v) }
func (e *EllipseNode) SetDirection(v ShapeDirection) error   { return setShapeDirection(&e.direction, v) }

// setShapeDirection validates and assigns a ShapeDirection (Normal=1 or
// Reversed=3; AE has no value 2 for parametric shapes).
func setShapeDirection(dst *ShapeDirection, v ShapeDirection) error {
	if v != ShapeDirectionNormal && v != ShapeDirectionReversed {
		return fmt.Errorf("SetDirection: invalid value %d (want 1=Normal or 3=Reversed)", v)
	}
	*dst = v
	return nil
}

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

// StarType selects a polystar's shape: a Star (alternating outer/inner radius
// points) or a Polygon (a convex N-gon — Inner Radius/Roundness have no effect).
// Matches AE's `ADBE Vector Star Type` enum.
type StarType int

const (
	// StarTypeStar is the alternating-point star (AE default).
	StarTypeStar StarType = 1
	// StarTypePolygon is a convex N-gon (Inner Radius/Roundness ignored).
	StarTypePolygon StarType = 2
)

// StarNode — `ADBE Vector Shape - Star`. A parametric polystar: a Star
// (alternating outer/inner points) or a Polygon (convex N-gon) per StarType.
// Defaults match AE: Type=Star, Points=5, Position=[0,0], Rotation=0,
// InnerRadius=50, OuterRadius=100, Inner/OuterRoundness=0. For a Polygon the
// Inner Radius / Inner Roundness sub-streams have no visual effect (AE hides
// them), but Points / Position / Rotation / Outer Radius / Outer Roundness apply.
//
// All sub-streams are animatable: Position is a Vec2 (spatial motion-path, same
// layout as Rect Position); the rest are 1D scalars.
type StarNode struct {
	starType       StarType
	points         *codec.PropertyStream[float64]
	position       *codec.PropertyStream[[2]float64]
	rotation       *codec.PropertyStream[float64]
	innerRadius    *codec.PropertyStream[float64]
	outerRadius    *codec.PropertyStream[float64]
	innerRoundness *codec.PropertyStream[float64]
	outerRoundness *codec.PropertyStream[float64]
}

// NewStarNode constructs a default-valued StarNode (AE's default 5-point star).
func NewStarNode() *StarNode {
	n := &StarNode{
		starType:       StarTypeStar,
		points:         codec.NewPropertyStream[float64](),
		position:       codec.NewPropertyStream[[2]float64](),
		rotation:       codec.NewPropertyStream[float64](),
		innerRadius:    codec.NewPropertyStream[float64](),
		outerRadius:    codec.NewPropertyStream[float64](),
		innerRoundness: codec.NewPropertyStream[float64](),
		outerRoundness: codec.NewPropertyStream[float64](),
	}
	_ = n.points.SetStaticValue(5)
	_ = n.position.SetStaticValue([2]float64{0, 0})
	_ = n.rotation.SetStaticValue(0)
	_ = n.innerRadius.SetStaticValue(50)
	_ = n.outerRadius.SetStaticValue(100)
	_ = n.innerRoundness.SetStaticValue(0)
	_ = n.outerRoundness.SetStaticValue(0)
	return n
}

// StarType returns whether this polystar is a Star or a Polygon.
func (n *StarNode) StarType() StarType { return n.starType }

// IsPolygon reports whether this polystar is a Polygon (vs a Star).
func (n *StarNode) IsPolygon() bool { return n.starType == StarTypePolygon }

// SetStarType selects Star (alternating points) or Polygon (convex N-gon). For a
// Polygon the Inner Radius/Roundness are ignored.
func (n *StarNode) SetStarType(t StarType) error {
	if t != StarTypeStar && t != StarTypePolygon {
		return fmt.Errorf("StarNode.SetStarType: %d out of range (1=star, 2=polygon)", t)
	}
	n.starType = t
	return nil
}

func (n *StarNode) Kind() ShapeNodeKind                      { return ShapeKindStar }
func (n *StarNode) Points() *PropertyStream[float64]         { return n.points }
func (n *StarNode) Position() *PropertyStream[[2]float64]    { return n.position }
func (n *StarNode) Rotation() *PropertyStream[float64]       { return n.rotation }
func (n *StarNode) InnerRadius() *PropertyStream[float64]    { return n.innerRadius }
func (n *StarNode) OuterRadius() *PropertyStream[float64]    { return n.outerRadius }
func (n *StarNode) InnerRoundness() *PropertyStream[float64] { return n.innerRoundness }
func (n *StarNode) OuterRoundness() *PropertyStream[float64] { return n.outerRoundness }

// SetPoints sets the number of star points. Rejects values < 3.
func (n *StarNode) SetPoints(v float64) error {
	if v < 3 {
		return fmt.Errorf("StarNode.SetPoints: %g out of range (want >= 3)", v)
	}
	return n.points.SetStaticValue(v)
}

// SetPosition sets the star's local position offset (px).
func (n *StarNode) SetPosition(v [2]float64) error { return n.position.SetStaticValue(v) }

// SetRotation sets the star's rotation in degrees.
func (n *StarNode) SetRotation(v float64) error { return n.rotation.SetStaticValue(v) }

// SetInnerRadius sets the inner radius (px, the valley between points). Rejects negatives.
func (n *StarNode) SetInnerRadius(v float64) error {
	if v < 0 {
		return fmt.Errorf("StarNode.SetInnerRadius: %g out of range (want >= 0)", v)
	}
	return n.innerRadius.SetStaticValue(v)
}

// SetOuterRadius sets the outer radius (px, the star tips). Rejects negatives.
func (n *StarNode) SetOuterRadius(v float64) error {
	if v < 0 {
		return fmt.Errorf("StarNode.SetOuterRadius: %g out of range (want >= 0)", v)
	}
	return n.outerRadius.SetStaticValue(v)
}

// SetInnerRoundness sets the inner-point roundness (percent).
func (n *StarNode) SetInnerRoundness(v float64) error { return n.innerRoundness.SetStaticValue(v) }

// SetOuterRoundness sets the outer-point (tip) roundness (percent).
func (n *StarNode) SetOuterRoundness(v float64) error { return n.outerRoundness.SetStaticValue(v) }

// Properties returns the escape-hatch β view.
func (n *StarNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Star",
		streams: map[string]any{
			"Points":         n.points,
			"Position":       n.position,
			"Rotation":       n.rotation,
			"InnerRadius":    n.innerRadius,
			"OuterRadius":    n.outerRadius,
			"InnerRoundness": n.innerRoundness,
			"OuterRoundness": n.outerRoundness,
		},
	}
}

// PathNode — `ADBE Vector Shape - Group`. Default = empty Vertices,
// Closed=true. V2.2 SetVertices builds linear segments (tangents=0); min
// 2 vertices is enforced runtime-side (conservative invariant — AE itself
// was not directly probed for 0/1-vertex rejection).
type PathNode struct {
	path *codec.PropertyStream[BezierPath]
}

// NewPathNode constructs an empty PathNode (Closed=true).
func NewPathNode() *PathNode {
	p := &PathNode{path: codec.NewPropertyStream[BezierPath]()}
	_ = p.path.SetStaticValue(BezierPath{Closed: true})
	return p
}

func (p *PathNode) Kind() ShapeNodeKind               { return ShapeKindPath }
func (p *PathNode) Path() *PropertyStream[BezierPath] { return p.path }

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
// Opacity=100, Blend Mode=Normal, Composite Order=Above Previous, Fill
// Rule=Nonzero Winding (runtime defaults; AE elides at default).
type FillNode struct {
	color          *codec.PropertyStream[[4]float64]
	opacity        *codec.PropertyStream[float64]
	blendMode      ShapeBlendMode
	compositeOrder ShapeCompositeOrder
	fillRule       FillRule
}

// NewFillNode constructs a default-valued FillNode.
func NewFillNode() *FillNode {
	f := &FillNode{
		color:          codec.NewPropertyStream[[4]float64](),
		opacity:        codec.NewPropertyStream[float64](),
		blendMode:      ShapeBlendModeNormal,
		compositeOrder: ShapeCompositeOrderAbovePrevious,
		fillRule:       FillRuleNonzeroWinding,
	}
	_ = f.color.SetStaticValue([4]float64{1, 1, 1, 1}) // white
	_ = f.opacity.SetStaticValue(100)
	return f
}

func (f *FillNode) Kind() ShapeNodeKind                 { return ShapeKindFill }
func (f *FillNode) Color() *PropertyStream[[4]float64]  { return f.color }
func (f *FillNode) Opacity() *PropertyStream[float64]   { return f.opacity }
func (f *FillNode) BlendMode() ShapeBlendMode           { return f.blendMode }
func (f *FillNode) CompositeOrder() ShapeCompositeOrder { return f.compositeOrder }
func (f *FillNode) FillRule() FillRule                  { return f.fillRule }
func (f *FillNode) SetColor(v [4]float64) error         { return f.color.SetStaticValue(v) }
func (f *FillNode) SetOpacity(v float64) error          { return f.opacity.SetStaticValue(v) }
func (f *FillNode) SetBlendMode(v ShapeBlendMode) error { return setShapeBlendMode(&f.blendMode, v) }
func (f *FillNode) SetCompositeOrder(v ShapeCompositeOrder) error {
	return setShapeCompositeOrder(&f.compositeOrder, v)
}

// SetFillRule sets the winding rule (NonzeroWinding=1 / EvenOdd=2).
func (f *FillNode) SetFillRule(v FillRule) error {
	if v != FillRuleNonzeroWinding && v != FillRuleEvenOdd {
		return fmt.Errorf("SetFillRule: invalid value %d (want 1=Nonzero or 2=EvenOdd)", v)
	}
	f.fillRule = v
	return nil
}

// setShapeBlendMode validates (>= 1) and assigns a ShapeBlendMode.
func setShapeBlendMode(dst *ShapeBlendMode, v ShapeBlendMode) error {
	if v < 1 {
		return fmt.Errorf("SetBlendMode: invalid value %d (want >= 1)", v)
	}
	*dst = v
	return nil
}

// setShapeCompositeOrder validates (AbovePrevious=1 / BelowPrevious=2).
func setShapeCompositeOrder(dst *ShapeCompositeOrder, v ShapeCompositeOrder) error {
	if v != ShapeCompositeOrderAbovePrevious && v != ShapeCompositeOrderBelowPrevious {
		return fmt.Errorf("SetCompositeOrder: invalid value %d (want 1 or 2)", v)
	}
	*dst = v
	return nil
}

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

// GradientFillNode — `ADBE Vector Graphic - G-Fill`. Models the gradient's
// color + alpha stops (`ADBE Vector Grad Colors`) plus the linear ramp direction
// (`ADBE Vector Grad Start Pt` / `End Pt`). The ramp runs from StartPoint to
// EndPoint in the shape's local coordinate space (AE default [0,0]→[100,0], a
// horizontal ramp); set them to a diagonal/vertical pair to rotate the gradient.
//
// `Grad Type` selects linear vs radial (StartPoint = centre, EndPoint = a point
// on the radius for radial). For a radial gradient the HiLite controls offset
// the bright centre (start color) off the geometric centre: HighlightLength is
// the shift magnitude as a percent of the radius (0 = centred, the default) and
// HighlightAngle is its direction in degrees. They have no visible effect on a
// linear gradient. Stops + direction + type + highlight are static (V2.2 does
// not model animated gradients). The serializer re-encodes the stops to prop.map
// XML and overwrites the Grad Type / Start Pt / End Pt / HiLite Length / HiLite
// Angle cdats + GCky/Utf8 chunk (length-variable; rifx recomputes the enclosing
// LIST sizes).
type GradientFillNode struct {
	gradient        *codec.Gradient
	gradientKfs     []GradientKeyframe
	startPoint      [2]float64
	endPoint        [2]float64
	gradientType    GradientType
	highlightLength float64
	highlightAngle  float64
}

// GradientKeyframe pairs a time (seconds) with a complete gradient value
// (color + alpha stops). Animated gradient stops are a list of these — each
// keyframe carries the full stop set at that time, and AE interpolates the
// stops between keyframes. See GradientFillNode.AddGradientKeyframe.
type GradientKeyframe struct {
	Time     float64
	Gradient *Gradient
}

// GradientType selects a gradient fill's ramp shape: linear (a straight band) or
// radial (concentric rings from StartPoint out to EndPoint). Matches AE's
// `ADBE Vector Grad Type` enum.
type GradientType int

const (
	// GradientLinear is a straight ramp from StartPoint to EndPoint (AE default).
	GradientLinear GradientType = 1
	// GradientRadial is concentric rings centred at StartPoint, reaching the
	// EndPoint color at radius |EndPoint − StartPoint|.
	GradientRadial GradientType = 2
)

// NewGradientFillNode constructs a default 2-stop black→white linear gradient
// (fully opaque) with AE's default horizontal ramp ([0,0]→[100,0]). Callers
// override stops via SetColorStops / SetAlphaStops, direction via
// SetStartPoint / SetEndPoint, and ramp shape via SetGradientType.
func NewGradientFillNode() *GradientFillNode {
	return &GradientFillNode{
		gradient:     defaultGradient(),
		startPoint:   [2]float64{0, 0},
		endPoint:     [2]float64{100, 0},
		gradientType: GradientLinear,
	}
}

// StartPoint returns the gradient ramp's start point (shape-local coords).
func (n *GradientFillNode) StartPoint() [2]float64 { return n.startPoint }

// EndPoint returns the gradient ramp's end point (shape-local coords).
func (n *GradientFillNode) EndPoint() [2]float64 { return n.endPoint }

// SetStartPoint sets the gradient ramp's start point (shape-local coords). The
// ramp direction is EndPoint − StartPoint; defaults to a horizontal [0,0]→[100,0].
func (n *GradientFillNode) SetStartPoint(v [2]float64) error { n.startPoint = v; return nil }

// SetEndPoint sets the gradient ramp's end point (shape-local coords).
func (n *GradientFillNode) SetEndPoint(v [2]float64) error { n.endPoint = v; return nil }

// GradientType returns the ramp shape (linear or radial).
func (n *GradientFillNode) GradientType() GradientType { return n.gradientType }

// SetGradientType selects linear (default) or radial ramp shape. For radial,
// StartPoint is the centre and EndPoint sets the outer radius.
func (n *GradientFillNode) SetGradientType(t GradientType) error {
	if t != GradientLinear && t != GradientRadial {
		return fmt.Errorf("gradient type %d out of range (1=linear, 2=radial)", t)
	}
	n.gradientType = t
	return nil
}

// HighlightLength returns the radial highlight offset magnitude (percent of the
// radius; 0 = centred).
func (n *GradientFillNode) HighlightLength() float64 { return n.highlightLength }

// HighlightAngle returns the radial highlight offset direction (degrees).
func (n *GradientFillNode) HighlightAngle() float64 { return n.highlightAngle }

// SetHighlightLength offsets a radial gradient's bright centre off the geometric
// centre by the given percent of the radius (-100..100; 0 = centred). No visible
// effect on a linear gradient. Returns an error if out of range.
func (n *GradientFillNode) SetHighlightLength(v float64) error {
	if v < -100 || v > 100 {
		return fmt.Errorf("highlight length %g out of range [-100,100]", v)
	}
	n.highlightLength = v
	return nil
}

// SetHighlightAngle sets the direction (degrees) of a radial gradient's highlight
// offset. Only meaningful together with a non-zero HighlightLength.
func (n *GradientFillNode) SetHighlightAngle(v float64) error {
	n.highlightAngle = v
	return nil
}

// defaultGradient returns a 2-stop black→white gradient with two opaque alpha
// stops — the values AE shows for a freshly-added gradient fill.
func defaultGradient() *codec.Gradient {
	return &codec.Gradient{
		Version: "4",
		ColorStops: []codec.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{0, 0, 0}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{1, 1, 1}},
		},
		AlphaStops: []codec.GradientAlphaStop{
			{Offset: 0, Midpoint: 0.5, Alpha: 1},
			{Offset: 1, Midpoint: 0.5, Alpha: 1},
		},
	}
}

func (n *GradientFillNode) Kind() ShapeNodeKind { return ShapeKindGradientFill }

// Gradient returns the live gradient (color + alpha stops). Mutating the
// returned struct's slices directly also works, but prefer SetColorStops /
// SetAlphaStops for range validation.
func (n *GradientFillNode) Gradient() *Gradient { return n.gradient }

// SetColorStops replaces the gradient's color stops. Requires ≥ 2 stops; each
// Offset/Midpoint in [0,1] and each Color component in [0,1].
func (n *GradientFillNode) SetColorStops(stops []GradientColorStop) error {
	return setGradientColorStops(n.gradient, stops)
}

// SetAlphaStops replaces the gradient's alpha (opacity) stops. Requires ≥ 2
// stops; each Offset/Midpoint/Alpha in [0,1].
func (n *GradientFillNode) SetAlphaStops(stops []GradientAlphaStop) error {
	return setGradientAlphaStops(n.gradient, stops)
}

// AddGradientKeyframe appends an animated-stops keyframe: the full gradient g
// (color + alpha stops) takes effect at `time` seconds, and AE interpolates the
// stops between keyframes (a colour sweep / flow). The first keyframe switches
// the node to animated mode — the static Gradient() value is then ignored on
// lower in favour of the keyframe list. Provide keyframes in ascending time.
// g's stops are validated (≥2 stops; offsets/colours in range).
//
// On-disk this writes the `ADBE Vector Grad Colors` stream as a keyframe
// time-table plus one prop.map-XML leaf per keyframe (see
// incidents/gradient-fill-write-re.md § animated color stops). Write-only:
// re-parsing surfaces the first keyframe's stops as the static value.
func (n *GradientFillNode) AddGradientKeyframe(time float64, g *Gradient) error {
	if g == nil {
		return fmt.Errorf("AddGradientKeyframe: nil gradient")
	}
	if err := setGradientColorStops(g, g.ColorStops); err != nil {
		return err
	}
	if err := setGradientAlphaStops(g, g.AlphaStops); err != nil {
		return err
	}
	n.gradientKfs = append(n.gradientKfs, GradientKeyframe{Time: time, Gradient: g})
	return nil
}

// GradientKeyframes returns the animated-stops keyframes (nil when the node is
// static). Used by the serializer.
func (n *GradientFillNode) GradientKeyframes() []GradientKeyframe { return n.gradientKfs }

// setGradientColorStops validates (≥2 stops; offset/midpoint/color ∈ [0,1])
// and replaces g.ColorStops. Shared by GradientFill / GradientStroke.
func setGradientColorStops(g *codec.Gradient, stops []GradientColorStop) error {
	if len(stops) < 2 {
		return fmt.Errorf("SetColorStops: need ≥ 2 stops, got %d", len(stops))
	}
	for i, s := range stops {
		if err := checkUnit("offset", i, s.Offset); err != nil {
			return err
		}
		if err := checkUnit("midpoint", i, s.Midpoint); err != nil {
			return err
		}
		for c, v := range s.Color {
			if v < 0 || v > 1 {
				return fmt.Errorf("SetColorStops: stop %d color[%d] = %g out of range [0,1]", i, c, v)
			}
		}
	}
	g.ColorStops = append([]GradientColorStop(nil), stops...)
	return nil
}

// setGradientAlphaStops validates (≥2 stops; offset/midpoint/alpha ∈ [0,1])
// and replaces g.AlphaStops. Shared by GradientFill / GradientStroke.
func setGradientAlphaStops(g *codec.Gradient, stops []GradientAlphaStop) error {
	if len(stops) < 2 {
		return fmt.Errorf("SetAlphaStops: need ≥ 2 stops, got %d", len(stops))
	}
	for i, s := range stops {
		if err := checkUnit("offset", i, s.Offset); err != nil {
			return err
		}
		if err := checkUnit("midpoint", i, s.Midpoint); err != nil {
			return err
		}
		if s.Alpha < 0 || s.Alpha > 1 {
			return fmt.Errorf("SetAlphaStops: stop %d alpha = %g out of range [0,1]", i, s.Alpha)
		}
	}
	g.AlphaStops = append([]GradientAlphaStop(nil), stops...)
	return nil
}

func checkUnit(what string, i int, v float64) error {
	if v < 0 || v > 1 {
		return fmt.Errorf("gradient stop %d %s = %g out of range [0,1]", i, what, v)
	}
	return nil
}

// Properties returns the escape-hatch β view.
func (n *GradientFillNode) Properties() *PropertyGroup {
	return &PropertyGroup{Name: "Gradient Fill"}
}

// GradientStrokeNode — `ADBE Vector Graphic - G-Stroke`. Models the gradient's
// color + alpha stops plus the ramp geometry (direction, type, highlight),
// symmetric to GradientFillNode: StartPoint/EndPoint set the ramp direction
// (AE default [0,0]→[100,0]), GradientType selects linear vs radial, and the
// HiLite controls offset a radial gradient's bright centre. Stroke geometry
// (width / cap / join / dashes / taper / wave) is NOT modeled (kept at the
// extracted template's values; deferred).
type GradientStrokeNode struct {
	gradient        *codec.Gradient
	gradientKfs     []GradientKeyframe
	startPoint      [2]float64
	endPoint        [2]float64
	gradientType    GradientType
	highlightLength float64
	highlightAngle  float64
	strokeWidth     float64
	lineCap         StrokeLineCap
	lineJoin        StrokeLineJoin
	miterLimit      float64
}

// NewGradientStrokeNode constructs a default 2-stop black→white linear gradient
// stroke with AE's default horizontal ramp ([0,0]→[100,0]). The stroke geometry
// (width / cap / join / miter) defaults to the embedded template's baked values
// (18px, round cap, round join, miter 4); override via the Set* methods.
func NewGradientStrokeNode() *GradientStrokeNode {
	return &GradientStrokeNode{
		gradient:     defaultGradient(),
		startPoint:   [2]float64{0, 0},
		endPoint:     [2]float64{100, 0},
		gradientType: GradientLinear,
		strokeWidth:  18,
		lineCap:      StrokeLineCapRound,
		lineJoin:     StrokeLineJoinRound,
		miterLimit:   4,
	}
}

// StrokeWidth returns the gradient stroke's width (pixels).
func (n *GradientStrokeNode) StrokeWidth() float64 { return n.strokeWidth }

// SetStrokeWidth sets the gradient stroke's width (pixels; must be ≥ 0).
func (n *GradientStrokeNode) SetStrokeWidth(v float64) error {
	if v < 0 {
		return fmt.Errorf("stroke width %g must be ≥ 0", v)
	}
	n.strokeWidth = v
	return nil
}

// LineCap returns the gradient stroke's end-cap style.
func (n *GradientStrokeNode) LineCap() StrokeLineCap { return n.lineCap }

// SetLineCap sets the gradient stroke's end-cap style (Butt / Round / Projecting).
func (n *GradientStrokeNode) SetLineCap(c StrokeLineCap) error {
	if c < StrokeLineCapButt || c > StrokeLineCapProjecting {
		return fmt.Errorf("invalid line cap %d (want 1..3)", c)
	}
	n.lineCap = c
	return nil
}

// LineJoin returns the gradient stroke's corner-join style.
func (n *GradientStrokeNode) LineJoin() StrokeLineJoin { return n.lineJoin }

// SetLineJoin sets the gradient stroke's corner-join style (Miter / Round / Bevel).
func (n *GradientStrokeNode) SetLineJoin(j StrokeLineJoin) error {
	if j < StrokeLineJoinMiter || j > StrokeLineJoinBevel {
		return fmt.Errorf("invalid line join %d (want 1..3)", j)
	}
	n.lineJoin = j
	return nil
}

// MiterLimit returns the gradient stroke's miter limit (only used with a miter join).
func (n *GradientStrokeNode) MiterLimit() float64 { return n.miterLimit }

// SetMiterLimit sets the gradient stroke's miter limit (only used when LineJoin
// is Miter; must be ≥ 1).
func (n *GradientStrokeNode) SetMiterLimit(v float64) error {
	if v < 1 {
		return fmt.Errorf("miter limit %g must be ≥ 1", v)
	}
	n.miterLimit = v
	return nil
}

func (n *GradientStrokeNode) Kind() ShapeNodeKind { return ShapeKindGradientStroke }

// StartPoint returns the gradient ramp's start point (shape-local coords).
func (n *GradientStrokeNode) StartPoint() [2]float64 { return n.startPoint }

// EndPoint returns the gradient ramp's end point (shape-local coords).
func (n *GradientStrokeNode) EndPoint() [2]float64 { return n.endPoint }

// SetStartPoint sets the gradient ramp's start point (shape-local coords). The
// ramp direction is EndPoint − StartPoint; defaults to a horizontal [0,0]→[100,0].
func (n *GradientStrokeNode) SetStartPoint(v [2]float64) error { n.startPoint = v; return nil }

// SetEndPoint sets the gradient ramp's end point (shape-local coords).
func (n *GradientStrokeNode) SetEndPoint(v [2]float64) error { n.endPoint = v; return nil }

// GradientType returns the ramp shape (linear or radial).
func (n *GradientStrokeNode) GradientType() GradientType { return n.gradientType }

// SetGradientType selects linear (default) or radial ramp shape. For radial,
// StartPoint is the centre and EndPoint sets the outer radius.
func (n *GradientStrokeNode) SetGradientType(t GradientType) error {
	if t != GradientLinear && t != GradientRadial {
		return fmt.Errorf("gradient type %d out of range (1=linear, 2=radial)", t)
	}
	n.gradientType = t
	return nil
}

// HighlightLength returns the radial highlight offset magnitude (percent of the
// radius; 0 = centred).
func (n *GradientStrokeNode) HighlightLength() float64 { return n.highlightLength }

// HighlightAngle returns the radial highlight offset direction (degrees).
func (n *GradientStrokeNode) HighlightAngle() float64 { return n.highlightAngle }

// SetHighlightLength offsets a radial gradient's bright centre off the geometric
// centre by the given percent of the radius (-100..100; 0 = centred). No visible
// effect on a linear gradient. Returns an error if out of range.
func (n *GradientStrokeNode) SetHighlightLength(v float64) error {
	if v < -100 || v > 100 {
		return fmt.Errorf("highlight length %g out of range [-100,100]", v)
	}
	n.highlightLength = v
	return nil
}

// SetHighlightAngle sets the direction (degrees) of a radial gradient's highlight
// offset. Only meaningful together with a non-zero HighlightLength.
func (n *GradientStrokeNode) SetHighlightAngle(v float64) error {
	n.highlightAngle = v
	return nil
}

// Gradient returns the live gradient (color + alpha stops).
func (n *GradientStrokeNode) Gradient() *Gradient { return n.gradient }

// SetColorStops replaces the gradient's color stops (≥2; ranges in [0,1]).
func (n *GradientStrokeNode) SetColorStops(stops []GradientColorStop) error {
	return setGradientColorStops(n.gradient, stops)
}

// SetAlphaStops replaces the gradient's alpha stops (≥2; ranges in [0,1]).
func (n *GradientStrokeNode) SetAlphaStops(stops []GradientAlphaStop) error {
	return setGradientAlphaStops(n.gradient, stops)
}

// AddGradientKeyframe appends an animated-stops keyframe to a gradient STROKE:
// the full gradient g (color + alpha stops) takes effect at `time` seconds, and
// AE interpolates the stops between keyframes (a colour sweep along the stroke).
// The first keyframe switches the node to animated mode — the static Gradient()
// value is then ignored on lower in favour of the keyframe list. Provide
// keyframes in ascending time. Identical mechanism to
// GradientFillNode.AddGradientKeyframe (the `ADBE Vector Grad Colors` stream is
// the same on fill and stroke); see incidents/gradient-fill-write-re.md §
// animated color stops. Write-only: re-parsing surfaces the first keyframe's
// stops as the static value.
func (n *GradientStrokeNode) AddGradientKeyframe(time float64, g *Gradient) error {
	if g == nil {
		return fmt.Errorf("AddGradientKeyframe: nil gradient")
	}
	if err := setGradientColorStops(g, g.ColorStops); err != nil {
		return err
	}
	if err := setGradientAlphaStops(g, g.AlphaStops); err != nil {
		return err
	}
	n.gradientKfs = append(n.gradientKfs, GradientKeyframe{Time: time, Gradient: g})
	return nil
}

// GradientKeyframes returns the animated-stops keyframes (nil when the node is
// static). Used by the serializer.
func (n *GradientStrokeNode) GradientKeyframes() []GradientKeyframe { return n.gradientKfs }

// Properties returns the escape-hatch β view.
func (n *GradientStrokeNode) Properties() *PropertyGroup {
	return &PropertyGroup{Name: "Gradient Stroke"}
}

// AddGradientStroke appends a default-valued GradientStrokeNode (2-stop
// black→white gradient) and returns it. Set stops via SetColorStops /
// SetAlphaStops.
func (g *VectorGroup) AddGradientStroke() (*GradientStrokeNode, error) {
	n := NewGradientStrokeNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddTrim appends a default-valued TrimNode (Start=0, End=100, Offset=0 — the
// no-op identity trim) and returns it. A Trim Paths filter reveals only the
// arc of the preceding paths between Start% and End% (offset by Offset
// degrees) — the canonical stroke line-draw / dash-reveal MG primitive. Place
// it AFTER the path-producing shapes it should trim (render order).
func (g *VectorGroup) AddTrim() (*TrimNode, error) {
	n := NewTrimNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddRepeater appends a default-valued RepeaterNode (3 copies, identity
// transform) and returns it. A Repeater duplicates the preceding paths N times,
// applying its Transform (position offset / rotation / scale / opacity falloff)
// cumulatively per copy — the canonical radial-burst / grid MG primitive. Place
// it AFTER the shapes it should duplicate (render order).
func (g *VectorGroup) AddRepeater() (*RepeaterNode, error) {
	n := NewRepeaterNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddRoundCorners appends a default-valued RoundCornersNode (Radius=10, AE's
// default) and returns it. A Round Corners filter rounds the corners of the
// preceding paths by Radius pixels — the canonical "soften the rectangle" MG
// primitive. Place it AFTER the shapes whose corners it should round (render
// order).
func (g *VectorGroup) AddRoundCorners() (*RoundCornersNode, error) {
	n := NewRoundCornersNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddOffsetPaths appends a default-valued OffsetPathsNode (Amount=10, AE's
// default) and returns it. An Offset Paths filter grows (positive) or shrinks
// (negative) the preceding paths by Amount pixels — the canonical outline /
// inflate MG primitive. Place it AFTER the shapes it should offset (render
// order).
func (g *VectorGroup) AddOffsetPaths() (*OffsetPathsNode, error) {
	n := NewOffsetPathsNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddMergePaths appends a default-valued MergePathsNode (Type=Merge) and returns
// it. A Merge Paths filter boolean-combines all the preceding paths in the group
// (Merge / Add / Subtract / Intersect / Exclude) into one path — the canonical
// compound-shape / cut-out MG primitive. Place it AFTER the ≥2 shapes it should
// combine (render order); the result is painted by the fills/strokes.
func (g *VectorGroup) AddMergePaths() (*MergePathsNode, error) {
	n := NewMergePathsNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddZigZag appends a default-valued ZigZagNode (Size=5, Detail=10 — AE's
// defaults) and returns it. A ZigZag filter distorts the preceding paths into a
// zigzag/wave: Size is the amplitude (px), Detail is the number of ridges per
// path segment. Place it AFTER the shapes whose edges it should distort (render
// order).
func (g *VectorGroup) AddZigZag() (*ZigZagNode, error) {
	n := NewZigZagNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddPuckerBloat appends a default-valued PuckerBloatNode (Amount=0, the no-op
// identity) and returns it. Pucker & Bloat bows the preceding paths' edges
// inward (negative Amount = pucker, concave) or outward (positive = bloat,
// convex) — the organic-blob / squish MG primitive. Place it AFTER the shapes it
// should distort (render order).
func (g *VectorGroup) AddPuckerBloat() (*PuckerBloatNode, error) {
	n := NewPuckerBloatNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddTwist appends a default-valued TwistNode (Angle=0, the no-op identity) and
// returns it. Twist rotates the preceding paths progressively — points farther
// from the center rotate more — bowing straight edges into spirals. Place it
// AFTER the shapes it should distort (render order).
func (g *VectorGroup) AddTwist() (*TwistNode, error) {
	n := NewTwistNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddWigglePaths appends a default-valued WigglePathsNode (Size=0, the no-op
// identity) and returns it. Wiggle Paths roughens the preceding paths with
// time-varying random displacement — the classic hand-drawn "boil" jitter.
// Place it AFTER the shapes it should distort (render order).
func (g *VectorGroup) AddWigglePaths() (*WigglePathsNode, error) {
	n := NewWigglePathsNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// AddWiggleTransform appends a default-valued WiggleTransformNode (zero
// amplitudes, the no-op identity) and returns it. Wiggle Transform randomly
// jitters a transform (Anchor/Position/Scale/Rotation) applied to the preceding
// paths over time — typically placed after a Repeater to scatter its copies.
// Place it AFTER the shapes it should affect (render order).
func (g *VectorGroup) AddWiggleTransform() (*WiggleTransformNode, error) {
	n := NewWiggleTransformNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// TrimNode — `ADBE Vector Filter - Trim` (Trim Paths). A path-filter that
// reveals only the portion of the preceding paths between Start% and End%,
// rotated by Offset degrees. Default Start=0, End=100, Offset=0 (identity, no
// trimming). Start/End are percentages (0..100); Offset is in degrees.
//
// `Trim Type` (Simultaneously/Individually) is AE-default (Simultaneously) and
// elided by AE; the serializer materializes the leaf via synthesis-insert when
// SetType selects Individually. Start/End/Offset are static — animated trim (the
// actual line-draw reveal) flips the cdat to a keyframe container via the same
// injectAnimatedStream path as the other shape scalars.
type TrimNode struct {
	start    *codec.PropertyStream[float64]
	end      *codec.PropertyStream[float64]
	offset   *codec.PropertyStream[float64]
	trimType TrimType
}

// TrimType selects how a Trim Paths filter treats multiple paths in its group
// (`ADBE Vector Trim Type`, AE's "Trim Multiple Shapes"). Stored on disk as a
// 1-based float64 enum index.
type TrimType int

const (
	// TrimTypeSimultaneously trims all paths as one combined length (AE default).
	TrimTypeSimultaneously TrimType = 1
	// TrimTypeIndividually trims each path independently to the same Start/End%.
	TrimTypeIndividually TrimType = 2
)

// NewTrimNode constructs a default (identity) TrimNode: Start=0, End=100,
// Offset=0, Type=Simultaneously.
func NewTrimNode() *TrimNode {
	n := &TrimNode{
		start:    codec.NewPropertyStream[float64](),
		end:      codec.NewPropertyStream[float64](),
		offset:   codec.NewPropertyStream[float64](),
		trimType: TrimTypeSimultaneously,
	}
	_ = n.start.SetStaticValue(0)
	_ = n.end.SetStaticValue(100)
	_ = n.offset.SetStaticValue(0)
	return n
}

func (n *TrimNode) Kind() ShapeNodeKind              { return ShapeKindTrim }
func (n *TrimNode) Start() *PropertyStream[float64]  { return n.start }
func (n *TrimNode) End() *PropertyStream[float64]    { return n.end }
func (n *TrimNode) Offset() *PropertyStream[float64] { return n.offset }

// Type returns how the trim treats multiple paths (Simultaneously / Individually).
func (n *TrimNode) Type() TrimType { return n.trimType }

// SetType selects Simultaneously (all paths as one combined length, the default)
// or Individually (each path trimmed to the same Start/End%). The Individually
// value is AE-default-elided; setting it materializes the `ADBE Vector Trim Type`
// leaf on lower. Only visible with multiple paths in the trim's group.
func (n *TrimNode) SetType(t TrimType) error {
	if t != TrimTypeSimultaneously && t != TrimTypeIndividually {
		return fmt.Errorf("TrimNode.SetType: %d out of range (1=Simultaneously, 2=Individually)", t)
	}
	n.trimType = t
	return nil
}

// SetStart sets the trim start percentage (0..100).
func (n *TrimNode) SetStart(v float64) error { return n.start.SetStaticValue(v) }

// SetEnd sets the trim end percentage (0..100).
func (n *TrimNode) SetEnd(v float64) error { return n.end.SetStaticValue(v) }

// SetOffset sets the trim offset in degrees.
func (n *TrimNode) SetOffset(v float64) error { return n.offset.SetStaticValue(v) }

// Properties returns the escape-hatch β view.
func (n *TrimNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Trim Paths",
		streams: map[string]any{
			"Start":  n.start,
			"End":    n.end,
			"Offset": n.offset,
		},
	}
}

// RepeaterNode — `ADBE Vector Filter - Repeater` (Repeater). Duplicates the
// preceding paths `Copies` times, applying `Transform` cumulatively per copy.
// `Copies` and `Offset` (which copy index the first instance starts at) are
// animatable scalars; the Transform (Anchor/Position/Scale/Rotation + Start/End
// Opacity) is static, modeled like StrokeTaper. `Order` (Composite — copies
// above/below) is AE-default (elided) and not modeled.
type RepeaterNode struct {
	copies    *codec.PropertyStream[float64]
	offset    *codec.PropertyStream[float64]
	transform *RepeaterTransform
}

// RepeaterTransform models the Repeater's nested `ADBE Vector Repeater
// Transform` group: the per-copy transform applied cumulatively. Anchor /
// Position / Scale are Vec2 (Scale in %); Rotation is degrees; Start/End
// Opacity are % applied to the first/last copy with a linear falloff between.
// All static (V2.2), stored as plain values like StrokeTaper.
type RepeaterTransform struct {
	anchor       [2]float64
	position     [2]float64
	scale        [2]float64
	rotation     float64
	startOpacity float64
	endOpacity   float64
}

// NewRepeaterNode constructs a default RepeaterNode: 3 copies, offset 0,
// identity transform (anchor [0,0], position [0,0], scale [100,100], rotation
// 0, start/end opacity 100).
func NewRepeaterNode() *RepeaterNode {
	n := &RepeaterNode{
		copies: codec.NewPropertyStream[float64](),
		offset: codec.NewPropertyStream[float64](),
		transform: &RepeaterTransform{
			scale:        [2]float64{100, 100},
			startOpacity: 100,
			endOpacity:   100,
		},
	}
	_ = n.copies.SetStaticValue(3)
	_ = n.offset.SetStaticValue(0)
	return n
}

func (n *RepeaterNode) Kind() ShapeNodeKind              { return ShapeKindRepeater }
func (n *RepeaterNode) Copies() *PropertyStream[float64] { return n.copies }
func (n *RepeaterNode) Offset() *PropertyStream[float64] { return n.offset }

// Transform returns the Repeater's per-copy Transform group.
func (n *RepeaterNode) Transform() *RepeaterTransform { return n.transform }

// SetCopies sets the number of copies (≥ 1).
func (n *RepeaterNode) SetCopies(v float64) error {
	if v < 1 {
		return fmt.Errorf("RepeaterNode.SetCopies: %g out of range (want ≥ 1)", v)
	}
	return n.copies.SetStaticValue(v)
}

// SetOffset sets the copy-index offset of the first instance.
func (n *RepeaterNode) SetOffset(v float64) error { return n.offset.SetStaticValue(v) }

func (t *RepeaterTransform) Anchor() [2]float64    { return t.anchor }
func (t *RepeaterTransform) Position() [2]float64  { return t.position }
func (t *RepeaterTransform) Scale() [2]float64     { return t.scale }
func (t *RepeaterTransform) Rotation() float64     { return t.rotation }
func (t *RepeaterTransform) StartOpacity() float64 { return t.startOpacity }
func (t *RepeaterTransform) EndOpacity() float64   { return t.endOpacity }

// SetAnchor sets the per-copy anchor point (px).
func (t *RepeaterTransform) SetAnchor(v [2]float64) error { t.anchor = v; return nil }

// SetPosition sets the per-copy position offset (px) — the spacing between
// copies. The MG grid/line knob.
func (t *RepeaterTransform) SetPosition(v [2]float64) error { t.position = v; return nil }

// SetScale sets the per-copy scale (%). Cumulative: copy k is scaled k times.
func (t *RepeaterTransform) SetScale(v [2]float64) error { t.scale = v; return nil }

// SetRotation sets the per-copy rotation (degrees) — the radial-burst knob.
func (t *RepeaterTransform) SetRotation(v float64) error { t.rotation = v; return nil }

// SetStartOpacity sets the first copy's opacity (%, 0..100).
func (t *RepeaterTransform) SetStartOpacity(v float64) error {
	if v < 0 || v > 100 {
		return fmt.Errorf("SetStartOpacity: %g out of range 0..100", v)
	}
	t.startOpacity = v
	return nil
}

// SetEndOpacity sets the last copy's opacity (%, 0..100) — falloff to End.
func (t *RepeaterTransform) SetEndOpacity(v float64) error {
	if v < 0 || v > 100 {
		return fmt.Errorf("SetEndOpacity: %g out of range 0..100", v)
	}
	t.endOpacity = v
	return nil
}

// Properties returns the escape-hatch β view.
func (n *RepeaterNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Repeater",
		streams: map[string]any{
			"Copies": n.copies,
			"Offset": n.offset,
		},
	}
}

// RoundCornersNode — `ADBE Vector Filter - RC` (Round Corners). A path-filter
// that rounds the corners of the preceding paths in the stack by `Radius`
// pixels. Its single sub-stream `ADBE Vector RoundCorner Radius` is an
// animatable 1D scalar (default 10, AE's default). Place it AFTER the
// path-producing shapes whose corners it should round (render order).
type RoundCornersNode struct {
	radius *codec.PropertyStream[float64]
}

// NewRoundCornersNode constructs a default RoundCornersNode (Radius=10).
func NewRoundCornersNode() *RoundCornersNode {
	n := &RoundCornersNode{radius: codec.NewPropertyStream[float64]()}
	_ = n.radius.SetStaticValue(10)
	return n
}

func (n *RoundCornersNode) Kind() ShapeNodeKind              { return ShapeKindRoundCorners }
func (n *RoundCornersNode) Radius() *PropertyStream[float64] { return n.radius }

// SetRadius sets the corner radius in pixels. Rejects negative values.
func (n *RoundCornersNode) SetRadius(v float64) error {
	if v < 0 {
		return fmt.Errorf("RoundCornersNode.SetRadius: %g out of range (want >= 0)", v)
	}
	return n.radius.SetStaticValue(v)
}

// Properties returns the escape-hatch β view.
func (n *RoundCornersNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Round Corners",
		streams: map[string]any{
			"Radius": n.radius,
		},
	}
}

// OffsetPathsNode — `ADBE Vector Filter - Offset` (Offset Paths). A path-filter
// that grows (positive Amount) or shrinks (negative Amount) the preceding paths
// in the stack by `Amount` pixels. Its headline sub-stream `ADBE Vector Offset
// Amount` is an animatable 1D scalar (default 10, AE's default). Place it AFTER
// the path-producing shapes it should offset (render order).
//
// Line Join / Miter Limit / Copy Offset are AE-default and elided in the
// extracted Amount-only template (no slot); Copies is materialized on demand by
// the serializer (synthesis-insert of the `ADBE Vector Offset Copies` leaf) when
// SetCopies is called, mirroring SetMaterialOption for default-elided leaves.
type OffsetPathsNode struct {
	amount    *codec.PropertyStream[float64]
	copies    float64
	copiesSet bool
}

// NewOffsetPathsNode constructs a default OffsetPathsNode (Amount=10, Copies=1).
func NewOffsetPathsNode() *OffsetPathsNode {
	n := &OffsetPathsNode{amount: codec.NewPropertyStream[float64](), copies: 1}
	_ = n.amount.SetStaticValue(10)
	return n
}

func (n *OffsetPathsNode) Kind() ShapeNodeKind              { return ShapeKindOffsetPaths }
func (n *OffsetPathsNode) Amount() *PropertyStream[float64] { return n.amount }

// SetAmount sets the offset amount in pixels (positive grows, negative shrinks).
func (n *OffsetPathsNode) SetAmount(v float64) error { return n.amount.SetStaticValue(v) }

// Copies returns the number of progressively-offset copies (default 1).
func (n *OffsetPathsNode) Copies() float64 { return n.copies }

// SetCopies sets the number of copies the Offset Paths filter stacks, each
// offset by a further Amount pixels — N nested outlines growing outward (or
// inward for a negative Amount). Must be >= 1. The `ADBE Vector Offset Copies`
// sub-stream is AE-default-elided; setting it materializes the leaf on lower.
func (n *OffsetPathsNode) SetCopies(v float64) error {
	if v < 1 {
		return fmt.Errorf("OffsetPathsNode.SetCopies: %g out of range (want >= 1)", v)
	}
	n.copies = v
	n.copiesSet = true
	return nil
}

// CopiesSet reports whether SetCopies was called (serializer splice trigger).
func (n *OffsetPathsNode) CopiesSet() bool { return n.copiesSet }

// Properties returns the escape-hatch β view.
func (n *OffsetPathsNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Offset Paths",
		streams: map[string]any{
			"Amount": n.amount,
		},
	}
}

// MergeType is the Merge Paths boolean mode (`ADBE Vector Merge Type`). Stored
// on disk as a 1-based float64 enum index.
type MergeType int

const (
	MergeTypeMerge     MergeType = 1 // default — union all paths, keep overlaps
	MergeTypeAdd       MergeType = 2
	MergeTypeSubtract  MergeType = 3 // upper paths cut holes in the lowest
	MergeTypeIntersect MergeType = 4
	MergeTypeExclude   MergeType = 5 // exclude overlapping regions
)

// MergePathsNode — `ADBE Vector Filter - Merge` (Merge Paths). A path-filter
// that boolean-combines all the paths below it in the group into a single path
// per its `Type`. Its only sub-stream `ADBE Vector Merge Type` is a non-animated
// enum (default Merge); modeled as a plain value like the Fill blend mode.
type MergePathsNode struct {
	mergeType MergeType
}

// NewMergePathsNode constructs a default MergePathsNode (Type=Merge).
func NewMergePathsNode() *MergePathsNode {
	return &MergePathsNode{mergeType: MergeTypeMerge}
}

func (n *MergePathsNode) Kind() ShapeNodeKind { return ShapeKindMergePaths }
func (n *MergePathsNode) Type() MergeType     { return n.mergeType }

// SetType sets the boolean merge mode. Rejects values outside 1..5.
func (n *MergePathsNode) SetType(v MergeType) error {
	if v < MergeTypeMerge || v > MergeTypeExclude {
		return fmt.Errorf("MergePathsNode.SetType: invalid value %d (want 1..5)", v)
	}
	n.mergeType = v
	return nil
}

// Properties returns the escape-hatch β view.
func (n *MergePathsNode) Properties() *PropertyGroup {
	return &PropertyGroup{Name: "Merge Paths"}
}

// ZigZagNode — `ADBE Vector Filter - Zigzag` (ZigZag). A path-filter that
// distorts the paths below it in the stack into a zigzag/wave. `Size` (amplitude,
// px) and `Detail` (ridges per path segment) are animatable 1D scalars (defaults
// 5 / 10, AE's defaults). `Points` (Corner/Smooth) selects whether each ridge is
// a sharp sawtooth corner (AE default) or a smooth scalloped wave; the non-default
// Smooth value is AE-default-elided and materialized by the serializer via
// synthesis-insert when SetPoints selects it.
type ZigZagNode struct {
	size   *codec.PropertyStream[float64]
	detail *codec.PropertyStream[float64]
	points ZigZagPoints
}

// ZigZagPoints selects whether a ZigZag filter's ridges are sharp corners or
// smooth scalloped waves (`ADBE Vector Zigzag Points`, AE's "Points" dropdown).
// Stored on disk as a 1-based float64 enum index.
type ZigZagPoints int

const (
	// ZigZagPointsCorner makes each ridge a sharp sawtooth corner (AE default).
	ZigZagPointsCorner ZigZagPoints = 1
	// ZigZagPointsSmooth makes each ridge a smooth scalloped wave.
	ZigZagPointsSmooth ZigZagPoints = 2
)

// NewZigZagNode constructs a default ZigZagNode (Size=5, Detail=10, Points=Corner).
func NewZigZagNode() *ZigZagNode {
	n := &ZigZagNode{
		size:   codec.NewPropertyStream[float64](),
		detail: codec.NewPropertyStream[float64](),
		points: ZigZagPointsCorner,
	}
	_ = n.size.SetStaticValue(5)
	_ = n.detail.SetStaticValue(10)
	return n
}

func (n *ZigZagNode) Kind() ShapeNodeKind              { return ShapeKindZigZag }
func (n *ZigZagNode) Size() *PropertyStream[float64]   { return n.size }
func (n *ZigZagNode) Detail() *PropertyStream[float64] { return n.detail }

// SetSize sets the zigzag amplitude in pixels. Rejects negative values.
func (n *ZigZagNode) SetSize(v float64) error {
	if v < 0 {
		return fmt.Errorf("ZigZagNode.SetSize: %g out of range (want >= 0)", v)
	}
	return n.size.SetStaticValue(v)
}

// SetDetail sets the number of ridges per path segment. Rejects negative values.
func (n *ZigZagNode) SetDetail(v float64) error {
	if v < 0 {
		return fmt.Errorf("ZigZagNode.SetDetail: %g out of range (want >= 0)", v)
	}
	return n.detail.SetStaticValue(v)
}

// Points returns whether the ridges are sharp corners or smooth waves.
func (n *ZigZagNode) Points() ZigZagPoints { return n.points }

// SetPoints selects Corner (sharp sawtooth ridges, the default) or Smooth
// (scalloped wave ridges). The Smooth value is AE-default-elided; setting it
// materializes the `ADBE Vector Zigzag Points` leaf on lower.
func (n *ZigZagNode) SetPoints(p ZigZagPoints) error {
	if p != ZigZagPointsCorner && p != ZigZagPointsSmooth {
		return fmt.Errorf("ZigZagNode.SetPoints: %d out of range (1=Corner, 2=Smooth)", p)
	}
	n.points = p
	return nil
}

// Properties returns the escape-hatch β view.
func (n *ZigZagNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "ZigZag",
		streams: map[string]any{
			"Size":   n.size,
			"Detail": n.detail,
		},
	}
}

// PuckerBloatNode — `ADBE Vector Filter - PB` (Pucker & Bloat). A path-filter
// that bows the paths below it inward (negative Amount = pucker) or outward
// (positive = bloat). Its single sub-stream `ADBE Vector PuckerBloat Amount` is
// an animatable 1D scalar (percent; default 0 = no distortion). Place it AFTER
// the path-producing shapes it should distort (render order).
type PuckerBloatNode struct {
	amount *codec.PropertyStream[float64]
}

// NewPuckerBloatNode constructs a default PuckerBloatNode (Amount=0, identity).
func NewPuckerBloatNode() *PuckerBloatNode {
	n := &PuckerBloatNode{amount: codec.NewPropertyStream[float64]()}
	_ = n.amount.SetStaticValue(0)
	return n
}

func (n *PuckerBloatNode) Kind() ShapeNodeKind              { return ShapeKindPuckerBloat }
func (n *PuckerBloatNode) Amount() *PropertyStream[float64] { return n.amount }

// SetAmount sets the pucker/bloat amount (percent; negative puckers/concave,
// positive bloats/convex).
func (n *PuckerBloatNode) SetAmount(v float64) error { return n.amount.SetStaticValue(v) }

// Properties returns the escape-hatch β view.
func (n *PuckerBloatNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Pucker & Bloat",
		streams: map[string]any{
			"Amount": n.amount,
		},
	}
}

// TwistNode — `ADBE Vector Filter - Twist`. A path-filter that rotates the
// paths below it progressively (more rotation farther from the twist center),
// bowing straight edges into spirals. Its headline sub-stream `ADBE Vector
// Twist Angle` is an animatable 1D scalar (degrees; default 0 = no twist).
// `Center` (Vec2, shape-local pixels relative to the path centre) is the pivot
// the twist rotates around; it defaults to [0,0] and is AE-default-elided,
// materialized by the serializer via synthesis-insert when SetCenter offsets it.
// Place it AFTER the path-producing shapes it should distort (render order).
type TwistNode struct {
	angle     *codec.PropertyStream[float64]
	center    [2]float64
	centerSet bool
}

// NewTwistNode constructs a default TwistNode (Angle=0, identity; Center [0,0]).
func NewTwistNode() *TwistNode {
	n := &TwistNode{angle: codec.NewPropertyStream[float64]()}
	_ = n.angle.SetStaticValue(0)
	return n
}

func (n *TwistNode) Kind() ShapeNodeKind             { return ShapeKindTwist }
func (n *TwistNode) Angle() *PropertyStream[float64] { return n.angle }

// SetAngle sets the twist angle (degrees; positive twists clockwise, negative
// counter-clockwise).
func (n *TwistNode) SetAngle(v float64) error { return n.angle.SetStaticValue(v) }

// Center returns the twist pivot (Vec2, shape-local pixels relative to the path
// centre). Default [0,0] (rotate about the path centre).
func (n *TwistNode) Center() [2]float64 { return n.center }

// CenterSet reports whether SetCenter has been called (so the serializer knows
// to materialize the otherwise-elided Center leaf).
func (n *TwistNode) CenterSet() bool { return n.centerSet }

// SetCenter offsets the twist pivot from the path centre (Vec2, shape-local
// pixels). The default [0,0] is AE-default-elided; a non-zero Center
// materializes the `ADBE Vector Twist Center` leaf on lower.
func (n *TwistNode) SetCenter(c [2]float64) error {
	n.center = c
	n.centerSet = true
	return nil
}

// Properties returns the escape-hatch β view.
func (n *TwistNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Twist",
		streams: map[string]any{
			"Angle": n.angle,
		},
	}
}

// WigglePathsNode — `ADBE Vector Filter - Roughen` (Wiggle Paths). A path-filter
// that roughens the paths below it with time-varying random displacement — the
// hand-drawn "boil" jitter. Four animatable 1D scalar sub-streams are modeled:
// Size (displacement amplitude), Detail (segments per unit length),
// WigglesPerSecond (the temporal frequency, `ADBE Vector Temporal Freq`), and
// RandomSeed. The Points (enum), Correlation, and Temporal/Spatial Phase
// sub-streams are left at their defaults and elided. Place it AFTER the
// path-producing shapes it should distort (render order).
type WigglePathsNode struct {
	size             *codec.PropertyStream[float64]
	detail           *codec.PropertyStream[float64]
	wigglesPerSecond *codec.PropertyStream[float64]
	randomSeed       *codec.PropertyStream[float64]
}

// NewWigglePathsNode constructs a default WigglePathsNode (Size=0 → no
// displacement, the identity; Detail=10, WigglesPerSecond=2, RandomSeed=0 match
// AE's filter defaults).
func NewWigglePathsNode() *WigglePathsNode {
	n := &WigglePathsNode{
		size:             codec.NewPropertyStream[float64](),
		detail:           codec.NewPropertyStream[float64](),
		wigglesPerSecond: codec.NewPropertyStream[float64](),
		randomSeed:       codec.NewPropertyStream[float64](),
	}
	_ = n.size.SetStaticValue(0)
	_ = n.detail.SetStaticValue(10)
	_ = n.wigglesPerSecond.SetStaticValue(2)
	_ = n.randomSeed.SetStaticValue(0)
	return n
}

func (n *WigglePathsNode) Kind() ShapeNodeKind                        { return ShapeKindWigglePaths }
func (n *WigglePathsNode) Size() *PropertyStream[float64]             { return n.size }
func (n *WigglePathsNode) Detail() *PropertyStream[float64]           { return n.detail }
func (n *WigglePathsNode) WigglesPerSecond() *PropertyStream[float64] { return n.wigglesPerSecond }
func (n *WigglePathsNode) RandomSeed() *PropertyStream[float64]       { return n.randomSeed }

// SetSize sets the wiggle displacement amplitude (pixels; 0 = no roughening).
func (n *WigglePathsNode) SetSize(v float64) error { return n.size.SetStaticValue(v) }

// SetDetail sets the wiggle detail (number of segments per path length — higher
// = finer, more frequent ridges).
func (n *WigglePathsNode) SetDetail(v float64) error { return n.detail.SetStaticValue(v) }

// SetWigglesPerSecond sets the temporal frequency (`ADBE Vector Temporal Freq`,
// the Wiggles/Second control — how fast the random edge churns over time).
func (n *WigglePathsNode) SetWigglesPerSecond(v float64) error {
	return n.wigglesPerSecond.SetStaticValue(v)
}

// SetRandomSeed sets the random seed selecting the displacement pattern.
func (n *WigglePathsNode) SetRandomSeed(v float64) error { return n.randomSeed.SetStaticValue(v) }

// Properties returns the escape-hatch β view.
func (n *WigglePathsNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Wiggle Paths",
		streams: map[string]any{
			"Size":             n.size,
			"Detail":           n.detail,
			"WigglesPerSecond": n.wigglesPerSecond,
			"RandomSeed":       n.randomSeed,
		},
	}
}

// WiggleTransformNode — `ADBE Vector Filter - Wiggler` (Wiggle Transform). A
// filter that randomly jitters a transform applied to the paths below it over
// time — usually placed after a Repeater to scatter its copies. WigglesPerSecond
// (the temporal frequency, `ADBE Vector Xform Temporal Freq`) and RandomSeed are
// animatable 1D scalars; the per-channel wiggle amplitudes live in the nested
// Transform group (Anchor/Position/Scale Vec2, Rotation degrees) modeled
// statically like the Repeater Transform. Correlation and Temporal/Spatial Phase
// are left at their defaults and elided.
type WiggleTransformNode struct {
	wigglesPerSecond *codec.PropertyStream[float64]
	randomSeed       *codec.PropertyStream[float64]
	transform        *WigglerTransform
}

// WigglerTransform models the Wiggle Transform's nested `ADBE Vector Wiggler
// Transform` group: the per-channel random wiggle AMPLITUDES (not absolute
// transform values). Anchor / Position / Scale are Vec2 (Scale amplitude in %),
// Rotation in degrees. A zero amplitude means that channel does not wiggle. All
// static (V2.2), stored as plain values like RepeaterTransform.
type WigglerTransform struct {
	anchor   [2]float64
	position [2]float64
	scale    [2]float64
	rotation float64
}

// NewWiggleTransformNode constructs a default WiggleTransformNode: all wiggle
// amplitudes zero (the identity — nothing wiggles), WigglesPerSecond=2 and
// RandomSeed=0 matching AE's filter defaults.
func NewWiggleTransformNode() *WiggleTransformNode {
	n := &WiggleTransformNode{
		wigglesPerSecond: codec.NewPropertyStream[float64](),
		randomSeed:       codec.NewPropertyStream[float64](),
		transform:        &WigglerTransform{},
	}
	_ = n.wigglesPerSecond.SetStaticValue(2)
	_ = n.randomSeed.SetStaticValue(0)
	return n
}

func (n *WiggleTransformNode) Kind() ShapeNodeKind { return ShapeKindWiggleTransform }

// WigglesPerSecond returns the temporal-frequency stream (`ADBE Vector Xform
// Temporal Freq`).
func (n *WiggleTransformNode) WigglesPerSecond() *PropertyStream[float64] {
	return n.wigglesPerSecond
}

// RandomSeed returns the random-seed stream.
func (n *WiggleTransformNode) RandomSeed() *PropertyStream[float64] { return n.randomSeed }

// Transform returns the per-channel wiggle-amplitude group.
func (n *WiggleTransformNode) Transform() *WigglerTransform { return n.transform }

// SetWigglesPerSecond sets the temporal frequency (how fast the transform churns).
func (n *WiggleTransformNode) SetWigglesPerSecond(v float64) error {
	return n.wigglesPerSecond.SetStaticValue(v)
}

// SetRandomSeed sets the random seed selecting the wiggle pattern.
func (n *WiggleTransformNode) SetRandomSeed(v float64) error { return n.randomSeed.SetStaticValue(v) }

// Anchor / Position / Scale / Rotation return the current wiggle amplitudes.
func (t *WigglerTransform) Anchor() [2]float64   { return t.anchor }
func (t *WigglerTransform) Position() [2]float64 { return t.position }
func (t *WigglerTransform) Scale() [2]float64    { return t.scale }
func (t *WigglerTransform) Rotation() float64    { return t.rotation }

// SetAnchor sets the anchor-point wiggle amplitude (pixels).
func (t *WigglerTransform) SetAnchor(v [2]float64) error { t.anchor = v; return nil }

// SetPosition sets the position wiggle amplitude (pixels).
func (t *WigglerTransform) SetPosition(v [2]float64) error { t.position = v; return nil }

// SetScale sets the scale wiggle amplitude (percent).
func (t *WigglerTransform) SetScale(v [2]float64) error { t.scale = v; return nil }

// SetRotation sets the rotation wiggle amplitude (degrees).
func (t *WigglerTransform) SetRotation(v float64) error { t.rotation = v; return nil }

// Properties returns the escape-hatch β view.
func (n *WiggleTransformNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Wiggle Transform",
		streams: map[string]any{
			"WigglesPerSecond": n.wigglesPerSecond,
			"RandomSeed":       n.randomSeed,
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
	color   *codec.PropertyStream[[4]float64]
	opacity *codec.PropertyStream[float64]
	width   *codec.PropertyStream[float64]

	// Line Cap / Line Join are enums; Miter Limit is a scalar. AE does not
	// animate them, so they are plain values rather than PropertyStreams.
	lineCap    StrokeLineCap
	lineJoin   StrokeLineJoin
	miterLimit float64

	blendMode      ShapeBlendMode
	compositeOrder ShapeCompositeOrder

	taper  *StrokeTaper
	wave   *StrokeWave
	dashes *StrokeDashes
}

// NewStrokeNode constructs a default-valued StrokeNode.
func NewStrokeNode() *StrokeNode {
	s := &StrokeNode{
		color:          codec.NewPropertyStream[[4]float64](),
		opacity:        codec.NewPropertyStream[float64](),
		width:          codec.NewPropertyStream[float64](),
		lineCap:        StrokeLineCapButt,
		lineJoin:       StrokeLineJoinMiter,
		miterLimit:     4,
		blendMode:      ShapeBlendModeNormal,
		compositeOrder: ShapeCompositeOrderAbovePrevious,
		taper:          newStrokeTaper(),
		wave:           newStrokeWave(),
		dashes:         newStrokeDashes(),
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

func (s *StrokeNode) LineCap() StrokeLineCap              { return s.lineCap }
func (s *StrokeNode) LineJoin() StrokeLineJoin            { return s.lineJoin }
func (s *StrokeNode) MiterLimit() float64                 { return s.miterLimit }
func (s *StrokeNode) BlendMode() ShapeBlendMode           { return s.blendMode }
func (s *StrokeNode) CompositeOrder() ShapeCompositeOrder { return s.compositeOrder }
func (s *StrokeNode) SetBlendMode(v ShapeBlendMode) error { return setShapeBlendMode(&s.blendMode, v) }
func (s *StrokeNode) SetCompositeOrder(v ShapeCompositeOrder) error {
	return setShapeCompositeOrder(&s.compositeOrder, v)
}

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

// Taper returns the stroke's Taper group (`ADBE Vector Stroke Taper`).
func (s *StrokeNode) Taper() *StrokeTaper { return s.taper }

// Wave returns the stroke's Wave group (`ADBE Vector Stroke Wave`).
func (s *StrokeNode) Wave() *StrokeWave { return s.wave }

// Dashes returns the stroke's Dashes group (`ADBE Vector Stroke Dashes`).
func (s *StrokeNode) Dashes() *StrokeDashes { return s.dashes }

// StrokeTaper models the Stroke "Taper" group's %-mode scalar controls
// (`ADBE Vector Stroke Taper`): Start/End Length, Start/End Width, Start/End
// Ease — all plain float64 percentages stored on disk as float64 BE at
// cdat[0:8]. AE does not animate them in V2.2, so they are stored as values,
// not PropertyStreams. All default to 0 (no taper).
//
// V2.2 supports only the always-active %-mode controls. The Length Units enum
// and the pixel-mode mirror streams (StartWidthPx/EndWidthPx) are AE-elided at
// the % default and not modeled — see lowerStrokeNode limitations.
type StrokeTaper struct {
	startLength, endLength float64
	startWidth, endWidth   float64
	startEase, endEase     float64
}

func newStrokeTaper() *StrokeTaper { return &StrokeTaper{} }

func (t *StrokeTaper) StartLength() float64 { return t.startLength }
func (t *StrokeTaper) EndLength() float64   { return t.endLength }
func (t *StrokeTaper) StartWidth() float64  { return t.startWidth }
func (t *StrokeTaper) EndWidth() float64    { return t.endWidth }
func (t *StrokeTaper) StartEase() float64   { return t.startEase }
func (t *StrokeTaper) EndEase() float64     { return t.endEase }

func (t *StrokeTaper) SetStartLength(v float64) error { t.startLength = v; return nil }
func (t *StrokeTaper) SetEndLength(v float64) error   { t.endLength = v; return nil }
func (t *StrokeTaper) SetStartWidth(v float64) error  { t.startWidth = v; return nil }
func (t *StrokeTaper) SetEndWidth(v float64) error    { t.endWidth = v; return nil }
func (t *StrokeTaper) SetStartEase(v float64) error   { t.startEase = v; return nil }
func (t *StrokeTaper) SetEndEase(v float64) error     { t.endEase = v; return nil }

// StrokeWave models the Stroke "Wave" group's Wavelength-mode scalars
// (`ADBE Vector Stroke Wave`): Amount (%), Wavelength (px), Phase (deg) — stored
// on disk as float64 BE at cdat[0:8]. Defaults: Amount 0, Wavelength 100,
// Phase 0.
//
// V2.2 supports only the Wavelength-mode controls. The Units enum (Wavelength
// vs Cycles) and the Cycles stream are AE-elided at the Wavelength default and
// not modeled — Wave is always emitted in Wavelength mode.
type StrokeWave struct {
	amount     float64
	wavelength float64
	phase      float64
}

func newStrokeWave() *StrokeWave { return &StrokeWave{wavelength: 100} }

func (w *StrokeWave) Amount() float64     { return w.amount }
func (w *StrokeWave) Wavelength() float64 { return w.wavelength }
func (w *StrokeWave) Phase() float64      { return w.phase }

func (w *StrokeWave) SetAmount(v float64) error     { w.amount = v; return nil }
func (w *StrokeWave) SetWavelength(v float64) error { w.wavelength = v; return nil }
func (w *StrokeWave) SetPhase(v float64) error      { w.phase = v; return nil }

// StrokeDashes models the Stroke "Dashes" group (`ADBE Vector Stroke Dashes`):
// a single Dash + Gap pair, stored on disk as float64 BE at cdat[0:8] —
// identical encoding to the other stroke scalars, nested one level deeper inside
// the group's LIST(tdgp). The group is hidden-by-default: a default stroke emits
// the Dashes group as an empty placeholder (solid line). When Enabled, the
// serializer swaps to a dashed stroke-body template that carries the Dash 1 /
// Gap 1 slots and overwrites them with these values.
//
// V2.2 models exactly one Dash + Gap pair (the common dashed/dotted-line case).
// Deferred: additional Dash 2/3 + Gap 2/3 pairs (AE emits only enabled pairs,
// each pair is a separate template variant), and Offset — which AE keeps hidden
// until a dash is enabled and refuses to set via ScriptingAPI, so no template
// can carry its slot. Dash/Gap are not animated in V2.2.
type StrokeDashes struct {
	enabled   bool
	dash, gap float64
}

// newStrokeDashes returns a disabled Dashes group with AE's default Dash/Gap
// of 10 (the values AE shows when you first add a dash).
func newStrokeDashes() *StrokeDashes { return &StrokeDashes{dash: 10, gap: 10} }

func (d *StrokeDashes) Enabled() bool { return d.enabled }
func (d *StrokeDashes) Dash() float64 { return d.dash }
func (d *StrokeDashes) Gap() float64  { return d.gap }

// Enable turns dashing on (the serializer emits the Dash + Gap slots). Disable
// reverts to a solid stroke.
func (d *StrokeDashes) Enable()  { d.enabled = true }
func (d *StrokeDashes) Disable() { d.enabled = false }

// SetDash sets the dash length and enables dashing. Rejects negative values.
func (d *StrokeDashes) SetDash(v float64) error {
	if v < 0 {
		return fmt.Errorf("StrokeDashes.SetDash: %g out of range (want >= 0)", v)
	}
	d.dash = v
	d.enabled = true
	return nil
}

// SetGap sets the gap length and enables dashing. Rejects negative values.
func (d *StrokeDashes) SetGap(v float64) error {
	if v < 0 {
		return fmt.Errorf("StrokeDashes.SetGap: %g out of range (want >= 0)", v)
	}
	d.gap = v
	d.enabled = true
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

// Float64Stream returns the PropertyStream[float64] under the given name,
// or an error if no stream by that name exists or it isn't the expected type.
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

// ColorStream returns the PropertyStream[[4]float64] (RGBA) under the
// given name.
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

// PathStream returns the PropertyStream[BezierPath] under the given
// name.
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
