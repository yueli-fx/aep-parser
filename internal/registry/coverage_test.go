package registry

import (
	"path/filepath"
	"testing"
)

func TestValidateCoveragePassesWhenArtifactTotalsMatch(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "registry/workflows/aep-understanding-generation/coverage.json", 2)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
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
	writeCoverageFixture(t, root, "registry/workflows/aep-understanding-generation/coverage.json", 3)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "matrix_totals_mismatch", "text", "tmp/matrix/text/matrix.json")
}

func TestValidateCoverageChecksContractGates(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "registry/workflows/aep-understanding-generation/coverage.json", 2)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)
	writeBoundaryGateReport(t, root, VersionBoundaryCheckSummary{Boundaries: 1, Matched: 1, CheckedCells: 3})
	addBoundaryContractGate(t, root, VersionBoundaryCheckSummary{Boundaries: 1, Matched: 1, CheckedCells: 3})

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	if report.Summary.ContractGates != 1 {
		t.Fatalf("contract gates = %d, want 1", report.Summary.ContractGates)
	}
}

func TestValidateCoverageReportsContractGateDrift(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "registry/workflows/aep-understanding-generation/coverage.json", 2)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)
	writeBoundaryGateReport(t, root, VersionBoundaryCheckSummary{Boundaries: 1, Matched: 1, CheckedCells: 2})
	addBoundaryContractGate(t, root, VersionBoundaryCheckSummary{Boundaries: 1, Matched: 1, CheckedCells: 3})

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "contract_gate_summary_mismatch", "registry-version-boundaries", "tmp/registry_version_boundaries.json")
}

func TestValidateCoverageReportsRecordRecipeVersionAndAtomLinks(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "registry/workflows/aep-understanding-generation/coverage.json", 4)
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

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
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

func TestValidateCoverageReportsAtomRowsWithHostOpenEvidenceLabels(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
		"coverage": []map[string]any{
			{
				"id":                        "text",
				"artifact":                  "tmp/matrix/text/matrix.json",
				"recipes":                   []string{"minimal-text-a", "minimal-text-b"},
				"writer_status":             "PD-6x6",
				"host_open_representatives": []string{"minimal-text-a"},
				"host_open_endpoint_evidence": map[string]any{
					"direct_hosts":   []string{"AE2020", "AE2025"},
					"inferred_hosts": []string{"AE2021", "AE2022", "AE2023", "AE2024"},
					"recipes":        []string{"minimal-text-b"},
					"artifact":       "tmp/matrix/text-host/matrix.json",
					"totals":         map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
				},
				"totals": map[string]any{"total": 4, "pass": 4, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-a", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text-host/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-b", "source_version": "AE2020", "target_version": "AE2020", "ae_open_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2025", "ae_open_version": "AE2025", "status": "pass"},
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
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-text-b.json", "required": true}},
			},
		},
	})

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	if len(report.AtomRows) != 2 {
		t.Fatalf("atom rows = %d, want 2: %+v", len(report.AtomRows), report.AtomRows)
	}
	rowA := findAtomCoverageRow(t, report, "text.a")
	if rowA.HostOpenEvidenceLevel != "direct_all_hosts_representative" {
		t.Fatalf("rowA host level = %q", rowA.HostOpenEvidenceLevel)
	}
	assertStringSet(t, rowA.SourceVersions, []string{"AE2020", "AE2025"})
	assertStringSet(t, rowA.TargetVersions, []string{"AE2020", "AE2025"})
	if rowA.WriterStatus != "PD-6x6" || rowA.Totals.Pass != 2 {
		t.Fatalf("rowA writer/totals = %q/%+v", rowA.WriterStatus, rowA.Totals)
	}

	rowB := findAtomCoverageRow(t, report, "text.b")
	if rowB.HostOpenEvidenceLevel != "direct_endpoint_hosts_pass" {
		t.Fatalf("rowB host level = %q", rowB.HostOpenEvidenceLevel)
	}
	assertStringSet(t, rowB.DirectHostVersions, []string{"AE2020", "AE2025"})
	assertStringSet(t, rowB.InferredHostVersions, []string{"AE2021", "AE2022", "AE2023", "AE2024"})
}

