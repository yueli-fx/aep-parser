package selfhost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type PseudoBehaviorWiringOptions struct {
	ProofPath string
	OutDir    string
}

type PseudoBehaviorWiringPlan struct {
	SchemaVersion int                          `json:"schema_version"`
	SourceProof   string                       `json:"source_proof"`
	OutputPath    string                       `json:"output_path,omitempty"`
	Summary       PseudoBehaviorWiringSummary  `json:"summary"`
	Families      []PseudoBehaviorWiringFamily `json:"families"`
	Boundaries    []string                     `json:"boundaries"`
}

type PseudoBehaviorWiringSummary struct {
	Families          int `json:"families"`
	Controls          int `json:"controls"`
	StaticControls    int `json:"static_controls"`
	KeyframeTasks     int `json:"keyframe_tasks"`
	ExpressionTasks   int `json:"expression_tasks"`
	ConflictTasks     int `json:"conflict_tasks"`
	GeneratedFamilies int `json:"generated_families"`
}

type PseudoBehaviorWiringFamily struct {
	MatchName    string                     `json:"match_name"`
	UID          string                     `json:"uid"`
	Status       string                     `json:"status"`
	GeneratedAEP string                     `json:"generated_aep,omitempty"`
	Tasks        []PseudoBehaviorWiringTask `json:"tasks"`
}

type PseudoBehaviorWiringTask struct {
	ParamMatchName string   `json:"param_match_name"`
	Label          string   `json:"label"`
	ControlKind    string   `json:"control_kind"`
	Action         string   `json:"action"`
	Phase          string   `json:"phase"`
	InputsNeeded   []string `json:"inputs_needed,omitempty"`
	Notes          []string `json:"notes,omitempty"`
}

func RunPseudoBehaviorWiringPlan(opts PseudoBehaviorWiringOptions) (PseudoBehaviorWiringPlan, error) {
	if opts.ProofPath == "" {
		opts.ProofPath = filepath.Join("tmp", "pseudo_controller_rebuild", "proof.json")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join("tmp", "pseudo_behavior_wiring")
	}
	data, err := os.ReadFile(opts.ProofPath)
	if err != nil {
		return PseudoBehaviorWiringPlan{}, err
	}
	var proof PseudoControllerRebuildProof
	if err := json.Unmarshal(data, &proof); err != nil {
		return PseudoBehaviorWiringPlan{}, fmt.Errorf("parse pseudo controller proof %s: %w", opts.ProofPath, err)
	}
	plan := BuildPseudoBehaviorWiringPlan(proof, opts)
	plan.SourceProof = opts.ProofPath
	plan.OutputPath = filepath.Join(opts.OutDir, "plan.json")
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return PseudoBehaviorWiringPlan{}, err
	}
	if err := writeIndentedJSON(plan.OutputPath, plan); err != nil {
		return PseudoBehaviorWiringPlan{}, err
	}
	return plan, nil
}

func BuildPseudoBehaviorWiringPlan(proof PseudoControllerRebuildProof, opts PseudoBehaviorWiringOptions) PseudoBehaviorWiringPlan {
	plan := PseudoBehaviorWiringPlan{
		SchemaVersion: 1,
		SourceProof:   opts.ProofPath,
		Boundaries: []string{
			"Controller structure is rebuildable from the pseudo-controller proof.",
			"Keyframe and expression data is not reconstructed by this package.",
			"Exact behavior requires later extraction from source project facts and targeted application to generated controls.",
		},
	}
	families := append([]PseudoControllerRebuildFamily(nil), proof.Families...)
	sort.Slice(families, func(i, j int) bool {
		if families[i].Status != families[j].Status {
			return families[i].Status == "generated"
		}
		return families[i].MatchName < families[j].MatchName
	})
	for _, sourceFamily := range families {
		family := buildPseudoBehaviorWiringFamily(sourceFamily)
		plan.Families = append(plan.Families, family)
		plan.Summary.Families++
		if sourceFamily.Status == "generated" {
			plan.Summary.GeneratedFamilies++
		}
		for _, task := range family.Tasks {
			plan.Summary.Controls++
			switch task.Action {
			case "static_control_preserved":
				plan.Summary.StaticControls++
			case "extract_keyframes_then_apply":
				plan.Summary.KeyframeTasks++
			case "extract_expression_then_apply":
				plan.Summary.ExpressionTasks++
			case "resolve_keyframe_expression_conflict":
				plan.Summary.ConflictTasks++
				plan.Summary.KeyframeTasks++
				plan.Summary.ExpressionTasks++
			}
		}
	}
	return plan
}

func buildPseudoBehaviorWiringFamily(source PseudoControllerRebuildFamily) PseudoBehaviorWiringFamily {
	controls := append([]PseudoControllerControlPlan(nil), source.Controls...)
	sort.Slice(controls, func(i, j int) bool { return controls[i].ParamMatchName < controls[j].ParamMatchName })
	family := PseudoBehaviorWiringFamily{
		MatchName:    source.MatchName,
		UID:          source.UID,
		Status:       source.Status,
		GeneratedAEP: source.GeneratedAEP,
		Tasks:        make([]PseudoBehaviorWiringTask, 0, len(controls)),
	}
	for _, control := range controls {
		family.Tasks = append(family.Tasks, pseudoBehaviorTaskFromControl(control))
	}
	return family
}

func pseudoBehaviorTaskFromControl(control PseudoControllerControlPlan) PseudoBehaviorWiringTask {
	task := PseudoBehaviorWiringTask{
		ParamMatchName: control.ParamMatchName,
		Label:          control.Label,
		ControlKind:    control.ControlKind,
	}
	hasKeyframes := stringSliceContains(control.BehaviorNotes, "keyframed")
	hasExpression := stringSliceContains(control.BehaviorNotes, "expression")
	switch {
	case hasKeyframes && hasExpression:
		task.Action = "resolve_keyframe_expression_conflict"
		task.Phase = "keyframes_then_expression"
		task.InputsNeeded = []string{"source_keyframes", "source_expression"}
		task.Notes = []string{"Apply extracted keyframes first, then apply expression only if source behavior confirms expression should override or reference the keyframes."}
	case hasKeyframes:
		task.Action = "extract_keyframes_then_apply"
		task.Phase = "keyframes"
		task.InputsNeeded = []string{"source_keyframes"}
	case hasExpression:
		task.Action = "extract_expression_then_apply"
		task.Phase = "expression"
		task.InputsNeeded = []string{"source_expression"}
	default:
		task.Action = "static_control_preserved"
		task.Phase = "none"
	}
	return task
}
