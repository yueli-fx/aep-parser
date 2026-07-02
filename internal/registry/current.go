package registry

import (
	"fmt"
	"path/filepath"
	"sort"
)

type CurrentValidationReport struct {
	SchemaVersion int                      `json:"schema_version"`
	Status        string                   `json:"status"`
	CurrentPath   string                   `json:"current_path"`
	Summary       CurrentValidationSummary `json:"summary"`
	Issues        []CurrentValidationIssue `json:"issues,omitempty"`
}

type CurrentValidationSummary struct {
	CoverageBatches int `json:"coverage_batches"`
	Tooling         int `json:"tooling"`
	Errors          int `json:"errors"`
}

type CurrentValidationIssue struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type currentFile struct {
	TruthSources struct {
		Current          string `json:"current"`
		Coverage         string `json:"coverage"`
		CapabilityLedger string `json:"capability_ledger"`
	} `json:"truth_sources"`
	CurrentState struct {
		LatestRecurringGate        currentMatrixGate           `json:"latest_recurring_gate"`
		LatestExplicitMatteGate    currentMatrixGate           `json:"latest_explicit_matte_gate"`
		LatestStandaloneVerifyGate currentStandaloneVerifyGate `json:"latest_standalone_verify_gate"`
	} `json:"current_state"`
	FrozenMarkdown         []currentPathRef       `json:"frozen_markdown"`
	HostOpenGapAudit       currentHostOpenGap     `json:"host_open_gap_audit"`
	Tooling                []currentTooling       `json:"tooling"`
	NextBatches            []currentNextBatch     `json:"next_batches"`
	CoverageBatches        []currentCoverageBatch `json:"coverage_batches"`
	CanonicalCoverageBatch string                 `json:"canonical_coverage_batch"`
}

type currentPathRef struct {
	Path string `json:"path"`
}

type currentMatrixGate struct {
	Artifact string         `json:"artifact"`
	Totals   CoverageTotals `json:"totals"`
}

type currentStandaloneVerifyGate struct {
	Artifact          string `json:"artifact"`
	Source            string `json:"source"`
	Target            string `json:"target"`
	MigrationReport   string `json:"migration_report"`
	Status            string `json:"status"`
	ProfileDiffStatus string `json:"profile_diff_status"`
	ProfileDiffCount  int    `json:"profile_diff_count"`
}

type currentHostOpenGap struct {
	Script          string                          `json:"script"`
	Artifact        string                          `json:"artifact"`
	Summary         currentHostOpenGapSummary       `json:"summary"`
	CompletedGroups []currentHostOpenCompletedGroup `json:"completed_groups"`
}

type currentHostOpenGapSummary struct {
	GapGroups  int `json:"gap_groups"`
	GapRecipes int `json:"gap_recipes"`
	Chunks     int `json:"chunks"`
}

type currentHostOpenCompletedGroup struct {
	CoverageID string   `json:"coverage_id"`
	Artifact   string   `json:"artifact"`
	Artifacts  []string `json:"artifacts"`
}

type currentTooling struct {
	ID     string `json:"id"`
	Script string `json:"script"`
}

type currentNextBatch struct {
	ID              string `json:"id"`
	Script          string `json:"script"`
	CandidateScript string `json:"candidate_script"`
}

type currentCoverageBatch struct {
	ID          string                      `json:"id"`
	Description string                      `json:"description"`
	Entries     []currentCoverageBatchEntry `json:"entries"`
}

type currentCoverageBatchEntry struct {
	CoverageID      string   `json:"coverage_id"`
	Out             string   `json:"out"`
	Matrix          string   `json:"matrix"`
	Ledger          string   `json:"ledger"`
	RecipeGlob      string   `json:"recipe_glob"`
	RecipePaths     []string `json:"recipe_paths"`
	AllowMatrixExit []int    `json:"allow_matrix_exit"`
}

type currentHostOpenGapReport struct {
	Summary currentHostOpenGapSummary `json:"summary"`
}

