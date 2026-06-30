package selfhost

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/host"
)

type RecipeDraftSmokeOptions struct {
	DraftsPath string
	CompileDir string
	ReparseDir string
	BatchDir   string
	BatchLimit int
	Runner     host.Runner
	WorkingDir string
}

type RecipeDraftSmokeResult struct {
	Compile RecipeDraftCase           `json:"compile"`
	Reparse RecipeDraftReparseSummary `json:"reparse"`
	Batch   RecipeDraftBatchSummary   `json:"batch"`
}

type RecipeDraftCase struct {
	CaseID             string `json:"case_id,omitempty"`
	ProjectPath        string `json:"project_path,omitempty"`
	Dir                string `json:"dir"`
	Recipe             string `json:"recipe"`
	Output             string `json:"output"`
	OutputBytes        int64  `json:"output_bytes"`
	ValidateJSON       string `json:"validate_json"`
	CompileJSON        string `json:"compile_json"`
	FactsJSON          string `json:"facts_json,omitempty"`
	ReparseSummaryJSON string `json:"reparse_summary_json,omitempty"`
	ExpectedCompCount  int    `json:"expected_comp_count"`
	ActualCompCount    int    `json:"actual_comp_count"`
	ExpectedLayerCount int    `json:"expected_layer_count"`
	ActualLayerCount   int    `json:"actual_layer_count"`
	Passed             bool   `json:"passed"`
}

type RecipeDraftReparseSummary struct {
	SchemaVersion      int    `json:"schema_version"`
	CompiledAEP        string `json:"compiled_aep"`
	FactsJSON          string `json:"facts_json"`
	ExpectedCompCount  int    `json:"expected_comp_count"`
	ActualCompCount    int    `json:"actual_comp_count"`
	ExpectedLayerCount int    `json:"expected_layer_count"`
	ActualLayerCount   int    `json:"actual_layer_count"`
	Passed             bool   `json:"passed"`
}

type RecipeDraftBatchSummary struct {
	SchemaVersion int               `json:"schema_version"`
	Requested     int               `json:"requested"`
	Attempted     int               `json:"attempted"`
	Passed        int               `json:"passed"`
	Cases         []RecipeDraftCase `json:"cases"`
}

type recipeDraftRow struct {
	ProjectPath string          `json:"project_path"`
	Recipe      json.RawMessage `json:"recipe"`
}

type recipeExpectedProfile struct {
	ExpectedProfile struct {
		CompCount  int  `json:"comp_count"`
		LayerCount *int `json:"layer_count"`
	} `json:"expected_profile"`
}

type techniqueFactsSummary struct {
	Summary struct {
		CompCount  int `json:"comp_count"`
		LayerCount int `json:"layer_count"`
	} `json:"summary"`
}

func RunRecipeDraftSmoke(ctx context.Context, opts RecipeDraftSmokeOptions) (RecipeDraftSmokeResult, error) {
	if opts.BatchLimit <= 0 {
		opts.BatchLimit = 3
	}
	if opts.Runner == nil {
		opts.Runner = host.ExecRunner{}
	}
	rows, err := readRecipeDraftRows(opts.DraftsPath, opts.BatchLimit)
	if err != nil {
		return RecipeDraftSmokeResult{}, err
	}
	if len(rows) == 0 {
		return RecipeDraftSmokeResult{}, fmt.Errorf("recipe_drafts.jsonl has no rows: %s", opts.DraftsPath)
	}

	compileCase, err := compileRecipeDraft(ctx, opts, rows[0], opts.CompileDir, "recipe_draft.json", "recipe_draft.aep")
	if err != nil {
		return RecipeDraftSmokeResult{}, err
	}
	reparse, err := reparseRecipeDraft(ctx, opts, compileCase.Output, compileCase.Recipe, opts.ReparseDir, "compiled_facts.json", "reparse_summary.json")
	if err != nil {
		return RecipeDraftSmokeResult{}, err
	}

	batch := RecipeDraftBatchSummary{
		SchemaVersion: 1,
		Requested:     opts.BatchLimit,
		Cases:         make([]RecipeDraftCase, 0, len(rows)),
	}
	for i, row := range rows {
		caseID := fmt.Sprintf("%03d", i+1)
		caseDir := filepath.Join(opts.BatchDir, caseID)
		c, err := compileRecipeDraft(ctx, opts, row, caseDir, "recipe.json", "recipe.aep")
		if err != nil {
			return RecipeDraftSmokeResult{}, err
		}
		c.CaseID = caseID
		c.ProjectPath = row.ProjectPath
		c.FactsJSON = filepath.Join(caseDir, "facts.json")
		c.ReparseSummaryJSON = filepath.Join(caseDir, "reparse_summary.json")
		reparse, err := reparseRecipeDraft(ctx, opts, c.Output, c.Recipe, caseDir, "facts.json", "reparse_summary.json")
		if err != nil {
			return RecipeDraftSmokeResult{}, err
		}
		c.ExpectedCompCount = reparse.ExpectedCompCount
		c.ActualCompCount = reparse.ActualCompCount
		c.ExpectedLayerCount = reparse.ExpectedLayerCount
		c.ActualLayerCount = reparse.ActualLayerCount
		c.Passed = reparse.Passed
		if !c.Passed {
			return RecipeDraftSmokeResult{}, fmt.Errorf("batch recipe draft reparse mismatch for case %s: comps %d/%d layers %d/%d", caseID, c.ActualCompCount, c.ExpectedCompCount, c.ActualLayerCount, c.ExpectedLayerCount)
		}
		batch.Cases = append(batch.Cases, c)
	}
	batch.Attempted = len(batch.Cases)
	for _, c := range batch.Cases {
		if c.Passed {
			batch.Passed++
		}
	}
	if err := writeIndentedJSON(filepath.Join(opts.BatchDir, "summary.json"), batch); err != nil {
		return RecipeDraftSmokeResult{}, err
	}
	if batch.Passed != batch.Attempted {
		return RecipeDraftSmokeResult{}, fmt.Errorf("recipe draft batch smoke passed %d/%d", batch.Passed, batch.Attempted)
	}

	return RecipeDraftSmokeResult{
		Compile: compileCase,
		Reparse: reparse,
		Batch:   batch,
	}, nil
}

