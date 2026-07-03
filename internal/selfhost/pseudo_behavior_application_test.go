package selfhost

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestRunPseudoBehaviorApplicationAppliesGeneratedKeyframes(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	payloadPath := filepath.Join(dir, "payloads.json")
	outDir := filepath.Join(dir, "application")
	writeJSONFile(t, proofPath, pseudoBehaviorApplicationProofFixture("generated"))
	writeJSONFile(t, payloadPath, pseudoBehaviorApplicationPayloadFixture())

	report, err := RunPseudoBehaviorApplication(PseudoBehaviorApplicationOptions{
		ProofPath:   proofPath,
		PayloadPath: payloadPath,
		OutDir:      outDir,
	})
	if err != nil {
		t.Fatalf("RunPseudoBehaviorApplication: %v", err)
	}

	if report.OutputPath != filepath.Join(outDir, "application.json") {
		t.Fatalf("output_path = %q", report.OutputPath)
	}
	if _, err := os.Stat(report.OutputPath); err != nil {
		t.Fatalf("application.json missing: %v", err)
	}
	if report.Summary.AppliedControls != 1 || report.Summary.VerifiedKeyframedControls != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	family := report.Families[0]
	if family.Status != "applied" || family.GeneratedAEP == "" {
		t.Fatalf("family = %+v", family)
	}
	if _, err := os.Stat(family.GeneratedAEP); err != nil {
		t.Fatalf("generated AEP missing: %v", err)
	}
	if family.Controls[0].Status != "applied_keyframes" || family.Controls[0].AppliedKeyframes != 2 {
		t.Fatalf("control = %+v", family.Controls[0])
	}
}

func TestRunPseudoBehaviorApplicationSkipsNonGeneratedFamily(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	payloadPath := filepath.Join(dir, "payloads.json")
	writeJSONFile(t, proofPath, pseudoBehaviorApplicationProofFixture("planned"))
	writeJSONFile(t, payloadPath, pseudoBehaviorApplicationPayloadFixture())

	report, err := RunPseudoBehaviorApplication(PseudoBehaviorApplicationOptions{
		ProofPath:   proofPath,
		PayloadPath: payloadPath,
		OutDir:      filepath.Join(dir, "application"),
	})
	if err != nil {
		t.Fatalf("RunPseudoBehaviorApplication: %v", err)
	}

	if report.Summary.SkippedFamilies != 1 || report.Summary.AppliedControls != 0 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if report.Families[0].Status != "skipped_not_generated" {
		t.Fatalf("family status = %q", report.Families[0].Status)
	}
}

func TestRunPseudoBehaviorApplicationAppliesMultipleGeneratedFamilies(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	payloadPath := filepath.Join(dir, "payloads.json")
	outDir := filepath.Join(dir, "application")
	writeJSONFile(t, proofPath, PseudoControllerRebuildProof{
		SchemaVersion: 1,
		Families: []PseudoControllerRebuildFamily{
			pseudoBehaviorApplicationFamilyFixture("Pseudo/RigA", "RigA"),
			pseudoBehaviorApplicationFamilyFixture("Pseudo/RigB", "RigB"),
		},
	})
	writeJSONFile(t, payloadPath, PseudoBehaviorPayloadReport{
		SchemaVersion: 1,
		Families: []PseudoBehaviorPayloadFamily{
			pseudoBehaviorApplicationPayloadFamilyFixture("Pseudo/RigA"),
			pseudoBehaviorApplicationPayloadFamilyFixture("Pseudo/RigB"),
		},
	})

	report, err := RunPseudoBehaviorApplication(PseudoBehaviorApplicationOptions{
		ProofPath:   proofPath,
		PayloadPath: payloadPath,
		OutDir:      outDir,
	})
	if err != nil {
		t.Fatalf("RunPseudoBehaviorApplication: %v", err)
	}

	if report.Summary.GeneratedFamilies != 2 || report.Summary.GeneratedAEPs != 2 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if report.Summary.AppliedControls != 2 || report.Summary.VerifiedKeyframedControls != 2 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	for _, family := range report.Families {
		if family.Status != "applied" || family.GeneratedAEP == "" {
			t.Fatalf("family = %+v", family)
		}
	}
}

