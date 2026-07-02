package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CoverageReport struct {
	SchemaVersion int                    `json:"schema_version"`
	Status        string                 `json:"status"`
	Summary       CoverageSummary        `json:"summary"`
	Records       []CoverageRecordReport `json:"records,omitempty"`
	AtomRows      []AtomCoverageRow      `json:"atom_rows,omitempty"`
	Issues        []CoverageIssue        `json:"issues"`
}

type CoverageSummary struct {
	Records   int `json:"records"`
	Artifacts int `json:"artifacts"`
	Atoms     int `json:"atoms"`
	Errors    int `json:"errors"`
}

type CoverageIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	RecordID string `json:"record_id,omitempty"`
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
}

type CoverageTotals struct {
	Total   int `json:"total"`
	Pass    int `json:"pass"`
	Blocked int `json:"blocked"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

type CoverageRecordReport struct {
	ID              string         `json:"id"`
	Artifact        string         `json:"artifact"`
	DeclaredRecipes []string       `json:"declared_recipes,omitempty"`
	ObservedRecipes []string       `json:"observed_recipes,omitempty"`
	SourceVersions  []string       `json:"source_versions,omitempty"`
	TargetVersions  []string       `json:"target_versions,omitempty"`
	AEOpenVersions  []string       `json:"ae_open_versions,omitempty"`
	AtomIDs         []string       `json:"atom_ids,omitempty"`
	Totals          CoverageTotals `json:"totals"`
}

type AtomCoverageRow struct {
	AtomID                string         `json:"atom_id"`
	RecordID              string         `json:"record_id"`
	Recipe                string         `json:"recipe,omitempty"`
	Artifact              string         `json:"artifact"`
	WriterStatus          string         `json:"writer_status,omitempty"`
	BoundaryStatus        string         `json:"boundary_status,omitempty"`
	BoundaryReason        string         `json:"boundary_reason,omitempty"`
	SourceVersions        []string       `json:"source_versions,omitempty"`
	TargetVersions        []string       `json:"target_versions,omitempty"`
	AEOpenVersions        []string       `json:"ae_open_versions,omitempty"`
	HostOpenEvidenceLevel string         `json:"host_open_evidence_level,omitempty"`
	DirectHostVersions    []string       `json:"direct_host_versions,omitempty"`
	InferredHostVersions  []string       `json:"inferred_host_versions,omitempty"`
	Totals                CoverageTotals `json:"totals"`
}

type coverageFile struct {
	SchemaVersion  int                `json:"schema_version"`
	RecurringGates []coverageArtifact `json:"recurring_gates"`
	Coverage       []coverageRecord   `json:"coverage"`
}

type coverageRecord struct {
	ID                       string           `json:"id"`
	Artifact                 string           `json:"artifact"`
	Recipes                  []string         `json:"recipes"`
	WriterStatus             string           `json:"writer_status"`
	HostOpenStatus           string           `json:"host_open_status"`
	HostOpenRepresentatives  []string         `json:"host_open_representatives"`
	Totals                   CoverageTotals   `json:"totals"`
	HostOpenEndpointEvidence coverageEndpoint `json:"host_open_endpoint_evidence"`
	Boundary                 coverageBoundary `json:"boundary"`
}

type coverageEndpoint struct {
	Artifact      string             `json:"artifact"`
	Recipes       []string           `json:"recipes"`
	DirectHosts   []string           `json:"direct_hosts"`
	InferredHosts []string           `json:"inferred_hosts"`
	Totals        CoverageTotals     `json:"totals"`
	Chunks        []coverageArtifact `json:"chunks"`
}

type coverageArtifact struct {
	ID       string         `json:"id"`
	Command  string         `json:"command"`
	Artifact string         `json:"artifact"`
	Totals   CoverageTotals `json:"totals"`
}

type coverageBoundary struct {
	Status           string                   `json:"status"`
	BlockedRecipeIDs []string                 `json:"blocked_recipe_ids"`
	Details          []coverageBoundaryDetail `json:"details"`
}

type coverageBoundaryDetail struct {
	Recipe string `json:"recipe"`
	Reason string `json:"reason"`
}

type matrixFile struct {
	Summary matrixSummary `json:"summary"`
	Cases   []matrixCase  `json:"cases"`
}

type matrixSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Blocked int `json:"blocked"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

type matrixCase struct {
	RecipeName    string `json:"recipe_name"`
	SourceVersion string `json:"source_version"`
	TargetVersion string `json:"target_version"`
	AEOpenVersion string `json:"ae_open_version"`
	Status        string `json:"status"`
}

type coverageArtifactRef struct {
	RecordID string
	Path     string
	Totals   CoverageTotals
}

func ValidateCoverage(root, coveragePath string) (CoverageReport, error) {
	var coverage coverageFile
	if err := readJSONPath(root, coveragePath, &coverage); err != nil {
		return CoverageReport{}, err
	}
	atomRefs := loadCoverageAtomRefs(root)

	refs := collectCoverageArtifactRefs(coverage)
	report := CoverageReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		Summary: CoverageSummary{
			Records:   len(coverage.RecurringGates) + len(coverage.Coverage),
			Artifacts: len(refs),
		},
	}
	for _, ref := range refs {
		report.checkArtifact(root, ref)
	}
	for _, record := range coverage.Coverage {
		report.checkCoverageRecord(root, record, atomRefs)
	}
	report.Summary.Atoms = countUniqueAtoms(report.Records, report.AtomRows)
	if report.Summary.Errors > 0 {
		report.Status = StatusFail
	}
	return report, nil
}

