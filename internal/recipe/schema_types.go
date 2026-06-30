package recipe

const SchemaVersion = 1

type Recipe struct {
	SchemaVersion   int             `json:"schema_version"`
	Project         ProjectSpec     `json:"project"`
	Comps           []CompSpec      `json:"comps"`
	ExpectedProfile ExpectedProfile `json:"expected_profile,omitempty"`
}

type ProjectSpec struct {
	Name          string `json:"name,omitempty"`
	TargetVersion string `json:"target_version,omitempty"`
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
	Type                  string             `json:"type"`
	Name                  string             `json:"name"`
	Label                 *float64           `json:"label,omitempty"`
	Comment               string             `json:"comment,omitempty"`
	Visible               *bool              `json:"visible,omitempty"`
	Solo                  *bool              `json:"solo,omitempty"`
	Locked                *bool              `json:"locked,omitempty"`
	MotionBlur            *bool              `json:"motion_blur,omitempty"`
	Shy                   *bool              `json:"shy,omitempty"`
	EffectsEnabled        *bool              `json:"effects_enabled,omitempty"`
	AudioEnabled          *bool              `json:"audio_enabled,omitempty"`
	FrameBlendEnabled     *bool              `json:"frame_blend_enabled,omitempty"`
	MarkersLocked         *bool              `json:"markers_locked,omitempty"`
	CollapseTransform     *bool              `json:"collapse_transform,omitempty"`
	Is3D                  *bool              `json:"is_3d,omitempty"`
	IsAdjust              *bool              `json:"is_adjust,omitempty"`
	IsNull                *bool              `json:"is_null,omitempty"`
	IsGuide               *bool              `json:"is_guide,omitempty"`
	SamplingBicubic       *bool              `json:"sampling_bicubic,omitempty"`
	FrameBlendPixelMotion *bool              `json:"frame_blend_pixel_motion,omitempty"`
	PreserveTransparency  *bool              `json:"preserve_transparency,omitempty"`
	Quality               string             `json:"quality,omitempty"`
	BlendingMode          string             `json:"blending_mode,omitempty"`
	TrackMatte            string             `json:"track_matte,omitempty"`
	Matte                 string             `json:"matte,omitempty"`
	AutoOrient            string             `json:"auto_orient,omitempty"`
	StartTime             *float64           `json:"start_time,omitempty"`
	InPoint               *float64           `json:"in_point,omitempty"`
	OutPoint              *float64           `json:"out_point,omitempty"`
	Parent                string             `json:"parent,omitempty"`
	Text                  string             `json:"text,omitempty"`
	TextStyle             *TextStyleSpec     `json:"text_style,omitempty"`
	TextAnimators         []TextAnimatorSpec `json:"text_animators,omitempty"`
	Camera                *CameraSpec        `json:"camera,omitempty"`
	Light                 *LightSpec         `json:"light,omitempty"`
	Shape                 *ShapeSpec         `json:"shape,omitempty"`
	Masks                 []MaskSpec         `json:"masks,omitempty"`
	Transform             Transform          `json:"transform,omitempty"`
	Effects               []Effect           `json:"effects,omitempty"`
}

type MaskSpec struct {
	Name           string                 `json:"name,omitempty"`
	Mode           string                 `json:"mode,omitempty"`
	Inverted       *bool                  `json:"inverted,omitempty"`
	Locked         *bool                  `json:"locked,omitempty"`
	Color          []float64              `json:"color,omitempty"`
	MotionBlur     string                 `json:"motion_blur,omitempty"`
	FeatherFalloff string                 `json:"feather_falloff,omitempty"`
	Opacity        *float64               `json:"opacity,omitempty"`
	Feather        []float64              `json:"feather,omitempty"`
	Expansion      *float64               `json:"expansion,omitempty"`
	Closed         *bool                  `json:"closed,omitempty"`
	Vertices       [][]float64            `json:"vertices"`
	PathKeyframes  []MaskPathKeyframeSpec `json:"path_keyframes,omitempty"`
}

type MaskPathKeyframeSpec struct {
	Time     float64     `json:"time"`
	Vertices [][]float64 `json:"vertices"`
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

type TextAnimatorSpec struct {
	Property             string           `json:"property"`
	Value                any              `json:"value,omitempty"`
	RangeStart           *float64         `json:"range_start,omitempty"`
	RangeEnd             *float64         `json:"range_end,omitempty"`
	RangeOffset          *float64         `json:"range_offset,omitempty"`
	RangeOffsetKeyframes []ScalarKeyframe `json:"range_offset_keyframes,omitempty"`
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
	Name                     string                      `json:"name,omitempty"`
	Width                    *float64                    `json:"width,omitempty"`
	Height                   *float64                    `json:"height,omitempty"`
	FrameRate                *float64                    `json:"frame_rate,omitempty"`
	Duration                 *float64                    `json:"duration,omitempty"`
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
	LayerName         string `json:"layer_name"`
	MatchName         string `json:"match_name"`
	Value             any    `json:"value,omitempty"`
	Expression        string `json:"expression,omitempty"`
	ExpressionEnabled *bool  `json:"expression_enabled,omitempty"`
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
	LayerName      string                     `json:"layer_name"`
	Name           string                     `json:"name,omitempty"`
	Mode           string                     `json:"mode,omitempty"`
	Inverted       *bool                      `json:"inverted,omitempty"`
	Locked         *bool                      `json:"locked,omitempty"`
	Color          []float64                  `json:"color,omitempty"`
	MotionBlur     string                     `json:"motion_blur,omitempty"`
	FeatherFalloff string                     `json:"feather_falloff,omitempty"`
	Opacity        *float64                   `json:"opacity,omitempty"`
	Feather        []float64                  `json:"feather,omitempty"`
	Expansion      *float64                   `json:"expansion,omitempty"`
	Closed         *bool                      `json:"closed,omitempty"`
	VertexCount    *int                       `json:"vertex_count,omitempty"`
	PathKeyframes  []ExpectedMaskPathKeyframe `json:"path_keyframes,omitempty"`
}

type ExpectedMaskPathKeyframe struct {
	Time        float64 `json:"time"`
	VertexCount *int    `json:"vertex_count,omitempty"`
}

type ExpectedEffect struct {
	LayerName string                `json:"layer_name"`
	MatchName string                `json:"match_name"`
	Params    []ExpectedEffectParam `json:"params,omitempty"`
}

type ExpectedEffectParam struct {
	MatchName         string `json:"match_name"`
	Value             any    `json:"value,omitempty"`
	Expression        string `json:"expression,omitempty"`
	ExpressionEnabled *bool  `json:"expression_enabled,omitempty"`
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
