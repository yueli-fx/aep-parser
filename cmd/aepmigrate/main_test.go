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