func collectCoverageArtifactRefs(coverage coverageFile) []coverageArtifactRef {
	var refs []coverageArtifactRef
	add := func(recordID, artifact string, totals CoverageTotals) {
		if artifact == "" {
			return
		}
		refs = append(refs, coverageArtifactRef{
			RecordID: recordID,
			Path:     filepath.ToSlash(artifact),
			Totals:   totals,
		})
	}
	for _, gate := range coverage.RecurringGates {
		add(gate.ID, gate.Artifact, gate.Totals)
	}
	for _, record := range coverage.Coverage {
		add(record.ID, record.Artifact, record.Totals)
		add(record.ID, record.HostOpenEndpointEvidence.Artifact, record.HostOpenEndpointEvidence.Totals)
		for _, chunk := range record.HostOpenEndpointEvidence.Chunks {
			add(record.ID, chunk.Artifact, chunk.Totals)
		}
	}
	return refs
}

func (r *CoverageReport) checkArtifact(root string, ref coverageArtifactRef) {
	var matrix matrixFile
	if err := readJSONPath(root, ref.Path, &matrix); err != nil {
		r.addError("missing_or_invalid_matrix", ref.RecordID, ref.Path, fmt.Sprintf("read matrix artifact: %v", err))
		return
	}
	actual := CoverageTotals{
		Total:   matrix.Summary.Total,
		Pass:    matrix.Summary.Passed,
		Blocked: matrix.Summary.Blocked,
		Failed:  matrix.Summary.Failed,
		Skipped: matrix.Summary.Skipped,
	}
	if actual != ref.Totals {
		r.addError("matrix_totals_mismatch", ref.RecordID, ref.Path, fmt.Sprintf("coverage totals %+v do not match matrix summary %+v", ref.Totals, actual))
	}
}

func (r *CoverageReport) checkCoverageRecord(root string, record coverageRecord, atomRefs coverageAtomRefs) {
	if record.Artifact == "" {
		return
	}
	var matrix matrixFile
	if err := readJSONPath(root, record.Artifact, &matrix); err != nil {
		return
	}
	detail := CoverageRecordReport{
		ID:              record.ID,
		Artifact:        filepath.ToSlash(record.Artifact),
		DeclaredRecipes: sortedStrings(record.Recipes),
		ObservedRecipes: matrixRecipes(matrix),
		SourceVersions:  matrixSourceVersions(matrix),
		TargetVersions:  matrixTargetVersions(matrix),
		AEOpenVersions:  matrixAEOpenVersions(matrix),
		Totals:          record.Totals,
	}
	detail.AtomIDs = atomRefs.atomIDsForRecord(detail.Artifact, detail.ObservedRecipes)
	r.Records = append(r.Records, detail)
	r.AtomRows = append(r.AtomRows, atomRowsForRecord(record, matrix, atomRefs)...)
	sort.Slice(r.AtomRows, func(i, j int) bool {
		if r.AtomRows[i].AtomID == r.AtomRows[j].AtomID {
			return r.AtomRows[i].Recipe < r.AtomRows[j].Recipe
		}
		return r.AtomRows[i].AtomID < r.AtomRows[j].AtomID
	})
	r.checkDeclaredRecipes(record.ID, detail.Artifact, detail.DeclaredRecipes, detail.ObservedRecipes)
}

