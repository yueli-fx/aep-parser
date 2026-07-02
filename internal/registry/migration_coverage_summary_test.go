package registry

import "testing"

func TestBuildMigrationCoverageSummaryIndexesRecipesDomainsAndHostEvidence(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", migrationCoverageSummaryFixture())
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2025", "status": "blocked"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-b", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/host/text/matrix.json", []map[string]any{
		{"recipe_name": "minimal-text-a", "source_version": "AE2020", "target_version": "AE2020", "ae_open_version": "AE2020", "status": "pass"},
		{"recipe_name": "minimal-text-a", "source_version": "AE2025", "target_version": "AE2025", "ae_open_version": "AE2025", "status": "pass"},
	})
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/shape/matrix.json", []map[string]any{
		{"recipe_name": "minimal-shape-a", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
	})

	summary, err := BuildMigrationCoverageSummary(root, "flightdeck/work/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if summary.SchemaVersion != 1 || summary.GeneratedFrom != "flightdeck/work/aep-understanding-generation/coverage.json" {
		t.Fatalf("summary identity = %d/%q", summary.SchemaVersion, summary.GeneratedFrom)
	}
	if summary.Totals.CoverageRecords != 2 || summary.Totals.Domains != 2 || summary.Totals.Recipes != 3 {
		t.Fatalf("totals = %+v, want two records, two domains, three recipes", summary.Totals)
	}
	if summary.Totals.WriterCases.Total != 5 || summary.Totals.WriterCases.Pass != 4 || summary.Totals.WriterCases.Blocked != 1 {
		t.Fatalf("writer totals = %+v", summary.Totals.WriterCases)
	}
	if summary.Totals.EndpointHostOpenCases.Total != 2 || summary.Totals.EndpointHostOpenCases.Pass != 2 {
		t.Fatalf("host totals = %+v", summary.Totals.EndpointHostOpenCases)
	}
	if summary.Totals.HostOpenEvidenceLevels.DirectEndpointHosts != 2 || summary.Totals.HostOpenEvidenceLevels.RepresentativeOnly != 1 {
		t.Fatalf("host evidence totals = %+v", summary.Totals.HostOpenEvidenceLevels)
	}
	textDomain := findMigrationDomain(t, summary.DomainRollup, "text")
	if textDomain.RecipeCount != 2 || textDomain.WriterTotals.Total != 4 || textDomain.HostOpenEvidenceLevels.DirectEndpointHosts != 2 {
		t.Fatalf("text domain = %+v", textDomain)
	}
	textRecipe := findMigrationRecipe(t, summary.RecipeIndex, "minimal-text-a")
	if textRecipe.BoundaryStatus != "none" {
		t.Fatalf("text recipe boundary status = %q, want none", textRecipe.BoundaryStatus)
	}
	if textRecipe.WriterMatrix.Artifact != "tmp/matrix/text/matrix.json" || textRecipe.WriterMatrix.Totals.Total != 2 || textRecipe.WriterMatrix.Totals.Blocked != 1 {
		t.Fatalf("text recipe writer matrix = %+v", textRecipe.WriterMatrix)
	}
	if textRecipe.HostOpenEvidence.EvidenceLevel != "direct_endpoint_hosts" || textRecipe.HostOpenEvidence.Matrix.Totals.Total != 2 {
		t.Fatalf("text recipe host evidence = %+v", textRecipe.HostOpenEvidence)
	}
	shapeRecipe := findMigrationRecipe(t, summary.RecipeIndex, "minimal-shape-a")
	if shapeRecipe.HostOpenEvidence.EvidenceLevel != "representative_only" {
		t.Fatalf("shape host evidence = %+v", shapeRecipe.HostOpenEvidence)
	}
}

func TestCheckMigrationCoverageSummaryReportsDrift(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", migrationCoverageSummaryFixture())
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/text/matrix.json", nil)
	writeMatrixFixtureWithCases(t, root, "tmp/host/text/matrix.json", nil)
	writeMatrixFixtureWithCases(t, root, "tmp/matrix/shape/matrix.json", nil)
	summary, err := BuildMigrationCoverageSummary(root, "flightdeck/work/aep-understanding-generation/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	summary.Totals.Recipes = 99
	writeJSON(t, root, "tmp/migration_coverage_summary.json", summary)

	report, err := CheckMigrationCoverageSummary(root, "flightdeck/work/aep-understanding-generation/coverage.json", "tmp/migration_coverage_summary.json")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail || report.Errors == 0 {
		t.Fatalf("check report = %+v, want failing drift report", report)
	}
}

func migrationCoverageSummaryFixture() map[string]any {
	return map[string]any{
		"schema_version":   1,
		"generated_at":     "2026-07-03T00:00:00Z",
		"writer_axes":      map[string]any{"source_writers": []string{"AE2020", "AE2025"}, "target_writers": []string{"AE2020", "AE2025"}},
		"host_open_axis":   map[string]any{"hosts": []string{"AE2020", "AE2025"}},
		"host_open_policy": map[string]any{"evidence_levels": []string{"direct_endpoint_hosts", "representative_only", "representative_covered", "excluded_known_boundary", "recorded_status_only"}},
		"recurring_gates":  []map[string]any{{"id": "gate", "artifact": "tmp/gate/matrix.json"}},
		"open_items":       []map[string]any{{"id": "follow-up"}},
		"coverage": []map[string]any{
			{
				"id":               "text",
				"domain":           "text",
				"scope":            "text scope",
				"artifact":         "tmp/matrix/text/matrix.json",
				"recipes":          []string{"minimal-text-a", "minimal-text-b"},
				"writer_status":    "PD-2x2",
				"writer_coverage":  "full",
				"totals":           map[string]any{"total": 4, "pass": 3, "blocked": 1, "failed": 0, "skipped": 0},
				"host_open_status": "direct_endpoint_hosts_pass",
				"host_open_endpoint_evidence": map[string]any{
					"open_mode":      "target_bound",
					"direct_hosts":   []string{"AE2020", "AE2025"},
					"inferred_hosts": []string{"AE2021", "AE2022", "AE2023", "AE2024"},
					"totals":         map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
					"artifact":       "tmp/host/text/matrix.json",
				},
			},
			{
				"id":                        "shape",
				"domain":                    "shape",
				"scope":                     "shape scope",
				"artifact":                  "tmp/matrix/shape/matrix.json",
				"recipes":                   []string{"minimal-shape-a"},
				"writer_status":             "PD-1x1",
				"writer_coverage":           "full",
				"totals":                    map[string]any{"total": 1, "pass": 1, "blocked": 0, "failed": 0, "skipped": 0},
				"host_open_status":          "OPEN-ALL-HOSTS representative",
				"host_open_representatives": []string{"minimal-shape-a"},
			},
		},
	}
}

func findMigrationDomain(t *testing.T, domains []MigrationCoverageDomainRollup, name string) MigrationCoverageDomainRollup {
	t.Helper()
	for _, domain := range domains {
		if domain.Domain == name {
			return domain
		}
	}
	t.Fatalf("domain %q not found in %+v", name, domains)
	return MigrationCoverageDomainRollup{}
}

func findMigrationRecipe(t *testing.T, recipes []MigrationCoverageRecipeIndexEntry, name string) MigrationCoverageRecipeIndexEntry {
	t.Helper()
	for _, recipe := range recipes {
		if recipe.Recipe == name {
			return recipe
		}
	}
	t.Fatalf("recipe %q not found in %+v", name, recipes)
	return MigrationCoverageRecipeIndexEntry{}
}
