package selfhost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type PseudoBehaviorApplicationOptions struct {
	ProofPath   string
	PayloadPath string
	OutDir      string
}

type PseudoBehaviorApplicationReport struct {
	SchemaVersion int                               `json:"schema_version"`
	SourceProof   string                            `json:"source_proof"`
	SourcePayload string                            `json:"source_payload"`
	OutputPath    string                            `json:"output_path,omitempty"`
	Summary       PseudoBehaviorApplicationSummary  `json:"summary"`
	Families      []PseudoBehaviorApplicationFamily `json:"families"`
	Boundaries    []string                          `json:"boundaries"`
}

type PseudoBehaviorApplicationSummary struct {
	Families                   int `json:"families"`
	GeneratedFamilies          int `json:"generated_families"`
	SkippedFamilies            int `json:"skipped_families"`
	GeneratedAEPs              int `json:"generated_aeps"`
	AppliedControls            int `json:"applied_controls"`
	SkippedControls            int `json:"skipped_controls"`
	ErrorControls              int `json:"error_controls"`
	VerifiedKeyframedControls  int `json:"verified_keyframed_controls"`
	ExpressionDeferredControls int `json:"expression_deferred_controls"`
}

type PseudoBehaviorApplicationFamily struct {
	MatchName    string                             `json:"match_name"`
	UID          string                             `json:"uid,omitempty"`
	Status       string                             `json:"status"`
	GeneratedAEP string                             `json:"generated_aep,omitempty"`
	Controls     []PseudoBehaviorApplicationControl `json:"controls,omitempty"`
	Verification PseudoBehaviorApplicationVerify    `json:"verification,omitempty"`
	Error        string                             `json:"error,omitempty"`
}

type PseudoBehaviorApplicationControl struct {
	ParamMatchName            string `json:"param_match_name"`
	Status                    string `json:"status"`
	AppliedKeyframes          int    `json:"applied_keyframes,omitempty"`
	VerifiedKeyframes         int    `json:"verified_keyframes,omitempty"`
	ExpressionDeferred        bool   `json:"expression_deferred,omitempty"`
	DeferredExpression        string `json:"deferred_expression,omitempty"`
	DeferredExpressionEnabled *bool  `json:"deferred_expression_enabled,omitempty"`
	ExpressionBoundary        string `json:"expression_boundary,omitempty"`
	Error                     string `json:"error,omitempty"`
}

type PseudoBehaviorApplicationVerify struct {
	Parsed            bool `json:"parsed,omitempty"`
	KeyframedControls int  `json:"keyframed_controls,omitempty"`
}

func RunPseudoBehaviorApplication(opts PseudoBehaviorApplicationOptions) (PseudoBehaviorApplicationReport, error) {
	if opts.ProofPath == "" {
		opts.ProofPath = filepath.Join("tmp", "pseudo_controller_rebuild", "proof.json")
	}
	if opts.PayloadPath == "" {
		opts.PayloadPath = filepath.Join("tmp", "pseudo_behavior_payloads", "payloads.json")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join("tmp", "pseudo_behavior_application")
	}
	proofData, err := os.ReadFile(opts.ProofPath)
	if err != nil {
		return PseudoBehaviorApplicationReport{}, err
	}
	var proof PseudoControllerRebuildProof
	if err := json.Unmarshal(proofData, &proof); err != nil {
		return PseudoBehaviorApplicationReport{}, fmt.Errorf("parse pseudo controller proof %s: %w", opts.ProofPath, err)
	}
	payloadData, err := os.ReadFile(opts.PayloadPath)
	if err != nil {
		return PseudoBehaviorApplicationReport{}, err
	}
	var payload PseudoBehaviorPayloadReport
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		return PseudoBehaviorApplicationReport{}, fmt.Errorf("parse pseudo behavior payloads %s: %w", opts.PayloadPath, err)
	}
	report := BuildPseudoBehaviorApplicationReport(proof, payload, opts)
	if err := os.MkdirAll(filepath.Join(opts.OutDir, "generated"), 0o755); err != nil {
		return PseudoBehaviorApplicationReport{}, err
	}
	for i := range report.Families {
		if report.Families[i].Status != "pending_application" {
			continue
		}
		source := proofFamilyByMatch(proof, report.Families[i].MatchName)
		payloadFamily := payloadFamilyByMatch(payload, report.Families[i].MatchName)
		if source == nil || payloadFamily == nil {
			continue
		}
		outPath := filepath.Join(opts.OutDir, "generated", safePseudoFamilySlug(source.UID)+".aep")
		applied, verified, err := writePseudoBehaviorApplicationAEP(outPath, *source, *payloadFamily)
		if err != nil {
			report.Families[i].Status = "error"
			report.Families[i].Error = err.Error()
			report.Summary.ErrorControls += len(report.Families[i].Controls)
			continue
		}
		report.Families[i].Status = "applied"
		report.Families[i].GeneratedAEP = outPath
		report.Families[i].Controls = applied
		report.Families[i].Verification = verified
		report.Summary.GeneratedAEPs++
		for _, control := range applied {
			switch control.Status {
			case "applied_keyframes":
				report.Summary.AppliedControls++
			case "expression_deferred":
				report.Summary.SkippedControls++
			case "error":
				report.Summary.ErrorControls++
			default:
				report.Summary.SkippedControls++
			}
			if control.ExpressionDeferred {
				report.Summary.ExpressionDeferredControls++
			}
			if control.VerifiedKeyframes > 0 {
				report.Summary.VerifiedKeyframedControls++
			}
		}
	}
	report.OutputPath = filepath.Join(opts.OutDir, "application.json")
	if err := writeSampleShellJSON(report.OutputPath, report); err != nil {
		return PseudoBehaviorApplicationReport{}, err
	}
	return report, nil
}

