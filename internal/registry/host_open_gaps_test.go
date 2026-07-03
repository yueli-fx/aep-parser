package registry

import "testing"

func TestPlanHostOpenGapsChunksPendingAndRepresentativeRecords(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"host_open_policy": map[string]any{
			"matrix_command_status": "available",
			"default_strategy":      "endpoint_direct",
			"broad_fanout_status":   "optional",
			"endpoint_inference": map[string]any{
				"label":          "endpoint",
				"direct_hosts":   []string{"AE2020", "AE2025"},
				"inferred_hosts": []string{"AE2021", "AE2022", "AE2023", "AE2024"},
			},
		},
		"coverage": []map[string]any{
			{
				"id":               "text",
				"domain":           "text",
				"recipes":          []string{"minimal-text-a", "minimal-text-b", "minimal-text-c"},
				"host_open_status": "pending_per_capability",
			},
			{
				"id":                        "shape",
				"domain":                    "shape",
				"recipes":                   []string{"minimal-shape-a", "minimal-shape-b"},
				"host_open_status":          "OPEN-ALL-HOSTS representative; pending for others",
				"host_open_representatives": []string{"minimal-shape-a"},
			},
			{
				"id":               "layer",
				"domain":           "layer",
				"recipes":          []string{"minimal-layer-a", "minimal-layer-boundary"},
				"host_open_status": "representative checks only",
				"boundary": map[string]any{
					"blocked_recipe_ids": []string{"minimal-layer-boundary"},
				},
			},
		},
	})

	report, err := PlanHostOpenGaps(root, "flightdeck/work/aep-understanding-generation/coverage.json", HostOpenGapPlanOptions{
		AERoot:         "E:/adobe",
		MaxAEOpenCases: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.GapGroups != 3 || report.Summary.GapRecipes != 5 || report.Summary.Chunks != 4 {
		t.Fatalf("summary = %+v, want 3 groups, 5 recipes, 4 chunks", report.Summary)
	}
	text := report.Gaps[0]
	if text.CoverageID != "text" || text.Classification != "direct_endpoint_gap" || len(text.Chunks) != 2 {
		t.Fatalf("text gap = %+v, want direct endpoint gap split into 2 chunks", text)
	}
	if text.Chunks[0].ExpectedAEOpenCases != 8 || text.Chunks[1].ExpectedAEOpenCases != 4 {
		t.Fatalf("text chunk cases = %+v", text.Chunks)
	}
	if text.Chunks[0].Command == "" || text.Chunks[0].Matrix != "registry/evidence/versioned-aep-migration/host_open_gaps/text/chunk-1/matrix.json" {
		t.Fatalf("text chunk = %+v", text.Chunks[0])
	}
	shape := report.Gaps[1]
	if shape.CoverageID != "shape" || len(shape.GapRecipes) != 1 || shape.GapRecipes[0] != "minimal-shape-b" {
		t.Fatalf("shape gap = %+v, want only non-representative recipe", shape)
	}
	layer := report.Gaps[2]
	if layer.CoverageID != "layer" || len(layer.GapRecipes) != 1 || layer.GapRecipes[0] != "minimal-layer-a" {
		t.Fatalf("layer gap = %+v, want boundary recipe excluded", layer)
	}
}

func TestPlanHostOpenGapsReportsZeroWhenCoverageIsComplete(t *testing.T) {
	root := newTestRegistryRoot(t)
	writeJSON(t, root, "flightdeck/work/aep-understanding-generation/coverage.json", map[string]any{
		"schema_version": 1,
		"host_open_policy": map[string]any{
			"endpoint_inference": map[string]any{
				"label":        "endpoint",
				"direct_hosts": []string{"AE2020", "AE2025"},
			},
		},
		"coverage": []map[string]any{
			{
				"id":               "text",
				"domain":           "text",
				"recipes":          []string{"minimal-text-a"},
				"host_open_status": "direct_endpoint_hosts_pass; AE2021-AE2024 inferred_by_endpoint",
			},
		},
	})

	report, err := PlanHostOpenGaps(root, "flightdeck/work/aep-understanding-generation/coverage.json", HostOpenGapPlanOptions{MaxAEOpenCases: 24})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.GapGroups != 0 || len(report.Gaps) != 0 {
		t.Fatalf("report = %+v, want no gaps", report)
	}
}
