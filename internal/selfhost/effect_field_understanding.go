package selfhost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type EffectFieldUnderstandingOptions struct {
	InventoryPath string
	OutPath       string
}

type EffectFieldUnderstanding struct {
	SchemaVersion   int                              `json:"schema_version"`
	SourceInventory string                           `json:"source_inventory"`
	OutputPath      string                           `json:"output_path,omitempty"`
	Summary         EffectFieldUnderstandingSummary  `json:"summary"`
	Effects         []EffectFieldUnderstandingEffect `json:"effects"`
}

type EffectFieldUnderstandingSummary struct {
	EffectKinds        int        `json:"effect_kinds"`
	EffectOccurrences  int        `json:"effect_occurrences"`
	ParamKinds         int        `json:"param_kinds"`
	ParamOccurrences   int        `json:"param_occurrences"`
	Classes            []CountRow `json:"classes,omitempty"`
	Reproducibility    []CountRow `json:"reproducibility,omitempty"`
	GenerationPolicies []CountRow `json:"generation_policies,omitempty"`
	StudyActions       []CountRow `json:"study_actions,omitempty"`
}

type EffectFieldUnderstandingEffect struct {
	MatchName          string   `json:"match_name"`
	Class              string   `json:"class"`
	Occurrences        int      `json:"occurrences"`
	ParamKinds         int      `json:"param_kinds"`
	ParamOccurrences   int      `json:"param_occurrences"`
	Capabilities       []string `json:"capabilities,omitempty"`
	Reproducibility    string   `json:"reproducibility"`
	GenerationPolicy   string   `json:"generation_policy"`
	FieldUnderstanding string   `json:"field_understanding"`
	StudyPriority      int      `json:"study_priority"`
	StudyActions       []string `json:"study_actions,omitempty"`
	Boundary           string   `json:"boundary,omitempty"`
}

func RunEffectFieldUnderstanding(opts EffectFieldUnderstandingOptions) (EffectFieldUnderstanding, error) {
	if opts.InventoryPath == "" {
		opts.InventoryPath = filepath.Join("tmp", "effect_field_inventory", "inventory.json")
	}
	if opts.OutPath == "" {
		opts.OutPath = filepath.Join("tmp", "effect_field_understanding", "understanding.json")
	}
	data, err := os.ReadFile(opts.InventoryPath)
	if err != nil {
		return EffectFieldUnderstanding{}, err
	}
	var inventory EffectFieldInventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return EffectFieldUnderstanding{}, err
	}
	understanding := BuildEffectFieldUnderstanding(inventory, opts.InventoryPath)
	understanding.OutputPath = opts.OutPath
	if err := os.MkdirAll(filepath.Dir(opts.OutPath), 0o755); err != nil {
		return EffectFieldUnderstanding{}, err
	}
	if err := writeIndentedJSON(opts.OutPath, understanding); err != nil {
		return EffectFieldUnderstanding{}, err
	}
	return understanding, nil
}

func BuildEffectFieldUnderstanding(inventory EffectFieldInventory, sourceInventory string) EffectFieldUnderstanding {
	understanding := EffectFieldUnderstanding{
		SchemaVersion:   1,
		SourceInventory: sourceInventory,
		Summary: EffectFieldUnderstandingSummary{
			EffectKinds:       inventory.Summary.EffectKinds,
			EffectOccurrences: inventory.Summary.EffectOccurrences,
			ParamKinds:        inventory.Summary.ParamKinds,
			ParamOccurrences:  inventory.Summary.ParamOccurrences,
		},
		Effects: make([]EffectFieldUnderstandingEffect, 0, len(inventory.Effects)),
	}
	if understanding.Summary.EffectKinds == 0 {
		understanding.Summary.EffectKinds = len(inventory.Effects)
	}
	classCounts := map[string]int{}
	reproCounts := map[string]int{}
	policyCounts := map[string]int{}
	actionCounts := map[string]int{}
	var effectOccurrences, paramKinds, paramOccurrences int
	for _, effect := range inventory.Effects {
		row := buildEffectFieldUnderstandingEffect(effect)
		understanding.Effects = append(understanding.Effects, row)
		effectOccurrences += row.Occurrences
		paramKinds += row.ParamKinds
		paramOccurrences += row.ParamOccurrences
		weight := row.Occurrences
		if weight == 0 {
			weight = 1
		}
		classCounts[row.Class] += weight
		reproCounts[row.Reproducibility] += weight
		policyCounts[row.GenerationPolicy] += weight
		for _, action := range row.StudyActions {
			actionCounts[action] += weight
		}
	}
	if understanding.Summary.EffectOccurrences == 0 {
		understanding.Summary.EffectOccurrences = effectOccurrences
	}
	if understanding.Summary.ParamKinds == 0 {
		understanding.Summary.ParamKinds = paramKinds
	}
	if understanding.Summary.ParamOccurrences == 0 {
		understanding.Summary.ParamOccurrences = paramOccurrences
	}
	sort.Slice(understanding.Effects, func(i, j int) bool {
		if understanding.Effects[i].StudyPriority != understanding.Effects[j].StudyPriority {
			return understanding.Effects[i].StudyPriority > understanding.Effects[j].StudyPriority
		}
		if understanding.Effects[i].Occurrences != understanding.Effects[j].Occurrences {
			return understanding.Effects[i].Occurrences > understanding.Effects[j].Occurrences
		}
		return understanding.Effects[i].MatchName < understanding.Effects[j].MatchName
	})
	understanding.Summary.Classes = countMapRows(classCounts)
	understanding.Summary.Reproducibility = countMapRows(reproCounts)
	understanding.Summary.GenerationPolicies = countMapRows(policyCounts)
	understanding.Summary.StudyActions = countMapRows(actionCounts)
	return understanding
}

