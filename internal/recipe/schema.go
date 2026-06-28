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
	Name            string    `json:"name"`
	Width           int       `json:"width"`
	Height          int       `json:"height"`
	FrameRate       float64   `json:"frame_rate"`
	Duration        float64   `json:"duration"`
	BackgroundColor []float64 `json:"background_color,omitempty"`
	Layers          []Layer   `json:"layers,omitempty"`
}

type Layer struct {
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Text      string     `json:"text,omitempty"`
	Shape     *ShapeSpec `json:"shape,omitempty"`
	Transform Transform  `json:"transform,omitempty"`
	Effects   []Effect   `json:"effects,omitempty"`
}

type ShapeSpec struct {
	Kind      string      `json:"kind"`
	Size      []float64   `json:"size,omitempty"`
	FillColor []float64   `json:"fill_color,omitempty"`
	Stroke    *StrokeSpec `json:"stroke,omitempty"`
}

type StrokeSpec struct {
	Color   []float64 `json:"color,omitempty"`
	Width   *float64  `json:"width,omitempty"`
	Opacity *float64  `json:"opacity,omitempty"`
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
	Position          []float64        `json:"position,omitempty"`
	Scale             []float64        `json:"scale,omitempty"`
	AnchorPoint       []float64        `json:"anchor_point,omitempty"`
	Rotation          *float64         `json:"rotation,omitempty"`
	Opacity           *float64         `json:"opacity,omitempty"`
	PositionKeyframes []VectorKeyframe `json:"position_keyframes,omitempty"`
}

type VectorKeyframe struct {
	Time  float64   `json:"time"`
	Value []float64 `json:"value"`
}

type ExpectedProfile struct {
	CompCount       *int               `json:"comp_count,omitempty"`
	LayerCount      *int               `json:"layer_count,omitempty"`
	TextLayerCount  *int               `json:"text_layer_count,omitempty"`
	ShapeLayerCount *int               `json:"shape_layer_count,omitempty"`
	Effects         []ExpectedEffect   `json:"effects,omitempty"`
	Properties      []ExpectedProperty `json:"properties,omitempty"`
}

type ExpectedProperty struct {
	LayerName string `json:"layer_name"`
	MatchName string `json:"match_name"`
	Value     any    `json:"value,omitempty"`
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
		for li, layer := range comp.Layers {
			layerPath := fmt.Sprintf("%s.layers[%d]", compPath, li)
			validateLayer(layer, layerPath, comp.Duration, recordCapability, addRefusal)
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
	default:
		addRefusal("unsupported_layer_type", layerPath+".type", fmt.Sprintf("unsupported layer type %q", layer.Type))
	}
	if layer.Name == "" {
		addRefusal("missing_layer_name", layerPath+".name", "layer name is required")
	}
	if layer.Type == "shape" && layer.Shape != nil {
		switch layer.Shape.Kind {
		case "rect", "ellipse":
			if layer.Shape.Kind == "rect" {
				recordCapability("RectNode.SetSize", layerPath+".shape.size")
			} else {
				recordCapability("EllipseNode.SetSize", layerPath+".shape.size")
			}
		default:
			addRefusal("unsupported_shape_kind", layerPath+".shape.kind", fmt.Sprintf("unsupported shape kind %q", layer.Shape.Kind))
		}
		if len(layer.Shape.FillColor) > 0 {
			recordCapability("FillNode.SetColor", layerPath+".shape.fill_color")
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
		validateVec(kf.Value, 2, kfPath+".value", addRefusal)
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

func usesTransform(t Transform) bool {
	return len(t.Position) > 0 ||
		len(t.Scale) > 0 ||
		len(t.AnchorPoint) > 0 ||
		t.Rotation != nil ||
		t.Opacity != nil ||
		len(t.PositionKeyframes) > 0
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
