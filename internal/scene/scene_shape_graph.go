// internal/aep/shape_graph.go
package scene

import (
	"fmt"

	"github.com/example/aep-parser/internal/codec"
)

// ShapeNodeKind identifies a shape-graph node's runtime kind. AE match-name
// strings (`ADBE Vector Shape - Rect` etc.) are intentionally NOT exposed
// here — they belong to the serializer, which maps Kind → match name during
// lowering.
type ShapeNodeKind int

const (
	ShapeKindRect            ShapeNodeKind = iota // `ADBE Vector Shape - Rect`
	ShapeKindEllipse                              // `ADBE Vector Shape - Ellipse`
	ShapeKindPath                                 // `ADBE Vector Shape - Group`
	ShapeKindFill                                 // `ADBE Vector Graphic - Fill`
	ShapeKindStroke                               // `ADBE Vector Graphic - Stroke`
	ShapeKindGroup                                // `ADBE Vector Group` (user-created nested group)
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
)

// ShapeNode is the runtime-facing shape-graph node interface. All concrete
// node types (RectNode / EllipseNode / PathNode / FillNode / StrokeNode and
// the container VectorGroup) satisfy it.
type ShapeNode interface {
	Kind() ShapeNodeKind
	Properties() *PropertyGroup // escape hatch; currently returns nil
}

// VectorGroup is the shape-graph container node. Every ShapeLayer carries
// one default RootGroup. Children render in order: Children[0] = bottom;
// Children[len-1] = top / most recently appended.
//
// Transform is the group-level Transform PropertyGroup placeholder. A
// default-serialized group has no `ADBE Vector Transform Group`; the field
// exists for the escape-hatch / hydration path on user-created nested groups.
type VectorGroup struct {
	Children  []ShapeNode
	Transform *PropertyGroup
}

// NewVectorGroup returns an empty group with an identity Transform placeholder.
func NewVectorGroup() *VectorGroup {
	return &VectorGroup{Transform: newGroupTransform()}
}

// @summary    Append a rectangle node to a vector group
// @description The new node is placed at the top of the render stack
//   (Children[len-1]). Call its setters to give it geometry.
// @returns    the created rectangle node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTrim_AEShipGate_AE2020,TestMGTrim_AEShipGate_AE2025
// @since      AE2020
// @alias      add rect,矩形,新增矩形节点
func (g *VectorGroup) AddRect() (*RectNode, error) {
	r := NewRectNode()
	g.Children = append(g.Children, r)
	return r, nil
}

// @summary    Append an ellipse node to a vector group
// @returns    the created ellipse node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Ellipse_AEShipGate_AE2020,TestV2_2_Ellipse_AEShipGate_AE2025
// @since      AE2020
// @alias      add ellipse,椭圆,新增椭圆节点
func (g *VectorGroup) AddEllipse() (*EllipseNode, error) {
	e := NewEllipseNode()
	g.Children = append(g.Children, e)
	return e, nil
}

// @summary    Append an empty closed path node to a vector group
// @description Caller must set vertices to give it geometry (minimum two
//   vertices).
// @returns    the created path node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPentagonPath_AEShipGate_AE2020,TestMGPentagonPath_AEShipGate_AE2025
// @since      AE2020
// @alias      add path,路径,新增路径节点,bezier
func (g *VectorGroup) AddPath() (*PathNode, error) {
	p := NewPathNode()
	g.Children = append(g.Children, p)
	return p, nil
}

// @summary    Append a star node to a vector group
// @description Defaults to a 5-point star (outer radius 100, inner radius 50).
//   The Star type is modeled; Polygon is selected via the type setter.
// @returns    the created star node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGStar_AEShipGate_AE2020,TestMGStar_AEShipGate_AE2025
// @since      AE2020
// @alias      add star,星形,多边形,新增星形节点,polystar
func (g *VectorGroup) AddStar() (*StarNode, error) {
	n := NewStarNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a fill node to a vector group
// @description Defaults to white at 100% opacity.
// @returns    the created fill node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Ellipse_AEShipGate_AE2020,TestV2_2_Ellipse_AEShipGate_AE2025
// @since      AE2020
// @alias      add fill,填充,新增填充节点,solid fill
func (g *VectorGroup) AddFill() (*FillNode, error) {
	f := NewFillNode()
	g.Children = append(g.Children, f)
	return f, nil
}

// @summary    Append a stroke node to a vector group
// @description Defaults to black, width 2, 100% opacity.
// @returns    the created stroke node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Stroke_AEShipGate_AE2020,TestV2_2_Stroke_AEShipGate_AE2025
// @since      AE2020
// @alias      add stroke,描边,新增描边节点,outline
func (g *VectorGroup) AddStroke() (*StrokeNode, error) {
	s := NewStrokeNode()
	g.Children = append(g.Children, s)
	return s, nil
}

// @summary    Append a gradient fill node to a vector group
// @description Defaults to a 2-stop black-to-white linear gradient, fully
//   opaque. Set the stops via the color and alpha stop setters.
// @returns    the created gradient fill node
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_GradientFill_AEShipGate_AE2020,TestV2_2_GradientFill_AEShipGate_AE2025
// @since      AE2020
// @alias      add gradient fill,渐变填充,新增渐变填充节点
func (g *VectorGroup) AddGradientFill() (*GradientFillNode, error) {
	n := NewGradientFillNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// BezierPath is the runtime geometry object. Vertices are the path control
// points; InTangents / OutTangents are the per-vertex bezier tangent offsets
// (zero = linear segment); Closed distinguishes closed shapes from open
// polylines.
type BezierPath struct {
	Vertices    [][2]float64
	InTangents  [][2]float64
	OutTangents [][2]float64
	Closed      bool
}

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

// RectNode is a `ADBE Vector Shape - Rect`. Default Size=[100,100],
// Position=[0,0], Roundness=0 (AE elides all three at default).
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

// @summary    Set the rectangle size
// @param      v  the width and height in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTrim_AEShipGate_AE2020,TestMGTrim_AEShipGate_AE2025
// @since      AE2020
// @alias      rect size,矩形尺寸,set size,宽高
func (r *RectNode) SetSize(v [2]float64) error { return r.size.SetStaticValue(v) }

// @summary    Set the rectangle position
// @param      v  the local position offset in pixels
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      rect position,矩形位置,set position
func (r *RectNode) SetPosition(v [2]float64) error { return r.position.SetStaticValue(v) }

// @summary    Set the rectangle corner roundness
// @param      v  the corner radius in pixels
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      rect roundness,矩形圆角,set roundness
func (r *RectNode) SetRoundness(v float64) error { return r.roundness.SetStaticValue(v) }

// @summary    Set the rectangle path direction
// @param      v  the path direction (Normal or Reversed)
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestV2_2_ShapeEnums_AEShipGate_AE2020,TestV2_2_ShapeEnums_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts Reversed and it survives a re-save
// @alias      rect direction,路径方向,shape direction,正向反向
func (r *RectNode) SetDirection(v ShapeDirection) error { return setShapeDirection(&r.direction, v) }

// Properties returns the escape-hatch view onto this RectNode's streams.
// Streams returned via PropertyGroup accessors are the same instances as the
// typed accessors — mutating one reflects through the other.
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

// EllipseNode is a `ADBE Vector Shape - Ellipse`. Default Size=[100,100],
// Position=[0,0]. AE child[1] is `ADBE Vector Shape Direction`, runtime-default
// Normal.
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

// @summary    Set the ellipse size
// @param      v  the width and height in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Ellipse_AEShipGate_AE2020,TestV2_2_Ellipse_AEShipGate_AE2025
// @since      AE2020
// @alias      ellipse size,椭圆尺寸,set size,宽高
func (e *EllipseNode) SetSize(v [2]float64) error { return e.size.SetStaticValue(v) }

// @summary    Set the ellipse position
// @param      v  the local position offset in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Ellipse_AEShipGate_AE2020,TestV2_2_Ellipse_AEShipGate_AE2025
// @since      AE2020
// @alias      ellipse position,椭圆位置,set position
func (e *EllipseNode) SetPosition(v [2]float64) error { return e.position.SetStaticValue(v) }

// @summary    Set the ellipse path direction
// @param      v  the path direction (Normal or Reversed)
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestV2_2_ShapeEnums_AEShipGate_AE2020,TestV2_2_ShapeEnums_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the direction and it survives a re-save
// @alias      ellipse direction,椭圆方向,path direction,正向反向
func (e *EllipseNode) SetDirection(v ShapeDirection) error { return setShapeDirection(&e.direction, v) }

// setShapeDirection validates and assigns a ShapeDirection (Normal=1 or
// Reversed=3; AE has no value 2 for parametric shapes).
func setShapeDirection(dst *ShapeDirection, v ShapeDirection) error {
	if v != ShapeDirectionNormal && v != ShapeDirectionReversed {
		return fmt.Errorf("SetDirection: invalid value %d (want 1=Normal or 3=Reversed)", v)
	}
	*dst = v
	return nil
}

// Properties returns the escape-hatch view.
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

// StarNode is a `ADBE Vector Shape - Star`: a parametric polystar, either a
// Star (alternating outer/inner points) or a Polygon (convex N-gon) per
// StarType. Defaults match AE: Type=Star, Points=5, Position=[0,0], Rotation=0,
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

// @summary    Select the polystar shape (Star or Polygon)
// @description For a Polygon the Inner Radius and Inner Roundness are ignored.
// @param      t  the polystar shape (Star or Polygon)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPolygon_AEShipGate_AE2020,TestMGPolygon_AEShipGate_AE2025
// @since      AE2020
// @alias      star type,星形类型,polygon,多边形,polystar type
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

// @summary    Set the number of star points
// @param      v  the point count (must be at least three)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGStar_AEShipGate_AE2020,TestMGStar_AEShipGate_AE2025
// @since      AE2020
// @alias      star points,星形点数,角数,polystar points
func (n *StarNode) SetPoints(v float64) error {
	if v < 3 {
		return fmt.Errorf("StarNode.SetPoints: %g out of range (want >= 3)", v)
	}
	return n.points.SetStaticValue(v)
}

// @summary    Set the star local position offset
// @param      v  the local position offset in pixels
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      star position,星形位置,polystar position
func (n *StarNode) SetPosition(v [2]float64) error { return n.position.SetStaticValue(v) }

// @summary    Set the star rotation in degrees
// @param      v  the rotation angle in degrees
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPolygon_AEShipGate_AE2020,TestMGPolygon_AEShipGate_AE2025
// @since      AE2020
// @alias      star rotation,星形旋转,polystar rotation,旋转角度
func (n *StarNode) SetRotation(v float64) error { return n.rotation.SetStaticValue(v) }

// @summary    Set the star inner radius
// @description The inner radius is the valley between points.
// @param      v  the inner radius in pixels (must be non-negative)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGStar_AEShipGate_AE2020,TestMGStar_AEShipGate_AE2025
// @since      AE2020
// @alias      star inner radius,星形内半径,inner radius
func (n *StarNode) SetInnerRadius(v float64) error {
	if v < 0 {
		return fmt.Errorf("StarNode.SetInnerRadius: %g out of range (want >= 0)", v)
	}
	return n.innerRadius.SetStaticValue(v)
}

// @summary    Set the star outer radius
// @description The outer radius is the distance to the star tips.
// @param      v  the outer radius in pixels (must be non-negative)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGStar_AEShipGate_AE2020,TestMGStar_AEShipGate_AE2025
// @since      AE2020
// @alias      star outer radius,星形外半径,outer radius
func (n *StarNode) SetOuterRadius(v float64) error {
	if v < 0 {
		return fmt.Errorf("StarNode.SetOuterRadius: %g out of range (want >= 0)", v)
	}
	return n.outerRadius.SetStaticValue(v)
}

// @summary    Set the star inner-point roundness
// @param      v  the inner-point roundness in percent
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      star inner roundness,星形内圆角,inner roundness
func (n *StarNode) SetInnerRoundness(v float64) error { return n.innerRoundness.SetStaticValue(v) }

// @summary    Set the star outer-point roundness
// @param      v  the outer-point roundness in percent
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      star outer roundness,星形外圆角,outer roundness
func (n *StarNode) SetOuterRoundness(v float64) error { return n.outerRoundness.SetStaticValue(v) }

// Properties returns the escape-hatch view.
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

// PathNode is a `ADBE Vector Shape - Group`. Default = empty Vertices,
// Closed=true. The vertex setter builds linear segments (tangents=0); a minimum
// of two vertices is enforced at runtime as a conservative invariant.
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

// Properties returns the escape-hatch view.
func (p *PathNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Path",
		streams: map[string]any{
			"Path": p.path,
		},
	}
}

// @summary    Replace the path vertices with linear segments
// @description Tangents are zeroed and the current Closed flag is preserved.
//   Requires at least two vertices.
// @param      verts  the ordered list of vertices in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPentagonPath_AEShipGate_AE2020,TestMGPentagonPath_AEShipGate_AE2025
// @since      AE2020
// @alias      set vertices,顶点,路径顶点,bezier path
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

// @summary    Toggle whether the path is closed
// @description Vertices and tangents are left undisturbed.
// @param      closed  true to close the path, false to leave it open
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPentagonPath_AEShipGate_AE2020,TestMGPentagonPath_AEShipGate_AE2025
// @since      AE2020
// @alias      set closed,路径闭合,closed path,开放路径
func (p *PathNode) SetClosed(closed bool) error {
	current, _ := p.path.StaticValue()
	current.Closed = closed
	return p.path.SetStaticValue(current)
}

// FillNode is a `ADBE Vector Graphic - Fill`. Default Color=[1,1,1,1] white,
// Opacity=100, Blend Mode=Normal, Composite Order=Above Previous, Fill
// Rule=Nonzero Winding (AE elides at default).
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

// @summary    Set the fill color
// @param      v  the RGBA color, each channel in zero to one
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Ellipse_AEShipGate_AE2020,TestV2_2_Ellipse_AEShipGate_AE2025
// @since      AE2020
// @alias      fill color,填充颜色,set color,RGBA
func (f *FillNode) SetColor(v [4]float64) error { return f.color.SetStaticValue(v) }

// @summary    Set the fill opacity
// @param      v  the opacity in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_FillKf_AEShipGate_AE2020,TestV2_2_FillKf_AEShipGate_AE2025
// @since      AE2020
// @alias      fill opacity,填充不透明度,set opacity
func (f *FillNode) SetOpacity(v float64) error { return f.opacity.SetStaticValue(v) }

// @summary    Set the fill blend mode
// @param      v  the blend mode as AE's 1-based index
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_ShapeEnums_AEShipGate_AE2020,TestV2_2_ShapeEnums_AEShipGate_AE2025
// @since      AE2020
// @alias      fill blend mode,填充混合模式,blend mode
func (f *FillNode) SetBlendMode(v ShapeBlendMode) error { return setShapeBlendMode(&f.blendMode, v) }

// @summary    Set the fill composite order
// @param      v  whether the fill composites above or below the previous
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_ShapeEnums_AEShipGate_AE2020,TestV2_2_ShapeEnums_AEShipGate_AE2025
// @since      AE2020
// @alias      fill composite order,填充合成顺序,composite order
func (f *FillNode) SetCompositeOrder(v ShapeCompositeOrder) error {
	return setShapeCompositeOrder(&f.compositeOrder, v)
}

// @summary    Set the fill winding rule
// @param      v  the winding rule (Nonzero Winding or Even-Odd)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGOffsetCopies_AEShipGate_AE2020,TestMGOffsetCopies_AEShipGate_AE2025
// @since      AE2020
// @alias      fill rule,填充规则,winding rule,even-odd,non-zero
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

// Properties returns the escape-hatch view.
func (f *FillNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Fill",
		streams: map[string]any{
			"Color":   f.color,
			"Opacity": f.opacity,
		},
	}
}

