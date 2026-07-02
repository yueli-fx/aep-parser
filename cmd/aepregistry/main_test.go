package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/registry"
)

func TestRunAuditWritesReportAndReturnsZero(t *testing.T) {
	root := newRegistryRoot(t)
	out := filepath.Join(root, "tmp", "registry_audit.json")

	code := run([]string{"audit", "-root", root, "-out", out})
	if code != 0 {
		t.Fatalf("run(audit) = %d, want 0", code)
	}

	report := readReport(t, out)
	if report.Status != registry.StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, registry.StatusPass, report.Issues)
	}
}

func TestRunAuditReturnsOneForRegistryErrors(t *testing.T) {
	root := newRegistryRoot(t)
	out := filepath.Join(root, "tmp", "registry_audit.json")
	if err := os.Remove(filepath.Join(root, "examples", "recipes", "text-basic.json")); err != nil {
		t.Fatal(err)
	}

	code := run([]string{"audit", "-root", root, "-out", out})
	if code != 1 {
		t.Fatalf("run(audit) = %d, want 1", code)
	}

	report := readReport(t, out)
	if report.Status != registry.StatusFail || report.Summary.Errors == 0 {
		t.Fatalf("status/errors = %q/%d, want fail/>0", report.Status, report.Summary.Errors)
	}
}

