package registry

import (
	"fmt"
	"math"
	"strings"
)

type HostOpenGapPlanOptions struct {
	AERoot         string
	MaxAEOpenCases int
}

type HostOpenGapPlanReport struct {
	SchemaVersion  int                     `json:"schema_version"`
	Status         string                  `json:"status"`
	CoverageSource string                  `json:"coverage_source"`
	HostOpenPolicy HostOpenGapPolicy       `json:"host_open_policy"`
	Planner        HostOpenGapPlanner      `json:"planner"`
	Summary        HostOpenGapSummary      `json:"summary"`
	Gaps           []HostOpenGapRecordPlan `json:"gaps"`
}

type HostOpenGapPolicy struct {
	MatrixCommandStatus string   `json:"matrix_command_status,omitempty"`
	DefaultStrategy     string   `json:"default_strategy,omitempty"`
	BroadFanoutStatus   string   `json:"broad_fanout_status,omitempty"`
	EndpointLabel       string   `json:"endpoint_label,omitempty"`
	DirectHosts         []string `json:"direct_hosts"`
	InferredHosts       []string `json:"inferred_hosts,omitempty"`
}

type HostOpenGapPlanner struct {
	MaxAEOpenCases  int      `json:"max_ae_open_cases"`
	AERoot          string   `json:"ae_root,omitempty"`
	Sources         []string `json:"sources"`
	Targets         []string `json:"targets"`
	AEOpenMode      string   `json:"ae_open_mode"`
	AEOpenHosts     []string `json:"ae_open_hosts"`
	CasesPerRecipe  int      `json:"cases_per_recipe"`
	RecipesPerChunk int      `json:"recipes_per_chunk"`
}

type HostOpenGapSummary struct {
	GapGroups  int `json:"gap_groups"`
	GapRecipes int `json:"gap_recipes"`
	Chunks     int `json:"chunks"`
}

type HostOpenGapRecordPlan struct {
	CoverageID                     string                 `json:"coverage_id"`
	Domain                         string                 `json:"domain,omitempty"`
	HostOpenStatus                 string                 `json:"host_open_status"`
	Classification                 string                 `json:"classification"`
	Recommendation                 string                 `json:"recommendation"`
	EndpointDirectHosts            []string               `json:"endpoint_direct_hosts"`
	InferredHostsAfterEndpointPass []string               `json:"inferred_hosts_after_endpoint_pass,omitempty"`
	TotalRecipes                   int                    `json:"total_recipes"`
	RepresentativeRecipes          []string               `json:"representative_recipes,omitempty"`
	ExcludedKnownBoundaryRecipes   []string               `json:"excluded_known_boundary_recipes,omitempty"`
	GapRecipes                     []string               `json:"gap_recipes"`
	Chunks                         []HostOpenGapChunkPlan `json:"chunks"`
}

type HostOpenGapChunkPlan struct {
	ID                  string   `json:"id"`
	Recipes             []string `json:"recipes"`
	ExpectedAEOpenCases int      `json:"expected_ae_open_cases"`
	Command             string   `json:"command"`
	Matrix              string   `json:"matrix"`
	Ledger              string   `json:"ledger"`
}