// GradientFillNode is a `ADBE Vector Graphic - G-Fill`. It models the
// gradient's color and alpha stops (`ADBE Vector Grad Colors`) plus the linear
// ramp direction (`ADBE Vector Grad Start Pt` / `End Pt`). The ramp runs from
// StartPoint to EndPoint in the shape's local coordinate space (AE default
// [0,0]→[100,0], a horizontal ramp); set them to a diagonal/vertical pair to
// rotate the gradient.
//
// Grad Type selects linear vs radial (StartPoint = centre, EndPoint = a point
// on the radius for radial). For a radial gradient the HiLite controls offset
// the bright centre (start color) off the geometric centre: HighlightLength is
// the shift magnitude as a percent of the radius (0 = centred, the default) and
// HighlightAngle is its direction in degrees. They have no visible effect on a
// linear gradient. Stops, direction, type and highlight are static (animated
// gradients are not modeled). The serializer re-encodes the stops to a property
// map and overwrites the Grad Type / Start Pt / End Pt / HiLite cdats plus the
// GCky/Utf8 chunk (length-variable; the enclosing LIST sizes are recomputed).
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
// stops between keyframes.
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

// NewGradientFillNode constructs a default 2-stop black-to-white linear gradient
// (fully opaque) with AE's default horizontal ramp ([0,0]→[100,0]). Callers
// override stops, direction and ramp shape via the Set methods.
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

// @summary    Set the gradient ramp start point
// @description The ramp direction is EndPoint minus StartPoint; it defaults to a
//   horizontal [0,0] to [100,0].
// @param      v  the start point in shape-local coordinates
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradientDir_AEShipGate_AE2020,TestMGGradientDir_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient start point,渐变起点,gradient direction,渐变方向
func (n *GradientFillNode) SetStartPoint(v [2]float64) error { n.startPoint = v; return nil }

// @summary    Set the gradient ramp end point
// @param      v  the end point in shape-local coordinates
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradientDir_AEShipGate_AE2020,TestMGGradientDir_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient end point,渐变终点,gradient direction,渐变方向
func (n *GradientFillNode) SetEndPoint(v [2]float64) error { n.endPoint = v; return nil }

// GradientType returns the ramp shape (linear or radial).
func (n *GradientFillNode) GradientType() GradientType { return n.gradientType }

// @summary    Select the gradient ramp shape (linear or radial)
// @description For radial, StartPoint is the centre and EndPoint sets the outer
//   radius.
// @param      t  the ramp shape (linear or radial)
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradientRadial_AEShipGate_AE2020,TestMGGradientRadial_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient type,渐变类型,linear,radial,线性,径向
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

// @summary    Offset a radial gradient's bright centre
// @description Shifts the bright centre off the geometric centre by a percent of
//   the radius (range -100 to 100; 0 = centred). It has no visible effect on a
//   linear gradient.
// @param      v  the offset magnitude as a percent of the radius
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradientHilite_AEShipGate_AE2020,TestMGGradientHilite_AEShipGate_AE2025
// @since      AE2020
// @alias      highlight length,高亮偏移,radial highlight,渐变高亮
func (n *GradientFillNode) SetHighlightLength(v float64) error {
	if v < -100 || v > 100 {
		return fmt.Errorf("highlight length %g out of range [-100,100]", v)
	}
	n.highlightLength = v
	return nil
}

// @summary    Set a radial gradient's highlight offset direction
// @description Only meaningful together with a non-zero highlight length.
// @param      v  the highlight direction in degrees
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradientHilite_AEShipGate_AE2020,TestMGGradientHilite_AEShipGate_AE2025
// @since      AE2020
// @alias      highlight angle,高亮角度,radial highlight angle,渐变高亮角度
func (n *GradientFillNode) SetHighlightAngle(v float64) error {
	n.highlightAngle = v
	return nil
}

// defaultGradient returns a 2-stop black-to-white gradient with two opaque alpha
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
// returned struct's slices directly also works, but prefer the stop setters for
// range validation.
func (n *GradientFillNode) Gradient() *Gradient { return n.gradient }

// @summary    Replace the gradient color stops
// @description Requires at least two stops; each offset, midpoint and color
//   component must lie in zero to one.
// @param      stops  the ordered list of color stops
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_GradientFill_AEShipGate_AE2020,TestV2_2_GradientFill_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient color stops,渐变色标,color stops,渐变颜色
func (n *GradientFillNode) SetColorStops(stops []GradientColorStop) error {
	return setGradientColorStops(n.gradient, stops)
}

// @summary    Replace the gradient alpha stops
// @description Requires at least two stops; each offset, midpoint and alpha must
//   lie in zero to one.
// @param      stops  the ordered list of alpha stops
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_GradientFill_AEShipGate_AE2020,TestV2_2_GradientFill_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient alpha stops,渐变透明度色标,alpha stops,渐变不透明度
func (n *GradientFillNode) SetAlphaStops(stops []GradientAlphaStop) error {
	return setGradientAlphaStops(n.gradient, stops)
}