func TestValidateCoverageReportsAtomRowBoundaryLabels(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
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
							"blocked": 5,
							"skipped": 30,
						},
					},
				},
				"totals": map[string]any{"total": 3, "pass": 1, "blocked": 1, "failed": 0, "skipped": 1},
			},
		},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/layer/matrix.json", []map[string]any{
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2024", "status": "blocked"},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2020", "target_version": "AE2020", "status": "skipped"},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version": 1,
		"capability_atoms": []map[string]any{
			{
				"id":           "layer.track_matte.explicit_source",
				"domain":       "layer",
				"tier":         "atom",
				"status":       "boundary",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-layer-explicit-matte.json", "required": true}},
			},
		},
	})

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	row := findAtomCoverageRow(t, report, "layer.track_matte.explicit_source")
	if row.BoundaryStatus != "known_matte_contract_boundary" {
		t.Fatalf("boundary status = %q", row.BoundaryStatus)
	}
	if row.BoundaryReason != "AE2025-only explicit matte source contract" {
		t.Fatalf("boundary reason = %q", row.BoundaryReason)
	}
	if row.HostOpenEvidenceLevel != "excluded_known_boundary" {
		t.Fatalf("host level = %q", row.HostOpenEvidenceLevel)
	}
}

func TestCoverageCellsReportsSourceTargetCasesWithBoundaryLabels(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
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
							"recipe": "minimal-layer-explicit-matte",
							"reason": "AE2025-only explicit matte source contract",
						},
					},
				},
				"totals": map[string]any{"total": 3, "pass": 1, "blocked": 1, "failed": 0, "skipped": 1},
			},
		},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/layer/matrix.json", []map[string]any{
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2024", "status": "blocked", "reason": "cannot downgrade explicit matte source"},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2020", "target_version": "AE2020", "status": "skipped", "reason": "source contract unavailable before AE2025"},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version": 1,
		"capability_atoms": []map[string]any{
			{
				"id":           "layer.track_matte.explicit_source",
				"domain":       "layer",
				"tier":         "atom",
				"status":       "boundary",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-layer-explicit-matte.json", "required": true}},
			},
		},
	})

	report, err := CoverageCells(root, "registry/workflows/aep-understanding-generation/coverage.json", CoverageCellFilter{
		AtomID: "layer.track_matte.explicit_source",
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass || report.Count != 3 {
		t.Fatalf("cells status/count = %q/%d, want pass/3", report.Status, report.Count)
	}
	blocked := findCoverageCell(t, report.Cells, "blocked")
	if blocked.Domain != "layer" || blocked.RecordID != "layer" || blocked.Recipe != "minimal-layer-explicit-matte" {
		t.Fatalf("blocked cell identity = %+v", blocked)
	}
	if blocked.BoundaryStatus != "known_matte_contract_boundary" || blocked.BoundaryReason != "AE2025-only explicit matte source contract" {
		t.Fatalf("blocked boundary labels = %+v", blocked)
	}
	if blocked.SourceVersion != "AE2025" || blocked.TargetVersion != "AE2024" {
		t.Fatalf("blocked versions = %+v", blocked)
	}
	if blocked.Reason != "cannot downgrade explicit matte source" {
		t.Fatalf("blocked reason = %q", blocked.Reason)
	}

	filtered, err := CoverageCells(root, "registry/workflows/aep-understanding-generation/coverage.json", CoverageCellFilter{
		AtomID:     "layer.track_matte.explicit_source",
		CaseStatus: "skipped",
	})
	if err != nil {
		t.Fatal(err)
	}
	if filtered.Count != 1 || filtered.Cells[0].Status != "skipped" {
		t.Fatalf("filtered cells = %+v, want one skipped cell", filtered.Cells)
	}

	domainFiltered, err := CoverageCells(root, "registry/workflows/aep-understanding-generation/coverage.json", CoverageCellFilter{
		Domain: "layer",
	})
	if err != nil {
		t.Fatal(err)
	}
	if domainFiltered.Count != 3 || domainFiltered.Filter.Domain != "layer" {
		t.Fatalf("domain filtered cells = count %d filter=%+v", domainFiltered.Count, domainFiltered.Filter)
	}
}

func TestSummarizeCoverageGroupsAtomRowsForQuickQueries(t *testing.T) {
	report := CoverageReport{
		Status: StatusPass,
		Summary: CoverageSummary{
			Records:   2,
			Artifacts: 3,
			Atoms:     3,
		},
		AtomRows: []AtomCoverageRow{
			{
				AtomID:                "text.a",
				Domain:                "text",
				RecordID:              "text",
				Recipe:                "minimal-text-a",
				WriterStatus:          "PD-6x6",
				SourceVersions:        []string{"AE2020", "AE2025"},
				TargetVersions:        []string{"AE2020", "AE2025"},
				HostOpenEvidenceLevel: "direct_endpoint_hosts_pass",
				DirectHostVersions:    []string{"AE2020", "AE2025"},
				InferredHostVersions:  []string{"AE2021", "AE2022", "AE2023", "AE2024"},
			},
			{
				AtomID:                "text.b",
				Domain:                "text",
				RecordID:              "text",
				Recipe:                "minimal-text-b",
				WriterStatus:          "PD-6x6",
				SourceVersions:        []string{"AE2020", "AE2025"},
				TargetVersions:        []string{"AE2020", "AE2025"},
				HostOpenEvidenceLevel: "direct_all_hosts_representative",
			},
			{
				AtomID:                "layer.boundary",
				Domain:                "layer",
				RecordID:              "layer",
				Recipe:                "minimal-layer-boundary",
				WriterStatus:          "boundary",
				BoundaryStatus:        "known_matte_contract_boundary",
				HostOpenEvidenceLevel: "excluded_known_boundary",
			},
		},
	}

	summary := SummarizeCoverage(report)
	if summary.Status != StatusPass || summary.AtomRows != 3 || summary.Atoms != 3 {
		t.Fatalf("summary status/rows/atoms = %q/%d/%d", summary.Status, summary.AtomRows, summary.Atoms)
	}
	if summary.Summary.AtomRows != 3 || summary.Summary.Atoms != 3 || summary.Summary.DirectHostAtoms != 1 || summary.Summary.InferredHostAtoms != 1 {
		t.Fatalf("nested summary = %+v, want scalar totals mirrored for machine-readable total queries", summary.Summary)
	}
	assertCoverageSummaryBucket(t, summary.ByRecord, "text", 2)
	assertCoverageSummaryBucket(t, summary.ByRecord, "layer", 1)
	assertCoverageSummaryBucket(t, summary.ByDomain, "text", 2)
	assertCoverageSummaryBucket(t, summary.ByDomain, "layer", 1)
	assertCoverageSummaryBucket(t, summary.ByWriterStatus, "PD-6x6", 2)
	assertCoverageSummaryBucket(t, summary.ByWriterStatus, "boundary", 1)
	assertCoverageSummaryBucket(t, summary.ByHostOpenEvidenceLevel, "direct_endpoint_hosts_pass", 1)
	assertCoverageSummaryBucket(t, summary.ByHostOpenEvidenceLevel, "direct_all_hosts_representative", 1)
	assertCoverageSummaryBucket(t, summary.ByHostOpenEvidenceLevel, "excluded_known_boundary", 1)
	assertCoverageSummaryBucket(t, summary.ByBoundaryStatus, "known_matte_contract_boundary", 1)
	if summary.DirectHostAtoms != 1 || summary.InferredHostAtoms != 1 {
		t.Fatalf("host atoms = direct %d inferred %d", summary.DirectHostAtoms, summary.InferredHostAtoms)
	}
}

func TestSummarizeCoverageReportsRecipeAtomCoverageGaps(t *testing.T) {
	report := CoverageReport{
		Status: StatusPass,
		Summary: CoverageSummary{
			Records:   1,
			Artifacts: 1,
			Atoms:     2,
		},
		Records: []CoverageRecordReport{
			{
				ID:              "text",
				DeclaredRecipes: []string{"minimal-text-a", "minimal-text-b", "minimal-text-c"},
				ObservedRecipes: []string{"minimal-text-a", "minimal-text-b", "minimal-text-c"},
			},
		},
		AtomRows: []AtomCoverageRow{
			{AtomID: "text.a", RecordID: "text", Recipe: "minimal-text-a"},
			{AtomID: "text.b", RecordID: "text", Recipe: "minimal-text-b"},
			{AtomID: "text.template", RecordID: "text"},
		},
	}

	summary := SummarizeCoverage(report)
	if summary.Summary.DeclaredRecipes != 3 || summary.Summary.ObservedRecipes != 3 || summary.Summary.AtomRowRecipes != 2 {
		t.Fatalf("recipe totals = %+v, want declared/observed/atom-row recipes 3/3/2", summary.Summary)
	}
	if summary.Summary.RecipesWithoutAtomRows != 1 || len(summary.RecipesWithoutAtomRows) != 1 || summary.RecipesWithoutAtomRows[0] != "minimal-text-c" {
		t.Fatalf("recipes without atom rows = %d/%v, want minimal-text-c", summary.Summary.RecipesWithoutAtomRows, summary.RecipesWithoutAtomRows)
	}
	if summary.Summary.AtomRowsWithoutRecipes != 1 || len(summary.AtomRowsWithoutRecipes) != 1 || summary.AtomRowsWithoutRecipes[0] != "text.template" {
		t.Fatalf("atom rows without recipes = %d/%v, want text.template", summary.Summary.AtomRowsWithoutRecipes, summary.AtomRowsWithoutRecipes)
	}
}

func TestSummarizeCoverageReportsObservedRecipesWithoutDeclaration(t *testing.T) {
	report := CoverageReport{
		Status: StatusPass,
		Summary: CoverageSummary{
			Records:   1,
			Artifacts: 1,
			Atoms:     3,
		},
		Records: []CoverageRecordReport{
			{
				ID:              "shape",
				DeclaredRecipes: []string{"minimal-shape-a", "minimal-shape-b"},
				ObservedRecipes: []string{"minimal-shape-a", "minimal-shape-b", "minimal-shape-c"},
			},
		},
		AtomRows: []AtomCoverageRow{
			{AtomID: "shape.a", RecordID: "shape", Recipe: "minimal-shape-a"},
			{AtomID: "shape.b", RecordID: "shape", Recipe: "minimal-shape-b"},
			{AtomID: "shape.c", RecordID: "shape", Recipe: "minimal-shape-c"},
		},
	}

	summary := SummarizeCoverage(report)
	if summary.Summary.ObservedRecipesWithoutDeclaration != 1 || summary.Summary.RecordsWithUndeclaredRecipes != 1 {
		t.Fatalf("undeclared recipe totals = %+v, want one undeclared recipe on one record", summary.Summary)
	}
	if len(summary.RecordsWithUndeclaredRecipes) != 1 || summary.RecordsWithUndeclaredRecipes[0].RecordID != "shape" {
		t.Fatalf("undeclared recipe records = %+v, want shape", summary.RecordsWithUndeclaredRecipes)
	}
	assertStringSet(t, summary.RecordsWithUndeclaredRecipes[0].Recipes, []string{"minimal-shape-c"})
}

func TestCoverageAxisSummarizesVersionAxisAndHostLabels(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
		"coverage": []map[string]any{
			{
				"id":                        "text",
				"artifact":                  "tmp/matrix/text/matrix.json",
				"recipes":                   []string{"minimal-text-a", "minimal-text-b"},
				"writer_status":             "PD-6x6",
				"host_open_representatives": []string{"minimal-text-a"},
				"host_open_endpoint_evidence": map[string]any{
					"direct_hosts":   []string{"AE2020", "AE2025"},
					"inferred_hosts": []string{"AE2021", "AE2022", "AE2023", "AE2024"},
					"recipes":        []string{"minimal-text-b"},
					"artifact":       "tmp/matrix/text-host/matrix.json",
					"totals":         map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
				},
				"totals": map[string]any{"total": 4, "pass": 4, "blocked": 0, "failed": 0, "skipped": 0},
			},
			{
				"id":            "layer",
				"artifact":      "tmp/matrix/layer/matrix.json",
				"recipes":       []string{"minimal-layer-explicit-matte"},
				"writer_status": "boundary",
				"boundary": map[string]any{
					"status":             "known_matte_contract_boundary",
					"blocked_recipe_ids": []string{"minimal-layer-explicit-matte"},
					"details": []map[string]any{
						{"recipe": "minimal-layer-explicit-matte", "reason": "AE2025-only explicit matte source contract"},
					},
				},
				"totals": map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0},
			},
		},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-a", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text-host/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-b", "source_version": "AE2020", "target_version": "AE2020", "ae_open_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2025", "ae_open_version": "AE2025", "status": "pass"},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/layer/matrix.json", []map[string]any{
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2024", "status": "blocked"},
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
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-text-b.json", "required": true}},
			},
			{
				"id":           "layer.track_matte.explicit_source",
				"domain":       "layer",
				"tier":         "atom",
				"status":       "boundary",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-layer-explicit-matte.json", "required": true}},
			},
		},
	})

	axis, err := CoverageAxis(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2020", "AE2025"})
	if err != nil {
		t.Fatal(err)
	}
	if axis.Status != StatusPass || axis.Summary.AtomRows != 3 {
		t.Fatalf("axis status/rows = %q/%d", axis.Status, axis.Summary.AtomRows)
	}
	if axis.Summary.WriterFullAxisRows != 2 || axis.Summary.BoundaryRows != 1 {
		t.Fatalf("writer summary = %+v", axis.Summary)
	}
	if axis.Summary.HostFullAxisRows != 2 || axis.Summary.HostMissingRows != 0 || axis.Summary.HostBoundaryRows != 1 {
		t.Fatalf("host summary = %+v", axis.Summary)
	}
	textA := findCoverageAxisRow(t, axis, "text.a")
	if textA.HostAxisStatus != "full_direct_all_hosts_axis" {
		t.Fatalf("text.a axis row = %+v", textA)
	}
	textB := findCoverageAxisRow(t, axis, "text.b")
	if textB.WriterAxisStatus != "full_source_target_axis" || textB.HostAxisStatus != "full_direct_or_inferred_axis" {
		t.Fatalf("text.b axis row = %+v", textB)
	}
	boundary := findCoverageAxisRow(t, axis, "layer.track_matte.explicit_source")
	if boundary.WriterAxisStatus != "boundary_source_contract" || boundary.HostAxisStatus != "excluded_known_boundary" {
		t.Fatalf("boundary axis row = %+v", boundary)
	}
	pair := findCoverageAxisPair(t, axis, "AE2025", "AE2024")
	if pair.Blocked != 1 {
		t.Fatalf("AE2025->AE2024 pair = %+v, want one blocked", pair)
	}
	if axis.Summary.SourceTargetPairs != 5 || axis.Summary.Cells != 6 {
		t.Fatalf("pair/cell summary = %+v", axis.Summary)
	}
	textDomain := findCoverageAxisDomain(t, axis, "text")
	if textDomain.AtomRows != 2 || textDomain.WriterFullAxisRows != 2 || textDomain.HostFullAxisRows != 2 || textDomain.Cells != 4 || textDomain.Pass != 4 {
		t.Fatalf("text domain summary = %+v", textDomain)
	}
	layerDomain := findCoverageAxisDomain(t, axis, "layer")
	if layerDomain.AtomRows != 1 || layerDomain.BoundaryRows != 1 || layerDomain.HostBoundaryRows != 1 || layerDomain.Cells != 2 || layerDomain.Blocked != 1 {
		t.Fatalf("layer domain summary = %+v", layerDomain)
	}
}

