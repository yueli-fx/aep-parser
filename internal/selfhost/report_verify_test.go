package selfhost

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyTechniqueReportAcceptsCompleteReport(t *testing.T) {
	dir := writeMinimalTechniqueReport(t)

	result, err := VerifyTechniqueReport(ReportVerifyOptions{OutDir: dir, MinProjects: 1})

	if err != nil {
		t.Fatalf("VerifyTechniqueReport: %v", err)
	}
	if result.ProjectCount != 1 || result.PatternCount != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestVerifyTechniqueReportRejectsMissingArtifact(t *testing.T) {
	dir := writeMinimalTechniqueReport(t)
	mustRemove(t, filepath.Join(dir, "recipe_drafts.jsonl"))

	_, err := VerifyTechniqueReport(ReportVerifyOptions{OutDir: dir, MinProjects: 1})

	if err == nil || !strings.Contains(err.Error(), "missing report artifact") {
		t.Fatalf("err = %v, want missing artifact", err)
	}
}

func writeMinimalTechniqueReport(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeJSONFile(t, filepath.Join(dir, "summary.json"), map[string]any{
		"project_count": 1,
		"error_count":   0,
		"totals": map[string]any{
			"comp_count":            1,
			"layer_count":           1,
			"effect_count":          1,
			"shape_operator_count":  1,
			"text_animator_count":   0,
			"dependency_count":      1,
			"text_layer_count":      0,
			"shape_layer_count":     1,
			"layer_role_counts":     map[string]any{"shape": 1},
			"unused_compatibility":  0,
			"unknown_compatibility": 0,
		},
	})
	writeJSONFile(t, filepath.Join(dir, "digest.json"), map[string]any{
		"project_count": 1,
		"patterns": []any{
			map[string]any{"id": "shape_system", "recreation_steps": []any{map[string]any{"id": "structure"}}},
		},
	})
	writeJSONFile(t, filepath.Join(dir, "manifest.json"), map[string]any{
		"project_count": 1,
		"error_count":   0,
		"git":           map[string]any{"commit": "abc123"},
		"artifacts": []any{
			"summary.json", "corpus.jsonl", "digest.json", "learning.md", "projects.csv", "report.md", "report.html", "recipe_drafts.jsonl",
		},
	})
	writeFile(t, filepath.Join(dir, "corpus.jsonl"), `{"path":"demo.aep","explanation":{"recreation_steps":[{"id":"structure"}]}}`+"\n")
	writeCSV(t, filepath.Join(dir, "projects.csv"), []string{"project_path", "readiness", "readiness_blockers"}, [][]string{{"demo.aep", "analysis_ready", ""}})
	writeCSV(t, filepath.Join(dir, "project_playbooks.csv"), []string{"project_path", "readiness", "overview", "ordered_steps"}, [][]string{{"demo.aep", "analysis_ready", "overview", "structure"}})
	writeCSV(t, filepath.Join(dir, "compositions.csv"), []string{"project_path", "name", "width", "height", "layer_count"}, [][]string{{"demo.aep", "Main", "1920", "1080", "1"}})
	writeCSV(t, filepath.Join(dir, "layers.csv"), []string{"project_path", "comp_name", "type", "role", "index"}, [][]string{{"demo.aep", "Main", "shape", "shape", "0"}})
	writeCSV(t, filepath.Join(dir, "recreation_steps.csv"), []string{"project_path", "step_id", "priority", "title", "summary"}, [][]string{{"demo.aep", "structure", "10", "Build", "Build structure"}})
	writeCSV(t, filepath.Join(dir, "patterns.csv"), []string{"id"}, [][]string{{"shape_system"}})
	writeCSV(t, filepath.Join(dir, "study_queue.csv"), []string{"path"}, [][]string{{"demo.aep"}})
	writeCSV(t, filepath.Join(dir, "study_tasks.csv"), []string{"project_path", "focus", "action", "rank"}, [][]string{{"demo.aep", "shape", "study", "1"}})
	writeCSV(t, filepath.Join(dir, "recreation_blockers.csv"), []string{"project_path", "readiness", "blocker_type", "blocker", "action"}, nil)
	writeCSV(t, filepath.Join(dir, "signal_layers.csv"), []string{"project_path", "layer_name", "role", "score", "signals"}, [][]string{{"demo.aep", "Shape", "shape", "1", "shape_operator:path"}})
	writeCSV(t, filepath.Join(dir, "effect_stacks.csv"), []string{"project_path", "comp_name", "match_name"}, [][]string{{"demo.aep", "Main", "ADBE Fill"}})
	writeCSV(t, filepath.Join(dir, "shape_operators.csv"), []string{"project_path", "comp_name", "layer_name", "family", "source"}, [][]string{{"demo.aep", "Main", "Shape", "path", "shape"}})
	writeCSV(t, filepath.Join(dir, "text_animators.csv"), []string{"project_path", "comp_name", "layer_name", "property_kind", "match_name"}, nil)
	writeCSV(t, filepath.Join(dir, "dependency_edges.csv"), []string{"project_path", "comp_name", "relation"}, [][]string{{"demo.aep", "Main", "source"}})
	writeCSV(t, filepath.Join(dir, "learning_actions.csv"), []string{"pattern", "action", "representative_project"}, [][]string{{"shape_system", "study", "demo.aep"}})
	writeCSV(t, filepath.Join(dir, "mechanisms.csv"), []string{"category", "name", "count", "action"}, [][]string{{"shape", "path", "1", "study"}})
	writeCSV(t, filepath.Join(dir, "mechanism_examples.csv"), []string{"category", "name", "project_path", "project_count", "action"}, [][]string{{"shape", "path", "demo.aep", "1", "study"}})
	coverageRows := [][]string{}
	for _, artifact := range []string{"projects.csv", "project_playbooks.csv", "compositions.csv", "layers.csv", "recreation_steps.csv", "effect_stacks.csv", "shape_operators.csv", "text_animators.csv", "dependency_edges.csv", "reconstruction_blueprints.jsonl", "recipe_drafts.jsonl"} {
		coverageRows = append(coverageRows, []string{artifact, "rows", "1", "1", "ok"})
	}
	writeCSV(t, filepath.Join(dir, "coverage_scorecard.csv"), []string{"artifact", "metric", "expected_count", "actual_count", "status"}, coverageRows)
	writeFile(t, filepath.Join(dir, "reconstruction_blueprints.jsonl"), `{"project_path":"demo.aep","readiness":"analysis_ready","counts":{},"phases":[{"id":"create_compositions"},{"id":"create_layers"},{"id":"apply_mechanisms"},{"id":"wire_dependencies"},{"id":"verify_recreation"}]}`+"\n")
	writeFile(t, filepath.Join(dir, "recipe_drafts.jsonl"), `{"project_path":"demo.aep","readiness":"analysis_ready","counts":{},"recipe":{"schema_version":1,"project":{},"comps":[{"name":"Main"}],"expected_profile":{"comp_count":1}},"gaps":[{"id":"layers"}]}`+"\n")
	writeCSV(t, filepath.Join(dir, "errors.csv"), []string{"path", "mode", "error"}, nil)
	writeFile(t, filepath.Join(dir, "learning.md"), "# Technique Learning Index\n\n## Pattern Playbook\n\n## Study Queue\n\nrecreation steps\n\n## Plugin Risk Queue\n\n## Readiness Queue\n")
	writeFile(t, filepath.Join(dir, "report.md"), "# Report\n\n## Pattern Representatives\n\n## Study Queue\n\nStep 1: Build\n")
	writeFile(t, filepath.Join(dir, "report.html"), strings.Join([]string{
		"Technique Corpus Dashboard",
		`data-tab="overview"`,
		`data-tab="patterns"`,
		`data-tab="projects"`,
		`data-tab="recipes"`,
		`data-tab="data"`,
		`id="project-search"`,
		`id="project-detail"`,
		`window.__TECHNIQUE_REPORT__`,
		"summary.json", "corpus.jsonl", "digest.json", "projects.csv", "project_playbooks.csv",
		"compositions.csv", "layers.csv", "recreation_steps.csv", "study_queue.csv",
		"study_tasks.csv", "recreation_blockers.csv", "signal_layers.csv", "effect_stacks.csv",
		"shape_operators.csv", "text_animators.csv", "dependency_edges.csv", "learning_actions.csv",
		"mechanisms.csv", "mechanism_examples.csv", "coverage_scorecard.csv",
		"reconstruction_blueprints.jsonl", "recipe_drafts.jsonl", "errors.csv",
	}, "\n"))
	return dir
}

func writeCSV(t *testing.T, path string, header []string, rows [][]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write(header); err != nil {
		t.Fatalf("Write header: %v", err)
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			t.Fatalf("Write row: %v", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatalf("csv flush: %v", err)
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func mustRemove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}