func TestRunInventoryWritesLocationSummary(t *testing.T) {
	root := newRegistryRoot(t)
	out := filepath.Join(root, "tmp", "registry_inventory.json")

	code := run([]string{"inventory", "-root", root, "-out", out})
	if code != 0 {
		t.Fatalf("run(inventory) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.InventoryReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Locations == 0 || len(report.Locations) == 0 {
		t.Fatalf("empty inventory report: %+v", report)
	}
	if !hasInventoryLocation(report, "recipes") {
		t.Fatalf("recipes location missing from inventory: %+v", report.Locations)
	}
}

func TestRunOwnershipWritesOwnershipSummary(t *testing.T) {
	root := newRegistryRoot(t)
	writeFile(t, root, "examples/recipes/unowned.json", "{}\n")
	out := filepath.Join(root, "tmp", "registry_ownership.json")

	code := run([]string{"ownership", "-root", root, "-out", out, "-sample-limit", "2"})
	if code != 0 {
		t.Fatalf("run(ownership) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.OwnershipReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Locations == 0 || len(report.Locations) == 0 {
		t.Fatalf("empty ownership report: %+v", report)
	}
	if !hasOwnershipLocation(report, "recipes") {
		t.Fatalf("recipes location missing from ownership: %+v", report.Locations)
	}
	if report.Summary.UnownedFiles == 0 {
		t.Fatalf("ownership report should expose unowned files: %+v", report.Summary)
	}
}

func TestRunLayoutWritesCleanupGuardrailReport(t *testing.T) {
	root := newRegistryRoot(t)
	writeFile(t, root, "examples/recipes/unowned.json", "{}\n")
	writeFile(t, root, "tmp/orphan/matrix.json", "{}\n")
	out := filepath.Join(root, "tmp", "registry_layout.json")

	code := run([]string{"layout", "-root", root, "-out", out, "-sample-limit", "1"})
	if code != 0 {
		t.Fatalf("run(layout) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.LayoutReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Locations == 0 || report.Summary.CleanupCandidateFiles == 0 || report.Summary.BlockedUnownedFiles == 0 {
		t.Fatalf("layout summary = %+v, want cleanup candidates and blocked unowned files", report.Summary)
	}
}

func TestRunCleanupWritesGeneratedCleanupReport(t *testing.T) {
	root := newRegistryRoot(t)
	writeFile(t, root, "tmp/orphan/matrix.json", "{}\n")
	out := filepath.Join(root, "tmp", "registry_generated_cleanup.json")

	code := run([]string{"cleanup", "-root", root, "-out", out, "-sample-limit", "1"})
	if code != 0 {
		t.Fatalf("run(cleanup) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.GeneratedCleanupReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Locations == 0 || report.Summary.Groups == 0 || report.Summary.CleanupCandidateFiles == 0 {
		t.Fatalf("cleanup summary = %+v, want generated cleanup candidates", report.Summary)
	}
}

func TestRunCleanupCanWriteExecutionDryRunReport(t *testing.T) {
	root := newRegistryRoot(t)
	writeFile(t, root, "tmp/orphan/matrix.json", "{}\n")
	out := filepath.Join(root, "tmp", "registry_generated_cleanup.json")
	execOut := filepath.Join(root, "tmp", "registry_generated_cleanup_execution.json")

	code := run([]string{"cleanup", "-root", root, "-out", out, "-exec-out", execOut, "-sample-limit", "1"})
	if code != 0 {
		t.Fatalf("run(cleanup -exec-out) = %d, want 0", code)
	}

	data, err := os.ReadFile(execOut)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.GeneratedCleanupExecutionReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Mode != "dry_run" || report.Summary.PlannedGroups == 0 || report.Summary.DeletedGroups != 0 {
		t.Fatalf("execution report = %+v, want dry-run planned cleanup", report.Summary)
	}
	if _, err := os.Stat(filepath.Join(root, "tmp", "orphan", "matrix.json")); err != nil {
		t.Fatalf("dry-run should keep orphan matrix: %v", err)
	}
}

func TestRunCoverageWritesReportAndReturnsOneForDrift(t *testing.T) {
	root := newRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 3)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)
	out := filepath.Join(root, "tmp", "registry_coverage.json")

	code := run([]string{
		"coverage",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
	})
	if code != 1 {
		t.Fatalf("run(coverage) = %d, want 1", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.CoverageReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != registry.StatusFail || report.Summary.Errors == 0 {
		t.Fatalf("status/errors = %q/%d, want fail/>0", report.Status, report.Summary.Errors)
	}
}

func TestRunCoverageCanWriteSummaryReport(t *testing.T) {
	root := newRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 2)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 2)
	out := filepath.Join(root, "tmp", "registry_coverage_summary.json")

	code := run([]string{
		"coverage",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-summary",
	})
	if code != 0 {
		t.Fatalf("run(coverage -summary) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var summary registry.CoverageSummaryReport
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Status != registry.StatusPass || summary.AtomRows != 0 {
		t.Fatalf("summary status/atom rows = %q/%d", summary.Status, summary.AtomRows)
	}
}

func TestRunCoverageCanWriteVersionAxisReport(t *testing.T) {
	root := newRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 2)
	writeJSON(t, root, "tmp/matrix/text/matrix.json", map[string]any{
		"schema_version": 1,
		"summary":        map[string]any{"total": 2, "passed": 2, "blocked": 0, "failed": 0, "skipped": 0},
		"cases": []map[string]any{
			{"recipe_name": "text-basic", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
			{"recipe_name": "text-basic", "source_version": "AE2025", "target_version": "AE2025", "status": "pass"},
		},
	})
	out := filepath.Join(root, "tmp", "registry_coverage_axis.json")

	code := run([]string{
		"coverage",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-axis",
		"-versions", "AE2020,AE2025",
	})
	if code != 0 {
		t.Fatalf("run(coverage -axis) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.CoverageAxisReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != registry.StatusPass || report.Summary.AtomRows != 1 || report.Summary.SourceTargetPairs != 4 {
		t.Fatalf("axis report = %+v", report.Summary)
	}
}

func TestRunCoverageAxisSupportsFocusedFilters(t *testing.T) {
	root := newRegistryRoot(t)
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":            "text",
				"artifact":      "tmp/matrix/text/matrix.json",
				"recipes":       []string{"text-basic"},
				"writer_status": "boundary",
				"totals":        map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0},
			},
		},
	})
	writeJSON(t, root, "tmp/matrix/text/matrix.json", map[string]any{
		"schema_version": 1,
		"summary":        map[string]any{"total": 2, "passed": 1, "blocked": 1, "failed": 0, "skipped": 0},
		"cases": []map[string]any{
			{"recipe_name": "text-basic", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
			{"recipe_name": "text-basic", "source_version": "AE2025", "target_version": "AE2024", "status": "blocked"},
		},
	})
	out := filepath.Join(root, "tmp", "registry_coverage_axis_blocked.json")

	code := run([]string{
		"coverage",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-axis",
		"-atom", "text.source.default",
		"-case-status", "blocked",
		"-versions", "AE2020,AE2024,AE2025",
	})
	if code != 0 {
		t.Fatalf("run(coverage -axis filtered) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.CoverageAxisReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.AtomRows != 1 || report.Summary.Cells != 1 || report.Summary.Blocked != 1 || report.Filter.CaseStatus != "blocked" {
		t.Fatalf("filtered axis report = %+v filter=%+v", report.Summary, report.Filter)
	}
}

func TestRunBoundariesWritesBoundaryCheckReport(t *testing.T) {
	root := newRegistryRoot(t)
	writeBoundaryCommandFixture(t, root, map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0})
	out := filepath.Join(root, "tmp", "registry_version_boundaries.json")

	code := run([]string{
		"boundaries",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-versions", "AE2020,AE2025",
	})
	if code != 0 {
		t.Fatalf("run(boundaries) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.VersionBoundaryCheckReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != registry.StatusPass || report.Summary.Matched != 1 || len(report.Boundaries) != 1 {
		t.Fatalf("boundary report = %+v", report)
	}
}

func TestRunBoundariesReturnsOneForBoundaryDrift(t *testing.T) {
	root := newRegistryRoot(t)
	writeBoundaryCommandFixture(t, root, map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0})
	out := filepath.Join(root, "tmp", "registry_version_boundaries.json")

	code := run([]string{
		"boundaries",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-versions", "AE2020,AE2025",
	})
	if code != 1 {
		t.Fatalf("run(boundaries drift) = %d, want 1", code)
	}
}

func TestRunGateWritesOrderedReports(t *testing.T) {
	root := newRegistryRoot(t)
	writeBoundaryCommandFixture(t, root, map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0})
	addCommandBoundaryContractGate(t, root)
	out := filepath.Join(root, "tmp", "registry_gate.json")

	code := run([]string{
		"gate",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-versions", "AE2020,AE2025",
	})
	if code != 0 {
		t.Fatalf("run(gate) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report gateReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != registry.StatusPass || report.Summary.Steps != 6 || report.Summary.Failed != 0 {
		t.Fatalf("gate report = %+v", report)
	}
	if !reflect.DeepEqual(report.VersionAxis, []string{"AE2020", "AE2025"}) {
		t.Fatalf("version axis = %v, want explicit AE2020/AE2025", report.VersionAxis)
	}
	wantOrder := []string{
		"registry_audit",
		"registry_version_boundaries",
		"registry_coverage",
		"registry_coverage_summary",
		"registry_coverage_axis",
		"registry_coverage_axis_boundary",
	}
	for i, want := range wantOrder {
		if report.Steps[i].ID != want {
			t.Fatalf("step %d = %q, want %q; steps=%+v", i, report.Steps[i].ID, want, report.Steps)
		}
	}
	for _, rel := range []string{
		"tmp/registry_audit.json",
		"tmp/registry_version_boundaries.json",
		"tmp/registry_coverage.json",
		"tmp/registry_coverage_summary.json",
		"tmp/registry_coverage_axis.json",
		"tmp/registry_coverage_axis_boundary.json",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("expected gate output %s: %v", rel, err)
		}
	}
}

