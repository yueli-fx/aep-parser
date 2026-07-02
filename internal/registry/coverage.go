package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CoverageReport struct {
	SchemaVersion int             `json:"schema_version"`
	Status        string          `json:"status"`
	Summary       CoverageSummary `json:"summary"`
	Issues        []CoverageIssue `json:"issues"`
}

type CoverageSummary struct {
	Records   int `json:"records"`
	Artifacts int `json:"artifacts"`
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

type coverageFile struct {
	SchemaVersion  int                `json:"schema_version"`
	RecurringGates []coverageArtifact `json:"recurring_gates"`
	Coverage       []coverageRecord   `json:"coverage"`
}

type coverageRecord struct {
	ID                       string           `json:"id"`
	Artifact                 string           `json:"artifact"`
	Totals                   CoverageTotals   `json:"totals"`
	HostOpenEndpointEvidence coverageEndpoint `json:"host_open_endpoint_evidence"`
}

type coverageEndpoint struct {
	Artifact string             `json:"artifact"`
	Totals   CoverageTotals     `json:"totals"`
	Chunks   []coverageArtifact `json:"chunks"`
}

type coverageArtifact struct {
	ID       string         `json:"id"`
	Command  string         `json:"command"`
	Artifact string         `json:"artifact"`
	Totals   CoverageTotals `json:"totals"`
}

type matrixFile struct {
	Summary matrixSummary `json:"summary"`
}

type matrixSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Blocked int `json:"blocked"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
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