func TestCoverageAxisWithFilterNarrowsRowsAndCells(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
		"coverage": []map[string]any{
			{
				"id":            "layer",
				"artifact":      "tmp/matrix/layer/matrix.json",
				"recipes":       []string{"minimal-layer-explicit-matte", "minimal-layer-parent"},
				"writer_status": "boundary",
				"boundary": map[string]any{
					"status":             "known_matte_contract_boundary",
					"blocked_recipe_ids": []string{"minimal-layer-explicit-matte"},
					"details": []map[string]any{
						{"recipe": "minimal-layer-explicit-matte", "reason": "AE2025-only explicit matte source contract"},
					},
				},
				"totals": map[string]any{"total": 3, "pass": 2, "blocked": 1, "failed": 0, "skipped": 0},
			},
			{
				"id":               "text",
				"artifact":         "tmp/matrix/text/matrix.json",
				"recipes":          []string{"minimal-text-source"},
				"writer_status":    "PD-6x6",
				"host_open_status": "OPEN-ALL-HOSTS",
				"totals":           map[string]any{"total": 1, "pass": 1, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/layer/matrix.json", []map[string]any{
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		{"recipe_name": "minimal-layer-explicit-matte", "source_version": "AE2025", "target_version": "AE2024", "status": "blocked"},
		{"recipe_name": "minimal-layer-parent", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-source", "source_version": "AE2024", "target_version": "AE2024", "status": "pass"},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version": 1,
		"capability_atoms": []map[string]any{
			{
				"id":           "layer.track_matte.explicit_source",
				"domain":       "layer",
				"tier":         "atom",
				"status":       "boundary",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-layer-explicit-matte.json", "required": true}},
			},
			{
				"id":           "layer.parent",
				"domain":       "layer",
				"tier":         "atom",
				"status":       "verified",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-layer-parent.json", "required": true}},
			},
			{
				"id":           "text.source.default",
				"domain":       "text",
				"tier":         "atom",
				"status":       "verified",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/minimal-text-source.json", "required": true}},
			},
		},
	})

	axis, err := CoverageAxisWithFilter(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2024", "AE2025"}, CoverageAxisFilter{
		AtomID: "layer.track_matte.explicit_source",
	})
	if err != nil {
		t.Fatal(err)
	}
	if axis.Summary.AtomRows != 1 || len(axis.Rows) != 1 || axis.Rows[0].AtomID != "layer.track_matte.explicit_source" {
		t.Fatalf("filtered rows = %+v summary=%+v", axis.Rows, axis.Summary)
	}
	if axis.Summary.Cells != 2 || axis.Summary.Blocked != 1 || axis.Summary.Pass != 1 {
		t.Fatalf("filtered cell summary = %+v", axis.Summary)
	}
	if axis.Filter.AtomID != "layer.track_matte.explicit_source" {
		t.Fatalf("filter not preserved: %+v", axis.Filter)
	}

	blockedOnly, err := CoverageAxisWithFilter(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2024", "AE2025"}, CoverageAxisFilter{
		AtomID:     "layer.track_matte.explicit_source",
		CaseStatus: "blocked",
	})
	if err != nil {
		t.Fatal(err)
	}
	if blockedOnly.Summary.AtomRows != 1 || len(blockedOnly.Rows) != 1 || blockedOnly.Summary.Cells != 1 || blockedOnly.Summary.Blocked != 1 || blockedOnly.Summary.Pass != 0 {
		t.Fatalf("blocked-only summary = %+v", blockedOnly.Summary)
	}

	boundaryAxisOnly, err := CoverageAxisWithFilter(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2024", "AE2025"}, CoverageAxisFilter{
		WriterAxisStatus: "boundary_source_contract",
	})
	if err != nil {
		t.Fatal(err)
	}
	if boundaryAxisOnly.Summary.AtomRows != 1 || len(boundaryAxisOnly.Rows) != 1 || boundaryAxisOnly.Rows[0].AtomID != "layer.track_matte.explicit_source" {
		t.Fatalf("boundary-axis rows = %+v summary=%+v", boundaryAxisOnly.Rows, boundaryAxisOnly.Summary)
	}
	if boundaryAxisOnly.Summary.Cells != 2 || boundaryAxisOnly.Summary.Blocked != 1 || boundaryAxisOnly.Summary.Pass != 1 {
		t.Fatalf("boundary-axis cell summary = %+v", boundaryAxisOnly.Summary)
	}

	missingHostOnly, err := CoverageAxisWithFilter(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2024", "AE2025"}, CoverageAxisFilter{
		HostAxisStatus: "missing_host_axis",
	})
	if err != nil {
		t.Fatal(err)
	}
	if missingHostOnly.Summary.AtomRows != 1 || missingHostOnly.Summary.HostMissingRows != 1 {
		t.Fatalf("missing-host summary = %+v", missingHostOnly.Summary)
	}

	textOnly, err := CoverageAxisWithFilter(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2024", "AE2025"}, CoverageAxisFilter{
		Domain: "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	if textOnly.Summary.AtomRows != 1 || len(textOnly.Rows) != 1 || textOnly.Rows[0].Domain != "text" || textOnly.Rows[0].AtomID != "text.source.default" {
		t.Fatalf("text-domain rows = %+v summary=%+v", textOnly.Rows, textOnly.Summary)
	}
	if textOnly.Summary.Cells != 1 || textOnly.Summary.Pass != 1 || textOnly.Filter.Domain != "text" {
		t.Fatalf("text-domain summary = %+v filter=%+v", textOnly.Summary, textOnly.Filter)
	}

	textMissingTarget, err := CoverageAxisWithFilter(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2024", "AE2025"}, CoverageAxisFilter{
		Domain:               "text",
		MissingTargetVersion: "AE2025",
	})
	if err != nil {
		t.Fatal(err)
	}
	if textMissingTarget.Summary.AtomRows != 1 || len(textMissingTarget.Rows) != 1 || textMissingTarget.Rows[0].AtomID != "text.source.default" {
		t.Fatalf("text missing-target rows = %+v summary=%+v", textMissingTarget.Rows, textMissingTarget.Summary)
	}
	if textMissingTarget.Summary.Cells != 1 || textMissingTarget.Filter.MissingTargetVersion != "AE2025" {
		t.Fatalf("text missing-target summary = %+v filter=%+v", textMissingTarget.Summary, textMissingTarget.Filter)
	}

	missingHostAE2025, err := CoverageAxisWithFilter(root, "registry/workflows/aep-understanding-generation/coverage.json", []string{"AE2024", "AE2025"}, CoverageAxisFilter{
		MissingHostVersion: "AE2025",
	})
	if err != nil {
		t.Fatal(err)
	}
	if missingHostAE2025.Summary.AtomRows != 1 || len(missingHostAE2025.Rows) != 1 || missingHostAE2025.Rows[0].AtomID != "layer.parent" {
		t.Fatalf("missing-host AE2025 rows = %+v summary=%+v", missingHostAE2025.Rows, missingHostAE2025.Summary)
	}
	if missingHostAE2025.Summary.Cells != 1 || missingHostAE2025.Filter.MissingHostVersion != "AE2025" {
		t.Fatalf("missing-host AE2025 summary = %+v filter=%+v", missingHostAE2025.Summary, missingHostAE2025.Filter)
	}
}