func TestRunGateDefaultsVersionAxisInReport(t *testing.T) {
	root := newRegistryRoot(t)
	writeCoverageFixture(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", 0)
	writeMatrixFixture(t, root, "tmp/matrix/text/matrix.json", 0)
	out := filepath.Join(root, "tmp", "registry_gate.json")

	code := run([]string{
		"gate",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
	})
	if code != 0 {
		t.Fatalf("run(gate default versions) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report gateReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.VersionAxis, registry.DefaultAEVersionAxis()) {
		t.Fatalf("version axis = %v, want default %v", report.VersionAxis, registry.DefaultAEVersionAxis())
	}
}

func TestRunGateReturnsOneForFailedStep(t *testing.T) {
	root := newRegistryRoot(t)
	writeBoundaryCommandFixture(t, root, map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0})
	addCommandBoundaryContractGate(t, root)
	if err := os.Remove(filepath.Join(root, "examples", "recipes", "text-basic.json")); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "tmp", "registry_gate.json")

	code := run([]string{
		"gate",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-versions", "AE2020,AE2025",
	})
	if code != 1 {
		t.Fatalf("run(gate failed) = %d, want 1", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report gateReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != registry.StatusFail || report.Summary.Failed == 0 {
		t.Fatalf("gate report = %+v, want failed step", report)
	}
}

func TestRunGateAssetPolicyWritesOwnershipAndCleanupReports(t *testing.T) {
	root := newRegistryRoot(t)
	out := filepath.Join(root, "tmp", "registry_asset_gate.json")

	code := run([]string{
		"gate",
		"-root", root,
		"-out", out,
		"-scope", "asset-policy",
	})
	if code != 0 {
		t.Fatalf("run(gate asset-policy) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report gateReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Scope != "asset-policy" || report.Status != registry.StatusPass || report.Summary.Steps != 5 || report.Summary.Failed != 0 {
		t.Fatalf("asset gate report = %+v", report)
	}
	wantOrder := []string{
		"registry_audit",
		"registry_ownership",
		"registry_layout",
		"registry_generated_cleanup",
		"registry_generated_cleanup_execution_dry_run",
	}
	for i, want := range wantOrder {
		if report.Steps[i].ID != want {
			t.Fatalf("asset step %d = %q, want %q; steps=%+v", i, report.Steps[i].ID, want, report.Steps)
		}
	}
	for _, rel := range []string{
		"tmp/registry_ownership.json",
		"tmp/registry_layout.json",
		"tmp/registry_generated_cleanup.json",
		"tmp/registry_generated_cleanup_execution.json",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("expected asset gate output %s: %v", rel, err)
		}
	}
}

func TestRunGateAssetPolicyReturnsOneForLayoutBlockers(t *testing.T) {
	root := newRegistryRoot(t)
	writeFile(t, root, "examples/recipes/unowned.json", "{}\n")
	out := filepath.Join(root, "tmp", "registry_asset_gate.json")

	code := run([]string{
		"gate",
		"-root", root,
		"-out", out,
		"-scope", "asset-policy",
	})
	if code != 1 {
		t.Fatalf("run(gate asset-policy blocked) = %d, want 1", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report gateReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != registry.StatusFail || report.Summary.Failed == 0 {
		t.Fatalf("asset gate report = %+v, want failed layout", report)
	}
	foundLayoutFailure := false
	for _, step := range report.Steps {
		if step.ID == "registry_layout" && step.Status == registry.StatusFail && step.Errors > 0 {
			foundLayoutFailure = true
		}
	}
	if !foundLayoutFailure {
		t.Fatalf("layout failure not found in %+v", report.Steps)
	}
}

func TestRunCoverageCanWriteFilteredAtomRows(t *testing.T) {
	root := newRegistryRoot(t)
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":            "text",
				"artifact":      "tmp/matrix/text/matrix.json",
				"recipes":       []string{"text-basic"},
				"writer_status": "PD-6x6",
				"totals":        map[string]any{"total": 1, "pass": 1, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	writeJSON(t, root, "tmp/matrix/text/matrix.json", map[string]any{
		"schema_version": 1,
		"summary":        map[string]any{"total": 1, "passed": 1, "blocked": 0, "failed": 0, "skipped": 0},
		"cases": []map[string]any{
			{"recipe_name": "text-basic", "source_version": "AE2020", "target_version": "AE2025", "status": "pass"},
		},
	})
	out := filepath.Join(root, "tmp", "registry_coverage_rows.json")

	code := run([]string{
		"coverage",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-rows",
		"-record", "text",
	})
	if code != 0 {
		t.Fatalf("run(coverage -rows) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.CoverageRowsReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Count != 1 || report.Rows[0].AtomID != "text.source.default" {
		t.Fatalf("rows report = %+v, want one text.source.default row", report)
	}
}

func TestRunCoverageCanWriteFilteredMatrixCells(t *testing.T) {
	root := newRegistryRoot(t)
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":            "text",
				"artifact":      "tmp/matrix/text/matrix.json",
				"recipes":       []string{"text-basic"},
				"writer_status": "PD-6x6",
				"totals":        map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0},
			},
		},
	})
	writeJSON(t, root, "tmp/matrix/text/matrix.json", map[string]any{
		"schema_version": 1,
		"summary":        map[string]any{"total": 2, "passed": 1, "blocked": 1, "failed": 0, "skipped": 0},
		"cases": []map[string]any{
			{"recipe_name": "text-basic", "source_version": "AE2020", "target_version": "AE2025", "status": "pass"},
			{"recipe_name": "text-basic", "source_version": "AE2025", "target_version": "AE2020", "status": "blocked", "reason": "blocked fixture"},
		},
	})
	out := filepath.Join(root, "tmp", "registry_coverage_cells.json")

	code := run([]string{
		"coverage",
		"-root", root,
		"-coverage", "flightdeck/work/aep-understanding-generation/coverage.json",
		"-out", out,
		"-cells",
		"-atom", "text.source.default",
		"-case-status", "blocked",
	})
	if code != 0 {
		t.Fatalf("run(coverage -cells) = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.CoverageCellsReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Count != 1 || report.Cells[0].Status != "blocked" || report.Cells[0].Reason != "blocked fixture" {
		t.Fatalf("cells report = %+v, want one blocked fixture cell", report)
	}
}

func newRegistryRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "examples/recipes/text-basic.json", "{}\n")
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeJSON(t, root, "registry/locations.json", map[string]any{
		"schema_version": 1,
		"locations": []map[string]any{
			{"id": "registry", "path": "registry", "class": "spec_registry", "tracked": true, "required": true, "lifecycle": "source_of_truth"},
			{"id": "recipes", "path": "examples/recipes", "class": "atomic_fixtures", "tracked": true, "required": true, "lifecycle": "source_contract"},
			{"id": "tmp", "path": "tmp", "class": "generated_evidence", "tracked": false, "required": false, "lifecycle": "disposable"},
		},
	})
	writeJSON(t, root, "registry/workflows.json", map[string]any{
		"schema_version": 1,
		"workflows": []map[string]any{
			{"id": "generate", "summary": "Generate AEP from recipes", "cross_platform": "go"},
		},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version": 1,
		"capability_atoms": []map[string]any{
			{
				"id":       "text.source.default",
				"domain":   "text",
				"tier":     "atom",
				"status":   "verified",
				"platform": map[string]any{"host_required": false, "os": []string{"windows", "macos", "linux"}},
				"version_axis": map[string]any{
					"min_supported":    "AE2020",
					"known_supported":  []string{"AE2020", "AE2025"},
					"expansion_policy": "append_new_ae_versions",
				},
				"workflows": []string{"generate"},
				"dependencies": []map[string]any{
					{"kind": "recipe", "path": "examples/recipes/text-basic.json", "required": true},
				},
			},
			{
				"id":       "registry.asset_map",
				"domain":   "registry",
				"tier":     "source_contract",
				"status":   "verified",
				"platform": map[string]any{"host_required": false, "os": []string{"windows", "macos", "linux"}},
				"version_axis": map[string]any{
					"min_supported":    "AE2020",
					"known_supported":  []string{"AE2020", "AE2025"},
					"expansion_policy": "append_new_ae_versions",
				},
				"workflows": []string{"generate"},
				"dependencies": []map[string]any{
					{"kind": "registry", "path": "registry/locations.json", "required": true},
					{"kind": "registry", "path": "registry/workflows.json", "required": true},
					{"kind": "registry", "path": "registry/capability_atoms.json", "required": true},
					{"kind": "registry", "path": "registry/evidence.json", "required": true},
				},
			},
		},
	})
	writeJSON(t, root, "registry/evidence.json", map[string]any{
		"schema_version": 1,
		"evidence_sets": []map[string]any{
			{"id": "matrix.text", "class": "generated_evidence", "artifact_path": "tmp/text-basic/matrix.json", "required": true, "workflows": []string{"generate"}},
		},
	})
	return root
}

