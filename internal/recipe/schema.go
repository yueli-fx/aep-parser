package recipe

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const SchemaVersion = 1

type Recipe struct {
	SchemaVersion   int             `json:"schema_version"`
	Project         ProjectSpec     `json:"project"`
	Comps           []CompSpec      `json:"comps"`
	ExpectedProfile ExpectedProfile `json:"expected_profile,omitempty"`
}

type ProjectSpec struct {
	Name string `json:"name,omitempty"`
}

type CompSpec struct {
	Name                     string              `json:"name"`
	Width                    int                 `json:"width"`
	Height                   int                 `json:"height"`
	FrameRate                float64             `json:"frame_rate"`
	Duration                 float64             `json:"duration"`
	BackgroundColor          []float64           `json:"background_color,omitempty"`
	Label                    *float64            `json:"label,omitempty"`
	Comment                  string              `json:"comment,omitempty"`
	Renderer                 string              `json:"renderer,omitempty"`
	ResolutionFactor         []float64           `json:"resolution_factor,omitempty"`
	PixelAspect              *float64            `json:"pixel_aspect,omitempty"`
	DisplayStartTime         *float64            `json:"display_start_time,omitempty"`
	FrameBlending            *bool               `json:"frame_blending,omitempty"`
	Draft3D                  *bool               `json:"draft_3d,omitempty"`
	HideShyLayers            *bool               `json:"hide_shy_layers,omitempty"`
	PreserveNestedFrameRate  *bool               `json:"preserve_nested_frame_rate,omitempty"`
	PreserveNestedResolution *bool               `json:"preserve_nested_resolution,omitempty"`
	MotionBlur               *CompMotionBlurSpec `json:"motion_blur,omitempty"`
	WorkArea                 *CompWorkAreaSpec   `json:"work_area,omitempty"`
	Layers                   []Layer             `json:"layers,omitempty"`
}

type CompWorkAreaSpec struct {
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
}

type CompMotionBlurSpec struct {
	Enabled             *bool    `json:"enabled,omitempty"`
	ShutterAngle        *float64 `json:"shutter_angle,omitempty"`
	ShutterPhase        *float64 `json:"shutter_phase,omitempty"`
	AdaptiveSampleLimit *float64 `json:"adaptive_sample_limit,omitempty"`
	SamplesPerFrame     *float64 `json:"samples_per_frame,omitempty"`
}

type Layer struct {
	Type                  string         `json:"type"`
	Name                  string         `json:"name"`
	Label                 *float64       `json:"label,omitempty"`
	Comment               string         `json:"comment,omitempty"`
	Visible               *bool          `json:"visible,omitempty"`
	Solo                  *bool          `json:"solo,omitempty"`
	Locked                *bool          `json:"locked,omitempty"`
	MotionBlur            *bool          `json:"motion_blur,omitempty"`
	Shy                   *bool          `json:"shy,omitempty"`
	EffectsEnabled        *bool          `json:"effects_enabled,omitempty"`
	AudioEnabled          *bool          `json:"audio_enabled,omitempty"`
	FrameBlendEnabled     *bool          `json:"frame_blend_enabled,omitempty"`
	MarkersLocked         *bool          `json:"markers_locked,omitempty"`
	CollapseTransform     *bool          `json:"collapse_transform,omitempty"`
	Is3D                  *bool          `json:"is_3d,omitempty"`
	IsAdjust              *bool          `json:"is_adjust,omitempty"`
	IsNull                *bool          `json:"is_null,omitempty"`
	IsGuide               *bool          `json:"is_guide,omitempty"`
	SamplingBicubic       *bool          `json:"sampling_bicubic,omitempty"`
	FrameBlendPixelMotion *bool          `json:"frame_blend_pixel_motion,omitempty"`
	PreserveTransparency  *bool          `json:"preserve_transparency,omitempty"`
	Quality               string         `json:"quality,omitempty"`
	BlendingMode          string         `json:"blending_mode,omitempty"`
	TrackMatte            string         `json:"track_matte,omitempty"`
	AutoOrient            string         `json:"auto_orient,omitempty"`
	StartTime             *float64       `json:"start_time,omitempty"`
	InPoint               *float64       `json:"in_point,omitempty"`
	OutPoint              *float64       `json:"out_point,omitempty"`
	Parent                string         `json:"parent,omitempty"`
	Text                  string         `json:"text,omitempty"`
	TextStyle             *TextStyleSpec `json:"text_style,omitempty"`
	Camera                *CameraSpec    `json:"camera,omitempty"`
	Light                 *LightSpec     `json:"light,omitempty"`
	Shape                 *ShapeSpec     `json:"shape,omitempty"`
	Masks                 []MaskSpec     `json:"masks,omitempty"`
	Transform             Transform      `json:"transform,omitempty"`
	Effects               []Effect       `json:"effects,omitempty"`
}

type MaskSpec struct {
	Name           string      `json:"name,omitempty"`
	Mode           string      `json:"mode,omitempty"`
	Inverted       *bool       `json:"inverted,omitempty"`
	Locked         *bool       `json:"locked,omitempty"`
	Color          []float64   `json:"color,omitempty"`
	MotionBlur     string      `json:"motion_blur,omitempty"`
	FeatherFalloff string      `json:"feather_falloff,omitempty"`
	Opacity        *float64    `json:"opacity,omitempty"`
	Feather        []float64   `json:"feather,omitempty"`
	Expansion      *float64    `json:"expansion,omitempty"`
	Closed         *bool       `json:"closed,omitempty"`
	Vertices       [][]float64 `json:"vertices"`
}

type LightSpec struct {
	Kind            string    `json:"kind,omitempty"`
	SourceLayer     string    `json:"source_layer,omitempty"`
	Intensity       *float64  `json:"intensity,omitempty"`
	Color           []float64 `json:"color,omitempty"`
	CastsShadows    *bool     `json:"casts_shadows,omitempty"`
	ShadowDarkness  *float64  `json:"shadow_darkness,omitempty"`
	ShadowDiffusion *float64  `json:"shadow_diffusion,omitempty"`
	FalloffType     *float64  `json:"falloff_type,omitempty"`
	FalloffStart    *float64  `json:"falloff_start,omitempty"`
	FalloffDistance *float64  `json:"falloff_distance,omitempty"`
	ConeAngle       *float64  `json:"cone_angle,omitempty"`
	ConeFeather     *float64  `json:"cone_feather,omitempty"`
}

type CameraSpec struct {
	Zoom                    *float64 `json:"zoom,omitempty"`
	DepthOfField            *bool    `json:"depth_of_field,omitempty"`
	FocusDistance           *float64 `json:"focus_distance,omitempty"`
	Aperture                *float64 `json:"aperture,omitempty"`
	BlurLevel               *float64 `json:"blur_level,omitempty"`
	IrisShape               *float64 `json:"iris_shape,omitempty"`
	IrisRotation            *float64 `json:"iris_rotation,omitempty"`
	IrisRoundness           *float64 `json:"iris_roundness,omitempty"`
	IrisAspectRatio         *float64 `json:"iris_aspect_ratio,omitempty"`
	IrisDiffractionFringe   *float64 `json:"iris_diffraction_fringe,omitempty"`
	IrisHighlightGain       *float64 `json:"iris_highlight_gain,omitempty"`
	IrisHighlightThreshold  *float64 `json:"iris_highlight_threshold,omitempty"`
	IrisHighlightSaturation *float64 `json:"iris_highlight_saturation,omitempty"`
}

type TextStyleSpec struct {
	RunIndex       int       `json:"run_index,omitempty"`
	ParagraphIndex int       `json:"paragraph_index,omitempty"`
	FontSize       *float64  `json:"font_size,omitempty"`
	FillColor      []float64 `json:"fill_color,omitempty"`
	Tracking       *float64  `json:"tracking,omitempty"`
	FauxBold       *bool     `json:"faux_bold,omitempty"`
	FauxItalic     *bool     `json:"faux_italic,omitempty"`
	ApplyStroke    *bool     `json:"apply_stroke,omitempty"`
	StrokeColor    []float64 `json:"stroke_color,omitempty"`
	StrokeWidth    *float64  `json:"stroke_width,omitempty"`
	Justification  string    `json:"justification,omitempty"`
}

type ShapeSpec struct {
	Kind               string               `json:"kind"`
	Size               []float64            `json:"size,omitempty"`
	Position           []float64            `json:"position,omitempty"`
	Roundness          *float64             `json:"roundness,omitempty"`
	Points             *float64             `json:"points,omitempty"`
	Rotation           *float64             `json:"rotation,omitempty"`
	InnerRadius        *float64             `json:"inner_radius,omitempty"`
	OuterRadius        *float64             `json:"outer_radius,omitempty"`
	InnerRoundness     *float64             `json:"inner_roundness,omitempty"`
	OuterRoundness     *float64             `json:"outer_roundness,omitempty"`
	FillColor          []float64            `json:"fill_color,omitempty"`
	FillOpacity        *float64             `json:"fill_opacity,omitempty"`
	FillBlendMode      *float64             `json:"fill_blend_mode,omitempty"`
	FillCompositeOrder string               `json:"fill_composite_order,omitempty"`
	FillRule           string               `json:"fill_rule,omitempty"`
	GradientFill       *GradientFillSpec    `json:"gradient_fill,omitempty"`
	GradientStroke     *GradientStrokeSpec  `json:"gradient_stroke,omitempty"`
	Stroke             *StrokeSpec          `json:"stroke,omitempty"`
	Trim               *TrimSpec            `json:"trim,omitempty"`
	RoundCorners       *RoundCornersSpec    `json:"round_corners,omitempty"`
	OffsetPaths        *OffsetPathsSpec     `json:"offset_paths,omitempty"`
	Repeater           *RepeaterSpec        `json:"repeater,omitempty"`
	MergePaths         *MergePathsSpec      `json:"merge_paths,omitempty"`
	ZigZag             *ZigZagSpec          `json:"zigzag,omitempty"`
	PuckerBloat        *PuckerBloatSpec     `json:"pucker_bloat,omitempty"`
	Twist              *TwistSpec           `json:"twist,omitempty"`
	WigglePaths        *WigglePathsSpec     `json:"wiggle_paths,omitempty"`
	WiggleTransform    *WiggleTransformSpec `json:"wiggle_transform,omitempty"`
}

type GradientFillSpec struct {
	Type            string                  `json:"type,omitempty"`
	StartPoint      []float64               `json:"start_point,omitempty"`
	EndPoint        []float64               `json:"end_point,omitempty"`
	HighlightLength *float64                `json:"highlight_length,omitempty"`
	HighlightAngle  *float64                `json:"highlight_angle,omitempty"`
	ColorStops      []GradientColorStopSpec `json:"color_stops,omitempty"`
	AlphaStops      []GradientAlphaStopSpec `json:"alpha_stops,omitempty"`
}

type GradientColorStopSpec struct {
	Offset   float64   `json:"offset"`
	Midpoint *float64  `json:"midpoint,omitempty"`
	Color    []float64 `json:"color"`
}

type GradientAlphaStopSpec struct {
	Offset   float64  `json:"offset"`
	Midpoint *float64 `json:"midpoint,omitempty"`
	Alpha    float64  `json:"alpha"`
}

type GradientStrokeSpec struct {
	Type            string                  `json:"type,omitempty"`
	StartPoint      []float64               `json:"start_point,omitempty"`
	EndPoint        []float64               `json:"end_point,omitempty"`
	HighlightLength *float64                `json:"highlight_length,omitempty"`
	HighlightAngle  *float64                `json:"highlight_angle,omitempty"`
	Width           *float64                `json:"width,omitempty"`
	LineCap         string                  `json:"line_cap,omitempty"`
	LineJoin        string                  `json:"line_join,omitempty"`
	MiterLimit      *float64                `json:"miter_limit,omitempty"`
	ColorStops      []GradientColorStopSpec `json:"color_stops,omitempty"`
	AlphaStops      []GradientAlphaStopSpec `json:"alpha_stops,omitempty"`
}

type StrokeSpec struct {
	Color          []float64         `json:"color,omitempty"`
	Width          *float64          `json:"width,omitempty"`
	Opacity        *float64          `json:"opacity,omitempty"`
	LineCap        string            `json:"line_cap,omitempty"`
	LineJoin       string            `json:"line_join,omitempty"`
	MiterLimit     *float64          `json:"miter_limit,omitempty"`
	CompositeOrder string            `json:"composite_order,omitempty"`
	Taper          *StrokeTaperSpec  `json:"taper,omitempty"`
	Wave           *StrokeWaveSpec   `json:"wave,omitempty"`
	Dashes         *StrokeDashesSpec `json:"dashes,omitempty"`
}

