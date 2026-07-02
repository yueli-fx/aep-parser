package registry

import (
	"fmt"
	"path/filepath"
)

type CoverageBatchListReport struct {
	SchemaVersion int                    `json:"schema_version"`
	Status        string                 `json:"status"`
	CurrentPath   string                 `json:"current_path"`
	Summary       CoverageBatchSummary   `json:"summary"`
	Batches       []CoverageBatchListRow `json:"batches"`
}

type CoverageBatchReport struct {
	SchemaVersion int                  `json:"schema_version"`
	Status        string               `json:"status"`
	CurrentPath   string               `json:"current_path"`
	CoveragePath  string               `json:"coverage_path"`
	BatchID       string               `json:"batch_id"`
	Summary       CoverageBatchSummary `json:"summary"`
	Entries       []CoverageBatchEntry `json:"entries,omitempty"`
	Issues        []CoverageBatchIssue `json:"issues,omitempty"`
}

type CoverageBatchSummary struct {
	Batches int `json:"batches,omitempty"`
	Entries int `json:"entries"`
	Errors  int `json:"errors"`
}

type CoverageBatchListRow struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Entries     int    `json:"entries"`
}

type CoverageBatchEntry struct {
	CoverageID string         `json:"coverage_id"`
	Matrix     string         `json:"matrix"`
	Ledger     string         `json:"ledger,omitempty"`
	Recipes    []string       `json:"recipes,omitempty"`
	Expected   CoverageTotals `json:"expected"`
	Actual     CoverageTotals `json:"actual"`
	Status     string         `json:"status"`
}

type CoverageBatchIssue struct {
	Code       string `json:"code"`
	CoverageID string `json:"coverage_id,omitempty"`
	Path       string `json:"path,omitempty"`
	Message    string `json:"message"`
}

func ListCoverageBatches(root, currentPath string) (CoverageBatchListReport, error) {
	var current currentFile
	if err := readJSONPath(root, currentPath, &current); err != nil {
		return CoverageBatchListReport{}, err
	}
	report := CoverageBatchListReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		CurrentPath:   filepath.ToSlash(currentPath),
		Summary: CoverageBatchSummary{
			Batches: len(current.CoverageBatches),
		},
	}
	for _, batch := range current.CoverageBatches {
		report.Summary.Entries += len(batch.Entries)
		report.Batches = append(report.Batches, CoverageBatchListRow{
			ID:          batch.ID,
			Description: batch.Description,
			Entries:     len(batch.Entries),
		})
	}
	return report, nil
}

func CheckCoverageBatch(root, currentPath, coveragePath, batchID string) (CoverageBatchReport, error) {
	var current currentFile
	if err := readJSONPath(root, currentPath, &current); err != nil {
		return CoverageBatchReport{}, err
	}
	if coveragePath == "" {
		coveragePath = current.TruthSources.Coverage
	}
	var coverage coverageFile
	if err := readJSONPath(root, coveragePath, &coverage); err != nil {
		return CoverageBatchReport{}, err
	}
	report := CoverageBatchReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		CurrentPath:   filepath.ToSlash(currentPath),
		CoveragePath:  filepath.ToSlash(coveragePath),
		BatchID:       batchID,
	}
	batch := findCurrentCoverageBatch(current.CoverageBatches, batchID)
	if batch == nil {
		report.addIssue("missing_coverage_batch", "", "", fmt.Sprintf("coverage batch %q not found", batchID))
		report.finish()
		return report, nil
	}
	records := map[string]coverageRecord{}
	for _, record := range coverage.Coverage {
		records[record.ID] = record
	}
	for _, entry := range batch.Entries {
		report.checkEntry(root, records, entry)
	}
	report.Summary.Entries = len(report.Entries)
	report.finish()
	return report, nil
}

func findCurrentCoverageBatch(batches []currentCoverageBatch, id string) *currentCoverageBatch {
	for i := range batches {
		if batches[i].ID == id {
			return &batches[i]
		}
	}
	return nil
}

func (r *CoverageBatchReport) checkEntry(root string, records map[string]coverageRecord, entry currentCoverageBatchEntry) {
	row := CoverageBatchEntry{
		CoverageID: entry.CoverageID,
		Matrix:     filepath.ToSlash(entry.Matrix),
		Ledger:     filepath.ToSlash(entry.Ledger),
		Status:     StatusPass,
	}
	record, ok := records[entry.CoverageID]
	if !ok {
		row.Status = StatusFail
		r.addIssue("unknown_coverage_id", entry.CoverageID, "", "coverage batch references unknown coverage id")
		r.Entries = append(r.Entries, row)
		return
	}
	row.Expected = record.Totals
	if entry.Matrix == "" || !fileExists(root, entry.Matrix) {
		row.Status = StatusFail
		r.addIssue("missing_matrix", entry.CoverageID, entry.Matrix, "matrix artifact not found")
		r.Entries = append(r.Entries, row)
		return
	}
	var matrix matrixFile
	if err := readJSONPath(root, entry.Matrix, &matrix); err != nil {
		row.Status = StatusFail
		r.addIssue("invalid_matrix", entry.CoverageID, entry.Matrix, fmt.Sprintf("read matrix: %v", err))
		r.Entries = append(r.Entries, row)
		return
	}
	row.Actual = CoverageTotals{
		Total:   matrix.Summary.Total,
		Pass:    matrix.Summary.Passed,
		Blocked: matrix.Summary.Blocked,
		Failed:  matrix.Summary.Failed,
		Skipped: matrix.Summary.Skipped,
	}
	row.Recipes = matrixRecipes(matrix)
	if row.Expected != row.Actual {
		row.Status = StatusFail
		r.addIssue("coverage_totals_mismatch", entry.CoverageID, entry.Matrix, fmt.Sprintf("coverage totals %+v do not match matrix summary %+v", row.Expected, row.Actual))
	}
	if entry.Ledger != "" && !fileExists(root, entry.Ledger) {
		row.Status = StatusFail
		r.addIssue("missing_ledger", entry.CoverageID, entry.Ledger, "ledger artifact not found")
	}
	r.checkEntryRecipes(root, entry)
	r.Entries = append(r.Entries, row)
}

func (r *CoverageBatchReport) checkEntryRecipes(root string, entry currentCoverageBatchEntry) {
	if entry.RecipeGlob != "" {
		matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(entry.RecipeGlob)))
		if err != nil || len(matches) == 0 {
			r.addIssue("empty_recipe_glob", entry.CoverageID, entry.RecipeGlob, "recipe_glob matched no files")
		}
	}
	for _, recipePath := range entry.RecipePaths {
		if !fileExists(root, recipePath) {
			r.addIssue("missing_recipe", entry.CoverageID, recipePath, "recipe path not found")
		}
	}
}

func (r *CoverageBatchReport) addIssue(code, coverageID, path, message string) {
	r.Issues = append(r.Issues, CoverageBatchIssue{
		Code:       code,
		CoverageID: coverageID,
		Path:       filepath.ToSlash(path),
		Message:    message,
	})
}

func (r *CoverageBatchReport) finish() {
	r.Summary.Errors = len(r.Issues)
	if r.Summary.Errors > 0 {
		r.Status = StatusFail
		return
	}
	r.Status = StatusPass
}
