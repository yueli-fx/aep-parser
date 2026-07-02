package registry

import "testing"

func TestInventoryRepositoryCountsRegisteredLocations(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeFile(t, root, "examples/recipes/a.json", "12345")
	writeFile(t, root, "examples/recipes/nested/b.json", "12")
	writeFile(t, root, "tmp/run/matrix.json", "123")
	writeValidRegistry(t, root, nil)

	report, err := InventoryRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", report.SchemaVersion)
	}
	if report.Summary.Locations != 5 {
		t.Fatalf("locations = %d, want 5", report.Summary.Locations)
	}
	if report.Summary.Files == 0 || report.Summary.Bytes == 0 {
		t.Fatalf("summary files/bytes = %d/%d, want non-zero", report.Summary.Files, report.Summary.Bytes)
	}

	recipes := findInventoryLocation(t, report, "recipes")
	if !recipes.Exists {
		t.Fatalf("recipes location should exist")
	}
	if recipes.Files != 2 || recipes.Bytes != 7 {
		t.Fatalf("recipes files/bytes = %d/%d, want 2/7", recipes.Files, recipes.Bytes)
	}

	tmp := findInventoryLocation(t, report, "tmp")
	if tmp.Files != 1 {
		t.Fatalf("tmp files = %d, want 1", tmp.Files)
	}
}

func findInventoryLocation(t *testing.T, report InventoryReport, id string) LocationInventory {
	t.Helper()
	for _, location := range report.Locations {
		if location.ID == id {
			return location
		}
	}
	t.Fatalf("inventory location %q not found in %+v", id, report.Locations)
	return LocationInventory{}
}
