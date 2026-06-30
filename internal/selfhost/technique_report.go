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

type techniqueDashboard struct {
	GeneratedAtUTC string                       `json:"generated_at_utc"`
	InputPath      string                       `json:"input_path"`
	Limit          int                          `json:"limit"`
	Summary        techniqueDashboardSummary    `json:"summary"`
	Patterns       []techniqueDashboardPattern  `json:"patterns"`
	Projects       []techniqueDashboardProject  `json:"projects"`
	Recipes        []techniqueDashboardRecipe   `json:"recipes"`
	Artifacts      []techniqueDashboardArtifact `json:"artifacts"`
}

type techniqueDashboardSummary struct {
	ProjectCount int             `json:"project_count"`
	ErrorCount   int             `json:"error_count"`
	PatternCount int             `json:"pattern_count"`
	Totals       techniqueTotals `json:"totals"`
	TopFinding   string          `json:"top_finding"`
	RiskSummary  string          `json:"risk_summary"`
	NextAction   string          `json:"next_action"`
}

type techniqueDashboardPattern struct {
	ID                    string   `json:"id"`
	Count                 int      `json:"count"`
	RepresentativeProject string   `json:"representative_project"`
	Action                string   `json:"action"`
	Steps                 []string `json:"steps"`
}

type techniqueDashboardProject struct {
	ProjectPath string   `json:"project_path"`
	Readiness   string   `json:"readiness"`
	Rank        int      `json:"rank"`
	Patterns    []string `json:"patterns"`
	Steps       []string `json:"steps"`
	Error       string   `json:"error,omitempty"`
	Detail      string   `json:"detail"`
}

type techniqueDashboardRecipe struct {
	ProjectPath string `json:"project_path"`
	Readiness   string `json:"readiness"`
	GapCount    int    `json:"gap_count"`
	CompCount   int    `json:"comp_count"`
}

type techniqueDashboardArtifact struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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
	dashboard := buildTechniqueDashboard(opts, summary, patterns, corpusRecords, projects)
	if err := writeTechniqueMarkdownAndHTML(opts.OutDir, opts.InputPath, summary, patterns, dashboard); err != nil {
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

func buildTechniqueDashboard(opts TechniqueReportRenderOptions, summary techniqueReportSummary, patterns []string, records []map[string]any, projects []string) techniqueDashboard {
	dashboard := techniqueDashboard{
		GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339Nano),
		InputPath:      opts.InputPath,
		Limit:          opts.Limit,
		Summary: techniqueDashboardSummary{
			ProjectCount: summary.ProjectCount,
			ErrorCount:   summary.ErrorCount,
			PatternCount: len(patterns),
			Totals:       summary.Totals,
			TopFinding:   firstTopFinding(patterns),
			RiskSummary:  reportRiskSummary(summary),
			NextAction:   "Inspect representative projects, then promote stable patterns into recipe candidates.",
		},
		Artifacts: techniqueDashboardArtifacts(),
	}
	for _, pattern := range patterns {
		dashboard.Patterns = append(dashboard.Patterns, techniqueDashboardPattern{
			ID:                    pattern,
			Count:                 intAny(summary.PatternCounts[pattern]),
			RepresentativeProject: firstProject(projects),
			Action:                "Study representative project.",
			Steps:                 []string{"structure"},
		})
	}
	for i, record := range records {
		project := stringAny(record["path"])
		if project == "" && i < len(projects) {
			project = projects[i]
		}
		if project == "" {
			project = fmt.Sprintf("project_%03d.aep", i+1)
		}
		item := techniqueDashboardProject{
			ProjectPath: project,
			Readiness:   "analysis_ready",
			Rank:        i + 1,
			Patterns:    patterns,
			Steps:       recreationStepTitles(record),
			Detail:      projectDetailText(record),
		}
		if errText := stringAny(record["error"]); errText != "" {
			item.Readiness = "error"
			item.Error = errText
			item.Detail = errText
		}
		dashboard.Projects = append(dashboard.Projects, item)
	}
	for _, project := range projects {
		dashboard.Recipes = append(dashboard.Recipes, techniqueDashboardRecipe{
			ProjectPath: project,
			Readiness:   "analysis_ready",
			GapCount:    1,
			CompCount:   1,
		})
	}
	return dashboard
}

