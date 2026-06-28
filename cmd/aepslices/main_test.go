package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/example/aep-parser/internal/sliceworkflow"
)

func TestRunPlanWritesJSONReport(t *testing.T) {
	root := repoRoot(t)
	out := filepath.Join(t.TempDir(), "slices.json")
	code := run([]string{
		"plan",
		"-aep", filepath.Join(root, "flightdeck", "showcase", "text", "text.aep"),
		"-dict", "",
		"-max", "1",
		"-json",
		"-out", out,
	})
	if code != 0 {
		t.Fatalf("run(plan) = %d, want 0", code)
	}

	report := readReport(t, out)
	if report.Mode != "plan" {
		t.Fatalf("Mode = %q, want plan", report.Mode)
	}
	if report.Summary.SelectedSliceCount != 1 || len(report.Slices) != 1 {
		t.Fatalf("selected slices = %d/%d, want 1/1", report.Summary.SelectedSliceCount, len(report.Slices))
	}
}

func TestRunDiagnoseReturnsOneForDifferentAEPs(t *testing.T) {
	root := repoRoot(t)
	out := filepath.Join(t.TempDir(), "diagnose.json")
	code := run([]string{
		"diagnose",
		"-expected", filepath.Join(root, "flightdeck", "showcase", "text", "text.aep"),
		"-actual", filepath.Join(root, "flightdeck", "showcase", "effects", "effects.aep"),
		"-dict", "",
		"-json",
		"-out", out,
	})
	if code != 1 {
		t.Fatalf("run(diagnose different) = %d, want 1", code)
	}

	report := readReport(t, out)
	if report.Mode != "diagnose" {
		t.Fatalf("Mode = %q, want diagnose", report.Mode)
	}
	if report.Summary.GapCount == 0 || len(report.GapReports) == 0 {
		t.Fatalf("gap summary = %+v reports=%d, want non-zero", report.Summary, len(report.GapReports))
	}
}

func TestRunReturnsUsageForMissingSubcommand(t *testing.T) {
	if code := run(nil); code != 2 {
		t.Fatalf("run(nil) = %d, want 2", code)
	}
}

func readReport(t *testing.T, path string) sliceworkflow.Report {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report sliceworkflow.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	return report
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
