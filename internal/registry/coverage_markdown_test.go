package registry

import (
	"strings"
	"testing"
)

func TestRenderCoverageMarkdownMirrorsLegacyCoverageSummary(t *testing.T) {
	root := newTestRegistryRoot(t)
	coveragePath := "registry/workflows/aep-understanding-generation/coverage.json"
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
		"recurring_gates": []map[string]any{
			{
				"id":       "smoke",
				"status":   "pass",
				"artifact": "tmp/matrix/smoke/matrix.json",
				"totals":   map[string]any{"total": 2, "pass": 2, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
		"coverage": []map[string]any{
			{
				"id":               "text",
				"domain":           "text",
				"writer_status":    "PD-6x6",
				"host_open_status": "OPEN-ALL-HOSTS",
				"artifact":         "tmp/matrix/text/matrix.json",
				"totals":           map[string]any{"total": 36, "pass": 36, "blocked": 0, "failed": 0, "skipped": 0},
			},
		},
		"open_items": []map[string]any{
			{"id": "text.kerning", "status": "planned", "scope": "text"},
		},
	})

	md, err := RenderCoverageMarkdown(root, coveragePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# Versioned AEP Migration Coverage",
		"Generated from `registry/workflows/aep-understanding-generation/coverage.json`.",
		"- Matrix command: go_primary",
		"- Endpoint inference: direct=AE2020,AE2025; inferred=AE2021,AE2022,AE2023,AE2024; label=edge-hosts",
		"| `smoke` | pass | 2/2 pass, blocked=0, failed=0, skipped=0 | `tmp/matrix/smoke/matrix.json` |",
		"| `text` | text | PD-6x6 | 36/36 pass, blocked=0, failed=0, skipped=0 | OPEN-ALL-HOSTS | `tmp/matrix/text/matrix.json` |",
		"- `text.kerning`: planned - text",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q:\n%s", want, md)
		}
	}
}
