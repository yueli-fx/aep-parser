package aepmigrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/recipe"
)

type MatrixStatus string

const (
	MatrixStatusPass    MatrixStatus = "pass"
	MatrixStatusBlocked MatrixStatus = "blocked"
	MatrixStatusFailed  MatrixStatus = "failed"
	MatrixStatusSkipped MatrixStatus = "skipped"
)

type MatrixOptions struct {
	RecipePaths      []string
	RecipeDir        string
	SourceLabels     []string
	TargetLabels     []string
	OutRoot          string
	AEInstallRoot    string
	AEHosts          map[string]string
	AEOpen           bool
	AEOpenJSXPath    string
	AEOpenTimeoutSec int
	MaxAEOpenCases   int
	Host             aehost.Host
	Capabilities     recipe.CapabilityIndex
}

type MatrixReport struct {
	SchemaVersion int               `json:"schema_version"`
	OutRoot       string            `json:"out_root"`
	AEHosts       map[string]string `json:"ae_hosts,omitempty"`
	Summary       MatrixSummary     `json:"summary"`
	Cases         []MatrixCase      `json:"cases"`
}

type MatrixSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Blocked int `json:"blocked"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

type MatrixCase struct {
	RecipePath        string       `json:"recipe_path"`
	RecipeName        string       `json:"recipe_name"`
	SourceVersion     string       `json:"source_version"`
	TargetVersion     string       `json:"target_version"`
	Status            MatrixStatus `json:"status"`
	Reason            string       `json:"reason,omitempty"`
	CaseDir           string       `json:"case_dir,omitempty"`
	SourcePath        string       `json:"source_path,omitempty"`
	OutputPath        string       `json:"output_path,omitempty"`
	CompileReportPath string       `json:"compile_report_path,omitempty"`
	ConvertReportPath string       `json:"convert_report_path,omitempty"`
}

func BuildAEHostMap(installRoot string) map[string]string {
	hosts := map[string]string{}
	if strings.TrimSpace(installRoot) == "" {
		return hosts
	}
	for year := 2020; year <= 2025; year++ {
		label := fmt.Sprintf("AE%d", year)
		path := filepath.Join(installRoot, fmt.Sprintf("Adobe After Effects %d", year), "Support Files", "AfterFX.exe")
		if _, err := os.Stat(path); err == nil {
			hosts[label] = path
		}
	}
	return hosts
}

func RunMatrix(opts MatrixOptions) (MatrixReport, error) {
	if opts.OutRoot == "" {
		return MatrixReport{}, fmt.Errorf("matrix out root is required")
	}
	recipePaths, err := matrixRecipePaths(opts)
	if err != nil {
		return MatrixReport{}, err
	}
	if len(recipePaths) == 0 {
		return MatrixReport{}, fmt.Errorf("matrix requires at least one recipe")
	}
	sourceLabels := opts.SourceLabels
	if len(sourceLabels) == 0 {
		sourceLabels = []string{"recipe"}
	}
	targetLabels := opts.TargetLabels
	if len(targetLabels) == 0 {
		targetLabels = []string{string(VersionAE2025)}
	}
	caseCount := len(recipePaths) * len(sourceLabels) * len(targetLabels)
	if opts.AEOpen && opts.MaxAEOpenCases > 0 && caseCount > opts.MaxAEOpenCases {
		return MatrixReport{}, fmt.Errorf("AE open matrix would run %d cases, above limit %d; narrow recipes/targets or set a higher -max-ae-open-cases", caseCount, opts.MaxAEOpenCases)
	}
	hosts := mergeHostMaps(BuildAEHostMap(opts.AEInstallRoot), opts.AEHosts)
	report := MatrixReport{
		SchemaVersion: SchemaVersion,
		OutRoot:       opts.OutRoot,
		AEHosts:       hosts,
	}
	for _, recipePath := range recipePaths {
		for _, sourceLabel := range sourceLabels {
			for _, targetLabel := range targetLabels {
				c := runMatrixCase(opts, hosts, recipePath, sourceLabel, targetLabel)
				report.Cases = append(report.Cases, c)
			}
		}
	}
	report.Summary = summarizeMatrix(report.Cases)
	if err := writeMatrixReport(filepath.Join(opts.OutRoot, "matrix.json"), report); err != nil {
		return MatrixReport{}, err
	}
	return report, nil
}

func matrixRecipePaths(opts MatrixOptions) ([]string, error) {
	seen := map[string]bool{}
	var paths []string
	add := func(path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		paths = append(paths, path)
	}
	for _, path := range opts.RecipePaths {
		add(path)
	}
	if opts.RecipeDir != "" {
		matches, err := filepath.Glob(filepath.Join(opts.RecipeDir, "*.json"))
		if err != nil {
			return nil, err
		}
		for _, path := range matches {
			add(path)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func runMatrixCase(opts MatrixOptions, hosts map[string]string, recipePath, sourceLabel, targetLabel string) MatrixCase {
	recipeName := strings.TrimSuffix(filepath.Base(recipePath), filepath.Ext(recipePath))
	caseDir := filepath.Join(opts.OutRoot, sanitizeMatrixName(recipeName), sanitizeMatrixName(sourceLabel+"_to_"+targetLabel))
	c := MatrixCase{
		RecipePath:    recipePath,
		RecipeName:    recipeName,
		SourceVersion: sourceLabel,
		TargetVersion: targetLabel,
		CaseDir:       caseDir,
	}
	sourceWriterLabel, sourceWriterOK := matrixSourceWriterLabel(sourceLabel)
	if !sourceWriterOK {
		c.Status = MatrixStatusSkipped
		c.Reason = "unsupported_source_writer"
		return c
	}
	target, targetWriterOK := matrixWriterTarget(targetLabel)
	if !targetWriterOK {
		c.Status = MatrixStatusSkipped
		c.Reason = "unsupported_writer_target"
		return c
	}
	if opts.AEOpen {
		if hosts[targetLabel] == "" {
			c.Status = MatrixStatusSkipped
			c.Reason = "missing_ae_host"
			return c
		}
		if opts.Host == nil {
			c.Status = MatrixStatusSkipped
			c.Reason = "missing_ae_host_runner"
			return c
		}
	}
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	sourcePath := filepath.Join(caseDir, "source.aep")
	compileReportPath := filepath.Join(caseDir, "compile_report.json")
	c.SourcePath = sourcePath
	c.CompileReportPath = compileReportPath
	rec, err := readMatrixRecipe(recipePath)
	if err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	if sourceWriterLabel != "" {
		rec.Project.TargetVersion = sourceWriterLabel
	}
	compileReport, err := recipe.CompileToFile(rec, sourcePath, opts.Capabilities)
	if writeErr := writeMatrixJSON(compileReportPath, compileReport); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	if !compileReport.Valid {
		c.Status = MatrixStatusBlocked
		c.Reason = "recipe_compile_blocked"
		return c
	}
	outputPath := filepath.Join(caseDir, "target.aep")
	convertReportPath := filepath.Join(caseDir, "convert_report.json")
	c.OutputPath = outputPath
	c.ConvertReportPath = convertReportPath
	convertOpts := ConvertOptions{
		InputPath:  sourcePath,
		OutputPath: outputPath,
		Target:     target,
	}
	if opts.AEOpen {
		jsPath := opts.AEOpenJSXPath
		if jsPath == "" {
			jsPath = filepath.Join("test_data", "generators", "verify_open.jsx")
		}
		jsPath, err = filepath.Abs(jsPath)
		if err != nil {
			c.Status = MatrixStatusFailed
			c.Reason = err.Error()
			return c
		}
		argsPath, err := filepath.Abs(filepath.Join(caseDir, "ae_open.args.json"))
		if err != nil {
			c.Status = MatrixStatusFailed
			c.Reason = err.Error()
			return c
		}
		donePath, err := filepath.Abs(filepath.Join(caseDir, "ae_open.done"))
		if err != nil {
			c.Status = MatrixStatusFailed
			c.Reason = err.Error()
			return c
		}
		convertOpts.AEOpen = &AEOpenOptions{
			Host:       opts.Host,
			AEPath:     hosts[targetLabel],
			JSXPath:    jsPath,
			ArgsPath:   argsPath,
			DonePath:   donePath,
			TimeoutSec: opts.AEOpenTimeoutSec,
		}
	}
	convertReport, err := Convert(convertOpts)
	if writeErr := writeMatrixJSON(convertReportPath, convertReport); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		c.Status = MatrixStatusFailed
		c.Reason = err.Error()
		return c
	}
	switch convertReport.Summary.Status {
	case StatusPass:
		c.Status = MatrixStatusPass
	case StatusBlocked:
		c.Status = MatrixStatusBlocked
		c.Reason = "convert_blocked"
	case StatusError:
		c.Status = MatrixStatusFailed
		c.Reason = "convert_error"
	default:
		c.Status = MatrixStatusFailed
		c.Reason = "convert_status_" + string(convertReport.Summary.Status)
	}
	return c
}

func matrixSourceWriterLabel(label string) (string, bool) {
	if strings.EqualFold(strings.TrimSpace(label), "recipe") {
		return "", true
	}
	target, ok := matrixWriterTarget(label)
	if !ok {
		return "", false
	}
	return string(target), true
}

func matrixWriterTarget(label string) (VersionLabel, bool) {
	target, err := ParseVersionLabel(label)
	if err != nil {
		return "", false
	}
	switch target {
	case VersionAE2020, VersionAE2022, VersionAE2025:
		return target, true
	default:
		return "", false
	}
}

func readMatrixRecipe(path string) (recipe.Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return recipe.Recipe{}, err
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(data, &rec); err != nil {
		return recipe.Recipe{}, err
	}
	return rec, nil
}

func mergeHostMaps(a, b map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

func summarizeMatrix(cases []MatrixCase) MatrixSummary {
	s := MatrixSummary{Total: len(cases)}
	for _, c := range cases {
		switch c.Status {
		case MatrixStatusPass:
			s.Passed++
		case MatrixStatusBlocked:
			s.Blocked++
		case MatrixStatusFailed:
			s.Failed++
		case MatrixStatusSkipped:
			s.Skipped++
		}
	}
	return s
}

func writeMatrixReport(path string, report MatrixReport) error {
	return writeMatrixJSON(path, report)
}

func writeMatrixJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func sanitizeMatrixName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unnamed"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
