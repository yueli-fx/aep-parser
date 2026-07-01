package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func TestRunAssessWritesJSONReport(t *testing.T) {
	input := writeTempProject(t, aep.TargetAE2020)
	outPath := filepath.Join(t.TempDir(), "assess.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"assess", "-in", input, "-target", "AE2020", "-out", outPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run assess = %d, stderr=%s", code, stderr.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var report struct {
		SchemaVersion int `json:"schema_version"`
		Summary       struct {
			Status string `json:"status"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if report.SchemaVersion != 1 || report.Summary.Status != "pass" {
		t.Fatalf("report = %+v", report)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("migration assess:")) {
		t.Fatalf("stdout missing report path: %s", stdout.String())
	}
}

func TestRunRejectsMissingAssessInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"assess", "-target", "AE2020"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run missing input = %d, want 2", code)
	}
}

func TestRunConvertWritesOutputAndReport(t *testing.T) {
	input := writeTempProjectWithOneComp(t, aep.TargetAE2020)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile report: %v", err)
	}
	if !bytes.Contains(data, []byte(`"profile_diff_count": 0`)) {
		t.Fatalf("report missing explicit profile_diff_count: %s", string(data))
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" {
		t.Fatalf("report status = %q, want pass", report.Summary.Status)
	}
	if report.Verification.ProfileDiffStatus != "pass" || report.Verification.ProfileDiffCount != 0 {
		t.Fatalf("profile diff verification = %+v, want pass with 0 diffs", report.Verification)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("migration convert:")) {
		t.Fatalf("stdout missing output path: %s", stdout.String())
	}
}

func TestRunConvertCanRunAEOpenGate(t *testing.T) {
	input := writeTempProjectWithOneComp(t, aep.TargetAE2020)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	host := &fakeCLIHost{doneBody: "PASS\nproject items.length=1\ncomp=Main layers.length=0\n"}
	var stdout, stderr bytes.Buffer

	code := runWithHost([]string{
		"convert",
		"-in", input,
		"-target", "AE2025",
		"-out", outPath,
		"-report", reportPath,
		"-ae-open",
		"-ae", "AfterFX.exe",
	}, &stdout, &stderr, host)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	report := readReportSummary(t, reportPath)
	if report.Verification.AEOpenStatus != "pass" {
		t.Fatalf("AE open verification = %+v, want pass", report.Verification)
	}
	if !host.called {
		t.Fatal("fake AE host was not called")
	}
	if !filepath.IsAbs(host.request.JSXPath) || !filepath.IsAbs(host.request.DonePath) {
		t.Fatalf("AE open request paths must be absolute: %+v", host.request)
	}
	argsPath := host.request.Env["AE_OPEN_ARGS"]
	if argsPath == "" || !filepath.IsAbs(filepath.FromSlash(argsPath)) {
		t.Fatalf("AE open args path must be absolute in env: %+v", host.request.Env)
	}
	if filepath.Dir(filepath.FromSlash(argsPath)) != filepath.Dir(outPath) {
		t.Fatalf("AE open args path should live beside output: %q", argsPath)
	}
}

func TestRunMatrixWritesAggregateReport(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "minimal.json")
	writeMinimalMatrixRecipe(t, recipePath)
	outRoot := filepath.Join(root, "matrix")
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"matrix",
		"-recipe", recipePath,
		"-sources", "AE2020",
		"-targets", "AE2020,AE2021,AE2025",
		"-out", outRoot,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run matrix = %d, stderr=%s", code, stderr.String())
	}
	reportPath := filepath.Join(outRoot, "matrix.json")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile matrix report: %v", err)
	}
	var report struct {
		Summary struct {
			Total   int `json:"total"`
			Passed  int `json:"passed"`
			Skipped int `json:"skipped"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Unmarshal matrix report: %v", err)
	}
	if report.Summary.Total != 3 || report.Summary.Passed != 2 || report.Summary.Skipped != 1 {
		t.Fatalf("matrix summary = %+v", report.Summary)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("migration matrix:")) {
		t.Fatalf("stdout missing matrix report path: %s", stdout.String())
	}
}

func TestRunMatrixWritesLedgerOut(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "minimal.json")
	writeMinimalMatrixRecipe(t, recipePath)
	outRoot := filepath.Join(root, "matrix")
	ledgerPath := filepath.Join(outRoot, "ledger.md")
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"matrix",
		"-recipe", recipePath,
		"-sources", "AE2020",
		"-targets", "AE2020",
		"-out", outRoot,
		"-ledger-out", ledgerPath,
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run matrix = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(outRoot, "matrix.json")); err != nil {
		t.Fatalf("matrix.json missing: %v", err)
	}
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("ReadFile ledger: %v", err)
	}
	if !bytes.Contains(data, []byte("| other | `minimal` | pass | 1 | 0 | 0 | 0 | AE2020 | AE2020 | - |")) {
		t.Fatalf("ledger missing recipe row:\n%s", string(data))
	}
	if !bytes.Contains(stdout.Bytes(), []byte("migration matrix:")) {
		t.Fatalf("stdout missing matrix path: %s", stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("migration ledger:")) {
		t.Fatalf("stdout missing ledger path: %s", stdout.String())
	}
}

