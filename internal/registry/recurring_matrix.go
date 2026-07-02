package registry

import (
	"fmt"
	"path/filepath"
)

type RecurringMatrixReport struct {
	SchemaVersion int                    `json:"schema_version"`
	Status        string                 `json:"status"`
	OutRoot       string                 `json:"out_root"`
	Summary       RecurringMatrixSummary `json:"summary"`
	Gates         []RecurringMatrixGate  `json:"gates"`
	Issues        []RecurringMatrixIssue `json:"issues,omitempty"`
}

type RecurringMatrixSummary struct {
	Gates  int `json:"gates"`
	Errors int `json:"errors"`
}

type RecurringMatrixGate struct {
	ID       string         `json:"id"`
	Kind     string         `json:"kind"`
	Artifact string         `json:"artifact"`
	Status   string         `json:"status"`
	Expected CoverageTotals `json:"expected,omitempty"`
	Actual   CoverageTotals `json:"actual,omitempty"`
}

type RecurringMatrixIssue struct {
	Code     string `json:"code"`
	GateID   string `json:"gate_id"`
	Artifact string `json:"artifact,omitempty"`
	Message  string `json:"message"`
}

type recurringVerifyReport struct {
	Summary struct {
		Status string `json:"status"`
	} `json:"summary"`
	Verification struct {
		ProfileDiffStatus string `json:"profile_diff_status"`
		ProfileDiffCount  int    `json:"profile_diff_count"`
	} `json:"verification"`
}

func CheckRecurringMatrixGates(root, outRoot string) (RecurringMatrixReport, error) {
	outRoot = filepath.ToSlash(outRoot)
	report := RecurringMatrixReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		OutRoot:       outRoot,
	}
	report.checkMatrix(root, "full-w2020-no-ae-matrix", pathJoinSlash(outRoot, "smoke_all", "matrix.json"), CoverageTotals{
		Total:   858,
		Pass:    852,
		Blocked: 0,
		Failed:  0,
		Skipped: 6,
	})
	report.checkVerify(root, "standalone-verify-cli-smoke", pathJoinSlash(outRoot, "smoke_all", "minimal-adjustment-layer", "AE2020_to_AE2025", "verify_report.json"))
	report.checkMatrix(root, "explicit-matte-ae2025-contract", pathJoinSlash(outRoot, "explicit_matte_ae2025", "matrix.json"), CoverageTotals{
		Total:   1,
		Pass:    1,
		Blocked: 0,
		Failed:  0,
		Skipped: 0,
	})
	report.Summary.Gates = len(report.Gates)
	report.Summary.Errors = len(report.Issues)
	if report.Summary.Errors > 0 {
		report.Status = StatusFail
	}
	return report, nil
}

func (r *RecurringMatrixReport) checkMatrix(root, id, artifact string, expected CoverageTotals) {
	gate := RecurringMatrixGate{
		ID:       id,
		Kind:     "matrix_summary",
		Artifact: artifact,
		Expected: expected,
		Status:   StatusPass,
	}
	var matrix matrixFile
	if err := readJSONPath(root, artifact, &matrix); err != nil {
		gate.Status = StatusFail
		r.Gates = append(r.Gates, gate)
		r.addIssue("missing_or_invalid_matrix", id, artifact, fmt.Sprintf("read matrix artifact: %v", err))
		return
	}
	gate.Actual = CoverageTotals{
		Total:   matrix.Summary.Total,
		Pass:    matrix.Summary.Passed,
		Blocked: matrix.Summary.Blocked,
		Failed:  matrix.Summary.Failed,
		Skipped: matrix.Summary.Skipped,
	}
	if gate.Actual != expected {
		gate.Status = StatusFail
		r.addIssue("matrix_summary_mismatch", id, artifact, fmt.Sprintf("summary drifted: got %+v want %+v", gate.Actual, expected))
	}
	r.Gates = append(r.Gates, gate)
}

func (r *RecurringMatrixReport) checkVerify(root, id, artifact string) {
	gate := RecurringMatrixGate{
		ID:       id,
		Kind:     "verify_report",
		Artifact: artifact,
		Status:   StatusPass,
	}
	var report recurringVerifyReport
	if err := readJSONPath(root, artifact, &report); err != nil {
		gate.Status = StatusFail
		r.Gates = append(r.Gates, gate)
		r.addIssue("missing_or_invalid_verify_report", id, artifact, fmt.Sprintf("read verify report: %v", err))
		return
	}
	if report.Summary.Status != StatusPass || report.Verification.ProfileDiffStatus != StatusPass || report.Verification.ProfileDiffCount != 0 {
		gate.Status = StatusFail
		r.addIssue("verify_report_mismatch", id, artifact, fmt.Sprintf("verify drifted: status=%s profile_diff_status=%s profile_diff_count=%d", report.Summary.Status, report.Verification.ProfileDiffStatus, report.Verification.ProfileDiffCount))
	}
	r.Gates = append(r.Gates, gate)
}

func (r *RecurringMatrixReport) addIssue(code, gateID, artifact, message string) {
	r.Issues = append(r.Issues, RecurringMatrixIssue{
		Code:     code,
		GateID:   gateID,
		Artifact: artifact,
		Message:  message,
	})
}

func pathJoinSlash(parts ...string) string {
	return filepath.ToSlash(filepath.Join(parts...))
}