func PlanHostOpenGaps(root, coveragePath string, opts HostOpenGapPlanOptions) (HostOpenGapPlanReport, error) {
	var coverage coverageFile
	if err := readJSONPath(root, coveragePath, &coverage); err != nil {
		return HostOpenGapPlanReport{}, err
	}
	directHosts := sortedStrings(coverage.HostOpenPolicy.EndpointInference.DirectHosts)
	if len(directHosts) == 0 {
		return HostOpenGapPlanReport{}, fmt.Errorf("host_open_policy.endpoint_inference.direct_hosts is empty")
	}
	maxCases := opts.MaxAEOpenCases
	if maxCases <= 0 {
		maxCases = 24
	}
	casesPerRecipe := len(directHosts) * len(directHosts)
	if casesPerRecipe <= 0 {
		return HostOpenGapPlanReport{}, fmt.Errorf("invalid endpoint case calculation")
	}
	recipesPerChunk := int(math.Max(1, math.Floor(float64(maxCases)/float64(casesPerRecipe))))
	report := HostOpenGapPlanReport{
		SchemaVersion:  1,
		Status:         StatusPass,
		CoverageSource: coveragePath,
		HostOpenPolicy: HostOpenGapPolicy{
			MatrixCommandStatus: coverage.HostOpenPolicy.MatrixCommandStatus,
			DefaultStrategy:     coverage.HostOpenPolicy.DefaultStrategy,
			BroadFanoutStatus:   coverage.HostOpenPolicy.BroadFanoutStatus,
			EndpointLabel:       coverage.HostOpenPolicy.EndpointInference.Label,
			DirectHosts:         directHosts,
			InferredHosts:       sortedStrings(coverage.HostOpenPolicy.EndpointInference.InferredHosts),
		},
		Planner: HostOpenGapPlanner{
			MaxAEOpenCases:  maxCases,
			AERoot:          opts.AERoot,
			Sources:         directHosts,
			Targets:         directHosts,
			AEOpenMode:      "target_bound",
			AEOpenHosts:     directHosts,
			CasesPerRecipe:  casesPerRecipe,
			RecipesPerChunk: recipesPerChunk,
		},
	}
	for _, record := range coverage.Coverage {
		gapRecipes, classification, recommendation := hostOpenGapRecipes(record)
		if len(gapRecipes) == 0 {
			continue
		}
		plan := HostOpenGapRecordPlan{
			CoverageID:                     record.ID,
			Domain:                         record.Domain,
			HostOpenStatus:                 record.HostOpenStatus,
			Classification:                 classification,
			Recommendation:                 recommendation,
			EndpointDirectHosts:            directHosts,
			InferredHostsAfterEndpointPass: sortedStrings(coverage.HostOpenPolicy.EndpointInference.InferredHosts),
			TotalRecipes:                   len(record.Recipes),
			RepresentativeRecipes:          sortedStrings(record.HostOpenRepresentatives),
			ExcludedKnownBoundaryRecipes:   sortedStrings(record.Boundary.BlockedRecipeIDs),
			GapRecipes:                     gapRecipes,
			Chunks:                         hostOpenGapChunks(record.ID, gapRecipes, directHosts, opts.AERoot, maxCases, recipesPerChunk),
		}
		report.Summary.GapRecipes += len(gapRecipes)
		report.Summary.Chunks += len(plan.Chunks)
		report.Gaps = append(report.Gaps, plan)
	}
	report.Summary.GapGroups = len(report.Gaps)
	return report, nil
}

func hostOpenGapRecipes(record coverageRecord) ([]string, string, string) {
	allRecipes := sortedStrings(record.Recipes)
	representatives := stringSet(record.HostOpenRepresentatives)
	boundaries := stringSet(record.Boundary.BlockedRecipeIDs)
	switch {
	case record.HostOpenStatus == "pending_per_capability":
		return allRecipes, "direct_endpoint_gap", "run_endpoint_direct_host_open"
	case strings.Contains(record.HostOpenStatus, "pending for others"):
		return filterRecipeSet(allRecipes, representatives), "partial_representative_gap", "run_endpoint_direct_host_open_for_non_representatives"
	case record.HostOpenStatus == "representative checks only":
		return filterRecipeSet(allRecipes, boundaries), "representative_only_gap", "run_endpoint_direct_host_open_for_non_boundary_recipes"
	default:
		return nil, "covered_or_not_required", "no_action"
	}
}

func filterRecipeSet(recipes []string, excluded map[string]bool) []string {
	var out []string
	for _, recipe := range recipes {
		if !excluded[recipe] {
			out = append(out, recipe)
		}
	}
	return out
}

func hostOpenGapChunks(recordID string, recipes, directHosts []string, aeRoot string, maxCases, recipesPerChunk int) []HostOpenGapChunkPlan {
	var chunks []HostOpenGapChunkPlan
	for start := 0; start < len(recipes); start += recipesPerChunk {
		end := start + recipesPerChunk
		if end > len(recipes) {
			end = len(recipes)
		}
		chunkRecipes := append([]string(nil), recipes[start:end]...)
		chunkID := fmt.Sprintf("chunk-%d", len(chunks)+1)
		outRoot := fmt.Sprintf("registry/evidence/versioned-aep-migration/host_open_gaps/%s/%s", recordID, chunkID)
		chunks = append(chunks, HostOpenGapChunkPlan{
			ID:                  chunkID,
			Recipes:             chunkRecipes,
			ExpectedAEOpenCases: len(chunkRecipes) * len(directHosts) * len(directHosts),
			Command:             hostOpenGapCommand(chunkRecipes, directHosts, aeRoot, maxCases, outRoot),
			Matrix:              outRoot + "/matrix.json",
			Ledger:              outRoot + "/ledger.md",
		})
	}
	return chunks
}

func hostOpenGapCommand(recipes, directHosts []string, aeRoot string, maxCases int, outRoot string) string {
	parts := []string{"go run ./cmd/aepmigrate matrix"}
	for _, recipe := range recipes {
		parts = append(parts, "-recipe examples/recipes/"+recipe+".json")
	}
	if aeRoot != "" {
		parts = append(parts, "-ae-root "+aeRoot)
	}
	axis := strings.Join(directHosts, ",")
	parts = append(parts,
		"-sources "+axis,
		"-targets "+axis,
		"-ae-open",
		fmt.Sprintf("-max-ae-open-cases %d", maxCases),
		"-out "+outRoot,
		"-ledger-out "+outRoot+"/ledger.md",
	)
	return strings.Join(parts, " ")
}