type StrokeTaperSpec struct {
	StartLength *float64 `json:"start_length,omitempty"`
	EndLength   *float64 `json:"end_length,omitempty"`
	StartWidth  *float64 `json:"start_width,omitempty"`
	EndWidth    *float64 `json:"end_width,omitempty"`
	StartEase   *float64 `json:"start_ease,omitempty"`
	EndEase     *float64 `json:"end_ease,omitempty"`
}

type StrokeWaveSpec struct {
	Amount     *float64 `json:"amount,omitempty"`
	Wavelength *float64 `json:"wavelength,omitempty"`
	Phase      *float64 `json:"phase,omitempty"`
}

type StrokeDashesSpec struct {
	Dash *float64 `json:"dash,omitempty"`
	Gap  *float64 `json:"gap,omitempty"`
}

type TrimSpec struct {
	Start  *float64 `json:"start,omitempty"`
	End    *float64 `json:"end,omitempty"`
	Offset *float64 `json:"offset,omitempty"`
}

type RoundCornersSpec struct {
	Radius *float64 `json:"radius,omitempty"`
}

type OffsetPathsSpec struct {
	Amount     *float64 `json:"amount,omitempty"`
	LineJoin   string   `json:"line_join,omitempty"`
	MiterLimit *float64 `json:"miter_limit,omitempty"`
	Copies     *float64 `json:"copies,omitempty"`
	CopyOffset *float64 `json:"copy_offset,omitempty"`
}

type RepeaterSpec struct {
	Copies       *float64  `json:"copies,omitempty"`
	Offset       *float64  `json:"offset,omitempty"`
	Order        string    `json:"order,omitempty"`
	Anchor       []float64 `json:"anchor,omitempty"`
	Position     []float64 `json:"position,omitempty"`
	Scale        []float64 `json:"scale,omitempty"`
	Rotation     *float64  `json:"rotation,omitempty"`
	StartOpacity *float64  `json:"start_opacity,omitempty"`
	EndOpacity   *float64  `json:"end_opacity,omitempty"`
}

type MergePathsSpec struct {
	Type string `json:"type,omitempty"`
}

type ZigZagSpec struct {
	Size   *float64 `json:"size,omitempty"`
	Detail *float64 `json:"detail,omitempty"`
	Points string   `json:"points,omitempty"`
}

type PuckerBloatSpec struct {
	Amount *float64 `json:"amount,omitempty"`
}

type TwistSpec struct {
	Angle  *float64  `json:"angle,omitempty"`
	Center []float64 `json:"center,omitempty"`
}

type WigglePathsSpec struct {
	Size             *float64 `json:"size,omitempty"`
	Detail           *float64 `json:"detail,omitempty"`
	WigglesPerSecond *float64 `json:"wiggles_per_second,omitempty"`
	RandomSeed       *float64 `json:"random_seed,omitempty"`
	Points           string   `json:"points,omitempty"`
	Correlation      *float64 `json:"correlation,omitempty"`
	TemporalPhase    *float64 `json:"temporal_phase,omitempty"`
	SpatialPhase     *float64 `json:"spatial_phase,omitempty"`
}

type WiggleTransformSpec struct {
	Anchor           []float64 `json:"anchor,omitempty"`
	Position         []float64 `json:"position,omitempty"`
	Scale            []float64 `json:"scale,omitempty"`
	Rotation         *float64  `json:"rotation,omitempty"`
	WigglesPerSecond *float64  `json:"wiggles_per_second,omitempty"`
	RandomSeed       *float64  `json:"random_seed,omitempty"`
	Correlation      *float64  `json:"correlation,omitempty"`
	TemporalPhase    *float64  `json:"temporal_phase,omitempty"`
	SpatialPhase     *float64  `json:"spatial_phase,omitempty"`
}

type Effect struct {
	MatchName string        `json:"match_name"`
	Params    []EffectParam `json:"params,omitempty"`
}

type EffectParam struct {
	MatchName  string          `json:"match_name"`
	Value      any             `json:"value,omitempty"`
	Expression *ExpressionSpec `json:"expression,omitempty"`
}

type Transform struct {
	Position             []float64            `json:"position,omitempty"`
	Scale                []float64            `json:"scale,omitempty"`
	AnchorPoint          []float64            `json:"anchor_point,omitempty"`
	Rotation             *float64             `json:"rotation,omitempty"`
	Opacity              *float64             `json:"opacity,omitempty"`
	PositionKeyframes    []VectorKeyframe     `json:"position_keyframes,omitempty"`
	AnchorPointKeyframes []VectorKeyframe     `json:"anchor_point_keyframes,omitempty"`
	ScaleKeyframes       []VectorKeyframe     `json:"scale_keyframes,omitempty"`
	RotationKeyframes    []ScalarKeyframe     `json:"rotation_keyframes,omitempty"`
	OpacityKeyframes     []ScalarKeyframe     `json:"opacity_keyframes,omitempty"`
	Expressions          TransformExpressions `json:"expressions,omitempty"`
}

type VectorKeyframe struct {
	Time    float64       `json:"time"`
	Value   []float64     `json:"value"`
	InEase  *TemporalEase `json:"in_ease,omitempty"`
	OutEase *TemporalEase `json:"out_ease,omitempty"`
}

type ScalarKeyframe struct {
	Time    float64       `json:"time"`
	Value   float64       `json:"value"`
	InEase  *TemporalEase `json:"in_ease,omitempty"`
	OutEase *TemporalEase `json:"out_ease,omitempty"`
}

type TemporalEase struct {
	Speed     float64 `json:"speed,omitempty"`
	Influence float64 `json:"influence"`
}

type TransformExpressions struct {
	Position    *ExpressionSpec `json:"position,omitempty"`
	AnchorPoint *ExpressionSpec `json:"anchor_point,omitempty"`
	Scale       *ExpressionSpec `json:"scale,omitempty"`
	Rotation    *ExpressionSpec `json:"rotation,omitempty"`
	Opacity     *ExpressionSpec `json:"opacity,omitempty"`
}

type ExpressionSpec struct {
	Source  string `json:"source"`
	Enabled *bool  `json:"enabled,omitempty"`
}

type ExpectedProfile struct {
	CompCount                *int                        `json:"comp_count,omitempty"`
	LayerCount               *int                        `json:"layer_count,omitempty"`
	TextLayerCount           *int                        `json:"text_layer_count,omitempty"`
	ShapeLayerCount          *int                        `json:"shape_layer_count,omitempty"`
	Label                    *float64                    `json:"label,omitempty"`
	Comment                  string                      `json:"comment,omitempty"`
	BackgroundColor          []float64                   `json:"background_color,omitempty"`
	ResolutionFactor         []float64                   `json:"resolution_factor,omitempty"`
	PixelAspect              *float64                    `json:"pixel_aspect,omitempty"`
	DisplayStartTime         *float64                    `json:"display_start_time,omitempty"`
	Renderer                 string                      `json:"renderer,omitempty"`
	Draft3D                  *bool                       `json:"draft_3d,omitempty"`
	FrameBlending            *bool                       `json:"frame_blending,omitempty"`
	HideShyLayers            *bool                       `json:"hide_shy_layers,omitempty"`
	PreserveNestedFrameRate  *bool                       `json:"preserve_nested_frame_rate,omitempty"`
	PreserveNestedResolution *bool                       `json:"preserve_nested_resolution,omitempty"`
	MotionBlur               *ExpectedMotionBlurSpec     `json:"motion_blur,omitempty"`
	WorkArea                 *ExpectedWorkAreaSpec       `json:"work_area,omitempty"`
	Layers                   []ExpectedLayer             `json:"layers,omitempty"`
	Effects                  []ExpectedEffect            `json:"effects,omitempty"`
	Properties               []ExpectedProperty          `json:"properties,omitempty"`
	TextStyles               []ExpectedTextStyle         `json:"text_styles,omitempty"`
	Keyframes                []ExpectedKeyframedProperty `json:"keyframes,omitempty"`
	Masks                    []ExpectedMask              `json:"masks,omitempty"`
}

type ExpectedLayer struct {
	Name         string               `json:"name"`
	Type         string               `json:"type,omitempty"`
	Quality      string               `json:"quality,omitempty"`
	BlendingMode string               `json:"blending_mode,omitempty"`
	AutoOrient   string               `json:"auto_orient,omitempty"`
	LightKind    string               `json:"light_kind,omitempty"`
	Source       string               `json:"source,omitempty"`
	SourceKind   string               `json:"source_kind,omitempty"`
	LightSource  string               `json:"light_source,omitempty"`
	Parent       string               `json:"parent,omitempty"`
	TrackMatte   string               `json:"track_matte,omitempty"`
	Matte        string               `json:"matte,omitempty"`
	Label        *float64             `json:"label,omitempty"`
	Comment      string               `json:"comment,omitempty"`
	Timing       *ExpectedLayerTiming `json:"timing,omitempty"`
	Flags        *ExpectedLayerFlags  `json:"flags,omitempty"`
}

type ExpectedLayerTiming struct {
	StartTime *float64 `json:"start_time,omitempty"`
	InPoint   *float64 `json:"in_point,omitempty"`
	OutPoint  *float64 `json:"out_point,omitempty"`
	Duration  *float64 `json:"duration,omitempty"`
	Stretch   *float64 `json:"stretch,omitempty"`
}

type ExpectedLayerFlags struct {
	Visible               *bool `json:"visible,omitempty"`
	Solo                  *bool `json:"solo,omitempty"`
	Shy                   *bool `json:"shy,omitempty"`
	Locked                *bool `json:"locked,omitempty"`
	Is3D                  *bool `json:"is_3d,omitempty"`
	IsAdjustment          *bool `json:"is_adjustment,omitempty"`
	IsNull                *bool `json:"is_null,omitempty"`
	IsGuide               *bool `json:"is_guide,omitempty"`
	MotionBlur            *bool `json:"motion_blur,omitempty"`
	EffectsEnabled        *bool `json:"effects_enabled,omitempty"`
	AudioEnabled          *bool `json:"audio_enabled,omitempty"`
	FrameBlendEnabled     *bool `json:"frame_blend_enabled,omitempty"`
	MarkersLocked         *bool `json:"markers_locked,omitempty"`
	FrameBlendPixelMotion *bool `json:"frame_blend_pixel_motion,omitempty"`
	CollapseTransform     *bool `json:"collapse_transform,omitempty"`
	SamplingBicubic       *bool `json:"sampling_bicubic,omitempty"`
	PreserveTransparency  *bool `json:"preserve_transparency,omitempty"`
}

type ExpectedWorkAreaSpec struct {
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
}

type ExpectedMotionBlurSpec struct {
	Enabled             *bool    `json:"enabled,omitempty"`
	ShutterAngle        *float64 `json:"shutter_angle,omitempty"`
	ShutterPhase        *float64 `json:"shutter_phase,omitempty"`
	AdaptiveSampleLimit *float64 `json:"adaptive_sample_limit,omitempty"`
	SamplesPerFrame     *float64 `json:"samples_per_frame,omitempty"`
}

type ExpectedProperty struct {
	LayerName  string `json:"layer_name"`
	MatchName  string `json:"match_name"`
	Value      any    `json:"value,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type ExpectedTextStyle struct {
	LayerName      string    `json:"layer_name"`
	RunIndex       int       `json:"run_index,omitempty"`
	ParagraphIndex int       `json:"paragraph_index,omitempty"`
	FontSize       *float64  `json:"font_size,omitempty"`
	FillColor      []float64 `json:"fill_color,omitempty"`
	Tracking       *float64  `json:"tracking,omitempty"`
	FauxBold       *bool     `json:"faux_bold,omitempty"`
	FauxItalic     *bool     `json:"faux_italic,omitempty"`
	ApplyStroke    *bool     `json:"apply_stroke,omitempty"`
	StrokeColor    []float64 `json:"stroke_color,omitempty"`
	StrokeWidth    *float64  `json:"stroke_width,omitempty"`
	Justification  string    `json:"justification,omitempty"`
}

type ExpectedKeyframedProperty struct {
	LayerName string             `json:"layer_name"`
	MatchName string             `json:"match_name"`
	Keyframes []ExpectedKeyframe `json:"keyframes"`
}

type ExpectedKeyframe struct {
	Time  float64 `json:"time"`
	Value any     `json:"value,omitempty"`
}

type ExpectedMask struct {
	LayerName      string    `json:"layer_name"`
	Name           string    `json:"name,omitempty"`
	Mode           string    `json:"mode,omitempty"`
	Inverted       *bool     `json:"inverted,omitempty"`
	Locked         *bool     `json:"locked,omitempty"`
	Color          []float64 `json:"color,omitempty"`
	MotionBlur     string    `json:"motion_blur,omitempty"`
	FeatherFalloff string    `json:"feather_falloff,omitempty"`
	Opacity        *float64  `json:"opacity,omitempty"`
	Feather        []float64 `json:"feather,omitempty"`
	Expansion      *float64  `json:"expansion,omitempty"`
	Closed         *bool     `json:"closed,omitempty"`
	VertexCount    *int      `json:"vertex_count,omitempty"`
}

type ExpectedEffect struct {
	LayerName string                `json:"layer_name"`
	MatchName string                `json:"match_name"`
	Params    []ExpectedEffectParam `json:"params,omitempty"`
}

type ExpectedEffectParam struct {
	MatchName  string `json:"match_name"`
	Value      any    `json:"value,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type Report struct {
	SchemaVersion int             `json:"schema_version"`
	Valid         bool            `json:"valid"`
	OutputPath    string          `json:"output_path,omitempty"`
	Capabilities  []CapabilityUse `json:"capabilities,omitempty"`
	Downgrades    []Downgrade     `json:"downgrades,omitempty"`
	ProfileChecks []ProfileCheck  `json:"profile_checks,omitempty"`
	Refusals      []Refusal       `json:"refusals,omitempty"`
}

