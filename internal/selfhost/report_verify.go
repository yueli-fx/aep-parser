package selfhost

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ReportVerifyOptions struct {
	OutDir      string
	MinProjects int
}

type ReportVerifyResult struct {
	ProjectCount int `json:"project_count"`
	PatternCount int `json:"pattern_count"`
}

type reportVerifySummary struct {
	ProjectCount int `json:"project_count"`
	ErrorCount   int `json:"error_count"`
	Totals       struct {
		CompCount          int `json:"comp_count"`
		LayerCount         int `json:"layer_count"`
		EffectCount        int `json:"effect_count"`
		ShapeOperatorCount int `json:"shape_operator_count"`
		TextAnimatorCount  int `json:"text_animator_count"`
		DependencyCount    int `json:"dependency_count"`
	} `json:"totals"`
}

type reportVerifyDigest struct {
	ProjectCount int              `json:"project_count"`
	Patterns     []map[string]any `json:"patterns"`
}

type reportVerifyManifest struct {
	ProjectCount int   `json:"project_count"`
	ErrorCount   int   `json:"error_count"`
	Artifacts    []any `json:"artifacts"`
	Git          struct {
		Commit string `json:"commit"`
	} `json:"git"`
}

func VerifyTechniqueReport(opts ReportVerifyOptions) (ReportVerifyResult, error) {
	if opts.MinProjects <= 0 {
		opts.MinProjects = 1
	}
	paths := reportArtifactPaths(opts.OutDir)
	for _, path := range paths.all {
		if _, err := os.Stat(path); err != nil {
			return ReportVerifyResult{}, fmt.Errorf("missing report artifact: %s", path)
		}
	}

	var summary reportVerifySummary
	if err := readIndentedJSON(paths.summary, &summary); err != nil {
		return ReportVerifyResult{}, err
	}
	var digest reportVerifyDigest
	if err := readIndentedJSON(paths.digest, &digest); err != nil {
		return ReportVerifyResult{}, err
	}
	var manifest reportVerifyManifest
	if err := readIndentedJSON(paths.manifest, &manifest); err != nil {
		return ReportVerifyResult{}, err
	}

	corpusLines, corpusRecords, err := readJSONLines(paths.corpus)
	if err != nil {
		return ReportVerifyResult{}, err
	}
	csvRows := map[string][]map[string]string{}
	for name, path := range paths.csvs {
		rows, err := readCSVRows(path)
		if err != nil {
			return ReportVerifyResult{}, err
		}
		csvRows[name] = rows
	}
	blueprints, err := readJSONLObjects(paths.reconstructionBlueprints)
	if err != nil {
		return ReportVerifyResult{}, err
	}
	recipeDrafts, err := readJSONLObjects(paths.recipeDrafts)
	if err != nil {
		return ReportVerifyResult{}, err
	}

	if summary.ProjectCount < opts.MinProjects {
		return ReportVerifyResult{}, fmt.Errorf("project_count %d is lower than MinProjects %d", summary.ProjectCount, opts.MinProjects)
	}
	expectedCorpusLines := summary.ProjectCount + summary.ErrorCount
	if len(corpusLines) != expectedCorpusLines {
		return ReportVerifyResult{}, fmt.Errorf("corpus line count %d does not match project_count + error_count %d", len(corpusLines), expectedCorpusLines)
	}
	if digest.ProjectCount != summary.ProjectCount {
		return ReportVerifyResult{}, fmt.Errorf("digest project_count %d does not match summary project_count %d", digest.ProjectCount, summary.ProjectCount)
	}
	if manifest.ProjectCount != summary.ProjectCount {
		return ReportVerifyResult{}, fmt.Errorf("manifest project_count %d does not match summary project_count %d", manifest.ProjectCount, summary.ProjectCount)
	}
	if manifest.ErrorCount != summary.ErrorCount {
		return ReportVerifyResult{}, fmt.Errorf("manifest error_count %d does not match summary error_count %d", manifest.ErrorCount, summary.ErrorCount)
	}
	if len(manifest.Artifacts) < 8 {
		return ReportVerifyResult{}, fmt.Errorf("manifest has no artifact inventory")
	}
	if manifest.Git.Commit == "" {
		return ReportVerifyResult{}, fmt.Errorf("manifest has no git commit")
	}
	if len(digest.Patterns) == 0 {
		return ReportVerifyResult{}, fmt.Errorf("digest has no pattern entries")
	}
	if err := verifyManifestArtifacts(opts.OutDir, manifest.Artifacts); err != nil {
		return ReportVerifyResult{}, err
	}
	if err := verifyEffectFieldReportSurface(opts.OutDir, manifest.Artifacts); err != nil {
		return ReportVerifyResult{}, err
	}

	if err := verifyReportRows(summary, digest, corpusRecords, csvRows, blueprints, recipeDrafts); err != nil {
		return ReportVerifyResult{}, err
	}
	if err := requireReportTexts(paths); err != nil {
		return ReportVerifyResult{}, err
	}

	return ReportVerifyResult{ProjectCount: summary.ProjectCount, PatternCount: len(digest.Patterns)}, nil
}

