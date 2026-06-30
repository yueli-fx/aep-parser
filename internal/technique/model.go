// Package technique derives project-understanding facts from normalized
// profiles. It does not parse or write AEP data directly.
package technique

import "github.com/yueli-fx/aep-parser/internal/profile"

const SchemaVersion = 1

type FactSet struct {
	SchemaVersion  int                 `json:"schema_version"`
	SourcePath     string              `json:"source_path,omitempty"`
	Summary        Summary             `json:"summary"`
	Comps          []CompFact          `json:"comps,omitempty"`
	Layers         []LayerFact         `json:"layers,omitempty"`
	Effects        []EffectFact        `json:"effects,omitempty"`
	TextAnimators  []TextAnimatorFact  `json:"text_animators,omitempty"`
	ShapeOperators []ShapeOperatorFact `json:"shape_operators,omitempty"`
	Dependencies   []DependencyFact    `json:"dependencies,omitempty"`
	Unknowns       []UnknownFact       `json:"unknowns,omitempty"`
}

type Summary struct {
	CompCount       int    `json:"comp_count"`
	LayerCount      int    `json:"layer_count"`
	EffectCount     int    `json:"effect_count"`
	TextLayerCount  int    `json:"text_layer_count,omitempty"`
	ShapeLayerCount int    `json:"shape_layer_count,omitempty"`
	UnknownCount    int    `json:"unknown_count,omitempty"`
	MainCompID      uint32 `json:"main_comp_id,omitempty"`
	MainCompName    string `json:"main_comp_name,omitempty"`
}

type CompFact struct {
	ID            uint32           `json:"id,omitempty"`
	Name          string           `json:"name"`
	Width         uint16           `json:"width"`
	Height        uint16           `json:"height"`
	FrameRate     float64          `json:"frame_rate"`
	Duration      float64          `json:"duration_seconds"`
	LayerCount    int              `json:"layer_count"`
	MainCandidate bool             `json:"main_candidate,omitempty"`
	Path          profile.PathRef  `json:"path"`
	Evidence      profile.Evidence `json:"evidence"`
}

type LayerFact struct {
	CompName   string           `json:"comp_name,omitempty"`
	ID         uint32           `json:"id,omitempty"`
	Index      int              `json:"index,omitempty"`
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	Role       string           `json:"role"`
	Confidence string           `json:"confidence"`
	Path       profile.PathRef  `json:"path"`
	Evidence   profile.Evidence `json:"evidence"`
}

type EffectFact struct {
	CompName          string           `json:"comp_name,omitempty"`
	LayerName         string           `json:"layer_name,omitempty"`
	MatchName         string           `json:"match_name"`
	DisplayName       string           `json:"display_name,omitempty"`
	DependencyClass   string           `json:"dependency_class,omitempty"`
	Occurrence        int              `json:"occurrence,omitempty"`
	ChangedParamCount int              `json:"changed_param_count,omitempty"`
	TunedParamCount   int              `json:"tuned_param_count,omitempty"`
	UnknownParamCount int              `json:"unknown_param_count,omitempty"`
	HasExpression     bool             `json:"has_expression,omitempty"`
	HasKeyframes      bool             `json:"has_keyframes,omitempty"`
	HasLayerRef       bool             `json:"has_layer_ref,omitempty"`
	Path              profile.PathRef  `json:"path"`
	Evidence          profile.Evidence `json:"evidence"`
}

type TextAnimatorFact struct {
	CompName       string           `json:"comp_name,omitempty"`
	LayerName      string           `json:"layer_name,omitempty"`
	PropertyKind   string           `json:"property_kind"`
	PropertyName   string           `json:"property_name,omitempty"`
	MatchName      string           `json:"match_name"`
	HasStaticValue bool             `json:"has_static_value,omitempty"`
	HasExpression  bool             `json:"has_expression,omitempty"`
	HasKeyframes   bool             `json:"has_keyframes,omitempty"`
	Path           profile.PathRef  `json:"path"`
	Evidence       profile.Evidence `json:"evidence"`
}

type ShapeOperatorFact struct {
	CompName  string           `json:"comp_name,omitempty"`
	LayerName string           `json:"layer_name,omitempty"`
	Family    string           `json:"family"`
	Source    string           `json:"source"`
	MatchName string           `json:"match_name,omitempty"`
	Path      profile.PathRef  `json:"path"`
	Evidence  profile.Evidence `json:"evidence"`
}

type DependencyFact struct {
	CompName   string           `json:"comp_name,omitempty"`
	SourceName string           `json:"source_name,omitempty"`
	SourceID   uint32           `json:"source_id,omitempty"`
	Relation   string           `json:"relation"`
	TargetName string           `json:"target_name,omitempty"`
	TargetID   uint32           `json:"target_id,omitempty"`
	TargetKind string           `json:"target_kind,omitempty"`
	Property   string           `json:"property,omitempty"`
	Path       profile.PathRef  `json:"path"`
	Evidence   profile.Evidence `json:"evidence"`
}

type UnknownFact struct {
	CompName   string           `json:"comp_name,omitempty"`
	LayerName  string           `json:"layer_name,omitempty"`
	EffectName string           `json:"effect_name,omitempty"`
	Path       string           `json:"path,omitempty"`
	Reason     string           `json:"reason"`
	Evidence   profile.Evidence `json:"evidence"`
}
