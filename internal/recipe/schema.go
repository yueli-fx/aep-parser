package recipe

import "fmt"

const SchemaVersion = 1

type Recipe struct {
	SchemaVersion int         `json:"schema_version"`
	Project       ProjectSpec `json:"project"`
	Comps         []CompSpec  `json:"comps"`
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
	Kind      string    `json:"kind"`
	Size      []float64 `json:"size,omitempty"`
	FillColor []float64 `json:"fill_color,omitempty"`
}

type Effect struct {
	MatchName string `json:"match_name"`
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

type Report struct {
	SchemaVersion int       `json:"schema_version"`
	Valid         bool      `json:"valid"`
	OutputPath    string    `json:"output_path,omitempty"`
	Refusals      []Refusal `json:"refusals,omitempty"`
}

type Refusal struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
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
	for ci, comp := range rec.Comps {
		compPath := fmt.Sprintf("comps[%d]", ci)
		if comp.Name == "" {
			addRefusal("missing_comp_name", compPath+".name", "comp name is required")
		}
		if comp.Width <= 0 || comp.Height <= 0 || comp.FrameRate <= 0 || comp.Duration <= 0 {
			addRefusal("invalid_comp_timing_or_size", compPath, "width, height, frame_rate, and duration must be positive")
		}
		for li, layer := range comp.Layers {
			layerPath := fmt.Sprintf("%s.layers[%d]", compPath, li)
			validateLayer(layer, layerPath, comp.Duration, caps, addRefusal)
		}
	}
	return report
}

func validateLayer(layer Layer, layerPath string, compDuration float64, caps CapabilityIndex, addRefusal func(string, string, string)) {
	switch layer.Type {
	case "solid", "text", "shape":
	default:
		addRefusal("unsupported_layer_type", layerPath+".type", fmt.Sprintf("unsupported layer type %q", layer.Type))
	}
	if layer.Name == "" {
		addRefusal("missing_layer_name", layerPath+".name", "layer name is required")
	}
	if layer.Type == "shape" && layer.Shape != nil {
		switch layer.Shape.Kind {
		case "rect", "ellipse":
		default:
			addRefusal("unsupported_shape_kind", layerPath+".shape.kind", fmt.Sprintf("unsupported shape kind %q", layer.Shape.Kind))
		}
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
		if effect.MatchName == "" {
			addRefusal("missing_effect_match_name", effectPath+".match_name", "effect match_name is required")
			continue
		}
		if caps.Lookup(effect.MatchName) == CapabilityUnsupported {
			addRefusal("unsupported_effect", effectPath+".match_name", fmt.Sprintf("effect %q is unsupported", effect.MatchName))
		}
	}
}

func validateVec(values []float64, want int, path string, addRefusal func(string, string, string)) {
	if len(values) == 0 {
		return
	}
	if len(values) != want {
		addRefusal("invalid_vector_size", path, fmt.Sprintf("expected %d values", want))
	}
}
