package registry

import (
	"path/filepath"
	"testing"
)

func TestValidateCoveragePassesWhenArtifactTotalsMatch(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 2)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)

	report, err := ValidateCoverage(root, "flightdeck/work/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	if report.Summary.Artifacts != 1 || report.Summary.Errors != 0 {
		t.Fatalf("summary = %+v, want 1 artifact and 0 errors", report.Summary)
	}
}

func TestValidateCoverageFailsWhenArtifactTotalsDrift(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 3)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)

	report, err := ValidateCoverage(root, "flightdeck/work/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "matrix_totals_mismatch", "text", "tmp/matrix/text/matrix.json")
}

func writeCoverageFixture(t *testing.T, root, rel string, total int) {
	t.Helper()
	writeJSON(t, root, rel, map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":       "text",
				"artifact": "tmp/matrix/text/matrix.json",
				"totals": map[string]any{
					"total":   total,
					"pass":    total,
					"blocked": 0,
					"failed":  0,
					"skipped": 0,
				},
			},
		},
	})
}

func writeMatrixFixture(t *testing.T, root, rel string, total int) {
	t.Helper()
	writeJSON(t, root, rel, map[string]any{
		"schema_version": 1,
		"out_root":       filepath.Dir(rel),
		"summary": map[string]any{
			"total":   total,
			"passed":  total,
			"blocked": 0,
			"failed":  0,
			"skipped": 0,
		},
		"cases": []map[string]any{},
	})
}

func assertCoverageIssue(t *testing.T, report CoverageReport, code, recordID, path string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Code == code && issue.RecordID == recordID && issue.Path == path {
			return
		}
	}
	t.Fatalf("issue %q/%q/%q not found in %+v", code, recordID, path, report.Issues)
}