type Refusal struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
}

type CapabilityUse struct {
	Path string `json:"path,omitempty"`
	CapabilityLookup
}

type Downgrade struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Query   string `json:"query,omitempty"`
	Message string `json:"message,omitempty"`
}

type ProfileCheck struct {
	Path     string `json:"path"`
	Passed   bool   `json:"passed"`
	Expected any    `json:"expected,omitempty"`
	Actual   any    `json:"actual,omitempty"`
	Message  string `json:"message,omitempty"`
}

func Validate(rec Recipe) Report {
	return ValidateWithCapabilities(rec, unknownCapabilities{})
}

func ValidateWithCapabilities(rec Recipe, caps CapabilityIndex) Report {
	if caps == nil {
		caps = unknownCapabilities{}
	}
	report := Report{SchemaVersion: SchemaVersion, Valid: true}
	addRefusal := func(code, path, message string) {
		report.Valid = false
		report.Refusals = append(report.Refusals, Refusal{Code: code, Path: path, Message: message})
	}
	recordCapability := func(query, path string) CapabilityLookup {
		lookup := caps.Lookup(query)
		if lookup.Query == "" {
			lookup.Query = query
		}
		report.Capabilities = append(report.Capabilities, CapabilityUse{Path: path, CapabilityLookup: lookup})
		switch {
		case lookup.Status == CapabilityUnknown:
			report.Downgrades = append(report.Downgrades, Downgrade{
				Code:    "unknown_capability",
				Path:    path,
				Query:   query,
				Message: fmt.Sprintf("capability %q is not present in the loaded index", query),
			})
		case lookup.Status == CapabilitySupported && lookup.Tier != "" && lookup.Tier != "stable":
			report.Downgrades = append(report.Downgrades, Downgrade{
				Code:    "non_stable_capability",
				Path:    path,
				Query:   query,
				Message: fmt.Sprintf("capability %q is %s", query, lookup.Tier),
			})
		}
		return lookup
	}

	if rec.SchemaVersion != SchemaVersion {
		addRefusal("unsupported_schema_version", "schema_version", fmt.Sprintf("schema_version must be %d", SchemaVersion))
	}
	if len(rec.Comps) == 0 {
		addRefusal("missing_comp", "comps", "at least one comp is required")
		return report
	}
	if len(rec.Comps) > 1 {
		addRefusal("too_many_comps", "comps", "first recipe slice supports exactly one comp")
	}
	validateExpectedProfile(rec.ExpectedProfile, addRefusal)
	for ci, comp := range rec.Comps {
		compPath := fmt.Sprintf("comps[%d]", ci)
		recordCapability("NewComposition", compPath)
		if comp.Name == "" {
			addRefusal("missing_comp_name", compPath+".name", "comp name is required")
		}
		if comp.Width <= 0 || comp.Height <= 0 || comp.FrameRate <= 0 || comp.Duration <= 0 {
			addRefusal("invalid_comp_timing_or_size", compPath, "width, height, frame_rate, and duration must be positive")
		}
		if len(comp.BackgroundColor) > 0 {
			recordCapability("SetBGColor", compPath+".background_color")
			validateRGBColor(comp.BackgroundColor, compPath+".background_color", "invalid_comp_background_color", addRefusal)
		}
		if comp.Label != nil {
			recordCapability("SetLabel", compPath+".label")
			validateCompLabel(*comp.Label, compPath+".label", addRefusal)
		}
		if comp.Comment != "" {
			recordCapability("SetComment", compPath+".comment")
		}
		if comp.Renderer != "" {
			recordCapability("SetRenderer", compPath+".renderer")
		}
		if len(comp.ResolutionFactor) > 0 {
			recordCapability("SetResolutionFactor", compPath+".resolution_factor")
			validateResolutionFactor(comp.ResolutionFactor, compPath+".resolution_factor", addRefusal)
		}
		if comp.PixelAspect != nil {
			recordCapability("SetPixelAspect", compPath+".pixel_aspect")
			validatePixelAspect(*comp.PixelAspect, compPath+".pixel_aspect", addRefusal)
		}
		if comp.DisplayStartTime != nil {
			recordCapability("SetDisplayStartTime", compPath+".display_start_time")
			validateDisplayStartTime(*comp.DisplayStartTime, compPath+".display_start_time", addRefusal)
		}
		if comp.FrameBlending != nil {
			recordCapability("SetFrameBlending", compPath+".frame_blending")
		}
		if comp.Draft3D != nil {
			recordCapability("SetDraft3D", compPath+".draft_3d")
		}
		if comp.HideShyLayers != nil {
			recordCapability("SetHideShyLayers", compPath+".hide_shy_layers")
		}
		if comp.PreserveNestedFrameRate != nil {
			recordCapability("SetPreserveNestedFrameRate", compPath+".preserve_nested_frame_rate")
		}
		if comp.PreserveNestedResolution != nil {
			recordCapability("SetPreserveNestedResolution", compPath+".preserve_nested_resolution")
		}
		if comp.MotionBlur != nil {
			validateCompMotionBlur(comp.MotionBlur, compPath+".motion_blur", recordCapability, addRefusal)
		}
		if comp.WorkArea != nil {
			validateCompWorkArea(comp.WorkArea, compPath+".work_area", comp.Duration, recordCapability, addRefusal)
		}
		layerNames := map[string]bool{}
		for _, layer := range comp.Layers {
			if layer.Name != "" {
				layerNames[layer.Name] = true
			}
		}
		for li, layer := range comp.Layers {
			layerPath := fmt.Sprintf("%s.layers[%d]", compPath, li)
			validateLayer(layer, layerPath, comp.Duration, recordCapability, addRefusal)
			if layer.Parent != "" {
				recordCapability("Layer.SetParent", layerPath+".parent")
				if !layerNames[layer.Parent] {
					addRefusal("unknown_layer_parent", layerPath+".parent", fmt.Sprintf("parent layer %q was not found in the comp", layer.Parent))
				}
			}
			if layer.Light != nil && layer.Light.SourceLayer != "" && !layerNames[layer.Light.SourceLayer] {
				addRefusal("unknown_light_source", layerPath+".light.source_layer", fmt.Sprintf("source layer %q was not found in the comp", layer.Light.SourceLayer))
			}
		}
	}
	return report
}

func validateExpectedProfile(expected ExpectedProfile, addRefusal func(string, string, string)) {
	if expected.CompCount != nil && *expected.CompCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.comp_count", "expected count must be non-negative")
	}
	if expected.LayerCount != nil && *expected.LayerCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.layer_count", "expected count must be non-negative")
	}
	if expected.TextLayerCount != nil && *expected.TextLayerCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.text_layer_count", "expected count must be non-negative")
	}
	if expected.ShapeLayerCount != nil && *expected.ShapeLayerCount < 0 {
		addRefusal("invalid_expected_profile", "expected_profile.shape_layer_count", "expected count must be non-negative")
	}
	if expected.Label != nil {
		validateCompLabel(*expected.Label, "expected_profile.label", addRefusal)
	}
	if len(expected.BackgroundColor) > 0 {
		validateRGBColor(expected.BackgroundColor, "expected_profile.background_color", "invalid_expected_profile", addRefusal)
	}
	if len(expected.ResolutionFactor) > 0 {
		validateResolutionFactor(expected.ResolutionFactor, "expected_profile.resolution_factor", addRefusal)
	}
	if expected.PixelAspect != nil {
		validatePixelAspect(*expected.PixelAspect, "expected_profile.pixel_aspect", addRefusal)
	}
	if expected.DisplayStartTime != nil {
		validateDisplayStartTime(*expected.DisplayStartTime, "expected_profile.display_start_time", addRefusal)
	}
	for i, layer := range expected.Layers {
		layerPath := fmt.Sprintf("expected_profile.layers[%d]", i)
		if layer.Name == "" {
			addRefusal("invalid_expected_profile", layerPath+".name", "layer name is required")
		}
		if layer.Label != nil {
			validateLayerLabel(*layer.Label, layerPath+".label", addRefusal)
		}
		if layer.Quality != "" {
			if _, err := layerQuality(layer.Quality); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".quality", "quality must be wireframe, draft, or best")
			}
		}
		if layer.BlendingMode != "" {
			if _, err := layerBlendingMode(layer.BlendingMode); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".blending_mode", "blending_mode is not supported")
			}
		}
		if layer.TrackMatte != "" {
			if _, err := layerTrackMatte(layer.TrackMatte); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".track_matte", "track_matte must be none, alpha, alpha_inverse, luma, or luma_inverse")
			}
		}
		if layer.AutoOrient != "" {
			if _, err := layerAutoOrient(layer.AutoOrient); err != nil {
				addRefusal("invalid_expected_profile", layerPath+".auto_orient", "auto_orient must be none, along_path, camera_or_point_of_interest, or characters_toward_camera")
			}
		}
		if layer.Timing != nil {
			validateExpectedLayerTiming(layer.Timing, layerPath+".timing", addRefusal)
		}
	}
	for i, effect := range expected.Effects {
		effectPath := fmt.Sprintf("expected_profile.effects[%d]", i)
		if effect.LayerName == "" {
			addRefusal("invalid_expected_profile", effectPath+".layer_name", "layer_name is required")
		}
		if effect.MatchName == "" {
			addRefusal("invalid_expected_profile", effectPath+".match_name", "effect match_name is required")
		}
		for pi, param := range effect.Params {
			paramPath := fmt.Sprintf("%s.params[%d]", effectPath, pi)
			if param.MatchName == "" {
				addRefusal("invalid_expected_profile", paramPath+".match_name", "param match_name is required")
			}
			if param.Expression == "" && param.Value == nil {
				addRefusal("invalid_expected_profile", paramPath, "expected param value or expression is required")
			}
			if param.Value != nil && !validEffectParamValue(param.Value) {
				addRefusal("invalid_expected_profile", paramPath+".value", "expected param value must be a number, boolean, or numeric array")
			}
		}
	}
	for i, prop := range expected.Properties {
		propPath := fmt.Sprintf("expected_profile.properties[%d]", i)
		if prop.LayerName == "" {
			addRefusal("invalid_expected_profile", propPath+".layer_name", "layer_name is required")
		}
		if prop.MatchName == "" {
			addRefusal("invalid_expected_profile", propPath+".match_name", "property match_name is required")
		}
		if prop.Expression == "" && prop.Value == nil {
			addRefusal("invalid_expected_profile", propPath, "expected property value or expression is required")
		}
		if prop.Value != nil && !validEffectParamValue(prop.Value) {
			addRefusal("invalid_expected_profile", propPath+".value", "expected property value must be a number, boolean, or numeric array")
		}
	}
	for i, style := range expected.TextStyles {
		stylePath := fmt.Sprintf("expected_profile.text_styles[%d]", i)
		if style.LayerName == "" {
			addRefusal("invalid_expected_profile", stylePath+".layer_name", "layer_name is required")
		}
		if style.RunIndex < 0 {
			addRefusal("invalid_expected_profile", stylePath+".run_index", "run_index must be non-negative")
		}
		if style.ParagraphIndex < 0 {
			addRefusal("invalid_expected_profile", stylePath+".paragraph_index", "paragraph_index must be non-negative")
		}
		if style.FontSize != nil && *style.FontSize <= 0 {
			addRefusal("invalid_expected_profile", stylePath+".font_size", "font_size must be positive")
		}
		if len(style.FillColor) > 0 {
			validateColor(style.FillColor, stylePath+".fill_color", "invalid_expected_profile", addRefusal)
		}
		if len(style.StrokeColor) > 0 {
			validateColor(style.StrokeColor, stylePath+".stroke_color", "invalid_expected_profile", addRefusal)
		}
		if style.StrokeWidth != nil && *style.StrokeWidth < 0 {
			addRefusal("invalid_expected_profile", stylePath+".stroke_width", "stroke_width must be non-negative")
		}
		if style.Justification != "" && !validTextJustification(style.Justification) {
			addRefusal("invalid_expected_profile", stylePath+".justification", "justification must be left, right, or center")
		}
	}
	for i, keyframed := range expected.Keyframes {
		kfPropPath := fmt.Sprintf("expected_profile.keyframes[%d]", i)
		if keyframed.LayerName == "" {
			addRefusal("invalid_expected_profile", kfPropPath+".layer_name", "layer_name is required")
		}
		if keyframed.MatchName == "" {
			addRefusal("invalid_expected_profile", kfPropPath+".match_name", "property match_name is required")
		}
		if len(keyframed.Keyframes) == 0 {
			addRefusal("invalid_expected_profile", kfPropPath+".keyframes", "at least one keyframe is required")
		}
		for ki, kf := range keyframed.Keyframes {
			kfPath := fmt.Sprintf("%s.keyframes[%d]", kfPropPath, ki)
			if kf.Time < 0 {
				addRefusal("invalid_expected_profile", kfPath+".time", "keyframe time must be non-negative")
			}
			if ki > 0 && kf.Time < keyframed.Keyframes[ki-1].Time {
				addRefusal("invalid_expected_profile", kfPath+".time", "keyframes must be sorted by time")
			}
			if !validEffectParamValue(kf.Value) {
				addRefusal("invalid_expected_profile", kfPath+".value", "expected keyframe value must be a number, boolean, or numeric array")
			}
		}
	}
	for i, mask := range expected.Masks {
		maskPath := fmt.Sprintf("expected_profile.masks[%d]", i)
		if mask.LayerName == "" {
			addRefusal("invalid_expected_profile", maskPath+".layer_name", "layer_name is required")
		}
		if mask.Mode != "" && !validMaskMode(mask.Mode) {
			addRefusal("invalid_expected_profile", maskPath+".mode", "mask mode is not supported")
		}
		if len(mask.Color) > 0 {
			validateRGBColor(mask.Color, maskPath+".color", "invalid_expected_profile", addRefusal)
		}
		if mask.MotionBlur != "" && !validMaskMotionBlur(mask.MotionBlur) {
			addRefusal("invalid_expected_profile", maskPath+".motion_blur", "mask motion_blur is not supported")
		}
		if mask.FeatherFalloff != "" && !validMaskFeatherFalloff(mask.FeatherFalloff) {
			addRefusal("invalid_expected_profile", maskPath+".feather_falloff", "mask feather_falloff is not supported")
		}
		if mask.Opacity != nil && (*mask.Opacity < 0 || *mask.Opacity > 1) {
			addRefusal("invalid_expected_profile", maskPath+".opacity", "mask opacity must be between 0 and 1")
		}
		validateMaskFeather(mask.Feather, maskPath+".feather", "invalid_expected_profile", addRefusal)
		if mask.VertexCount != nil && *mask.VertexCount < 0 {
			addRefusal("invalid_expected_profile", maskPath+".vertex_count", "vertex_count must be non-negative")
		}
	}
}

