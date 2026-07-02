package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type MigrationCoverageSummary struct {
	SchemaVersion       int                                 `json:"schema_version"`
	GeneratedFrom       string                              `json:"generated_from"`
	CoverageGeneratedAt string                              `json:"coverage_generated_at,omitempty"`
	Axes                MigrationCoverageAxes               `json:"axes"`
	Totals              MigrationCoverageTotals             `json:"totals"`
	HostOpenPolicy      MigrationCoverageHostOpenPolicy     `json:"host_open_policy"`
	RecurringGates      []MigrationCoverageArtifact         `json:"recurring_gates,omitempty"`
	DomainRollup        []MigrationCoverageDomainRollup     `json:"domain_rollup"`
	RecipeIndex         []MigrationCoverageRecipeIndexEntry `json:"recipe_index"`
	Coverage            []MigrationCoverageRecordSummary    `json:"coverage"`
	Boundaries          []MigrationCoverageBoundary         `json:"boundaries,omitempty"`
	OpenItems           []json.RawMessage                   `json:"open_items,omitempty"`
}

type MigrationCoverageHostOpenPolicy struct {
	MatrixCommandStatus string                             `json:"matrix_command_status,omitempty"`
	DefaultStrategy     string                             `json:"default_strategy,omitempty"`
	BroadFanoutStatus   string                             `json:"broad_fanout_status,omitempty"`
	EvidenceLevels      []string                           `json:"evidence_levels,omitempty"`
	EndpointInference   MigrationCoverageEndpointInference `json:"endpoint_inference,omitempty"`
}

type MigrationCoverageEndpointInference struct {
	Label         string   `json:"label,omitempty"`
	DirectHosts   []string `json:"direct_hosts,omitempty"`
	InferredHosts []string `json:"inferred_hosts,omitempty"`
}

type MigrationCoverageArtifact struct {
	ID       string         `json:"id,omitempty"`
	Command  string         `json:"command,omitempty"`
	Artifact string         `json:"artifact,omitempty"`
	Totals   CoverageTotals `json:"totals,omitempty"`
}

type MigrationCoverageAxes struct {
	SourceWriters []string `json:"source_writers"`
	TargetWriters []string `json:"target_writers"`
	HostOpenHosts []string `json:"host_open_hosts"`
}

type MigrationCoverageTotals struct {
	CoverageRecords        int                             `json:"coverage_records"`
	Domains                int                             `json:"domains"`
	Recipes                int                             `json:"recipes"`
	WriterCases            CoverageTotals                  `json:"writer_cases"`
	EndpointHostOpenCases  CoverageTotals                  `json:"endpoint_host_open_cases"`
	HostOpenEvidenceLevels MigrationHostOpenEvidenceCounts `json:"host_open_evidence_levels"`
	KnownBoundaries        int                             `json:"known_boundaries"`
	OpenItems              int                             `json:"open_items"`
}

type MigrationHostOpenEvidenceCounts struct {
	DirectEndpointHosts   int `json:"direct_endpoint_hosts"`
	RepresentativeOnly    int `json:"representative_only"`
	RepresentativeCovered int `json:"representative_covered"`
	ExcludedKnownBoundary int `json:"excluded_known_boundary"`
	RecordedStatusOnly    int `json:"recorded_status_only"`
}

type MigrationCoverageDomainRollup struct {
	Domain                 string                          `json:"domain"`
	CoverageIDs            []string                        `json:"coverage_ids"`
	RecipeCount            int                             `json:"recipe_count"`
	WriterTotals           CoverageTotals                  `json:"writer_totals"`
	WriterStatuses         []string                        `json:"writer_statuses"`
	HostOpenStatuses       []string                        `json:"host_open_statuses"`
	HostOpenEvidenceLevels MigrationHostOpenEvidenceCounts `json:"host_open_evidence_levels"`
	BoundaryIDs            []string                        `json:"boundary_ids"`
}

type MigrationCoverageRecordSummary struct {
	ID       string                     `json:"id"`
	Domain   string                     `json:"domain"`
	Scope    string                     `json:"scope,omitempty"`
	Recipes  []string                   `json:"recipes"`
	Writer   MigrationCoverageWriter    `json:"writer"`
	HostOpen MigrationCoverageHostOpen  `json:"host_open"`
	Boundary *MigrationCoverageBoundary `json:"boundary,omitempty"`
}

