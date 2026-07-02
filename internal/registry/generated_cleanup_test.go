package registry

import "testing"

func TestGeneratedCleanupClassifiesReferencedMixedAndUnreferencedGroups(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "tmp/migration_matrix_text/matrix.json", "{}\n")
	writeFile(t, root, "tmp/migration_matrix_text/log.txt", "log\n")
	writeFile(t, root, "tmp/migration_assess.json", "{}\n")
	writeFile(t, root, "tmp/technique_smoke/report.json", "{}\n")
	writeFile(t, root, "test_data/generated/ship-gate/generated.aep", "aep\n")
	writeJSON(t, root, "registry/locations.json", map[string]any{
		"schema_version": 1,
		"locations": []map[string]any{
			{"id": "tmp_evidence", "path": "tmp", "class": "generated_evidence", "tracked": false, "required": false, "lifecycle": "disposable"},
			{"id": "generated_test_data", "path": "test_data/generated", "class": "generated_evidence", "tracked": false, "required": false, "lifecycle": "rebuildable"},
		},
	})
	writeJSON(t, root, "registry/workflows.json", map[string]any{
		"schema_version": 1,
		"workflows": []map[string]any{
			{"id": "migrate", "summary": "Migrate", "cross_platform": "go"},
			{"id": "host_open", "summary": "Host open", "cross_platform": "host_required"},
			{"id": "generate", "summary": "Generate", "cross_platform": "go"},
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

	report, err := GeneratedCleanupRepository(root, GeneratedCleanupOptions{SampleLimit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", report.SchemaVersion)
	}
	if report.Summary.Locations != 2 || report.Summary.Groups != 4 || report.Summary.Files != 5 {
		t.Fatalf("summary counts = %+v, want 2 locations, 4 groups, 5 files", report.Summary)
	}
	if report.Summary.ReferencedFiles != 1 || report.Summary.UnreferencedFiles != 4 || report.Summary.CleanupCandidateFiles != 4 || report.Summary.MixedGroups != 1 {
		t.Fatalf("summary refs = %+v, want one referenced, four cleanup candidates, one mixed group", report.Summary)
	}

	matrix := findGeneratedGroup(t, report, "tmp_evidence", "migration_matrix_text")
	if matrix.Action != "review_mixed_registered_generated" || matrix.ProducerCategory != "version_matrix" {
		t.Fatalf("matrix action/category = %q/%q", matrix.Action, matrix.ProducerCategory)
	}
	if matrix.ReferencedFiles != 1 || matrix.UnreferencedFiles != 1 || len(matrix.EvidenceIDs) != 1 || matrix.EvidenceIDs[0] != "matrix.text" {
		t.Fatalf("matrix refs/evidence = %+v", matrix)
	}
	if matrix.CleanupOperation != "preserve_paths_then_review_unreferenced_siblings" || matrix.CleanupTarget != "tmp/migration_matrix_text" {
		t.Fatalf("matrix cleanup operation/target = %q/%q", matrix.CleanupOperation, matrix.CleanupTarget)
	}
	if len(matrix.PreservePaths) != 1 || matrix.PreservePaths[0] != "tmp/migration_matrix_text/matrix.json" {
		t.Fatalf("matrix preserve paths = %+v", matrix.PreservePaths)
	}
	if len(matrix.DeleteSamples) != 1 || matrix.DeleteSamples[0] != "tmp/migration_matrix_text/log.txt" {
		t.Fatalf("matrix delete samples = %+v", matrix.DeleteSamples)
	}

	technique := findGeneratedGroup(t, report, "tmp_evidence", "technique_smoke")
	if technique.Action != "cleanup_candidate" || technique.ProducerCategory != "technique_learning" {
		t.Fatalf("technique action/category = %q/%q", technique.Action, technique.ProducerCategory)
	}
	if technique.CleanupOperation != "delete_directory_tree" || technique.CleanupTarget != "tmp/technique_smoke" {
		t.Fatalf("technique cleanup operation/target = %q/%q", technique.CleanupOperation, technique.CleanupTarget)
	}

	migrationRootFiles := findGeneratedGroup(t, report, "tmp_evidence", "migration")
	if migrationRootFiles.CleanupOperation != "delete_file_prefix_matches" || migrationRootFiles.CleanupTarget != "tmp/migration*" {
		t.Fatalf("migration root file operation/target = %q/%q", migrationRootFiles.CleanupOperation, migrationRootFiles.CleanupTarget)
	}

	shipGate := findGeneratedGroup(t, report, "generated_test_data", "ship-gate")
	if shipGate.Action != "cleanup_candidate" || shipGate.ProducerCategory != "ship_gate_fixtures" {
		t.Fatalf("ship-gate action/category = %q/%q", shipGate.Action, shipGate.ProducerCategory)
	}
}

func findGeneratedGroup(t *testing.T, report GeneratedCleanupReport, locationID, groupName string) GeneratedCleanupGroup {
	t.Helper()
	for _, location := range report.Locations {
		if location.ID != locationID {
			continue
		}
		for _, group := range location.Groups {
			if group.Group == groupName {
				return group
			}
		}
	}
	t.Fatalf("generated cleanup group %s/%s not found in %+v", locationID, groupName, report.Locations)
	return GeneratedCleanupGroup{}
}
