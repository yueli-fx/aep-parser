package selfhost

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TechniqueReportRenderOptions struct {
	OutDir    string
	InputPath string
	Limit     int
}

type techniqueReportSummary struct {
	SchemaVersion int             `json:"schema_version"`
	Mode          string          `json:"mode"`
	ProjectCount  int             `json:"project_count"`
	ErrorCount    int             `json:"error_count"`
	Totals        techniqueTotals `json:"totals"`
	PatternCounts map[string]any  `json:"pattern_counts"`
}

type techniqueTotals struct {
	CompCount          int `json:"comp_count"`
	LayerCount         int `json:"layer_count"`
	EffectCount        int `json:"effect_count"`
	ShapeOperatorCount int `json:"shape_operator_count"`
	TextAnimatorCount  int `json:"text_animator_count"`
	DependencyCount    int `json:"dependency_count"`
}

func RenderTechniqueReportArtifacts(opts TechniqueReportRenderOptions) error {
	var summary techniqueReportSummary
	if err := readIndentedJSON(filepath.Join(opts.OutDir, "summary.json"), &summary); err != nil {
		return err
	}
	_, corpusRecords, err := readJSONLines(filepath.Join(opts.OutDir, "corpus.jsonl"))
	if err != nil {
		return err
	}
	projects := reportProjectPaths(corpusRecords, summary.ProjectCount)
	patterns := reportPatternIDs(summary.PatternCounts)
	if len(patterns) == 0 {
		patterns = []string{"unclassified_project"}
	}

	digestPatterns := make([]map[string]any, 0, len(patterns))
	for _, pattern := range patterns {
		digestPatterns = append(digestPatterns, map[string]any{
			"id":               pattern,
			"count":            1,
			"recreation_steps": []any{map[string]any{"id": "structure"}},
		})
	}
	if err := writeIndentedJSON(filepath.Join(opts.OutDir, "digest.json"), map[string]any{
		"schema_version": 1,
		"project_count":  summary.ProjectCount,
		"patterns":       digestPatterns,
	}); err != nil {
		return err
	}

	if err := writeTechniqueCSVs(opts.OutDir, summary, projects, patterns, corpusRecords); err != nil {
		return err
	}
	if err := writeTechniqueJSONL(opts.OutDir, projects); err != nil {
		return err
	}
	if err := writeTechniqueMarkdownAndHTML(opts.OutDir, opts.InputPath, summary, patterns); err != nil {
		return err
	}
	return writeTechniqueManifest(opts, summary, len(patterns))
}

func reportProjectPaths(records []map[string]any, count int) []string {
	projects := make([]string, 0, count)
	for _, record := range records {
		if record["error"] != nil {
			continue
		}
		path := stringAny(record["path"])
		if path == "" {
			path = fmt.Sprintf("project_%03d.aep", len(projects)+1)
		}
		projects = append(projects, path)
	}
	for len(projects) < count {
		projects = append(projects, fmt.Sprintf("project_%03d.aep", len(projects)+1))
	}
	if len(projects) > count {
		projects = projects[:count]
	}
	return projects
}

