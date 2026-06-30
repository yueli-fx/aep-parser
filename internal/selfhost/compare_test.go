package selfhost

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareReportsWritesScalarAndCountDiffs(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "base")
	newDir := filepath.Join(root, "new")
	outDir := filepath.Join(root, "out")
	writeJSONFile(t, filepath.Join(baseDir, "summary.json"), map[string]any{
		"project_count": 2,
		"error_count":   0,
		"totals": map[string]any{
			"comp_count":           3,
			"layer_count":          10,
			"effect_count":         4,
			"text_animator_count":  1,
			"shape_operator_count": 2,
			"dependency_count":     5,
		},
		"readiness_counts": map[string]any{"analysis_ready": 1, "plugin_blocked": 1},
		"effect_counts":    map[string]any{"ADBE Fill": 2},
	})
	writeJSONFile(t, filepath.Join(newDir, "summary.json"), map[string]any{
		"project_count": 3,
		"error_count":   1,
		"totals": map[string]any{
			"comp_count":           4,
			"layer_count":          12,
			"effect_count":         7,
			"text_animator_count":  2,
			"shape_operator_count": 5,
			"dependency_count":     9,
		},
		"readiness_counts": map[string]any{"analysis_ready": 2, "plugin_blocked": 1},
		"effect_counts":    map[string]any{"ADBE Fill": 1, "ADBE Slider Control": 3},
	})
	writeJSONFile(t, filepath.Join(baseDir, "digest.json"), map[string]any{
		"patterns": []any{"kinetic_text"},
	})
	writeJSONFile(t, filepath.Join(newDir, "digest.json"), map[string]any{
		"patterns": []any{"kinetic_text", "precomp_effect_pipeline"},
	})

	result, err := CompareReports(CompareOptions{BaseDir: baseDir, NewDir: newDir, OutDir: outDir, Top: 10})

	if err != nil {
		t.Fatalf("CompareReports: %v", err)
	}
	if result.BaseDir != baseDir || result.NewDir != newDir {
		t.Fatalf("result dirs = %q %q", result.BaseDir, result.NewDir)
	}
	assertScalarDiff(t, result, "projects", 2, 3, 1)
	assertScalarDiff(t, result, "patterns", 1, 2, 1)
	assertCountDiff(t, result, "effects", "ADBE Slider Control", 0, 3, 3)
	assertCountDiff(t, result, "effects", "ADBE Fill", 2, 1, -1)

	compareJSON := filepath.Join(outDir, "compare.json")
	if _, err := os.Stat(compareJSON); err != nil {
		t.Fatalf("compare.json missing: %v", err)
	}
	compareMD, err := os.ReadFile(filepath.Join(outDir, "compare.md"))
	if err != nil {
		t.Fatalf("compare.md missing: %v", err)
	}
	if !strings.Contains(string(compareMD), "# Technique Report Compare") {
		t.Fatalf("compare.md missing title:\n%s", string(compareMD))
	}
}

func TestCompareReportsUsesDefaultOutDirAndTopLimit(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "base")
	newDir := filepath.Join(root, "new")
	writeJSONFile(t, filepath.Join(baseDir, "summary.json"), map[string]any{
		"project_count": 0,
		"totals":        map[string]any{},
		"effect_counts": map[string]any{"a": 1, "b": 1, "c": 1},
	})
	writeJSONFile(t, filepath.Join(newDir, "summary.json"), map[string]any{
		"project_count": 0,
		"totals":        map[string]any{},
		"effect_counts": map[string]any{"a": 10, "b": 8, "c": 6},
	})
	writeJSONFile(t, filepath.Join(baseDir, "digest.json"), map[string]any{"patterns": []any{}})
	writeJSONFile(t, filepath.Join(newDir, "digest.json"), map[string]any{"patterns": []any{}})

	result, err := CompareReports(CompareOptions{BaseDir: baseDir, NewDir: newDir, Top: 2})

	if err != nil {
		t.Fatalf("CompareReports: %v", err)
	}
	if len(result.CountDiffs) != 2 {
		t.Fatalf("count diffs = %d, want 2: %+v", len(result.CountDiffs), result.CountDiffs)
	}
	if _, err := os.Stat(filepath.Join(newDir, "compare", "compare.json")); err != nil {
		t.Fatalf("default compare.json missing: %v", err)
	}
}

func TestCompareReportsWritesEmptyCountDiffArrayForIdenticalReports(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "report")
	writeJSONFile(t, filepath.Join(baseDir, "summary.json"), map[string]any{
		"project_count": 1,
		"totals":        map[string]any{"comp_count": 1},
		"effect_counts": map[string]any{"ADBE Fill": 2},
	})
	writeJSONFile(t, filepath.Join(baseDir, "digest.json"), map[string]any{"patterns": []any{"kinetic_text"}})

	_, err := CompareReports(CompareOptions{BaseDir: baseDir, NewDir: baseDir})

	if err != nil {
		t.Fatalf("CompareReports: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(baseDir, "compare", "compare.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(data), `"count_diffs": null`) {
		t.Fatalf("count_diffs must be an empty array, not null:\n%s", string(data))
	}
	if !strings.Contains(string(data), `"count_diffs": []`) {
		t.Fatalf("count_diffs empty array missing:\n%s", string(data))
	}
}

func assertScalarDiff(t *testing.T, result CompareResult, name string, base, next, delta int) {
	t.Helper()
	for _, row := range result.ScalarDiffs {
		if row.Name == name {
			if row.Base != base || row.New != next || row.Delta != delta {
				t.Fatalf("scalar %s = %+v", name, row)
			}
			return
		}
	}
	t.Fatalf("missing scalar diff %q in %+v", name, result.ScalarDiffs)
}

func assertCountDiff(t *testing.T, result CompareResult, group, name string, base, next, delta int) {
	t.Helper()
	for _, row := range result.CountDiffs {
		if row.Group == group && row.Name == name {
			if row.Base != base || row.New != next || row.Delta != delta {
				t.Fatalf("count %s/%s = %+v", group, name, row)
			}
			return
		}
	}
	t.Fatalf("missing count diff %q/%q in %+v", group, name, result.CountDiffs)
}

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
