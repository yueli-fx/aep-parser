# Technique Report Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the self-hosted `report.html` long dump with a static single-page dashboard that shows summary first and reveals project/pattern details on demand.

**Architecture:** Keep the report generator in `internal/selfhost/technique_report.go` for this pass. Build a small dashboard data model from existing `summary.json`, `corpus.jsonl`, generated CSV-style rows, blueprints, and recipe drafts, then render one self-contained HTML file with embedded JSON and plain JavaScript tabs/detail panels. Existing JSON/CSV/JSONL artifacts remain canonical.

**Tech Stack:** Go standard library, `encoding/json`, static HTML/CSS/JavaScript, existing `internal/selfhost` report verifier.

---

## Files

- Modify: `internal/selfhost/technique_report.go`
  - Add dashboard data structs.
  - Collect top-level counts, pattern rows, project rows, artifact links, and selected detail payloads.
  - Replace the current minimal HTML string with a real static dashboard renderer.
- Modify: `internal/selfhost/technique_report_test.go`
  - Add tests for dashboard tabs, summary-first content, project detail hooks, artifact links, and non-eager detail rendering.
- Modify: `internal/selfhost/report_verify.go`
  - Update required HTML text/hooks from the old long page vocabulary to the dashboard vocabulary while keeping existing artifact-link checks.
- Modify: `internal/selfhost/report_verify_test.go`
  - Update the minimal fixture HTML to satisfy dashboard verification.
- Optional modify: `flightdeck/work/aep-understanding-generation/technique-report-dashboard-spec.md`
  - Only if implementation reveals a small spec clarification.

## Task 1: Add Dashboard HTML Contract Tests

**Files:**
- Modify: `internal/selfhost/technique_report_test.go`

- [ ] **Step 1: Write failing tests for dashboard structure**

Add imports:

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)
```

Add this test:

```go
func TestRenderTechniqueReportArtifactsWritesHumanDashboard(t *testing.T) {
	dir := t.TempDir()
	writeDashboardFixtureInputs(t, dir)

	if err := RenderTechniqueReportArtifacts(TechniqueReportRenderOptions{
		OutDir:    dir,
		InputPath: "data/samples",
		Limit:     2,
	}); err != nil {
		t.Fatalf("RenderTechniqueReportArtifacts: %v", err)
	}

	htmlBytes, err := os.ReadFile(filepath.Join(dir, "report.html"))
	if err != nil {
		t.Fatalf("ReadFile report.html: %v", err)
	}
	html := string(htmlBytes)
	for _, want := range []string{
		"Technique Corpus Dashboard",
		`data-tab="overview"`,
		`data-tab="patterns"`,
		`data-tab="projects"`,
		`data-tab="recipes"`,
		`data-tab="data"`,
		`id="project-search"`,
		`id="project-detail"`,
		`window.__TECHNIQUE_REPORT__`,
		"summary.json",
		"corpus.jsonl",
		"recipe_drafts.jsonl",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("report.html missing %q\n%s", want, html)
		}
	}
}
```

- [ ] **Step 2: Add non-eager detail test**

Add this test below it:

```go
func TestRenderTechniqueReportArtifactsDoesNotEagerlyRenderAllProjectDetails(t *testing.T) {
	dir := t.TempDir()
	writeDashboardFixtureInputs(t, dir)

	if err := RenderTechniqueReportArtifacts(TechniqueReportRenderOptions{
		OutDir:    dir,
		InputPath: "data/samples",
		Limit:     2,
	}); err != nil {
		t.Fatalf("RenderTechniqueReportArtifacts: %v", err)
	}

	htmlBytes, err := os.ReadFile(filepath.Join(dir, "report.html"))
	if err != nil {
		t.Fatalf("ReadFile report.html: %v", err)
	}
	html := string(htmlBytes)
	if strings.Contains(html, `<section class="project-detail-card"`) {
		t.Fatalf("project details are rendered eagerly into the DOM")
	}
	if !strings.Contains(html, `"project_path":"alpha.aep"`) || !strings.Contains(html, `"project_path":"beta.aep"`) {
		t.Fatalf("embedded dashboard data should still include selectable projects")
	}
}
```

- [ ] **Step 3: Add fixture helper**

Add this helper to the test file:

```go
func writeDashboardFixtureInputs(t *testing.T, dir string) {
	t.Helper()
	writeJSONFile(t, filepath.Join(dir, "summary.json"), map[string]any{
		"schema_version": 1,
		"mode":           "explain",
		"project_count":  2,
		"error_count":    1,
		"totals": map[string]any{
			"comp_count":           3,
			"layer_count":          6,
			"effect_count":         2,
			"shape_operator_count": 2,
			"text_animator_count":  1,
			"dependency_count":     2,
		},
		"pattern_counts": map[string]any{"shape_system": 2, "text_animation": 1},
	})
	writeFile(t, filepath.Join(dir, "corpus.jsonl"),
		`{"path":"alpha.aep","explanation":{"recreation_steps":[{"id":"structure","priority":10,"title":"Build structure","summary":"Create comps and layers."}]}}`+"\n"+
			`{"path":"beta.aep","explanation":{"recreation_steps":[{"id":"text","priority":20,"title":"Animate text","summary":"Recreate text animator timing."}]}}`+"\n"+
			`{"path":"broken.aep","mode":"explain","error":"parse failed"}`+"\n")
}
```

- [ ] **Step 4: Run the focused tests and verify failure**

Run:

```powershell
go test ./internal/selfhost -run 'TestRenderTechniqueReportArtifactsWritesHumanDashboard|TestRenderTechniqueReportArtifactsDoesNotEagerlyRenderAllProjectDetails' -count=1
```

Expected: both new tests fail because `report.html` is still the old line-oriented minimal output.

## Task 2: Add Dashboard Data Model

**Files:**
- Modify: `internal/selfhost/technique_report.go`

- [ ] **Step 1: Add dashboard structs near `techniqueTotals`**

```go
type techniqueDashboard struct {
	GeneratedAtUTC string                     `json:"generated_at_utc"`
	InputPath      string                     `json:"input_path"`
	Limit          int                        `json:"limit"`
	Summary        techniqueDashboardSummary  `json:"summary"`
	Patterns       []techniqueDashboardPattern `json:"patterns"`
	Projects       []techniqueDashboardProject `json:"projects"`
	Recipes        []techniqueDashboardRecipe  `json:"recipes"`
	Artifacts      []techniqueDashboardArtifact `json:"artifacts"`
}

