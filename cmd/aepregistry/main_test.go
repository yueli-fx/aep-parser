package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func hasInventoryLocation(report registry.InventoryReport, id string) bool {
	for _, location := range report.Locations {
		if location.ID == id {
			return true
		}
	}
	return false
}
