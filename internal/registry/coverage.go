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
	Records       int `json:"records"`
	Artifacts     int `json:"artifacts"`
	ContractGates int `json:"contract_gates"`
	Atoms         int `json:"atoms"`
	Errors        int `json:"errors"`
}

type CoverageSummaryReport struct {
	SchemaVersion           int                     `json:"schema_version"`
	Status                  string                  `json:"status"`
	Records                 int                     `json:"records"`
	Artifacts               int                     `json:"artifacts"`
	ContractGates           int                     `json:"contract_gates"`
	Atoms                   int                     `json:"atoms"`
	AtomRows                int                     `json:"atom_rows"`
	Errors                  int                     `json:"errors"`
	DirectHostAtoms         int                     `json:"direct_host_atoms"`
	InferredHostAtoms       int                     `json:"inferred_host_atoms"`
	ByDomain                []CoverageSummaryBucket `json:"by_domain,omitempty"`
	ByRecord                []CoverageSummaryBucket `json:"by_record,omitempty"`
	ByWriterStatus          []CoverageSummaryBucket `json:"by_writer_status,omitempty"`
	ByHostOpenEvidenceLevel []CoverageSummaryBucket `json:"by_host_open_evidence_level,omitempty"`
	ByBoundaryStatus        []CoverageSummaryBucket `json:"by_boundary_status,omitempty"`
}

type CoverageSummaryBucket struct {
	Name     string `json:"name"`
	AtomRows int    `json:"atom_rows"`
}

type CoverageRowsReport struct {
	SchemaVersion int               `json:"schema_version"`
	Status        string            `json:"status"`
	Count         int               `json:"count"`
	Filter        CoverageRowFilter `json:"filter"`
	Rows          []AtomCoverageRow `json:"rows"`
}

type CoverageCellsReport struct {
	SchemaVersion int                `json:"schema_version"`
	Status        string             `json:"status"`
	Count         int                `json:"count"`
	Filter        CoverageCellFilter `json:"filter"`
	Cells         []CoverageCell     `json:"cells"`
}

type CoverageRowFilter struct {
	RecordID              string `json:"record_id,omitempty"`
	Domain                string `json:"domain,omitempty"`
	AtomID                string `json:"atom_id,omitempty"`
	WriterStatus          string `json:"writer_status,omitempty"`
	HostOpenEvidenceLevel string `json:"host_open_evidence_level,omitempty"`
	BoundaryStatus        string `json:"boundary_status,omitempty"`
}

