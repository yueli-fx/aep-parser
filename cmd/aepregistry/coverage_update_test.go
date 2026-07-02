package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCoverageUpdateWritesCoverageRecord(t *testing.T) {
	root := newRegistryRoot(t)
	coveragePath := "flightdeck/work/aep-understanding-generation/coverage.json"
	matrixPath := "tmp/matrix/text/matrix.json"
	writeJSON(t, root, coveragePath, map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":               "text",
				"domain":           "text",
				"scope":            "text smoke",
				"host_open_status": "pending",
			},
		},
	})
	writeJSON(t, root, matrixPath, map[string]any{
		"schema_version": 1,
		"summary": map[string]any{
			"total":   1,
			"passed":  1,
			"blocked": 0,
			"failed":  0,
			"skipped": 0,
		},
		"cases": []map[string]any{
			{"recipe_name": "minimal-text", "source_version": "AE2020", "target_version": "AE2025", "status": "pass"},
		},
	})
	out := filepath.Join(root, "tmp", "coverage_update_report.json")

	code := run([]string{
		"coverage-update",
		"-root", root,
		"-coverage", coveragePath,
		"-id", "text",
		"-matrix", matrixPath,
		"-out", out,
	})
	if code != 0 {
		t.Fatalf("run(coverage-update) = %d, want 0", code)
	}
	var report struct {
		Status       string `json:"status"`
		CoverageID   string `json:"coverage_id"`
		WriterStatus string `json:"writer_status"`
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "pass" || report.CoverageID != "text" || report.WriterStatus != "PD-6x6" {
		t.Fatalf("report = %+v", report)
	}
	var coverage map[string]any
	data, err = os.ReadFile(filepath.Join(root, filepath.FromSlash(coveragePath)))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &coverage); err != nil {
		t.Fatal(err)
	}
	record := coverage["coverage"].([]any)[0].(map[string]any)
	if record["artifact"] != "tmp/matrix/text/matrix.json" {
		t.Fatalf("artifact = %q", record["artifact"])
	}
}
