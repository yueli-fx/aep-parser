package selfhost

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestBuildPseudoBehaviorPayloadReportExtractsKeyframesExpressionsAndMisses(t *testing.T) {
	exprEnabled := true
	prof := &profile.Profile{
		Meta: profile.Meta{Path: "samples/rig.aep"},
		Comps: []profile.Composition{{
			Name: "Main",
			Layers: []profile.Layer{{
				Name: "Controller",
				Effects: []profile.Effect{{
					MatchName: "Pseudo/Rig",
					Params: []profile.Property{
						{
							Name:      "Animated",
							MatchName: "Pseudo/Rig-0002",
							Keyframes: []profile.Keyframe{
								{Time: 0, Value: 1.0, InInterp: "linear", OutInterp: "linear"},
								{Time: 1.5, Value: 12.0, InInterp: "bezier", OutInterp: "bezier"},
							},
						},
						{
							Name:              "Driven",
							MatchName:         "Pseudo/Rig-0003",
							Expression:        "time * 2",
							ExpressionEnabled: &exprEnabled,
						},
					},
				}},
			}},
		}},
	}
	plan := PseudoBehaviorWiringPlan{
		SchemaVersion: 1,
		Families: []PseudoBehaviorWiringFamily{{
			MatchName: "Pseudo/Rig",
			Status:    "generated",
			Tasks: []PseudoBehaviorWiringTask{
				{ParamMatchName: "Pseudo/Rig-0001", Action: "static_control_preserved", Phase: "none"},
				{ParamMatchName: "Pseudo/Rig-0002", Action: "extract_keyframes_then_apply", Phase: "keyframes"},
				{ParamMatchName: "Pseudo/Rig-0003", Action: "extract_expression_then_apply", Phase: "expression"},
				{ParamMatchName: "Pseudo/Rig-0004", Action: "extract_keyframes_then_apply", Phase: "keyframes"},
			},
		}},
	}

	report := BuildPseudoBehaviorPayloadReport(plan, []*profile.Profile{prof}, PseudoBehaviorPayloadOptions{PlanPath: "plan.json", Root: "samples", MaxExamples: 2})

	if report.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", report.SchemaVersion)
	}
	if report.Summary.BehaviorTasks != 3 || report.Summary.ControlsWithPayloads != 2 || report.Summary.ControlsMissingPayloads != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if report.Summary.KeyframePayloads != 1 || report.Summary.ExpressionPayloads != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	assertPayloadControl(t, report, "Pseudo/Rig-0002", "payload_found", 1, 2, "")
	assertPayloadControl(t, report, "Pseudo/Rig-0003", "payload_found", 1, 0, "time * 2")
	assertPayloadControl(t, report, "Pseudo/Rig-0004", "missing_payload", 0, 0, "")
}

func TestRunPseudoBehaviorPayloadExtractionRejectsMalformedPlan(t *testing.T) {
	dir := t.TempDir()
	planPath := filepath.Join(dir, "plan.json")
	writeFile(t, planPath, `{"schema_version":`)

	_, err := RunPseudoBehaviorPayloadExtraction(PseudoBehaviorPayloadOptions{
		PlanPath: planPath,
		Root:     filepath.Join(dir, "samples"),
		OutDir:   filepath.Join(dir, "payloads"),
	})

	if err == nil {
		t.Fatalf("err = nil, want malformed plan error")
	}
}

func TestRunPseudoBehaviorPayloadExtractionWritesJSONWithMissingPayload(t *testing.T) {
	dir := t.TempDir()
	planPath := filepath.Join(dir, "plan.json")
	outDir := filepath.Join(dir, "payloads")
	sampleRoot := filepath.Join(dir, "samples")
	if err := os.MkdirAll(sampleRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writeJSONFile(t, planPath, PseudoBehaviorWiringPlan{
		SchemaVersion: 1,
		Families: []PseudoBehaviorWiringFamily{{
			MatchName: "Pseudo/Rig",
			Tasks: []PseudoBehaviorWiringTask{{
				ParamMatchName: "Pseudo/Rig-0002",
				Action:         "extract_keyframes_then_apply",
				Phase:          "keyframes",
			}},
		}},
	})

	report, err := RunPseudoBehaviorPayloadExtraction(PseudoBehaviorPayloadOptions{
		PlanPath: planPath,
		Root:     sampleRoot,
		OutDir:   outDir,
	})
	if err != nil {
		t.Fatalf("RunPseudoBehaviorPayloadExtraction: %v", err)
	}

	if report.OutputPath != filepath.Join(outDir, "payloads.json") {
		t.Fatalf("output_path = %q", report.OutputPath)
	}
	if _, err := os.Stat(report.OutputPath); err != nil {
		t.Fatalf("payloads.json missing: %v", err)
	}
	if report.Summary.BehaviorTasks != 1 || report.Summary.ControlsMissingPayloads != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}

func assertPayloadControl(t *testing.T, report PseudoBehaviorPayloadReport, paramMatchName, status string, examples, keyframes int, expression string) {
	t.Helper()
	for _, family := range report.Families {
		for _, control := range family.Controls {
			if control.ParamMatchName != paramMatchName {
				continue
			}
			if control.Status != status {
				t.Fatalf("%s status = %q, want %q", paramMatchName, control.Status, status)
			}
			if len(control.Examples) != examples {
				t.Fatalf("%s examples = %d, want %d", paramMatchName, len(control.Examples), examples)
			}
			if examples == 0 {
				return
			}
			if len(control.Examples[0].Keyframes) != keyframes {
				t.Fatalf("%s keyframes = %+v", paramMatchName, control.Examples[0].Keyframes)
			}
			if control.Examples[0].Expression != expression {
				t.Fatalf("%s expression = %q, want %q", paramMatchName, control.Examples[0].Expression, expression)
			}
			return
		}
	}
	t.Fatalf("control %q missing from %+v", paramMatchName, report.Families)
}
