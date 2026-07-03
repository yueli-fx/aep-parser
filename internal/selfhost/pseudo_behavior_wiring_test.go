package selfhost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPseudoBehaviorWiringPlanDerivesStaticKeyframeAndExpressionTasks(t *testing.T) {
	proof := pseudoBehaviorProofFixture()

	plan := BuildPseudoBehaviorWiringPlan(proof, PseudoBehaviorWiringOptions{})

	if plan.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", plan.SchemaVersion)
	}
	if plan.Summary.Families != 1 || plan.Summary.Controls != 4 {
		t.Fatalf("summary = %+v", plan.Summary)
	}
	if plan.Summary.StaticControls != 1 || plan.Summary.KeyframeTasks != 2 || plan.Summary.ExpressionTasks != 2 || plan.Summary.ConflictTasks != 1 {
		t.Fatalf("summary = %+v", plan.Summary)
	}
	family := plan.Families[0]
	assertBehaviorTask(t, family, "Pseudo/Rig-0001", "static_control_preserved", "none")
	assertBehaviorTask(t, family, "Pseudo/Rig-0002", "extract_keyframes_then_apply", "keyframes")
	assertBehaviorTask(t, family, "Pseudo/Rig-0003", "extract_expression_then_apply", "expression")
	assertBehaviorTask(t, family, "Pseudo/Rig-0004", "resolve_keyframe_expression_conflict", "keyframes_then_expression")
}

func TestRunPseudoBehaviorWiringPlanWritesJSON(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	outDir := filepath.Join(dir, "behavior")
	writeJSONFile(t, proofPath, pseudoBehaviorProofFixture())

	plan, err := RunPseudoBehaviorWiringPlan(PseudoBehaviorWiringOptions{
		ProofPath: proofPath,
		OutDir:    outDir,
	})
	if err != nil {
		t.Fatalf("RunPseudoBehaviorWiringPlan: %v", err)
	}

	if plan.OutputPath != filepath.Join(outDir, "plan.json") {
		t.Fatalf("output_path = %q", plan.OutputPath)
	}
	if _, err := os.Stat(plan.OutputPath); err != nil {
		t.Fatalf("plan.json missing: %v", err)
	}
	if plan.Summary.KeyframeTasks != 2 || plan.Summary.ExpressionTasks != 2 {
		t.Fatalf("summary = %+v", plan.Summary)
	}
}

func TestRunPseudoBehaviorWiringPlanRejectsMalformedInput(t *testing.T) {
	dir := t.TempDir()
	proofPath := filepath.Join(dir, "proof.json")
	writeFile(t, proofPath, `{"schema_version":`)

	_, err := RunPseudoBehaviorWiringPlan(PseudoBehaviorWiringOptions{
		ProofPath: proofPath,
		OutDir:    filepath.Join(dir, "behavior"),
	})

	if err == nil {
		t.Fatalf("err = nil, want malformed proof error")
	}
}

func assertBehaviorTask(t *testing.T, family PseudoBehaviorWiringFamily, paramMatchName, action, phase string) {
	t.Helper()
	for _, task := range family.Tasks {
		if task.ParamMatchName != paramMatchName {
			continue
		}
		if task.Action != action || task.Phase != phase {
			t.Fatalf("%s task = %+v, want action %q phase %q", paramMatchName, task, action, phase)
		}
		return
	}
	t.Fatalf("task %q missing from %+v", paramMatchName, family.Tasks)
}

func pseudoBehaviorProofFixture() PseudoControllerRebuildProof {
	return PseudoControllerRebuildProof{
		SchemaVersion: 1,
		OutputPath:    "tmp/pseudo_controller_rebuild/proof.json",
		Summary: PseudoControllerRebuildSummary{
			PseudoFamilies:    1,
			SelectedFamilies:  1,
			GeneratedFamilies: 1,
			GeneratedControls: 4,
		},
		Families: []PseudoControllerRebuildFamily{{
			MatchName:    "Pseudo/Rig",
			UID:          "Rig",
			Status:       "generated",
			GeneratedAEP: "tmp/pseudo_controller_rebuild/generated/rig.aep",
			Controls: []PseudoControllerControlPlan{
				{ParamMatchName: "Pseudo/Rig-0001", Label: "Static", ControlKind: "slider", ValueType: "number"},
				{ParamMatchName: "Pseudo/Rig-0002", Label: "Animated", ControlKind: "slider", ValueType: "number", BehaviorNotes: []string{"keyframed"}},
				{ParamMatchName: "Pseudo/Rig-0003", Label: "Driven", ControlKind: "slider", ValueType: "number", BehaviorNotes: []string{"expression"}},
				{ParamMatchName: "Pseudo/Rig-0004", Label: "Both", ControlKind: "slider", ValueType: "number", BehaviorNotes: []string{"keyframed", "expression"}},
			},
		}},
	}
}