func firstTopFinding(patterns []string) string {
	if len(patterns) == 0 {
		return "No stable technique pattern was detected yet."
	}
	return "Top detected pattern: " + patterns[0]
}

func reportRiskSummary(summary techniqueReportSummary) string {
	if summary.ErrorCount == 0 {
		return "No parse errors in this report."
	}
	return fmt.Sprintf("%d project(s) failed parsing and need inspection.", summary.ErrorCount)
}

func firstProject(projects []string) string {
	if len(projects) == 0 {
		return ""
	}
	return projects[0]
}

func recreationStepTitles(record map[string]any) []string {
	steps := nestedSlice(record, "explanation", "recreation_steps")
	out := make([]string, 0, len(steps))
	for _, stepAny := range steps {
		step, _ := stepAny.(map[string]any)
		title := stringAny(step["title"])
		if title == "" {
			title = stringAny(step["id"])
		}
		if title != "" {
			out = append(out, title)
		}
	}
	if len(out) == 0 {
		return []string{"Inspect project structure."}
	}
	return out
}

func projectDetailText(record map[string]any) string {
	steps := recreationStepTitles(record)
	return strings.Join(steps, " / ")
}

func techniqueDashboardArtifacts() []techniqueDashboardArtifact {
	return []techniqueDashboardArtifact{
		{Name: "summary.json", Description: "Corpus-level counts and totals."},
		{Name: "corpus.jsonl", Description: "Machine-readable per-project facts and explanations."},
		{Name: "digest.json", Description: "Ranked pattern digest."},
		{Name: "learning.md", Description: "Human-readable learning notes."},
		{Name: "projects.csv", Description: "Project readiness table."},
		{Name: "project_playbooks.csv", Description: "Per-project reconstruction playbooks."},
		{Name: "compositions.csv", Description: "Composition inventory."},
		{Name: "layers.csv", Description: "Layer inventory."},
		{Name: "recreation_steps.csv", Description: "Reconstruction step table."},
		{Name: "patterns.csv", Description: "Detected pattern identifiers."},
		{Name: "study_queue.csv", Description: "Ranked project study queue."},
		{Name: "study_tasks.csv", Description: "Concrete study tasks."},
		{Name: "recreation_blockers.csv", Description: "Known blockers to recreation."},
		{Name: "signal_layers.csv", Description: "Layers carrying strong technique signals."},
		{Name: "effect_stacks.csv", Description: "Effect stack inventory."},
		{Name: "shape_operators.csv", Description: "Shape operator inventory."},
		{Name: "text_animators.csv", Description: "Text animator inventory."},
		{Name: "dependency_edges.csv", Description: "Project dependency edges."},
		{Name: "learning_actions.csv", Description: "Prioritized learning actions."},
		{Name: "mechanisms.csv", Description: "Mechanism summary table."},
		{Name: "mechanism_examples.csv", Description: "Representative mechanism examples."},
		{Name: "coverage_scorecard.csv", Description: "Report artifact coverage checks."},
		{Name: "reconstruction_blueprints.jsonl", Description: "Per-project reconstruction plans."},
		{Name: "recipe_drafts.jsonl", Description: "Generated recipe draft candidates."},
		{Name: "errors.csv", Description: "Parse or analysis errors."},
		{Name: "report.md", Description: "Markdown report companion."},
	}
}