type CoverageCellFilter struct {
	RecordID              string `json:"record_id,omitempty"`
	Domain                string `json:"domain,omitempty"`
	AtomID                string `json:"atom_id,omitempty"`
	Recipe                string `json:"recipe,omitempty"`
	CaseStatus            string `json:"case_status,omitempty"`
	WriterStatus          string `json:"writer_status,omitempty"`
	HostOpenEvidenceLevel string `json:"host_open_evidence_level,omitempty"`
	BoundaryStatus        string `json:"boundary_status,omitempty"`
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
	Domain                string         `json:"domain,omitempty"`
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

type CoverageCell struct {
	AtomID                string `json:"atom_id"`
	Domain                string `json:"domain,omitempty"`
	RecordID              string `json:"record_id"`
	Recipe                string `json:"recipe,omitempty"`
	Artifact              string `json:"artifact"`
	WriterStatus          string `json:"writer_status,omitempty"`
	BoundaryStatus        string `json:"boundary_status,omitempty"`
	BoundaryReason        string `json:"boundary_reason,omitempty"`
	HostOpenEvidenceLevel string `json:"host_open_evidence_level,omitempty"`
	SourceVersion         string `json:"source_version,omitempty"`
	TargetVersion         string `json:"target_version,omitempty"`
	AEOpenVersion         string `json:"ae_open_version,omitempty"`
	Status                string `json:"status"`
	Reason                string `json:"reason,omitempty"`
}

type coverageFile struct {
	SchemaVersion  int                    `json:"schema_version"`
	RecurringGates []coverageArtifact     `json:"recurring_gates"`
	ContractGates  []coverageContractGate `json:"contract_gates"`
	Coverage       []coverageRecord       `json:"coverage"`
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

type coverageContractGate struct {
	ID       string                      `json:"id"`
	Kind     string                      `json:"kind"`
	Command  string                      `json:"command"`
	Artifact string                      `json:"artifact"`
	Status   string                      `json:"status"`
	Summary  VersionBoundaryCheckSummary `json:"summary,omitempty"`
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
	Reason        string `json:"reason,omitempty"`
}

type coverageArtifactRef struct {
	RecordID string
	Path     string
	Totals   CoverageTotals
}

func ValidateCoverage(root, coveragePath string) (CoverageReport, error) {
	return validateCoverage(root, coveragePath, true)
}

func validateCoverage(root, coveragePath string, checkContractGates bool) (CoverageReport, error) {
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
			Records:       len(coverage.RecurringGates) + len(coverage.Coverage),
			Artifacts:     len(refs),
			ContractGates: len(coverage.ContractGates),
		},
	}
	for _, ref := range refs {
		report.checkArtifact(root, ref)
	}
	if checkContractGates {
		for _, gate := range coverage.ContractGates {
			report.checkContractGate(root, gate)
		}
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

func SummarizeCoverage(report CoverageReport) CoverageSummaryReport {
	summary := CoverageSummaryReport{
		SchemaVersion: 1,
		Status:        report.Status,
		Records:       report.Summary.Records,
		Artifacts:     report.Summary.Artifacts,
		ContractGates: report.Summary.ContractGates,
		Atoms:         report.Summary.Atoms,
		AtomRows:      len(report.AtomRows),
		Errors:        report.Summary.Errors,
	}
	byDomain := map[string]int{}
	byRecord := map[string]int{}
	byWriterStatus := map[string]int{}
	byHostOpenEvidenceLevel := map[string]int{}
	byBoundaryStatus := map[string]int{}
	for _, row := range report.AtomRows {
		incrementBucket(byDomain, row.Domain)
		incrementBucket(byRecord, row.RecordID)
		incrementBucket(byWriterStatus, row.WriterStatus)
		incrementBucket(byHostOpenEvidenceLevel, row.HostOpenEvidenceLevel)
		incrementBucket(byBoundaryStatus, row.BoundaryStatus)
		if len(row.DirectHostVersions) > 0 {
			summary.DirectHostAtoms++
		}
		if len(row.InferredHostVersions) > 0 {
			summary.InferredHostAtoms++
		}
	}
	summary.ByDomain = coverageSummaryBuckets(byDomain)
	summary.ByRecord = coverageSummaryBuckets(byRecord)
	summary.ByWriterStatus = coverageSummaryBuckets(byWriterStatus)
	summary.ByHostOpenEvidenceLevel = coverageSummaryBuckets(byHostOpenEvidenceLevel)
	summary.ByBoundaryStatus = coverageSummaryBuckets(byBoundaryStatus)
	return summary
}

func FilterCoverageRows(report CoverageReport, filter CoverageRowFilter) []AtomCoverageRow {
	var rows []AtomCoverageRow
	for _, row := range report.AtomRows {
		if filter.RecordID != "" && row.RecordID != filter.RecordID {
			continue
		}
		if filter.Domain != "" && row.Domain != filter.Domain {
			continue
		}
		if filter.AtomID != "" && row.AtomID != filter.AtomID {
			continue
		}
		if filter.WriterStatus != "" && row.WriterStatus != filter.WriterStatus {
			continue
		}
		if filter.HostOpenEvidenceLevel != "" && row.HostOpenEvidenceLevel != filter.HostOpenEvidenceLevel {
			continue
		}
		if filter.BoundaryStatus != "" && row.BoundaryStatus != filter.BoundaryStatus {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func CoverageRows(report CoverageReport, filter CoverageRowFilter) CoverageRowsReport {
	rows := FilterCoverageRows(report, filter)
	return CoverageRowsReport{
		SchemaVersion: 1,
		Status:        report.Status,
		Count:         len(rows),
		Filter:        filter,
		Rows:          rows,
	}
}

func CoverageCells(root, coveragePath string, filter CoverageCellFilter) (CoverageCellsReport, error) {
	return coverageCells(root, coveragePath, filter, true)
}

func coverageCells(root, coveragePath string, filter CoverageCellFilter, checkContractGates bool) (CoverageCellsReport, error) {
	report, err := validateCoverage(root, coveragePath, checkContractGates)
	if err != nil {
		return CoverageCellsReport{}, err
	}
	var coverage coverageFile
	if err := readJSONPath(root, coveragePath, &coverage); err != nil {
		return CoverageCellsReport{}, err
	}
	refs := loadCoverageAtomRefs(root)
	cells := coverageCellsForRecords(root, coverage.Coverage, refs, filter)
	return CoverageCellsReport{
		SchemaVersion: 1,
		Status:        report.Status,
		Count:         len(cells),
		Filter:        filter,
		Cells:         cells,
	}, nil
}

func coverageCellsForRecords(root string, records []coverageRecord, refs coverageAtomRefs, filter CoverageCellFilter) []CoverageCell {
	var cells []CoverageCell
	for _, record := range records {
		if record.Artifact == "" {
			continue
		}
		var matrix matrixFile
		if err := readJSONPath(root, record.Artifact, &matrix); err != nil {
			continue
		}
		for _, c := range matrix.Cases {
			if c.RecipeName == "" {
				continue
			}
			atomIDs := refs.byRecipe[c.RecipeName]
			if len(atomIDs) == 0 {
				atomIDs = refs.byEvidence[filepath.ToSlash(record.Artifact)]
			}
			for _, atomID := range atomIDs {
				cell := CoverageCell{
					AtomID:                atomID,
					Domain:                refs.domainFor(atomID),
					RecordID:              record.ID,
					Recipe:                c.RecipeName,
					Artifact:              filepath.ToSlash(record.Artifact),
					WriterStatus:          record.WriterStatus,
					BoundaryStatus:        boundaryStatus(record, c.RecipeName),
					BoundaryReason:        boundaryReason(record, c.RecipeName),
					HostOpenEvidenceLevel: hostOpenEvidenceLevel(record, c.RecipeName),
					SourceVersion:         c.SourceVersion,
					TargetVersion:         c.TargetVersion,
					AEOpenVersion:         c.AEOpenVersion,
					Status:                c.Status,
					Reason:                c.Reason,
				}
				if coverageCellMatches(cell, filter) {
					cells = append(cells, cell)
				}
			}
		}
	}
	sort.Slice(cells, func(i, j int) bool {
		left := coverageCellSortKey(cells[i])
		right := coverageCellSortKey(cells[j])
		for idx := range left {
			if left[idx] != right[idx] {
				return left[idx] < right[idx]
			}
		}
		return false
	})
	return cells
}

func coverageCellMatches(cell CoverageCell, filter CoverageCellFilter) bool {
	if filter.RecordID != "" && cell.RecordID != filter.RecordID {
		return false
	}
	if filter.Domain != "" && cell.Domain != filter.Domain {
		return false
	}
	if filter.AtomID != "" && cell.AtomID != filter.AtomID {
		return false
	}
	if filter.Recipe != "" && cell.Recipe != filter.Recipe {
		return false
	}
	if filter.CaseStatus != "" && cell.Status != filter.CaseStatus {
		return false
	}
	if filter.WriterStatus != "" && cell.WriterStatus != filter.WriterStatus {
		return false
	}
	if filter.HostOpenEvidenceLevel != "" && cell.HostOpenEvidenceLevel != filter.HostOpenEvidenceLevel {
		return false
	}
	if filter.BoundaryStatus != "" && cell.BoundaryStatus != filter.BoundaryStatus {
		return false
	}
	return true
}

func coverageCellSortKey(cell CoverageCell) [7]string {
	return [7]string{
		cell.AtomID,
		cell.RecordID,
		cell.Recipe,
		cell.SourceVersion,
		cell.TargetVersion,
		cell.AEOpenVersion,
		cell.Status,
	}
}

func incrementBucket(buckets map[string]int, name string) {
	if name == "" {
		return
	}
	buckets[name]++
}

func coverageSummaryBuckets(counts map[string]int) []CoverageSummaryBucket {
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	buckets := make([]CoverageSummaryBucket, 0, len(names))
	for _, name := range names {
		buckets = append(buckets, CoverageSummaryBucket{Name: name, AtomRows: counts[name]})
	}
	return buckets
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

func (r *CoverageReport) checkContractGate(root string, gate coverageContractGate) {
	if gate.Artifact == "" {
		r.addError("missing_contract_gate_artifact", gate.ID, "", "contract gate artifact is required")
		return
	}
	switch gate.Kind {
	case "version_boundaries":
		r.checkVersionBoundaryContractGate(root, gate)
	default:
		r.addError("unknown_contract_gate_kind", gate.ID, gate.Artifact, fmt.Sprintf("unknown contract gate kind %q", gate.Kind))
	}
}

func (r *CoverageReport) checkVersionBoundaryContractGate(root string, gate coverageContractGate) {
	var actual VersionBoundaryCheckReport
	if err := readJSONPath(root, gate.Artifact, &actual); err != nil {
		r.addError("missing_or_invalid_contract_gate", gate.ID, gate.Artifact, fmt.Sprintf("read contract gate artifact: %v", err))
		return
	}
	if gate.Status != "" && actual.Status != gate.Status {
		r.addError("contract_gate_status_mismatch", gate.ID, gate.Artifact, fmt.Sprintf("contract gate status %q does not match artifact status %q", gate.Status, actual.Status))
	}
	if gate.Summary != (VersionBoundaryCheckSummary{}) && gate.Summary != actual.Summary {
		r.addError("contract_gate_summary_mismatch", gate.ID, gate.Artifact, fmt.Sprintf("contract gate summary %+v does not match artifact summary %+v", gate.Summary, actual.Summary))
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
	byRecipe       map[string][]string
	byEvidence     map[string][]string
	domainByAtomID map[string]string
}

func loadCoverageAtomRefs(root string) coverageAtomRefs {
	var atoms atomsFile
	if err := readJSONPath(root, "registry/capability_atoms.json", &atoms); err != nil {
		return coverageAtomRefs{byRecipe: map[string][]string{}, byEvidence: map[string][]string{}, domainByAtomID: map[string]string{}}
	}
	refs := coverageAtomRefs{byRecipe: map[string][]string{}, byEvidence: map[string][]string{}, domainByAtomID: map[string]string{}}
	for _, atom := range atoms.CapabilityAtoms {
		if atom.ID != "" {
			refs.domainByAtomID[atom.ID] = atom.Domain
		}
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

func (r coverageAtomRefs) domainFor(atomID string) string {
	return r.domainByAtomID[atomID]
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
				Domain:                refs.domainFor(atomID),
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
			Domain:                refs.domainFor(atomID),
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
