package aepmigrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