func validateExpectedLayerTiming(timing *ExpectedLayerTiming, path string, addRefusal func(string, string, string)) {
	if timing.StartTime != nil && *timing.StartTime < 0 {
		addRefusal("invalid_expected_profile", path+".start_time", "start_time must be non-negative")
	}
	if timing.InPoint != nil && *timing.InPoint < 0 {
		addRefusal("invalid_expected_profile", path+".in_point", "in_point must be non-negative")
	}
	if timing.OutPoint != nil && *timing.OutPoint < 0 {
		addRefusal("invalid_expected_profile", path+".out_point", "out_point must be non-negative")
	}
	if timing.InPoint != nil && timing.OutPoint != nil && *timing.OutPoint < *timing.InPoint {
		addRefusal("invalid_expected_profile", path+".out_point", "out_point must be greater than or equal to in_point")
	}
	if timing.Duration != nil && *timing.Duration < 0 {
		addRefusal("invalid_expected_profile", path+".duration", "duration must be non-negative")
	}
}

func validateCompMotionBlur(spec *CompMotionBlurSpec, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	if spec.Enabled != nil {
		recordCapability("SetCompMotionBlur", path+".enabled")
	}
	if spec.ShutterAngle != nil {
		recordCapability("SetShutterAngle", path+".shutter_angle")
		if *spec.ShutterAngle < 0 || *spec.ShutterAngle > 720 || !isWholeNumber(*spec.ShutterAngle) {
			addRefusal("invalid_comp_motion_blur_shutter_angle", path+".shutter_angle", "motion_blur shutter_angle must be an integer between 0 and 720")
		}
	}
	if spec.ShutterPhase != nil {
		recordCapability("SetShutterPhase", path+".shutter_phase")
		if !isWholeNumber(*spec.ShutterPhase) {
			addRefusal("invalid_comp_motion_blur_shutter_phase", path+".shutter_phase", "motion_blur shutter_phase must be an integer")
		}
	}
	if spec.AdaptiveSampleLimit != nil {
		recordCapability("SetMotionBlurAdaptiveSampleLimit", path+".adaptive_sample_limit")
		if *spec.AdaptiveSampleLimit < 0 || !isWholeNumber(*spec.AdaptiveSampleLimit) {
			addRefusal("invalid_comp_motion_blur_adaptive_sample_limit", path+".adaptive_sample_limit", "motion_blur adaptive_sample_limit must be a non-negative integer")
		}
	}
	if spec.SamplesPerFrame != nil {
		recordCapability("SetMotionBlurSamplesPerFrame", path+".samples_per_frame")
		if *spec.SamplesPerFrame < 0 || !isWholeNumber(*spec.SamplesPerFrame) {
			addRefusal("invalid_comp_motion_blur_samples_per_frame", path+".samples_per_frame", "motion_blur samples_per_frame must be a non-negative integer")
		}
	}
}

func validateCompWorkArea(spec *CompWorkAreaSpec, path string, compDuration float64, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	recordCapability("SetWorkArea", path)
	if spec.Start == nil || spec.End == nil {
		addRefusal("invalid_comp_work_area", path, "work_area start and end are required")
		return
	}
	if *spec.Start < 0 || *spec.End < *spec.Start || *spec.End > compDuration {
		addRefusal("invalid_comp_work_area", path, "work_area must satisfy 0 <= start <= end <= comp duration")
	}
}

func validateResolutionFactor(values []float64, path string, addRefusal func(string, string, string)) {
	if len(values) != 2 {
		addRefusal("invalid_comp_resolution_factor", path, "resolution_factor must have exactly two values")
		return
	}
	for i, value := range values {
		if value <= 0 || value > 65535 || !isWholeNumber(value) {
			addRefusal("invalid_comp_resolution_factor", fmt.Sprintf("%s[%d]", path, i), "resolution_factor values must be positive integers in uint16 range")
		}
	}
}

func validatePixelAspect(value float64, path string, addRefusal func(string, string, string)) {
	if value <= 0 {
		addRefusal("invalid_comp_pixel_aspect", path, "pixel_aspect must be positive")
	}
}

func validateDisplayStartTime(value float64, path string, addRefusal func(string, string, string)) {
	if value < 0 {
		addRefusal("invalid_comp_display_start_time", path, "display_start_time must be non-negative")
	}
}

func validateCompLabel(value float64, path string, addRefusal func(string, string, string)) {
	if value < 0 || value > 16 || !isWholeNumber(value) {
		addRefusal("invalid_comp_label", path, "label must be an integer between 0 and 16")
	}
}

func validateLayerLabel(value float64, path string, addRefusal func(string, string, string)) {
	if value < 0 || value > 16 || !isWholeNumber(value) {
		addRefusal("invalid_layer_label", path, "label must be an integer between 0 and 16")
	}
}

