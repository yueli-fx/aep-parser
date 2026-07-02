package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteGeneratedCleanupDryRunPlansOnlyCleanupCandidates(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "tmp/migration_matrix_text/matrix.json", "{}\n")
	writeFile(t, root, "tmp/migration_matrix_text/log.txt", "log\n")
	writeFile(t, root, "tmp/registry_audit.json", "{}\n")
	writeFile(t, root, "tmp/technique_smoke/report.json", "{}\n")
	writeGeneratedCleanupRegistry(t, root)

	cleanup, err := GeneratedCleanupRepository(root, GeneratedCleanupOptions{SampleLimit: 2})
	if err != nil {
		t.Fatal(err)
	}
	exec := ExecuteGeneratedCleanup(root, cleanup, GeneratedCleanupExecutionOptions{
		ExcludeProducers: []string{"registry_report"},
	})

	if exec.Mode != "dry_run" {
		t.Fatalf("mode = %q, want dry_run", exec.Mode)
	}
	if exec.Summary.PlannedGroups != 1 || exec.Summary.SkippedGroups != 2 || exec.Summary.DeletedGroups != 0 || exec.Summary.Errors != 0 {
		t.Fatalf("summary = %+v, want one planned, two skipped, no deletes/errors", exec.Summary)
	}
	technique := findGeneratedExecutionOp(t, exec, "tmp_evidence", "technique_smoke")
	if technique.Status != "planned" || technique.Reason != "dry_run" || technique.CleanupOperation != "delete_directory_tree" {
		t.Fatalf("technique op = %+v", technique)
	}
	matrix := findGeneratedExecutionOp(t, exec, "tmp_evidence", "migration_matrix_text")
	if matrix.Status != "skipped" || matrix.Reason != "not_cleanup_candidate" {
		t.Fatalf("matrix op = %+v", matrix)
	}
	registryReport := findGeneratedExecutionOp(t, exec, "tmp_evidence", "registry")
	if registryReport.Status != "skipped" || registryReport.Reason != "producer_excluded" {
		t.Fatalf("registry op = %+v", registryReport)
	}
	if !testPathExists(root, "tmp/technique_smoke/report.json") {
		t.Fatalf("dry run deleted technique report")
	}
}

func TestExecuteGeneratedCleanupApplyDeletesOnlyCleanupCandidateTargets(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "tmp/migration_matrix_text/matrix.json", "{}\n")
	writeFile(t, root, "tmp/migration_matrix_text/log.txt", "log\n")
	writeFile(t, root, "tmp/technique_smoke/report.json", "{}\n")
	writeFile(t, root, "tmp/migration_assess.json", "{}\n")
	writeGeneratedCleanupRegistry(t, root)

	cleanup, err := GeneratedCleanupRepository(root, GeneratedCleanupOptions{SampleLimit: 2})
	if err != nil {
		t.Fatal(err)
	}
	exec := ExecuteGeneratedCleanup(root, cleanup, GeneratedCleanupExecutionOptions{
		Apply: true,
	})

	if exec.Mode != "apply" {
		t.Fatalf("mode = %q, want apply", exec.Mode)
	}
	if exec.Summary.DeletedGroups != 2 || exec.Summary.SkippedGroups != 1 || exec.Summary.Errors != 0 {
		t.Fatalf("summary = %+v, want two deleted groups, one skipped mixed group", exec.Summary)
	}
	if testPathExists(root, "tmp/technique_smoke/report.json") {
		t.Fatalf("apply kept cleanup candidate directory")
	}
	if testPathExists(root, "tmp/migration_assess.json") {
		t.Fatalf("apply kept cleanup candidate prefix file")
	}
	if !testPathExists(root, "tmp/migration_matrix_text/matrix.json") {
		t.Fatalf("apply deleted registered preserve path")
	}
	if !testPathExists(root, "tmp/migration_matrix_text/log.txt") {
		t.Fatalf("apply deleted mixed group sibling without preserve handling")
	}
}

func TestExecuteGeneratedCleanupRejectsTargetsOutsideGeneratedLocations(t *testing.T) {
	report := GeneratedCleanupReport{
		Locations: []GeneratedCleanupLocation{
			{ID: "tmp_evidence", Path: "tmp", Groups: []GeneratedCleanupGroup{
				{
					LocationID:        "tmp_evidence",
					Group:             "bad",
					ProducerCategory:  "version_matrix",
					Action:            "cleanup_candidate",
					CleanupOperation:  "delete_directory_tree",
					CleanupTarget:     "examples/recipes",
					UnreferencedFiles: 1,
					Files:             1,
				},
			}},
		},
	}

	exec := ExecuteGeneratedCleanup(t.TempDir(), report, GeneratedCleanupExecutionOptions{Apply: true})
	if exec.Summary.Errors != 1 {
		t.Fatalf("summary = %+v, want one error", exec.Summary)
	}
	op := exec.Operations[0]
	if op.Status != "error" {
		t.Fatalf("op status = %q, want error", op.Status)
	}
}

func writeGeneratedCleanupRegistry(t *testing.T, root string) {
	t.Helper()
	writeJSON(t, root, "registry/locations.json", map[string]any{
		"schema_version": 1,
		"locations": []map[string]any{
			{"id": "tmp_evidence", "path": "tmp", "class": "generated_evidence", "tracked": false, "required": false, "lifecycle": "disposable"},
		},
	})
	writeJSON(t, root, "registry/workflows.json", map[string]any{
		"schema_version": 1,
		"workflows": []map[string]any{
			{"id": "migrate", "summary": "Migrate", "cross_platform": "go"},
			{"id": "host_open", "summary": "Host open", "cross_platform": "host_required"},
			{"id": "learn", "summary": "Learn", "cross_platform": "go"},
			{"id": "profile", "summary": "Profile", "cross_platform": "go"},
			{"id": "ai_generate", "summary": "AI generate", "cross_platform": "go"},
		},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version":   1,
		"capability_atoms": []map[string]any{},
	})
	writeJSON(t, root, "registry/evidence.json", map[string]any{
		"schema_version": 1,
		"evidence_sets": []map[string]any{
			{"id": "matrix.text", "class": "generated_evidence", "artifact_path": "tmp/migration_matrix_text/matrix.json", "required": true, "workflows": []string{"migrate", "host_open"}},
		},
	})
}

func findGeneratedExecutionOp(t *testing.T, report GeneratedCleanupExecutionReport, locationID, group string) GeneratedCleanupExecutionOp {
	t.Helper()
	for _, op := range report.Operations {
		if op.LocationID == locationID && op.Group == group {
			return op
		}
	}
	t.Fatalf("execution op %s/%s not found in %+v", locationID, group, report.Operations)
	return GeneratedCleanupExecutionOp{}
}

func testPathExists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}