func ValidateCurrent(root, currentPath string) (CurrentValidationReport, error) {
	var current currentFile
	if err := readJSONPath(root, currentPath, &current); err != nil {
		return CurrentValidationReport{}, err
	}
	report := CurrentValidationReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		CurrentPath:   filepath.ToSlash(currentPath),
		Summary: CurrentValidationSummary{
			CoverageBatches: len(current.CoverageBatches),
			Tooling:         len(current.Tooling),
		},
	}

	report.checkPath(root, "truth_sources.current", current.TruthSources.Current, true)
	report.checkPath(root, "truth_sources.coverage", current.TruthSources.Coverage, true)
	report.checkPath(root, "truth_sources.capability_ledger", current.TruthSources.CapabilityLedger, false)

	var coverage coverageFile
	if current.TruthSources.Coverage != "" {
		if err := readJSONPath(root, current.TruthSources.Coverage, &coverage); err != nil {
			report.addError("missing_or_invalid_coverage", "truth_sources.coverage", current.TruthSources.Coverage, fmt.Sprintf("read coverage: %v", err))
		}
	}
	coverageIDs := map[string]bool{}
	for _, record := range coverage.Coverage {
		if record.ID != "" {
			coverageIDs[record.ID] = true
		}
	}

	report.checkMatrixGate(root, "current_state.latest_recurring_gate", current.CurrentState.LatestRecurringGate)
	report.checkMatrixGate(root, "current_state.latest_explicit_matte_gate", current.CurrentState.LatestExplicitMatteGate)
	report.checkStandaloneVerify(root, current.CurrentState.LatestStandaloneVerifyGate)
	report.checkFrozenMarkdown(root, current.FrozenMarkdown)
	report.checkHostOpenGapAudit(root, current.HostOpenGapAudit)
	report.checkTooling(root, current.Tooling)
	report.checkNextBatches(root, current.NextBatches)
	report.checkCoverageBatches(root, current.CoverageBatches, coverageIDs)
	report.checkCanonicalCoverageBatch(current.CanonicalCoverageBatch, current.CoverageBatches, coverage.Coverage)

	report.Summary.Errors = len(report.Issues)
	if report.Summary.Errors > 0 {
		report.Status = StatusFail
	}
	return report, nil
}

func (r *CurrentValidationReport) checkPath(root, field, path string, required bool) {
	if path == "" {
		if required {
			r.addError("missing_path", field, "", field+" is empty")
		}
		return
	}
	if !fileExists(root, path) {
		r.addError("missing_path", field, path, field+" not found")
	}
}

func (r *CurrentValidationReport) checkMatrixGate(root, field string, gate currentMatrixGate) {
	if gate.Artifact == "" {
		return
	}
	r.checkPath(root, field+".artifact", gate.Artifact, true)
	var matrix matrixFile
	if err := readJSONPath(root, gate.Artifact, &matrix); err != nil {
		r.addError("missing_or_invalid_matrix_gate", field, gate.Artifact, fmt.Sprintf("read matrix gate: %v", err))
		return
	}
	actual := CoverageTotals{
		Total:   matrix.Summary.Total,
		Pass:    matrix.Summary.Passed,
		Blocked: matrix.Summary.Blocked,
		Failed:  matrix.Summary.Failed,
		Skipped: matrix.Summary.Skipped,
	}
	if gate.Totals != actual {
		r.addError("matrix_gate_totals_mismatch", field+".totals", gate.Artifact, fmt.Sprintf("current totals %+v do not match matrix summary %+v", gate.Totals, actual))
	}
}

func (r *CurrentValidationReport) checkStandaloneVerify(root string, gate currentStandaloneVerifyGate) {
	if gate.Artifact == "" {
		return
	}
	for field, path := range map[string]string{
		"artifact":         gate.Artifact,
		"source":           gate.Source,
		"target":           gate.Target,
		"migration_report": gate.MigrationReport,
	} {
		r.checkPath(root, "current_state.latest_standalone_verify_gate."+field, path, true)
	}
	var report recurringVerifyReport
	if err := readJSONPath(root, gate.Artifact, &report); err != nil {
		r.addError("missing_or_invalid_verify_gate", "current_state.latest_standalone_verify_gate", gate.Artifact, fmt.Sprintf("read verify gate: %v", err))
		return
	}
	if gate.Status != report.Summary.Status {
		r.addError("verify_gate_status_mismatch", "current_state.latest_standalone_verify_gate.status", gate.Artifact, fmt.Sprintf("current status %q does not match report %q", gate.Status, report.Summary.Status))
	}
	if gate.ProfileDiffStatus != report.Verification.ProfileDiffStatus {
		r.addError("verify_gate_profile_status_mismatch", "current_state.latest_standalone_verify_gate.profile_diff_status", gate.Artifact, fmt.Sprintf("current profile status %q does not match report %q", gate.ProfileDiffStatus, report.Verification.ProfileDiffStatus))
	}
	if gate.ProfileDiffCount != report.Verification.ProfileDiffCount {
		r.addError("verify_gate_profile_count_mismatch", "current_state.latest_standalone_verify_gate.profile_diff_count", gate.Artifact, fmt.Sprintf("current profile diff count %d does not match report %d", gate.ProfileDiffCount, report.Verification.ProfileDiffCount))
	}
}

func (r *CurrentValidationReport) checkFrozenMarkdown(root string, docs []currentPathRef) {
	for i, doc := range docs {
		r.checkPath(root, fmt.Sprintf("frozen_markdown.%d.path", i), doc.Path, true)
	}
}

