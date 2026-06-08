// internal/aep/shape_graph.go
package aep

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
	ShapeKindRect    ShapeNodeKind = iota // `ADBE Vector Shape - Rect`
	ShapeKindEllipse                      // `ADBE Vector Shape - Ellipse`
	ShapeKindPath                         // `ADBE Vector Shape - Group`
	ShapeKindFill                         // `ADBE Vector Graphic - Fill`
	ShapeKindStroke                       // `ADBE Vector Graphic - Stroke`
	ShapeKindGroup                        // `ADBE Vector Group` (V2.3+ user-created nested group)
	ShapeKindGradientFill                 // `ADBE Vector Graphic - G-Fill`
	// V2.3+ candidates: PolyStar / GradientStroke / Trim / Merge / Repeater /
	// Transform.
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
func (r *RectNode) Size() *codec.PropertyStream[[2]float64]     { return r.size }
func (r *RectNode) Position() *codec.PropertyStream[[2]float64] { return r.position }
func (r *RectNode) Roundness() *codec.PropertyStream[float64]   { return r.roundness }
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
func (e *EllipseNode) Size() *codec.PropertyStream[[2]float64]     { return e.size }
func (e *EllipseNode) Position() *codec.PropertyStream[[2]float64] { return e.position }
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

func (p *PathNode) Kind() ShapeNodeKind                { return ShapeKindPath }
func (p *PathNode) Path() *codec.PropertyStream[BezierPath]  { return p.path }

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

func (f *FillNode) Kind() ShapeNodeKind                  { return ShapeKindFill }
func (f *FillNode) Color() *codec.PropertyStream[[4]float64]   { return f.color }
func (f *FillNode) Opacity() *codec.PropertyStream[float64]    { return f.opacity }
func (f *FillNode) BlendMode() ShapeBlendMode            { return f.blendMode }
func (f *FillNode) CompositeOrder() ShapeCompositeOrder  { return f.compositeOrder }
func (f *FillNode) FillRule() FillRule                   { return f.fillRule }
func (f *FillNode) SetColor(v [4]float64) error          { return f.color.SetStaticValue(v) }
func (f *FillNode) SetOpacity(v float64) error           { return f.opacity.SetStaticValue(v) }
func (f *FillNode) SetBlendMode(v ShapeBlendMode) error  { return setShapeBlendMode(&f.blendMode, v) }
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
// color + alpha stops (`ADBE Vector Grad Colors`), the headline of a gradient
// fill. The ramp geometry (`Grad Type` / `Start Pt` / `End Pt`) is NOT modeled:
// the embedded template was extracted from an AE-saved fixture where those were
// default and therefore elided (no cdat slot to overwrite — same elision trap
// as Stroke Taper Units). AE reconstructs the default linear ramp on open.
//
// Stops are static (V2.2 does not model animated gradients). The serializer
// re-encodes the stops to prop.map XML and overwrites the GCky/Utf8 chunk
// (length-variable; rifx recomputes the enclosing LIST sizes).
type GradientFillNode struct {
	gradient *codec.Gradient
}

// NewGradientFillNode constructs a default 2-stop black→white linear gradient
// (fully opaque). Callers override via SetColorStops / SetAlphaStops.
func NewGradientFillNode() *GradientFillNode {
	return &GradientFillNode{gradient: defaultGradient()}
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
func (n *GradientFillNode) Gradient() *codec.Gradient { return n.gradient }

// SetColorStops replaces the gradient's color stops. Requires ≥ 2 stops; each
// Offset/Midpoint in [0,1] and each Color component in [0,1].
func (n *GradientFillNode) SetColorStops(stops []codec.GradientColorStop) error {
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
	n.gradient.ColorStops = append([]codec.GradientColorStop(nil), stops...)
	return nil
}

// SetAlphaStops replaces the gradient's alpha (opacity) stops. Requires ≥ 2
// stops; each Offset/Midpoint/Alpha in [0,1].
func (n *GradientFillNode) SetAlphaStops(stops []codec.GradientAlphaStop) error {
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
	n.gradient.AlphaStops = append([]codec.GradientAlphaStop(nil), stops...)
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
		color:      codec.NewPropertyStream[[4]float64](),
		opacity:    codec.NewPropertyStream[float64](),
		width:      codec.NewPropertyStream[float64](),
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
func (s *StrokeNode) Color() *codec.PropertyStream[[4]float64] { return s.color }
func (s *StrokeNode) Opacity() *codec.PropertyStream[float64]  { return s.opacity }
func (s *StrokeNode) Width() *codec.PropertyStream[float64]    { return s.width }
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

// Float64Stream returns the codec.PropertyStream[float64] under the given name, or
// an error if no stream by that name exists or it isn't the expected type.
// Mutations on the returned stream are visible through the typed accessor.
func (pg *PropertyGroup) Float64Stream(name string) (*codec.PropertyStream[float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*codec.PropertyStream[float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not float64", pg.Name, name)
	}
	return ps, nil
}

// Vec2Stream returns the codec.PropertyStream[[2]float64] under the given name.
func (pg *PropertyGroup) Vec2Stream(name string) (*codec.PropertyStream[[2]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*codec.PropertyStream[[2]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [2]float64", pg.Name, name)
	}
	return ps, nil
}

// Vec3Stream returns the codec.PropertyStream[[3]float64] under the given name.
func (pg *PropertyGroup) Vec3Stream(name string) (*codec.PropertyStream[[3]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*codec.PropertyStream[[3]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [3]float64", pg.Name, name)
	}
	return ps, nil
}

// ColorStream returns the codec.PropertyStream[[4]float64] (RGBA) under the given name.
func (pg *PropertyGroup) ColorStream(name string) (*codec.PropertyStream[[4]float64], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*codec.PropertyStream[[4]float64])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not [4]float64", pg.Name, name)
	}
	return ps, nil
}

// PathStream returns the codec.PropertyStream[BezierPath] under the given name.
func (pg *PropertyGroup) PathStream(name string) (*codec.PropertyStream[BezierPath], error) {
	v, ok := pg.streams[name]
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q not found", pg.Name, name)
	}
	ps, ok := v.(*codec.PropertyStream[BezierPath])
	if !ok {
		return nil, fmt.Errorf("PropertyGroup %q: stream %q is not BezierPath", pg.Name, name)
	}
	return ps, nil
}
