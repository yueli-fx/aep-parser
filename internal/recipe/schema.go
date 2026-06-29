package recipe

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
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
	AutoOrient            string         `json:"auto_orient,omitempty"`
	StartTime             *float64       `json:"start_time,omitempty"`
	InPoint               *float64       `json:"in_point,omitempty"`
	OutPoint              *float64       `json:"out_point,omitempty"`
	Parent                string         `json:"parent,omitempty"`
	Text                  string         `json:"text,omitempty"`
	TextStyle             *TextStyleSpec `json:"text_style,omitempty"`
	Shape                 *ShapeSpec     `json:"shape,omitempty"`
	Transform             Transform      `json:"transform,omitempty"`
	Effects               []Effect       `json:"effects,omitempty"`
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
	Kind            string               `json:"kind"`
	Size            []float64            `json:"size,omitempty"`
	Position        []float64            `json:"position,omitempty"`
	Roundness       *float64             `json:"roundness,omitempty"`
	Points          *float64             `json:"points,omitempty"`
	Rotation        *float64             `json:"rotation,omitempty"`
	InnerRadius     *float64             `json:"inner_radius,omitempty"`
	OuterRadius     *float64             `json:"outer_radius,omitempty"`
	InnerRoundness  *float64             `json:"inner_roundness,omitempty"`
	OuterRoundness  *float64             `json:"outer_roundness,omitempty"`
	FillColor       []float64            `json:"fill_color,omitempty"`
	FillOpacity     *float64             `json:"fill_opacity,omitempty"`
	Stroke          *StrokeSpec          `json:"stroke,omitempty"`
	Trim            *TrimSpec            `json:"trim,omitempty"`
	RoundCorners    *RoundCornersSpec    `json:"round_corners,omitempty"`
	OffsetPaths     *OffsetPathsSpec     `json:"offset_paths,omitempty"`
	Repeater        *RepeaterSpec        `json:"repeater,omitempty"`
	MergePaths      *MergePathsSpec      `json:"merge_paths,omitempty"`
	ZigZag          *ZigZagSpec          `json:"zigzag,omitempty"`
	PuckerBloat     *PuckerBloatSpec     `json:"pucker_bloat,omitempty"`
	Twist           *TwistSpec           `json:"twist,omitempty"`
	WigglePaths     *WigglePathsSpec     `json:"wiggle_paths,omitempty"`
	WiggleTransform *WiggleTransformSpec `json:"wiggle_transform,omitempty"`
}

type StrokeSpec struct {
	Color      []float64         `json:"color,omitempty"`
	Width      *float64          `json:"width,omitempty"`
	Opacity    *float64          `json:"opacity,omitempty"`
	LineCap    string            `json:"line_cap,omitempty"`
	LineJoin   string            `json:"line_join,omitempty"`
	MiterLimit *float64          `json:"miter_limit,omitempty"`
	Taper      *StrokeTaperSpec  `json:"taper,omitempty"`
	Wave       *StrokeWaveSpec   `json:"wave,omitempty"`
	Dashes     *StrokeDashesSpec `json:"dashes,omitempty"`
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
	MatchName string `json:"match_name"`
	Value     any    `json:"value,omitempty"`
}

type Transform struct {
	Position             []float64        `json:"position,omitempty"`
	Scale                []float64        `json:"scale,omitempty"`
	AnchorPoint          []float64        `json:"anchor_point,omitempty"`
	Rotation             *float64         `json:"rotation,omitempty"`
	Opacity              *float64         `json:"opacity,omitempty"`
	PositionKeyframes    []VectorKeyframe `json:"position_keyframes,omitempty"`
	AnchorPointKeyframes []VectorKeyframe `json:"anchor_point_keyframes,omitempty"`
	ScaleKeyframes       []VectorKeyframe `json:"scale_keyframes,omitempty"`
	RotationKeyframes    []ScalarKeyframe `json:"rotation_keyframes,omitempty"`
	OpacityKeyframes     []ScalarKeyframe `json:"opacity_keyframes,omitempty"`
}

type VectorKeyframe struct {
	Time  float64   `json:"time"`
	Value []float64 `json:"value"`
}

type ScalarKeyframe struct {
	Time  float64 `json:"time"`
	Value float64 `json:"value"`
}

type ExpectedProfile struct {
	CompCount       *int                        `json:"comp_count,omitempty"`
	LayerCount      *int                        `json:"layer_count,omitempty"`
	TextLayerCount  *int                        `json:"text_layer_count,omitempty"`
	ShapeLayerCount *int                        `json:"shape_layer_count,omitempty"`
	Renderer        string                      `json:"renderer,omitempty"`
	MotionBlur      *ExpectedMotionBlurSpec     `json:"motion_blur,omitempty"`
	WorkArea        *ExpectedWorkAreaSpec       `json:"work_area,omitempty"`
	Effects         []ExpectedEffect            `json:"effects,omitempty"`
	Properties      []ExpectedProperty          `json:"properties,omitempty"`
	TextStyles      []ExpectedTextStyle         `json:"text_styles,omitempty"`
	Keyframes       []ExpectedKeyframedProperty `json:"keyframes,omitempty"`
}

type ExpectedWorkAreaSpec struct {
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
}

type ExpectedMotionBlurSpec struct {
	ShutterAngle        *float64 `json:"shutter_angle,omitempty"`
	ShutterPhase        *float64 `json:"shutter_phase,omitempty"`
	AdaptiveSampleLimit *float64 `json:"adaptive_sample_limit,omitempty"`
	SamplesPerFrame     *float64 `json:"samples_per_frame,omitempty"`
}

type ExpectedProperty struct {
	LayerName string `json:"layer_name"`
	MatchName string `json:"match_name"`
	Value     any    `json:"value,omitempty"`
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

type ExpectedEffect struct {
	LayerName string                `json:"layer_name"`
	MatchName string                `json:"match_name"`
	Params    []ExpectedEffectParam `json:"params,omitempty"`
}

type ExpectedEffectParam struct {
	MatchName string `json:"match_name"`
	Value     any    `json:"value,omitempty"`
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
			if !validEffectParamValue(param.Value) {
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
		if !validEffectParamValue(prop.Value) {
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
	if layer.AutoOrient != "" {
		recordCapability("Layer.SetAutoOrient", layerPath+".auto_orient")
		if _, err := layerAutoOrient(layer.AutoOrient); err != nil {
			addRefusal("invalid_layer_auto_orient", layerPath+".auto_orient", "auto_orient must be none, along_path, camera_or_point_of_interest, or characters_toward_camera")
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
	if usesTransform(layer.Transform) {
		recordCapability("SetLayerTransform", layerPath+".transform")
	}
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
	}
	for i, kf := range layer.Transform.RotationKeyframes {
		kfPath := fmt.Sprintf("%s.transform.rotation_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.RotationKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
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