func verifyManifestArtifacts(outDir string, artifacts []any) error {
	for _, artifact := range artifacts {
		name := stringAny(artifact)
		if name == "" {
			return fmt.Errorf("manifest contains invalid artifact entry: %s", compactJSON(artifact))
		}
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			return fmt.Errorf("missing manifest artifact %s: %w", name, err)
		}
	}
	return nil
}

func verifyEffectFieldReportSurface(outDir string, artifacts []any) error {
	hasSummary := manifestArtifactExists(artifacts, "effect_field_summary.json")
	hasQueue := manifestArtifactExists(artifacts, "effect_field_study_queue.csv")
	if !hasSummary && !hasQueue {
		return nil
	}
	if hasSummary != hasQueue {
		return fmt.Errorf("effect field report surface must include both summary and study queue artifacts")
	}
	var summary effectFieldReportSummary
	if err := readIndentedJSON(filepath.Join(outDir, "effect_field_summary.json"), &summary); err != nil {
		return fmt.Errorf("effect_field_summary.json: %w", err)
	}
	if summary.SchemaVersion != 1 {
		return fmt.Errorf("effect_field_summary.json schema_version %d, want 1", summary.SchemaVersion)
	}
	if summary.Summary.EffectKinds <= 0 {
		return fmt.Errorf("effect_field_summary.json has no effect kinds")
	}
	if len(summary.TopStudyTargets) == 0 {
		return fmt.Errorf("effect_field_summary.json has no top study targets")
	}
	rows, err := readCSVRows(filepath.Join(outDir, "effect_field_study_queue.csv"))
	if err != nil {
		return fmt.Errorf("effect_field_study_queue.csv: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("effect_field_study_queue.csv has no rows")
	}
	for _, row := range rows {
		if missing(row, "match_name", "class", "reproducibility", "generation_policy", "study_priority") {
			return fmt.Errorf("effect_field_study_queue.csv contains incomplete row: %s", compactJSON(row))
		}
		if intRow(row, "study_priority") <= 0 {
			return fmt.Errorf("effect_field_study_queue.csv row has invalid study_priority: %s", compactJSON(row))
		}
	}
	return nil
}

func manifestArtifactExists(artifacts []any, name string) bool {
	for _, artifact := range artifacts {
		if stringAny(artifact) == name {
			return true
		}
	}
	return false
}

func verifyReportRows(summary reportVerifySummary, digest reportVerifyDigest, corpusRecords []map[string]any, rows map[string][]map[string]string, blueprints, recipeDrafts []map[string]any) error {
	if len(rows["projects"]) != summary.ProjectCount {
		return fmt.Errorf("projects.csv row count %d does not match summary project_count %d", len(rows["projects"]), summary.ProjectCount)
	}
	if len(rows["project_playbooks"]) != summary.ProjectCount {
		return fmt.Errorf("project_playbooks.csv row count %d does not match summary project_count %d", len(rows["project_playbooks"]), summary.ProjectCount)
	}
	for _, row := range rows["project_playbooks"] {
		if missing(row, "project_path", "readiness", "overview", "ordered_steps") {
			return fmt.Errorf("project_playbooks.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	if len(rows["compositions"]) != summary.Totals.CompCount {
		return fmt.Errorf("compositions.csv row count %d does not match summary totals.comp_count %d", len(rows["compositions"]), summary.Totals.CompCount)
	}
	for _, row := range rows["compositions"] {
		if missing(row, "project_path", "name") || intRow(row, "width") <= 0 || intRow(row, "height") <= 0 || intRow(row, "layer_count") < 0 {
			return fmt.Errorf("compositions.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	if len(rows["layers"]) != summary.Totals.LayerCount {
		return fmt.Errorf("layers.csv row count %d does not match summary totals.layer_count %d", len(rows["layers"]), summary.Totals.LayerCount)
	}
	for _, row := range rows["layers"] {
		if missing(row, "project_path", "comp_name", "type", "role") || intRow(row, "index") < 0 {
			return fmt.Errorf("layers.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	expectedSteps := 0
	for _, record := range corpusRecords {
		expectedSteps += len(nestedSlice(record, "explanation", "recreation_steps"))
	}
	if len(rows["recreation_steps"]) != expectedSteps {
		return fmt.Errorf("recreation_steps.csv row count %d does not match corpus recreation step count %d", len(rows["recreation_steps"]), expectedSteps)
	}
	for _, row := range rows["recreation_steps"] {
		if missing(row, "project_path", "step_id", "title", "summary") || intRow(row, "priority") <= 0 {
			return fmt.Errorf("recreation_steps.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	if len(rows["study_queue"]) != summary.ProjectCount {
		return fmt.Errorf("study_queue.csv row count %d does not match summary project_count %d", len(rows["study_queue"]), summary.ProjectCount)
	}
	if len(rows["study_tasks"]) == 0 {
		return fmt.Errorf("study_tasks.csv has no rows")
	}
	for _, row := range rows["study_tasks"] {
		if missing(row, "project_path", "focus", "action") || intRow(row, "rank") <= 0 {
			return fmt.Errorf("study_tasks.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	for _, row := range rows["recreation_blockers"] {
		if missing(row, "project_path", "blocker_type", "blocker", "action") {
			return fmt.Errorf("recreation_blockers.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	projectsWithBlockers := 0
	for _, row := range rows["projects"] {
		if strings.TrimSpace(row["readiness_blockers"]) != "" {
			projectsWithBlockers++
		}
	}
	if projectsWithBlockers > 0 && len(rows["recreation_blockers"]) == 0 {
		return fmt.Errorf("recreation_blockers.csv has no rows but projects.csv reports readiness blockers")
	}
	if len(rows["signal_layers"]) == 0 {
		return fmt.Errorf("signal_layers.csv has no rows")
	}
	for _, row := range rows["signal_layers"] {
		if missing(row, "project_path", "layer_name", "role", "signals") || intRow(row, "score") <= 0 {
			return fmt.Errorf("signal_layers.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	if err := verifyCountedRows("effect_stacks.csv", rows["effect_stacks"], summary.Totals.EffectCount, "project_path", "comp_name", "match_name"); err != nil {
		return err
	}
	if err := verifyCountedRows("shape_operators.csv", rows["shape_operators"], summary.Totals.ShapeOperatorCount, "project_path", "comp_name", "layer_name", "family", "source"); err != nil {
		return err
	}
	if err := verifyCountedRows("text_animators.csv", rows["text_animators"], summary.Totals.TextAnimatorCount, "project_path", "comp_name", "layer_name", "property_kind", "match_name"); err != nil {
		return err
	}
	if err := verifyCountedRows("dependency_edges.csv", rows["dependency_edges"], summary.Totals.DependencyCount, "project_path", "comp_name", "relation"); err != nil {
		return err
	}
	if len(rows["errors"]) != summary.ErrorCount {
		return fmt.Errorf("errors.csv row count %d does not match summary error_count %d", len(rows["errors"]), summary.ErrorCount)
	}
	if len(rows["patterns"]) != len(digest.Patterns) {
		return fmt.Errorf("patterns.csv row count %d does not match digest pattern count %d", len(rows["patterns"]), len(digest.Patterns))
	}
	if len(rows["learning_actions"]) != len(digest.Patterns) {
		return fmt.Errorf("learning_actions.csv row count %d does not match digest pattern count %d", len(rows["learning_actions"]), len(digest.Patterns))
	}
	for _, row := range rows["learning_actions"] {
		if missing(row, "pattern", "action", "representative_project") {
			return fmt.Errorf("learning_actions.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	if err := verifyNonEmptyRows("mechanisms.csv", rows["mechanisms"], "category", "name", "action"); err != nil {
		return err
	}
	for _, row := range rows["mechanisms"] {
		if intRow(row, "count") <= 0 {
			return fmt.Errorf("mechanisms.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	if err := verifyNonEmptyRows("mechanism_examples.csv", rows["mechanism_examples"], "category", "name", "project_path", "action"); err != nil {
		return err
	}
	for _, row := range rows["mechanism_examples"] {
		if intRow(row, "project_count") <= 0 {
			return fmt.Errorf("mechanism_examples.csv contains incomplete row: %s", compactJSON(row))
		}
	}
	if err := verifyCoverage(rows["coverage_scorecard"]); err != nil {
		return err
	}
	if err := verifyBlueprints(summary.ProjectCount, blueprints); err != nil {
		return err
	}
	if err := verifyRecipeDrafts(summary.ProjectCount, recipeDrafts); err != nil {
		return err
	}
	if summary.ProjectCount > 0 && expectedSteps == 0 {
		return fmt.Errorf("corpus has no explanation.recreation_steps entries")
	}
	patternsWithSteps := 0
	for _, pattern := range digest.Patterns {
		if len(anySlice(pattern["recreation_steps"])) > 0 {
			patternsWithSteps++
		}
	}
	if len(digest.Patterns) > 0 && patternsWithSteps == 0 {
		return fmt.Errorf("digest patterns have no recreation_steps entries")
	}
	return nil
}

func verifyCountedRows(name string, rows []map[string]string, expected int, required ...string) error {
	if len(rows) != expected {
		return fmt.Errorf("%s row count %d does not match summary totals.%s %d", name, len(rows), strings.TrimSuffix(strings.ReplaceAll(name, ".csv", ""), "s"), expected)
	}
	for _, row := range rows {
		if missing(row, required...) {
			return fmt.Errorf("%s contains incomplete row: %s", name, compactJSON(row))
		}
	}
	return nil
}

func verifyNonEmptyRows(name string, rows []map[string]string, required ...string) error {
	if len(rows) == 0 {
		return fmt.Errorf("%s has no rows", name)
	}
	for _, row := range rows {
		if missing(row, required...) {
			return fmt.Errorf("%s contains incomplete row: %s", name, compactJSON(row))
		}
	}
	return nil
}

func verifyCoverage(rows []map[string]string) error {
	if len(rows) == 0 {
		return fmt.Errorf("coverage_scorecard.csv has no rows")
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if missing(row, "artifact", "metric", "status") {
			return fmt.Errorf("coverage_scorecard.csv contains incomplete row: %s", compactJSON(row))
		}
		if row["status"] != "ok" {
			return fmt.Errorf("coverage_scorecard.csv contains non-ok row: %s", compactJSON(row))
		}
		if intRow(row, "expected_count") != intRow(row, "actual_count") {
			return fmt.Errorf("coverage_scorecard.csv count mismatch: %s", compactJSON(row))
		}
		seen[row["artifact"]] = true
	}
	for _, artifact := range []string{"projects.csv", "project_playbooks.csv", "compositions.csv", "layers.csv", "recreation_steps.csv", "effect_stacks.csv", "shape_operators.csv", "text_animators.csv", "dependency_edges.csv", "reconstruction_blueprints.jsonl", "recipe_drafts.jsonl"} {
		if !seen[artifact] {
			return fmt.Errorf("coverage_scorecard.csv missing artifact row: %s", artifact)
		}
	}
	return nil
}

func verifyBlueprints(projectCount int, rows []map[string]any) error {
	if len(rows) != projectCount {
		return fmt.Errorf("reconstruction_blueprints.jsonl line count %d does not match summary project_count %d", len(rows), projectCount)
	}
	required := map[string]bool{"create_compositions": true, "create_layers": true, "apply_mechanisms": true, "wire_dependencies": true, "verify_recreation": true}
	for _, row := range rows {
		if stringAny(row["project_path"]) == "" || stringAny(row["readiness"]) == "" || row["counts"] == nil || row["phases"] == nil {
			return fmt.Errorf("reconstruction_blueprints.jsonl contains incomplete row: %s", compactJSON(row))
		}
		phases := anySlice(row["phases"])
		if len(phases) < 5 {
			return fmt.Errorf("reconstruction_blueprints.jsonl row has too few phases: %s", compactJSON(row))
		}
		seen := map[string]bool{}
		for _, phase := range phases {
			if obj, ok := phase.(map[string]any); ok {
				seen[stringAny(obj["id"])] = true
			}
		}
		for id := range required {
			if !seen[id] {
				return fmt.Errorf("reconstruction_blueprints.jsonl row missing phase %s: %s", id, stringAny(row["project_path"]))
			}
		}
	}
	return nil
}

func verifyRecipeDrafts(projectCount int, rows []map[string]any) error {
	if len(rows) != projectCount {
		return fmt.Errorf("recipe_drafts.jsonl line count %d does not match summary project_count %d", len(rows), projectCount)
	}
	for _, row := range rows {
		if stringAny(row["project_path"]) == "" || stringAny(row["readiness"]) == "" || row["counts"] == nil || row["recipe"] == nil || row["gaps"] == nil {
			return fmt.Errorf("recipe_drafts.jsonl contains incomplete row: %s", compactJSON(row))
		}
		recipe, _ := row["recipe"].(map[string]any)
		comps := anySlice(recipe["comps"])
		expected, _ := recipe["expected_profile"].(map[string]any)
		if intAny(recipe["schema_version"]) != 1 || recipe["project"] == nil || recipe["expected_profile"] == nil {
			return fmt.Errorf("recipe_drafts.jsonl row has incomplete recipe draft: %s", compactJSON(row))
		}
		if len(comps) <= 0 {
			return fmt.Errorf("recipe_drafts.jsonl row has no comps: %s", stringAny(row["project_path"]))
		}
		if intAny(expected["comp_count"]) != len(comps) {
			return fmt.Errorf("recipe_drafts.jsonl expected_profile comp_count mismatch: %s", stringAny(row["project_path"]))
		}
		if expected["layer_count"] != nil || expected["text_layer_count"] != nil || expected["shape_layer_count"] != nil {
			return fmt.Errorf("recipe_drafts.jsonl recipe expected_profile must only claim materialized fields: %s", stringAny(row["project_path"]))
		}
		if len(anySlice(row["gaps"])) <= 0 {
			return fmt.Errorf("recipe_drafts.jsonl row has no explicit gaps: %s", stringAny(row["project_path"]))
		}
	}
	return nil
}

func requireReportTexts(paths reportPaths) error {
	for _, req := range []struct {
		path    string
		pattern string
	}{
		{paths.learning, `^## Pattern Playbook$`},
		{paths.learning, `^## Study Queue$`},
		{paths.learning, `recreation steps`},
		{paths.learning, `^## Plugin Risk Queue$`},
		{paths.learning, `^## Readiness Queue$`},
		{paths.report, `^## Pattern Representatives$`},
		{paths.report, `^## Study Queue$`},
		{paths.report, `Step [0-9]+:`},
	} {
		if err := requireText(req.path, req.pattern); err != nil {
			return err
		}
	}
	for _, pattern := range []string{
		`Technique Corpus Dashboard`,
		`data-tab="overview"`,
		`data-tab="patterns"`,
		`data-tab="projects"`,
		`data-tab="recipes"`,
		`data-tab="data"`,
		`id="project-search"`,
		`id="project-detail"`,
		`window\.__TECHNIQUE_REPORT__`,
		`summary\.json`,
		`corpus\.jsonl`,
		`digest\.json`,
		`projects\.csv`,
		`project_playbooks\.csv`,
		`compositions\.csv`,
		`layers\.csv`,
		`recreation_steps\.csv`,
		`study_queue\.csv`,
		`study_tasks\.csv`,
		`recreation_blockers\.csv`,
		`signal_layers\.csv`,
		`effect_stacks\.csv`,
		`shape_operators\.csv`,
		`text_animators\.csv`,
		`dependency_edges\.csv`,
		`learning_actions\.csv`,
		`mechanisms\.csv`,
		`mechanism_examples\.csv`,
		`coverage_scorecard\.csv`,
		`reconstruction_blueprints\.jsonl`,
		`recipe_drafts\.jsonl`,
		`errors\.csv`,
	} {
		if err := requireText(paths.html, pattern); err != nil {
			return err
		}
	}
	return nil
}

type reportPaths struct {
	summary                  string
	corpus                   string
	digest                   string
	learning                 string
	reconstructionBlueprints string
	recipeDrafts             string
	manifest                 string
	report                   string
	html                     string
	csvs                     map[string]string
	all                      []string
}

func reportArtifactPaths(outDir string) reportPaths {
	csvs := map[string]string{
		"projects":            filepath.Join(outDir, "projects.csv"),
		"project_playbooks":   filepath.Join(outDir, "project_playbooks.csv"),
		"compositions":        filepath.Join(outDir, "compositions.csv"),
		"layers":              filepath.Join(outDir, "layers.csv"),
		"recreation_steps":    filepath.Join(outDir, "recreation_steps.csv"),
		"patterns":            filepath.Join(outDir, "patterns.csv"),
		"study_queue":         filepath.Join(outDir, "study_queue.csv"),
		"study_tasks":         filepath.Join(outDir, "study_tasks.csv"),
		"recreation_blockers": filepath.Join(outDir, "recreation_blockers.csv"),
		"signal_layers":       filepath.Join(outDir, "signal_layers.csv"),
		"effect_stacks":       filepath.Join(outDir, "effect_stacks.csv"),
		"shape_operators":     filepath.Join(outDir, "shape_operators.csv"),
		"text_animators":      filepath.Join(outDir, "text_animators.csv"),
		"dependency_edges":    filepath.Join(outDir, "dependency_edges.csv"),
		"learning_actions":    filepath.Join(outDir, "learning_actions.csv"),
		"mechanisms":          filepath.Join(outDir, "mechanisms.csv"),
		"mechanism_examples":  filepath.Join(outDir, "mechanism_examples.csv"),
		"coverage_scorecard":  filepath.Join(outDir, "coverage_scorecard.csv"),
		"errors":              filepath.Join(outDir, "errors.csv"),
	}
	paths := reportPaths{
		summary:                  filepath.Join(outDir, "summary.json"),
		corpus:                   filepath.Join(outDir, "corpus.jsonl"),
		digest:                   filepath.Join(outDir, "digest.json"),
		learning:                 filepath.Join(outDir, "learning.md"),
		reconstructionBlueprints: filepath.Join(outDir, "reconstruction_blueprints.jsonl"),
		recipeDrafts:             filepath.Join(outDir, "recipe_drafts.jsonl"),
		manifest:                 filepath.Join(outDir, "manifest.json"),
		report:                   filepath.Join(outDir, "report.md"),
		html:                     filepath.Join(outDir, "report.html"),
		csvs:                     csvs,
	}
	paths.all = []string{paths.summary, paths.corpus, paths.digest, paths.learning, paths.reconstructionBlueprints, paths.recipeDrafts, paths.manifest, paths.report, paths.html}
	for _, path := range csvs {
		paths.all = append(paths.all, path)
	}
	return paths
}

func readCSVRows(path string) ([]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	header := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		row := map[string]string{}
		for i, name := range header {
			if i < len(record) {
				row[name] = record[i]
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func readJSONLines(path string) ([]string, []map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var lines []string
	var rows []map[string]any
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, nil, err
		}
		lines = append(lines, line)
		rows = append(rows, row)
	}
	return lines, rows, nil
}

func readJSONLObjects(path string) ([]map[string]any, error) {
	_, rows, err := readJSONLines(path)
	return rows, err
}

func requireText(path, pattern string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	matched, err := regexp.MatchString("(?m)"+pattern, text)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("missing pattern %q in %s", pattern, path)
	}
	return nil
}

func missing(row map[string]string, fields ...string) bool {
	for _, field := range fields {
		if strings.TrimSpace(row[field]) == "" {
			return true
		}
	}
	return false
}

func intRow(row map[string]string, field string) int {
	value, _ := strconv.Atoi(strings.TrimSpace(row[field]))
	return value
}

func intAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	default:
		return 0
	}
}

func stringAny(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func anySlice(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	default:
		return nil
	}
}

func nestedSlice(row map[string]any, first, second string) []any {
	child, ok := row[first].(map[string]any)
	if !ok {
		return nil
	}
	return anySlice(child[second])
}

func compactJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}