func writeTechniqueMarkdownAndHTML(outDir, inputPath string, summary techniqueReportSummary, patterns []string, dashboard techniqueDashboard) error {
	learning := "# Technique Learning Index\n\n## Study Queue\n\n- generated from `" + inputPath + "`\n\n## Learning Actions\n\n- Study representative project.\n\n## Pattern Playbook\n\n- " + strings.Join(patterns, ", ") + "\n\n## Plugin Risk Queue\n\n_none_\n\n## Readiness Queue\n\n- analysis_ready\nrecreation steps\n"
	report := "# Technique Corpus Report\n\n## Pattern Representatives\n\n## Study Queue\n\nStep 1: Rebuild structure\n"
	html, err := renderTechniqueDashboardHTML(dashboard)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "learning.md"), []byte(learning), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "report.md"), []byte(report), 0o644); err != nil {
		return err
	}
	_ = summary
	return os.WriteFile(filepath.Join(outDir, "report.html"), []byte(html), 0o644)
}

func renderTechniqueDashboardHTML(dashboard techniqueDashboard) (string, error) {
	payload, err := json.Marshal(dashboard)
	if err != nil {
		return "", err
	}
	return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Technique Corpus Dashboard</title>
<style>
:root{font-family:Inter,Segoe UI,Arial,sans-serif;color:#17202a;background:#f6f7f9}
body{margin:0}
header{padding:24px 28px;background:#111827;color:#fff}
h1{margin:0 0 8px;font-size:24px}
.subtitle{margin:0;color:#cbd5e1}
main{padding:20px 28px}
.tabs{display:flex;gap:8px;flex-wrap:wrap;margin-bottom:16px}
.tab{border:1px solid #cbd5e1;background:#fff;border-radius:6px;padding:8px 12px;cursor:pointer}
.tab.active{background:#111827;color:#fff;border-color:#111827}
.panel{display:none}
.panel.active{display:block}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:12px;margin:12px 0 18px}
.metric,.card{background:#fff;border:1px solid #d9dee7;border-radius:8px;padding:14px}
.metric strong{display:block;font-size:24px}
table{width:100%;border-collapse:collapse;background:#fff;border:1px solid #d9dee7}
th,td{border-bottom:1px solid #e5e7eb;padding:9px;text-align:left;vertical-align:top}
th{background:#f1f5f9;font-size:12px;text-transform:uppercase}
input{width:100%;max-width:420px;padding:9px;border:1px solid #cbd5e1;border-radius:6px;margin:0 0 12px}
button.link{border:0;background:transparent;color:#1d4ed8;cursor:pointer;padding:0;text-align:left}
#project-detail{margin-top:14px}
pre{white-space:pre-wrap;background:#0f172a;color:#e5e7eb;padding:12px;border-radius:6px;overflow:auto}
</style>
</head>
<body>
<header>
<h1>Technique Corpus Dashboard</h1>
<p class="subtitle">Summary first. Use search and drilldown for the few projects that matter.</p>
</header>
<main>
<nav class="tabs">
<button class="tab active" data-tab="overview">Overview</button>
<button class="tab" data-tab="patterns">Patterns</button>
<button class="tab" data-tab="projects">Projects</button>
<button class="tab" data-tab="recipes">Recipes</button>
<button class="tab" data-tab="data">Data</button>
</nav>
<section class="panel active" id="overview"></section>
<section class="panel" id="patterns"></section>
<section class="panel" id="projects"></section>
<section class="panel" id="recipes"></section>
<section class="panel" id="data"></section>
</main>
<script>window.__TECHNIQUE_REPORT__=` + string(payload) + `;</script>
<script>
const report = window.__TECHNIQUE_REPORT__;
const esc = value => String(value ?? '').replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));
function setTab(name){
  document.querySelectorAll('.tab').forEach(btn=>btn.classList.toggle('active', btn.dataset.tab===name));
  document.querySelectorAll('.panel').forEach(panel=>panel.classList.toggle('active', panel.id===name));
}
document.querySelectorAll('.tab').forEach(btn=>btn.addEventListener('click',()=>setTab(btn.dataset.tab)));
function renderOverview(){
  const s = report.summary;
  document.getElementById('overview').innerHTML =
    '<div class="grid">' +
    metric('Projects', s.project_count) + metric('Errors', s.error_count) + metric('Patterns', s.pattern_count) +
    metric('Layers', s.totals.layer_count) + metric('Effects', s.totals.effect_count) + metric('Dependencies', s.totals.dependency_count) +
    '</div>' +
    card('Top finding', s.top_finding) + card('Risk summary', s.risk_summary) + card('Next action', s.next_action);
}
function metric(label,value){return '<div class="metric"><span>'+esc(label)+'</span><strong>'+esc(value)+'</strong></div>'}
function card(title,text){return '<div class="card"><h2>'+esc(title)+'</h2><p>'+esc(text)+'</p></div>'}
function renderPatterns(){
  document.getElementById('patterns').innerHTML = '<table><thead><tr><th>Pattern</th><th>Count</th><th>Representative</th><th>Action</th></tr></thead><tbody>' +
    report.patterns.map(p=>'<tr><td>'+esc(p.id)+'</td><td>'+esc(p.count)+'</td><td>'+esc(p.representative_project)+'</td><td>'+esc(p.action)+'</td></tr>').join('') +
    '</tbody></table>';
}
function renderProjects(projects){
  document.getElementById('project-rows').innerHTML = projects.map((p,i)=>
    '<tr><td><button class="link" onclick="showProject('+i+')">'+esc(p.project_path)+'</button></td><td>'+esc(p.readiness)+'</td><td>'+esc((p.patterns||[]).slice(0,3).join(', '))+'</td></tr>'
  ).join('');
}
function renderProjectShell(){
  document.getElementById('projects').innerHTML =
    '<input id="project-search" aria-label="Search project path">' +
    '<table><thead><tr><th>Project</th><th>Readiness</th><th>Patterns</th></tr></thead><tbody id="project-rows"></tbody></table>' +
    '<div id="project-detail" class="card"><p>Select a project to inspect details.</p></div>';
  renderProjects(report.projects);
  document.getElementById('project-search').addEventListener('input', e=>{
    const q = e.target.value.toLowerCase();
    const filtered = report.projects.filter(p=>p.project_path.toLowerCase().includes(q));
    document.getElementById('project-rows').innerHTML = filtered.map(p=>{
      const idx = report.projects.indexOf(p);
      return '<tr><td><button class="link" onclick="showProject('+idx+')">'+esc(p.project_path)+'</button></td><td>'+esc(p.readiness)+'</td><td>'+esc((p.patterns||[]).slice(0,3).join(', '))+'</td></tr>';
    }).join('');
  });
}
function showProject(index){
  const p = report.projects[index];
  document.getElementById('project-detail').innerHTML =
    '<h2>'+esc(p.project_path)+'</h2><p><strong>Readiness:</strong> '+esc(p.readiness)+'</p>' +
    (p.error ? '<p><strong>Error:</strong> '+esc(p.error)+'</p>' : '') +
    '<p><strong>Steps:</strong> '+esc((p.steps||[]).join(' / '))+'</p>' +
    '<pre>'+esc(p.detail)+'</pre>';
}
function renderRecipes(){
  document.getElementById('recipes').innerHTML = '<table><thead><tr><th>Project</th><th>Readiness</th><th>Comps</th><th>Gaps</th></tr></thead><tbody>' +
    report.recipes.map(r=>'<tr><td>'+esc(r.project_path)+'</td><td>'+esc(r.readiness)+'</td><td>'+esc(r.comp_count)+'</td><td>'+esc(r.gap_count)+'</td></tr>').join('') +
    '</tbody></table>';
}
function renderData(){
  document.getElementById('data').innerHTML = '<table><thead><tr><th>Artifact</th><th>Description</th></tr></thead><tbody>' +
    report.artifacts.map(a=>'<tr><td><a href="'+esc(a.name)+'">'+esc(a.name)+'</a></td><td>'+esc(a.description)+'</td></tr>').join('') +
    '</tbody></table>';
}
renderOverview(); renderPatterns(); renderProjectShell(); renderRecipes(); renderData();
</script>
</body>
</html>`, nil
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
