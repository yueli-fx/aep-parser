package selfhost

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type FinalizeOptions struct {
	OutRoot   string
	RunRoot   string
	RunID     string
	InputPath string
	Now       time.Time
}

type FinalizeResult struct {
	OutcomeStatus string `json:"outcome_status"`
	LatestIndex   string `json:"latest_index"`
}

type selfhostManifest struct {
	ProjectCount int `json:"project_count"`
	ErrorCount   int `json:"error_count"`
	PatternCount int `json:"pattern_count"`
}

type recipeCompileReport struct {
	Valid bool `json:"valid"`
}

type recipeBatchSummary struct {
	Requested int `json:"requested"`
	Attempted int `json:"attempted"`
	Passed    int `json:"passed"`
}

type reportCompareSummary struct {
	CountDiffs []CountDiff `json:"count_diffs"`
}

type selfhostHistoryEntry struct {
	SchemaVersion              int    `json:"schema_version"`
	GeneratedAtUTC             string `json:"generated_at_utc"`
	RunID                      string `json:"run_id"`
	RunRoot                    string `json:"run_root"`
	InputPath                  string `json:"input_path"`
	ParsedProjects             int    `json:"parsed_projects"`
	ParseErrors                int    `json:"parse_errors"`
	TechniquePatterns          int    `json:"technique_patterns"`
	PartialErrorGateErrors     int    `json:"partial_error_gate_errors"`
	PartialToFullCountDiffs    int    `json:"partial_to_full_count_diffs"`
	RecipeDraftCompileValid    bool   `json:"recipe_draft_compile_valid"`
	CompiledDraftReparsePassed bool   `json:"compiled_draft_reparse_passed"`
	BatchRequested             int    `json:"batch_requested"`
	BatchAttempted             int    `json:"batch_attempted"`
	BatchPassed                int    `json:"batch_passed"`
	StudyQueueTop              int    `json:"study_queue_top"`
	LearningActionsTop         int    `json:"learning_actions_top"`
	CoverageSummary            int    `json:"coverage_summary"`
}

