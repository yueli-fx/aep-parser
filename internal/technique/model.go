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

type Portrait struct {
	SchemaVersion  int                `json:"schema_version"`
	SourcePath     string             `json:"source_path,omitempty"`
	Fingerprint    FingerprintSummary `json:"fingerprint"`
	SignalLayers   []SignalLayer      `json:"signal_layers,omitempty"`
	Mechanisms     MechanismSummary   `json:"mechanisms"`
	Graph          GraphSummary       `json:"graph"`
	TechniqueHints []TechniqueHint    `json:"technique_hints,omitempty"`
	Unknowns       UnknownSummary     `json:"unknowns"`
}

type FingerprintSummary struct {
	CompCount          int            `json:"comp_count"`
	LayerCount         int            `json:"layer_count"`
	EffectCount        int            `json:"effect_count"`
	TextLayerCount     int            `json:"text_layer_count,omitempty"`
	ShapeLayerCount    int            `json:"shape_layer_count,omitempty"`
	TextAnimatorCount  int            `json:"text_animator_count,omitempty"`
	ShapeOperatorCount int            `json:"shape_operator_count,omitempty"`
	DependencyCount    int            `json:"dependency_count,omitempty"`
	UnknownCount       int            `json:"unknown_count,omitempty"`
	LayerRoleCounts    map[string]int `json:"layer_role_counts"`
}

type SignalLayer struct {
	CompName  string   `json:"comp_name,omitempty"`
	LayerName string   `json:"layer_name"`
	Role      string   `json:"role,omitempty"`
	Score     int      `json:"score"`
	Signals   []string `json:"signals"`
}

type MechanismSummary struct {
	EffectClassCounts           map[string]int `json:"effect_class_counts"`
	EffectMatchCounts           map[string]int `json:"effect_match_counts"`
	ThirdPartyEffectMatchCounts map[string]int `json:"third_party_effect_match_counts,omitempty"`
	TextAnimatorKindCounts      map[string]int `json:"text_animator_kind_counts"`
	ShapeFamilyCounts           map[string]int `json:"shape_family_counts"`
	ReproducibilityCounts       map[string]int `json:"reproducibility_counts"`
}

type GraphSummary struct {
	EdgeCount      int            `json:"edge_count"`
	RelationCounts map[string]int `json:"relation_counts"`
}

type TechniqueHint struct {
	ID         string   `json:"id"`
	Confidence string   `json:"confidence"`
	Signals    []string `json:"signals,omitempty"`
}

type UnknownSummary struct {
	Count int `json:"count"`
}

type Explanation struct {
	SchemaVersion        int                    `json:"schema_version"`
	SourcePath           string                 `json:"source_path,omitempty"`
	Portrait             Portrait               `json:"portrait"`
	Overview             []string               `json:"overview,omitempty"`
	Archetypes           []ProjectArchetype     `json:"archetypes,omitempty"`
	RecreationReadiness  RecreationReadiness    `json:"recreation_readiness"`
	Techniques           []TechniqueExplanation `json:"techniques,omitempty"`
	TopSignalLayers      []SignalLayer          `json:"top_signal_layers,omitempty"`
	ReproducibilityNotes []string               `json:"reproducibility_notes,omitempty"`
	UnknownNotes         []string               `json:"unknown_notes,omitempty"`
}

type ProjectArchetype struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Score   int      `json:"score"`
	Summary string   `json:"summary"`
	Signals []string `json:"signals,omitempty"`
}

type RecreationReadiness struct {
	Status   string   `json:"status"`
	Summary  string   `json:"summary"`
	Blockers []string `json:"blockers,omitempty"`
	Notes    []string `json:"notes,omitempty"`
}

type TechniqueExplanation struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Confidence string   `json:"confidence,omitempty"`
	Summary    string   `json:"summary"`
	Evidence   []string `json:"evidence,omitempty"`
}
