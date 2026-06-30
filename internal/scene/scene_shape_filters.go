// Code moved from scene_shape_graph.go; keep behavior-only edits out of split commits.
package scene

import (
	"fmt"
	"github.com/yueli-fx/aep-parser/internal/codec"
)

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