func BuildPseudoBehaviorApplicationReport(proof PseudoControllerRebuildProof, payload PseudoBehaviorPayloadReport, opts PseudoBehaviorApplicationOptions) PseudoBehaviorApplicationReport {
	report := PseudoBehaviorApplicationReport{
		SchemaVersion: 1,
		SourceProof:   opts.ProofPath,
		SourcePayload: opts.PayloadPath,
		Boundaries: []string{
			"Only generated pseudo families are applied.",
			"Keyframe application uses existing effect-param animation APIs for scalar and vector controls.",
			"Expression payload application is deferred.",
			"This does not prove full render-equivalent rig behavior.",
		},
	}
	for _, family := range proof.Families {
		row := PseudoBehaviorApplicationFamily{
			MatchName: family.MatchName,
			UID:       family.UID,
		}
		report.Summary.Families++
		if family.Status != "generated" {
			row.Status = "skipped_not_generated"
			report.Summary.SkippedFamilies++
			report.Families = append(report.Families, row)
			continue
		}
		report.Summary.GeneratedFamilies++
		if payloadFamilyByMatch(payload, family.MatchName) == nil {
			row.Status = "skipped_no_payload_family"
			report.Summary.SkippedFamilies++
			report.Families = append(report.Families, row)
			continue
		}
		row.Status = "pending_application"
		report.Families = append(report.Families, row)
	}
	return report
}