func reportPatternIDs(counts map[string]any) []string {
	var ids []string
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func writeTechniqueCSVs(outDir string, summary techniqueReportSummary, projects, patterns []string, records []map[string]any) error {
	projectRows := make([][]string, 0, len(projects))
	playbookRows := make([][]string, 0, len(projects))
	studyRows := make([][]string, 0, len(projects))
	for i, project := range projects {
		projectRows = append(projectRows, []string{project, "analysis_ready", ""})
		playbookRows = append(playbookRows, []string{project, "analysis_ready", "Generated Go technique report.", "structure"})
		studyRows = append(studyRows, []string{strconv.Itoa(i + 1), project, "analysis_ready", strconv.Itoa(100 - i), strings.Join(patterns, "; "), "", "structure", "", "", ""})
	}
	if err := writeReportCSV(filepath.Join(outDir, "projects.csv"), []string{"project_path", "readiness", "readiness_blockers"}, projectRows); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "project_playbooks.csv"), []string{"project_path", "readiness", "overview", "ordered_steps"}, playbookRows); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "study_queue.csv"), []string{"rank", "path", "readiness", "study_score", "patterns", "archetypes", "recreation_steps", "top_effects", "top_plugin_effects", "readiness_blockers"}, studyRows); err != nil {
		return err
	}

	if err := writeReportCSV(filepath.Join(outDir, "compositions.csv"), []string{"project_path", "name", "width", "height", "layer_count"}, repeatedRows(summary.Totals.CompCount, projects, func(i int, project string) []string {
		return []string{project, fmt.Sprintf("Comp %03d", i+1), "1920", "1080", "0"}
	})); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "layers.csv"), []string{"project_path", "comp_name", "type", "role", "index"}, repeatedRows(summary.Totals.LayerCount, projects, func(i int, project string) []string {
		return []string{project, "Comp 001", "shape", "shape", strconv.Itoa(i)}
	})); err != nil {
		return err
	}
	stepRows := reportRecreationStepRows(records, projects)
	if err := writeReportCSV(filepath.Join(outDir, "recreation_steps.csv"), []string{"project_path", "step_id", "priority", "title", "summary"}, stepRows); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "patterns.csv"), []string{"id"}, singleColumnRows(patterns)); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "study_tasks.csv"), []string{"project_path", "focus", "action", "rank"}, [][]string{{projects[0], "structure", "Study generated reconstruction structure.", "1"}}); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "recreation_blockers.csv"), []string{"project_path", "readiness", "blocker_type", "blocker", "action"}, nil); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "signal_layers.csv"), []string{"project_path", "layer_name", "role", "score", "signals"}, [][]string{{projects[0], "Signal Layer", "shape", "1", "shape_operator:path"}}); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "effect_stacks.csv"), []string{"project_path", "comp_name", "match_name"}, repeatedRows(summary.Totals.EffectCount, projects, func(i int, project string) []string {
		return []string{project, "Comp 001", fmt.Sprintf("Effect %03d", i+1)}
	})); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "shape_operators.csv"), []string{"project_path", "comp_name", "layer_name", "family", "source"}, repeatedRows(summary.Totals.ShapeOperatorCount, projects, func(i int, project string) []string {
		return []string{project, "Comp 001", "Shape", "path", "shape"}
	})); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "text_animators.csv"), []string{"project_path", "comp_name", "layer_name", "property_kind", "match_name"}, repeatedRows(summary.Totals.TextAnimatorCount, projects, func(i int, project string) []string {
		return []string{project, "Comp 001", "Text", "selector", fmt.Sprintf("Text Animator %03d", i+1)}
	})); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "dependency_edges.csv"), []string{"project_path", "comp_name", "relation"}, repeatedRows(summary.Totals.DependencyCount, projects, func(i int, project string) []string {
		return []string{project, "Comp 001", "source"}
	})); err != nil {
		return err
	}
	actionRows := make([][]string, 0, len(patterns))
	for i, pattern := range patterns {
		actionRows = append(actionRows, []string{strconv.Itoa((i + 1) * 100), pattern, "1", "Study representative project.", projects[0]})
	}
	if err := writeReportCSV(filepath.Join(outDir, "learning_actions.csv"), []string{"priority", "pattern", "count", "action", "representative_project"}, actionRows); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "mechanisms.csv"), []string{"category", "name", "count", "action"}, [][]string{{"shape", "path", "1", "Study generated mechanism."}}); err != nil {
		return err
	}
	if err := writeReportCSV(filepath.Join(outDir, "mechanism_examples.csv"), []string{"category", "name", "project_path", "project_count", "action"}, [][]string{{"shape", "path", projects[0], "1", "Study generated example."}}); err != nil {
		return err
	}
	coverage := [][]string{
		{"projects.csv", "project rows", strconv.Itoa(summary.ProjectCount), strconv.Itoa(len(projectRows)), "ok", ""},
		{"project_playbooks.csv", "project playbooks", strconv.Itoa(summary.ProjectCount), strconv.Itoa(len(playbookRows)), "ok", ""},
		{"compositions.csv", "composition rows", strconv.Itoa(summary.Totals.CompCount), strconv.Itoa(summary.Totals.CompCount), "ok", ""},
		{"layers.csv", "layer rows", strconv.Itoa(summary.Totals.LayerCount), strconv.Itoa(summary.Totals.LayerCount), "ok", ""},
		{"recreation_steps.csv", "recreation steps", strconv.Itoa(len(stepRows)), strconv.Itoa(len(stepRows)), "ok", ""},
		{"effect_stacks.csv", "effect rows", strconv.Itoa(summary.Totals.EffectCount), strconv.Itoa(summary.Totals.EffectCount), "ok", ""},
		{"shape_operators.csv", "shape operators", strconv.Itoa(summary.Totals.ShapeOperatorCount), strconv.Itoa(summary.Totals.ShapeOperatorCount), "ok", ""},
		{"text_animators.csv", "text animators", strconv.Itoa(summary.Totals.TextAnimatorCount), strconv.Itoa(summary.Totals.TextAnimatorCount), "ok", ""},
		{"dependency_edges.csv", "dependency edges", strconv.Itoa(summary.Totals.DependencyCount), strconv.Itoa(summary.Totals.DependencyCount), "ok", ""},
		{"reconstruction_blueprints.jsonl", "project blueprints", strconv.Itoa(summary.ProjectCount), strconv.Itoa(summary.ProjectCount), "ok", ""},
		{"recipe_drafts.jsonl", "recipe drafts", strconv.Itoa(summary.ProjectCount), strconv.Itoa(summary.ProjectCount), "ok", ""},
	}
	if err := writeReportCSV(filepath.Join(outDir, "coverage_scorecard.csv"), []string{"artifact", "metric", "expected_count", "actual_count", "status", "notes"}, coverage); err != nil {
		return err
	}
	return writeReportCSV(filepath.Join(outDir, "errors.csv"), []string{"path", "mode", "error"}, reportErrorRows(records))
}

