package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunOutcomePrintsLatestOutcome(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "latest_outcome.json"), `{
  "outcome_status": {"status": "pass"},
  "outcome_summary": {"headline": "pass · 2 projects · recipe smoke 1/1"},
  "corpus": {"parsed_projects": 2, "technique_patterns": 4, "parse_errors": 0},
  "closed_loop": {"batch_passed": 1, "batch_attempted": 1},
  "stable_outputs": {"open_target": "tmp/out/latest_outcome.html"},
  "action_plan": {"next_actions": [
    {"priority": 100, "title": "Resolve top blocker", "detail": "plugin X"}
  ]}
}`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"outcome", "-out-root", root}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run outcome = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Self-Hosted Outcome",
		"Headline: pass · 2 projects · recipe smoke 1/1",
		"Corpus:   2 projects, 4 patterns, 0 parse errors",
		"Smoke:    1/1",
		"Open:     tmp/out/latest_outcome.html",
		"Next Actions",
		"- P100: Resolve top blocker - plugin X",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
}

func TestRunOutcomePrintsExplicitJSONPath(t *testing.T) {
	root := t.TempDir()
	jsonPath := filepath.Join(root, "custom_outcome.json")
	writeFile(t, jsonPath, `{
  "outcome_status": {"status": "warn"},
  "outcome_summary": {"headline": "warn · custom outcome"},
  "corpus": {"parsed_projects": 5, "technique_patterns": 8, "parse_errors": 1},
  "closed_loop": {"batch_passed": 2, "batch_attempted": 3},
  "stable_outputs": {"open_target": "tmp/out/custom.html"}
}`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"outcome", "-json-path", jsonPath}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run outcome = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Headline: warn · custom outcome",
		"Status:   warn",
		"Corpus:   5 projects, 8 patterns, 1 parse errors",
		"Smoke:    2/3",
		"Open:     tmp/out/custom.html",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
}

func TestRunStatusPrintsProcessWatchOutcomeAndLogs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "watch_process.json"), `{
  "mode": "started",
  "pid": 999999,
  "command": "pwsh -File scripts/watch.ps1",
  "stdout_log": "tmp/out/watch.out.log",
  "stderr_log": "tmp/out/watch.err.log"
}`)
	writeFile(t, filepath.Join(root, "watch_status.json"), `{
  "mode": "watch",
  "completed_iterations": 3,
  "last_exit_code": 0
}`)
	writeFile(t, filepath.Join(root, "latest_outcome.json"), `{
  "outcome_status": {"status": "pass"},
  "outcome_summary": {"headline": "pass · 2 projects"},
  "stable_outputs": {"open_target": "tmp/out/latest_outcome.html"}
}`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"status", "-out-root", root}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run status = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Self-Hosted Status",
		"Watch Process",
		"- mode: started",
		"- pid: 999999",
		"- state: not-running",
		"Watch Status",
		"- completed iterations: 3",
		"- last exit code: 0",
		"Latest Outcome",
		"- headline: pass · 2 projects",
		"Logs",
		"- stdout: tmp/out/watch.out.log",
		"- stderr: tmp/out/watch.err.log",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
}

func TestRunRejectsMissingOutcome(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"outcome", "-out-root", t.TempDir()}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("run outcome missing = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "latest_outcome.json") {
		t.Fatalf("stderr = %q, want latest_outcome.json", stderr.String())
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