// @summary    Append an animated gradient-stops keyframe
// @description The full gradient takes effect at the given time and AE
//   interpolates the stops between keyframes. The first keyframe switches the
//   node to animated mode; the static gradient value is then ignored in favour
//   of the keyframe list. Provide keyframes in ascending time. The keyframe
//   stops are validated. This is a write-only path: re-parsing surfaces the
//   first keyframe's stops as the static value.
// @param      time  the keyframe time in seconds
// @param      g     the gradient value (color and alpha stops) at that time
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestGradientAnim_AEShipGate_AE2020,TestGradientAnim_AEShipGate_AE2025
// @since      AE2020
// @incident   gradient-fill-write-re
// @alias      animated gradient,渐变动画,gradient keyframe,渐变关键帧,color sweep
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

// Properties returns the escape-hatch view.
func (n *GradientFillNode) Properties() *PropertyGroup {
	return &PropertyGroup{Name: "Gradient Fill"}
}

// GradientStrokeNode is a `ADBE Vector Graphic - G-Stroke`. It models the
// gradient's color and alpha stops plus the ramp geometry (direction, type,
// highlight), symmetric to GradientFillNode: StartPoint/EndPoint set the ramp
// direction (AE default [0,0]→[100,0]), GradientType selects linear vs radial,
// and the HiLite controls offset a radial gradient's bright centre. Stroke
// geometry (width / cap / join / dashes / taper / wave) is partially modeled.
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

// NewGradientStrokeNode constructs a default 2-stop black-to-white linear
// gradient stroke with AE's default horizontal ramp ([0,0]→[100,0]). The stroke
// geometry defaults to width 18, round cap, round join and miter 4; override via
// the Set methods.
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

// @summary    Set the gradient stroke width
// @param      v  the stroke width in pixels (must be non-negative)
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeGeom_AEShipGate_AE2020,TestMGGradStrokeGeom_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke width,渐变描边宽度,stroke width
func (n *GradientStrokeNode) SetStrokeWidth(v float64) error {
	if v < 0 {
		return fmt.Errorf("stroke width %g must be ≥ 0", v)
	}
	n.strokeWidth = v
	return nil
}

// LineCap returns the gradient stroke's end-cap style.
func (n *GradientStrokeNode) LineCap() StrokeLineCap { return n.lineCap }

// @summary    Set the gradient stroke end-cap style
// @param      c  the end-cap style (Butt, Round or Projecting)
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeStyle_AEShipGate_AE2020,TestMGGradStrokeStyle_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke line cap,渐变描边端点,line cap,butt,round,projecting
func (n *GradientStrokeNode) SetLineCap(c StrokeLineCap) error {
	if c < StrokeLineCapButt || c > StrokeLineCapProjecting {
		return fmt.Errorf("invalid line cap %d (want 1..3)", c)
	}
	n.lineCap = c
	return nil
}

// LineJoin returns the gradient stroke's corner-join style.
func (n *GradientStrokeNode) LineJoin() StrokeLineJoin { return n.lineJoin }

// @summary    Set the gradient stroke corner-join style
// @param      j  the corner-join style (Miter, Round or Bevel)
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeStyle_AEShipGate_AE2020,TestMGGradStrokeStyle_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke line join,渐变描边接头,line join,miter,round,bevel
func (n *GradientStrokeNode) SetLineJoin(j StrokeLineJoin) error {
	if j < StrokeLineJoinMiter || j > StrokeLineJoinBevel {
		return fmt.Errorf("invalid line join %d (want 1..3)", j)
	}
	n.lineJoin = j
	return nil
}

// MiterLimit returns the gradient stroke's miter limit (only used with a miter join).
func (n *GradientStrokeNode) MiterLimit() float64 { return n.miterLimit }

// @summary    Set the gradient stroke miter limit
// @description Only used when the corner join is Miter.
// @param      v  the miter limit ratio (must be at least one)
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeStyle_AEShipGate_AE2020,TestMGGradStrokeStyle_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke miter limit,渐变描边斜接限制,miter limit
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

// @summary    Set the gradient stroke ramp start point
// @description The ramp direction is EndPoint minus StartPoint; it defaults to a
//   horizontal [0,0] to [100,0].
// @param      v  the start point in shape-local coordinates
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeGeom_AEShipGate_AE2020,TestMGGradStrokeGeom_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke start point,渐变描边起点,gradient direction
func (n *GradientStrokeNode) SetStartPoint(v [2]float64) error { n.startPoint = v; return nil }

// @summary    Set the gradient stroke ramp end point
// @param      v  the end point in shape-local coordinates
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeGeom_AEShipGate_AE2020,TestMGGradStrokeGeom_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke end point,渐变描边终点,gradient direction
func (n *GradientStrokeNode) SetEndPoint(v [2]float64) error { n.endPoint = v; return nil }

// GradientType returns the ramp shape (linear or radial).
func (n *GradientStrokeNode) GradientType() GradientType { return n.gradientType }

// @summary    Select the gradient stroke ramp shape (linear or radial)
// @description For radial, StartPoint is the centre and EndPoint sets the outer
//   radius.
// @param      t  the ramp shape (linear or radial)
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeGeom_AEShipGate_AE2020,TestMGGradStrokeGeom_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke type,渐变描边类型,linear,radial,线性,径向
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

// @summary    Offset a radial gradient stroke's bright centre
// @description Shifts the bright centre off the geometric centre by a percent of
//   the radius (range -100 to 100; 0 = centred). It has no visible effect on a
//   linear gradient.
// @param      v  the offset magnitude as a percent of the radius
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeGeom_AEShipGate_AE2020,TestMGGradStrokeGeom_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke highlight length,渐变描边高亮偏移,radial highlight stroke
func (n *GradientStrokeNode) SetHighlightLength(v float64) error {
	if v < -100 || v > 100 {
		return fmt.Errorf("highlight length %g out of range [-100,100]", v)
	}
	n.highlightLength = v
	return nil
}

// @summary    Set a radial gradient stroke's highlight offset direction
// @description Only meaningful together with a non-zero highlight length.
// @param      v  the highlight direction in degrees
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestMGGradStrokeGeom_AEShipGate_AE2020,TestMGGradStrokeGeom_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke highlight angle,渐变描边高亮角度,radial highlight angle stroke
func (n *GradientStrokeNode) SetHighlightAngle(v float64) error {
	n.highlightAngle = v
	return nil
}

// Gradient returns the live gradient (color + alpha stops).
func (n *GradientStrokeNode) Gradient() *Gradient { return n.gradient }

// @summary    Replace the gradient stroke color stops
// @description Requires at least two stops; each value must lie in zero to one.
// @param      stops  the ordered list of color stops
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_GradientStroke_AEShipGate_AE2020,TestV2_2_GradientStroke_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke color stops,渐变描边色标,color stops
func (n *GradientStrokeNode) SetColorStops(stops []GradientColorStop) error {
	return setGradientColorStops(n.gradient, stops)
}

// @summary    Replace the gradient stroke alpha stops
// @description Requires at least two stops; each value must lie in zero to one.
// @param      stops  the ordered list of alpha stops
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_GradientStroke_AEShipGate_AE2020,TestV2_2_GradientStroke_AEShipGate_AE2025
// @since      AE2020
// @alias      gradient stroke alpha stops,渐变描边透明度色标,alpha stops
func (n *GradientStrokeNode) SetAlphaStops(stops []GradientAlphaStop) error {
	return setGradientAlphaStops(n.gradient, stops)
}

// @summary    Append an animated gradient-stops keyframe to a stroke
// @description The full gradient takes effect at the given time and AE
//   interpolates the stops between keyframes (a color sweep along the stroke).
//   The first keyframe switches the node to animated mode; the static gradient
//   value is then ignored in favour of the keyframe list. Provide keyframes in
//   ascending time. The underlying color stream is the same on fill and stroke.
//   This is a write-only path: re-parsing surfaces the first keyframe's stops as
//   the static value.
// @param      time  the keyframe time in seconds
// @param      g     the gradient value (color and alpha stops) at that time
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestGradientStrokeAnim_AEShipGate_AE2020,TestGradientStrokeAnim_AEShipGate_AE2025
// @since      AE2020
// @incident   gradient-fill-write-re
// @alias      animated gradient stroke,渐变描边动画,gradient stroke keyframe,色标关键帧
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

// Properties returns the escape-hatch view.
func (n *GradientStrokeNode) Properties() *PropertyGroup {
	return &PropertyGroup{Name: "Gradient Stroke"}
}

