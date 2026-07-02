package aepmigrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAEHostMapDiscoversInstalledAfterEffectsVersions(t *testing.T) {
	root := t.TempDir()
	for _, version := range []string{"2020", "2021", "2022", "2023", "2024", "2025"} {
		exe := filepath.Join(root, "Adobe After Effects "+version, "Support Files", "AfterFX.exe")
		if err := os.MkdirAll(filepath.Dir(exe), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(exe, []byte("fake"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	hosts := BuildAEHostMap(root)

	for _, label := range []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024", "AE2025"} {
		if hosts[label] == "" {
			t.Fatalf("host %s was not discovered in %+v", label, hosts)
		}
	}
}

func TestRunMatrixConvertsRecipeAcrossNativeWriterTargets(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "minimal.json")
	writeMinimalMatrixRecipe(t, recipePath)
	outRoot := filepath.Join(root, "matrix")

	report, err := RunMatrix(MatrixOptions{
		RecipePaths: []string{recipePath},
		SourceLabels: []string{
			"AE2020",
		},
		TargetLabels: []string{
			"AE2020",
			"AE2021",
			"AE2025",
		},
		OutRoot: outRoot,
	})
	if err != nil {
		t.Fatalf("RunMatrix: %v", err)
	}
	if report.Summary.Total != 3 {
		t.Fatalf("total = %d, want 3", report.Summary.Total)
	}
	if report.Summary.Passed != 3 {
		t.Fatalf("passed = %d, want 3; cases=%+v", report.Summary.Passed, report.Cases)
	}
	if report.Summary.Skipped != 0 {
		t.Fatalf("skipped = %d, want 0; cases=%+v", report.Summary.Skipped, report.Cases)
	}
	assertMatrixCaseStatus(t, report.Cases, "AE2020", MatrixStatusPass)
	assertMatrixCaseStatus(t, report.Cases, "AE2021", MatrixStatusPass)
	assertMatrixCaseStatus(t, report.Cases, "AE2025", MatrixStatusPass)

	matrixJSON := filepath.Join(outRoot, "matrix.json")
	if _, err := os.Stat(matrixJSON); err != nil {
		t.Fatalf("matrix.json was not written: %v", err)
	}
	data, err := os.ReadFile(matrixJSON)
	if err != nil {
		t.Fatalf("read matrix.json: %v", err)
	}
	var disk MatrixReport
	if err := json.Unmarshal(data, &disk); err != nil {
		t.Fatalf("unmarshal matrix.json: %v", err)
	}
	if disk.Summary.Total != report.Summary.Total {
		t.Fatalf("disk total = %d, want %d", disk.Summary.Total, report.Summary.Total)
	}
}

func TestRunMatrixSkipsRecipeWhenSourceWriterViolatesRecipeContract(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join("..", "..", "examples", "recipes", "minimal-layer-explicit-matte.json")
	outRoot := filepath.Join(root, "matrix")

	report, err := RunMatrix(MatrixOptions{
		RecipePaths:  []string{recipePath},
		SourceLabels: []string{"AE2020"},
		TargetLabels: []string{"AE2025"},
		OutRoot:      outRoot,
	})
	if err != nil {
		t.Fatalf("RunMatrix: %v", err)
	}
	if report.Summary.Total != 1 || report.Summary.Skipped != 1 || report.Summary.Blocked != 0 {
		t.Fatalf("summary = %+v, want total=1 skipped=1 blocked=0; cases=%+v", report.Summary, report.Cases)
	}
	if got := report.Cases[0].Status; got != MatrixStatusSkipped {
		t.Fatalf("status = %s, want %s; case=%+v", got, MatrixStatusSkipped, report.Cases[0])
	}
	if got := report.Cases[0].Reason; got != "source_contract_unsupported" {
		t.Fatalf("reason = %q, want source_contract_unsupported; case=%+v", got, report.Cases[0])
	}
}

func TestMatrixAllPresetsExpandWritersAndAEOpenHosts(t *testing.T) {
	sourceLabels := expandMatrixSourceLabels([]string{"all"})
	targetLabels := expandMatrixTargetLabels([]string{"all"})
	aeOpenLabels := matrixAEOpenLabels([]string{"all"})

	wantWriters := []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024", "AE2025"}
	if strings.Join(sourceLabels, ",") != strings.Join(wantWriters, ",") {
		t.Fatalf("source all = %v, want %v", sourceLabels, wantWriters)
	}
	if strings.Join(targetLabels, ",") != strings.Join(wantWriters, ",") {
		t.Fatalf("target all = %v, want %v", targetLabels, wantWriters)
	}
	wantHosts := []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024", "AE2025"}
	if strings.Join(aeOpenLabels, ",") != strings.Join(wantHosts, ",") {
		t.Fatalf("ae-versions all = %v, want %v", aeOpenLabels, wantHosts)
	}
}