func writeJSON(t *testing.T, root, rel string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, rel, string(append(data, '\n')))
}

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readReport(t *testing.T, path string) registry.AuditReport {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report registry.AuditReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	return report
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

func writeBoundaryCommandFixture(t *testing.T, root string, expectedCells map[string]any) {
	t.Helper()
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"coverage": []map[string]any{
			{
				"id":            "text",
				"artifact":      "tmp/matrix/text/matrix.json",
				"recipes":       []string{"text-basic"},
				"writer_status": "boundary",
				"boundary": map[string]any{
					"status":             "known_text_boundary",
					"blocked_recipe_ids": []string{"text-basic"},
				},
				"totals": map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0},
			},
		},
	})
	writeJSON(t, root, "tmp/matrix/text/matrix.json", map[string]any{
		"schema_version": 1,
		"summary":        map[string]any{"total": 2, "passed": 1, "blocked": 1, "failed": 0, "skipped": 0},
		"cases": []map[string]any{
			{"recipe_name": "text-basic", "source_version": "AE2020", "target_version": "AE2020", "status": "pass"},
			{"recipe_name": "text-basic", "source_version": "AE2025", "target_version": "AE2020", "status": "blocked"},
		},
	})
	writeJSON(t, root, "registry/version_boundaries.json", map[string]any{
		"schema_version": 1,
		"version_boundaries": []map[string]any{
			{
				"id":                       "text.boundary",
				"atom_id":                  "text.source.default",
				"recipe":                   "text-basic",
				"feature":                  "test text boundary",
				"policy":                   "known_source_contract_boundary",
				"coverage_boundary_status": "known_text_boundary",
				"source_contract": map[string]any{
					"min_source_version":        "AE2025",
					"available_source_versions": []string{"AE2025"},
				},
				"target_contract": map[string]any{
					"supported_targets": []string{"AE2025"},
				},
				"expected_cells": expectedCells,
				"evidence":       []map[string]any{{"kind": "matrix", "path": "tmp/matrix/text/matrix.json", "required": true}},
			},
		},
	})
}

