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

	tmp := findOwnershipLocation(t, report, "tmp")
	if tmp.Files != 2 || tmp.OwnedFiles != 1 || tmp.UnownedFiles != 1 {
		t.Fatalf("tmp files/owned/unowned = %d/%d/%d, want 2/1/1", tmp.Files, tmp.OwnedFiles, tmp.UnownedFiles)
	}
	if report.Summary.UnownedFiles < 2 {
		t.Fatalf("summary unowned = %d, want at least 2", report.Summary.UnownedFiles)
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