func buildEffectFieldUnderstandingEffect(effect EffectFieldInventoryEffect) EffectFieldUnderstandingEffect {
	reproducibility, generationPolicy, fieldUnderstanding, boundary := effectFieldUnderstandingPolicy(effect.Class)
	actions := effectFieldStudyActions(effect)
	row := EffectFieldUnderstandingEffect{
		MatchName:          effect.MatchName,
		Class:              effect.Class,
		Occurrences:        effect.Occurrences,
		ParamKinds:         effect.ParamKinds,
		ParamOccurrences:   effect.ParamOccurrences,
		Capabilities:       append([]string(nil), effect.InferredCapabilities...),
		Reproducibility:    reproducibility,
		GenerationPolicy:   generationPolicy,
		FieldUnderstanding: fieldUnderstanding,
		StudyActions:       actions,
		Boundary:           boundary,
	}
	row.StudyPriority = effectFieldStudyPriority(row)
	return row
}

func effectFieldUnderstandingPolicy(class string) (string, string, string, string) {
	switch class {
	case "native_supported":
		return "go_native_exact", "emit_native", "native_effect_can_be_emitted_by_go", ""
	case "native_template_gap":
		return "native_template_possible", "add_native_template", "param_names_and_values_parseable", "Native or bundled effect appears parseable, but generation needs a template/API support gap closed first."
	case "pseudo":
		return "pseudo_rebuildable", "rebuild_pseudo_controls", "param_names_and_values_parseable", "Pseudo effect fields are parseable controls; rebuild can use BuildPseudoEffect only when the control model is sufficient."
	case "third_party":
		return "third_party_plugin_required", "preserve_as_dependency", "param_names_and_values_parseable", "Rendering equivalence requires the plugin unless a separate native approximation recipe is authored."
	default:
		return "unknown_or_alias_pending", "classify_or_alias", "class_or_alias_needs_review", "Effect class or alias needs review before any generation claim."
	}
}

func effectFieldStudyActions(effect EffectFieldInventoryEffect) []string {
	actions := map[string]struct{}{}
	switch effect.Class {
	case "pseudo":
		actions["summarize_param_groups"] = struct{}{}
		actions["map_controller_inputs"] = struct{}{}
	case "third_party":
		actions["summarize_param_groups"] = struct{}{}
		actions["preserve_plugin_dependency"] = struct{}{}
	case "native_template_gap":
		actions["extract_native_template"] = struct{}{}
	case "unknown_or_alias":
		actions["classify_or_alias"] = struct{}{}
	}
	if effect.Class != "native_supported" && hasEffectCapability(effect.InferredCapabilities, "layer_ref") {
		actions["map_controller_inputs"] = struct{}{}
	}
	if hasEffectCapability(effect.InferredCapabilities, "expression") || hasEffectCapability(effect.InferredCapabilities, "keyframes") {
		actions["summarize_motion_controls"] = struct{}{}
	}
	out := make([]string, 0, len(actions))
	for action := range actions {
		out = append(out, action)
	}
	sort.Strings(out)
	return out
}

func effectFieldStudyPriority(effect EffectFieldUnderstandingEffect) int {
	score := effect.Occurrences*10 + effect.ParamKinds
	if effect.Class == "third_party" || effect.Class == "pseudo" {
		score += 100
	}
	if hasEffectCapability(effect.Capabilities, "keyframes") {
		score += 50
	}
	if hasEffectCapability(effect.Capabilities, "expression") {
		score += 50
	}
	if hasEffectCapability(effect.Capabilities, "layer_ref") {
		score += 30
	}
	if hasEffectCapability(effect.Capabilities, "color") || hasEffectCapability(effect.Capabilities, "position") {
		score += 20
	}
	return score
}

func hasEffectCapability(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func (e EffectFieldUnderstandingEffect) String() string {
	return fmt.Sprintf("%s %s %s", e.MatchName, e.Reproducibility, e.GenerationPolicy)
}