func TestRunPseudoBehaviorApplicationAppliesVectorKeyframes(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	payloadPath := filepath.Join(dir, "payloads.json")
	outDir := filepath.Join(dir, "application")
	writeJSONFile(t, proofPath, PseudoControllerRebuildProof{
		SchemaVersion: 1,
		Families: []PseudoControllerRebuildFamily{{
			MatchName: "Pseudo/PointRig",
			UID:       "PointRig",
			Status:    "generated",
			Controls: []PseudoControllerControlPlan{
				{ParamMatchName: "Pseudo/PointRig-0001", Label: "Point", ControlKind: "point", ValueType: "vector2", DefaultValue: []any{0.1, 0.2}, BehaviorNotes: []string{"keyframed"}},
			},
		}},
	})
	writeJSONFile(t, payloadPath, PseudoBehaviorPayloadReport{
		SchemaVersion: 1,
		Families: []PseudoBehaviorPayloadFamily{{
			MatchName: "Pseudo/PointRig",
			Status:    "generated",
			Controls: []PseudoBehaviorPayloadControl{{
				ParamMatchName: "Pseudo/PointRig-0001",
				Action:         "extract_keyframes_then_apply",
				Phase:          "keyframes",
				Status:         "payload_found",
				Examples: []PseudoBehaviorPayloadExample{{
					EffectMatchName: "Pseudo/PointRig",
					ParamMatchName:  "Pseudo/PointRig-0001",
					Keyframes: []profile.Keyframe{
						{Time: 0, Value: []float64{0.1, 0.2}},
						{Time: 1, Value: []float64{0.8, 0.6}},
					},
				}},
			}},
		}},
	})

	report, err := RunPseudoBehaviorApplication(PseudoBehaviorApplicationOptions{
		ProofPath:   proofPath,
		PayloadPath: payloadPath,
		OutDir:      outDir,
	})
	if err != nil {
		t.Fatalf("RunPseudoBehaviorApplication: %v", err)
	}

	if report.Summary.AppliedControls != 1 || report.Summary.VerifiedKeyframedControls != 1 || report.Summary.ErrorControls != 0 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	control := report.Families[0].Controls[0]
	if control.Status != "applied_keyframes" || control.VerifiedKeyframes != 2 {
		t.Fatalf("control = %+v", control)
	}
}

func TestRunPseudoBehaviorApplicationDefersExpressionPayloads(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	payloadPath := filepath.Join(dir, "payloads.json")
	outDir := filepath.Join(dir, "application")
	exprEnabled := true
	writeJSONFile(t, proofPath, PseudoControllerRebuildProof{
		SchemaVersion: 1,
		Families: []PseudoControllerRebuildFamily{{
			MatchName: "Pseudo/ExprRig",
			UID:       "ExprRig",
			Status:    "generated",
			Controls: []PseudoControllerControlPlan{
				{ParamMatchName: "Pseudo/ExprRig-0001", Label: "Driven", ControlKind: "slider", ValueType: "number", DefaultValue: 0.0, BehaviorNotes: []string{"expression"}},
				{ParamMatchName: "Pseudo/ExprRig-0002", Label: "Both", ControlKind: "slider", ValueType: "number", DefaultValue: 0.0, BehaviorNotes: []string{"keyframed", "expression"}},
			},
		}},
	})
	writeJSONFile(t, payloadPath, PseudoBehaviorPayloadReport{
		SchemaVersion: 1,
		Families: []PseudoBehaviorPayloadFamily{{
			MatchName: "Pseudo/ExprRig",
			Status:    "generated",
			Controls: []PseudoBehaviorPayloadControl{
				{
					ParamMatchName: "Pseudo/ExprRig-0001",
					Action:         "extract_expression_then_apply",
					Phase:          "expression",
					Status:         "payload_found",
					Examples: []PseudoBehaviorPayloadExample{{
						EffectMatchName:   "Pseudo/ExprRig",
						ParamMatchName:    "Pseudo/ExprRig-0001",
						Expression:        "time * 2",
						ExpressionEnabled: &exprEnabled,
					}},
				},
				{
					ParamMatchName: "Pseudo/ExprRig-0002",
					Action:         "resolve_keyframe_expression_conflict",
					Phase:          "keyframes_then_expression",
					Status:         "payload_found",
					Examples: []PseudoBehaviorPayloadExample{{
						EffectMatchName:   "Pseudo/ExprRig",
						ParamMatchName:    "Pseudo/ExprRig-0002",
						Keyframes:         []profile.Keyframe{{Time: 0, Value: 1.0}, {Time: 1, Value: 4.0}},
						Expression:        "value + time",
						ExpressionEnabled: &exprEnabled,
					}},
				},
			},
		}},
	})

	report, err := RunPseudoBehaviorApplication(PseudoBehaviorApplicationOptions{
		ProofPath:   proofPath,
		PayloadPath: payloadPath,
		OutDir:      outDir,
	})
	if err != nil {
		t.Fatalf("RunPseudoBehaviorApplication: %v", err)
	}

	if report.Summary.AppliedControls != 1 || report.Summary.VerifiedKeyframedControls != 1 || report.Summary.ExpressionDeferredControls != 2 || report.Summary.ErrorControls != 0 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	exprOnly := report.Families[0].Controls[0]
	if exprOnly.Status != "expression_deferred" || exprOnly.DeferredExpression != "time * 2" || !exprOnly.ExpressionDeferred {
		t.Fatalf("exprOnly = %+v", exprOnly)
	}
	conflict := report.Families[0].Controls[1]
	if conflict.Status != "applied_keyframes" || conflict.AppliedKeyframes != 2 || conflict.DeferredExpression != "value + time" || !conflict.ExpressionDeferred {
		t.Fatalf("conflict = %+v", conflict)
	}
}

