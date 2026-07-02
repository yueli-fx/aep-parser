package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListCoverageBatchesReportsEntries(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageBatchFixture(t, root, 2)

	report, err := ListCoverageBatches(root, "flightdeck/work/aep-understanding-generation/current.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass || report.Summary.Batches != 1 || report.Summary.Entries != 1 {
		t.Fatalf("report = %+v, want one listed batch", report)
	}
	if len(report.Batches) != 1 || report.Batches[0].ID != "all" || report.Batches[0].Entries != 1 {
		t.Fatalf("batches = %+v, want all/1", report.Batches)
	}
}

func TestCheckCoverageBatchSkipRunPassesForMatchingArtifacts(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageBatchFixture(t, root, 2)

	report, err := CheckCoverageBatch(root, "flightdeck/work/aep-understanding-generation/current.json", "flightdeck/work/aep-understanding-generation/coverage.json", "all")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass || report.Summary.Errors != 0 || report.Summary.Entries != 1 {
		t.Fatalf("report = %+v, want passing batch check", report)
	}
}

func TestCheckCoverageBatchReportsCoverageTotalsDrift(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageBatchFixture(t, root, 3)

	report, err := CheckCoverageBatch(root, "flightdeck/work/aep-understanding-generation/current.json", "flightdeck/work/aep-understanding-generation/coverage.json", "all")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail", report.Status)
	}
	assertCoverageBatchIssue(t, report, "coverage_totals_mismatch", "text", "tmp/matrix/text/matrix.json")
}

func TestSyncCoverageBatchFromMatricesUpdatesCoverageCandidate(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageBatchFixture(t, root, 0)

	report, err := SyncCoverageBatchFromMatrices(root, "flightdeck/work/aep-understanding-generation/current.json", "flightdeck/work/aep-understanding-generation/coverage.json", "all")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass || report.Summary.Errors != 0 || report.Summary.Entries != 1 {
		t.Fatalf("report = %+v, want passing synced batch", report)
	}

	check, err := CheckCoverageBatch(root, "flightdeck/work/aep-understanding-generation/current.json", "flightdeck/work/aep-understanding-generation/coverage.json", "all")
	if err != nil {
		t.Fatal(err)
	}
	if check.Status != StatusPass {
		t.Fatalf("check after sync = %+v, want pass", check)
	}
}

func TestSyncCoverageBatchFromMatricesAcceptsAbsoluteCoveragePath(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageBatchFixture(t, root, 0)
	source := filepath.Join(root, filepath.FromSlash("flightdeck/work/aep-understanding-generation/coverage.json"))
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	absoluteCoverage := filepath.Join(root, "tmp", "coverage-candidate.json")
	if err := os.MkdirAll(filepath.Dir(absoluteCoverage), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absoluteCoverage, data, 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := SyncCoverageBatchFromMatrices(root, "flightdeck/work/aep-understanding-generation/current.json", absoluteCoverage, "all")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("report = %+v, want pass for absolute coverage path", report)
	}
}

func TestPlanCoverageBatchMatricesBuildsStableCommands(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageBatchFixture(t, root, 2)
	writeFile(t, root, "examples/recipes/glob-b.json", "{}\n")
	writeFile(t, root, "examples/recipes/glob-a.json", "{}\n")
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/current.json", map[string]any{
		"coverage_batches": []map[string]any{
			{
				"id": "all",
				"entries": []map[string]any{
					{
						"coverage_id":  "text-paths",
						"out":          "tmp/matrix/text-paths",
						"ledger":       "tmp/matrix/text-paths/ledger.md",
						"recipe_paths": []string{"examples/recipes/text-basic.json"},
					},
					{
						"coverage_id":       "text-glob",
						"out":               "tmp/matrix/text-glob",
						"ledger":            "tmp/matrix/text-glob/ledger.md",
						"recipe_glob":       "examples/recipes/glob-*.json",
						"allow_matrix_exit": []int{1},
					},
				},
			},
		},
	})

	plan, err := PlanCoverageBatchMatrices(root, "flightdeck/work/aep-understanding-generation/current.json", "all")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Status != StatusPass || plan.Summary.Entries != 2 {
		t.Fatalf("plan = %+v, want two entries", plan)
	}
	first := plan.Entries[0]
	if first.CoverageID != "text-paths" || first.Out != "tmp/matrix/text-paths" || first.Ledger != "tmp/matrix/text-paths/ledger.md" {
		t.Fatalf("first entry = %+v", first)
	}
	assertStringSet(t, first.Args, []string{"run", "./cmd/aepmigrate", "matrix", "-recipe", "examples/recipes/text-basic.json", "-sources", "all", "-targets", "all", "-out", "tmp/matrix/text-paths", "-ledger-out", "tmp/matrix/text-paths/ledger.md"})
	second := plan.Entries[1]
	assertStringSet(t, second.Recipes, []string{"examples/recipes/glob-a.json", "examples/recipes/glob-b.json"})
	if len(second.AllowExitCodes) != 1 || second.AllowExitCodes[0] != 1 {
		t.Fatalf("allow exit codes = %+v", second.AllowExitCodes)
	}
}

func writeCoverageBatchFixture(t *testing.T, root string, coverageTotal int) {
	t.Helper()
	writeFile(t, root, "examples/recipes/text-basic.json", "{}\n")
	writeMatrixFixtureWithSummary(t, root, "tmp/matrix/text/matrix.json", map[string]any{
		"total": 2, "passed": 2, "blocked": 0, "failed": 0, "skipped": 0,
	})
	writeFile(t, root, "tmp/matrix/text/ledger.md", "# ledger\n")
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":       "text",
				"artifact": "tmp/matrix/text/matrix.json",
				"ledger":   "tmp/matrix/text/ledger.md",
				"recipes":  []string{"text-basic"},
				"totals":   map[string]any{"total": coverageTotal, "pass": coverageTotal, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/current.json", map[string]any{
		"truth_sources": map[string]any{
			"current":  "flightdeck/work/aep-understanding-generation/current.json",
			"coverage": "flightdeck/work/aep-understanding-generation/coverage.json",
		},
		"coverage_batches": []map[string]any{
			{
				"id":          "all",
				"description": "all current coverage",
				"entries": []map[string]any{
					{
						"coverage_id":  "text",
						"out":          "tmp/matrix/text",
						"matrix":       "tmp/matrix/text/matrix.json",
						"ledger":       "tmp/matrix/text/ledger.md",
						"recipe_paths": []string{"examples/recipes/text-basic.json"},
					},
				},
			},
		},
	})
}

func assertCoverageBatchIssue(t *testing.T, report CoverageBatchReport, code, coverageID, path string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Code == code && issue.CoverageID == coverageID && issue.Path == path {
			return
		}
	}
	t.Fatalf("issue %q/%q/%q not found in %+v", code, coverageID, path, report.Issues)
}
