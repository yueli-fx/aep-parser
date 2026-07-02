package aepmigrate

import (
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
	AEOpenLabels     []string
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
	AEOpenVersion     string       `json:"ae_open_version,omitempty"`
	AEOpenPath        string       `json:"ae_open_path,omitempty"`
	Status            MatrixStatus `json:"status"`
	Reason            string       `json:"reason,omitempty"`
	CaseDir           string       `json:"case_dir,omitempty"`
	SourcePath        string       `json:"source_path,omitempty"`
	OutputPath        string       `json:"output_path,omitempty"`
	CompileReportPath string       `json:"compile_report_path,omitempty"`
	ConvertReportPath string       `json:"convert_report_path,omitempty"`
}

var matrixWriterLabels = []string{
	string(VersionAE2020),
	string(VersionAE2021),
	string(VersionAE2022),
	string(VersionAE2023),
	string(VersionAE2024),
	string(VersionAE2025),
}

var matrixAEHostLabels = []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024", "AE2025"}

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
	sourceLabels := expandMatrixSourceLabels(opts.SourceLabels)
	if len(sourceLabels) == 0 {
		sourceLabels = []string{"recipe"}
	}
	targetLabels := expandMatrixTargetLabels(opts.TargetLabels)
	if len(targetLabels) == 0 {
		targetLabels = []string{string(VersionAE2025)}
	}
	aeOpenLabels := matrixAEOpenLabels(opts.AEOpenLabels)
	caseCount := len(recipePaths) * len(sourceLabels) * len(targetLabels)
	if opts.AEOpen && len(aeOpenLabels) > 0 {
		caseCount *= len(aeOpenLabels)
	}
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
				if opts.AEOpen && len(aeOpenLabels) > 0 {
					for _, aeOpenLabel := range aeOpenLabels {
						c := runMatrixCase(opts, hosts, recipePath, sourceLabel, targetLabel, aeOpenLabel, true)
						report.Cases = append(report.Cases, c)
					}
					continue
				}
				c := runMatrixCase(opts, hosts, recipePath, sourceLabel, targetLabel, "", false)
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
