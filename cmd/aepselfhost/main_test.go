package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/host"
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

	code := run([]string{"outcome", "-out-root", root}, &stdout, &stderr, testPlatform())

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

	code := run([]string{"outcome", "-json-path", jsonPath}, &stdout, &stderr, testPlatform())

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

	code := run([]string{"status", "-out-root", root}, &stdout, &stderr, testPlatform())

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

func TestRunVerifyDryRunUsesGoGate(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := run([]string{"verify", "-out-root", root, "-limit", "3", "-open", "-dry-run"}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run verify dry-run = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if strings.Contains(out, "pwsh") {
		t.Fatalf("dry-run should not depend on pwsh:\n%s", out)
	}
	for _, want := range []string{
		"DRY RUN technique selfhost verify",
		"out_root: " + root,
		"limit: 3",
		"open: true",
		"gate engine: go",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
}

func TestRunRejectsMissingOutcome(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"outcome", "-out-root", t.TempDir()}, &stdout, &stderr, testPlatform())

	if code != 1 {
		t.Fatalf("run outcome missing = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "latest_outcome.json") {
		t.Fatalf("stderr = %q, want latest_outcome.json", stderr.String())
	}
}

func TestRunWatchDryRunWritesStatus(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"watch",
		"-out-root", root,
		"-duration-minutes", "10",
		"-interval-seconds", "15",
		"-iterations", "2",
		"-limit", "7",
		"-open-first",
		"-dry-run",
	}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run watch dry-run = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"DRY RUN technique selfhost watch",
		"verify command:",
		"aepselfhost verify",
		"-limit 7",
		"-open",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
	var status watchFile
	readFileJSON(t, filepath.Join(root, "watch_status.json"), &status)
	if status.Mode != "dry_run" {
		t.Fatalf("watch status mode = %q, want dry_run", status.Mode)
	}
	if status.DurationMinutes != 10 || status.IntervalSeconds != 15 || status.PlannedIterations != 2 {
		t.Fatalf("watch status = %+v", status)
	}
	if !strings.Contains(status.VerifyCommand, "-limit 7") {
		t.Fatalf("verify command = %q, want limit", status.VerifyCommand)
	}
}