func validateLayer(layer Layer, layerPath string, compDuration float64, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	switch layer.Type {
	case "solid":
		recordCapability("NewSolidLayer", layerPath)
	case "text":
		recordCapability("NewTextLayer", layerPath)
		if layer.Text != "" {
			recordCapability("Layer.SetText", layerPath+".text")
		}
	case "shape":
		recordCapability("NewShapeLayer", layerPath)
	case "null":
		recordCapability("NewNullLayer", layerPath)
	case "adjustment":
		recordCapability("NewAdjustmentLayer", layerPath)
	case "camera":
		recordCapability("NewCameraLayer", layerPath)
	case "light":
		recordCapability("NewLightLayer", layerPath)
	default:
		addRefusal("unsupported_layer_type", layerPath+".type", fmt.Sprintf("unsupported layer type %q", layer.Type))
	}
	if layer.Name == "" {
		addRefusal("missing_layer_name", layerPath+".name", "layer name is required")
	}
	if layer.Label != nil {
		recordCapability("Layer.SetLabel", layerPath+".label")
		validateLayerLabel(*layer.Label, layerPath+".label", addRefusal)
	}
	if layer.Comment != "" {
		recordCapability("Layer.SetComment", layerPath+".comment")
	}
	if layer.Visible != nil {
		recordCapability("Layer.SetVisible", layerPath+".visible")
	}
	if layer.Solo != nil {
		recordCapability("Layer.SetSolo", layerPath+".solo")
	}
	if layer.Locked != nil {
		recordCapability("Layer.SetLocked", layerPath+".locked")
	}
	if layer.MotionBlur != nil {
		recordCapability("Layer.SetMotionBlur", layerPath+".motion_blur")
	}
	if layer.Shy != nil {
		recordCapability("Layer.SetShy", layerPath+".shy")
	}
	if layer.EffectsEnabled != nil {
		recordCapability("Layer.SetEffectsEnabled", layerPath+".effects_enabled")
	}
	if layer.AudioEnabled != nil {
		recordCapability("Layer.SetAudioEnabled", layerPath+".audio_enabled")
	}
	if layer.FrameBlendEnabled != nil {
		recordCapability("Layer.SetFrameBlendEnabled", layerPath+".frame_blend_enabled")
	}
	if layer.MarkersLocked != nil {
		recordCapability("Layer.SetMarkersLocked", layerPath+".markers_locked")
	}
	if layer.CollapseTransform != nil {
		recordCapability("Layer.SetCollapseTransform", layerPath+".collapse_transform")
	}
	if layer.Is3D != nil {
		recordCapability("Layer.SetIs3D", layerPath+".is_3d")
	}
	if layer.IsAdjust != nil {
		recordCapability("Layer.SetIsAdjust", layerPath+".is_adjust")
	}
	if layer.IsNull != nil {
		recordCapability("Layer.SetIsNull", layerPath+".is_null")
	}
	if layer.IsGuide != nil {
		recordCapability("Layer.SetIsGuide", layerPath+".is_guide")
	}
	if layer.SamplingBicubic != nil {
		recordCapability("Layer.SetSamplingBicubic", layerPath+".sampling_bicubic")
	}
	if layer.FrameBlendPixelMotion != nil {
		recordCapability("Layer.SetFrameBlendPixelMotion", layerPath+".frame_blend_pixel_motion")
	}
	if layer.PreserveTransparency != nil {
		recordCapability("Layer.SetPreserveTransparency", layerPath+".preserve_transparency")
	}
	if layer.Quality != "" {
		recordCapability("Layer.SetQuality", layerPath+".quality")
		if _, err := layerQuality(layer.Quality); err != nil {
			addRefusal("invalid_layer_quality", layerPath+".quality", "quality must be wireframe, draft, or best")
		}
	}
	if layer.BlendingMode != "" {
		recordCapability("Layer.SetBlendingMode", layerPath+".blending_mode")
		if _, err := layerBlendingMode(layer.BlendingMode); err != nil {
			addRefusal("invalid_layer_blending_mode", layerPath+".blending_mode", "blending_mode is not supported")
		}
	}
	if layer.TrackMatte != "" {
		recordCapability("Layer.SetTrackMatte", layerPath+".track_matte")
		if _, err := layerTrackMatte(layer.TrackMatte); err != nil {
			addRefusal("invalid_layer_track_matte", layerPath+".track_matte", "track_matte must be none, alpha, alpha_inverse, luma, or luma_inverse")
		}
	}
	if layer.AutoOrient != "" {
		recordCapability("Layer.SetAutoOrient", layerPath+".auto_orient")
		if _, err := layerAutoOrient(layer.AutoOrient); err != nil {
			addRefusal("invalid_layer_auto_orient", layerPath+".auto_orient", "auto_orient must be none, along_path, camera_or_point_of_interest, or characters_toward_camera")
		}
	}
	if layer.Camera != nil {
		if layer.Type != "camera" {
			addRefusal("camera_options_on_non_camera_layer", layerPath+".camera", "camera options require type camera")
		}
		if layer.Camera.Zoom != nil {
			recordCapability("Layer.SetCameraZoom", layerPath+".camera.zoom")
		}
		if layer.Camera.DepthOfField != nil {
			recordCapability("Layer.SetCameraDepthOfField", layerPath+".camera.depth_of_field")
		}
		if layer.Camera.FocusDistance != nil {
			recordCapability("Layer.SetCameraFocusDistance", layerPath+".camera.focus_distance")
		}
		if layer.Camera.Aperture != nil {
			recordCapability("Layer.SetCameraAperture", layerPath+".camera.aperture")
		}
		if layer.Camera.BlurLevel != nil {
			recordCapability("Layer.SetCameraBlurLevel", layerPath+".camera.blur_level")
		}
		if layer.Camera.IrisShape != nil {
			recordCapability("SetIrisShape", layerPath+".camera.iris_shape")
		}
		if layer.Camera.IrisRotation != nil {
			recordCapability("SetIrisRotation", layerPath+".camera.iris_rotation")
		}
		if layer.Camera.IrisRoundness != nil {
			recordCapability("SetIrisRoundness", layerPath+".camera.iris_roundness")
		}
		if layer.Camera.IrisAspectRatio != nil {
			recordCapability("SetIrisAspectRatio", layerPath+".camera.iris_aspect_ratio")
		}
		if layer.Camera.IrisDiffractionFringe != nil {
			recordCapability("SetIrisDiffractionFringe", layerPath+".camera.iris_diffraction_fringe")
		}
		if layer.Camera.IrisHighlightGain != nil {
			recordCapability("SetIrisHighlightGain", layerPath+".camera.iris_highlight_gain")
		}
		if layer.Camera.IrisHighlightThreshold != nil {
			recordCapability("SetIrisHighlightThreshold", layerPath+".camera.iris_highlight_threshold")
		}
		if layer.Camera.IrisHighlightSaturation != nil {
			recordCapability("SetIrisHighlightSaturation", layerPath+".camera.iris_highlight_saturation")
		}
	}
	if layer.Light != nil {
		if layer.Type != "light" {
			addRefusal("light_options_on_non_light_layer", layerPath+".light", "light options require type light")
		}
		if layer.Light.Kind != "" {
			recordCapability("SetLightKind", layerPath+".light.kind")
		}
		if layer.Light.SourceLayer != "" {
			recordCapability("SetLightSource", layerPath+".light.source_layer")
		}
		if layer.Light.Intensity != nil {
			recordCapability("SetLightIntensity", layerPath+".light.intensity")
		}
		if len(layer.Light.Color) > 0 {
			recordCapability("SetLightColor", layerPath+".light.color")
			validateColor(layer.Light.Color, layerPath+".light.color", "invalid_light_color", addRefusal)
		}
		if layer.Light.CastsShadows != nil {
			recordCapability("SetLightCastsShadows", layerPath+".light.casts_shadows")
		}
		if layer.Light.ShadowDarkness != nil {
			recordCapability("SetLightShadowDarkness", layerPath+".light.shadow_darkness")
		}
		if layer.Light.ShadowDiffusion != nil {
			recordCapability("SetLightShadowDiffusion", layerPath+".light.shadow_diffusion")
		}
		if layer.Light.FalloffType != nil {
			recordCapability("SetLightFalloffType", layerPath+".light.falloff_type")
		}
		if layer.Light.FalloffStart != nil {
			recordCapability("SetLightFalloffStart", layerPath+".light.falloff_start")
		}
		if layer.Light.FalloffDistance != nil {
			recordCapability("SetLightFalloffDistance", layerPath+".light.falloff_distance")
		}
		if layer.Light.ConeAngle != nil {
			recordCapability("SetLightConeAngle", layerPath+".light.cone_angle")
		}
		if layer.Light.ConeFeather != nil {
			recordCapability("SetLightConeFeather", layerPath+".light.cone_feather")
		}
	}
	if layer.StartTime != nil {
		recordCapability("Layer.SetStartTime", layerPath+".start_time")
	}
	if layer.InPoint != nil {
		recordCapability("Layer.SetInPoint", layerPath+".in_point")
		if *layer.InPoint < 0 || *layer.InPoint > compDuration {
			addRefusal("invalid_layer_in_point", layerPath+".in_point", "in_point must be between 0 and comp duration")
		}
	}
	if layer.OutPoint != nil {
		recordCapability("Layer.SetOutPoint", layerPath+".out_point")
		if *layer.OutPoint < 0 || *layer.OutPoint > compDuration {
			addRefusal("invalid_layer_out_point", layerPath+".out_point", "out_point must be between 0 and comp duration")
		}
		if layer.InPoint != nil && *layer.OutPoint < *layer.InPoint {
			addRefusal("invalid_layer_out_point", layerPath+".out_point", "out_point must be greater than or equal to in_point")
		}
	}
	if layer.TextStyle != nil {
		if layer.Type != "text" {
			addRefusal("text_style_on_non_text_layer", layerPath+".text_style", "text_style is only supported on text layers")
		} else {
			validateTextStyle(*layer.TextStyle, layerPath+".text_style", recordCapability, addRefusal)
		}
	}
	if layer.Type == "shape" && layer.Shape != nil {
		switch layer.Shape.Kind {
		case "rect", "ellipse":
			if layer.Shape.Kind == "rect" {
				recordCapability("RectNode.SetSize", layerPath+".shape.size")
				if len(layer.Shape.Position) > 0 {
					recordCapability("RectNode.SetPosition", layerPath+".shape.position")
				}
				if layer.Shape.Roundness != nil {
					recordCapability("RectNode.SetRoundness", layerPath+".shape.roundness")
					if *layer.Shape.Roundness < 0 {
						addRefusal("invalid_shape_roundness", layerPath+".shape.roundness", "shape roundness must be non-negative")
					}
				}
			} else {
				recordCapability("EllipseNode.SetSize", layerPath+".shape.size")
				if len(layer.Shape.Position) > 0 {
					recordCapability("EllipseNode.SetPosition", layerPath+".shape.position")
				}
				if layer.Shape.Roundness != nil {
					addRefusal("unsupported_shape_roundness", layerPath+".shape.roundness", "roundness is only supported for rect shapes")
				}
			}
		case "star", "polygon":
			recordCapability("VectorGroup.AddStar", layerPath+".shape.kind")
			if layer.Shape.Kind == "polygon" {
				recordCapability("StarNode.SetStarType", layerPath+".shape.kind")
			}
			if layer.Shape.Points != nil {
				recordCapability("StarNode.SetPoints", layerPath+".shape.points")
				if *layer.Shape.Points < 3 {
					addRefusal("invalid_shape_star_points", layerPath+".shape.points", "star points must be at least 3")
				}
			}
			if len(layer.Shape.Position) > 0 {
				recordCapability("StarNode.SetPosition", layerPath+".shape.position")
			}
			if layer.Shape.Rotation != nil {
				recordCapability("StarNode.SetRotation", layerPath+".shape.rotation")
			}
			if layer.Shape.InnerRadius != nil {
				recordCapability("StarNode.SetInnerRadius", layerPath+".shape.inner_radius")
				if *layer.Shape.InnerRadius < 0 {
					addRefusal("invalid_shape_star_inner_radius", layerPath+".shape.inner_radius", "star inner_radius must be non-negative")
				}
			}
			if layer.Shape.OuterRadius != nil {
				recordCapability("StarNode.SetOuterRadius", layerPath+".shape.outer_radius")
				if *layer.Shape.OuterRadius < 0 {
					addRefusal("invalid_shape_star_outer_radius", layerPath+".shape.outer_radius", "star outer_radius must be non-negative")
				}
			}
			if layer.Shape.InnerRoundness != nil {
				recordCapability("StarNode.SetInnerRoundness", layerPath+".shape.inner_roundness")
			}
			if layer.Shape.OuterRoundness != nil {
				recordCapability("StarNode.SetOuterRoundness", layerPath+".shape.outer_roundness")
			}
			if layer.Shape.Roundness != nil {
				addRefusal("unsupported_shape_roundness", layerPath+".shape.roundness", "roundness is only supported for rect shapes")
			}
		default:
			addRefusal("unsupported_shape_kind", layerPath+".shape.kind", fmt.Sprintf("unsupported shape kind %q", layer.Shape.Kind))
		}
		validateVec(layer.Shape.Position, 2, layerPath+".shape.position", addRefusal)
		if len(layer.Shape.FillColor) > 0 {
			recordCapability("FillNode.SetColor", layerPath+".shape.fill_color")
		}
		if layer.Shape.FillOpacity != nil {
			recordCapability("FillNode.SetOpacity", layerPath+".shape.fill_opacity")
			if *layer.Shape.FillOpacity < 0 || *layer.Shape.FillOpacity > 100 {
				addRefusal("invalid_shape_fill_opacity", layerPath+".shape.fill_opacity", "fill opacity must be between 0 and 100")
			}
		}
		if layer.Shape.FillBlendMode != nil {
			recordCapability("FillNode.SetBlendMode", layerPath+".shape.fill_blend_mode")
			if !validShapeBlendMode(*layer.Shape.FillBlendMode) {
				addRefusal("invalid_shape_fill_blend_mode", layerPath+".shape.fill_blend_mode", "fill_blend_mode must be an integer of at least 1")
			}
		}
		if layer.Shape.FillCompositeOrder != "" {
			recordCapability("FillNode.SetCompositeOrder", layerPath+".shape.fill_composite_order")
			if !validShapeCompositeOrder(layer.Shape.FillCompositeOrder) {
				addRefusal("invalid_shape_fill_composite_order", layerPath+".shape.fill_composite_order", "fill_composite_order must be above_previous or below_previous")
			}
		}
		if layer.Shape.FillRule != "" {
			recordCapability("FillNode.SetFillRule", layerPath+".shape.fill_rule")
			if !validFillRule(layer.Shape.FillRule) {
				addRefusal("invalid_shape_fill_rule", layerPath+".shape.fill_rule", "fill_rule must be nonzero_winding or even_odd")
			}
		}
		if layer.Shape.GradientFill != nil {
			gradientPath := layerPath + ".shape.gradient_fill"
			recordCapability("VectorGroup.AddGradientFill", gradientPath)
			if layer.Shape.GradientFill.Type != "" {
				recordCapability("GradientFillNode.SetGradientType", gradientPath+".type")
				if !validGradientType(layer.Shape.GradientFill.Type) {
					addRefusal("invalid_shape_gradient_fill_type", gradientPath+".type", "gradient_fill type must be linear or radial")
				}
			}
			if len(layer.Shape.GradientFill.StartPoint) > 0 {
				recordCapability("GradientFillNode.SetStartPoint", gradientPath+".start_point")
				validateVec(layer.Shape.GradientFill.StartPoint, 2, gradientPath+".start_point", addRefusal)
			}
			if len(layer.Shape.GradientFill.EndPoint) > 0 {
				recordCapability("GradientFillNode.SetEndPoint", gradientPath+".end_point")
				validateVec(layer.Shape.GradientFill.EndPoint, 2, gradientPath+".end_point", addRefusal)
			}
			if layer.Shape.GradientFill.HighlightLength != nil {
				recordCapability("GradientFillNode.SetHighlightLength", gradientPath+".highlight_length")
				if *layer.Shape.GradientFill.HighlightLength < -100 || *layer.Shape.GradientFill.HighlightLength > 100 {
					addRefusal("invalid_shape_gradient_fill_highlight_length", gradientPath+".highlight_length", "gradient_fill highlight_length must be between -100 and 100")
				}
			}
			if layer.Shape.GradientFill.HighlightAngle != nil {
				recordCapability("GradientFillNode.SetHighlightAngle", gradientPath+".highlight_angle")
			}
			if len(layer.Shape.GradientFill.ColorStops) > 0 {
				recordCapability("GradientFillNode.SetColorStops", gradientPath+".color_stops")
				validateGradientColorStops(layer.Shape.GradientFill.ColorStops, gradientPath+".color_stops", "invalid_shape_gradient_fill_color_stops", "gradient_fill", addRefusal)
			}
			if len(layer.Shape.GradientFill.AlphaStops) > 0 {
				recordCapability("GradientFillNode.SetAlphaStops", gradientPath+".alpha_stops")
				validateGradientAlphaStops(layer.Shape.GradientFill.AlphaStops, gradientPath+".alpha_stops", "invalid_shape_gradient_fill_alpha_stops", "gradient_fill", addRefusal)
			}
		}
		if layer.Shape.GradientStroke != nil {
			gradientPath := layerPath + ".shape.gradient_stroke"
			recordCapability("VectorGroup.AddGradientStroke", gradientPath)
			if layer.Shape.GradientStroke.Type != "" {
				recordCapability("GradientStrokeNode.SetGradientType", gradientPath+".type")
				if !validGradientType(layer.Shape.GradientStroke.Type) {
					addRefusal("invalid_shape_gradient_stroke_type", gradientPath+".type", "gradient_stroke type must be linear or radial")
				}
			}
			if len(layer.Shape.GradientStroke.StartPoint) > 0 {
				recordCapability("GradientStrokeNode.SetStartPoint", gradientPath+".start_point")
				validateVec(layer.Shape.GradientStroke.StartPoint, 2, gradientPath+".start_point", addRefusal)
			}
			if len(layer.Shape.GradientStroke.EndPoint) > 0 {
				recordCapability("GradientStrokeNode.SetEndPoint", gradientPath+".end_point")
				validateVec(layer.Shape.GradientStroke.EndPoint, 2, gradientPath+".end_point", addRefusal)
			}
			if layer.Shape.GradientStroke.HighlightLength != nil {
				recordCapability("GradientStrokeNode.SetHighlightLength", gradientPath+".highlight_length")
				if *layer.Shape.GradientStroke.HighlightLength < -100 || *layer.Shape.GradientStroke.HighlightLength > 100 {
					addRefusal("invalid_shape_gradient_stroke_highlight_length", gradientPath+".highlight_length", "gradient_stroke highlight_length must be between -100 and 100")
				}
			}
			if layer.Shape.GradientStroke.HighlightAngle != nil {
				recordCapability("GradientStrokeNode.SetHighlightAngle", gradientPath+".highlight_angle")
			}
			if layer.Shape.GradientStroke.Width != nil {
				recordCapability("GradientStrokeNode.SetStrokeWidth", gradientPath+".width")
				if *layer.Shape.GradientStroke.Width < 0 {
					addRefusal("invalid_shape_gradient_stroke_width", gradientPath+".width", "gradient_stroke width must be non-negative")
				}
			}
			if layer.Shape.GradientStroke.LineCap != "" {
				recordCapability("GradientStrokeNode.SetLineCap", gradientPath+".line_cap")
				if !validStrokeLineCap(layer.Shape.GradientStroke.LineCap) {
					addRefusal("invalid_shape_gradient_stroke_line_cap", gradientPath+".line_cap", "gradient_stroke line_cap must be butt, round, or projecting")
				}
			}
			if layer.Shape.GradientStroke.LineJoin != "" {
				recordCapability("GradientStrokeNode.SetLineJoin", gradientPath+".line_join")
				if !validStrokeLineJoin(layer.Shape.GradientStroke.LineJoin) {
					addRefusal("invalid_shape_gradient_stroke_line_join", gradientPath+".line_join", "gradient_stroke line_join must be miter, round, or bevel")
				}
			}
			if layer.Shape.GradientStroke.MiterLimit != nil {
				recordCapability("GradientStrokeNode.SetMiterLimit", gradientPath+".miter_limit")
				if *layer.Shape.GradientStroke.MiterLimit < 1 {
					addRefusal("invalid_shape_gradient_stroke_miter_limit", gradientPath+".miter_limit", "gradient_stroke miter_limit must be at least 1")
				}
			}
			if len(layer.Shape.GradientStroke.ColorStops) > 0 {
				recordCapability("GradientStrokeNode.SetColorStops", gradientPath+".color_stops")
				validateGradientColorStops(layer.Shape.GradientStroke.ColorStops, gradientPath+".color_stops", "invalid_shape_gradient_stroke_color_stops", "gradient_stroke", addRefusal)
			}
			if len(layer.Shape.GradientStroke.AlphaStops) > 0 {
				recordCapability("GradientStrokeNode.SetAlphaStops", gradientPath+".alpha_stops")
				validateGradientAlphaStops(layer.Shape.GradientStroke.AlphaStops, gradientPath+".alpha_stops", "invalid_shape_gradient_stroke_alpha_stops", "gradient_stroke", addRefusal)
			}
		}
		if layer.Shape.Stroke != nil {
			strokePath := layerPath + ".shape.stroke"
			recordCapability("VectorGroup.AddStroke", strokePath)
			if len(layer.Shape.Stroke.Color) > 0 {
				recordCapability("StrokeNode.SetColor", strokePath+".color")
				validateColor(layer.Shape.Stroke.Color, strokePath+".color", "invalid_shape_stroke_color", addRefusal)
			}
			if layer.Shape.Stroke.Width != nil {
				recordCapability("StrokeNode.SetWidth", strokePath+".width")
				if *layer.Shape.Stroke.Width < 0 {
					addRefusal("invalid_shape_stroke_width", strokePath+".width", "stroke width must be non-negative")
				}
			}
			if layer.Shape.Stroke.Opacity != nil {
				recordCapability("StrokeNode.SetOpacity", strokePath+".opacity")
				if *layer.Shape.Stroke.Opacity < 0 || *layer.Shape.Stroke.Opacity > 100 {
					addRefusal("invalid_shape_stroke_opacity", strokePath+".opacity", "stroke opacity must be between 0 and 100")
				}
			}
			if layer.Shape.Stroke.LineCap != "" {
				recordCapability("StrokeNode.SetLineCap", strokePath+".line_cap")
				if !validStrokeLineCap(layer.Shape.Stroke.LineCap) {
					addRefusal("invalid_shape_stroke_line_cap", strokePath+".line_cap", "stroke line_cap must be butt, round, or projecting")
				}
			}
			if layer.Shape.Stroke.LineJoin != "" {
				recordCapability("StrokeNode.SetLineJoin", strokePath+".line_join")
				if !validStrokeLineJoin(layer.Shape.Stroke.LineJoin) {
					addRefusal("invalid_shape_stroke_line_join", strokePath+".line_join", "stroke line_join must be miter, round, or bevel")
				}
			}
			if layer.Shape.Stroke.MiterLimit != nil {
				recordCapability("StrokeNode.SetMiterLimit", strokePath+".miter_limit")
				if *layer.Shape.Stroke.MiterLimit < 1 {
					addRefusal("invalid_shape_stroke_miter_limit", strokePath+".miter_limit", "stroke miter_limit must be at least 1")
				}
			}
			if layer.Shape.Stroke.CompositeOrder != "" {
				recordCapability("StrokeNode.SetCompositeOrder", strokePath+".composite_order")
				if !validShapeCompositeOrder(layer.Shape.Stroke.CompositeOrder) {
					addRefusal("invalid_shape_stroke_composite_order", strokePath+".composite_order", "stroke composite_order must be above_previous or below_previous")
				}
			}
			if layer.Shape.Stroke.Taper != nil {
				taperPath := strokePath + ".taper"
				if layer.Shape.Stroke.Taper.StartLength != nil {
					recordCapability("StrokeTaper.SetStartLength", taperPath+".start_length")
				}
				if layer.Shape.Stroke.Taper.EndLength != nil {
					recordCapability("StrokeTaper.SetEndLength", taperPath+".end_length")
				}
				if layer.Shape.Stroke.Taper.StartWidth != nil {
					recordCapability("StrokeTaper.SetStartWidth", taperPath+".start_width")
				}
				if layer.Shape.Stroke.Taper.EndWidth != nil {
					recordCapability("StrokeTaper.SetEndWidth", taperPath+".end_width")
				}
				if layer.Shape.Stroke.Taper.StartEase != nil {
					recordCapability("StrokeTaper.SetStartEase", taperPath+".start_ease")
				}
				if layer.Shape.Stroke.Taper.EndEase != nil {
					recordCapability("StrokeTaper.SetEndEase", taperPath+".end_ease")
				}
			}
			if layer.Shape.Stroke.Wave != nil {
				wavePath := strokePath + ".wave"
				if layer.Shape.Stroke.Wave.Amount != nil {
					recordCapability("StrokeWave.SetAmount", wavePath+".amount")
				}
				if layer.Shape.Stroke.Wave.Wavelength != nil {
					recordCapability("StrokeWave.SetWavelength", wavePath+".wavelength")
				}
				if layer.Shape.Stroke.Wave.Phase != nil {
					recordCapability("StrokeWave.SetPhase", wavePath+".phase")
				}
			}
			if layer.Shape.Stroke.Dashes != nil {
				dashesPath := strokePath + ".dashes"
				if layer.Shape.Stroke.Dashes.Dash != nil {
					recordCapability("StrokeDashes.SetDash", dashesPath+".dash")
					if *layer.Shape.Stroke.Dashes.Dash < 0 {
						addRefusal("invalid_shape_stroke_dash", dashesPath+".dash", "stroke dash must be non-negative")
					}
				}
				if layer.Shape.Stroke.Dashes.Gap != nil {
					recordCapability("StrokeDashes.SetGap", dashesPath+".gap")
					if *layer.Shape.Stroke.Dashes.Gap < 0 {
						addRefusal("invalid_shape_stroke_gap", dashesPath+".gap", "stroke gap must be non-negative")
					}
				}
			}
		}
		if layer.Shape.Trim != nil {
			trimPath := layerPath + ".shape.trim"
			recordCapability("VectorGroup.AddTrim", trimPath)
			if layer.Shape.Trim.Start != nil && (*layer.Shape.Trim.Start < 0 || *layer.Shape.Trim.Start > 100) {
				addRefusal("invalid_shape_trim_start", trimPath+".start", "trim start must be between 0 and 100")
			}
			if layer.Shape.Trim.End != nil && (*layer.Shape.Trim.End < 0 || *layer.Shape.Trim.End > 100) {
				addRefusal("invalid_shape_trim_end", trimPath+".end", "trim end must be between 0 and 100")
			}
		}
		if layer.Shape.RoundCorners != nil {
			roundPath := layerPath + ".shape.round_corners"
			recordCapability("VectorGroup.AddRoundCorners", roundPath)
			if layer.Shape.RoundCorners.Radius != nil {
				recordCapability("RoundCornersNode.SetRadius", roundPath+".radius")
				if *layer.Shape.RoundCorners.Radius < 0 {
					addRefusal("invalid_shape_round_corners_radius", roundPath+".radius", "round corners radius must be non-negative")
				}
			}
		}
		if layer.Shape.OffsetPaths != nil {
			offsetPath := layerPath + ".shape.offset_paths"
			recordCapability("VectorGroup.AddOffsetPaths", offsetPath)
			if layer.Shape.OffsetPaths.Amount != nil {
				recordCapability("OffsetPathsNode.SetAmount", offsetPath+".amount")
			}
			if layer.Shape.OffsetPaths.LineJoin != "" {
				recordCapability("OffsetPathsNode.SetLineJoin", offsetPath+".line_join")
				if !validOffsetLineJoin(layer.Shape.OffsetPaths.LineJoin) {
					addRefusal("invalid_shape_offset_line_join", offsetPath+".line_join", "offset line_join must be miter, round, or bevel")
				}
			}
			if layer.Shape.OffsetPaths.MiterLimit != nil {
				recordCapability("OffsetPathsNode.SetMiterLimit", offsetPath+".miter_limit")
				if *layer.Shape.OffsetPaths.MiterLimit < 1 {
					addRefusal("invalid_shape_offset_miter_limit", offsetPath+".miter_limit", "offset miter_limit must be at least 1")
				}
			}
			if layer.Shape.OffsetPaths.Copies != nil {
				recordCapability("OffsetPathsNode.SetCopies", offsetPath+".copies")
				if *layer.Shape.OffsetPaths.Copies < 1 {
					addRefusal("invalid_shape_offset_copies", offsetPath+".copies", "offset copies must be at least 1")
				}
			}
			if layer.Shape.OffsetPaths.CopyOffset != nil {
				recordCapability("OffsetPathsNode.SetCopyOffset", offsetPath+".copy_offset")
			}
		}
		if layer.Shape.Repeater != nil {
			repeaterPath := layerPath + ".shape.repeater"
			recordCapability("VectorGroup.AddRepeater", repeaterPath)
			if layer.Shape.Repeater.Copies != nil {
				recordCapability("RepeaterNode.SetCopies", repeaterPath+".copies")
				if *layer.Shape.Repeater.Copies < 1 {
					addRefusal("invalid_shape_repeater_copies", repeaterPath+".copies", "repeater copies must be at least 1")
				}
			}
			if layer.Shape.Repeater.Offset != nil {
				recordCapability("RepeaterNode.SetOffset", repeaterPath+".offset")
			}
			if layer.Shape.Repeater.Order != "" {
				recordCapability("RepeaterNode.SetOrder", repeaterPath+".order")
				if !validRepeaterOrder(layer.Shape.Repeater.Order) {
					addRefusal("invalid_shape_repeater_order", repeaterPath+".order", "repeater order must be below or above")
				}
			}
			if len(layer.Shape.Repeater.Anchor) > 0 {
				recordCapability("RepeaterTransform.SetAnchor", repeaterPath+".anchor")
				validateVec(layer.Shape.Repeater.Anchor, 2, repeaterPath+".anchor", addRefusal)
			}
			if len(layer.Shape.Repeater.Position) > 0 {
				recordCapability("RepeaterTransform.SetPosition", repeaterPath+".position")
				validateVec(layer.Shape.Repeater.Position, 2, repeaterPath+".position", addRefusal)
			}
			if len(layer.Shape.Repeater.Scale) > 0 {
				recordCapability("RepeaterTransform.SetScale", repeaterPath+".scale")
				validateVec(layer.Shape.Repeater.Scale, 2, repeaterPath+".scale", addRefusal)
			}
			if layer.Shape.Repeater.Rotation != nil {
				recordCapability("RepeaterTransform.SetRotation", repeaterPath+".rotation")
			}
			if layer.Shape.Repeater.StartOpacity != nil {
				recordCapability("RepeaterTransform.SetStartOpacity", repeaterPath+".start_opacity")
				if *layer.Shape.Repeater.StartOpacity < 0 || *layer.Shape.Repeater.StartOpacity > 100 {
					addRefusal("invalid_shape_repeater_start_opacity", repeaterPath+".start_opacity", "repeater start_opacity must be between 0 and 100")
				}
			}
			if layer.Shape.Repeater.EndOpacity != nil {
				recordCapability("RepeaterTransform.SetEndOpacity", repeaterPath+".end_opacity")
				if *layer.Shape.Repeater.EndOpacity < 0 || *layer.Shape.Repeater.EndOpacity > 100 {
					addRefusal("invalid_shape_repeater_end_opacity", repeaterPath+".end_opacity", "repeater end_opacity must be between 0 and 100")
				}
			}
		}
		if layer.Shape.MergePaths != nil {
			mergePath := layerPath + ".shape.merge_paths"
			recordCapability("VectorGroup.AddMergePaths", mergePath)
			if layer.Shape.MergePaths.Type != "" {
				recordCapability("MergePathsNode.SetType", mergePath+".type")
				if !validMergePathsType(layer.Shape.MergePaths.Type) {
					addRefusal("invalid_shape_merge_paths_type", mergePath+".type", "merge_paths type must be merge, add, subtract, intersect, or exclude")
				}
			}
		}
		if layer.Shape.ZigZag != nil {
			zigZagPath := layerPath + ".shape.zigzag"
			recordCapability("VectorGroup.AddZigZag", zigZagPath)
			if layer.Shape.ZigZag.Size != nil {
				recordCapability("ZigZagNode.SetSize", zigZagPath+".size")
				if *layer.Shape.ZigZag.Size < 0 {
					addRefusal("invalid_shape_zigzag_size", zigZagPath+".size", "zigzag size must be non-negative")
				}
			}
			if layer.Shape.ZigZag.Detail != nil {
				recordCapability("ZigZagNode.SetDetail", zigZagPath+".detail")
				if *layer.Shape.ZigZag.Detail < 0 {
					addRefusal("invalid_shape_zigzag_detail", zigZagPath+".detail", "zigzag detail must be non-negative")
				}
			}
			if layer.Shape.ZigZag.Points != "" {
				recordCapability("ZigZagNode.SetPoints", zigZagPath+".points")
				if !validZigZagPoints(layer.Shape.ZigZag.Points) {
					addRefusal("invalid_shape_zigzag_points", zigZagPath+".points", "zigzag points must be corner or smooth")
				}
			}
		}
		if layer.Shape.PuckerBloat != nil {
			puckerBloatPath := layerPath + ".shape.pucker_bloat"
			recordCapability("VectorGroup.AddPuckerBloat", puckerBloatPath)
			if layer.Shape.PuckerBloat.Amount != nil {
				recordCapability("PuckerBloatNode.SetAmount", puckerBloatPath+".amount")
			}
		}
		if layer.Shape.Twist != nil {
			twistPath := layerPath + ".shape.twist"
			recordCapability("VectorGroup.AddTwist", twistPath)
			if layer.Shape.Twist.Angle != nil {
				recordCapability("TwistNode.SetAngle", twistPath+".angle")
			}
			if len(layer.Shape.Twist.Center) > 0 {
				recordCapability("TwistNode.SetCenter", twistPath+".center")
				validateVec(layer.Shape.Twist.Center, 2, twistPath+".center", addRefusal)
			}
		}
		if layer.Shape.WigglePaths != nil {
			wigglePath := layerPath + ".shape.wiggle_paths"
			recordCapability("VectorGroup.AddWigglePaths", wigglePath)
			if layer.Shape.WigglePaths.Size != nil {
				recordCapability("WigglePathsNode.SetSize", wigglePath+".size")
			}
			if layer.Shape.WigglePaths.Detail != nil {
				recordCapability("WigglePathsNode.SetDetail", wigglePath+".detail")
			}
			if layer.Shape.WigglePaths.WigglesPerSecond != nil {
				recordCapability("WigglePathsNode.SetWigglesPerSecond", wigglePath+".wiggles_per_second")
			}
			if layer.Shape.WigglePaths.RandomSeed != nil {
				recordCapability("WigglePathsNode.SetRandomSeed", wigglePath+".random_seed")
			}
			if layer.Shape.WigglePaths.Points != "" {
				recordCapability("WigglePathsNode.SetPoints", wigglePath+".points")
				if !validRoughenPoints(layer.Shape.WigglePaths.Points) {
					addRefusal("invalid_shape_wiggle_paths_points", wigglePath+".points", "wiggle_paths points must be corner or smooth")
				}
			}
			if layer.Shape.WigglePaths.Correlation != nil {
				recordCapability("WigglePathsNode.SetCorrelation", wigglePath+".correlation")
				if *layer.Shape.WigglePaths.Correlation < 0 || *layer.Shape.WigglePaths.Correlation > 100 {
					addRefusal("invalid_shape_wiggle_paths_correlation", wigglePath+".correlation", "wiggle_paths correlation must be between 0 and 100")
				}
			}
			if layer.Shape.WigglePaths.TemporalPhase != nil {
				recordCapability("WigglePathsNode.SetTemporalPhase", wigglePath+".temporal_phase")
			}
			if layer.Shape.WigglePaths.SpatialPhase != nil {
				recordCapability("WigglePathsNode.SetSpatialPhase", wigglePath+".spatial_phase")
			}
		}
		if layer.Shape.WiggleTransform != nil {
			wigglePath := layerPath + ".shape.wiggle_transform"
			recordCapability("VectorGroup.AddWiggleTransform", wigglePath)
			if len(layer.Shape.WiggleTransform.Anchor) > 0 {
				recordCapability("WigglerTransform.SetAnchor", wigglePath+".anchor")
				validateVec(layer.Shape.WiggleTransform.Anchor, 2, wigglePath+".anchor", addRefusal)
			}
			if len(layer.Shape.WiggleTransform.Position) > 0 {
				recordCapability("WigglerTransform.SetPosition", wigglePath+".position")
				validateVec(layer.Shape.WiggleTransform.Position, 2, wigglePath+".position", addRefusal)
			}
			if len(layer.Shape.WiggleTransform.Scale) > 0 {
				recordCapability("WigglerTransform.SetScale", wigglePath+".scale")
				validateVec(layer.Shape.WiggleTransform.Scale, 2, wigglePath+".scale", addRefusal)
			}
			if layer.Shape.WiggleTransform.Rotation != nil {
				recordCapability("WigglerTransform.SetRotation", wigglePath+".rotation")
			}
			if layer.Shape.WiggleTransform.WigglesPerSecond != nil {
				recordCapability("WiggleTransformNode.SetWigglesPerSecond", wigglePath+".wiggles_per_second")
			}
			if layer.Shape.WiggleTransform.RandomSeed != nil {
				recordCapability("WiggleTransformNode.SetRandomSeed", wigglePath+".random_seed")
			}
			if layer.Shape.WiggleTransform.Correlation != nil {
				recordCapability("WiggleTransformNode.SetCorrelation", wigglePath+".correlation")
				if *layer.Shape.WiggleTransform.Correlation < 0 || *layer.Shape.WiggleTransform.Correlation > 100 {
					addRefusal("invalid_shape_wiggle_transform_correlation", wigglePath+".correlation", "wiggle_transform correlation must be between 0 and 100")
				}
			}
			if layer.Shape.WiggleTransform.TemporalPhase != nil {
				recordCapability("WiggleTransformNode.SetTemporalPhase", wigglePath+".temporal_phase")
			}
			if layer.Shape.WiggleTransform.SpatialPhase != nil {
				recordCapability("WiggleTransformNode.SetSpatialPhase", wigglePath+".spatial_phase")
			}
		}
	}
	for i, mask := range layer.Masks {
		maskPath := fmt.Sprintf("%s.masks[%d]", layerPath, i)
		recordCapability("AddMask", maskPath)
		if layer.Type == "camera" || layer.Type == "light" {
			addRefusal("mask_on_unsupported_layer_type", maskPath, "masks are not supported on camera or light layers")
		}
		if mask.Mode != "" {
			recordCapability("Mask.SetMode", maskPath+".mode")
			if !validMaskMode(mask.Mode) {
				addRefusal("invalid_mask_mode", maskPath+".mode", "mask mode must be add, subtract, intersect, lighten, darken, difference, or none")
			}
		}
		if mask.Inverted != nil {
			recordCapability("Mask.SetInverted", maskPath+".inverted")
		}
		if mask.Locked != nil {
			recordCapability("Mask.SetLocked", maskPath+".locked")
		}
		if len(mask.Color) > 0 {
			recordCapability("Mask.SetColor", maskPath+".color")
			validateRGBColor(mask.Color, maskPath+".color", "invalid_mask_color", addRefusal)
		}
		if mask.MotionBlur != "" {
			recordCapability("Mask.SetMaskMotionBlur", maskPath+".motion_blur")
			if !validMaskMotionBlur(mask.MotionBlur) {
				addRefusal("invalid_mask_motion_blur", maskPath+".motion_blur", "mask motion_blur must be same_as_layer, on, or off")
			}
		}
		if mask.FeatherFalloff != "" {
			recordCapability("Mask.SetFeatherFalloff", maskPath+".feather_falloff")
			if !validMaskFeatherFalloff(mask.FeatherFalloff) {
				addRefusal("invalid_mask_feather_falloff", maskPath+".feather_falloff", "mask feather_falloff must be smooth or linear")
			}
		}
		if mask.Opacity != nil {
			recordCapability("Mask.SetOpacity", maskPath+".opacity")
			if *mask.Opacity < 0 || *mask.Opacity > 1 {
				addRefusal("invalid_mask_opacity", maskPath+".opacity", "mask opacity must be between 0 and 1")
			}
		}
		if len(mask.Feather) > 0 {
			recordCapability("Mask.SetFeather", maskPath+".feather")
			validateMaskFeather(mask.Feather, maskPath+".feather", "invalid_mask_feather", addRefusal)
		}
		if mask.Expansion != nil {
			recordCapability("Mask.SetExpansion", maskPath+".expansion")
		}
		if len(mask.Vertices) < 3 {
			addRefusal("invalid_mask_vertices", maskPath+".vertices", "mask vertices must include at least 3 points")
		}
		for vi, vertex := range mask.Vertices {
			validateVec(vertex, 2, fmt.Sprintf("%s.vertices[%d]", maskPath, vi), addRefusal)
		}
	}
	if usesTransform(layer.Transform) {
		recordCapability("SetLayerTransform", layerPath+".transform")
	}
	validateTransformExpressions(layer.Transform.Expressions, layerPath+".transform.expressions", recordCapability, addRefusal)
	validateVec(layer.Transform.Position, 2, layerPath+".transform.position", addRefusal)
	validateVec(layer.Transform.Scale, 2, layerPath+".transform.scale", addRefusal)
	validateVec(layer.Transform.AnchorPoint, 2, layerPath+".transform.anchor_point", addRefusal)
	for i, kf := range layer.Transform.PositionKeyframes {
		kfPath := fmt.Sprintf("%s.transform.position_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.PositionKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateVec(kf.Value, 2, kfPath+".value", addRefusal)
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.AnchorPointKeyframes {
		kfPath := fmt.Sprintf("%s.transform.anchor_point_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.AnchorPointKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateVec(kf.Value, 2, kfPath+".value", addRefusal)
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.ScaleKeyframes {
		kfPath := fmt.Sprintf("%s.transform.scale_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.ScaleKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateVec(kf.Value, 2, kfPath+".value", addRefusal)
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.RotationKeyframes {
		kfPath := fmt.Sprintf("%s.transform.rotation_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.RotationKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.OpacityKeyframes {
		kfPath := fmt.Sprintf("%s.transform.opacity_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.OpacityKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		if kf.Value < 0 || kf.Value > 100 {
			addRefusal("invalid_opacity_keyframe_value", kfPath+".value", "opacity keyframe value must be between 0 and 100")
		}
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, effect := range layer.Effects {
		effectPath := fmt.Sprintf("%s.effects[%d]", layerPath, i)
		lookup := recordCapability("AddEffect", effectPath)
		if effect.MatchName == "" {
			addRefusal("missing_effect_match_name", effectPath+".match_name", "effect match_name is required")
			continue
		}
		if lookup.Status == CapabilityUnsupported {
			addRefusal("unsupported_effect_api", effectPath, "AddEffect is not supported by the loaded capability index")
			continue
		}
		if !supportedEffect(effect.MatchName) {
			addRefusal("unsupported_effect", effectPath+".match_name", fmt.Sprintf("effect %q is not in SupportedEffects", effect.MatchName))
			continue
		}
		for pi, param := range effect.Params {
			paramPath := fmt.Sprintf("%s.params[%d]", effectPath, pi)
			recordCapability("SetEffectParam", paramPath)
			if param.MatchName == "" {
				addRefusal("missing_effect_param_match_name", paramPath+".match_name", "effect param match_name is required")
			}
			if !validEffectParamValue(param.Value) {
				addRefusal("unsupported_effect_param_value", paramPath+".value", "effect param value must be a number, boolean, or numeric array")
			}
			if param.Expression != nil {
				recordCapability("Property.SetExpression", paramPath+".expression.source")
				if param.Expression.Source == "" {
					addRefusal("missing_effect_param_expression_source", paramPath+".expression.source", "effect param expression source is required")
				}
				if param.Expression.Enabled != nil {
					recordCapability("Property.SetExpressionEnabled", paramPath+".expression.enabled")
				}
			}
		}
	}
}

func validateTextStyle(style TextStyleSpec, stylePath string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	if style.RunIndex < 0 {
		addRefusal("invalid_text_style_run_index", stylePath+".run_index", "run_index must be non-negative")
	}
	if style.ParagraphIndex < 0 {
		addRefusal("invalid_text_style_paragraph_index", stylePath+".paragraph_index", "paragraph_index must be non-negative")
	}
	if style.FontSize != nil {
		recordCapability("Layer.SetRunFontSize", stylePath+".font_size")
		if *style.FontSize <= 0 {
			addRefusal("invalid_text_font_size", stylePath+".font_size", "font_size must be positive")
		}
	}
	if len(style.FillColor) > 0 {
		recordCapability("Layer.SetRunFillColor", stylePath+".fill_color")
		validateColor(style.FillColor, stylePath+".fill_color", "invalid_text_fill_color", addRefusal)
	}
	if style.Tracking != nil {
		recordCapability("Layer.SetRunTracking", stylePath+".tracking")
	}
	if style.FauxBold != nil {
		recordCapability("Layer.SetRunFauxBold", stylePath+".faux_bold")
	}
	if style.FauxItalic != nil {
		recordCapability("Layer.SetRunFauxItalic", stylePath+".faux_italic")
	}
	if style.ApplyStroke != nil {
		recordCapability("Layer.SetRunApplyStroke", stylePath+".apply_stroke")
	}
	if len(style.StrokeColor) > 0 {
		recordCapability("Layer.SetRunStrokeColor", stylePath+".stroke_color")
		validateColor(style.StrokeColor, stylePath+".stroke_color", "invalid_text_stroke_color", addRefusal)
	}
	if style.StrokeWidth != nil {
		recordCapability("Layer.SetRunStrokeWidth", stylePath+".stroke_width")
		if *style.StrokeWidth < 0 {
			addRefusal("invalid_text_stroke_width", stylePath+".stroke_width", "stroke_width must be non-negative")
		}
	}
	if style.Justification != "" {
		recordCapability("Layer.SetParagraphJustification", stylePath+".justification")
		if !validTextJustification(style.Justification) {
			addRefusal("invalid_text_justification", stylePath+".justification", "justification must be left, right, or center")
		}
	}
}

func validTextJustification(value string) bool {
	switch value {
	case "left", "right", "center":
		return true
	default:
		return false
	}
}

func validOffsetLineJoin(value string) bool {
	switch value {
	case "miter", "round", "bevel":
		return true
	default:
		return false
	}
}

func validStrokeLineCap(value string) bool {
	switch value {
	case "butt", "round", "projecting":
		return true
	default:
		return false
	}
}

func validStrokeLineJoin(value string) bool {
	switch value {
	case "miter", "round", "bevel":
		return true
	default:
		return false
	}
}

func validShapeCompositeOrder(value string) bool {
	switch value {
	case "above_previous", "below_previous":
		return true
	default:
		return false
	}
}

func validShapeBlendMode(value float64) bool {
	return value >= 1 && isWholeNumber(value)
}

func validFillRule(value string) bool {
	switch value {
	case "nonzero_winding", "even_odd":
		return true
	default:
		return false
	}
}

func validGradientType(value string) bool {
	switch value {
	case "linear", "radial":
		return true
	default:
		return false
	}
}

func validRepeaterOrder(value string) bool {
	switch value {
	case "below", "above":
		return true
	default:
		return false
	}
}

func validMergePathsType(value string) bool {
	switch value {
	case "merge", "add", "subtract", "intersect", "exclude":
		return true
	default:
		return false
	}
}

func validZigZagPoints(value string) bool {
	switch value {
	case "corner", "smooth":
		return true
	default:
		return false
	}
}

func validRoughenPoints(value string) bool {
	switch value {
	case "corner", "smooth":
		return true
	default:
		return false
	}
}

func usesTransform(t Transform) bool {
	return len(t.Position) > 0 ||
		len(t.Scale) > 0 ||
		len(t.AnchorPoint) > 0 ||
		t.Rotation != nil ||
		t.Opacity != nil ||
		len(t.PositionKeyframes) > 0 ||
		len(t.AnchorPointKeyframes) > 0 ||
		len(t.ScaleKeyframes) > 0 ||
		len(t.RotationKeyframes) > 0 ||
		len(t.OpacityKeyframes) > 0
}

func validateVec(values []float64, want int, path string, addRefusal func(string, string, string)) {
	if len(values) == 0 {
		return
	}
	if len(values) != want {
		addRefusal("invalid_vector_size", path, fmt.Sprintf("expected %d values", want))
	}
}

func validateKeyframeEase(ease *TemporalEase, path string, addRefusal func(string, string, string)) {
	if ease == nil {
		return
	}
	if ease.Influence <= 0 || ease.Influence > 1 {
		addRefusal("invalid_keyframe_ease_influence", path+".influence", "keyframe ease influence must be greater than 0 and at most 1")
	}
}

func validMaskMode(value string) bool {
	switch value {
	case "none", "add", "subtract", "intersect", "lighten", "darken", "difference":
		return true
	default:
		return false
	}
}

func validMaskMotionBlur(value string) bool {
	switch value {
	case "same_as_layer", "on", "off":
		return true
	default:
		return false
	}
}

func validMaskFeatherFalloff(value string) bool {
	switch value {
	case "smooth", "linear":
		return true
	default:
		return false
	}
}

func validateMaskFeather(values []float64, path, code string, addRefusal func(string, string, string)) {
	if len(values) == 0 {
		return
	}
	if len(values) != 2 {
		addRefusal(code, path, "mask feather must have 2 values")
		return
	}
	for i, value := range values {
		if value < 0 {
			addRefusal(code, fmt.Sprintf("%s[%d]", path, i), "mask feather values must be non-negative")
		}
	}
}

func validateTransformExpressions(expressions TransformExpressions, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	validateTransformExpression(expressions.Position, path+".position", recordCapability, addRefusal)
	validateTransformExpression(expressions.AnchorPoint, path+".anchor_point", recordCapability, addRefusal)
	validateTransformExpression(expressions.Scale, path+".scale", recordCapability, addRefusal)
	validateTransformExpression(expressions.Rotation, path+".rotation", recordCapability, addRefusal)
	validateTransformExpression(expressions.Opacity, path+".opacity", recordCapability, addRefusal)
}

func validateTransformExpression(expression *ExpressionSpec, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	if expression == nil {
		return
	}
	recordCapability("Property.SetExpression", path+".source")
	if expression.Source == "" {
		addRefusal("missing_transform_expression_source", path+".source", "transform expression source is required")
	}
	if expression.Enabled != nil {
		recordCapability("Property.SetExpressionEnabled", path+".enabled")
	}
}

func validateColor(values []float64, path, code string, addRefusal func(string, string, string)) {
	if len(values) != 3 && len(values) != 4 {
		addRefusal(code, path, "color must have 3 or 4 channels")
		return
	}
	for i, value := range values {
		if value < 0 || value > 255 {
			addRefusal(code, fmt.Sprintf("%s[%d]", path, i), "color channels must be between 0 and 255")
		}
	}
}

func validateGradientColorStops(stops []GradientColorStopSpec, path, code, label string, addRefusal func(string, string, string)) {
	if len(stops) < 2 {
		addRefusal(code, path, label+" color_stops must include at least 2 stops")
	}
	for i, stop := range stops {
		stopPath := fmt.Sprintf("%s[%d]", path, i)
		if stop.Offset < 0 || stop.Offset > 1 {
			addRefusal(code, stopPath+".offset", label+" stop offset must be between 0 and 1")
		}
		if stop.Midpoint != nil && (*stop.Midpoint < 0 || *stop.Midpoint > 1) {
			addRefusal(code, stopPath+".midpoint", label+" stop midpoint must be between 0 and 1")
		}
		validateColor(stop.Color, stopPath+".color", code, addRefusal)
	}
}

func validateGradientAlphaStops(stops []GradientAlphaStopSpec, path, code, label string, addRefusal func(string, string, string)) {
	if len(stops) < 2 {
		addRefusal(code, path, label+" alpha_stops must include at least 2 stops")
	}
	for i, stop := range stops {
		stopPath := fmt.Sprintf("%s[%d]", path, i)
		if stop.Offset < 0 || stop.Offset > 1 {
			addRefusal(code, stopPath+".offset", label+" alpha stop offset must be between 0 and 1")
		}
		if stop.Midpoint != nil && (*stop.Midpoint < 0 || *stop.Midpoint > 1) {
			addRefusal(code, stopPath+".midpoint", label+" alpha stop midpoint must be between 0 and 1")
		}
		if stop.Alpha < 0 || stop.Alpha > 1 {
			addRefusal(code, stopPath+".alpha", label+" alpha stop alpha must be between 0 and 1")
		}
	}
}

func validateRGBColor(values []float64, path, code string, addRefusal func(string, string, string)) {
	if len(values) != 3 {
		addRefusal(code, path, "color must have 3 channels")
		return
	}
	for i, value := range values {
		if value < 0 || value > 255 {
			addRefusal(code, fmt.Sprintf("%s[%d]", path, i), "color channels must be between 0 and 255")
		}
	}
}

func isWholeNumber(value float64) bool {
	return value == float64(int64(value))
}

func validEffectParamValue(value any) bool {
	switch v := value.(type) {
	case float64, int, bool:
		return true
	case []float64:
		return len(v) > 0
	case []any:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			if _, ok := item.(float64); !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func supportedEffect(matchName string) bool {
	for _, effect := range aep.SupportedEffects() {
		if effect == matchName {
			return true
		}
	}
	return false
}