func TestRunMatrixAEOpenRejectsLargeCaseCountByDefault(t *testing.T) {
	root := t.TempDir()
	recipeDir := filepath.Join(root, "recipes")
	if err := os.MkdirAll(recipeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeMinimalMatrixRecipe(t, filepath.Join(recipeDir, "minimal-a.json"))
	writeMinimalMatrixRecipe(t, filepath.Join(recipeDir, "minimal-b.json"))
	outRoot := filepath.Join(root, "matrix")
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"matrix",
		"-recipes", recipeDir,
		"-sources", "AE2020",
		"-targets", "AE2020,AE2022,AE2025",
		"-out", outRoot,
		"-ae-open",
		"-max-ae-open-cases", "5",
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run matrix = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "AE open matrix would run 6 cases") {
		t.Fatalf("stderr missing AE-open case limit: %s", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(outRoot, "matrix.json")); !os.IsNotExist(err) {
		t.Fatalf("matrix.json exists or stat failed unexpectedly: %v", err)
	}
}

func TestRunMatrixAEOpenVersionsExpandCaseLimit(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "minimal.json")
	writeMinimalMatrixRecipe(t, recipePath)
	outRoot := filepath.Join(root, "matrix")
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"matrix",
		"-recipe", recipePath,
		"-sources", "AE2020",
		"-targets", "AE2020",
		"-out", outRoot,
		"-ae-open",
		"-ae-versions", "AE2020,AE2021,AE2025",
		"-max-ae-open-cases", "2",
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run matrix = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "AE open matrix would run 3 cases") {
		t.Fatalf("stderr missing AE-open version case limit: %s", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(outRoot, "matrix.json")); !os.IsNotExist(err) {
		t.Fatalf("matrix.json exists or stat failed unexpectedly: %v", err)
	}
}