func writePseudoBehaviorApplicationAEP(path string, family PseudoControllerRebuildFamily, payloadFamily PseudoBehaviorPayloadFamily) ([]PseudoBehaviorApplicationControl, PseudoBehaviorApplicationVerify, error) {
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Pseudo Behavior Application", 640, 360, 30, 5)
	if err != nil {
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	if _, err := aep.NewShapeLayer(comp, "Controller Host"); err != nil {
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	host := reopened.Compositions[0].LayerByName("Controller Host")
	if host == nil {
		return nil, PseudoBehaviorApplicationVerify{}, fmt.Errorf("generated pseudo application host layer missing")
	}
	controls := make([]aep.PseudoControl, 0, len(family.Controls))
	for _, plan := range family.Controls {
		controls = append(controls, pseudoControlFromPlan(plan))
	}
	fx, err := aep.BuildPseudoEffect(host, family.UID, "Rebuild", family.UID+" Rebuild", controls)
	if err != nil {
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	rows := applyPseudoBehaviorControls(host, fx, payloadFamily)
	file, err := os.Create(path)
	if err != nil {
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	if err := reopened.WriteAEP(file); err != nil {
		_ = file.Close()
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	if err := file.Close(); err != nil {
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	prof, err := openProfile(path)
	if err != nil {
		return nil, PseudoBehaviorApplicationVerify{}, err
	}
	verify := verifyAppliedPseudoBehavior(prof, family.MatchName, rows)
	for i := range rows {
		rows[i].VerifiedKeyframes = verifiedKeyframesForControl(prof, family.MatchName, rows[i].ParamMatchName)
	}
	return rows, verify, nil
}

func applyPseudoBehaviorControls(layer *aep.Layer, fx *aep.Effect, payloadFamily PseudoBehaviorPayloadFamily) []PseudoBehaviorApplicationControl {
	rows := make([]PseudoBehaviorApplicationControl, 0, len(payloadFamily.Controls))
	for _, payloadControl := range payloadFamily.Controls {
		row := PseudoBehaviorApplicationControl{ParamMatchName: payloadControl.ParamMatchName}
		if len(payloadControl.Examples) == 0 {
			row.Status = "skipped_no_examples"
			rows = append(rows, row)
			continue
		}
		example := payloadControl.Examples[0]
		if payloadExampleHasExpression(example) {
			markExpressionDeferred(&row, example)
		}
		if len(example.Keyframes) == 0 {
			if row.ExpressionDeferred {
				row.Status = "expression_deferred"
			} else {
				row.Status = "skipped_no_keyframes"
			}
			rows = append(rows, row)
			continue
		}
		scalar, ok := scalarKeyframesFromProfile(example.Keyframes)
		if !ok {
			vector, vectorOK := vectorKeyframesFromProfile(example.Keyframes)
			if !vectorOK {
				row.Status = "error"
				row.Error = "keyframes are neither scalar nor 2/3/4-component vectors"
				rows = append(rows, row)
				continue
			}
			if _, err := aep.AnimateEffectParamVec(layer, fx, payloadControl.ParamMatchName, vector); err != nil {
				row.Status = "error"
				row.Error = err.Error()
				rows = append(rows, row)
				continue
			}
			row.Status = "applied_keyframes"
			row.AppliedKeyframes = len(vector)
			rows = append(rows, row)
			continue
		}
		if _, err := aep.AnimateEffectParam(layer, fx, payloadControl.ParamMatchName, scalar); err != nil {
			row.Status = "error"
			row.Error = err.Error()
			rows = append(rows, row)
			continue
		}
		row.Status = "applied_keyframes"
		row.AppliedKeyframes = len(scalar)
		rows = append(rows, row)
	}
	return rows
}

func payloadExampleHasExpression(example PseudoBehaviorPayloadExample) bool {
	return example.Expression != "" || example.ExpressionEnabled != nil
}

func markExpressionDeferred(row *PseudoBehaviorApplicationControl, example PseudoBehaviorPayloadExample) {
	row.ExpressionDeferred = true
	row.DeferredExpression = example.Expression
	row.DeferredExpressionEnabled = cloneBoolPtr(example.ExpressionEnabled)
	row.ExpressionBoundary = "expression payload captured as source fact; application is deferred"
}

func scalarKeyframesFromProfile(keyframes []profile.Keyframe) ([]aep.ScalarKeyframe, bool) {
	out := make([]aep.ScalarKeyframe, 0, len(keyframes))
	for _, keyframe := range keyframes {
		value, ok := scalarValue(keyframe.Value)
		if !ok {
			return nil, false
		}
		out = append(out, aep.ScalarKeyframe{Time: keyframe.Time, Value: value})
	}
	return out, len(out) >= 2
}

func vectorKeyframesFromProfile(keyframes []profile.Keyframe) ([]aep.VectorKeyframe, bool) {
	out := make([]aep.VectorKeyframe, 0, len(keyframes))
	width := 0
	for _, keyframe := range keyframes {
		value, ok := vectorValue(keyframe.Value)
		if !ok {
			return nil, false
		}
		if width == 0 {
			width = len(value)
		}
		if len(value) != width {
			return nil, false
		}
		out = append(out, aep.VectorKeyframe{Time: keyframe.Time, Value: value})
	}
	return out, len(out) >= 2 && width >= 2 && width <= 4
}

func vectorValue(value any) ([]float64, bool) {
	switch v := value.(type) {
	case []float64:
		if len(v) < 2 || len(v) > 4 {
			return nil, false
		}
		return append([]float64(nil), v...), true
	case []any:
		if len(v) < 2 || len(v) > 4 {
			return nil, false
		}
		out := make([]float64, 0, len(v))
		for _, item := range v {
			number, ok := scalarValue(item)
			if !ok {
				return nil, false
			}
			out = append(out, number)
		}
		return out, true
	default:
		return nil, false
	}
}

func scalarValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func verifyAppliedPseudoBehavior(prof *profile.Profile, effectMatchName string, rows []PseudoBehaviorApplicationControl) PseudoBehaviorApplicationVerify {
	verify := PseudoBehaviorApplicationVerify{Parsed: prof != nil}
	for _, row := range rows {
		if row.Status == "applied_keyframes" && verifiedKeyframesForControl(prof, effectMatchName, row.ParamMatchName) > 0 {
			verify.KeyframedControls++
		}
	}
	return verify
}

func verifiedKeyframesForControl(prof *profile.Profile, effectMatchName, paramMatchName string) int {
	if prof == nil {
		return 0
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			for _, effect := range layer.Effects {
				if effect.MatchName != effectMatchName {
					continue
				}
				for _, param := range effect.Params {
					if param.MatchName == paramMatchName {
						return len(param.Keyframes)
					}
				}
			}
		}
	}
	return 0
}

func proofFamilyByMatch(proof PseudoControllerRebuildProof, matchName string) *PseudoControllerRebuildFamily {
	for i := range proof.Families {
		if proof.Families[i].MatchName == matchName {
			return &proof.Families[i]
		}
	}
	return nil
}

func payloadFamilyByMatch(payload PseudoBehaviorPayloadReport, matchName string) *PseudoBehaviorPayloadFamily {
	for i := range payload.Families {
		if payload.Families[i].MatchName == matchName {
			return &payload.Families[i]
		}
	}
	return nil
}