func TestRunStartWatchDryRunWritesProcessStatus(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := run([]string{
		"start-watch",
		"-out-root", root,
		"-duration-minutes", "10",
		"-interval-seconds", "15",
		"-iterations", "2",
		"-limit", "7",
		"-open-first",
		"-dry-run",
	}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run start-watch dry-run = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"DRY RUN technique selfhost background start",
		"stdout:",
		"stderr:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
	var status processFile
	readFileJSON(t, filepath.Join(root, "watch_process.json"), &status)
	if status.Mode != "dry_run" {
		t.Fatalf("process status mode = %q, want dry_run", status.Mode)
	}
	if !strings.Contains(status.Command, "watch") {
		t.Fatalf("process command = %q, want watch", status.Command)
	}
	if status.StdoutLog == "" || status.StderrLog == "" {
		t.Fatalf("process logs = stdout %q stderr %q", status.StdoutLog, status.StderrLog)
	}
}

func TestRunWatchRejectsZeroWork(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"watch", "-out-root", t.TempDir(), "-duration-minutes", "0"}, &stdout, &stderr, testPlatform())

	if code != 1 {
		t.Fatalf("run watch zero work = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "Iterations and DurationMinutes cannot both be 0") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunCompareReportsWritesArtifacts(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "base")
	newDir := filepath.Join(root, "new")
	outDir := filepath.Join(root, "out")
	writeFile(t, filepath.Join(baseDir, "summary.json"), `{
  "project_count": 1,
  "totals": {"comp_count": 1, "layer_count": 2, "effect_count": 3},
  "effect_counts": {"ADBE Fill": 2}
}`)
	writeFile(t, filepath.Join(newDir, "summary.json"), `{
  "project_count": 2,
  "totals": {"comp_count": 2, "layer_count": 4, "effect_count": 6},
  "effect_counts": {"ADBE Fill": 1, "ADBE Slider Control": 3}
}`)
	writeFile(t, filepath.Join(baseDir, "digest.json"), `{"patterns":["kinetic_text"]}`)
	writeFile(t, filepath.Join(newDir, "digest.json"), `{"patterns":["kinetic_text","precomp_effect_pipeline"]}`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"compare-reports", "-base", baseDir, "-new", newDir, "-out", outDir, "-top", "5"}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run compare-reports = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "compare json:") {
		t.Fatalf("stdout missing compare json:\n%s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "compare.json")); err != nil {
		t.Fatalf("compare.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "compare.md")); err != nil {
		t.Fatalf("compare.md missing: %v", err)
	}
}

func TestRunRecipeSmokeWritesArtifacts(t *testing.T) {
	root := t.TempDir()
	fullReport := filepath.Join(root, "full_report")
	writeFile(t, filepath.Join(fullReport, "recipe_drafts.jsonl"), `{"project_path":"demo.aep","recipe":{"schema_version":1,"expected_profile":{"comp_count":1,"layer_count":0}}}`+"\n")
	platform := testPlatform()
	platform.Runner = fakeRecipeSmokeRunner{t: t}
	var stdout, stderr bytes.Buffer

	code := run([]string{"recipe-smoke", "-full-report", fullReport, "-run-root", root, "-batch-limit", "1"}, &stdout, &stderr, platform)

	if code != 0 {
		t.Fatalf("run recipe-smoke = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "recipe draft batch summary:") {
		t.Fatalf("stdout missing batch summary:\n%s", stdout.String())
	}
	for _, path := range []string{
		filepath.Join(root, "recipe_draft_compile", "recipe_draft.aep"),
		filepath.Join(root, "recipe_draft_reparse", "reparse_summary.json"),
		filepath.Join(root, "recipe_draft_batch", "summary.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
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

func readFileJSON(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
}

func testPlatform() host.Platform {
	return host.Platform{
		Runner:           fakeRunner{},
		ProcessInspector: fakeInspector{},
		BrowserOpener:    host.NoopBrowserOpener{},
	}
}

type fakeRunner struct{}

func (fakeRunner) Run(context.Context, host.Command) host.Result {
	return host.Result{ExitCode: 0}
}

func (fakeRunner) Start(context.Context, host.Command) (host.Process, error) {
	return fakeProcess{pid: 12345}, nil
}

type fakeRecipeSmokeRunner struct {
	t *testing.T
}

func (r fakeRecipeSmokeRunner) Run(_ context.Context, cmd host.Command) host.Result {
	joined := strings.Join(append([]string{cmd.Name}, cmd.Args...), " ")
	switch {
	case strings.Contains(joined, "./cmd/aeprecipe validate"):
		writeCommandOutput(r.t, cmd.Stdout, `{"valid":true}`+"\n")
	case strings.Contains(joined, "./cmd/aeprecipe compile"):
		writeCommandOutput(r.t, cmd.Stdout, `{"valid":true}`+"\n")
		outPath := argAfter(cmd.Args, "-out")
		if outPath == "" {
			r.t.Fatalf("compile command missing -out: %+v", cmd.Args)
		}
		writeFile(r.t, outPath, "fake aep")
	case strings.Contains(joined, "./cmd/aeptechnique"):
		outPath := argAfter(cmd.Args, "-out")
		if outPath == "" {
			r.t.Fatalf("technique command missing -out: %+v", cmd.Args)
		}
		writeFile(r.t, outPath, `{"summary":{"comp_count":1,"layer_count":0}}`+"\n")
	default:
		r.t.Fatalf("unexpected command: %s", joined)
	}
	return host.Result{ExitCode: 0}
}

func (r fakeRecipeSmokeRunner) Start(context.Context, host.Command) (host.Process, error) {
	r.t.Fatalf("Start should not be called")
	return nil, nil
}

type fakeProcess struct {
	pid int
}

func (p fakeProcess) PID() int {
	return p.pid
}

func (p fakeProcess) Release() error {
	return nil
}

type fakeInspector struct{}

func (fakeInspector) IsRunning(pid int) bool {
	return pid == os.Getpid()
}

func writeCommandOutput(t *testing.T, w io.Writer, text string) {
	t.Helper()
	if _, err := io.WriteString(w, text); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
}

func argAfter(args []string, name string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1]
		}
	}
	return ""
}