type MigrationCoverageWriter struct {
	Status   string         `json:"status"`
	Sources  []string       `json:"sources"`
	Targets  []string       `json:"targets"`
	Coverage string         `json:"coverage,omitempty"`
	Totals   CoverageTotals `json:"totals"`
	Artifact string         `json:"artifact,omitempty"`
}

type MigrationCoverageHostOpen struct {
	Status                       string                  `json:"status"`
	EvidenceLevel                string                  `json:"evidence_level"`
	OpenMode                     string                  `json:"open_mode,omitempty"`
	DirectHosts                  []string                `json:"direct_hosts,omitempty"`
	InferredHosts                []string                `json:"inferred_hosts,omitempty"`
	Totals                       CoverageTotals          `json:"totals,omitempty"`
	Representatives              []string                `json:"representatives,omitempty"`
	ExcludedKnownBoundaryRecipes []string                `json:"excluded_known_boundary_recipes,omitempty"`
	Matrix                       MigrationCoverageMatrix `json:"matrix,omitempty"`
}

type MigrationCoverageBoundary struct {
	ID               string                            `json:"id"`
	Domain           string                            `json:"domain"`
	Status           string                            `json:"status"`
	BlockedRecipeIDs []string                          `json:"blocked_recipe_ids"`
	Details          []MigrationCoverageBoundaryDetail `json:"details"`
}

type MigrationCoverageBoundaryDetail struct {
	Recipe string `json:"recipe"`
	Reason string `json:"reason"`
}

type MigrationCoverageRecipeIndexEntry struct {
	Recipe           string                    `json:"recipe"`
	CoverageID       string                    `json:"coverage_id"`
	Domain           string                    `json:"domain"`
	WriterStatus     string                    `json:"writer_status"`
	WriterSources    []string                  `json:"writer_sources"`
	WriterTargets    []string                  `json:"writer_targets"`
	WriterMatrix     MigrationCoverageMatrix   `json:"writer_matrix"`
	HostOpenStatus   string                    `json:"host_open_status"`
	HostOpenEvidence MigrationCoverageHostOpen `json:"host_open_evidence"`
	BoundaryStatus   string                    `json:"boundary_status"`
}

type MigrationCoverageMatrix struct {
	Artifact   string                      `json:"artifact,omitempty"`
	Totals     CoverageTotals              `json:"totals"`
	StatusSets MigrationCoverageStatusSets `json:"status_sets"`
}

type MigrationCoverageStatusSets struct {
	Pass    []string `json:"pass"`
	Blocked []string `json:"blocked"`
	Failed  []string `json:"failed"`
	Skipped []string `json:"skipped"`
}

type MigrationCoverageSummaryCheckReport struct {
	SchemaVersion int                                  `json:"schema_version"`
	Status        string                               `json:"status"`
	Errors        int                                  `json:"errors"`
	CoveragePath  string                               `json:"coverage_path"`
	SummaryPath   string                               `json:"summary_path"`
	Issues        []MigrationCoverageSummaryCheckIssue `json:"issues,omitempty"`
}

type MigrationCoverageSummaryCheckIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type MigrationCoverageSummaryQuery struct {
	Totals        bool
	Domain        string
	CoverageID    string
	Recipe        string
	EvidenceLevel string
}

type MigrationCoverageEvidenceLevelQueryResult struct {
	EvidenceLevel string                                 `json:"evidence_level"`
	Count         int                                    `json:"count"`
	Recipes       []MigrationCoverageEvidenceLevelRecipe `json:"recipes"`
}

type MigrationCoverageEvidenceLevelRecipe struct {
	Recipe           string                    `json:"recipe"`
	CoverageID       string                    `json:"coverage_id"`
	Domain           string                    `json:"domain"`
	HostOpenEvidence MigrationCoverageHostOpen `json:"host_open_evidence"`
}

type MigrationCoverageOverviewQueryResult struct {
	Totals     MigrationCoverageTotals          `json:"totals"`
	Domains    []MigrationCoverageDomainSummary `json:"domains"`
	Boundaries []MigrationCoverageBoundary      `json:"boundaries"`
	OpenItems  []json.RawMessage                `json:"open_items,omitempty"`
}

