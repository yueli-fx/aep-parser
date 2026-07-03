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

func TestRenderTechniqueReportArtifactsWritesEffectFieldSurfaceWhenAvailable(t *testing.T) {
	dir := t.TempDir()
	writeDashboardFixtureInputs(t, dir)
	understandingPath := filepath.Join(dir, "understanding.json")
	writeEffectFieldUnderstandingFixture(t, understandingPath)

	if err := RenderTechniqueReportArtifacts(TechniqueReportRenderOptions{
		OutDir:                       dir,
		InputPath:                    "data/samples",
		Limit:                        2,
		EffectFieldUnderstandingPath: understandingPath,
	}); err != nil {
		t.Fatalf("RenderTechniqueReportArtifacts: %v", err)
	}

	var summary map[string]any
	if err := readIndentedJSON(filepath.Join(dir, "effect_field_summary.json"), &summary); err != nil {
		t.Fatalf("read effect_field_summary.json: %v", err)
	}
	summaryMetrics, _ := summary["summary"].(map[string]any)
	targets, _ := summary["top_study_targets"].([]any)
	if intAny(summaryMetrics["effect_kinds"]) != 2 || len(targets) == 0 {
		t.Fatalf("effect_field_summary.json = %+v", summary)
	}
	rows, err := readCSVRows(filepath.Join(dir, "effect_field_study_queue.csv"))
	if err != nil {
		t.Fatalf("read effect_field_study_queue.csv: %v", err)
	}
	if len(rows) != 2 || rows[0]["match_name"] != "tc Particular" || rows[0]["generation_policy"] != "preserve_as_dependency" {
		t.Fatalf("effect field study rows = %+v", rows)
	}
	var manifest reportVerifyManifest
	if err := readIndentedJSON(filepath.Join(dir, "manifest.json"), &manifest); err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !manifestHasArtifact(manifest, "effect_field_summary.json") || !manifestHasArtifact(manifest, "effect_field_study_queue.csv") {
		t.Fatalf("manifest artifacts = %+v", manifest.Artifacts)
	}
	if _, err := VerifyTechniqueReport(ReportVerifyOptions{OutDir: dir, MinProjects: 1}); err != nil {
		t.Fatalf("VerifyTechniqueReport: %v", err)
	}
}

func TestRenderTechniqueReportArtifactsSkipsEffectFieldSurfaceWhenMissing(t *testing.T) {
	dir := t.TempDir()
	writeDashboardFixtureInputs(t, dir)

	if err := RenderTechniqueReportArtifacts(TechniqueReportRenderOptions{
		OutDir:                       dir,
		InputPath:                    "data/samples",
		Limit:                        2,
		EffectFieldUnderstandingPath: filepath.Join(dir, "missing-understanding.json"),
	}); err != nil {
		t.Fatalf("RenderTechniqueReportArtifacts: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "effect_field_summary.json")); !os.IsNotExist(err) {
		t.Fatalf("effect_field_summary.json exists or stat failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "effect_field_study_queue.csv")); !os.IsNotExist(err) {
		t.Fatalf("effect_field_study_queue.csv exists or stat failed: %v", err)
	}
	if _, err := VerifyTechniqueReport(ReportVerifyOptions{OutDir: dir, MinProjects: 1}); err != nil {
		t.Fatalf("VerifyTechniqueReport: %v", err)
	}
}

func TestRenderTechniqueReportArtifactsRejectsMalformedEffectFieldUnderstanding(t *testing.T) {
	dir := t.TempDir()
	writeDashboardFixtureInputs(t, dir)
	understandingPath := filepath.Join(dir, "understanding.json")
	writeFile(t, understandingPath, `{"schema_version":`)

	err := RenderTechniqueReportArtifacts(TechniqueReportRenderOptions{
		OutDir:                       dir,
		InputPath:                    "data/samples",
		Limit:                        2,
		EffectFieldUnderstandingPath: understandingPath,
	})

	if err == nil || !strings.Contains(err.Error(), "effect field understanding") {
		t.Fatalf("err = %v, want malformed effect field understanding", err)
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

func writeEffectFieldUnderstandingFixture(t *testing.T, path string) {
	t.Helper()
	writeJSONFile(t, path, EffectFieldUnderstanding{
		SchemaVersion:   1,
		SourceInventory: "tmp/effect_field_inventory/inventory.json",
		Summary: EffectFieldUnderstandingSummary{
			EffectKinds:       2,
			EffectOccurrences: 12,
			ParamKinds:        7,
			ParamOccurrences:  32,
			Reproducibility: []CountRow{
				{Name: "third_party_plugin_required", Count: 8},
				{Name: "go_native_exact", Count: 4},
			},
			GenerationPolicies: []CountRow{
				{Name: "preserve_as_dependency", Count: 8},
				{Name: "emit_native", Count: 4},
			},
		},
		Effects: []EffectFieldUnderstandingEffect{
			{
				MatchName:          "tc Particular",
				Class:              "third_party",
				Occurrences:        8,
				ParamKinds:         5,
				ParamOccurrences:   28,
				Capabilities:       []string{"color", "keyframes"},
				Reproducibility:    "third_party_plugin_required",
				GenerationPolicy:   "preserve_as_dependency",
				FieldUnderstanding: "param_names_and_values_parseable",
				StudyPriority:      710,
				StudyActions:       []string{"preserve_plugin_dependency", "summarize_param_groups"},
				Boundary:           "Rendering equivalence requires the plugin.",
			},
			{
				MatchName:          "ADBE Fill",
				Class:              "native_supported",
				Occurrences:        4,
				ParamKinds:         2,
				ParamOccurrences:   4,
				Capabilities:       []string{"color"},
				Reproducibility:    "go_native_exact",
				GenerationPolicy:   "emit_native",
				FieldUnderstanding: "native_effect_can_be_emitted_by_go",
				StudyPriority:      62,
			},
		},
	})
}

func manifestHasArtifact(manifest reportVerifyManifest, name string) bool {
	for _, artifact := range manifest.Artifacts {
		if stringAny(artifact) == name {
			return true
		}
	}
	return false
}
