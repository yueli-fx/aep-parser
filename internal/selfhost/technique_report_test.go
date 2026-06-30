package selfhost

import (
	"path/filepath"
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