func readRecipeDraftRows(path string, limit int) ([]recipeDraftRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rows []recipeDraftRow
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var row recipeDraftRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if len(row.Recipe) == 0 {
			return nil, fmt.Errorf("recipe draft row missing recipe in %s", path)
		}
		rows = append(rows, row)
		if limit > 0 && len(rows) >= limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

func compileRecipeDraft(ctx context.Context, opts RecipeDraftSmokeOptions, row recipeDraftRow, dir, recipeName, outputName string) (RecipeDraftCase, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return RecipeDraftCase{}, err
	}
	recipePath := filepath.Join(dir, recipeName)
	outputPath := filepath.Join(dir, outputName)
	validatePath := filepath.Join(dir, "validate.json")
	compilePath := filepath.Join(dir, "compile.json")
	if err := writeRawIndentedJSON(recipePath, row.Recipe); err != nil {
		return RecipeDraftCase{}, err
	}
	if err := runGoToFile(ctx, opts.Runner, opts.WorkingDir, validatePath, "run", "./cmd/aeprecipe", "validate", "-recipe", recipePath, "-json"); err != nil {
		return RecipeDraftCase{}, err
	}
	if err := runGoToFile(ctx, opts.Runner, opts.WorkingDir, compilePath, "run", "./cmd/aeprecipe", "compile", "-recipe", recipePath, "-out", outputPath, "-json"); err != nil {
		return RecipeDraftCase{}, err
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		return RecipeDraftCase{}, fmt.Errorf("missing compiled recipe draft: %s: %w", outputPath, err)
	}
	return RecipeDraftCase{
		ProjectPath:  row.ProjectPath,
		Dir:          dir,
		Recipe:       recipePath,
		Output:       outputPath,
		OutputBytes:  info.Size(),
		ValidateJSON: validatePath,
		CompileJSON:  compilePath,
	}, nil
}

func reparseRecipeDraft(ctx context.Context, opts RecipeDraftSmokeOptions, compiledAEP, recipePath, dir, factsName, summaryName string) (RecipeDraftReparseSummary, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return RecipeDraftReparseSummary{}, err
	}
	factsPath := filepath.Join(dir, factsName)
	summaryPath := filepath.Join(dir, summaryName)
	if err := runGoToFile(ctx, opts.Runner, opts.WorkingDir, "", "run", "./cmd/aeptechnique", "-in", compiledAEP, "-mode", "facts", "-out", factsPath); err != nil {
		return RecipeDraftReparseSummary{}, err
	}
	expected, err := readExpectedProfile(recipePath)
	if err != nil {
		return RecipeDraftReparseSummary{}, err
	}
	actual, err := readFactsSummary(factsPath)
	if err != nil {
		return RecipeDraftReparseSummary{}, err
	}
	summary := RecipeDraftReparseSummary{
		SchemaVersion:      1,
		CompiledAEP:        compiledAEP,
		FactsJSON:          factsPath,
		ExpectedCompCount:  expected.ExpectedProfile.CompCount,
		ActualCompCount:    actual.Summary.CompCount,
		ExpectedLayerCount: 0,
		ActualLayerCount:   actual.Summary.LayerCount,
	}
	if expected.ExpectedProfile.LayerCount != nil {
		summary.ExpectedLayerCount = *expected.ExpectedProfile.LayerCount
	}
	summary.Passed = summary.ExpectedCompCount == summary.ActualCompCount && summary.ExpectedLayerCount == summary.ActualLayerCount
	if err := writeIndentedJSON(summaryPath, summary); err != nil {
		return RecipeDraftReparseSummary{}, err
	}
	if !summary.Passed {
		return RecipeDraftReparseSummary{}, fmt.Errorf("compiled recipe draft reparse mismatch: comps %d/%d layers %d/%d", summary.ActualCompCount, summary.ExpectedCompCount, summary.ActualLayerCount, summary.ExpectedLayerCount)
	}
	return summary, nil
}

func runGoToFile(ctx context.Context, runner host.Runner, dir string, outPath string, args ...string) error {
	var outFile *os.File
	if outPath != "" {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		file, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer file.Close()
		outFile = file
	}
	cmd := host.Command{
		Name:   "go",
		Args:   args,
		Dir:    dir,
		Stdout: outFile,
		Stderr: outFile,
	}
	result := runner.Run(ctx, cmd)
	if result.ExitCode != 0 {
		return fmt.Errorf("go %s failed with exit code %d", strings.Join(args, " "), result.ExitCode)
	}
	return nil
}

func writeRawIndentedJSON(path string, raw json.RawMessage) error {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	return writeIndentedJSON(path, value)
}

func readExpectedProfile(path string) (recipeExpectedProfile, error) {
	var profile recipeExpectedProfile
	if err := readIndentedJSON(path, &profile); err != nil {
		return recipeExpectedProfile{}, err
	}
	return profile, nil
}

func readFactsSummary(path string) (techniqueFactsSummary, error) {
	var summary techniqueFactsSummary
	if err := readIndentedJSON(path, &summary); err != nil {
		return techniqueFactsSummary{}, err
	}
	return summary, nil
}