func (r *CoverageReport) checkDeclaredRecipes(recordID, artifact string, declared, observed []string) {
	if len(declared) == 0 {
		return
	}
	declaredSet := stringSet(declared)
	observedSet := stringSet(observed)
	for _, recipe := range declared {
		if !observedSet[recipe] {
			r.addError("matrix_missing_declared_recipe", recordID, artifact, fmt.Sprintf("declared recipe %q was not found in matrix cases", recipe))
		}
	}
	for _, recipe := range observed {
		if !declaredSet[recipe] {
			r.addError("matrix_unexpected_recipe", recordID, artifact, fmt.Sprintf("matrix recipe %q is not declared by the coverage record", recipe))
		}
	}
}

func (r *CoverageReport) addError(code, recordID, path, message string) {
	r.Summary.Errors++
	r.Issues = append(r.Issues, CoverageIssue{
		Code:     code,
		Severity: SeverityError,
		RecordID: recordID,
		Path:     filepath.ToSlash(path),
		Message:  message,
	})
}

type coverageAtomRefs struct {
	byRecipe   map[string][]string
	byEvidence map[string][]string
}

func loadCoverageAtomRefs(root string) coverageAtomRefs {
	var atoms atomsFile
	if err := readJSONPath(root, "registry/capability_atoms.json", &atoms); err != nil {
		return coverageAtomRefs{byRecipe: map[string][]string{}, byEvidence: map[string][]string{}}
	}
	refs := coverageAtomRefs{byRecipe: map[string][]string{}, byEvidence: map[string][]string{}}
	for _, atom := range atoms.CapabilityAtoms {
		for _, dependency := range atom.Dependencies {
			path := filepath.ToSlash(dependency.Path)
			switch dependency.Kind {
			case "recipe":
				recipe := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				refs.byRecipe[recipe] = append(refs.byRecipe[recipe], atom.ID)
			case "evidence":
				refs.byEvidence[path] = append(refs.byEvidence[path], atom.ID)
			}
		}
	}
	for recipe := range refs.byRecipe {
		sort.Strings(refs.byRecipe[recipe])
	}
	for evidence := range refs.byEvidence {
		sort.Strings(refs.byEvidence[evidence])
	}
	return refs
}

func (r coverageAtomRefs) atomIDsForRecord(artifact string, recipes []string) []string {
	ids := map[string]bool{}
	for _, id := range r.byEvidence[filepath.ToSlash(artifact)] {
		ids[id] = true
	}
	for _, recipe := range recipes {
		for _, id := range r.byRecipe[recipe] {
			ids[id] = true
		}
	}
	return sortedKeys(ids)
}

func atomRowsForRecord(record coverageRecord, matrix matrixFile, refs coverageAtomRefs) []AtomCoverageRow {
	stats := matrixStatsByRecipe(matrix)
	var rows []AtomCoverageRow
	added := map[string]bool{}
	for _, recipe := range sortedRecipeStatsKeys(stats) {
		for _, atomID := range refs.byRecipe[recipe] {
			stat := stats[recipe]
			rows = append(rows, AtomCoverageRow{
				AtomID:                atomID,
				RecordID:              record.ID,
				Recipe:                recipe,
				Artifact:              filepath.ToSlash(record.Artifact),
				WriterStatus:          record.WriterStatus,
				BoundaryStatus:        boundaryStatus(record, recipe),
				BoundaryReason:        boundaryReason(record, recipe),
				SourceVersions:        sortedKeys(stat.sourceVersions),
				TargetVersions:        sortedKeys(stat.targetVersions),
				AEOpenVersions:        sortedKeys(stat.aeOpenVersions),
				HostOpenEvidenceLevel: hostOpenEvidenceLevel(record, recipe),
				DirectHostVersions:    directHostVersions(record, recipe),
				InferredHostVersions:  inferredHostVersions(record, recipe),
				Totals:                stat.totals,
			})
			added[atomID] = true
		}
	}
	for _, atomID := range refs.byEvidence[filepath.ToSlash(record.Artifact)] {
		if added[atomID] {
			continue
		}
		rows = append(rows, AtomCoverageRow{
			AtomID:                atomID,
			RecordID:              record.ID,
			Artifact:              filepath.ToSlash(record.Artifact),
			WriterStatus:          record.WriterStatus,
			SourceVersions:        matrixSourceVersions(matrix),
			TargetVersions:        matrixTargetVersions(matrix),
			AEOpenVersions:        matrixAEOpenVersions(matrix),
			HostOpenEvidenceLevel: "recorded_status_only",
			Totals:                record.Totals,
		})
	}
	return rows
}

func sortedRecipeStatsKeys(values map[string]recipeMatrixStats) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

