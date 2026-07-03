package selfhost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

type PseudoBehaviorPayloadOptions struct {
	PlanPath    string
	Root        string
	OutDir      string
	Limit       int
	MaxExamples int
}

type PseudoBehaviorPayloadReport struct {
	SchemaVersion int                           `json:"schema_version"`
	SourcePlan    string                        `json:"source_plan"`
	SampleRoot    string                        `json:"sample_root"`
	OutputPath    string                        `json:"output_path,omitempty"`
	Summary       PseudoBehaviorPayloadSummary  `json:"summary"`
	Families      []PseudoBehaviorPayloadFamily `json:"families"`
	Boundaries    []string                      `json:"boundaries"`
}

type PseudoBehaviorPayloadSummary struct {
	PlannedFamilies         int `json:"planned_families"`
	PlannedControls         int `json:"planned_controls"`
	BehaviorTasks           int `json:"behavior_tasks"`
	ScannedProjects         int `json:"scanned_projects"`
	FailedProjects          int `json:"failed_projects"`
	ControlsWithPayloads    int `json:"controls_with_payloads"`
	ControlsMissingPayloads int `json:"controls_missing_payloads"`
	KeyframePayloads        int `json:"keyframe_payloads"`
	ExpressionPayloads      int `json:"expression_payloads"`
}

type PseudoBehaviorPayloadFamily struct {
	MatchName string                         `json:"match_name"`
	UID       string                         `json:"uid,omitempty"`
	Status    string                         `json:"status,omitempty"`
	Controls  []PseudoBehaviorPayloadControl `json:"controls"`
}

type PseudoBehaviorPayloadControl struct {
	ParamMatchName string                         `json:"param_match_name"`
	Label          string                         `json:"label,omitempty"`
	ControlKind    string                         `json:"control_kind,omitempty"`
	Action         string                         `json:"action"`
	Phase          string                         `json:"phase"`
	Status         string                         `json:"status"`
	Examples       []PseudoBehaviorPayloadExample `json:"examples,omitempty"`
	MissReason     string                         `json:"miss_reason,omitempty"`
}

type PseudoBehaviorPayloadExample struct {
	ProjectPath       string             `json:"project_path"`
	CompName          string             `json:"comp_name,omitempty"`
	LayerName         string             `json:"layer_name,omitempty"`
	EffectMatchName   string             `json:"effect_match_name"`
	ParamMatchName    string             `json:"param_match_name"`
	StaticValue       any                `json:"static_value,omitempty"`
	Keyframes         []profile.Keyframe `json:"keyframes,omitempty"`
	Expression        string             `json:"expression,omitempty"`
	ExpressionEnabled *bool              `json:"expression_enabled,omitempty"`
}

func RunPseudoBehaviorPayloadExtraction(opts PseudoBehaviorPayloadOptions) (PseudoBehaviorPayloadReport, error) {
	if opts.PlanPath == "" {
		opts.PlanPath = filepath.Join("tmp", "pseudo_behavior_wiring", "plan.json")
	}
	if opts.Root == "" {
		opts.Root = filepath.Join("data", "samples")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join("tmp", "pseudo_behavior_payloads")
	}
	data, err := os.ReadFile(opts.PlanPath)
	if err != nil {
		return PseudoBehaviorPayloadReport{}, err
	}
	var plan PseudoBehaviorWiringPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return PseudoBehaviorPayloadReport{}, fmt.Errorf("parse pseudo behavior plan %s: %w", opts.PlanPath, err)
	}
	inputs, err := DiscoverSampleShellInputs(opts.Root, opts.Limit)
	if err != nil {
		return PseudoBehaviorPayloadReport{}, err
	}
	profiles := make([]*profile.Profile, 0, len(inputs))
	failed := 0
	for _, input := range inputs {
		prof, err := openProfile(input)
		if err != nil {
			failed++
			continue
		}
		profiles = append(profiles, prof)
	}
	report := BuildPseudoBehaviorPayloadReport(plan, profiles, opts)
	report.Summary.ScannedProjects = len(profiles)
	report.Summary.FailedProjects = failed
	report.OutputPath = filepath.Join(opts.OutDir, "payloads.json")
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return PseudoBehaviorPayloadReport{}, err
	}
	if err := writeSampleShellJSON(report.OutputPath, report); err != nil {
		return PseudoBehaviorPayloadReport{}, err
	}
	return report, nil
}