func FinalizeSelfhostRun(_ context.Context, opts FinalizeOptions) (FinalizeResult, error) {
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	if opts.RunID == "" {
		opts.RunID = filepath.Base(opts.RunRoot)
	}
	generatedAt := opts.Now.UTC().Format(time.RFC3339Nano)

	var fullManifest, partialManifest selfhostManifest
	if err := readIndentedJSON(filepath.Join(opts.RunRoot, "full_report", "manifest.json"), &fullManifest); err != nil {
		return FinalizeResult{}, err
	}
	if err := readIndentedJSON(filepath.Join(opts.RunRoot, "partial_report", "manifest.json"), &partialManifest); err != nil {
		return FinalizeResult{}, err
	}
	var compare reportCompareSummary
	if err := readIndentedJSON(filepath.Join(opts.RunRoot, "compare_partial_to_full", "compare.json"), &compare); err != nil {
		return FinalizeResult{}, err
	}
	var compileReport recipeCompileReport
	if err := readIndentedJSON(filepath.Join(opts.RunRoot, "recipe_draft_compile", "compile.json"), &compileReport); err != nil {
		return FinalizeResult{}, err
	}
	var reparse RecipeDraftReparseSummary
	if err := readIndentedJSON(filepath.Join(opts.RunRoot, "recipe_draft_reparse", "reparse_summary.json"), &reparse); err != nil {
		return FinalizeResult{}, err
	}
	var batch recipeBatchSummary
	if err := readIndentedJSON(filepath.Join(opts.RunRoot, "recipe_draft_batch", "summary.json"), &batch); err != nil {
		return FinalizeResult{}, err
	}
	studyRows, err := readCSVRows(filepath.Join(opts.RunRoot, "full_report", "study_queue.csv"))
	if err != nil {
		return FinalizeResult{}, err
	}
	actionRows, err := readCSVRows(filepath.Join(opts.RunRoot, "full_report", "learning_actions.csv"))
	if err != nil {
		return FinalizeResult{}, err
	}
	coverageRows, err := readCSVRows(filepath.Join(opts.RunRoot, "full_report", "coverage_scorecard.csv"))
	if err != nil {
		return FinalizeResult{}, err
	}

	status := "pass"
	reasons := []string{"self-hosted gate passed with no full-report parse errors and recipe smoke checks passing"}
	if fullManifest.ErrorCount != 0 || !compileReport.Valid || !reparse.Passed || batch.Passed != batch.Attempted {
		status = "fail"
		reasons = nil
		if fullManifest.ErrorCount != 0 {
			reasons = append(reasons, fmt.Sprintf("full report has %d parse error(s)", fullManifest.ErrorCount))
		}
		if !compileReport.Valid {
			reasons = append(reasons, "recipe draft compile validation is not valid")
		}
		if !reparse.Passed {
			reasons = append(reasons, "compiled recipe draft reparse did not match expected counts")
		}
		if batch.Passed != batch.Attempted {
			reasons = append(reasons, fmt.Sprintf("recipe draft batch smoke passed %d/%d", batch.Passed, batch.Attempted))
		}
	}

	latestOutcomeJSON := filepath.Join(opts.OutRoot, "latest_outcome.json")
	latestOutcomeMD := filepath.Join(opts.OutRoot, "latest_outcome.md")
	latestOutcomeHTML := filepath.Join(opts.OutRoot, "latest_outcome.html")
	latestEffectivenessJSON := filepath.Join(opts.OutRoot, "latest_effectiveness.json")
	latestEffectivenessMD := filepath.Join(opts.OutRoot, "latest_effectiveness.md")
	latestIndex := filepath.Join(opts.OutRoot, "latest_index.html")
	latestRun := filepath.Join(opts.OutRoot, "latest_run.txt")
	historyJSONL := filepath.Join(opts.OutRoot, "history.jsonl")
	historyCSV := filepath.Join(opts.OutRoot, "history.csv")

	entry := selfhostHistoryEntry{
		SchemaVersion:              1,
		GeneratedAtUTC:             generatedAt,
		RunID:                      opts.RunID,
		RunRoot:                    opts.RunRoot,
		InputPath:                  opts.InputPath,
		ParsedProjects:             fullManifest.ProjectCount,
		ParseErrors:                fullManifest.ErrorCount,
		TechniquePatterns:          fullManifest.PatternCount,
		PartialErrorGateErrors:     partialManifest.ErrorCount,
		PartialToFullCountDiffs:    len(compare.CountDiffs),
		RecipeDraftCompileValid:    compileReport.Valid,
		CompiledDraftReparsePassed: reparse.Passed,
		BatchRequested:             batch.Requested,
		BatchAttempted:             batch.Attempted,
		BatchPassed:                batch.Passed,
		StudyQueueTop:              len(studyRows),
		LearningActionsTop:         len(actionRows),
		CoverageSummary:            len(coverageRows),
	}

	outcome := map[string]any{
		"schema_version":   1,
		"generated_at_utc": generatedAt,
		"input_path":       opts.InputPath,
		"run_id":           opts.RunID,
		"run_root":         opts.RunRoot,
		"stable_outputs": map[string]any{
			"open_target":               latestOutcomeHTML,
			"latest_outcome":            latestOutcomeMD,
			"latest_outcome_html":       latestOutcomeHTML,
			"latest_outcome_json":       latestOutcomeJSON,
			"latest_effectiveness_json": latestEffectivenessJSON,
			"latest_index":              latestIndex,
		},
		"outcome_status": map[string]any{
			"status":         status,
			"has_regression": status != "pass",
			"reasons":        reasons,
		},
		"outcome_summary": map[string]any{
			"headline":     fmt.Sprintf("%s · %d projects · recipe smoke %d/%d", status, fullManifest.ProjectCount, batch.Passed, batch.Attempted),
			"recipe_smoke": fmt.Sprintf("%d/%d", batch.Passed, batch.Attempted),
		},
		"corpus": map[string]any{
			"parsed_projects":    fullManifest.ProjectCount,
			"parse_errors":       fullManifest.ErrorCount,
			"technique_patterns": fullManifest.PatternCount,
		},
		"closed_loop": map[string]any{
			"batch_passed":    batch.Passed,
			"batch_attempted": batch.Attempted,
		},
	}
	effectiveness := map[string]any{
		"schema_version":   1,
		"generated_at_utc": generatedAt,
		"input_path":       opts.InputPath,
		"run_id":           opts.RunID,
		"run_root":         opts.RunRoot,
		"stable_outputs": map[string]any{
			"open_target":               latestOutcomeHTML,
			"latest_run":                latestRun,
			"latest_outcome":            latestOutcomeMD,
			"latest_outcome_html":       latestOutcomeHTML,
			"latest_outcome_json":       latestOutcomeJSON,
			"latest_effectiveness":      latestEffectivenessMD,
			"latest_effectiveness_json": latestEffectivenessJSON,
			"history_jsonl":             historyJSONL,
			"history_csv":               historyCSV,
			"latest_index":              latestIndex,
		},
		"corpus":                outcome["corpus"],
		"closed_loop":           outcome["closed_loop"],
		"outcome_status":        outcome["outcome_status"],
		"outcome_summary":       outcome["outcome_summary"],
		"learning_signals":      map[string]any{"study_queue_top": studyRows, "learning_actions_top": actionRows, "coverage_summary": coverageRows},
		"history_recent":        []selfhostHistoryEntry{entry},
		"history_delta":         map[string]any{"has_previous": false, "parsed_projects_delta": 0, "technique_patterns_delta": 0, "batch_passed_delta": 0},
		"action_plan":           map[string]any{"next_actions": []any{}},
		"primary_artifacts":     map[string]any{"full_report_html": filepath.Join(opts.RunRoot, "full_report", "report.html")},
		"reconstruction_status": map[string]any{"blocker_count": 0},
	}

	if err := os.MkdirAll(opts.OutRoot, 0o755); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(filepath.Join(opts.RunRoot, "acceptance.json"), map[string]any{"schema_version": 1, "run_id": opts.RunID, "steps": []any{}}); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(filepath.Join(opts.RunRoot, "effectiveness.json"), effectiveness); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(latestOutcomeJSON, outcome); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(latestEffectivenessJSON, effectiveness); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(filepath.Join(opts.RunRoot, "acceptance.md"), []byte("# Technique Self-Hosted Acceptance\n"), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestOutcomeMD, []byte("# Technique Self-Hosted Outcome\n\n- status: "+status+"\n"), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestEffectivenessMD, []byte("# Technique Self-Hosted Effectiveness\n\n## Outcome Status\n"), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestOutcomeHTML, []byte("<!doctype html><title>Technique Self-Hosted Outcome</title><a href=\"latest_index.html\">Latest full index</a>"), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	index := "<!doctype html><title>Technique Self-Hosted Index</title><a href=\"latest_outcome.json\">Latest outcome JSON</a><a href=\"latest_outcome.html\">Latest outcome HTML</a><a href=\"latest_effectiveness.json\">Latest effectiveness JSON</a><a href=\"history.jsonl\">History JSONL</a>"
	if err := os.WriteFile(latestIndex, []byte(index), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestRun, []byte(opts.RunRoot+"\n"), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := appendHistory(historyJSONL, historyCSV, entry); err != nil {
		return FinalizeResult{}, err
	}
	return FinalizeResult{OutcomeStatus: status, LatestIndex: latestIndex}, nil
}

func appendHistory(jsonlPath, csvPath string, entry selfhostHistoryEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(jsonlPath), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(jsonlPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	exists := true
	if _, err := os.Stat(csvPath); err != nil {
		exists = false
	}
	csvFile, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	if !exists {
		if err := writer.Write([]string{"run_id", "parsed_projects", "technique_patterns", "batch_passed", "batch_attempted"}); err != nil {
			return err
		}
	}
	if err := writer.Write([]string{entry.RunID, strconv.Itoa(entry.ParsedProjects), strconv.Itoa(entry.TechniquePatterns), strconv.Itoa(entry.BatchPassed), strconv.Itoa(entry.BatchAttempted)}); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}