func (r *CurrentValidationReport) checkHostOpenGapAudit(root string, audit currentHostOpenGap) {
	if audit.Script == "" && audit.Artifact == "" && len(audit.CompletedGroups) == 0 {
		return
	}
	r.checkPath(root, "host_open_gap_audit.script", audit.Script, true)
	for _, group := range audit.CompletedGroups {
		field := "host_open_gap_audit.completed_groups." + valueOr(group.CoverageID, "unknown")
		r.checkPath(root, field+".artifact", group.Artifact, false)
		for i, artifact := range group.Artifacts {
			r.checkPath(root, fmt.Sprintf("%s.artifacts.%d", field, i), artifact, true)
		}
	}
	if audit.Artifact == "" || !fileExists(root, audit.Artifact) {
		return
	}
	var actual currentHostOpenGapReport
	if err := readJSONPath(root, audit.Artifact, &actual); err != nil {
		r.addError("missing_or_invalid_host_open_gap_audit", "host_open_gap_audit.artifact", audit.Artifact, fmt.Sprintf("read host-open gap audit: %v", err))
		return
	}
	if audit.Summary != actual.Summary {
		r.addError("host_open_gap_audit_summary_mismatch", "host_open_gap_audit.summary", audit.Artifact, fmt.Sprintf("current summary %+v does not match audit %+v", audit.Summary, actual.Summary))
	}
}

func (r *CurrentValidationReport) checkTooling(root string, tools []currentTooling) {
	for _, tool := range tools {
		if tool.ID == "" {
			r.addError("missing_tooling_id", "tooling", "", "tooling entry missing id")
		}
		r.checkPath(root, "tooling."+valueOr(tool.ID, "unknown")+".script", tool.Script, true)
	}
}

func (r *CurrentValidationReport) checkNextBatches(root string, batches []currentNextBatch) {
	for _, batch := range batches {
		field := "next_batches." + valueOr(batch.ID, "unknown")
		r.checkPath(root, field+".script", batch.Script, false)
		if batch.CandidateScript != "" && !fileExists(root, batch.CandidateScript) {
			r.addError("missing_candidate_script", field+".candidate_script", batch.CandidateScript, "candidate script not found")
		}
	}
}

func (r *CurrentValidationReport) checkCoverageBatches(root string, batches []currentCoverageBatch, coverageIDs map[string]bool) {
	for _, batch := range batches {
		field := "coverage-batches." + valueOr(batch.ID, "unknown")
		if batch.ID == "" {
			r.addError("missing_coverage_batch_id", field, "", "coverage batch missing id")
		}
		if len(batch.Entries) == 0 {
			r.addError("empty_coverage_batch", field, "", "coverage batch has no entries")
			continue
		}
		for _, entry := range batch.Entries {
			entryField := field + ".entries." + valueOr(entry.CoverageID, "unknown")
			if entry.CoverageID == "" {
				r.addError("missing_coverage_batch_entry_id", entryField, "", "coverage batch entry missing coverage_id")
			} else if !coverageIDs[entry.CoverageID] {
				r.addError("unknown_coverage_batch_id", entryField, "", "coverage batch references unknown coverage id")
			}
			r.checkPath(root, entryField+".matrix", entry.Matrix, true)
			r.checkPath(root, entryField+".ledger", entry.Ledger, false)
			r.checkRecipeRefs(root, entryField, entry)
		}
	}
}

func (r *CurrentValidationReport) checkRecipeRefs(root, field string, entry currentCoverageBatchEntry) {
	if entry.RecipeGlob != "" {
		matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(entry.RecipeGlob)))
		if err != nil || len(matches) == 0 {
			r.addError("empty_recipe_glob", field+".recipe_glob", entry.RecipeGlob, "recipe_glob matched no files")
		}
	}
	for _, recipePath := range entry.RecipePaths {
		r.checkPath(root, field+".recipe_paths", recipePath, true)
	}
}

func (r *CurrentValidationReport) checkCanonicalCoverageBatch(canonicalID string, batches []currentCoverageBatch, coverage []coverageRecord) {
	if canonicalID == "" {
		return
	}
	var canonical *currentCoverageBatch
	for i := range batches {
		if batches[i].ID == canonicalID {
			canonical = &batches[i]
			break
		}
	}
	if canonical == nil {
		r.addError("missing_canonical_coverage_batch", "canonical_coverage_batch", "", "canonical coverage batch not found")
		return
	}
	artifactIDs := map[string]bool{}
	for _, record := range coverage {
		if record.Artifact != "" {
			artifactIDs[record.ID] = true
		}
	}
	canonicalIDs := map[string]bool{}
	for _, entry := range canonical.Entries {
		if entry.CoverageID != "" {
			canonicalIDs[entry.CoverageID] = true
		}
	}
	missing := setDifference(artifactIDs, canonicalIDs)
	extra := setDifference(canonicalIDs, artifactIDs)
	if len(missing) > 0 || len(extra) > 0 {
		r.addError("canonical_coverage_batch_mismatch", "coverage-batches."+canonicalID, "", fmt.Sprintf("canonical coverage batch mismatch: missing=%v extra=%v", missing, extra))
	}
}

func setDifference(left, right map[string]bool) []string {
	var values []string
	for value := range left {
		if !right[value] {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	return values
}

func (r *CurrentValidationReport) addError(code, field, path, message string) {
	r.Issues = append(r.Issues, CurrentValidationIssue{
		Code:    code,
		Field:   field,
		Path:    filepath.ToSlash(path),
		Message: message,
	})
}
