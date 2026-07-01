package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
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

func TestRunConvertWritesBlockedReportWithoutOutput(t *testing.T) {
	input := writeTempProjectWithOneSolidLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	reportPath := filepath.Join(t.TempDir(), "convert.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"convert", "-in", input, "-target", "AE2025", "-out", outPath, "-report", reportPath}, &stdout, &stderr)
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

func readReportSummary(t *testing.T, path string) struct {
	Summary struct {
		Status string `json:"status"`
	} `json:"summary"`
	Verification struct {
		ProfileDiffStatus string `json:"profile_diff_status"`
		ProfileDiffCount  int    `json:"profile_diff_count"`
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
		} `json:"verification"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Unmarshal report: %v", err)
	}
	return report
}
