package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditRepositoryPassesForDeclaredAssets(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "examples/recipes/text-basic.json", "{}\n")
	writeFile(t, root, "internal/serializer/templates/text/source.json", "{}\n")
	writeFile(t, root, "test_data/fixtures/text-basic.aep", "fixture\n")
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeValidRegistry(t, root, []map[string]any{
		{
			"id":       "text.source.default",
			"domain":   "text",
			"tier":     "atom",
			"status":   "verified",
			"platform": map[string]any{"host_required": false, "os": []string{"windows", "macos", "linux"}},
			"version_axis": map[string]any{
				"min_supported":    "AE2020",
				"known_supported":  []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024", "AE2025"},
				"expansion_policy": "append_new_ae_versions",
			},
			"workflows": []string{"generate", "migrate", "profile"},
			"dependencies": []map[string]any{
				{"kind": "recipe", "path": "examples/recipes/text-basic.json", "required": true},
				{"kind": "template", "path": "internal/serializer/templates/text/source.json", "required": true},
				{"kind": "fixture", "path": "test_data/fixtures/text-basic.aep", "required": true},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	if report.Summary.Errors != 0 || report.Summary.Warnings != 0 {
		t.Fatalf("summary errors/warnings = %d/%d, want 0/0", report.Summary.Errors, report.Summary.Warnings)
	}
}

func TestAuditRepositoryReportsMissingRequiredDependency(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeValidRegistry(t, root, []map[string]any{
		{
			"id":       "text.animator.tracking",
			"domain":   "text",
			"tier":     "atom",
			"status":   "planned",
			"platform": map[string]any{"host_required": false, "os": []string{"windows", "macos", "linux"}},
			"version_axis": map[string]any{
				"min_supported":    "AE2020",
				"known_supported":  []string{"AE2020", "AE2025"},
				"expansion_policy": "append_new_ae_versions",
			},
			"workflows": []string{"generate"},
			"dependencies": []map[string]any{
				{"kind": "recipe", "path": "examples/recipes/text-animator-tracking.json", "required": true},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q", report.Status, StatusFail)
	}
	assertIssue(t, report, "missing_dependency", SeverityError, "text.animator.tracking", "examples/recipes/text-animator-tracking.json")
}

func TestAuditRepositoryAcceptsVersionBoundaryContract(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeValidRegistry(t, root, []map[string]any{
		{
			"id":       "layer.track_matte.explicit_source",
			"domain":   "layer",
			"tier":     "boundary",
			"status":   "boundary",
			"platform": map[string]any{"host_required": true, "os": []string{"windows", "macos"}},
			"version_axis": map[string]any{
				"min_supported":    "AE2020",
				"known_supported":  []string{"AE2020", "AE2025"},
				"expansion_policy": "append_new_ae_versions",
			},
			"workflows": []string{"generate", "migrate"},
		},
	})
	writeJSON(t, root, "registry/version_boundaries.json", map[string]any{
		"schema_version": 1,
		"version_boundaries": []map[string]any{
			{
				"id":      "ae2025.explicit_matte_source",
				"atom_id": "layer.track_matte.explicit_source",
				"recipe":  "minimal-layer-explicit-matte",
				"feature": "AE2025 explicit matte source references",
				"policy":  "known_source_contract_boundary",
				"source_contract": map[string]any{
					"min_source_version":          "AE2025",
					"available_source_versions":   []string{"AE2025"},
					"unavailable_source_versions": []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024"},
				},
				"target_contract": map[string]any{
					"supported_targets":         []string{"AE2025"},
					"blocked_downgrade_targets": []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024"},
				},
				"expected_cells": map[string]any{"total": 36, "pass": 1, "blocked": 5, "failed": 0, "skipped": 30},
				"evidence":       []map[string]any{{"kind": "matrix", "path": "tmp/text-basic/matrix.json", "required": true}},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; issues: %+v", report.Status, StatusPass, report.Issues)
	}
	if report.Summary.VersionBoundaries != 1 {
		t.Fatalf("version boundaries = %d, want 1", report.Summary.VersionBoundaries)
	}
}

func TestAuditRepositoryReportsUnknownVersionBoundaryAtom(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeValidRegistry(t, root, nil)
	writeJSON(t, root, "registry/version_boundaries.json", map[string]any{
		"schema_version": 1,
		"version_boundaries": []map[string]any{
			{
				"id":      "ae2025.explicit_matte_source",
				"atom_id": "missing.atom",
				"recipe":  "minimal-layer-explicit-matte",
				"feature": "AE2025 explicit matte source references",
				"policy":  "known_source_contract_boundary",
				"source_contract": map[string]any{
					"min_source_version":        "AE2025",
					"available_source_versions": []string{"AE2025"},
				},
				"target_contract": map[string]any{
					"supported_targets": []string{"AE2025"},
				},
				"evidence": []map[string]any{{"kind": "matrix", "path": "tmp/text-basic/matrix.json", "required": true}},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	assertBoundaryIssue(t, report, "unknown_version_boundary_atom", SeverityError, "ae2025.explicit_matte_source", "missing.atom", "")
}

func TestAuditRepositoryReportsUnknownWorkflowReferences(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeValidRegistry(t, root, []map[string]any{
		{
			"id":       "ai.workflow_generation",
			"domain":   "ai",
			"tier":     "workflow",
			"status":   "planned",
			"platform": map[string]any{"host_required": false, "os": []string{"windows", "macos", "linux"}},
			"version_axis": map[string]any{
				"min_supported":    "AE2020",
				"known_supported":  []string{"AE2020", "AE2025"},
				"expansion_policy": "append_new_ae_versions",
			},
			"workflows": []string{"not-a-real-workflow"},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	assertIssue(t, report, "unknown_workflow", SeverityError, "ai.workflow_generation", "not-a-real-workflow")
}

func TestAuditRepositoryTreatsMissingOptionalGeneratedEvidenceAsWarning(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeValidRegistry(t, root, nil)
	writeJSON(t, root, "registry/evidence.json", map[string]any{
		"schema_version": 1,
		"evidence_sets": []map[string]any{
			{
				"id":            "matrix.text.optional",
				"class":         "generated_evidence",
				"artifact_path": "tmp/missing-matrix/matrix.json",
				"required":      false,
				"workflows":     []string{"migrate"},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q", report.Status, StatusPass)
	}
	assertIssue(t, report, "missing_optional_evidence", SeverityWarning, "", "tmp/missing-matrix/matrix.json")
}

func TestAuditRepositoryReportsDependenciesOutsideRegisteredLocations(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "orphan/asset.json", "{}\n")
	writeValidRegistry(t, root, []map[string]any{
		{
			"id":       "orphan.atom",
			"domain":   "orphan",
			"tier":     "atom",
			"status":   "planned",
			"platform": map[string]any{"host_required": false, "os": []string{"windows", "macos", "linux"}},
			"version_axis": map[string]any{
				"min_supported":    "AE2020",
				"known_supported":  []string{"AE2020", "AE2025"},
				"expansion_policy": "append_new_ae_versions",
			},
			"workflows": []string{"generate"},
			"dependencies": []map[string]any{
				{"kind": "recipe", "path": "orphan/asset.json", "required": true},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	assertIssue(t, report, "unregistered_dependency_location", SeverityError, "orphan.atom", "orphan/asset.json")
}

func TestAuditRepositoryAcceptsRequiredGlobDependencyWhenItMatches(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "test_data/generators/verify_text.jsx", "// jsx\n")
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeValidRegistryWithGeneratorLocation(t, root, []map[string]any{
		{
			"id":       "fixture_generators.verify",
			"domain":   "fixture",
			"tier":     "source_family",
			"status":   "verified",
			"platform": map[string]any{"host_required": true, "os": []string{"windows", "macos"}},
			"version_axis": map[string]any{
				"min_supported":    "AE2020",
				"known_supported":  []string{"AE2020", "AE2025"},
				"expansion_policy": "append_new_ae_versions",
			},
			"workflows": []string{"generate"},
			"dependencies": []map[string]any{
				{"kind": "glob", "path": "test_data/generators/verify_*.jsx", "required": true},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusPass {
		t.Fatalf("status = %q, want pass; issues: %+v", report.Status, report.Issues)
	}
}

func TestAuditRepositoryReportsMissingRequiredGlobDependency(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeValidRegistryWithGeneratorLocation(t, root, []map[string]any{
		{
			"id":       "fixture_generators.verify",
			"domain":   "fixture",
			"tier":     "source_family",
			"status":   "verified",
			"platform": map[string]any{"host_required": true, "os": []string{"windows", "macos"}},
			"version_axis": map[string]any{
				"min_supported":    "AE2020",
				"known_supported":  []string{"AE2020", "AE2025"},
				"expansion_policy": "append_new_ae_versions",
			},
			"workflows": []string{"generate"},
			"dependencies": []map[string]any{
				{"kind": "glob", "path": "test_data/generators/verify_*.jsx", "required": true},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusFail {
		t.Fatalf("status = %q, want fail", report.Status)
	}
	assertIssue(t, report, "missing_dependency", SeverityError, "fixture_generators.verify", "test_data/generators/verify_*.jsx")
}

func TestAuditRepositoryReportsEvidenceOutsideRegisteredLocations(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "orphan/matrix.json", "{}\n")
	writeValidRegistry(t, root, nil)
	writeJSON(t, root, "registry/evidence.json", map[string]any{
		"schema_version": 1,
		"evidence_sets": []map[string]any{
			{
				"id":            "orphan.matrix",
				"class":         "generated_evidence",
				"artifact_path": "orphan/matrix.json",
				"required":      true,
				"workflows":     []string{"migrate"},
			},
		},
	})

	report, err := AuditRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	assertIssue(t, report, "unregistered_evidence_location", SeverityError, "", "orphan/matrix.json")
}

func newTestRegistryRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{
		"registry",
		"examples/recipes",
		"internal/serializer/templates/text",
		"test_data/fixtures",
		"tmp",
	} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func writeValidRegistry(t *testing.T, root string, atoms []map[string]any) {
	t.Helper()
	writeJSON(t, root, "registry/locations.json", map[string]any{
		"schema_version": 1,
		"locations": []map[string]any{
			{"id": "registry", "path": "registry", "class": "spec_registry", "tracked": true, "required": true, "lifecycle": "source_of_truth"},
			{"id": "recipes", "path": "examples/recipes", "class": "atomic_fixtures", "tracked": true, "required": true, "lifecycle": "source_contract"},
			{"id": "templates", "path": "internal/serializer/templates", "class": "source_contract", "tracked": true, "required": true, "lifecycle": "source_contract"},
			{"id": "fixtures", "path": "test_data/fixtures", "class": "reverse_reference", "tracked": true, "required": true, "lifecycle": "regression_corpus"},
			{"id": "tmp", "path": "tmp", "class": "generated_evidence", "tracked": false, "required": false, "lifecycle": "disposable"},
		},
	})
	writeJSON(t, root, "registry/workflows.json", map[string]any{
		"schema_version": 1,
		"workflows": []map[string]any{
			{"id": "generate", "summary": "Generate AEP from recipes", "cross_platform": "go"},
			{"id": "migrate", "summary": "Convert AEP between supported versions", "cross_platform": "go_plus_host_optional"},
			{"id": "profile", "summary": "Parse detailed AEP profile data", "cross_platform": "go"},
		},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version":   1,
		"capability_atoms": atoms,
	})
	writeJSON(t, root, "registry/evidence.json", map[string]any{
		"schema_version": 1,
		"evidence_sets": []map[string]any{
			{
				"id":            "matrix.text",
				"class":         "generated_evidence",
				"artifact_path": "tmp/text-basic/matrix.json",
				"required":      true,
				"workflows":     []string{"migrate"},
			},
		},
	})
}

func writeValidRegistryWithGeneratorLocation(t *testing.T, root string, atoms []map[string]any) {
	t.Helper()
	writeValidRegistry(t, root, atoms)
	writeJSON(t, root, "registry/locations.json", map[string]any{
		"schema_version": 1,
		"locations": []map[string]any{
			{"id": "registry", "path": "registry", "class": "spec_registry", "tracked": true, "required": true, "lifecycle": "source_of_truth"},
			{"id": "recipes", "path": "examples/recipes", "class": "atomic_fixtures", "tracked": true, "required": true, "lifecycle": "source_contract"},
			{"id": "templates", "path": "internal/serializer/templates", "class": "source_contract", "tracked": true, "required": true, "lifecycle": "source_contract"},
			{"id": "fixtures", "path": "test_data/fixtures", "class": "reverse_reference", "tracked": true, "required": true, "lifecycle": "regression_corpus"},
			{"id": "fixture_generators", "path": "test_data/generators", "class": "reverse_reference", "tracked": true, "required": true, "lifecycle": "regression_source"},
			{"id": "tmp", "path": "tmp", "class": "generated_evidence", "tracked": false, "required": false, "lifecycle": "disposable"},
		},
	})
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

func assertIssue(t *testing.T, report AuditReport, code, severity, atomID, path string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Code == code && issue.Severity == severity && issue.AtomID == atomID && issue.Path == path {
			return
		}
	}
	t.Fatalf("issue %s/%s atom=%q path=%q not found in %+v", code, severity, atomID, path, report.Issues)
}

func assertBoundaryIssue(t *testing.T, report AuditReport, code, severity, boundaryID, atomID, path string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Code == code && issue.Severity == severity && issue.BoundaryID == boundaryID && issue.AtomID == atomID && issue.Path == path {
			return
		}
	}
	t.Fatalf("issue %s/%s boundary=%q atom=%q path=%q not found in %+v", code, severity, boundaryID, atomID, path, report.Issues)
}
