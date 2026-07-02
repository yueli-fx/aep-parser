package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CoverageUpdateOptions struct {
	ID             string
	CoveragePath   string
	MatrixPath     string
	Domain         string
	Scope          string
	WriterStatus   string
	HostOpenStatus string
	LedgerPath     string
}

type CoverageUpdateReport struct {
	SchemaVersion  int            `json:"schema_version"`
	Status         string         `json:"status"`
	CoverageID     string         `json:"coverage_id"`
	CoveragePath   string         `json:"coverage_path"`
	MatrixPath     string         `json:"matrix_path"`
	LedgerPath     string         `json:"ledger_path,omitempty"`
	Created        bool           `json:"created"`
	Recipes        []string       `json:"recipes"`
	SourceWriters  []string       `json:"source_writers"`
	TargetWriters  []string       `json:"target_writers"`
	WriterStatus   string         `json:"writer_status"`
	WriterCoverage string         `json:"writer_coverage"`
	HostOpenStatus string         `json:"host_open_status"`
	Totals         CoverageTotals `json:"totals"`
}

func UpdateCoverageFromMatrix(root string, opts CoverageUpdateOptions) (CoverageUpdateReport, error) {
	if opts.ID == "" {
		return CoverageUpdateReport{}, fmt.Errorf("coverage id is required")
	}
	if opts.CoveragePath == "" {
		return CoverageUpdateReport{}, fmt.Errorf("coverage path is required")
	}
	if opts.MatrixPath == "" {
		return CoverageUpdateReport{}, fmt.Errorf("matrix path is required")
	}

	var matrix matrixFile
	if err := readJSONPath(root, opts.MatrixPath, &matrix); err != nil {
		return CoverageUpdateReport{}, fmt.Errorf("read matrix: %w", err)
	}
	coverage, err := readJSONMap(root, opts.CoveragePath)
	if err != nil {
		return CoverageUpdateReport{}, fmt.Errorf("read coverage: %w", err)
	}

	records, err := coverageRecordsMap(coverage)
	if err != nil {
		return CoverageUpdateReport{}, err
	}
	record, created, err := findOrCreateCoverageRecord(records, opts)
	if err != nil {
		return CoverageUpdateReport{}, err
	}
	if created {
		coverage["coverage"] = append(records, record)
	}

	recipes := matrixUniqueValues(matrix, func(c matrixCase) string { return c.RecipeName })
	sourceWriters := matrixUniqueValues(matrix, func(c matrixCase) string { return c.SourceVersion })
	targetWriters := matrixUniqueValues(matrix, func(c matrixCase) string { return c.TargetVersion })
	totals := CoverageTotals{
		Total:   matrix.Summary.Total,
		Pass:    matrix.Summary.Passed,
		Blocked: matrix.Summary.Blocked,
		Failed:  matrix.Summary.Failed,
		Skipped: matrix.Summary.Skipped,
	}

	if opts.Domain != "" {
		record["domain"] = opts.Domain
	}
	if opts.Scope != "" {
		record["scope"] = opts.Scope
	}
	writerStatus := opts.WriterStatus
	if writerStatus == "" {
		writerStatus = writerStatusFromTotals(totals)
	}
	record["writer_status"] = writerStatus
	writerCoverage := fmt.Sprintf("%s sources into %s targets", strings.Join(sourceWriters, ","), strings.Join(targetWriters, ","))
	record["writer_coverage"] = writerCoverage
	hostOpenStatus := opts.HostOpenStatus
	if hostOpenStatus == "" {
		hostOpenStatus = stringValue(record["host_open_status"])
	}
	if hostOpenStatus == "" {
		hostOpenStatus = "pending"
	}
	record["host_open_status"] = hostOpenStatus

	matrixPath := repoRelativePath(root, opts.MatrixPath)
	ledgerPath := opts.LedgerPath
	if ledgerPath == "" {
		ledgerPath = defaultCoverageLedgerPath(root, opts.MatrixPath)
	}
	record["artifact"] = matrixPath
	if ledgerPath != "" {
		ledgerPath = repoRelativePath(root, ledgerPath)
		record["ledger"] = ledgerPath
	}
	record["totals"] = map[string]any{
		"total":   totals.Total,
		"pass":    totals.Pass,
		"blocked": totals.Blocked,
		"failed":  totals.Failed,
		"skipped": totals.Skipped,
	}
	record["recipes"] = stringsToAny(recipes)
	record["recipe_count"] = len(recipes)
	coverage["generated_at"] = time.Now().Format("2006-01-02")

	if err := writeJSONMap(root, opts.CoveragePath, coverage); err != nil {
		return CoverageUpdateReport{}, fmt.Errorf("write coverage: %w", err)
	}
	return CoverageUpdateReport{
		SchemaVersion:  1,
		Status:         StatusPass,
		CoverageID:     opts.ID,
		CoveragePath:   filepath.ToSlash(opts.CoveragePath),
		MatrixPath:     matrixPath,
		LedgerPath:     ledgerPath,
		Created:        created,
		Recipes:        recipes,
		SourceWriters:  sourceWriters,
		TargetWriters:  targetWriters,
		WriterStatus:   writerStatus,
		WriterCoverage: writerCoverage,
		HostOpenStatus: hostOpenStatus,
		Totals:         totals,
	}, nil
}

func coverageRecordsMap(coverage map[string]any) ([]any, error) {
	raw, ok := coverage["coverage"]
	if !ok || raw == nil {
		return nil, fmt.Errorf("coverage array is required")
	}
	records, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("coverage must be an array")
	}
	return records, nil
}

func findOrCreateCoverageRecord(records []any, opts CoverageUpdateOptions) (map[string]any, bool, error) {
	for _, raw := range records {
		record, ok := raw.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("coverage record must be an object")
		}
		if record["id"] == opts.ID {
			return record, false, nil
		}
	}
	if opts.Domain == "" || opts.Scope == "" {
		return nil, false, fmt.Errorf("coverage id %q does not exist; provide domain and scope to create it", opts.ID)
	}
	return map[string]any{
		"id":               opts.ID,
		"domain":           opts.Domain,
		"scope":            opts.Scope,
		"writer_status":    "",
		"writer_coverage":  "",
		"host_open_status": "",
		"artifact":         "",
		"ledger":           "",
		"totals":           map[string]any{},
	}, true, nil
}

func matrixUniqueValues(matrix matrixFile, pick func(matrixCase) string) []string {
	values := map[string]bool{}
	for _, c := range matrix.Cases {
		value := pick(c)
		if value != "" {
			values[value] = true
		}
	}
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func writerStatusFromTotals(t CoverageTotals) string {
	if t.Failed > 0 {
		return "failed"
	}
	if t.Blocked > 0 || t.Skipped > 0 {
		return "boundary"
	}
	return "PD-6x6"
}

func defaultCoverageLedgerPath(root, matrixPath string) string {
	matrixFullPath := resolveRegistryPath(root, matrixPath)
	candidate := filepath.Join(filepath.Dir(matrixFullPath), "ledger.md")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

func repoRelativePath(root, path string) string {
	fullPath := resolveRegistryPath(root, path)
	rootFullPath := resolveRegistryPath(root, ".")
	if rel, err := filepath.Rel(rootFullPath, fullPath); err == nil && !strings.HasPrefix(rel, "..") && rel != "." {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}

func resolveRegistryPath(root, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func stringValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func stringsToAny(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func readJSONMap(root, path string) (map[string]any, error) {
	fullPath := resolveRegistryPath(root, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func writeJSONMap(root, path string, value map[string]any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fullPath := resolveRegistryPath(root, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, append(data, '\n'), 0o644)
}