func findCoverageCell(t *testing.T, cells []CoverageCell, status string) CoverageCell {
	t.Helper()
	for _, cell := range cells {
		if cell.Status == status {
			return cell
		}
	}
	t.Fatalf("coverage cell with status %q not found in %+v", status, cells)
	return CoverageCell{}
}

func findCoverageAxisRow(t *testing.T, report CoverageAxisReport, atomID string) CoverageAxisRow {
	t.Helper()
	for _, row := range report.Rows {
		if row.AtomID == atomID {
			return row
		}
	}
	t.Fatalf("axis row %q not found in %+v", atomID, report.Rows)
	return CoverageAxisRow{}
}

func findCoverageAxisDomain(t *testing.T, report CoverageAxisReport, domain string) CoverageAxisDomainSummary {
	t.Helper()
	for _, summary := range report.ByDomain {
		if summary.Domain == domain {
			return summary
		}
	}
	t.Fatalf("axis domain %q not found in %+v", domain, report.ByDomain)
	return CoverageAxisDomainSummary{}
}

func findCoverageAxisPair(t *testing.T, report CoverageAxisReport, source, target string) CoverageAxisPairSummary {
	t.Helper()
	for _, pair := range report.ByVersionPair {
		if pair.SourceVersion == source && pair.TargetVersion == target {
			return pair
		}
	}
	t.Fatalf("axis pair %s->%s not found in %+v", source, target, report.ByVersionPair)
	return CoverageAxisPairSummary{}
}

