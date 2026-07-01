package aepmigrate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aehost"
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

func TestRunMatrixAEOpenUsesAbsoluteWorkerPaths(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "minimal.json")
	writeMinimalMatrixRecipe(t, recipePath)
	outRoot := filepath.Join(root, "matrix")
	host := &fakeMatrixHost{doneBody: "PASS\nproject items.length=1\ncomp=Main layers.length=0\n"}

	report, err := RunMatrix(MatrixOptions{
		RecipePaths: []string{recipePath},
		SourceLabels: []string{
			"AE2020",
		},
		TargetLabels: []string{
			"AE2025",
		},
		OutRoot: outRoot,
		AEHosts: map[string]string{
			"AE2025": "AfterFX.exe",
		},
		AEOpen: true,
		Host:   host,
	})
	if err != nil {
		t.Fatalf("RunMatrix: %v", err)
	}
	if report.Summary.Passed != 1 {
		t.Fatalf("passed = %d, want 1; cases=%+v", report.Summary.Passed, report.Cases)
	}
	if !host.called {
		t.Fatal("fake AE host was not called")
	}
	if !filepath.IsAbs(host.request.JSXPath) {
		t.Fatalf("JSXPath = %q, want absolute", host.request.JSXPath)
	}
	if !filepath.IsAbs(host.request.DonePath) {
		t.Fatalf("DonePath = %q, want absolute", host.request.DonePath)
	}
	if got := host.request.Env["AE_OPEN_ARGS"]; got == "" || !filepath.IsAbs(filepath.FromSlash(got)) {
		t.Fatalf("AE_OPEN_ARGS = %q, want absolute", got)
	}
}

func TestRunMatrixAEOpenCanValidateOneTargetAcrossMultipleAEHosts(t *testing.T) {
	root := t.TempDir()
	recipePath := filepath.Join(root, "minimal.json")
	writeMinimalMatrixRecipe(t, recipePath)
	outRoot := filepath.Join(root, "matrix")
	host := &fakeMatrixHost{doneBody: "PASS\nproject items.length=1\ncomp=Main layers.length=0\n"}

	report, err := RunMatrix(MatrixOptions{
		RecipePaths: []string{recipePath},
		SourceLabels: []string{
			"AE2020",
		},
		TargetLabels: []string{
			"AE2020",
		},
		AEOpenLabels: []string{
			"AE2020",
			"AE2021",
			"AE2025",
		},
		OutRoot: outRoot,
		AEHosts: map[string]string{
			"AE2020": "AfterFX2020.exe",
			"AE2021": "AfterFX2021.exe",
			"AE2025": "AfterFX2025.exe",
		},
		AEOpen: true,
		Host:   host,
	})
	if err != nil {
		t.Fatalf("RunMatrix: %v", err)
	}
	if report.Summary.Total != 3 || report.Summary.Passed != 3 {
		t.Fatalf("summary = %+v, want total=3 passed=3; cases=%+v", report.Summary, report.Cases)
	}
	if len(host.requests) != 3 {
		t.Fatalf("AE host calls = %d, want 3", len(host.requests))
	}
	wantPaths := map[string]bool{
		"AfterFX2020.exe": true,
		"AfterFX2021.exe": true,
		"AfterFX2025.exe": true,
	}
	for _, req := range host.requests {
		delete(wantPaths, req.AEPath)
	}
	if len(wantPaths) != 0 {
		t.Fatalf("missing AE host requests for %+v; requests=%+v", wantPaths, host.requests)
	}
	wantLabels := map[string]bool{
		"AE2020": true,
		"AE2021": true,
		"AE2025": true,
	}
	for _, c := range report.Cases {
		if c.TargetVersion != "AE2020" {
			t.Fatalf("target = %s, want AE2020; case=%+v", c.TargetVersion, c)
		}
		delete(wantLabels, c.AEOpenVersion)
	}
	if len(wantLabels) != 0 {
		t.Fatalf("missing AE open versions %+v; cases=%+v", wantLabels, report.Cases)
	}
}

func TestRunMatrixAEOpenRefusesCaseCountAboveLimit(t *testing.T) {
	root := t.TempDir()
	firstRecipePath := filepath.Join(root, "minimal-a.json")
	secondRecipePath := filepath.Join(root, "minimal-b.json")
	writeMinimalMatrixRecipe(t, firstRecipePath)
	writeMinimalMatrixRecipe(t, secondRecipePath)
	outRoot := filepath.Join(root, "matrix")
	host := &fakeMatrixHost{doneBody: "PASS\n"}

	_, err := RunMatrix(MatrixOptions{
		RecipePaths: []string{
			firstRecipePath,
			secondRecipePath,
		},
		SourceLabels: []string{
			"AE2020",
		},
		TargetLabels: []string{
			"AE2020",
			"AE2025",
		},
		OutRoot: outRoot,
		AEHosts: map[string]string{
			"AE2020": "AfterFX2020.exe",
			"AE2025": "AfterFX2025.exe",
		},
		AEOpen:         true,
		MaxAEOpenCases: 3,
		Host:           host,
	})
	if err == nil || !strings.Contains(err.Error(), "AE open matrix would run 4 cases") {
		t.Fatalf("RunMatrix error = %v, want AE open case limit", err)
	}
	if host.called {
		t.Fatal("AE host was called despite exceeding the AE-open case limit")
	}
	if _, statErr := os.Stat(filepath.Join(outRoot, "matrix.json")); !os.IsNotExist(statErr) {
		t.Fatalf("matrix.json exists or stat failed unexpectedly: %v", statErr)
	}
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

type fakeMatrixHost struct {
	called   bool
	request  aehost.ScriptRequest
	requests []aehost.ScriptRequest
	doneBody string
}

func (h *fakeMatrixHost) Available(context.Context) aehost.Availability {
	return aehost.Availability{Status: aehost.CapabilityAvailable}
}

func (h *fakeMatrixHost) RunScript(_ context.Context, req aehost.ScriptRequest) (aehost.ScriptResult, error) {
	h.called = true
	h.request = req
	h.requests = append(h.requests, req)
	if err := os.MkdirAll(filepath.Dir(req.DonePath), 0o755); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	if err := os.WriteFile(req.DonePath, []byte(h.doneBody), 0o644); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	return aehost.ScriptResult{ExitCode: 0, DonePath: req.DonePath}, nil
}

func assertMatrixCaseStatus(t *testing.T, cases []MatrixCase, target string, want MatrixStatus) {
	t.Helper()
	for _, c := range cases {
		if c.TargetVersion == target {
			if c.Status != want {
				t.Fatalf("%s status = %s, want %s; case=%+v", target, c.Status, want, c)
			}
			return
		}
	}
	t.Fatalf("case for target %s not found in %+v", target, cases)
}
