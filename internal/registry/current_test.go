package registry

import "testing"

func TestValidateCurrentPassesForReferencedArtifacts(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCurrentValidationFixture(t, root)

	report, err := ValidateCurrent(root, "registry/workflows/aep-understanding-generation/current.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass || report.Summary.Errors != 0 {
		t.Fatalf("report = %+v, want pass", report)
	}
	if report.Summary.CoverageBatches != 1 || report.Summary.Tooling != 1 {
		t.Fatalf("summary = %+v, want one batch and one tooling entry", report.Summary)
	}
}

func TestValidateCurrentReportsUnknownCoverageBatchID(t *testing.T) {
	root := newTestRegistryRoot(t)
	current := writeCurrentValidationFixture(t, root)
	current["coverage_batches"].([]map[string]any)[0]["entries"] = []map[string]any{
		{
			"coverage_id":  "missing",
			"matrix":       "tmp/matrix/text/matrix.json",
			"ledger":       "tmp/matrix/text/ledger.md",
			"recipe_paths": []string{"examples/recipes/text-basic.json"},
		},
	}
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/current.json", current)

	report, err := ValidateCurrent(root, "registry/workflows/aep-understanding-generation/current.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail", report.Status)
	}
	assertCurrentIssue(t, report, "unknown_coverage_batch_id", "coverage-batches.all.entries.missing", "")
}

func TestValidateCurrentReportsCanonicalCoverageBatchMismatch(t *testing.T) {
	root := newTestRegistryRoot(t)
	current := writeCurrentValidationFixture(t, root)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{"id": "text", "artifact": "tmp/matrix/text/matrix.json"},
			{"id": "shape", "artifact": "tmp/matrix/shape/matrix.json"},
		},
	})
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/current.json", current)

	report, err := ValidateCurrent(root, "registry/workflows/aep-understanding-generation/current.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail", report.Status)
	}
	assertCurrentIssue(t, report, "canonical_coverage_batch_mismatch", "coverage-batches.all", "")
}

func writeCurrentValidationFixture(t *testing.T, root string) map[string]any {
	t.Helper()
	currentPath := "registry/workflows/aep-understanding-generation/current.json"
	coveragePath := "registry/workflows/aep-understanding-generation/coverage.json"
	writeJSON(t, root, coveragePath, map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{"id": "text", "artifact": "tmp/matrix/text/matrix.json"},
		},
	})
	writeMatrixFixtureWithSummary(t, root, "tmp/matrix/text/matrix.json", map[string]any{
		"total": 2, "passed": 2, "blocked": 0, "failed": 0, "skipped": 0,
	})
	writeFile(t, root, "tmp/matrix/text/ledger.md", "# ledger\n")
	writeFile(t, root, "examples/recipes/text-basic.json", "{}\n")
	writeFile(t, root, "scripts/migration/tool.ps1", "")
	writeFile(t, root, "registry/workflows/aep-understanding-generation/frozen.md", "# frozen\n")

	current := map[string]any{
		"truth_sources": map[string]any{
			"current":  currentPath,
			"coverage": coveragePath,
		},
		"current_state": map[string]any{
			"latest_recurring_gate": map[string]any{
				"artifact": "tmp/matrix/text/matrix.json",
				"totals":   map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
		"frozen_markdown": []map[string]any{
			{"path": "registry/workflows/aep-understanding-generation/frozen.md"},
		},
		"tooling": []map[string]any{
			{"id": "tool", "script": "scripts/migration/tool.ps1"},
		},
		"coverage_batches": []map[string]any{
			{
				"id": "all",
				"entries": []map[string]any{
					{
						"coverage_id":  "text",
						"matrix":       "tmp/matrix/text/matrix.json",
						"ledger":       "tmp/matrix/text/ledger.md",
						"recipe_paths": []string{"examples/recipes/text-basic.json"},
					},
				},
			},
		},
		"canonical_coverage_batch": "all",
	}
	writeJSON(t, root, currentPath, current)
	return current
}

func assertCurrentIssue(t *testing.T, report CurrentValidationReport, code, field, path string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Code == code && issue.Field == field && issue.Path == path {
			return
		}
	}
	t.Fatalf("issue %q field=%q path=%q not found in %+v", code, field, path, report.Issues)
}