func TestFilterCoverageRowsSelectsAtomRowsForFocusedQueries(t *testing.T) {
	report := CoverageReport{
		AtomRows: []AtomCoverageRow{
			{
				AtomID:                "text.a",
				RecordID:              "text",
				WriterStatus:          "PD-6x6",
				HostOpenEvidenceLevel: "direct_endpoint_hosts_pass",
			},
			{
				AtomID:                "text.b",
				RecordID:              "text",
				WriterStatus:          "boundary",
				HostOpenEvidenceLevel: "excluded_known_boundary",
			},
			{
				AtomID:                "shape.a",
				RecordID:              "shape",
				WriterStatus:          "PD-6x6",
				HostOpenEvidenceLevel: "direct_all_hosts_representative",
			},
		},
	}

	rows := FilterCoverageRows(report, CoverageRowFilter{
		RecordID:              "text",
		WriterStatus:          "PD-6x6",
		HostOpenEvidenceLevel: "direct_endpoint_hosts_pass",
	})
	if len(rows) != 1 || rows[0].AtomID != "text.a" {
		t.Fatalf("rows = %+v, want text.a only", rows)
	}

	rows = FilterCoverageRows(report, CoverageRowFilter{AtomID: "shape.a"})
	if len(rows) != 1 || rows[0].RecordID != "shape" {
		t.Fatalf("rows = %+v, want shape.a only", rows)
	}
}

