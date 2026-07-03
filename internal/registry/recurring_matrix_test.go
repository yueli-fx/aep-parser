package registry

import "testing"

func TestCheckRecurringMatrixGatesPassesForExpectedArtifacts(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeMatrixFixtureWithSummary(t, root, "tmp/migration_matrix_verify/smoke_all/matrix.json", map[string]any{
		"total": 906, "passed": 900, "blocked": 0, "failed": 0, "skipped": 6,
	})
	writeJSON(t, root, "tmp/migration_matrix_verify/smoke_all/minimal-adjustment-layer/AE2020_to_AE2025/verify_report.json", map[string]any{
		"summary":      map[string]any{"status": "pass"},
		"verification": map[string]any{"profile_diff_status": "pass", "profile_diff_count": 0},
	})
	writeMatrixFixtureWithSummary(t, root, "tmp/migration_matrix_verify/explicit_matte_ae2025/matrix.json", map[string]any{
		"total": 1, "passed": 1, "blocked": 0, "failed": 0, "skipped": 0,
	})

	report, err := CheckRecurringMatrixGates(root, "tmp/migration_matrix_verify")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass || report.Summary.Errors != 0 || report.Summary.Gates != 3 {
		t.Fatalf("report = %+v, want three passing gates", report)
	}
}

func TestCheckRecurringMatrixGatesReportsDrift(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeMatrixFixtureWithSummary(t, root, "tmp/migration_matrix_verify/smoke_all/matrix.json", map[string]any{
		"total": 906, "passed": 899, "blocked": 0, "failed": 1, "skipped": 6,
	})
	writeJSON(t, root, "tmp/migration_matrix_verify/smoke_all/minimal-adjustment-layer/AE2020_to_AE2025/verify_report.json", map[string]any{
		"summary":      map[string]any{"status": "pass"},
		"verification": map[string]any{"profile_diff_status": "pass", "profile_diff_count": 0},
	})
	writeMatrixFixtureWithSummary(t, root, "tmp/migration_matrix_verify/explicit_matte_ae2025/matrix.json", map[string]any{
		"total": 1, "passed": 1, "blocked": 0, "failed": 0, "skipped": 0,
	})

	report, err := CheckRecurringMatrixGates(root, "tmp/migration_matrix_verify")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail || report.Summary.Errors == 0 {
		t.Fatalf("report = %+v, want failing drift report", report)
	}
}

func writeMatrixFixtureWithSummary(t *testing.T, root, rel string, summary map[string]any) {
	t.Helper()
	writeJSON(t, root, rel, map[string]any{
		"schema_version": 1,
		"summary":        summary,
		"cases":          []map[string]any{},
	})
}
