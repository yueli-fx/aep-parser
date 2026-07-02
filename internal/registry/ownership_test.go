package registry

import "testing"

func TestOwnershipRepositoryReportsOwnedAndUnownedFiles(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "examples/recipes/text-basic.json", "{}\n")
	writeFile(t, root, "examples/recipes/text-extra.json", "{}\n")
	writeFile(t, root, "tmp/text-basic/matrix.json", "{}\n")
	writeFile(t, root, "tmp/unused/matrix.json", "{}\n")
	writeValidRegistry(t, root, []map[string]any{
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
	})

	report, err := OwnershipRepository(root, OwnershipOptions{SampleLimit: 1})
	if err != nil {
		t.Fatal(err)
	}
	recipes := findOwnershipLocation(t, report, "recipes")
	if recipes.Files != 2 || recipes.OwnedFiles != 1 || recipes.UnownedFiles != 1 {
		t.Fatalf("recipes files/owned/unowned = %d/%d/%d, want 2/1/1", recipes.Files, recipes.OwnedFiles, recipes.UnownedFiles)
	}
	if len(recipes.UnownedSamples) != 1 || recipes.UnownedSamples[0] != "examples/recipes/text-extra.json" {
		t.Fatalf("recipes unowned samples = %+v", recipes.UnownedSamples)
	}
	if len(recipes.UnownedGroups) != 1 || recipes.UnownedGroups[0].Name != "text" || recipes.UnownedGroups[0].Files != 1 {
		t.Fatalf("recipes unowned groups = %+v", recipes.UnownedGroups)
	}

	tmp := findOwnershipLocation(t, report, "tmp")
	if tmp.Files != 2 || tmp.OwnedFiles != 1 || tmp.UnownedFiles != 1 {
		t.Fatalf("tmp files/owned/unowned = %d/%d/%d, want 2/1/1", tmp.Files, tmp.OwnedFiles, tmp.UnownedFiles)
	}
	if len(tmp.UnownedGroups) != 1 || tmp.UnownedGroups[0].Name != "unused" || tmp.UnownedGroups[0].Files != 1 {
		t.Fatalf("tmp unowned groups = %+v", tmp.UnownedGroups)
	}
	if report.Summary.UnownedFiles < 2 {
		t.Fatalf("summary unowned = %d, want at least 2", report.Summary.UnownedFiles)
	}
}

func TestOwnershipRepositoryGroupsUnownedFlatRecipeFilesByDomain(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "examples/recipes/minimal-camera-zoom.json", "{}\n")
	writeFile(t, root, "examples/recipes/minimal-camera-focus-distance.json", "{}\n")
	writeFile(t, root, "examples/recipes/minimal-comp-label.json", "{}\n")
	writeFile(t, root, "examples/recipes/minimal-text-style.json", "{}\n")
	writeValidRegistry(t, root, nil)

	report, err := OwnershipRepository(root, OwnershipOptions{SampleLimit: 0})
	if err != nil {
		t.Fatal(err)
	}
	recipes := findOwnershipLocation(t, report, "recipes")
	want := []UnownedGroup{
		{Name: "camera", Files: 2},
		{Name: "comp", Files: 1},
		{Name: "text", Files: 1},
	}
	if !sameGroups(recipes.UnownedGroups, want) {
		t.Fatalf("recipe groups = %+v, want %+v", recipes.UnownedGroups, want)
	}
}

func TestOwnershipRepositoryGroupsUnownedNestedFilesByFirstChild(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "internal/serializer/templates/effects/a.bin", "a")
	writeFile(t, root, "internal/serializer/templates/effects/b.bin", "b")
	writeFile(t, root, "internal/serializer/templates/text/a.bin", "a")
	writeValidRegistry(t, root, nil)

	report, err := OwnershipRepository(root, OwnershipOptions{SampleLimit: 0})
	if err != nil {
		t.Fatal(err)
	}
	templates := findOwnershipLocation(t, report, "templates")
	want := []UnownedGroup{
		{Name: "effects", Files: 2},
		{Name: "text", Files: 1},
	}
	if !sameGroups(templates.UnownedGroups, want) {
		t.Fatalf("template groups = %+v, want %+v", templates.UnownedGroups, want)
	}
}

func TestOwnershipRepositoryTreatsDirectoryDependencyAsOwningSubtree(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "internal/serializer/templates/text/a.bin", "a")
	writeFile(t, root, "internal/serializer/templates/text/b.bin", "b")
	writeValidRegistry(t, root, []map[string]any{
		{
			"id":       "text.templates",
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
				{"kind": "template_dir", "path": "internal/serializer/templates/text", "required": true},
			},
		},
	})

	report, err := OwnershipRepository(root, OwnershipOptions{SampleLimit: 5})
	if err != nil {
		t.Fatal(err)
	}
	templates := findOwnershipLocation(t, report, "templates")
	if templates.Files != 2 || templates.OwnedFiles != 2 || templates.UnownedFiles != 0 {
		t.Fatalf("templates files/owned/unowned = %d/%d/%d, want 2/2/0", templates.Files, templates.OwnedFiles, templates.UnownedFiles)
	}
}

func TestOwnershipRepositoryTreatsGlobDependencyAsOwningMatches(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "test_data/generators/verify_text.jsx", "// verify\n")
	writeFile(t, root, "test_data/generators/verify_shape.jsx", "// verify\n")
	writeFile(t, root, "test_data/generators/re_text.jsx", "// re\n")
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

	report, err := OwnershipRepository(root, OwnershipOptions{SampleLimit: 5})
	if err != nil {
		t.Fatal(err)
	}
	generators := findOwnershipLocation(t, report, "fixture_generators")
	if generators.Files != 3 || generators.OwnedFiles != 2 || generators.UnownedFiles != 1 {
		t.Fatalf("generators files/owned/unowned = %d/%d/%d, want 3/2/1", generators.Files, generators.OwnedFiles, generators.UnownedFiles)
	}
	if len(generators.UnownedSamples) != 1 || generators.UnownedSamples[0] != "test_data/generators/re_text.jsx" {
		t.Fatalf("generators unowned samples = %+v", generators.UnownedSamples)
	}
}

func TestOwnershipRepositoryTreatsVersionBoundaryEvidenceAsOwned(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "tmp/boundary/matrix.json", "{}\n")
	writeFile(t, root, "tmp/unused/matrix.json", "{}\n")
	writeValidRegistry(t, root, nil)
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
					"min_source_version":        "AE2025",
					"available_source_versions": []string{"AE2025"},
				},
				"target_contract": map[string]any{
					"supported_targets": []string{"AE2025"},
				},
				"evidence": []map[string]any{{"kind": "matrix", "path": "tmp/boundary/matrix.json", "required": true}},
			},
		},
	})

	report, err := OwnershipRepository(root, OwnershipOptions{SampleLimit: 5})
	if err != nil {
		t.Fatal(err)
	}
	tmp := findOwnershipLocation(t, report, "tmp")
	if tmp.Files != 2 || tmp.OwnedFiles != 1 || tmp.UnownedFiles != 1 {
		t.Fatalf("tmp files/owned/unowned = %d/%d/%d, want 2/1/1", tmp.Files, tmp.OwnedFiles, tmp.UnownedFiles)
	}
}

func findOwnershipLocation(t *testing.T, report OwnershipReport, id string) LocationOwnership {
	t.Helper()
	for _, location := range report.Locations {
		if location.ID == id {
			return location
		}
	}
	t.Fatalf("ownership location %q not found in %+v", id, report.Locations)
	return LocationOwnership{}
}

func sameGroups(got, want []UnownedGroup) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
