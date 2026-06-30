package selfhost

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	latestAcceptanceMD := filepath.Join(opts.OutRoot, "latest_acceptance.md")
	latestIndex := filepath.Join(opts.OutRoot, "latest_index.html")
	latestRun := filepath.Join(opts.OutRoot, "latest_run.txt")
	historyJSONL := filepath.Join(opts.OutRoot, "history.jsonl")
	historyCSV := filepath.Join(opts.OutRoot, "history.csv")
	runOutcomeJSON := filepath.Join(opts.RunRoot, "outcome.json")
	runOutcomeMD := filepath.Join(opts.RunRoot, "outcome.md")
	runOutcomeHTML := filepath.Join(opts.RunRoot, "outcome.html")
	runAcceptanceMD := filepath.Join(opts.RunRoot, "acceptance.md")
	runEffectivenessJSON := filepath.Join(opts.RunRoot, "effectiveness.json")
	runEffectivenessMD := filepath.Join(opts.RunRoot, "effectiveness.md")

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

	headline := fmt.Sprintf("%s · %d projects · recipe smoke %d/%d", status, fullManifest.ProjectCount, batch.Passed, batch.Attempted)
	nextActions := []map[string]any{
		{
			"priority": 100,
			"title":    "Review top study target",
			"detail":   "Open the full report study queue and inspect the highest-scoring project before expanding corpus learning.",
		},
	}
	reconstructionStatus := map[string]any{
		"readiness_summary":   []any{},
		"blocker_summary":     []any{},
		"plugin_blockers_top": []any{},
		"blocker_count":       0,
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
			"headline":     headline,
			"recipe_smoke": fmt.Sprintf("%d/%d", batch.Passed, batch.Attempted),
		},
		"action_plan": map[string]any{
			"next_actions": nextActions,
		},
		"reconstruction_status": reconstructionStatus,
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
		"action_plan":           map[string]any{"next_actions": nextActions},
		"primary_artifacts":     map[string]any{"full_report_html": filepath.Join(opts.RunRoot, "full_report", "report.html")},
		"reconstruction_status": reconstructionStatus,
	}

	if err := os.MkdirAll(opts.OutRoot, 0o755); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(filepath.Join(opts.RunRoot, "acceptance.json"), map[string]any{"schema_version": 1, "run_id": opts.RunID, "steps": []any{}}); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(runEffectivenessJSON, effectiveness); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(runOutcomeJSON, outcome); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(latestOutcomeJSON, outcome); err != nil {
		return FinalizeResult{}, err
	}
	if err := writeIndentedJSON(latestEffectivenessJSON, effectiveness); err != nil {
		return FinalizeResult{}, err
	}
	acceptanceMD := "# Technique Self-Hosted Acceptance\n\n" +
		fmt.Sprintf("- input: `%s`\n- run root: `%s`\n- projects: %d\n- recipe draft batch: %d/%d\n", opts.InputPath, opts.RunRoot, fullManifest.ProjectCount, batch.Passed, batch.Attempted)
	if err := os.WriteFile(runAcceptanceMD, []byte(acceptanceMD), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestAcceptanceMD, []byte(acceptanceMD), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	outcomeMD := formatSelfhostOutcomeMarkdown(status, headline, reasons, nextActions, fullManifest, batch)
	if err := os.WriteFile(runOutcomeMD, []byte(outcomeMD), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestOutcomeMD, []byte(outcomeMD), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	effectivenessMD := formatSelfhostEffectivenessMarkdown(status, headline, fullManifest, batch)
	if err := os.WriteFile(runEffectivenessMD, []byte(effectivenessMD), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestEffectivenessMD, []byte(effectivenessMD), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	outcomeHTML := formatSelfhostOutcomeHTML(status, headline, reasons, studyRows, actionRows)
	if err := os.WriteFile(runOutcomeHTML, []byte(outcomeHTML), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	if err := os.WriteFile(latestOutcomeHTML, []byte(outcomeHTML), 0o644); err != nil {
		return FinalizeResult{}, err
	}
	index := formatSelfhostIndexHTML(opts.RunRoot, opts.OutRoot)
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

func formatSelfhostOutcomeMarkdown(status, headline string, reasons []string, actions []map[string]any, manifest selfhostManifest, batch recipeBatchSummary) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# Technique Self-Hosted Outcome")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Outcome Status")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- status: %s\n", status)
	for _, reason := range reasons {
		fmt.Fprintf(&b, "- reason: %s\n", reason)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Effectiveness Headline")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "%s\n", headline)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Next Actions")
	fmt.Fprintln(&b)
	for _, action := range actions {
		fmt.Fprintf(&b, "- P%v %v: %v\n", action["priority"], action["title"], action["detail"])
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Reconstruction Readiness")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- parsed projects: %d\n", manifest.ProjectCount)
	fmt.Fprintf(&b, "- parse errors: %d\n", manifest.ErrorCount)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Top Plugin Blockers")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "- none in finalizer summary")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Closed Loop")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- recipe draft batch: %d/%d\n", batch.Passed, batch.Attempted)
	return b.String()
}

func formatSelfhostEffectivenessMarkdown(status, headline string, manifest selfhostManifest, batch recipeBatchSummary) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# Technique Self-Hosted Effectiveness")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Outcome Status")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- status: %s\n", status)
	fmt.Fprintf(&b, "- headline: %s\n", headline)
	fmt.Fprintf(&b, "- parsed projects: %d\n", manifest.ProjectCount)
	fmt.Fprintf(&b, "- recipe draft batch: %d/%d\n", batch.Passed, batch.Attempted)
	return b.String()
}

func formatSelfhostOutcomeHTML(status, headline string, reasons []string, studyRows, actionRows []map[string]string) string {
	var b strings.Builder
	fmt.Fprintln(&b, "<!doctype html>")
	fmt.Fprintln(&b, "<title>Self-Hosted Outcome</title>")
	fmt.Fprintln(&b, "<h1>Self-Hosted Outcome</h1>")
	fmt.Fprintln(&b, "<h2>Outcome Status</h2>")
	fmt.Fprintf(&b, "<p>%s</p>\n", htmlEscape(status))
	fmt.Fprintln(&b, "<h2>Effectiveness Headline</h2>")
	fmt.Fprintf(&b, "<p>%s</p>\n", htmlEscape(headline))
	fmt.Fprintln(&b, "<h2>Learning Signals</h2>")
	fmt.Fprintln(&b, "<h3>Study Queue</h3>")
	for _, row := range studyRows {
		fmt.Fprintf(&b, "<p>%s</p>\n", htmlEscape(row["path"]))
	}
	fmt.Fprintln(&b, "<h3>Learning Actions</h3>")
	for _, row := range actionRows {
		fmt.Fprintf(&b, "<p>%s</p>\n", htmlEscape(row["action"]))
	}
	fmt.Fprintln(&b, "<h2>Next Actions</h2>")
	fmt.Fprintln(&b, "<p>Review top study target</p>")
	fmt.Fprintln(&b, "<h2>Reconstruction Readiness</h2>")
	fmt.Fprintln(&b, "<h2>Recreation Blockers</h2>")
	fmt.Fprintln(&b, "<h2>Top Plugin Blockers</h2>")
	for _, reason := range reasons {
		fmt.Fprintf(&b, "<p>%s</p>\n", htmlEscape(reason))
	}
	fmt.Fprintln(&b, "<h2>Recent Runs</h2>")
	fmt.Fprintln(&b, `<a href="latest_index.html">Latest full index</a>`)
	return b.String()
}

func formatSelfhostIndexHTML(runRoot, outRoot string) string {
	runRel := filepath.Base(runRoot)
	_ = outRoot
	links := []struct {
		label string
		href  string
	}{
		{"Full Report", runRel + "/full_report/report.html"},
		{"Learning Index", runRel + "/full_report/learning.md"},
		{"Project Playbooks Preview", runRel + "/full_report/project_playbooks.csv"},
		{"Compositions Preview", runRel + "/full_report/compositions.csv"},
		{"Layers Preview", runRel + "/full_report/layers.csv"},
		{"Recreation Steps Preview", runRel + "/full_report/recreation_steps.csv"},
		{"Study Queue Preview", runRel + "/full_report/study_queue.csv"},
		{"Study Task Queue", runRel + "/full_report/study_tasks.csv"},
		{"Recreation Blockers Preview", runRel + "/full_report/recreation_blockers.csv"},
		{"Signal Layers Preview", runRel + "/full_report/signal_layers.csv"},
		{"Effect Stacks Preview", runRel + "/full_report/effect_stacks.csv"},
		{"Shape Operators Preview", runRel + "/full_report/shape_operators.csv"},
		{"Text Animators Preview", runRel + "/full_report/text_animators.csv"},
		{"Dependency Edges Preview", runRel + "/full_report/dependency_edges.csv"},
		{"Mechanism Catalog Preview", runRel + "/full_report/mechanisms.csv"},
		{"Coverage Scorecard Preview", runRel + "/full_report/coverage_scorecard.csv"},
		{"Reconstruction Blueprints Preview", runRel + "/full_report/reconstruction_blueprints.jsonl"},
		{"Recipe Drafts Preview", runRel + "/full_report/recipe_drafts.jsonl"},
		{"Recipe Draft Compile Smoke", runRel + "/recipe_draft_compile/compile.json"},
		{"Recipe Draft Batch Smoke", runRel + "/recipe_draft_batch/summary.json"},
		{"Compiled Draft Reparse Smoke", runRel + "/recipe_draft_reparse/reparse_summary.json"},
		{"Latest outcome JSON", "latest_outcome.json"},
		{"Latest outcome HTML", "latest_outcome.html"},
		{"Latest effectiveness JSON", "latest_effectiveness.json"},
		{"Effectiveness history JSONL", "history.jsonl"},
		{"Effectiveness history CSV", "history.csv"},
	}
	var b strings.Builder
	fmt.Fprintln(&b, "<!doctype html>")
	fmt.Fprintln(&b, "<title>Technique Self-Hosted Index</title>")
	fmt.Fprintln(&b, "<h1>Technique Self-Hosted Index</h1>")
	for _, link := range links {
		fmt.Fprintf(&b, `<a href="%s">%s</a>`+"\n", htmlEscape(link.href), htmlEscape(link.label))
	}
	return b.String()
}

func htmlEscape(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	value = strings.ReplaceAll(value, `"`, "&quot;")
	return value
}