func TestRunPseudoBehaviorApplicationRejectsMalformedPayload(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	payloadPath := filepath.Join(dir, "payloads.json")
	writeJSONFile(t, proofPath, pseudoBehaviorApplicationProofFixture("generated"))
	writeFile(t, payloadPath, `{"schema_version":`)

	_, err := RunPseudoBehaviorApplication(PseudoBehaviorApplicationOptions{
		ProofPath:   proofPath,
		PayloadPath: payloadPath,
		OutDir:      filepath.Join(dir, "application"),
	})

	if err == nil {
		t.Fatalf("err = nil, want malformed payload error")
	}
}

func pseudoBehaviorApplicationFamilyFixture(matchName, uid string) PseudoControllerRebuildFamily {
	return PseudoControllerRebuildFamily{
		MatchName: matchName,
		UID:       uid,
		Status:    "generated",
		Controls: []PseudoControllerControlPlan{
			{ParamMatchName: matchName + "-0001", Label: "Static", ControlKind: "slider", ValueType: "number", DefaultValue: 5.0},
			{ParamMatchName: matchName + "-0002", Label: "Animated", ControlKind: "slider", ValueType: "number", DefaultValue: 0.0, BehaviorNotes: []string{"keyframed"}},
		},
	}
}

func pseudoBehaviorApplicationPayloadFamilyFixture(matchName string) PseudoBehaviorPayloadFamily {
	return PseudoBehaviorPayloadFamily{
		MatchName: matchName,
		Status:    "generated",
		Controls: []PseudoBehaviorPayloadControl{{
			ParamMatchName: matchName + "-0002",
			Action:         "extract_keyframes_then_apply",
			Phase:          "keyframes",
			Status:         "payload_found",
			Examples: []PseudoBehaviorPayloadExample{{
				EffectMatchName: matchName,
				ParamMatchName:  matchName + "-0002",
				Keyframes: []profile.Keyframe{
					{Time: 0, Value: 0.0},
					{Time: 1, Value: 25.0},
				},
			}},
		}},
	}
}

func pseudoBehaviorApplicationProofFixture(status string) PseudoControllerRebuildProof {
	return PseudoControllerRebuildProof{
		SchemaVersion: 1,
		Summary: PseudoControllerRebuildSummary{
			SelectedFamilies:  1,
			GeneratedFamilies: 1,
			GeneratedControls: 2,
		},
		Families: []PseudoControllerRebuildFamily{{
			MatchName: "Pseudo/Rig",
			UID:       "Rig",
			Status:    status,
			Controls: []PseudoControllerControlPlan{
				{ParamMatchName: "Pseudo/Rig-0001", Label: "Static", ControlKind: "slider", ValueType: "number", DefaultValue: 5.0},
				{ParamMatchName: "Pseudo/Rig-0002", Label: "Animated", ControlKind: "slider", ValueType: "number", DefaultValue: 0.0, BehaviorNotes: []string{"keyframed"}},
			},
		}},
	}
}

func pseudoBehaviorApplicationPayloadFixture() PseudoBehaviorPayloadReport {
	return PseudoBehaviorPayloadReport{
		SchemaVersion: 1,
		Families: []PseudoBehaviorPayloadFamily{{
			MatchName: "Pseudo/Rig",
			Status:    "generated",
			Controls: []PseudoBehaviorPayloadControl{{
				ParamMatchName: "Pseudo/Rig-0002",
				Action:         "extract_keyframes_then_apply",
				Phase:          "keyframes",
				Status:         "payload_found",
				Examples: []PseudoBehaviorPayloadExample{{
					EffectMatchName: "Pseudo/Rig",
					ParamMatchName:  "Pseudo/Rig-0002",
					Keyframes: []profile.Keyframe{
						{Time: 0, Value: 0.0},
						{Time: 1, Value: 25.0},
					},
				}},
			}},
		}},
	}
}