type recipeMatrixStats struct {
	sourceVersions map[string]bool
	targetVersions map[string]bool
	aeOpenVersions map[string]bool
	totals         CoverageTotals
}

func matrixStatsByRecipe(matrix matrixFile) map[string]recipeMatrixStats {
	stats := map[string]recipeMatrixStats{}
	for _, c := range matrix.Cases {
		if c.RecipeName == "" {
			continue
		}
		stat := stats[c.RecipeName]
		if stat.sourceVersions == nil {
			stat.sourceVersions = map[string]bool{}
			stat.targetVersions = map[string]bool{}
			stat.aeOpenVersions = map[string]bool{}
		}
		if c.SourceVersion != "" {
			stat.sourceVersions[c.SourceVersion] = true
		}
		if c.TargetVersion != "" {
			stat.targetVersions[c.TargetVersion] = true
		}
		if c.AEOpenVersion != "" {
			stat.aeOpenVersions[c.AEOpenVersion] = true
		}
		stat.totals.Total++
		switch c.Status {
		case "pass":
			stat.totals.Pass++
		case "blocked":
			stat.totals.Blocked++
		case "failed":
			stat.totals.Failed++
		case "skipped":
			stat.totals.Skipped++
		}
		stats[c.RecipeName] = stat
	}
	return stats
}

func hostOpenEvidenceLevel(record coverageRecord, recipe string) string {
	if boundaryStatus(record, recipe) != "" {
		return "excluded_known_boundary"
	}
	if stringSet(record.HostOpenEndpointEvidence.Recipes)[recipe] {
		return "direct_endpoint_hosts_pass"
	}
	if stringSet(record.HostOpenRepresentatives)[recipe] {
		return "direct_all_hosts_representative"
	}
	if strings.Contains(record.HostOpenStatus, "OPEN-ALL-HOSTS") {
		return "direct_all_hosts_record"
	}
	if record.HostOpenStatus != "" {
		return "recorded_status_only"
	}
	return ""
}

func boundaryStatus(record coverageRecord, recipe string) string {
	if recipe == "" || record.Boundary.Status == "" {
		return ""
	}
	if stringSet(record.Boundary.BlockedRecipeIDs)[recipe] {
		return record.Boundary.Status
	}
	return ""
}

func boundaryReason(record coverageRecord, recipe string) string {
	if boundaryStatus(record, recipe) == "" {
		return ""
	}
	for _, detail := range record.Boundary.Details {
		if detail.Recipe == recipe {
			return detail.Reason
		}
	}
	return ""
}

func directHostVersions(record coverageRecord, recipe string) []string {
	if stringSet(record.HostOpenEndpointEvidence.Recipes)[recipe] {
		return sortedStrings(record.HostOpenEndpointEvidence.DirectHosts)
	}
	return nil
}

func inferredHostVersions(record coverageRecord, recipe string) []string {
	if stringSet(record.HostOpenEndpointEvidence.Recipes)[recipe] {
		return sortedStrings(record.HostOpenEndpointEvidence.InferredHosts)
	}
	return nil
}

func countUniqueAtoms(records []CoverageRecordReport, rows []AtomCoverageRow) int {
	ids := map[string]bool{}
	for _, record := range records {
		for _, id := range record.AtomIDs {
			ids[id] = true
		}
	}
	for _, row := range rows {
		ids[row.AtomID] = true
	}
	return len(ids)
}

func matrixRecipes(matrix matrixFile) []string {
	values := map[string]bool{}
	for _, c := range matrix.Cases {
		if c.RecipeName != "" {
			values[c.RecipeName] = true
		}
	}
	return sortedKeys(values)
}

func matrixSourceVersions(matrix matrixFile) []string {
	values := map[string]bool{}
	for _, c := range matrix.Cases {
		if c.SourceVersion != "" {
			values[c.SourceVersion] = true
		}
	}
	return sortedKeys(values)
}

func matrixTargetVersions(matrix matrixFile) []string {
	values := map[string]bool{}
	for _, c := range matrix.Cases {
		if c.TargetVersion != "" {
			values[c.TargetVersion] = true
		}
	}
	return sortedKeys(values)
}

func matrixAEOpenVersions(matrix matrixFile) []string {
	values := map[string]bool{}
	for _, c := range matrix.Cases {
		if c.AEOpenVersion != "" {
			values[c.AEOpenVersion] = true
		}
	}
	return sortedKeys(values)
}

func stringSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	return set
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func readJSONPath(root, path string, value any) error {
	fullPath := path
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(root, filepath.FromSlash(path))
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
