package registry

import "testing"

func TestLayoutRepositoryClassifiesCleanupAndAtomizationActions(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "examples/recipes/text-basic.json", "{}\n")
	writeFile(t, root, "examples/recipes/text-extra.json", "{}\n")
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeFile(t, root, "tmp/orphan/matrix.json", "{}\n")
	writeFile(t, root, "data/samples/local.aep", "local\n")
	writeJSON(t, root, "registry/locations.json", map[string]any{
		"schema_version": 1,
		"locations": []map[string]any{
			{"id": "recipes", "path": "examples/recipes", "class": "atomic_fixtures", "tracked": true, "required": true, "lifecycle": "source_contract"},
			{"id": "tmp", "path": "tmp", "class": "generated_evidence", "tracked": false, "required": false, "lifecycle": "disposable"},
			{"id": "local_samples", "path": "data/samples", "class": "local_corpus", "tracked": false, "required": false, "lifecycle": "local_only"},
		},
	})
	writeJSON(t, root, "registry/workflows.json", map[string]any{
		"schema_version": 1,
		"workflows":      []map[string]any{{"id": "migrate", "summary": "Migrate", "cross_platform": "go"}},
	})
	writeJSON(t, root, "registry/capability_atoms.json", map[string]any{
		"schema_version": 1,
		"capability_atoms": []map[string]any{
			{
				"id":           "text.source.default",
				"domain":       "text",
				"tier":         "atom",
				"status":       "verified",
				"workflows":    []string{"migrate"},
				"dependencies": []map[string]any{{"kind": "recipe", "path": "examples/recipes/text-basic.json", "required": true}},
			},
		},
	})
	writeJSON(t, root, "registry/evidence.json", map[string]any{
		"schema_version": 1,
		"evidence_sets": []map[string]any{
			{"id": "matrix.text", "class": "generated_evidence", "artifact_path": "tmp/text-basic/matrix.json", "required": true, "workflows": []string{"migrate"}},
		},
	})

	report, err := LayoutRepository(root, LayoutOptions{SampleLimit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", report.SchemaVersion)
	}
	if report.Summary.Locations != 3 || report.Summary.CleanupCandidateFiles != 1 || report.Summary.BlockedUnownedFiles != 1 || report.Summary.LocalOnlyFiles != 1 {
		t.Fatalf("summary = %+v, want one cleanup candidate, one blocked unowned, one local-only file", report.Summary)
	}

	recipes := findLayoutLocation(t, report, "recipes")
	if recipes.Action != "atomize_or_register_unowned" || recipes.CleanupSafety != "blocked_until_owned" {
		t.Fatalf("recipes action/safety = %q/%q", recipes.Action, recipes.CleanupSafety)
	}
	if recipes.UnownedFiles != 1 || len(recipes.UnownedGroups) != 1 || recipes.UnownedGroups[0].Name != "text" {
		t.Fatalf("recipes unowned = %d groups=%+v", recipes.UnownedFiles, recipes.UnownedGroups)
	}

	tmp := findLayoutLocation(t, report, "tmp")
	if tmp.Action != "review_generated_cleanup" || tmp.CleanupCandidateFiles != 1 {
		t.Fatalf("tmp action/candidates = %q/%d", tmp.Action, tmp.CleanupCandidateFiles)
	}

	local := findLayoutLocation(t, report, "local_samples")
	if local.Action != "exclude_local_only" || local.LocalOnlyFiles != 1 {
		t.Fatalf("local action/files = %q/%d", local.Action, local.LocalOnlyFiles)
	}
}

func findLayoutLocation(t *testing.T, report LayoutReport, id string) LayoutLocation {
	t.Helper()
	for _, location := range report.Locations {
		if location.ID == id {
			return location
		}
	}
	t.Fatalf("layout location %q not found in %+v", id, report.Locations)
	return LayoutLocation{}
}
