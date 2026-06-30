package selfhost

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinalizeSelfhostRunWritesLatestOutcomeAndHistory(t *testing.T) {
	root := t.TempDir()
	runRoot := filepath.Join(root, "20260701T000000Z")
	writeFinalizeRunFixture(t, runRoot)

	result, err := FinalizeSelfhostRun(context.Background(), FinalizeOptions{
		OutRoot:   root,
		RunRoot:   runRoot,
		RunID:     "20260701T000000Z",
		InputPath: "data/samples",
	})

	if err != nil {
		t.Fatalf("FinalizeSelfhostRun: %v", err)
	}
	if result.OutcomeStatus != "pass" {
		t.Fatalf("outcome status = %q", result.OutcomeStatus)
	}
	for _, path := range []string{
		filepath.Join(runRoot, "acceptance.json"),
		filepath.Join(runRoot, "effectiveness.json"),
		filepath.Join(root, "latest_outcome.json"),
		filepath.Join(root, "latest_outcome.md"),
		filepath.Join(root, "latest_outcome.html"),
		filepath.Join(root, "latest_effectiveness.json"),
		filepath.Join(root, "latest_index.html"),
		filepath.Join(root, "history.jsonl"),
		filepath.Join(root, "history.csv"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected output %s: %v", path, err)
		}
	}
	index, err := os.ReadFile(filepath.Join(root, "latest_index.html"))
	if err != nil {
		t.Fatalf("ReadFile latest_index: %v", err)
	}
	if !strings.Contains(string(index), "latest_outcome.json") || !strings.Contains(string(index), "history.jsonl") {
		t.Fatalf("latest_index missing stable links:\n%s", string(index))
	}
	for _, want := range []string{
		"Full Report",
		"Project Playbooks Preview",
		"Recipe Draft Compile Smoke",
		"Recipe Draft Batch Smoke",
		"Compiled Draft Reparse Smoke",
		"latest_effectiveness.json",
	} {
		if !strings.Contains(string(index), want) {
			t.Fatalf("latest_index missing %q:\n%s", want, string(index))
		}
	}
	outcomeMD, err := os.ReadFile(filepath.Join(root, "latest_outcome.md"))
	if err != nil {
		t.Fatalf("ReadFile latest_outcome.md: %v", err)
	}
	for _, want := range []string{"Outcome Status", "Effectiveness Headline", "Next Actions", "Reconstruction Readiness", "Top Plugin Blockers"} {
		if !strings.Contains(string(outcomeMD), want) {
			t.Fatalf("latest_outcome.md missing %q:\n%s", want, string(outcomeMD))
		}
	}
	outcomeHTML, err := os.ReadFile(filepath.Join(root, "latest_outcome.html"))
	if err != nil {
		t.Fatalf("ReadFile latest_outcome.html: %v", err)
	}
	for _, want := range []string{"Self-Hosted Outcome", "Learning Signals", "Study Queue", "Learning Actions", "Recent Runs"} {
		if !strings.Contains(string(outcomeHTML), want) {
			t.Fatalf("latest_outcome.html missing %q:\n%s", want, string(outcomeHTML))
		}
	}
	var effectiveness struct {
		ReconstructionStatus struct {
			ReadinessSummary  []any `json:"readiness_summary"`
			BlockerSummary    []any `json:"blocker_summary"`
			PluginBlockersTop []any `json:"plugin_blockers_top"`
		} `json:"reconstruction_status"`
		ActionPlan struct {
			NextActions []any `json:"next_actions"`
		} `json:"action_plan"`
	}
	readIndentedJSON(filepath.Join(root, "latest_effectiveness.json"), &effectiveness)
	if effectiveness.ReconstructionStatus.ReadinessSummary == nil || effectiveness.ReconstructionStatus.BlockerSummary == nil || effectiveness.ReconstructionStatus.PluginBlockersTop == nil {
		t.Fatalf("latest_effectiveness.json missing reconstruction status arrays")
	}
	if effectiveness.ActionPlan.NextActions == nil {
		t.Fatalf("latest_effectiveness.json missing next actions")
	}
}

func writeFinalizeRunFixture(t *testing.T, runRoot string) {
	t.Helper()
	writeJSONFile(t, filepath.Join(runRoot, "full_report", "manifest.json"), map[string]any{
		"project_count": 1,
		"error_count":   0,
		"pattern_count": 2,
	})
	writeJSONFile(t, filepath.Join(runRoot, "partial_report", "manifest.json"), map[string]any{
		"project_count": 1,
		"error_count":   1,
		"pattern_count": 1,
	})
	writeJSONFile(t, filepath.Join(runRoot, "compare_partial_to_full", "compare.json"), map[string]any{
		"count_diffs": []any{map[string]any{"group": "patterns"}},
	})
	writeJSONFile(t, filepath.Join(runRoot, "recipe_draft_compile", "compile.json"), map[string]any{"valid": true})
	writeFile(t, filepath.Join(runRoot, "recipe_draft_compile", "recipe_draft.aep"), "fake aep")
	writeJSONFile(t, filepath.Join(runRoot, "recipe_draft_reparse", "reparse_summary.json"), map[string]any{
		"passed":               true,
		"expected_comp_count":  1,
		"actual_comp_count":    1,
		"expected_layer_count": 0,
		"actual_layer_count":   0,
	})
	writeJSONFile(t, filepath.Join(runRoot, "recipe_draft_batch", "summary.json"), map[string]any{
		"requested": 1,
		"attempted": 1,
		"passed":    1,
	})
	writeCSV(t, filepath.Join(runRoot, "full_report", "study_queue.csv"), []string{"rank", "path", "readiness", "study_score"}, [][]string{{"1", "demo.aep", "analysis_ready", "10"}})
	writeCSV(t, filepath.Join(runRoot, "full_report", "learning_actions.csv"), []string{"priority", "pattern", "count", "action", "representative_project"}, [][]string{{"100", "shape", "1", "study", "demo.aep"}})
	writeCSV(t, filepath.Join(runRoot, "full_report", "coverage_scorecard.csv"), []string{"artifact", "metric", "expected_count", "actual_count", "status", "notes"}, [][]string{{"projects.csv", "rows", "1", "1", "ok", ""}})
	writeCSV(t, filepath.Join(runRoot, "full_report", "recreation_blockers.csv"), []string{"project_path", "readiness", "blocker_type", "blocker", "action"}, nil)
}
