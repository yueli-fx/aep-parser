// Code moved from scene_shape_graph.go; keep behavior-only edits out of split commits.
package scene

import (
	"fmt"
	"github.com/yueli-fx/aep-parser/internal/codec"
)

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
