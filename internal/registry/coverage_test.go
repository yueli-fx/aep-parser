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

func TestValidateCoverageReportsRecordRecipeVersionAndAtomLinks(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 4)
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version": 1,
		"capability_atoms": []map[string]any{
			{
				"id":           "text.a",
				"domain":       "text",
				"tier":         "atom",
				"status":       "verified",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-text-a.json", "required": true}},
			},
			{
				"id":           "text.b",
				"domain":       "text",
				"tier":         "atom",
				"status":       "verified",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "evidence", "path": "tmp/matrix/text/matrix.json", "required": true}},
			},
		},
	})

	report, err := ValidateCoverage(root, "flightdeck/work/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	if report.Summary.Atoms != 2 {
		t.Fatalf("atoms = %d, want 2", report.Summary.Atoms)
	}
	if len(report.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(report.Records))
	}
	record := report.Records[0]
	assertStringSet(t, record.ObservedRecipes, []string{"minimal-text-a", "minimal-text-b"})
	assertStringSet(t, record.SourceVersions, []string{"AE2020", "AE2025"})
	assertStringSet(t, record.TargetVersions, []string{"AE2020", "AE2025"})
	assertStringSet(t, record.AtomIDs, []string{"text.a", "text.b"})
	if record.Artifact != "tmp/matrix/text/matrix.json" {
		t.Fatalf("artifact = %q", record.Artifact)
	}
}

func TestValidateCoverageFailsWhenDeclaredRecipesDoNotMatchMatrix(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 1)
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-c", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
	})

	report, err := ValidateCoverage(root, "flightdeck/work/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "matrix_missing_declared_recipe", "text", "tmp/matrix/text/matrix.json")
	assertCoverageIssue(t, report, "matrix_unexpected_recipe", "text", "tmp/matrix/text/matrix.json")
}

func writeCoverageFixture(t *testing.T, root, rel string, total int) {
	t.Helper()
	writeJSON(t, root, rel, map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":       "text",
				"artifact": "tmp/matrix/text/matrix.json",
				"recipes":  []string{"minimal-text-a", "minimal-text-b"},
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
	cases := make([]map[string]any, 0, total)
	for i := 0; i < total; i++ {
		recipeName := "minimal-text-a"
		if i%2 == 1 {
			recipeName = "minimal-text-b"
		}
		cases = append(cases, map[string]any{
			"recipe_name":    recipeName,
			"source_version": "AE2020",
			"target_version": "AE2020",
			"status":         "pass",
		})
	}
	writeMatrixFixtureWithCases(t, root, rel, cases)
}

func writeMatrixFixtureWithCases(t *testing.T, root, rel string, cases []map[string]any) {
	t.Helper()
	total := len(cases)
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
		"cases": cases,
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

func assertStringSet(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("set = %+v, want %+v", got, want)
	}
	seen := map[string]bool{}
	for _, value := range got {
		seen[value] = true
	}
	for _, value := range want {
		if !seen[value] {
			t.Fatalf("set = %+v, want %+v", got, want)
		}
	}
}
