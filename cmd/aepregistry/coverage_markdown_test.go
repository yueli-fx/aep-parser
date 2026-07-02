package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCoverageMDWritesMarkdown(t *testing.T) {
	root := newRegistryRoot(t)
	coveragePath := "flightdeck/work/aep-understanding-generation/coverage.json"
	writeJSON(t, root, coveragePath, map[string]any{
		"schema_version": 1,
		"host_open_policy": map[string]any{
			"matrix_command_status": "go_primary",
			"default_strategy":      "target_bound",
			"broad_fanout_status":   "bounded",
			"endpoint_inference": map[string]any{
				"label":          "edge-hosts",
				"direct_hosts":   []string{"AE2020", "AE2025"},
				"inferred_hosts": []string{"AE2021", "AE2022", "AE2023", "AE2024"},
			},
		},
		"coverage": []map[string]any{
			{
				"id":               "text",
				"domain":           "text",
				"writer_status":    "PD-6x6",
				"host_open_status": "OPEN-ALL-HOSTS",
				"artifact":         "tmp/matrix/text/matrix.json",
				"totals":           map[string]any{"total": 1, "pass": 1, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
	})
	out := filepath.Join(root, "tmp", "migration_coverage.md")

	code := run([]string{
		"coverage-md",
		"-root", root,
		"-coverage", coveragePath,
		"-out", out,
	})
	if code != 0 {
		t.Fatalf("run(coverage-md) = %d, want 0", code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	md := string(data)
	if !strings.Contains(md, "# Versioned AEP Migration Coverage") {
		t.Fatalf("markdown header missing:\n%s", md)
	}
	if !strings.Contains(md, "| `text` | text | PD-6x6 | 1/1 pass, blocked=0, failed=0, skipped=0 | OPEN-ALL-HOSTS | `tmp/matrix/text/matrix.json` |") {
		t.Fatalf("coverage row missing:\n%s", md)
	}
}
