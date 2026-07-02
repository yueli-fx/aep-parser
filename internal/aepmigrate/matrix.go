package aepmigrate

import (
	"path/filepath"

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

var matrixWriterLabels = SupportedVersionStrings()

var matrixAEHostLabels = SupportedVersionStrings()

func RunMatrix(opts MatrixOptions) (MatrixReport, error) {
	plan, err := prepareMatrixRun(opts)
	if err != nil {
		return MatrixReport{}, err
	}
	report := MatrixReport{
		SchemaVersion: SchemaVersion,
		OutRoot:       opts.OutRoot,
		AEHosts:       plan.hosts,
	}
	report.Cases = runMatrixCases(opts, plan)
	report.Summary = summarizeMatrix(report.Cases)
	if err := writeMatrixReport(filepath.Join(opts.OutRoot, "matrix.json"), report); err != nil {
		return MatrixReport{}, err
	}
	return report, nil
}
