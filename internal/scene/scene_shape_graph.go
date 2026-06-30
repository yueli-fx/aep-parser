// internal/aep/shape_graph.go
package scene

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/codec"
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