type MigrationCoverageDomainSummary struct {
	Domain                 string                          `json:"domain"`
	RecipeCount            int                             `json:"recipe_count"`
	WriterTotals           CoverageTotals                  `json:"writer_totals"`
	HostOpenEvidenceLevels MigrationHostOpenEvidenceCounts `json:"host_open_evidence_levels"`
	BoundaryIDs            []string                        `json:"boundary_ids"`
}

func BuildMigrationCoverageSummary(root, coveragePath string) (MigrationCoverageSummary, error) {
	var coverage coverageFile
	if err := readJSONPath(root, coveragePath, &coverage); err != nil {
		return MigrationCoverageSummary{}, err
	}
	summary := MigrationCoverageSummary{
		SchemaVersion:       1,
		GeneratedFrom:       coveragePath,
		CoverageGeneratedAt: coverage.GeneratedAt,
		Axes: MigrationCoverageAxes{
			SourceWriters: sortedStrings(coverage.WriterAxes.SourceWriters),
			TargetWriters: sortedStrings(coverage.WriterAxes.TargetWriters),
			HostOpenHosts: sortedStrings(coverage.HostOpenAxis.Hosts),
		},
		HostOpenPolicy: migrationHostOpenPolicy(coverage.HostOpenPolicy),
		RecurringGates: migrationCoverageArtifacts(coverage.RecurringGates),
		OpenItems:      coverage.OpenItems,
	}
	domains := map[string]*MigrationCoverageDomainRollup{}
	allRecipes := map[string]bool{}
	for _, record := range coverage.Coverage {
		recipes, err := coverageRecordRecipes(root, record)
		if err != nil {
			return MigrationCoverageSummary{}, err
		}
		matrixCases, err := matrixCasesByRecipe(root, record.Artifact)
		if err != nil {
			return MigrationCoverageSummary{}, err
		}
		hostOpenCases, err := hostOpenCasesByRecipe(root, record)
		if err != nil {
			return MigrationCoverageSummary{}, err
		}
		addCoverageTotals(&summary.Totals.WriterCases, record.Totals)
		domain := migrationDomain(domains, record.Domain)
		domain.CoverageIDs = appendUniqueSorted(domain.CoverageIDs, record.ID)
		domain.RecipeCount += len(recipes)
		addCoverageTotals(&domain.WriterTotals, record.Totals)
		domain.WriterStatuses = appendUniqueSorted(domain.WriterStatuses, record.WriterStatus)
		domain.HostOpenStatuses = appendUniqueSorted(domain.HostOpenStatuses, record.HostOpenStatus)
		recordSummary := MigrationCoverageRecordSummary{
			ID:      record.ID,
			Domain:  record.Domain,
			Scope:   record.Scope,
			Recipes: recipes,
			Writer: MigrationCoverageWriter{
				Status:   record.WriterStatus,
				Sources:  sortedStrings(coverage.WriterAxes.SourceWriters),
				Targets:  sortedStrings(coverage.WriterAxes.TargetWriters),
				Coverage: record.WriterCoverage,
				Totals:   record.Totals,
				Artifact: filepath.ToSlash(record.Artifact),
			},
			HostOpen: migrationRecordHostOpen(record),
		}
		if hasCoverageBoundary(record) {
			boundary := MigrationCoverageBoundary{
				ID:               record.ID,
				Domain:           record.Domain,
				Status:           record.Boundary.Status,
				BlockedRecipeIDs: sortedStrings(record.Boundary.BlockedRecipeIDs),
				Details:          migrationBoundaryDetails(record.Boundary.Details),
			}
			recordSummary.Boundary = &boundary
			summary.Boundaries = append(summary.Boundaries, boundary)
			domain.BoundaryIDs = appendUniqueSorted(domain.BoundaryIDs, record.ID)
		}
		for _, recipe := range recipes {
			allRecipes[recipe] = true
			hostOpen := migrationRecipeHostOpen(record, recipe, hostOpenCases[recipe])
			incrementHostOpenEvidence(&summary.Totals.HostOpenEvidenceLevels, hostOpen.EvidenceLevel)
			incrementHostOpenEvidence(&domain.HostOpenEvidenceLevels, hostOpen.EvidenceLevel)
			summary.RecipeIndex = append(summary.RecipeIndex, MigrationCoverageRecipeIndexEntry{
				Recipe:           recipe,
				CoverageID:       record.ID,
				Domain:           record.Domain,
				WriterStatus:     record.WriterStatus,
				WriterSources:    sortedStrings(coverage.WriterAxes.SourceWriters),
				WriterTargets:    sortedStrings(coverage.WriterAxes.TargetWriters),
				WriterMatrix:     migrationWriterMatrix(record.Artifact, matrixCases[recipe]),
				HostOpenStatus:   record.HostOpenStatus,
				HostOpenEvidence: hostOpen,
				BoundaryStatus:   migrationBoundaryStatus(record, recipe),
			})
		}
		if hasEndpointEvidence(record) {
			addCoverageTotals(&summary.Totals.EndpointHostOpenCases, record.HostOpenEndpointEvidence.Totals)
		}
		summary.Coverage = append(summary.Coverage, recordSummary)
	}
	summary.Totals.CoverageRecords = len(coverage.Coverage)
	summary.Totals.Domains = len(domains)
	summary.Totals.Recipes = len(allRecipes)
	summary.Totals.KnownBoundaries = len(summary.Boundaries)
	summary.Totals.OpenItems = len(coverage.OpenItems)
	summary.DomainRollup = migrationDomainRollups(domains)
	sort.Slice(summary.RecipeIndex, func(i, j int) bool {
		return summary.RecipeIndex[i].Recipe < summary.RecipeIndex[j].Recipe
	})
	return summary, nil
}

