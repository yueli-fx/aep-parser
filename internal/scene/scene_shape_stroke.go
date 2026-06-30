// Code moved from scene_shape_graph.go; keep behavior-only edits out of split commits.
package scene

import (
	"fmt"
	"github.com/yueli-fx/aep-parser/internal/codec"
)

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
// on disk as float64 big-endian at cdat[0:8]. Defaults: Amount is 0,
// Wavelength is 100, Phase is 0.
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