// @summary    Append a gradient stroke node to a vector group
// @description Defaults to a 2-stop black-to-white gradient. Set the stops via
//   the color and alpha stop setters.
// @returns    the created gradient stroke node
// @domain     gradient
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_GradientStroke_AEShipGate_AE2020,TestV2_2_GradientStroke_AEShipGate_AE2025
// @since      AE2020
// @alias      add gradient stroke,渐变描边,新增渐变描边节点
func (g *VectorGroup) AddGradientStroke() (*GradientStrokeNode, error) {
	n := NewGradientStrokeNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a trim-paths node to a vector group
// @description Defaults to the no-op identity trim (Start 0, End 100, Offset 0).
//   A Trim Paths filter reveals only the arc of the preceding paths between
//   Start and End percent, offset by Offset degrees — the stroke line-draw
//   primitive. Place it after the path-producing shapes it should trim.
// @returns    the created trim-paths node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTrim_AEShipGate_AE2020,TestMGTrim_AEShipGate_AE2025
// @since      AE2020
// @alias      add trim,修剪路径,trim paths,新增修剪节点,line draw reveal
func (g *VectorGroup) AddTrim() (*TrimNode, error) {
	n := NewTrimNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a repeater node to a vector group
// @description Defaults to 3 copies with an identity transform. A Repeater
//   duplicates the preceding paths N times, applying its transform (position,
//   rotation, scale, opacity falloff) cumulatively per copy. Place it after the
//   shapes it should duplicate.
// @returns    the created repeater node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRepeater_AEShipGate_AE2020,TestMGRepeater_AEShipGate_AE2025
// @since      AE2020
// @alias      add repeater,重复器,新增重复器节点,radial burst,grid
func (g *VectorGroup) AddRepeater() (*RepeaterNode, error) {
	n := NewRepeaterNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a round-corners node to a vector group
// @description Defaults to radius 10. A Round Corners filter rounds the corners
//   of the preceding paths by Radius pixels. Place it after the shapes whose
//   corners it should round.
// @returns    the created round-corners node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRoundCorners_AEShipGate_AE2020,TestMGRoundCorners_AEShipGate_AE2025
// @since      AE2020
// @alias      add round corners,圆角,新增圆角节点,soften edges
func (g *VectorGroup) AddRoundCorners() (*RoundCornersNode, error) {
	n := NewRoundCornersNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append an offset-paths node to a vector group
// @description Defaults to amount 10. An Offset Paths filter grows (positive) or
//   shrinks (negative) the preceding paths by Amount pixels. Place it after the
//   shapes it should offset.
// @returns    the created offset-paths node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGOffset_AEShipGate_AE2020,TestMGOffset_AEShipGate_AE2025
// @since      AE2020
// @alias      add offset paths,偏移路径,新增偏移路径节点,inflate,grow,shrink
func (g *VectorGroup) AddOffsetPaths() (*OffsetPathsNode, error) {
	n := NewOffsetPathsNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a merge-paths node to a vector group
// @description Defaults to the Merge mode. A Merge Paths filter boolean-combines
//   all the preceding paths in the group into one path. Place it after the two
//   or more shapes it should combine; the result is painted by the fills and
//   strokes.
// @returns    the created merge-paths node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGMerge_AEShipGate_AE2020,TestMGMerge_AEShipGate_AE2025
// @since      AE2020
// @alias      add merge paths,合并路径,新增合并路径节点,boolean,cut-out
func (g *VectorGroup) AddMergePaths() (*MergePathsNode, error) {
	n := NewMergePathsNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a zigzag node to a vector group
// @description Defaults to size 5, detail 10. A ZigZag filter distorts the
//   preceding paths into a zigzag wave: Size is the amplitude in pixels, Detail
//   is the number of ridges per path segment. Place it after the shapes whose
//   edges it should distort.
// @returns    the created zigzag node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGZigZag_AEShipGate_AE2020,TestMGZigZag_AEShipGate_AE2025
// @since      AE2020
// @alias      add zigzag,锯齿,新增锯齿节点,wave distort
func (g *VectorGroup) AddZigZag() (*ZigZagNode, error) {
	n := NewZigZagNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a pucker-and-bloat node to a vector group
// @description Defaults to amount 0 (no-op identity). Pucker and Bloat bows the
//   preceding paths inward (negative Amount = pucker, concave) or outward
//   (positive = bloat, convex). Place it after the shapes it should distort.
// @returns    the created pucker-and-bloat node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPuckerBloat_AEShipGate_AE2020,TestMGPuckerBloat_AEShipGate_AE2025
// @since      AE2020
// @alias      add pucker bloat,向内凹向外凸,新增膨胀收缩节点,pucker,bloat,organic
func (g *VectorGroup) AddPuckerBloat() (*PuckerBloatNode, error) {
	n := NewPuckerBloatNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a twist node to a vector group
// @description Defaults to angle 0 (no-op identity). Twist rotates the preceding
//   paths progressively — points farther from the centre rotate more — bowing
//   straight edges into spirals. Place it after the shapes it should distort.
// @returns    the created twist node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTwist_AEShipGate_AE2020,TestMGTwist_AEShipGate_AE2025
// @since      AE2020
// @alias      add twist,扭曲,新增扭曲节点,spiral
func (g *VectorGroup) AddTwist() (*TwistNode, error) {
	n := NewTwistNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a wiggle-paths node to a vector group
// @description Defaults to size 0 (no-op identity). Wiggle Paths roughens the
//   preceding paths with time-varying random displacement — the hand-drawn boil
//   jitter. Place it after the shapes it should distort.
// @returns    the created wiggle-paths node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggle_AEShipGate_AE2020,TestMGWiggle_AEShipGate_AE2025
// @since      AE2020
// @alias      add wiggle paths,路径抖动,新增路径抖动节点,roughen,boil,jitter
func (g *VectorGroup) AddWigglePaths() (*WigglePathsNode, error) {
	n := NewWigglePathsNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// @summary    Append a wiggle-transform node to a vector group
// @description Defaults to zero amplitudes (no-op identity). Wiggle Transform
//   randomly jitters a transform (anchor, position, scale, rotation) applied to
//   the preceding paths over time — typically placed after a Repeater to scatter
//   its copies. Place it after the shapes it should affect.
// @returns    the created wiggle-transform node
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggleTransform_AEShipGate_AE2020,TestMGWiggleTransform_AEShipGate_AE2025
// @since      AE2020
// @alias      add wiggle transform,变换抖动,新增变换抖动节点,scatter,random jitter
func (g *VectorGroup) AddWiggleTransform() (*WiggleTransformNode, error) {
	n := NewWiggleTransformNode()
	g.Children = append(g.Children, n)
	return n, nil
}

// TrimNode is a `ADBE Vector Filter - Trim` (Trim Paths). It reveals only the
// portion of the preceding paths between Start and End percent, rotated by
// Offset degrees. Default Start=0, End=100, Offset=0 (identity, no trimming).
// Start/End are percentages (0..100); Offset is in degrees.
//
// Trim Type (Simultaneously/Individually) is AE-default (Simultaneously) and
// elided by AE; the serializer materializes the leaf when the type setter
// selects Individually. Start/End/Offset are static — animated trim flips the
// cdat to a keyframe container.
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

// @summary    Select how the trim treats multiple paths
// @description Simultaneously trims all paths as one combined length (the
//   default); Individually trims each path to the same start and end percent.
//   The Individually value materializes a leaf that AE elides at the default.
//   Only visible with multiple paths in the trim's group.
// @param      t  the trim mode (Simultaneously or Individually)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTrimType_AEShipGate_AE2020,TestMGTrimType_AEShipGate_AE2025
// @since      AE2020
// @alias      trim type,修剪类型,simultaneously,individually,同时修剪,单独修剪
func (n *TrimNode) SetType(t TrimType) error {
	if t != TrimTypeSimultaneously && t != TrimTypeIndividually {
		return fmt.Errorf("TrimNode.SetType: %d out of range (1=Simultaneously, 2=Individually)", t)
	}
	n.trimType = t
	return nil
}

// @summary    Set the trim start percentage
// @param      v  the start percentage (zero to one hundred)
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      trim start,修剪起始,start percentage
func (n *TrimNode) SetStart(v float64) error { return n.start.SetStaticValue(v) }

// @summary    Set the trim end percentage
// @param      v  the end percentage (zero to one hundred)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTrim_AEShipGate_AE2020,TestMGTrim_AEShipGate_AE2025
// @since      AE2020
// @alias      trim end,修剪结束,end percentage,line draw reveal
func (n *TrimNode) SetEnd(v float64) error { return n.end.SetStaticValue(v) }

// @summary    Set the trim offset in degrees
// @param      v  the trim offset in degrees
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      trim offset,修剪偏移,trim rotation,degrees
func (n *TrimNode) SetOffset(v float64) error { return n.offset.SetStaticValue(v) }

// Properties returns the escape-hatch view.
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

// RepeaterNode is a `ADBE Vector Filter - Repeater` (Repeater). It duplicates
// the preceding paths Copies times, applying Transform cumulatively per copy.
// Copies and Offset (which copy index the first instance starts at) are
// animatable scalars; the Transform (anchor, position, scale, rotation, start
// and end opacity) is static. Order (whether each copy stacks below or above the
// previous) defaults to Below and is AE-default-elided, materialized by the
// serializer when the order setter selects Above. With a single-color fill the
// Order makes no visible difference; it matters only when copies are visually
// distinguishable.
type RepeaterNode struct {
	copies    *codec.PropertyStream[float64]
	offset    *codec.PropertyStream[float64]
	order     RepeaterOrder
	orderSet  bool
	transform *RepeaterTransform
}

// RepeaterOrder selects whether each Repeater copy composites below (default) or
// above the previous one (`ADBE Vector Repeater Order`, AE's "Composite"
// dropdown). Stored on disk as a 1-based float64 enum index.
type RepeaterOrder int

const (
	// RepeaterOrderBelow stacks each copy below the previous (AE default).
	RepeaterOrderBelow RepeaterOrder = 1
	// RepeaterOrderAbove stacks each copy above the previous.
	RepeaterOrderAbove RepeaterOrder = 2
)

// RepeaterTransform models the Repeater's nested `ADBE Vector Repeater
// Transform` group: the per-copy transform applied cumulatively. Anchor /
// Position / Scale are Vec2 (Scale in %); Rotation is degrees; Start/End
// Opacity are % applied to the first/last copy with a linear falloff between.
// All static, stored as plain values.
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
		order:  RepeaterOrderBelow,
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

// Order returns whether copies composite below (default) or above the previous.
func (n *RepeaterNode) Order() RepeaterOrder { return n.order }

// OrderSet reports whether SetOrder was called (serializer splice trigger).
func (n *RepeaterNode) OrderSet() bool { return n.orderSet }

// @summary    Select whether copies composite below or above the previous
// @description The Above value materializes a leaf that AE elides at the default.
//   With a single-color fill this has no visible effect.
// @param      o  the composite order (Below or Above)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRepeaterOrder_AEShipGate_AE2020,TestMGRepeaterOrder_AEShipGate_AE2025
// @since      AE2020
// @alias      repeater order,重复器合成顺序,composite order,above,below
func (n *RepeaterNode) SetOrder(o RepeaterOrder) error {
	if o != RepeaterOrderBelow && o != RepeaterOrderAbove {
		return fmt.Errorf("RepeaterNode.SetOrder: %d out of range (1=Below, 2=Above)", o)
	}
	n.order = o
	n.orderSet = true
	return nil
}

// @summary    Set the number of repeater copies
// @param      v  the copy count (must be at least one)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRepeater_AEShipGate_AE2020,TestMGRepeater_AEShipGate_AE2025
// @since      AE2020
// @alias      repeater copies,重复器份数,copies count
func (n *RepeaterNode) SetCopies(v float64) error {
	if v < 1 {
		return fmt.Errorf("RepeaterNode.SetCopies: %g out of range (want ≥ 1)", v)
	}
	return n.copies.SetStaticValue(v)
}

// @summary    Set the copy-index offset of the first instance
// @param      v  the copy-index offset of the first instance
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      repeater offset,重复器偏移,copy offset
func (n *RepeaterNode) SetOffset(v float64) error { return n.offset.SetStaticValue(v) }

func (t *RepeaterTransform) Anchor() [2]float64    { return t.anchor }
func (t *RepeaterTransform) Position() [2]float64  { return t.position }
func (t *RepeaterTransform) Scale() [2]float64     { return t.scale }
func (t *RepeaterTransform) Rotation() float64     { return t.rotation }
func (t *RepeaterTransform) StartOpacity() float64 { return t.startOpacity }
func (t *RepeaterTransform) EndOpacity() float64   { return t.endOpacity }

// @summary    Set the per-copy anchor point
// @param      v  the per-copy anchor point in pixels
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      repeater anchor,重复器锚点,per-copy anchor
func (t *RepeaterTransform) SetAnchor(v [2]float64) error { t.anchor = v; return nil }

// @summary    Set the per-copy position offset
// @description This is the spacing between copies — the grid and line knob.
// @param      v  the per-copy position offset in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRepeater_AEShipGate_AE2020,TestMGRepeater_AEShipGate_AE2025
// @since      AE2020
// @alias      repeater position,重复器位置偏移,spacing,per-copy position
func (t *RepeaterTransform) SetPosition(v [2]float64) error { t.position = v; return nil }

// @summary    Set the per-copy scale
// @description The scale is cumulative: copy k is scaled k times.
// @param      v  the per-copy scale in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRepeaterOrder_AEShipGate_AE2020,TestMGRepeaterOrder_AEShipGate_AE2025
// @since      AE2020
// @alias      repeater scale,重复器缩放,per-copy scale
func (t *RepeaterTransform) SetScale(v [2]float64) error { t.scale = v; return nil }

// @summary    Set the per-copy rotation
// @description This is the radial-burst knob.
// @param      v  the per-copy rotation in degrees
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      repeater rotation,重复器旋转,radial burst,per-copy rotation
func (t *RepeaterTransform) SetRotation(v float64) error { t.rotation = v; return nil }

// @summary    Set the first copy's opacity
// @param      v  the start opacity in percent (zero to one hundred)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRepeaterOrder_AEShipGate_AE2020,TestMGRepeaterOrder_AEShipGate_AE2025
// @since      AE2020
// @alias      repeater start opacity,重复器起始不透明度,start opacity falloff
func (t *RepeaterTransform) SetStartOpacity(v float64) error {
	if v < 0 || v > 100 {
		return fmt.Errorf("SetStartOpacity: %g out of range 0..100", v)
	}
	t.startOpacity = v
	return nil
}

// @summary    Set the last copy's opacity
// @description This is the falloff target across copies.
// @param      v  the end opacity in percent (zero to one hundred)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRepeaterOrder_AEShipGate_AE2020,TestMGRepeaterOrder_AEShipGate_AE2025
// @since      AE2020
// @alias      repeater end opacity,重复器结束不透明度,end opacity falloff
func (t *RepeaterTransform) SetEndOpacity(v float64) error {
	if v < 0 || v > 100 {
		return fmt.Errorf("SetEndOpacity: %g out of range 0..100", v)
	}
	t.endOpacity = v
	return nil
}

// Properties returns the escape-hatch view.
func (n *RepeaterNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Repeater",
		streams: map[string]any{
			"Copies": n.copies,
			"Offset": n.offset,
		},
	}
}

// RoundCornersNode is a `ADBE Vector Filter - RC` (Round Corners). It rounds the
// corners of the preceding paths in the stack by Radius pixels. Its single
// sub-stream `ADBE Vector RoundCorner Radius` is an animatable 1D scalar
// (default 10). Place it after the path-producing shapes whose corners it should
// round.
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

// @summary    Set the corner radius
// @param      v  the corner radius in pixels (must be non-negative)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGRoundCorners_AEShipGate_AE2020,TestMGRoundCorners_AEShipGate_AE2025
// @since      AE2020
// @alias      round corners radius,圆角半径,corner radius
func (n *RoundCornersNode) SetRadius(v float64) error {
	if v < 0 {
		return fmt.Errorf("RoundCornersNode.SetRadius: %g out of range (want >= 0)", v)
	}
	return n.radius.SetStaticValue(v)
}

// Properties returns the escape-hatch view.
func (n *RoundCornersNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Round Corners",
		streams: map[string]any{
			"Radius": n.radius,
		},
	}
}

// OffsetPathsNode is a `ADBE Vector Filter - Offset` (Offset Paths). It grows
// (positive Amount) or shrinks (negative Amount) the preceding paths in the
// stack by Amount pixels. Its headline sub-stream `ADBE Vector Offset Amount` is
// an animatable 1D scalar (default 10). Place it after the path-producing shapes
// it should offset.
//
// Line Join / Miter Limit / Copies / Copy Offset are AE-default and elided in
// the Amount-only template; each is materialized on demand by the serializer
// when its setter is called.
type OffsetPathsNode struct {
	amount        *codec.PropertyStream[float64]
	lineJoin      StrokeLineJoin
	lineJoinSet   bool
	miterLimit    float64
	miterLimitSet bool
	copies        float64
	copiesSet     bool
	copyOffset    float64
	copyOffsetSet bool
}

// NewOffsetPathsNode constructs a default OffsetPathsNode (Amount=10, Line
// Join=Miter, Miter Limit=4, Copies=1, Copy Offset=1).
func NewOffsetPathsNode() *OffsetPathsNode {
	n := &OffsetPathsNode{
		amount:     codec.NewPropertyStream[float64](),
		lineJoin:   StrokeLineJoinMiter,
		miterLimit: 4,
		copies:     1,
		copyOffset: 1,
	}
	_ = n.amount.SetStaticValue(10)
	return n
}

func (n *OffsetPathsNode) Kind() ShapeNodeKind              { return ShapeKindOffsetPaths }
func (n *OffsetPathsNode) Amount() *PropertyStream[float64] { return n.amount }

// @summary    Set the offset amount
// @description A positive amount grows the paths, a negative amount shrinks them.
// @param      v  the offset amount in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGOffset_AEShipGate_AE2020,TestMGOffset_AEShipGate_AE2025
// @since      AE2020
// @alias      offset amount,偏移量,grow shrink,offset paths amount
func (n *OffsetPathsNode) SetAmount(v float64) error { return n.amount.SetStaticValue(v) }

// LineJoin returns the corner join used where the grown outline turns
// (Miter/Round/Bevel; default Miter). Uses the same enum as Stroke Line Join.
func (n *OffsetPathsNode) LineJoin() StrokeLineJoin { return n.lineJoin }

// LineJoinSet reports whether SetLineJoin was called (serializer splice trigger).
func (n *OffsetPathsNode) LineJoinSet() bool { return n.lineJoinSet }

// @summary    Select the corner join for the offset outline
// @description Miter is a sharp point, Round, and Bevel is a flat cut. The
//   default Miter materializes a leaf that AE elides at the default.
// @param      j  the corner join (Miter, Round or Bevel)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGOffsetExtras_AEShipGate_AE2020,TestMGOffsetExtras_AEShipGate_AE2025
// @since      AE2020
// @alias      offset line join,偏移接头,offset corner join,miter,bevel
func (n *OffsetPathsNode) SetLineJoin(j StrokeLineJoin) error {
	if j < StrokeLineJoinMiter || j > StrokeLineJoinBevel {
		return fmt.Errorf("OffsetPathsNode.SetLineJoin: %d out of range (1=Miter, 2=Round, 3=Bevel)", j)
	}
	n.lineJoin = j
	n.lineJoinSet = true
	return nil
}

// MiterLimit returns the miter clip ratio (default 4; only affects Miter joins
// at sharp angles).
func (n *OffsetPathsNode) MiterLimit() float64 { return n.miterLimit }

// MiterLimitSet reports whether SetMiterLimit was called (serializer splice trigger).
func (n *OffsetPathsNode) MiterLimitSet() bool { return n.miterLimitSet }

// @summary    Set the offset miter clip ratio
// @description A sharp corner whose miter would extend past limit times width is
//   clipped flat to a bevel. The default 4 materializes a leaf that AE elides at
//   the default.
// @param      v  the miter clip ratio (must be at least one)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGOffsetExtras_AEShipGate_AE2020,TestMGOffsetExtras_AEShipGate_AE2025
// @since      AE2020
// @alias      offset miter limit,偏移斜接限制,miter limit
func (n *OffsetPathsNode) SetMiterLimit(v float64) error {
	if v < 1 {
		return fmt.Errorf("OffsetPathsNode.SetMiterLimit: %g out of range (want >= 1)", v)
	}
	n.miterLimit = v
	n.miterLimitSet = true
	return nil
}

// Copies returns the number of progressively-offset copies (default 1).
func (n *OffsetPathsNode) Copies() float64 { return n.copies }

// @summary    Set the number of stacked offset copies
// @description Each copy is offset by a further Amount pixels — N nested
//   outlines growing outward (or inward for a negative amount). The sub-stream
//   is AE-default-elided; setting it materializes the leaf.
// @param      v  the copy count (must be at least one)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGOffsetCopies_AEShipGate_AE2020,TestMGOffsetCopies_AEShipGate_AE2025
// @since      AE2020
// @alias      offset copies,偏移副本数,nested outlines,copies
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

// CopyOffset returns the per-copy offset multiplier applied when Copies > 1
// (default 1 — successive copies step by one further Amount each).
func (n *OffsetPathsNode) CopyOffset() float64 { return n.copyOffset }

// CopyOffsetSet reports whether SetCopyOffset was called (serializer splice trigger).
func (n *OffsetPathsNode) CopyOffsetSet() bool { return n.copyOffsetSet }

// @summary    Scale the spacing between successive offset copies
// @description Only meaningful when there is more than one copy. The default 1
//   materializes a leaf that AE elides at the default.
// @param      v  the per-copy spacing multiplier
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGOffsetExtras_AEShipGate_AE2020,TestMGOffsetExtras_AEShipGate_AE2025
// @since      AE2020
// @alias      offset copy offset,偏移副本间距,copy spacing multiplier
func (n *OffsetPathsNode) SetCopyOffset(v float64) error {
	n.copyOffset = v
	n.copyOffsetSet = true
	return nil
}

// Properties returns the escape-hatch view.
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

// MergePathsNode is a `ADBE Vector Filter - Merge` (Merge Paths). It
// boolean-combines all the paths below it in the group into a single path per
// its Type. Its only sub-stream `ADBE Vector Merge Type` is a non-animated enum
// (default Merge); modeled as a plain value.
type MergePathsNode struct {
	mergeType MergeType
}

// NewMergePathsNode constructs a default MergePathsNode (Type=Merge).
func NewMergePathsNode() *MergePathsNode {
	return &MergePathsNode{mergeType: MergeTypeMerge}
}

func (n *MergePathsNode) Kind() ShapeNodeKind { return ShapeKindMergePaths }
func (n *MergePathsNode) Type() MergeType     { return n.mergeType }

// @summary    Set the boolean merge mode
// @param      v  the merge mode index (one through five)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGMerge_AEShipGate_AE2020,TestMGMerge_AEShipGate_AE2025
// @since      AE2020
// @alias      merge type,合并模式,boolean mode,merge,subtract,intersect,exclude
func (n *MergePathsNode) SetType(v MergeType) error {
	if v < MergeTypeMerge || v > MergeTypeExclude {
		return fmt.Errorf("MergePathsNode.SetType: invalid value %d (want 1..5)", v)
	}
	n.mergeType = v
	return nil
}

// Properties returns the escape-hatch view.
func (n *MergePathsNode) Properties() *PropertyGroup {
	return &PropertyGroup{Name: "Merge Paths"}
}

// ZigZagNode is a `ADBE Vector Filter - Zigzag` (ZigZag). It distorts the paths
// below it in the stack into a zigzag wave. Size (amplitude, pixels) and Detail
// (ridges per path segment) are animatable 1D scalars (defaults 5 / 10). Points
// (Corner/Smooth) selects whether each ridge is a sharp sawtooth corner (AE
// default) or a smooth scalloped wave; the non-default Smooth value is
// AE-default-elided and materialized by the serializer when its setter selects
// it.
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

// @summary    Set the zigzag amplitude
// @param      v  the zigzag amplitude in pixels (must be non-negative)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGZigZag_AEShipGate_AE2020,TestMGZigZag_AEShipGate_AE2025
// @since      AE2020
// @alias      zigzag size,锯齿幅度,amplitude
func (n *ZigZagNode) SetSize(v float64) error {
	if v < 0 {
		return fmt.Errorf("ZigZagNode.SetSize: %g out of range (want >= 0)", v)
	}
	return n.size.SetStaticValue(v)
}

// @summary    Set the number of ridges per path segment
// @param      v  the ridge count per segment (must be non-negative)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGZigZag_AEShipGate_AE2020,TestMGZigZag_AEShipGate_AE2025
// @since      AE2020
// @alias      zigzag detail,锯齿密度,ridges per segment
func (n *ZigZagNode) SetDetail(v float64) error {
	if v < 0 {
		return fmt.Errorf("ZigZagNode.SetDetail: %g out of range (want >= 0)", v)
	}
	return n.detail.SetStaticValue(v)
}

// Points returns whether the ridges are sharp corners or smooth waves.
func (n *ZigZagNode) Points() ZigZagPoints { return n.points }

// @summary    Select sharp-corner or smooth-wave ridges
// @description Corner is sharp sawtooth ridges (the default); Smooth is
//   scalloped wave ridges and materializes a leaf that AE elides at the default.
// @param      p  the ridge style (Corner or Smooth)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGZigZagPoints_AEShipGate_AE2020,TestMGZigZagPoints_AEShipGate_AE2025
// @since      AE2020
// @alias      zigzag points,锯齿形状,corner smooth,wave type
func (n *ZigZagNode) SetPoints(p ZigZagPoints) error {
	if p != ZigZagPointsCorner && p != ZigZagPointsSmooth {
		return fmt.Errorf("ZigZagNode.SetPoints: %d out of range (1=Corner, 2=Smooth)", p)
	}
	n.points = p
	return nil
}

// Properties returns the escape-hatch view.
func (n *ZigZagNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "ZigZag",
		streams: map[string]any{
			"Size":   n.size,
			"Detail": n.detail,
		},
	}
}

// PuckerBloatNode is a `ADBE Vector Filter - PB` (Pucker & Bloat). It bows the
// paths below it inward (negative Amount = pucker) or outward (positive = bloat).
// Its single sub-stream `ADBE Vector PuckerBloat Amount` is an animatable 1D
// scalar (percent; default 0 = no distortion). Place it after the path-producing
// shapes it should distort.
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

// @summary    Set the pucker or bloat amount
// @description A negative amount puckers (concave), a positive amount bloats
//   (convex).
// @param      v  the pucker or bloat amount in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGPuckerBloat_AEShipGate_AE2020,TestMGPuckerBloat_AEShipGate_AE2025
// @since      AE2020
// @alias      pucker bloat amount,膨胀收缩量,pucker,bloat
func (n *PuckerBloatNode) SetAmount(v float64) error { return n.amount.SetStaticValue(v) }

// Properties returns the escape-hatch view.
func (n *PuckerBloatNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Pucker & Bloat",
		streams: map[string]any{
			"Amount": n.amount,
		},
	}
}

// TwistNode is a `ADBE Vector Filter - Twist`. It rotates the paths below it
// progressively (more rotation farther from the twist centre), bowing straight
// edges into spirals. Its headline sub-stream `ADBE Vector Twist Angle` is an
// animatable 1D scalar (degrees; default 0 = no twist). Center (Vec2,
// shape-local pixels relative to the path centre) is the pivot the twist rotates
// around; it defaults to [0,0] and is AE-default-elided, materialized by the
// serializer when its setter offsets it. Place it after the path-producing
// shapes it should distort.
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

// @summary    Set the twist angle
// @description A positive angle twists clockwise, a negative angle
//   counter-clockwise.
// @param      v  the twist angle in degrees
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTwist_AEShipGate_AE2020,TestMGTwist_AEShipGate_AE2025
// @since      AE2020
// @alias      twist angle,扭曲角度,spiral angle
func (n *TwistNode) SetAngle(v float64) error { return n.angle.SetStaticValue(v) }

// Center returns the twist pivot (Vec2, shape-local pixels relative to the path
// centre). Default [0,0] (rotate about the path centre).
func (n *TwistNode) Center() [2]float64 { return n.center }

// CenterSet reports whether SetCenter has been called (so the serializer knows
// to materialize the otherwise-elided Center leaf).
func (n *TwistNode) CenterSet() bool { return n.centerSet }

// @summary    Offset the twist pivot from the path centre
// @description The default [0,0] is AE-default-elided; a non-zero centre
//   materializes the leaf.
// @param      c  the twist pivot in shape-local pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGTwistCenter_AEShipGate_AE2020,TestMGTwistCenter_AEShipGate_AE2025
// @since      AE2020
// @alias      twist center,扭曲中心,pivot,twist pivot
func (n *TwistNode) SetCenter(c [2]float64) error {
	n.center = c
	n.centerSet = true
	return nil
}

// Properties returns the escape-hatch view.
func (n *TwistNode) Properties() *PropertyGroup {
	return &PropertyGroup{
		Name: "Twist",
		streams: map[string]any{
			"Angle": n.angle,
		},
	}
}

// WigglePathsNode is a `ADBE Vector Filter - Roughen` (Wiggle Paths). It
// roughens the paths below it with time-varying random displacement — the
// hand-drawn boil jitter. Four animatable 1D scalar sub-streams are modeled:
// Size (displacement amplitude), Detail (segments per unit length),
// WigglesPerSecond (the temporal frequency, `ADBE Vector Temporal Freq`), and
// RandomSeed. The Points enum plus the Correlation and Temporal/Spatial Phase
// sub-streams are left at their defaults and elided. Place it after the
// path-producing shapes it should distort.
type WigglePathsNode struct {
	size             *codec.PropertyStream[float64]
	detail           *codec.PropertyStream[float64]
	wigglesPerSecond *codec.PropertyStream[float64]
	randomSeed       *codec.PropertyStream[float64]

	points           RoughenPoints
	correlation      float64
	correlationSet   bool
	temporalPhase    float64
	temporalPhaseSet bool
	spatialPhase     float64
	spatialPhaseSet  bool
}

// RoughenPoints selects whether the random roughen displacement breaks the path
// into sharp corner spikes or smooth scalloped bumps (`ADBE Vector Roughen
// Points`, AE's "Points" dropdown on Wiggle Paths). Stored on disk as a 1-based
// float64 enum index.
type RoughenPoints int

const (
	// RoughenPointsCorner makes each displaced segment a sharp corner (AE default).
	RoughenPointsCorner RoughenPoints = 1
	// RoughenPointsSmooth makes each displaced segment a smooth scalloped bump.
	RoughenPointsSmooth RoughenPoints = 2
)

// NewWigglePathsNode constructs a default WigglePathsNode (Size=0 → no
// displacement, the identity; Detail=10, WigglesPerSecond=2, RandomSeed=0 match
// AE's filter defaults; Points=Corner, Correlation=50, Temporal/Spatial Phase=0).
func NewWigglePathsNode() *WigglePathsNode {
	n := &WigglePathsNode{
		size:             codec.NewPropertyStream[float64](),
		detail:           codec.NewPropertyStream[float64](),
		wigglesPerSecond: codec.NewPropertyStream[float64](),
		randomSeed:       codec.NewPropertyStream[float64](),
		points:           RoughenPointsCorner,
		correlation:      50,
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

// Points / Correlation / TemporalPhase / SpatialPhase return the modulation
// sub-stream values. CorrelationSet / TemporalPhaseSet / SpatialPhaseSet report
// whether each was explicitly set (controls materialization on lower).
func (n *WigglePathsNode) Points() RoughenPoints  { return n.points }
func (n *WigglePathsNode) Correlation() float64   { return n.correlation }
func (n *WigglePathsNode) CorrelationSet() bool   { return n.correlationSet }
func (n *WigglePathsNode) TemporalPhase() float64 { return n.temporalPhase }
func (n *WigglePathsNode) TemporalPhaseSet() bool { return n.temporalPhaseSet }
func (n *WigglePathsNode) SpatialPhase() float64  { return n.spatialPhase }
func (n *WigglePathsNode) SpatialPhaseSet() bool  { return n.spatialPhaseSet }

// @summary    Set the wiggle displacement amplitude
// @param      v  the displacement amplitude in pixels (0 = no roughening)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggle_AEShipGate_AE2020,TestMGWiggle_AEShipGate_AE2025
// @since      AE2020
// @alias      wiggle size,路径抖动幅度,displacement amplitude
func (n *WigglePathsNode) SetSize(v float64) error { return n.size.SetStaticValue(v) }

// @summary    Set the wiggle detail
// @description Higher detail produces finer, more frequent ridges.
// @param      v  the number of segments per path length
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggle_AEShipGate_AE2020,TestMGWiggle_AEShipGate_AE2025
// @since      AE2020
// @alias      wiggle detail,路径抖动细节,segments per unit length
func (n *WigglePathsNode) SetDetail(v float64) error { return n.detail.SetStaticValue(v) }

// @summary    Set the wiggle temporal frequency
// @description This is how fast the random edge churns over time.
// @param      v  the temporal frequency in wiggles per second
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggle_AEShipGate_AE2020,TestMGWiggle_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural modulation parameter that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle frequency,路径抖动频率,wiggles per second,temporal frequency
func (n *WigglePathsNode) SetWigglesPerSecond(v float64) error {
	return n.wigglesPerSecond.SetStaticValue(v)
}

// @summary    Set the wiggle random seed
// @description The seed selects which displacement pattern is generated.
// @param      v  the random seed selecting the displacement pattern
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggle_AEShipGate_AE2020,TestMGWiggle_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural modulation parameter that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle random seed,路径抖动随机种子,random seed
func (n *WigglePathsNode) SetRandomSeed(v float64) error { return n.randomSeed.SetStaticValue(v) }

// @summary    Select sharp-corner or smooth-bump roughen ridges
// @description Corner is sharp displaced spikes (the default); Smooth is rounded
//   scalloped bumps and materializes a leaf that AE elides at the default.
// @param      p  the ridge style (Corner or Smooth)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggleMod_AEShipGate_AE2020,TestMGWiggleMod_AEShipGate_AE2025
// @since      AE2020
// @alias      wiggle points,路径抖动形状,corner smooth,roughen points
func (n *WigglePathsNode) SetPoints(p RoughenPoints) error {
	if p != RoughenPointsCorner && p != RoughenPointsSmooth {
		return fmt.Errorf("WigglePathsNode.SetPoints: %d out of range (1=Corner, 2=Smooth)", p)
	}
	n.points = p
	return nil
}

// @summary    Set how correlated successive random displacements are
// @description Range 0 to 100: 0 means each point jitters independently
//   (jagged), 100 means neighbours move together (smooth coherent boil). The
//   default 50 is elided; a non-default value materializes the leaf.
// @param      v  the correlation in percent (zero to one hundred)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggleMod_AEShipGate_AE2020,TestMGWiggleMod_AEShipGate_AE2025
// @since      AE2020
// @alias      wiggle correlation,路径抖动相关性,roughen correlation
func (n *WigglePathsNode) SetCorrelation(v float64) error {
	if v < 0 || v > 100 {
		return fmt.Errorf("WigglePathsNode.SetCorrelation: %g out of range (want 0..100)", v)
	}
	n.correlation = v
	n.correlationSet = true
	return nil
}

// @summary    Set the temporal phase into the noise sequence
// @description This selects a different time-slice of the same seeded random
//   pattern. The default is elided; a non-default value materializes the leaf.
//   Like the random seed, a phase shift produces a statistically-equivalent
//   alternate edge with no categorically-correct pixel result, so it is
//   roundtrip-verified.
// @param      v  the temporal phase in degrees
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggleModRT_AEShipGate_AE2020,TestMGWiggleModRT_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural noise-phase sample that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle temporal phase,路径抖动时间相位,roughen temporal phase
func (n *WigglePathsNode) SetTemporalPhase(v float64) error {
	n.temporalPhase = v
	n.temporalPhaseSet = true
	return nil
}

// @summary    Set the spatial phase into the noise field
// @description This selects a different spatial offset of the same seeded random
//   pattern. The default is elided; a non-default value materializes the leaf.
//   Roundtrip-verified for the same reason as the temporal phase.
// @param      v  the spatial phase in degrees
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggleModRT_AEShipGate_AE2020,TestMGWiggleModRT_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural noise-phase sample that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle spatial phase,路径抖动空间相位,roughen spatial phase
func (n *WigglePathsNode) SetSpatialPhase(v float64) error {
	n.spatialPhase = v
	n.spatialPhaseSet = true
	return nil
}

// Properties returns the escape-hatch view.
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

// WiggleTransformNode is a `ADBE Vector Filter - Wiggler` (Wiggle Transform). It
// randomly jitters a transform applied to the paths below it over time — usually
// placed after a Repeater to scatter its copies. WigglesPerSecond (the temporal
// frequency, `ADBE Vector Xform Temporal Freq`) and RandomSeed are animatable 1D
// scalars; the per-channel wiggle amplitudes live in the nested Transform group
// (anchor, position, scale as Vec2, rotation in degrees) modeled statically.
// Correlation and Temporal/Spatial Phase are left at their defaults and elided.
type WiggleTransformNode struct {
	wigglesPerSecond *codec.PropertyStream[float64]
	randomSeed       *codec.PropertyStream[float64]
	transform        *WigglerTransform

	correlation      float64
	correlationSet   bool
	temporalPhase    float64
	temporalPhaseSet bool
	spatialPhase     float64
	spatialPhaseSet  bool
}

// WigglerTransform models the Wiggle Transform's nested `ADBE Vector Wiggler
// Transform` group: the per-channel random wiggle AMPLITUDES (not absolute
// transform values). Anchor / Position / Scale are Vec2 (Scale amplitude in %),
// Rotation in degrees. A zero amplitude means that channel does not wiggle. All
// static, stored as plain values.
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
		correlation:      50,
	}
	_ = n.wigglesPerSecond.SetStaticValue(2)
	_ = n.randomSeed.SetStaticValue(0)
	return n
}

// Correlation / TemporalPhase / SpatialPhase return the modulation sub-stream
// values; the Set getters report whether each was explicitly set (controls
// materialization on lower).
func (n *WiggleTransformNode) Correlation() float64   { return n.correlation }
func (n *WiggleTransformNode) CorrelationSet() bool   { return n.correlationSet }
func (n *WiggleTransformNode) TemporalPhase() float64 { return n.temporalPhase }
func (n *WiggleTransformNode) TemporalPhaseSet() bool { return n.temporalPhaseSet }
func (n *WiggleTransformNode) SpatialPhase() float64  { return n.spatialPhase }
func (n *WiggleTransformNode) SpatialPhaseSet() bool  { return n.spatialPhaseSet }

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

// @summary    Set the wiggle-transform temporal frequency
// @description This is how fast the transform churns over time.
// @param      v  the temporal frequency in wiggles per second
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggleTransform_AEShipGate_AE2020,TestMGWiggleTransform_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural modulation parameter that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle transform frequency,变换抖动频率,wiggles per second
func (n *WiggleTransformNode) SetWigglesPerSecond(v float64) error {
	return n.wigglesPerSecond.SetStaticValue(v)
}

// @summary    Set the wiggle-transform random seed
// @description The seed selects which wiggle pattern is generated.
// @param      v  the random seed selecting the wiggle pattern
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggleTransform_AEShipGate_AE2020,TestMGWiggleTransform_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural modulation parameter that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle transform random seed,变换抖动随机种子,random seed
func (n *WiggleTransformNode) SetRandomSeed(v float64) error { return n.randomSeed.SetStaticValue(v) }

// @summary    Set how correlated the random transform jitter is
// @description Range 0 to 100 (default 50); a non-default value materializes the
//   leaf. The wiggle is a per-frame random transform offset and correlation
//   modulates its temporal smoothness, with no single-frame
//   categorically-correct pixel result, so it is roundtrip-verified.
// @param      v  the correlation in percent (zero to one hundred)
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggleModRT_AEShipGate_AE2020,TestMGWiggleModRT_AEShipGate_AE2025
// @since      AE2020
// @boundary   a time-varying jitter modulation that cannot be single-frame pixel-gated; AE accepts and reads back the value
// @alias      wiggle transform correlation,变换抖动相关性
func (n *WiggleTransformNode) SetCorrelation(v float64) error {
	if v < 0 || v > 100 {
		return fmt.Errorf("WiggleTransformNode.SetCorrelation: %g out of range (want 0..100)", v)
	}
	n.correlation = v
	n.correlationSet = true
	return nil
}

// @summary    Set the transform-wiggle temporal phase
// @description This selects a different time-slice of the seeded random pattern.
//   The default is elided; a non-default value materializes the leaf.
//   Roundtrip-verified (a noise-phase selection, not pixel-gatable).
// @param      v  the temporal phase in degrees
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggleModRT_AEShipGate_AE2020,TestMGWiggleModRT_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural noise-phase sample that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle transform temporal phase,变换抖动时间相位
func (n *WiggleTransformNode) SetTemporalPhase(v float64) error {
	n.temporalPhase = v
	n.temporalPhaseSet = true
	return nil
}

// @summary    Set the transform-wiggle spatial phase
// @description This selects a different spatial offset of the seeded random
//   pattern. The default is elided; a non-default value materializes the leaf.
//   Roundtrip-verified (a noise-phase selection, not pixel-gatable).
// @param      v  the spatial phase in degrees
// @domain     shape
// @stability  alpha
// @verify     ae-accept
// @gate       TestMGWiggleModRT_AEShipGate_AE2020,TestMGWiggleModRT_AEShipGate_AE2025
// @since      AE2020
// @boundary   a procedural noise-phase sample that cannot be pixel-gated; AE accepts and reads back the value
// @alias      wiggle transform spatial phase,变换抖动空间相位
func (n *WiggleTransformNode) SetSpatialPhase(v float64) error {
	n.spatialPhase = v
	n.spatialPhaseSet = true
	return nil
}

// Anchor / Position / Scale / Rotation return the current wiggle amplitudes.
func (t *WigglerTransform) Anchor() [2]float64   { return t.anchor }
func (t *WigglerTransform) Position() [2]float64 { return t.position }
func (t *WigglerTransform) Scale() [2]float64    { return t.scale }
func (t *WigglerTransform) Rotation() float64    { return t.rotation }

// @summary    Set the anchor-point wiggle amplitude
// @param      v  the anchor-point wiggle amplitude in pixels
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      wiggler anchor amplitude,抖动锚点幅度,anchor wiggle
func (t *WigglerTransform) SetAnchor(v [2]float64) error { t.anchor = v; return nil }

// @summary    Set the position wiggle amplitude
// @param      v  the position wiggle amplitude in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggleTransform_AEShipGate_AE2020,TestMGWiggleTransform_AEShipGate_AE2025
// @since      AE2020
// @alias      wiggler position amplitude,抖动位置幅度,position wiggle
func (t *WigglerTransform) SetPosition(v [2]float64) error { t.position = v; return nil }

// @summary    Set the scale wiggle amplitude
// @param      v  the scale wiggle amplitude in percent
// @domain     shape
// @stability  stable
// @verify     ae-accept
// @gate       TestShapeGeom_AEShipGate_AE2020,TestShapeGeom_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving; AE accepts the value and it survives a re-save
// @alias      wiggler scale amplitude,抖动缩放幅度,scale wiggle
func (t *WigglerTransform) SetScale(v [2]float64) error { t.scale = v; return nil }

// @summary    Set the rotation wiggle amplitude
// @param      v  the rotation wiggle amplitude in degrees
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestMGWiggleTransform_AEShipGate_AE2020,TestMGWiggleTransform_AEShipGate_AE2025
// @since      AE2020
// @alias      wiggler rotation amplitude,抖动旋转幅度,rotation wiggle
func (t *WigglerTransform) SetRotation(v float64) error { t.rotation = v; return nil }

// Properties returns the escape-hatch view.
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

// StrokeNode is a `ADBE Vector Graphic - Stroke`. Default Color=[0,0,0,1]
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

// @summary    Set the stroke color
// @param      v  the RGBA color, each channel in zero to one
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Stroke_AEShipGate_AE2020,TestV2_2_Stroke_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke color,描边颜色,set color,RGBA
func (s *StrokeNode) SetColor(v [4]float64) error { return s.color.SetStaticValue(v) }

// @summary    Set the stroke opacity
// @param      v  the opacity in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeKf_AEShipGate_AE2020,TestV2_2_StrokeKf_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke opacity,描边不透明度,set opacity
func (s *StrokeNode) SetOpacity(v float64) error { return s.opacity.SetStaticValue(v) }

// @summary    Set the stroke width
// @param      v  the stroke width in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Stroke_AEShipGate_AE2020,TestV2_2_Stroke_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke width,描边宽度,set width
func (s *StrokeNode) SetWidth(v float64) error { return s.width.SetStaticValue(v) }

func (s *StrokeNode) LineCap() StrokeLineCap              { return s.lineCap }
func (s *StrokeNode) LineJoin() StrokeLineJoin            { return s.lineJoin }
func (s *StrokeNode) MiterLimit() float64                 { return s.miterLimit }
func (s *StrokeNode) BlendMode() ShapeBlendMode           { return s.blendMode }
func (s *StrokeNode) CompositeOrder() ShapeCompositeOrder { return s.compositeOrder }

// @summary    Set the stroke blend mode
// @param      v  the blend mode as AE's 1-based index
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_ShapeEnums_AEShipGate_AE2020,TestV2_2_ShapeEnums_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke blend mode,描边混合模式,blend mode
func (s *StrokeNode) SetBlendMode(v ShapeBlendMode) error { return setShapeBlendMode(&s.blendMode, v) }

// @summary    Set the stroke composite order
// @param      v  whether the stroke composites above or below the previous
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_ShapeEnums_AEShipGate_AE2020,TestV2_2_ShapeEnums_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke composite order,描边合成顺序,composite order
func (s *StrokeNode) SetCompositeOrder(v ShapeCompositeOrder) error {
	return setShapeCompositeOrder(&s.compositeOrder, v)
}

// @summary    Set the stroke end-cap style
// @param      v  the end-cap style (Butt, Round or Projecting)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Stroke_AEShipGate_AE2020,TestV2_2_Stroke_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke line cap,描边端点,line cap,butt,round,projecting
func (s *StrokeNode) SetLineCap(v StrokeLineCap) error {
	if v < StrokeLineCapButt || v > StrokeLineCapProjecting {
		return fmt.Errorf("StrokeNode.SetLineCap: invalid value %d (want 1..3)", v)
	}
	s.lineCap = v
	return nil
}

// @summary    Set the stroke corner-join style
// @param      v  the corner-join style (Miter, Round or Bevel)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Stroke_AEShipGate_AE2020,TestV2_2_Stroke_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke line join,描边接头,line join,miter,round,bevel
func (s *StrokeNode) SetLineJoin(v StrokeLineJoin) error {
	if v < StrokeLineJoinMiter || v > StrokeLineJoinBevel {
		return fmt.Errorf("StrokeNode.SetLineJoin: invalid value %d (want 1..3)", v)
	}
	s.lineJoin = v
	return nil
}

// @summary    Set the stroke miter limit
// @description AE only applies it when the corner join is Miter, but the value
//   is stored regardless.
// @param      v  the miter limit ratio (must be at least one)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_Stroke_AEShipGate_AE2020,TestV2_2_Stroke_AEShipGate_AE2025
// @since      AE2020
// @alias      stroke miter limit,描边斜接限制,miter limit
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

// StrokeTaper models the Stroke "Taper" group's percent-mode scalar controls
// (`ADBE Vector Stroke Taper`): Start/End Length, Start/End Width, Start/End
// Ease — all plain float64 percentages stored on disk as float64 big-endian at
// cdat[0:8]. AE does not animate them, so they are stored as values, not
// PropertyStreams. All default to 0 (no taper).
//
// Only the always-active percent-mode controls are modeled. The Length Units
// enum and the pixel-mode mirror streams are AE-elided at the percent default
// and not modeled.
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

// @summary    Set the taper start length
// @param      v  the taper start length in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      taper start length,渐细起始长度,start taper length
func (t *StrokeTaper) SetStartLength(v float64) error { t.startLength = v; return nil }

// @summary    Set the taper end length
// @param      v  the taper end length in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      taper end length,渐细结束长度,end taper length
func (t *StrokeTaper) SetEndLength(v float64) error { t.endLength = v; return nil }

// @summary    Set the taper start width
// @param      v  the taper start width in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      taper start width,渐细起始宽度,start taper width
func (t *StrokeTaper) SetStartWidth(v float64) error { t.startWidth = v; return nil }

// @summary    Set the taper end width
// @param      v  the taper end width in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      taper end width,渐细结束宽度,end taper width
func (t *StrokeTaper) SetEndWidth(v float64) error { t.endWidth = v; return nil }

// @summary    Set the taper start ease
// @param      v  the taper start ease in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      taper start ease,渐细起始缓动,start ease
func (t *StrokeTaper) SetStartEase(v float64) error { t.startEase = v; return nil }

// @summary    Set the taper end ease
// @param      v  the taper end ease in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      taper end ease,渐细结束缓动,end ease
func (t *StrokeTaper) SetEndEase(v float64) error { t.endEase = v; return nil }

// StrokeWave models the Stroke "Wave" group's wavelength-mode scalars
// (`ADBE Vector Stroke Wave`): Amount (%), Wavelength (px), Phase (deg) — stored
// on disk as float64 big-endian at cdat[0:8]. Defaults: Amount 0, Wavelength
// 100, Phase 0.
//
// Only the wavelength-mode controls are modeled. The Units enum and the Cycles
// stream are AE-elided at the wavelength default and not modeled — Wave is
// always emitted in wavelength mode.
type StrokeWave struct {
	amount     float64
	wavelength float64
	phase      float64
}

func newStrokeWave() *StrokeWave { return &StrokeWave{wavelength: 100} }

func (w *StrokeWave) Amount() float64     { return w.amount }
func (w *StrokeWave) Wavelength() float64 { return w.wavelength }
func (w *StrokeWave) Phase() float64      { return w.phase }

// @summary    Set the stroke wave amount
// @param      v  the wave amount in percent
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      wave amount,波浪幅度,stroke wave amount
func (w *StrokeWave) SetAmount(v float64) error { w.amount = v; return nil }

// @summary    Set the stroke wavelength
// @param      v  the wavelength in pixels
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      wave wavelength,波浪波长,stroke wavelength
func (w *StrokeWave) SetWavelength(v float64) error { w.wavelength = v; return nil }

// @summary    Set the stroke wave phase
// @param      v  the wave phase in degrees
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeTaperWave_AEShipGate_AE2020,TestV2_2_StrokeTaperWave_AEShipGate_AE2025
// @since      AE2020
// @alias      wave phase,波浪相位,stroke wave phase
func (w *StrokeWave) SetPhase(v float64) error { w.phase = v; return nil }

// StrokeDashes models the Stroke "Dashes" group (`ADBE Vector Stroke Dashes`):
// a single Dash + Gap pair, stored on disk as float64 big-endian at cdat[0:8] —
// identical encoding to the other stroke scalars, nested one level deeper inside
// the group's LIST(tdgp). The group is hidden-by-default: a default stroke emits
// the Dashes group as an empty placeholder (solid line). When enabled, the
// serializer swaps to a dashed stroke-body template that carries the Dash 1 /
// Gap 1 slots and overwrites them with these values.
//
// Exactly one Dash + Gap pair is modeled (the common dashed/dotted-line case).
// Additional Dash/Gap pairs and Offset are not modeled. Dash/Gap are not
// animated.
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

// @summary    Enable dashing on the stroke
// @description The serializer emits the Dash and Gap slots.
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeDashes_AEShipGate_AE2020,TestV2_2_StrokeDashes_AEShipGate_AE2025
// @since      AE2020
// @alias      enable dashes,启用虚线,dashed stroke,enable
func (d *StrokeDashes) Enable() { d.enabled = true }

// @summary    Disable dashing on the stroke
// @description Reverts to a solid stroke.
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeDashes_AEShipGate_AE2020,TestV2_2_StrokeDashes_AEShipGate_AE2025
// @since      AE2020
// @alias      disable dashes,禁用虚线,solid stroke,disable
func (d *StrokeDashes) Disable() { d.enabled = false }

// @summary    Set the dash length and enable dashing
// @param      v  the dash length in pixels (must be non-negative)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeDashes_AEShipGate_AE2020,TestV2_2_StrokeDashes_AEShipGate_AE2025
// @since      AE2020
// @alias      dash length,虚线段长度,dash,dashes
func (d *StrokeDashes) SetDash(v float64) error {
	if v < 0 {
		return fmt.Errorf("StrokeDashes.SetDash: %g out of range (want >= 0)", v)
	}
	d.dash = v
	d.enabled = true
	return nil
}

// @summary    Set the gap length and enable dashing
// @param      v  the gap length in pixels (must be non-negative)
// @domain     shape
// @stability  stable
// @verify     render-pixel
// @gate       TestV2_2_StrokeDashes_AEShipGate_AE2020,TestV2_2_StrokeDashes_AEShipGate_AE2025
// @since      AE2020
// @alias      gap length,虚线间距,gap,dashes gap
func (d *StrokeDashes) SetGap(v float64) error {
	if v < 0 {
		return fmt.Errorf("StrokeDashes.SetGap: %g out of range (want >= 0)", v)
	}
	d.gap = v
	d.enabled = true
	return nil
}

// Properties returns the escape-hatch view.
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

// PropertyGroup is the escape-hatch surface. Currently the minimal struct —
// Name and the empty Children / streams maps — so the field exists on
// VectorGroup.Transform / node Properties() but none of the typed lookup methods
// are wired yet.
type PropertyGroup struct {
	Name      string
	Children  map[string]*PropertyGroup
	streams   map[string]any
	Separated bool // dimension separation reserved for nested-group support
}

// newGroupTransform constructs the identity Transform PropertyGroup placeholder
// for a fresh VectorGroup. A RootGroup default-serialized form has no
// `ADBE Vector Transform Group` — this placeholder stays nil-children until the
// escape hatch wires it.
func newGroupTransform() *PropertyGroup {
	return &PropertyGroup{Name: "Transform"}
}

// Child returns the nested PropertyGroup by name, or nil if not present.
// Use for walking deeper-than-leaf escape-hatch trees.
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