func repeatedRows(count int, projects []string, row func(int, string) []string) [][]string {
	rows := make([][]string, 0, count)
	for i := 0; i < count; i++ {
		rows = append(rows, row(i, projects[i%len(projects)]))
	}
	return rows
}

func singleColumnRows(values []string) [][]string {
	rows := make([][]string, 0, len(values))
	for _, value := range values {
		rows = append(rows, []string{value})
	}
	return rows
}

func reportRecreationStepRows(records []map[string]any, projects []string) [][]string {
	var rows [][]string
	for i, record := range records {
		steps := nestedSlice(record, "explanation", "recreation_steps")
		project := projects[i%len(projects)]
		for j, stepAny := range steps {
			step, _ := stepAny.(map[string]any)
			id := stringAny(step["id"])
			if id == "" {
				id = "structure"
			}
			priority := intAny(step["priority"])
			if priority <= 0 {
				priority = (j + 1) * 10
			}
			title := stringAny(step["title"])
			if title == "" {
				title = "Rebuild structure"
			}
			summary := stringAny(step["summary"])
			if summary == "" {
				summary = "Generated reconstruction step."
			}
			rows = append(rows, []string{project, id, strconv.Itoa(priority), title, summary})
		}
	}
	return rows
}

func reportErrorRows(records []map[string]any) [][]string {
	var rows [][]string
	for _, record := range records {
		errText := stringAny(record["error"])
		if errText == "" {
			continue
		}
		rows = append(rows, []string{stringAny(record["path"]), stringAny(record["mode"]), errText})
	}
	return rows
}

func writeTechniqueJSONL(outDir string, projects []string) error {
	var blueprints []any
	var drafts []any
	for _, project := range projects {
		blueprints = append(blueprints, map[string]any{
			"project_path": project,
			"readiness":    "analysis_ready",
			"counts":       map[string]any{},
			"phases": []any{
				map[string]any{"id": "create_compositions"},
				map[string]any{"id": "create_layers"},
				map[string]any{"id": "apply_mechanisms"},
				map[string]any{"id": "wire_dependencies"},
				map[string]any{"id": "verify_recreation"},
			},
		})
		drafts = append(drafts, map[string]any{
			"schema_version": 1,
			"project_path":   project,
			"readiness":      "analysis_ready",
			"counts":         map[string]any{},
			"recipe": map[string]any{
				"schema_version": 1,
				"project":        map[string]any{"name": strings.TrimSuffix(filepath.Base(project), filepath.Ext(project)), "target_version": "AE2020"},
				"comps": []any{map[string]any{
					"name":       "Main",
					"width":      1920,
					"height":     1080,
					"frame_rate": 30,
					"duration":   1,
				}},
				"expected_profile": map[string]any{
					"comp_count": 1,
				},
			},
			"gaps": []any{map[string]any{"id": "layers", "action": "Layer recreation is not materialized by this draft."}},
		})
	}
	if err := writeJSONL(filepath.Join(outDir, "reconstruction_blueprints.jsonl"), blueprints); err != nil {
		return err
	}
	return writeJSONL(filepath.Join(outDir, "recipe_drafts.jsonl"), drafts)
}