func BuildPseudoBehaviorPayloadReport(plan PseudoBehaviorWiringPlan, profiles []*profile.Profile, opts PseudoBehaviorPayloadOptions) PseudoBehaviorPayloadReport {
	maxExamples := opts.MaxExamples
	if maxExamples <= 0 {
		maxExamples = 5
	}
	report := PseudoBehaviorPayloadReport{
		SchemaVersion: 1,
		SourcePlan:    opts.PlanPath,
		SampleRoot:    opts.Root,
		Boundaries: []string{
			"Extracted payloads are source facts only.",
			"Applying payloads to generated pseudo controls is a later package.",
			"Expression semantic correctness is not proven by this extraction.",
		},
	}
	report.Summary.PlannedFamilies = len(plan.Families)
	for _, sourceFamily := range plan.Families {
		family := PseudoBehaviorPayloadFamily{
			MatchName: sourceFamily.MatchName,
			UID:       sourceFamily.UID,
			Status:    sourceFamily.Status,
			Controls:  []PseudoBehaviorPayloadControl{},
		}
		for _, task := range sourceFamily.Tasks {
			report.Summary.PlannedControls++
			if task.Action == "" || task.Action == "static_control_preserved" {
				continue
			}
			report.Summary.BehaviorTasks++
			control := PseudoBehaviorPayloadControl{
				ParamMatchName: task.ParamMatchName,
				Label:          task.Label,
				ControlKind:    task.ControlKind,
				Action:         task.Action,
				Phase:          task.Phase,
				Status:         "missing_payload",
				MissReason:     "no matching keyframe or expression payload in scanned samples",
			}
			control.Examples = extractPseudoBehaviorExamples(profiles, sourceFamily.MatchName, task, maxExamples)
			if len(control.Examples) > 0 {
				control.Status = "payload_found"
				control.MissReason = ""
				report.Summary.ControlsWithPayloads++
				if payloadExamplesHaveKeyframes(control.Examples) {
					report.Summary.KeyframePayloads++
				}
				if payloadExamplesHaveExpression(control.Examples) {
					report.Summary.ExpressionPayloads++
				}
			} else {
				report.Summary.ControlsMissingPayloads++
			}
			family.Controls = append(family.Controls, control)
		}
		if len(family.Controls) > 0 {
			report.Families = append(report.Families, family)
		}
	}
	return report
}

func extractPseudoBehaviorExamples(profiles []*profile.Profile, effectMatchName string, task PseudoBehaviorWiringTask, maxExamples int) []PseudoBehaviorPayloadExample {
	var examples []PseudoBehaviorPayloadExample
	for _, prof := range profiles {
		if prof == nil {
			continue
		}
		for _, comp := range prof.Comps {
			for _, layer := range comp.Layers {
				for _, effect := range layer.Effects {
					if effect.MatchName != effectMatchName {
						continue
					}
					for _, param := range effect.Params {
						if param.MatchName != task.ParamMatchName || !pseudoBehaviorParamHasPayload(task, param) {
							continue
						}
						examples = append(examples, PseudoBehaviorPayloadExample{
							ProjectPath:       prof.Meta.Path,
							CompName:          comp.Name,
							LayerName:         layer.Name,
							EffectMatchName:   effect.MatchName,
							ParamMatchName:    param.MatchName,
							StaticValue:       param.StaticValue,
							Keyframes:         append([]profile.Keyframe(nil), param.Keyframes...),
							Expression:        param.Expression,
							ExpressionEnabled: cloneBoolPtr(param.ExpressionEnabled),
						})
						if len(examples) >= maxExamples {
							return examples
						}
					}
				}
			}
		}
	}
	return examples
}

func pseudoBehaviorParamHasPayload(task PseudoBehaviorWiringTask, param profile.Property) bool {
	switch task.Action {
	case "extract_keyframes_then_apply":
		return len(param.Keyframes) > 0
	case "extract_expression_then_apply":
		return param.Expression != "" || param.ExpressionEnabled != nil
	case "resolve_keyframe_expression_conflict":
		return len(param.Keyframes) > 0 || param.Expression != "" || param.ExpressionEnabled != nil
	default:
		return false
	}
}

func payloadExamplesHaveKeyframes(examples []PseudoBehaviorPayloadExample) bool {
	for _, example := range examples {
		if len(example.Keyframes) > 0 {
			return true
		}
	}
	return false
}

func payloadExamplesHaveExpression(examples []PseudoBehaviorPayloadExample) bool {
	for _, example := range examples {
		if example.Expression != "" || example.ExpressionEnabled != nil {
			return true
		}
	}
	return false
}

func cloneBoolPtr(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
