package selfhost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTechniqueReportArtifactsPassesReportVerification(t *testing.T) {
	dir := t.TempDir()
	writeJSONFile(t, filepath.Join(dir, "summary.json"), map[string]any{
		"schema_version": 1,
		"mode":           "explain",
		"project_count":  1,
		"error_count":    0,
		"totals": map[string]any{
			"comp_count":           2,
			"layer_count":          3,
			"effect_count":         2,
			"shape_operator_count": 2,
			"text_animator_count":  1,
			"dependency_count":     2,
		},
		"pattern_counts": map[string]any{"shape_system": 1},
	})
	writeFile(t, filepath.Join(dir, "corpus.jsonl"), `{"path":"demo.aep","explanation":{"recreation_steps":[{"id":"structure","priority":10,"title":"Build","summary":"Build structure"}]}}`+"\n")

	if err := RenderTechniqueReportArtifacts(TechniqueReportRenderOptions{
		OutDir:    dir,
		InputPath: "data/samples",
		Limit:     1,
	}); err != nil {
		t.Fatalf("RenderTechniqueReportArtifacts: %v", err)
	}
	if _, err := VerifyTechniqueReport(ReportVerifyOptions{OutDir: dir, MinProjects: 1}); err != nil {
		t.Fatalf("VerifyTechniqueReport: %v", err)
	}
}

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