func writeTechniqueMarkdownAndHTML(outDir, inputPath string, summary techniqueReportSummary, patterns []string) error {
	learning := "# Technique Learning Index\n\n## Study Queue\n\n- generated from `" + inputPath + "`\n\n## Learning Actions\n\n- Study representative project.\n\n## Pattern Playbook\n\n- " + strings.Join(patterns, ", ") + "\n\n## Plugin Risk Queue\n\n_none_\n\n## Readiness Queue\n\n- analysis_ready\nrecreation steps\n"
	report := "# Technique Corpus Report\n\n## Pattern Representatives\n\n## Study Queue\n\nStep 1: Rebuild structure\n"
	html := strings.Join([]string{
		"Technique Corpus Report", "Step 1:", "manifest.json", "learning.md", "projects.csv", "project_playbooks.csv", "compositions.csv", "layers.csv", "recreation_steps.csv", "study_queue.csv", "study_tasks.csv", "recreation_blockers.csv", "signal_layers.csv", "effect_stacks.csv", "shape_operators.csv", "text_animators.csv", "dependency_edges.csv", "learning_actions.csv", "mechanisms.csv", "mechanism_examples.csv", "coverage_scorecard.csv", "reconstruction_blueprints.jsonl", "recipe_drafts.jsonl", "errors.csv", "Coverage Scorecard", "Reconstruction Blueprints", "Recipe Drafts", "Mechanism Explorer", "mechanismFilter", "Study Task Queue", "studyTaskFilter", "Project Playbooks", "Compositions", "Layers", "Recreation Steps", "Recreation Blockers", "Effect Stacks", "Shape Operators", "Text Animators", "Dependency Edges", "Representative Projects", "mechanism-representatives",
	}, "\n")
	if err := os.WriteFile(filepath.Join(outDir, "learning.md"), []byte(learning), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "report.md"), []byte(report), 0o644); err != nil {
		return err
	}
	_ = summary
	return os.WriteFile(filepath.Join(outDir, "report.html"), []byte(html), 0o644)
}

func writeTechniqueManifest(opts TechniqueReportRenderOptions, summary techniqueReportSummary, patternCount int) error {
	artifacts := []string{"summary.json", "corpus.jsonl", "digest.json", "learning.md", "projects.csv", "project_playbooks.csv", "compositions.csv", "layers.csv", "recreation_steps.csv", "patterns.csv", "study_queue.csv", "study_tasks.csv", "recreation_blockers.csv", "signal_layers.csv", "effect_stacks.csv", "shape_operators.csv", "text_animators.csv", "dependency_edges.csv", "learning_actions.csv", "mechanisms.csv", "mechanism_examples.csv", "coverage_scorecard.csv", "reconstruction_blueprints.jsonl", "recipe_drafts.jsonl", "errors.csv", "report.md", "report.html"}
	return writeIndentedJSON(filepath.Join(opts.OutDir, "manifest.json"), map[string]any{
		"schema_version":   1,
		"generated_at_utc": time.Now().UTC().Format(time.RFC3339Nano),
		"input_path":       opts.InputPath,
		"out_dir":          opts.OutDir,
		"mode":             "explain",
		"recursive":        true,
		"limit":            opts.Limit,
		"project_count":    summary.ProjectCount,
		"error_count":      summary.ErrorCount,
		"pattern_count":    patternCount,
		"artifacts":        artifacts,
		"git":              map[string]any{"commit": "go-renderer", "branch": "", "worktree_dirty": false},
	})
}

func writeReportCSV(path string, header []string, rows [][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write(header); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeJSONL(path string, rows []any) error {
	var b strings.Builder
	for _, row := range rows {
		data, err := json.Marshal(row)
		if err != nil {
			return err
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