type techniqueDashboardSummary struct {
	ProjectCount        int             `json:"project_count"`
	ErrorCount          int             `json:"error_count"`
	PatternCount        int             `json:"pattern_count"`
	Totals              techniqueTotals `json:"totals"`
	TopFinding          string          `json:"top_finding"`
	RiskSummary         string          `json:"risk_summary"`
	NextAction          string          `json:"next_action"`
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
```

- [ ] **Step 2: Run compile to verify expected struct formatting errors only if any**

Run:

```powershell
go test ./internal/selfhost -run TestRenderTechniqueReportArtifactsPassesReportVerification -count=1
```

Expected: compile succeeds; behavior still fails the new dashboard tests from Task 1.

## Task 3: Build Dashboard Data From Existing Artifacts

**Files:**
- Modify: `internal/selfhost/technique_report.go`

- [ ] **Step 1: Change `RenderTechniqueReportArtifacts` to pass data into renderer**

Replace:

```go
if err := writeTechniqueMarkdownAndHTML(opts.OutDir, opts.InputPath, summary, patterns); err != nil {
	return err
}
```

With:

```go
dashboard := buildTechniqueDashboard(opts, summary, patterns, corpusRecords, projects)
if err := writeTechniqueMarkdownAndHTML(opts.OutDir, opts.InputPath, summary, patterns, dashboard); err != nil {
	return err
}
```

- [ ] **Step 2: Add data builder functions before `writeTechniqueMarkdownAndHTML`**

```go
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
		{Name: "projects.csv", Description: "Project readiness table."},
		{Name: "patterns.csv", Description: "Detected pattern identifiers."},
		{Name: "reconstruction_blueprints.jsonl", Description: "Per-project reconstruction plans."},
		{Name: "recipe_drafts.jsonl", Description: "Generated recipe draft candidates."},
		{Name: "errors.csv", Description: "Parse or analysis errors."},
	}
}
```

- [ ] **Step 3: Update the `writeTechniqueMarkdownAndHTML` signature**

Change:

```go
func writeTechniqueMarkdownAndHTML(outDir, inputPath string, summary techniqueReportSummary, patterns []string) error {
```

To:

```go
func writeTechniqueMarkdownAndHTML(outDir, inputPath string, summary techniqueReportSummary, patterns []string, dashboard techniqueDashboard) error {
```

- [ ] **Step 4: Run focused tests**

Run:

```powershell
go test ./internal/selfhost -run TestRenderTechniqueReportArtifacts -count=1
```

Expected: compile succeeds; dashboard HTML tests still fail until Task 4 renders real HTML.

## Task 4: Render Static Dashboard HTML

**Files:**
- Modify: `internal/selfhost/technique_report.go`

- [ ] **Step 1: Replace minimal HTML generation**

Inside `writeTechniqueMarkdownAndHTML`, replace the current `html := strings.Join(...)` assignment with:

```go
html, err := renderTechniqueDashboardHTML(dashboard)
if err != nil {
	return err
}
```

- [ ] **Step 2: Add HTML renderer**

Add after `writeTechniqueMarkdownAndHTML`:

```go
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
```

- [ ] **Step 3: Run focused tests**

Run:

```powershell
go test ./internal/selfhost -run 'TestRenderTechniqueReportArtifactsWritesHumanDashboard|TestRenderTechniqueReportArtifactsDoesNotEagerlyRenderAllProjectDetails' -count=1
```

Expected: both dashboard tests pass.

## Task 5: Update Report Verification For Dashboard Vocabulary

**Files:**
- Modify: `internal/selfhost/report_verify.go`
- Modify: `internal/selfhost/report_verify_test.go`

- [ ] **Step 1: Update `requireReportTexts` dashboard requirements**

In `requireReportTexts`, replace only the HTML pattern list with:

```go
for _, pattern := range []string{
	`Technique Corpus Dashboard`,
	`data-tab="overview"`,
	`data-tab="patterns"`,
	`data-tab="projects"`,
	`data-tab="recipes"`,
	`data-tab="data"`,
	`id="project-search"`,
	`id="project-detail"`,
	`window\.__TECHNIQUE_REPORT__`,
	`summary\.json`,
	`corpus\.jsonl`,
	`digest\.json`,
	`projects\.csv`,
	`project_playbooks\.csv`,
	`compositions\.csv`,
	`layers\.csv`,
	`recreation_steps\.csv`,
	`study_queue\.csv`,
	`study_tasks\.csv`,
	`recreation_blockers\.csv`,
	`signal_layers\.csv`,
	`effect_stacks\.csv`,
	`shape_operators\.csv`,
	`text_animators\.csv`,
	`dependency_edges\.csv`,
	`learning_actions\.csv`,
	`mechanisms\.csv`,
	`mechanism_examples\.csv`,
	`coverage_scorecard\.csv`,
	`reconstruction_blueprints\.jsonl`,
	`recipe_drafts\.jsonl`,
	`errors\.csv`,
} {
	if err := requireText(paths.html, pattern); err != nil {
		return err
	}
}
```

Keep the `learning.md` and `report.md` requirements unchanged.

- [ ] **Step 2: Update minimal verifier fixture HTML**

In `writeMinimalTechniqueReport`, replace the `report.html` fixture string with a small dashboard-compatible string:

```go
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
```

- [ ] **Step 3: Run verifier tests**

Run:

```powershell
go test ./internal/selfhost -run 'TestVerifyTechniqueReport|TestRenderTechniqueReportArtifactsPassesReportVerification' -count=1
```

Expected: tests pass.

## Task 6: Verify CLI Integration

**Files:**
- No code changes expected.

- [ ] **Step 1: Run focused command tests**

Run:

```powershell
go test ./cmd/aepselfhost ./internal/selfhost -count=1
```

Expected: both packages pass.

- [ ] **Step 2: Generate a small real report**

Run:

```powershell
go run ./cmd/aepselfhost technique-report -limit 3 -verify
```

Expected: command exits `0` and prints paths for `summary.json`, `corpus.jsonl`, `manifest.json`, and `report.html`.

- [ ] **Step 3: Inspect generated HTML size and hooks**

Run:

```powershell
Select-String -Path tmp\technique_showcase_report\report.html -Pattern 'Technique Corpus Dashboard','id="project-search"','window.__TECHNIQUE_REPORT__'
```

Expected: all three patterns are present.

## Task 7: Full Verification And Commit

**Files:**
- Commit all modified Go/test files and any spec clarification.

- [ ] **Step 1: Run full verification**

Run:

```powershell
go test ./... -count=1
go vet ./...
git diff --check
```

Expected: all commands exit `0`.

- [ ] **Step 2: Review changed file list**

Run:

```powershell
git status --short
git diff --stat
```

Expected: changes are limited to the selfhost renderer, selfhost tests/verifier tests, and possibly the dashboard spec.

- [ ] **Step 3: Commit**

Run:

```powershell
git add internal/selfhost/technique_report.go internal/selfhost/technique_report_test.go internal/selfhost/report_verify.go internal/selfhost/report_verify_test.go flightdeck/work/aep-understanding-generation/technique-report-dashboard-spec.md
git commit -m "feat: render technique report as dashboard"
```

Expected: commit succeeds.

## Self-Review

- Spec coverage: Overview, Patterns, Projects, Recipes, Data, static single-file rendering, summary-first UX, lazy detail rendering, and artifact links are covered by Tasks 1-5.
- Scope control: the plan does not redesign canonical machine artifacts and does not introduce a web server.
- Type consistency: all new structs use existing helper functions (`stringAny`, `intAny`, `nestedSlice`) and existing `techniqueTotals`.
- Verification: focused tests, CLI generation, full `go test`, `go vet`, and `git diff --check` are included.