func TestValidateCoverageFailsWhenDeclaredRecipesDoNotMatchMatrix(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeCoverageFixture(t, root, "registry/workflows/aep-understanding-generation/coverage.json", 1)
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-c", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
	})

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "matrix_missing_declared_recipe", "text", "tmp/matrix/text/matrix.json")
	assertCoverageIssue(t, report, "matrix_unexpected_recipe", "text", "tmp/matrix/text/matrix.json")
}

func TestValidateCoverageRequiresHostOpenPolicy(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":       "text",
				"artifact": "tmp/matrix/text/matrix.json",
				"recipes":  []string{"minimal-text-a", "minimal-text-b"},
				"totals":   map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)

	report, err := ValidateCoverage(root, "registry/workflows/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "missing_host_open_policy", "", "registry/workflows/aep-understanding-generation/coverage.json")
}

func TestValidateCoverageReportsUncoveredRecipes(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "examples/recipes/minimal-text-a.json", "{}")
	writeFile(t, root, "examples/recipes/minimal-text-b.json", "{}")
	writeFile(t, root, "examples/recipes/minimal-text-c.json", "{}")
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
		"coverage": []map[string]any{
			{
				"id":       "text",
				"artifact": "tmp/matrix/text/matrix.json",
				"recipes":  []string{"minimal-text-a", "minimal-text-b"},
				"totals":   map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)

	report, err := ValidateCoverageWithOptions(root, "registry/workflows/aep-understanding-generation/coverage.json", CoverageValidationOptions{
		RequireAllRecipes: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "uncovered_recipe", "minimal-text-c", "examples/recipes/minimal-text-c.json")
}

func TestValidateCoverageCanRequireLedgers(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
		"coverage": []map[string]any{
			{
				"id":       "text",
				"artifact": "tmp/matrix/text/matrix.json",
				"ledger":   "tmp/matrix/text/ledger.md",
				"recipes":  []string{"minimal-text-a", "minimal-text-b"},
				"totals":   map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)

	report, err := ValidateCoverageWithOptions(root, "registry/workflows/aep-understanding-generation/coverage.json", CoverageValidationOptions{
		RequireLedgers: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertCoverageIssue(t, report, "missing_ledger", "text", "tmp/matrix/text/ledger.md")
}

func writeCoverageFixture(t *testing.T, root, rel string, total int) {
	t.Helper()
	writeJSON(t, root, rel, map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
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

func validHostOpenPolicyFixture() map[string]any {
	return map[string]any{
		"available_flags": []string{"-ae-open", "-ae-versions", "-max-ae-open-cases"},
		"evidence_levels": []string{
			"direct_all_hosts",
			"inferred_by_endpoint",
			"pending_per_capability",
			"direct_endpoint_hosts",
			"representative_only",
			"representative_covered",
			"excluded_known_boundary",
			"recorded_status_only",
		},
		"endpoint_inference": map[string]any{
			"enabled":        true,
			"open_mode":      "target_bound",
			"direct_hosts":   []string{"AE2020", "AE2025"},
			"inferred_hosts": []string{"AE2021", "AE2022", "AE2023", "AE2024"},
		},
	}
}

func addBoundaryContractGate(t *testing.T, root string, summary VersionBoundaryCheckSummary) {
	t.Helper()
	writeJSON(t, root, "registry/workflows/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version":   1,
		"host_open_policy": validHostOpenPolicyFixture(),
		"contract_gates": []map[string]any{
			{
				"id":       "registry-version-boundaries",
				"kind":     "version_boundaries",
				"command":  "go run ./cmd/aepregistry boundaries -root . -out tmp/registry_version_boundaries.json",
				"artifact": "tmp/registry_version_boundaries.json",
				"status":   StatusPass,
				"summary":  summary,
			},
		},
		"coverage": []map[string]any{
			{
				"id":       "text",
				"artifact": "tmp/matrix/text/matrix.json",
				"totals": map[string]any{
					"total":   2,
					"pass":    2,
					"blocked": 0,
					"failed":  0,
					"skipped": 0,
				},
			},
		},
	})
}

func writeBoundaryGateReport(t *testing.T, root string, summary VersionBoundaryCheckSummary) {
	t.Helper()
	writeJSON(t, root, "tmp/registry_version_boundaries.json", map[string]any{
		"schema_version": 1,
		"status":         StatusPass,
		"summary":        summary,
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
	summary := map[string]any{
		"total":   len(cases),
		"passed":  0,
		"blocked": 0,
		"failed":  0,
		"skipped": 0,
	}
	for _, c := range cases {
		switch c["status"] {
		case "blocked":
			summary["blocked"] = summary["blocked"].(int) + 1
		case "failed":
			summary["failed"] = summary["failed"].(int) + 1
		case "skipped":
			summary["skipped"] = summary["skipped"].(int) + 1
		default:
			summary["passed"] = summary["passed"].(int) + 1
		}
	}
	writeJSON(t, root, rel, map[string]any{
		"schema_version": 1,
		"out_root":       filepath.Dir(rel),
		"summary":        summary,
		"cases":          cases,
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

func findAtomCoverageRow(t *testing.T, report CoverageReport, atomID string) AtomCoverageRow {
	t.Helper()
	for _, row := range report.AtomRows {
		if row.AtomID == atomID {
			return row
		}
	}
	t.Fatalf("atom row %q not found in %+v", atomID, report.AtomRows)
	return AtomCoverageRow{}
}

func assertCoverageSummaryBucket(t *testing.T, buckets []CoverageSummaryBucket, name string, atomRows int) {
	t.Helper()
	for _, bucket := range buckets {
		if bucket.Name == name {
			if bucket.AtomRows != atomRows {
				t.Fatalf("bucket %q atom rows = %d, want %d", name, bucket.AtomRows, atomRows)
			}
			return
		}
	}
	t.Fatalf("bucket %q not found in %+v", name, buckets)
}