func addCommandBoundaryContractGate(t *testing.T, root string) {
	t.Helper()
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"contract_gates": []map[string]any{
			{
				"id":       "registry-version-boundaries",
				"kind":     "version_boundaries",
				"command":  "go run ./cmd/aepregistry boundaries -root . -out tmp/registry_version_boundaries.json",
				"artifact": "tmp/registry_version_boundaries.json",
				"status":   registry.StatusPass,
				"summary": map[string]any{
					"boundaries":        1,
					"matched":           1,
					"missing_rows":      0,
					"duplicate_rows":    0,
					"mismatched_totals": 0,
					"mismatched_status": 0,
					"checked_cells":     0,
					"missing_cells":     0,
					"mismatched_cells":  0,
					"errors":            0,
				},
			},
		},
		"coverage": []map[string]any{
			{
				"id":            "text",
				"artifact":      "tmp/matrix/text/matrix.json",
				"recipes":       []string{"text-basic"},
				"writer_status": "boundary",
				"boundary": map[string]any{
					"status":             "known_text_boundary",
					"blocked_recipe_ids": []string{"text-basic"},
				},
				"totals": map[string]any{"total": 2, "pass": 1, "blocked": 1, "failed": 0, "skipped": 0},
			},
		},
	})
}

func hasInventoryLocation(report registry.InventoryReport, id string) bool {
	for _, location := range report.Locations {
		if location.ID == id {
			return true
		}
	}
	return false
}

func hasOwnershipLocation(report registry.OwnershipReport, id string) bool {
	for _, location := range report.Locations {
		if location.ID == id {
			return true
		}
	}
	return false
}
