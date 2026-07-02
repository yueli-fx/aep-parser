package registry

import "testing"

func TestCheckVersionBoundariesMatchesCoverageAxisTotals(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeBoundaryCheckFixture(t, root, CoverageTotals{Total: 3, Pass: 1, Blocked: 1, Failed: 0, Skipped: 1}, boundaryCheckCases("blocked"))

	report, err := CheckVersionBoundaries(root, "flightdeck/work/aep-understanding-generation/coverage.json", []string{"AE2020", "AE2024", "AE2025"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	if report.Summary.Boundaries != 1 || report.Summary.Matched != 1 || len(report.Boundaries) != 1 {
		t.Fatalf("summary/boundaries = %+v/%+v", report.Summary, report.Boundaries)
	}
	if !report.Boundaries[0].Match || report.Boundaries[0].ActualBoundaryStatus != "known_matte_contract_boundary" {
		t.Fatalf("boundary check = %+v", report.Boundaries[0])
	}
	if report.Summary.CheckedCells != 3 || report.Boundaries[0].CheckedCells != 3 {
		t.Fatalf("checked cells summary/boundary = %d/%d, want 3/3", report.Summary.CheckedCells, report.Boundaries[0].CheckedCells)
	}
}

func TestCheckVersionBoundariesReportsCoverageDrift(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeBoundaryCheckFixture(t, root, CoverageTotals{Total: 3, Pass: 1, Blocked: 0, Failed: 0, Skipped: 2}, boundaryCheckCases("blocked"))

	report, err := CheckVersionBoundaries(root, "flightdeck/work/aep-understanding-generation/coverage.json", []string{"AE2020", "AE2024", "AE2025"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	if report.Summary.MismatchedTotals != 1 || report.Summary.Errors != 1 {
		t.Fatalf("summary = %+v, want one totals mismatch", report.Summary)
	}
	if len(report.Issues) != 1 || report.Issues[0].Code != "boundary_totals_mismatch" {
		t.Fatalf("issues = %+v, want totals mismatch", report.Issues)
	}
}

func TestCheckVersionBoundariesReportsCellPolicyDrift(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeBoundaryCheckFixture(t, root, CoverageTotals{Total: 3, Pass: 2, Blocked: 0, Failed: 0, Skipped: 1}, boundaryCheckCases("pass"))

	report, err := CheckVersionBoundaries(root, "flightdeck/work/aep-understanding-generation/coverage.json", []string{"AE2020", "AE2024", "AE2025"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	if report.Summary.MismatchedCells != 1 || report.Summary.MismatchedTotals != 0 {
		t.Fatalf("summary = %+v, want one cell mismatch and no totals mismatch", report.Summary)
	}
	if len(report.Boundaries) != 1 || len(report.Boundaries[0].CellMismatches) != 1 {
		t.Fatalf("boundary mismatches = %+v", report.Boundaries)
	}
	mismatch := report.Boundaries[0].CellMismatches[0]
	if mismatch.SourceVersion != "AE2025" || mismatch.TargetVersion != "AE2024" || mismatch.ExpectedStatus != "blocked" || mismatch.ActualStatus != "pass" {
		t.Fatalf("cell mismatch = %+v", mismatch)
	}
}

func writeBoundaryCheckFixture(t *testing.T, root string, expected CoverageTotals, cases []map[string]any) {
	t.Helper()
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":            "layer",
				"artifact":      "tmp/matrix/layer/matrix.json",
				"recipes":       []string{"minimal-layer-explicit-matte"},
				"writer_status": "boundary",
				"boundary": map[string]any{
					"status":             "known_matte_contract_boundary",
					"blocked_recipe_ids": []string{"minimal-layer-explicit-matte"},
					"details": []map[string]any{
						{
							"recipe":  "minimal-layer-explicit-matte",
							"reason":  "AE2025-only explicit matte source contract",
							"pass":    1,
							"blocked": 1,
							"skipped": 1,
						},
					},
				},
				"totals": map[string]any{"total": 3, "pass": 1, "blocked": 1, "failed": 0, "skipped": 1},
			},
		},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/layer/matrix.json", cases)
	writeValidRegistry(t, root, []map[string]any{
		{
			"id":           "layer.track_matte.explicit_source",
			"domain":       "layer",
			"tier":         "boundary",
			"status":       "boundary",
			"workflows":    []string{"migrate"},
			"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-layer-explicit-matte.json", "required": true}},
		},
	})
	writeJSON(t, root, "registry/version_boundaries.json", map[string]any{
		"schema_version": 1,
		"version_boundaries": []map[string]any{
			{
				"id":                       "ae2025.explicit_matte_source",
				"atom_id":                  "layer.track_matte.explicit_source",
				"recipe":                   "minimal-layer-explicit-matte",
				"feature":                  "AE2025 explicit matte source references",
				"policy":                   "known_source_contract_boundary",
				"coverage_boundary_status": "known_matte_contract_boundary",
				"source_contract": map[string]any{
					"min_source_version":        "AE2025",
					"available_source_versions": []string{"AE2025"},
				},
				"target_contract": map[string]any{
					"supported_targets": []string{"AE2025"},
				},
				"expected_cells": expected,
				"cell_policy": []map[string]any{
					{
						"source_versions": []string{"AE2020"},
						"target_versions": []string{"AE2020"},
						"status":          "skipped",
					},
					{
						"source_versions": []string{"AE2025"},
						"target_versions": []string{"AE2025"},
						"status":          "pass",
					},
					{
						"source_versions": []string{"AE2025"},
						"target_versions": []string{"AE2024"},
						"status":          "blocked",
					},
				},
				"evidence": []map[string]any{{"kind": "matrix", "path": "tmp/matrix/layer/matrix.json", "required": true}},
			},
		},
	})
}

func boundaryCheckCases(ae2025ToAE2024Status string) []map[string]any {
	return []map[string]any{
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2024", "status": ae2025ToAE2024Status},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2020", "target_version": "AE2020", "status": "skipped"},
	}
}