func TestRunLedgerWritesMarkdown(t *testing.T) {
	root := t.TempDir()
	matrixPath := filepath.Join(root, "matrix.json")
	ledgerPath := filepath.Join(root, "ledger.md")
	if err := os.WriteFile(matrixPath, []byte(`{
  "schema_version": 1,
  "out_root": "tmp/matrix",
  "summary": {
    "total": 2,
    "passed": 1,
    "blocked": 1,
    "failed": 0,
    "skipped": 0
  },
  "cases": [
    {
      "recipe_name": "minimal-text-animator-skew",
      "source_version": "AE2020",
      "target_version": "AE2020",
      "ae_open_version": "AE2020",
      "status": "pass"
    },
    {
      "recipe_name": "minimal-layer-explicit-matte",
      "source_version": "AE2025",
      "target_version": "AE2020",
      "status": "blocked",
      "reason": "convert_blocked"
    }
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"ledger",
		"-matrix", matrixPath,
		"-out", ledgerPath,
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run ledger = %d, stderr=%s", code, stderr.String())
	}
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("ReadFile ledger: %v", err)
	}
	if !bytes.Contains(data, []byte("| text | `minimal-text-animator-skew` | pass | 1 | 0 | 0 | 0 | AE2020 | AE2020 | AE2020 |")) {
		t.Fatalf("ledger missing text row:\n%s", string(data))
	}
	if !bytes.Contains(data, []byte("| layer | `minimal-layer-explicit-matte` | blocked | 0 | 1 | 0 | 0 | AE2025 | AE2020 | - |")) {
		t.Fatalf("ledger missing layer row:\n%s", string(data))
	}
	if !bytes.Contains(stdout.Bytes(), []byte("migration ledger:")) {
		t.Fatalf("stdout missing ledger path: %s", stdout.String())
	}
}

func TestRunConvertWritesDefaultNullLayerOutput(t *testing.T) {
	input := writeTempProjectWithOneNullLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesDefaultSolidLayerOutput(t *testing.T) {
	input := writeTempProjectWithOneSolidLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesDefaultAdjustmentLayerOutput(t *testing.T) {
	input := writeTempProjectWithOneAdjustmentLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesDefaultCameraLayerOutput(t *testing.T) {
	input := writeTempProjectWithOneCameraLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesDefaultLightLayerOutput(t *testing.T) {
	input := writeTempProjectWithOneLightLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesDefaultPrecompLayerOutput(t *testing.T) {
	input := writeTempProjectWithOnePrecompLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesDefaultTextLayerOutput(t *testing.T) {
	input := writeTempProjectWithDefaultTextLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesMovedNullLayerOutput(t *testing.T) {
	input := writeTempProjectWithMovedNullLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesDefaultShapeLayerOutput(t *testing.T) {
	input := writeTempProjectWithDefaultShapeLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesRectFillShapeLayerOutput(t *testing.T) {
	input := writeTempProjectWithRectFillShapeLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run convert = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("converted output missing: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "pass" || report.Verification.ProfileDiffStatus != "pass" {
		t.Fatalf("report = %+v, want pass with profile diff pass", report)
	}
}

func TestRunConvertWritesBlockedReportWithoutOutput(t *testing.T) {
	input := writeTempRecipe(t, filepath.Join("..", "..", "examples", "recipes", "minimal-layer-explicit-matte.json"))
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2020", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run blocked convert = %d, want 1; stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Fatalf("blocked output exists or stat failed unexpectedly: %v", err)
	}
	report := readReportSummary(t, reportPath)
	if report.Summary.Status != "blocked" {
		t.Fatalf("report status = %q, want blocked", report.Summary.Status)
	}
}

func writeTempRecipe(t *testing.T, recipePath string) string {
	t.Helper()
	data, err := os.ReadFile(recipePath)
	if err != nil {
		t.Fatalf("ReadFile recipe: %v", err)
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("Unmarshal recipe: %v", err)
	}
	path := filepath.Join(t.TempDir(), "source.aep")
	report, err := recipe.CompileToFile(rec, path, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("recipe report invalid: refusals=%+v", report.Refusals)
	}
	return path
}

func writeTempProject(t *testing.T, target aep.AETarget) string {
	t.Helper()
	project := aep.NewProject(target)
	path := filepath.Join(t.TempDir(), "source.aep")
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	return path
}

func writeTempProjectWithOneComp(t *testing.T, target aep.AETarget) string {
	t.Helper()
	project := aep.NewProject(target)
	if _, err := aep.NewComposition(project, "Main", 640, 360, 24, 2.5); err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	return writeProject(t, project, "one-comp.aep")
}

func writeTempProjectWithOneSolidLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "Solid", 640, 360, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	return writeProject(t, project, "one-layer.aep")
}

func writeTempProjectWithOneTextLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewTextLayer(comp, "Title"); err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	return writeProject(t, project, "one-text-layer.aep")
}

func writeTempProjectWithDefaultTextLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewTextLayer(comp, "Title")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := layer.SetText("Hello"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	return writeProject(t, project, "default-text-layer.aep")
}

func writeTempProjectWithMovedNullLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewNullLayer(comp, "Controller")
	if err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	transform := aep.NewLayerTransform()
	if err := transform.Position().SetStaticValue([2]float64{320, 180}); err != nil {
		t.Fatalf("Position.SetStaticValue: %v", err)
	}
	if err := aep.SetLayerTransform(layer, transform); err != nil {
		t.Fatalf("SetLayerTransform: %v", err)
	}
	return writeProject(t, project, "moved-null-layer.aep")
}

func writeTempProjectWithKeyframedNullLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewNullLayer(comp, "Controller")
	if err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	transform := aep.NewLayerTransform()
	if err := transform.Position().AddKeyframeLinear(0, [2]float64{0, 0}); err != nil {
		t.Fatalf("Position.AddKeyframeLinear 0: %v", err)
	}
	if err := transform.Position().AddKeyframeLinear(1, [2]float64{320, 180}); err != nil {
		t.Fatalf("Position.AddKeyframeLinear 1: %v", err)
	}
	if err := aep.SetLayerTransform(layer, transform); err != nil {
		t.Fatalf("SetLayerTransform: %v", err)
	}
	return writeProject(t, project, "keyframed-null-layer.aep")
}

func writeTempProjectWithCommentedTextLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewTextLayer(comp, "Title")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := layer.SetComment("unsupported text layer comment"); err != nil {
		t.Fatalf("SetComment: %v", err)
	}
	return writeProject(t, project, "commented-text-layer.aep")
}

func writeTempProjectWithDefaultShapeLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewShapeLayer(comp, "Shape"); err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	return writeProject(t, project, "default-shape-layer.aep")
}

func writeTempProjectWithRectFillShapeLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	shape, err := aep.NewShapeLayer(comp, "Card")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, err := shape.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{320, 180}); err != nil {
		t.Fatalf("Rect.SetSize: %v", err)
	}
	fill, err := shape.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 0.25, 0.5, 1}); err != nil {
		t.Fatalf("Fill.SetColor: %v", err)
	}
	return writeProject(t, project, "rect-fill-shape-layer.aep")
}

func writeTempProjectWithOneAdjustmentLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewAdjustmentLayer(comp, "Grade"); err != nil {
		t.Fatalf("NewAdjustmentLayer: %v", err)
	}
	return writeProject(t, project, "one-adjustment-layer.aep")
}

func writeTempProjectWithOneCameraLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewCameraLayer(comp, "Camera"); err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}
	return writeProject(t, project, "one-camera-layer.aep")
}

func writeTempProjectWithOneLightLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewLightLayer(comp, "Light"); err != nil {
		t.Fatalf("NewLightLayer: %v", err)
	}
	return writeProject(t, project, "one-light-layer.aep")
}

func writeTempProjectWithOnePrecompLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	source, err := aep.NewComposition(project, "Source", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition source: %v", err)
	}
	main, err := aep.NewComposition(project, "Main", 1920, 1080, 30, 3)
	if err != nil {
		t.Fatalf("NewComposition main: %v", err)
	}
	if _, err := aep.NewPrecompLayer(main, source, "Nested Source"); err != nil {
		t.Fatalf("NewPrecompLayer: %v", err)
	}
	return writeProject(t, project, "one-precomp-layer.aep")
}

func writeTempProjectWithOneNullLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewNullLayer(comp, "Controller"); err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	return writeProject(t, project, "one-null-layer.aep")
}

func writeProject(t *testing.T, project *aep.Project, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	return path
}

func writeMinimalMatrixRecipe(t *testing.T, path string) {
	t.Helper()
	data := []byte(`{
  "schema_version": 1,
  "project": {
    "name": "Matrix fixture",
    "target_version": "AE2020"
  },
  "comps": [
    {
      "name": "Main",
      "width": 640,
      "height": 360,
      "frame_rate": 24,
      "duration": 2.5
    }
  ]
}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readReportSummary(t *testing.T, path string) struct {
	Summary struct {
		Status string `json:"status"`
	} `json:"summary"`
	Verification struct {
		ProfileDiffStatus string `json:"profile_diff_status"`
		ProfileDiffCount  int    `json:"profile_diff_count"`
		AEOpenStatus      string `json:"ae_open_status"`
	} `json:"verification"`
} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile report: %v", err)
	}
	var report struct {
		Summary struct {
			Status string `json:"status"`
		} `json:"summary"`
		Verification struct {
			ProfileDiffStatus string `json:"profile_diff_status"`
			ProfileDiffCount  int    `json:"profile_diff_count"`
			AEOpenStatus      string `json:"ae_open_status"`
		} `json:"verification"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Unmarshal report: %v", err)
	}
	return report
}

type fakeCLIHost struct {
	called   bool
	request  aehost.ScriptRequest
	doneBody string
}

func (h *fakeCLIHost) Available(context.Context) aehost.Availability {
	return aehost.Availability{Status: aehost.CapabilityAvailable}
}

func (h *fakeCLIHost) RunScript(_ context.Context, req aehost.ScriptRequest) (aehost.ScriptResult, error) {
	h.called = true
	h.request = req
	if err := os.MkdirAll(filepath.Dir(req.DonePath), 0o755); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	if err := os.WriteFile(req.DonePath, []byte(h.doneBody), 0o644); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	return aehost.ScriptResult{ExitCode: 0, DonePath: req.DonePath}, nil
}
