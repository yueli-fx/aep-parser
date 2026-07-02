package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateCoverageFromMatrixPreservesRecordFields(t *testing.T) {
	root := newTestRegistryRoot(t)
	coveragePath := "flightdeck/work/aep-understanding-generation/coverage.json"
	matrixPath := "tmp/matrix/text/matrix.json"
	writeJSON(t, root, coveragePath, map[string]any{
		"schema_version": 1,
		"generated_at":   "2026-01-01",
		"coverage": []map[string]any{
			{
				"id":               "text",
				"domain":           "text",
				"scope":            "old scope",
				"writer_status":    "old",
				"writer_coverage":  "old coverage",
				"host_open_status": "existing-host",
				"custom_note":      "keep me",
				"host_open_endpoint_evidence": map[string]any{
					"scope": "non_representative_recipes",
				},
				"artifact": "old/matrix.json",
				"totals":   map[string]any{"total": 1, "pass": 1, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeMatrixFixtureWithCases(t, root, matrixPath, []map[string]any{
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2024", "status": "blocked"},
	})
	writeFile(t, root, "tmp/matrix/text/ledger.md", "ledger")

	report, err := UpdateCoverageFromMatrix(root, CoverageUpdateOptions{
		ID:           "text",
		CoveragePath: coveragePath,
		MatrixPath:   matrixPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Created || report.WriterStatus != "boundary" || report.Totals.Blocked != 1 {
		t.Fatalf("report = %+v", report)
	}

	coverage := readCoverageMap(t, root, coveragePath)
	record := firstCoverageRecord(t, coverage, "text")
	if record["custom_note"] != "keep me" {
		t.Fatalf("custom field was not preserved: %+v", record)
	}
	endpoint := record["host_open_endpoint_evidence"].(map[string]any)
	if endpoint["scope"] != "non_representative_recipes" {
		t.Fatalf("endpoint evidence was not preserved: %+v", endpoint)
	}
	if record["artifact"] != "tmp/matrix/text/matrix.json" || record["ledger"] != "tmp/matrix/text/ledger.md" {
		t.Fatalf("artifact/ledger = %q/%q", record["artifact"], record["ledger"])
	}
	if record["writer_status"] != "boundary" {
		t.Fatalf("writer_status = %q", record["writer_status"])
	}
	if record["writer_coverage"] != "AE2020,AE2025 sources into AE2024,AE2025 targets" {
		t.Fatalf("writer_coverage = %q", record["writer_coverage"])
	}
	if record["host_open_status"] != "existing-host" {
		t.Fatalf("host_open_status = %q", record["host_open_status"])
	}
	assertJSONStrings(t, record["recipes"], []string{"minimal-text-a", "minimal-text-b"})
	if record["recipe_count"].(float64) != 2 {
		t.Fatalf("recipe_count = %v", record["recipe_count"])
	}
	totals := record["totals"].(map[string]any)
	if totals["total"].(float64) != 2 || totals["pass"].(float64) != 1 || totals["blocked"].(float64) != 1 {
		t.Fatalf("totals = %+v", totals)
	}
	if coverage["generated_at"] == "2026-01-01" {
		t.Fatalf("generated_at was not refreshed")
	}
}

func TestUpdateCoverageFromMatrixCreatesRecordWhenDomainAndScopeProvided(t *testing.T) {
	root := newTestRegistryRoot(t)
	coveragePath := "flightdeck/work/aep-understanding-generation/coverage.json"
	matrixPath := "tmp/matrix/effects/matrix.json"
	writeJSON(t, root, coveragePath, map[string]any{
		"schema_version": 1,
		"coverage":       []map[string]any{},
	})
	writeMatrixFixtureWithCases(t, root, matrixPath, []map[string]any{
		{"recipe_name": "minimal-effect", "source_version": "AE2020", "target_version": "AE2025", "status": "pass"},
	})

	report, err := UpdateCoverageFromMatrix(root, CoverageUpdateOptions{
		ID:             "effects-static",
		CoveragePath:   coveragePath,
		MatrixPath:     matrixPath,
		Domain:         "effects",
		Scope:          "static effect smoke",
		HostOpenStatus: "pending_endpoint",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Created || report.WriterStatus != "PD-6x6" {
		t.Fatalf("report = %+v", report)
	}
	record := firstCoverageRecord(t, readCoverageMap(t, root, coveragePath), "effects-static")
	if record["domain"] != "effects" || record["scope"] != "static effect smoke" {
		t.Fatalf("created record domain/scope = %+v", record)
	}
	if record["host_open_status"] != "pending_endpoint" {
		t.Fatalf("host_open_status = %q", record["host_open_status"])
	}
	assertJSONStrings(t, record["recipes"], []string{"minimal-effect"})
}

func readCoverageMap(t *testing.T, root, rel string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func firstCoverageRecord(t *testing.T, coverage map[string]any, id string) map[string]any {
	t.Helper()
	records := coverage["coverage"].([]any)
	for _, raw := range records {
		record := raw.(map[string]any)
		if record["id"] == id {
			return record
		}
	}
	t.Fatalf("coverage record %q not found in %+v", id, records)
	return nil
}

func assertJSONStrings(t *testing.T, value any, want []string) {
	t.Helper()
	items := value.([]any)
	if len(items) != len(want) {
		t.Fatalf("strings = %+v, want %+v", items, want)
	}
	for i, item := range items {
		if item != want[i] {
			t.Fatalf("strings = %+v, want %+v", items, want)
		}
	}
}