func QueryMigrationCoverageSummary(summary MigrationCoverageSummary, query MigrationCoverageSummaryQuery) (any, error) {
	if query.Totals {
		return summary.Totals, nil
	}
	if query.Recipe != "" {
		var matches []MigrationCoverageRecipeIndexEntry
		for _, recipe := range summary.RecipeIndex {
			if recipe.Recipe == query.Recipe {
				matches = append(matches, recipe)
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("recipe not found in coverage summary: %s", query.Recipe)
		}
		return matches, nil
	}
	if query.Domain != "" {
		var matches []MigrationCoverageDomainRollup
		for _, domain := range summary.DomainRollup {
			if domain.Domain == query.Domain {
				matches = append(matches, domain)
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("domain not found in coverage summary: %s", query.Domain)
		}
		return matches, nil
	}
	if query.CoverageID != "" {
		var matches []MigrationCoverageRecordSummary
		for _, record := range summary.Coverage {
			if record.ID == query.CoverageID {
				matches = append(matches, record)
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("coverage id not found in coverage summary: %s", query.CoverageID)
		}
		return matches, nil
	}
	if query.EvidenceLevel != "" {
		if !stringSet(summary.HostOpenPolicy.EvidenceLevels)[query.EvidenceLevel] {
			return nil, fmt.Errorf("host-open evidence level is not declared by coverage policy: %s", query.EvidenceLevel)
		}
		result := MigrationCoverageEvidenceLevelQueryResult{EvidenceLevel: query.EvidenceLevel}
		for _, recipe := range summary.RecipeIndex {
			if recipe.HostOpenEvidence.EvidenceLevel != query.EvidenceLevel {
				continue
			}
			result.Recipes = append(result.Recipes, MigrationCoverageEvidenceLevelRecipe{
				Recipe:           recipe.Recipe,
				CoverageID:       recipe.CoverageID,
				Domain:           recipe.Domain,
				HostOpenEvidence: recipe.HostOpenEvidence,
			})
		}
		result.Count = len(result.Recipes)
		return result, nil
	}
	domains := make([]MigrationCoverageDomainSummary, 0, len(summary.DomainRollup))
	for _, domain := range summary.DomainRollup {
		domains = append(domains, MigrationCoverageDomainSummary{
			Domain:                 domain.Domain,
			RecipeCount:            domain.RecipeCount,
			WriterTotals:           domain.WriterTotals,
			HostOpenEvidenceLevels: domain.HostOpenEvidenceLevels,
			BoundaryIDs:            domain.BoundaryIDs,
		})
	}
	return MigrationCoverageOverviewQueryResult{
		Totals:     summary.Totals,
		Domains:    domains,
		Boundaries: summary.Boundaries,
		OpenItems:  summary.OpenItems,
	}, nil
}

func CheckMigrationCoverageSummary(root, coveragePath, summaryPath string) (MigrationCoverageSummaryCheckReport, error) {
	expected, err := BuildMigrationCoverageSummary(root, coveragePath)
	if err != nil {
		return MigrationCoverageSummaryCheckReport{}, err
	}
	var actual MigrationCoverageSummary
	if err := readJSONPath(root, summaryPath, &actual); err != nil {
		return MigrationCoverageSummaryCheckReport{}, err
	}
	report := MigrationCoverageSummaryCheckReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		CoveragePath:  coveragePath,
		SummaryPath:   summaryPath,
	}
	expectedJSON, err := canonicalJSON(expected)
	if err != nil {
		return MigrationCoverageSummaryCheckReport{}, err
	}
	actualJSON, err := canonicalJSON(actual)
	if err != nil {
		return MigrationCoverageSummaryCheckReport{}, err
	}
	if !bytes.Equal(expectedJSON, actualJSON) {
		report.Issues = append(report.Issues, MigrationCoverageSummaryCheckIssue{
			Code:    "migration_coverage_summary_drift",
			Message: "summary JSON does not match current coverage ledger and matrix artifacts",
		})
	}
	declared := stringSet(expected.HostOpenPolicy.EvidenceLevels)
	if len(declared) > 0 {
		for _, level := range []string{"direct_endpoint_hosts", "representative_only", "representative_covered", "excluded_known_boundary", "recorded_status_only"} {
			if !declared[level] {
				report.Issues = append(report.Issues, MigrationCoverageSummaryCheckIssue{
					Code:    "undeclared_host_open_evidence_level",
					Message: fmt.Sprintf("coverage host_open_policy.evidence_levels does not declare %q", level),
				})
			}
		}
	}
	report.Errors = len(report.Issues)
	if report.Errors > 0 {
		report.Status = StatusFail
	}
	return report, nil
}

func coverageRecordRecipes(root string, record coverageRecord) ([]string, error) {
	if len(record.Recipes) > 0 {
		return sortedStrings(record.Recipes), nil
	}
	if record.RecipePattern == "" {
		return nil, nil
	}
	matches, err := filepath.Glob(filepath.Join(root, "examples", "recipes", filepath.FromSlash(record.RecipePattern)))
	if err != nil {
		return nil, err
	}
	recipes := make([]string, 0, len(matches))
	for _, match := range matches {
		recipes = append(recipes, stringsTrimExt(filepath.Base(match)))
	}
	return sortedStrings(recipes), nil
}

func migrationHostOpenPolicy(policy coverageHostOpenPolicy) MigrationCoverageHostOpenPolicy {
	return MigrationCoverageHostOpenPolicy{
		MatrixCommandStatus: policy.MatrixCommandStatus,
		DefaultStrategy:     policy.DefaultStrategy,
		BroadFanoutStatus:   policy.BroadFanoutStatus,
		EvidenceLevels:      sortedStrings(policy.EvidenceLevels),
		EndpointInference: MigrationCoverageEndpointInference{
			Label:         policy.EndpointInference.Label,
			DirectHosts:   sortedStrings(policy.EndpointInference.DirectHosts),
			InferredHosts: sortedStrings(policy.EndpointInference.InferredHosts),
		},
	}
}

func migrationCoverageArtifacts(artifacts []coverageArtifact) []MigrationCoverageArtifact {
	out := make([]MigrationCoverageArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		out = append(out, MigrationCoverageArtifact{
			ID:       artifact.ID,
			Command:  artifact.Command,
			Artifact: filepath.ToSlash(artifact.Artifact),
			Totals:   artifact.Totals,
		})
	}
	return out
}

func migrationBoundaryDetails(details []coverageBoundaryDetail) []MigrationCoverageBoundaryDetail {
	out := make([]MigrationCoverageBoundaryDetail, 0, len(details))
	for _, detail := range details {
		out = append(out, MigrationCoverageBoundaryDetail{
			Recipe: detail.Recipe,
			Reason: detail.Reason,
		})
	}
	return out
}

func stringsTrimExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func matrixCasesByRecipe(root, path string) (map[string][]matrixCase, error) {
	cases := map[string][]matrixCase{}
	if path == "" || !fileExists(root, path) {
		return cases, nil
	}
	var matrix matrixFile
	if err := readJSONPath(root, path, &matrix); err != nil {
		return nil, err
	}
	for _, c := range matrix.Cases {
		cases[c.RecipeName] = append(cases[c.RecipeName], c)
	}
	return cases, nil
}

func hostOpenCasesByRecipe(root string, record coverageRecord) (map[string][]matrixCase, error) {
	cases := map[string][]matrixCase{}
	if !hasEndpointEvidence(record) {
		return cases, nil
	}
	artifacts := hostOpenArtifacts(record.HostOpenEndpointEvidence)
	for _, artifact := range artifacts {
		grouped, err := matrixCasesByRecipe(root, artifact)
		if err != nil {
			return nil, err
		}
		for recipe, recipeCases := range grouped {
			cases[recipe] = append(cases[recipe], recipeCases...)
		}
	}
	return cases, nil
}

func hostOpenArtifacts(endpoint coverageEndpoint) []string {
	var artifacts []string
	for _, chunk := range endpoint.Chunks {
		if chunk.Artifact != "" {
			artifacts = append(artifacts, chunk.Artifact)
		}
	}
	if len(artifacts) == 0 && endpoint.Artifact != "" {
		artifacts = append(artifacts, endpoint.Artifact)
	}
	return sortedStrings(artifacts)
}

func migrationDomain(domains map[string]*MigrationCoverageDomainRollup, name string) *MigrationCoverageDomainRollup {
	domain := domains[name]
	if domain == nil {
		domain = &MigrationCoverageDomainRollup{Domain: name}
		domains[name] = domain
	}
	return domain
}

func migrationDomainRollups(domains map[string]*MigrationCoverageDomainRollup) []MigrationCoverageDomainRollup {
	names := make([]string, 0, len(domains))
	for name := range domains {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]MigrationCoverageDomainRollup, 0, len(names))
	for _, name := range names {
		out = append(out, *domains[name])
	}
	return out
}

func migrationRecordHostOpen(record coverageRecord) MigrationCoverageHostOpen {
	hostOpen := MigrationCoverageHostOpen{Status: record.HostOpenStatus, EvidenceLevel: "recorded_status_only"}
	if hasEndpointEvidence(record) {
		hostOpen.EvidenceLevel = "direct_endpoint_hosts"
		hostOpen.OpenMode = record.HostOpenEndpointEvidence.OpenMode
		hostOpen.DirectHosts = sortedStrings(record.HostOpenEndpointEvidence.DirectHosts)
		hostOpen.InferredHosts = sortedStrings(record.HostOpenEndpointEvidence.InferredHosts)
		hostOpen.Totals = record.HostOpenEndpointEvidence.Totals
		hostOpen.ExcludedKnownBoundaryRecipes = sortedStrings(record.HostOpenEndpointEvidence.ExcludedKnownBoundaryRecipes)
		return hostOpen
	}
	if len(record.HostOpenRepresentatives) > 0 {
		hostOpen.EvidenceLevel = "representative_only"
		hostOpen.Representatives = sortedStrings(record.HostOpenRepresentatives)
	}
	return hostOpen
}

func migrationRecipeHostOpen(record coverageRecord, recipe string, cases []matrixCase) MigrationCoverageHostOpen {
	hostOpen := MigrationCoverageHostOpen{Status: record.HostOpenStatus, EvidenceLevel: "recorded_status_only"}
	representatives := stringSet(record.HostOpenRepresentatives)
	if representatives[recipe] {
		hostOpen.EvidenceLevel = "representative_only"
		return hostOpen
	}
	if len(record.HostOpenRepresentatives) > 0 && !hasEndpointEvidence(record) {
		hostOpen.EvidenceLevel = "representative_covered"
		hostOpen.Representatives = sortedStrings(record.HostOpenRepresentatives)
		return hostOpen
	}
	if !hasEndpointEvidence(record) {
		return hostOpen
	}
	endpoint := record.HostOpenEndpointEvidence
	hostOpen.EvidenceLevel = "direct_endpoint_hosts"
	hostOpen.OpenMode = endpoint.OpenMode
	hostOpen.DirectHosts = sortedStrings(endpoint.DirectHosts)
	hostOpen.InferredHosts = sortedStrings(endpoint.InferredHosts)
	if stringSet(endpoint.ExcludedKnownBoundaryRecipes)[recipe] {
		hostOpen.EvidenceLevel = "excluded_known_boundary"
		return hostOpen
	}
	hostOpen.Matrix = migrationHostOpenMatrix(cases)
	return hostOpen
}

func migrationWriterMatrix(artifact string, cases []matrixCase) MigrationCoverageMatrix {
	matrix := migrationMatrix(cases, false)
	matrix.Artifact = filepath.ToSlash(artifact)
	return matrix
}

func migrationHostOpenMatrix(cases []matrixCase) MigrationCoverageMatrix {
	return migrationMatrix(cases, true)
}

func migrationMatrix(cases []matrixCase, includeAEOpen bool) MigrationCoverageMatrix {
	sort.Slice(cases, func(i, j int) bool {
		left := matrixCaseSortKey(cases[i])
		right := matrixCaseSortKey(cases[j])
		for idx := range left {
			if left[idx] != right[idx] {
				return left[idx] < right[idx]
			}
		}
		return false
	})
	matrix := MigrationCoverageMatrix{}
	for _, c := range cases {
		pair := c.SourceVersion + "->" + c.TargetVersion
		if includeAEOpen {
			pair += "@" + c.AEOpenVersion
		}
		matrix.Totals.Total++
		switch c.Status {
		case "pass":
			matrix.Totals.Pass++
			matrix.StatusSets.Pass = append(matrix.StatusSets.Pass, pair)
		case "blocked":
			matrix.Totals.Blocked++
			matrix.StatusSets.Blocked = append(matrix.StatusSets.Blocked, pair)
		case "failed":
			matrix.Totals.Failed++
			matrix.StatusSets.Failed = append(matrix.StatusSets.Failed, pair)
		case "skipped":
			matrix.Totals.Skipped++
			matrix.StatusSets.Skipped = append(matrix.StatusSets.Skipped, pair)
		}
	}
	return matrix
}

func matrixCaseSortKey(c matrixCase) [3]string {
	return [3]string{c.SourceVersion, c.TargetVersion, c.AEOpenVersion}
}

func hasEndpointEvidence(record coverageRecord) bool {
	endpoint := record.HostOpenEndpointEvidence
	return endpoint.OpenMode != "" || endpoint.Artifact != "" || len(endpoint.Chunks) > 0 || len(endpoint.DirectHosts) > 0 || endpoint.Totals.Total > 0
}

func hasCoverageBoundary(record coverageRecord) bool {
	return record.Boundary.Status != "" || len(record.Boundary.BlockedRecipeIDs) > 0 || len(record.Boundary.Details) > 0
}

func migrationBoundaryStatus(record coverageRecord, recipe string) string {
	status := boundaryStatus(record, recipe)
	if status == "" {
		return "none"
	}
	return status
}

func addCoverageTotals(acc *CoverageTotals, totals CoverageTotals) {
	acc.Total += totals.Total
	acc.Pass += totals.Pass
	acc.Blocked += totals.Blocked
	acc.Failed += totals.Failed
	acc.Skipped += totals.Skipped
}

func incrementHostOpenEvidence(counts *MigrationHostOpenEvidenceCounts, level string) {
	switch level {
	case "direct_endpoint_hosts":
		counts.DirectEndpointHosts++
	case "representative_only":
		counts.RepresentativeOnly++
	case "representative_covered":
		counts.RepresentativeCovered++
	case "excluded_known_boundary":
		counts.ExcludedKnownBoundary++
	case "recorded_status_only":
		counts.RecordedStatusOnly++
	}
}

func appendUniqueSorted(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	values = append(values, value)
	sort.Strings(values)
	return values
}

func fileExists(root, path string) bool {
	_, err := os.Stat(resolveRegistryPath(root, path))
	return err == nil
}

func canonicalJSON(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	return json.Marshal(decoded)
}
