package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/host"
	"github.com/yueli-fx/aep-parser/internal/rifx"
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

func TestRunVerifyUsesGoOrchestration(t *testing.T) {
	root := t.TempDir()
	platform := testPlatform()
	verifyRunner := &fakeVerifyRunner{t: t}
	platform.Runner = verifyRunner
	var stdout, stderr bytes.Buffer

	code := run([]string{"verify", "-out-root", root, "-limit", "1"}, &stdout, &stderr, platform)

	if code != 0 {
		t.Fatalf("run verify = %d, stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if _, err := os.Stat(filepath.Join(root, "latest_outcome.json")); err != nil {
		t.Fatalf("latest_outcome.json missing: %v", err)
	}
	for _, cmd := range verifyRunner.commands {
		joined := strings.Join(append([]string{cmd.Name}, cmd.Args...), " ")
		if strings.Contains(joined, ".ps1") {
			t.Fatalf("verify should not call selfhost PS script: %s", joined)
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

func TestRunSampleShellBatchWritesFailureSummary(t *testing.T) {
	root := t.TempDir()
	outDir := filepath.Join(root, "out")
	sampleRoot := filepath.Join(root, "samples")
	writeFile(t, filepath.Join(sampleRoot, "bad.aep"), "not an aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"sample-shell-batch", "-root", sampleRoot, "-out", outDir, "-limit", "1"}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run sample-shell-batch = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "batch summary:") {
		t.Fatalf("stdout missing batch summary:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "effect_params 0/0") {
		t.Fatalf("stdout missing effect param count:\n%s", stdout.String())
	}
	var summary struct {
		Summary struct {
			Total     int `json:"total"`
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
		} `json:"summary"`
		Failures []struct {
			InputPath string `json:"input_path"`
			Error     string `json:"error"`
		} `json:"failures"`
	}
	readFileJSON(t, filepath.Join(outDir, "batch_summary.json"), &summary)
	if summary.Summary.Total != 1 || summary.Summary.Succeeded != 0 || summary.Summary.Failed != 1 {
		t.Fatalf("summary = %+v", summary.Summary)
	}
	if len(summary.Failures) != 1 || summary.Failures[0].InputPath == "" || summary.Failures[0].Error == "" {
		t.Fatalf("failures = %+v", summary.Failures)
	}
}

func TestRunExtractEffectTemplatesWritesCandidates(t *testing.T) {
	root := t.TempDir()
	samplePath := filepath.Join(root, "samples", "one.aep")
	writeMainTestAEPWithEffect(t, samplePath, "ADBE Missing FX")
	summaryPath := filepath.Join(root, "batch_summary.json")
	writeFile(t, summaryPath, `{
  "schema_version": 1,
  "effect_work_items": [
    {"match_name":"ADBE Missing FX","count":1,"class":"native_template_gap","action":"add_effect_template"}
  ]
}`)
	outDir := filepath.Join(root, "candidates")
	var stdout, stderr bytes.Buffer

	code := run([]string{"extract-effect-templates", "-summary", summaryPath, "-root", filepath.Join(root, "samples"), "-out", outDir}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run extract-effect-templates = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"extraction summary:",
		"requested 1 hit 1 missing 0",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "effect_adbe_missing_fx.bin")); err != nil {
		t.Fatalf("candidate missing: %v", err)
	}
}

func TestRunAuditEffectTemplatesWritesSummary(t *testing.T) {
	root := t.TempDir()
	writeMainTestEffectTemplateCandidate(t, filepath.Join(root, "effect_adbe_demo.bin"))
	var stdout, stderr bytes.Buffer

	code := run([]string{"audit-effect-templates", "-candidates", root}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run audit-effect-templates = %d, stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"candidate audit:",
		"total 1 candidate 0 review 0 reject 1 parse_error 0",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "candidate_audit.json")); err != nil {
		t.Fatalf("candidate audit missing: %v", err)
	}
}

func TestRunEffectFieldInventoryWritesJSON(t *testing.T) {
	root := t.TempDir()
	outPath := filepath.Join(root, "inventory.json")
	var stdout, stderr bytes.Buffer

	code := run([]string{"effect-field-inventory", "-root", root, "-out", outPath}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run effect-field-inventory = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "effect field inventory:") {
		t.Fatalf("stdout missing inventory path:\n%s", stdout.String())
	}
	var inventory struct {
		SchemaVersion int `json:"schema_version"`
		Summary       struct {
			ProjectCount int `json:"project_count"`
			EffectKinds  int `json:"effect_kinds"`
		} `json:"summary"`
		Effects []any `json:"effects"`
	}
	readFileJSON(t, outPath, &inventory)
	if inventory.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", inventory.SchemaVersion)
	}
	if inventory.Summary.ProjectCount != 0 || inventory.Summary.EffectKinds != 0 || len(inventory.Effects) != 0 {
		t.Fatalf("inventory = %+v, want empty valid report", inventory)
	}
}

func TestRunEffectFieldUnderstandingWritesJSON(t *testing.T) {
	root := t.TempDir()
	inventoryPath := filepath.Join(root, "inventory.json")
	outPath := filepath.Join(root, "understanding.json")
	writeFile(t, inventoryPath, `{
  "schema_version": 1,
  "summary": {
    "effect_kinds": 1,
    "effect_occurrences": 2,
    "param_kinds": 3,
    "param_occurrences": 4
  },
  "effects": [
    {
      "match_name": "tc Particular",
      "class": "third_party",
      "occurrences": 2,
      "param_kinds": 3,
      "param_occurrences": 4,
      "inferred_capabilities": ["keyframes", "layer_ref", "scalar"]
    }
  ]
}`)
	var stdout, stderr bytes.Buffer

	code := run([]string{"effect-field-understanding", "-inventory", inventoryPath, "-out", outPath}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run effect-field-understanding = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "effect field understanding:") {
		t.Fatalf("stdout missing understanding path:\n%s", stdout.String())
	}
	var understanding struct {
		SchemaVersion int `json:"schema_version"`
		Summary       struct {
			EffectKinds int `json:"effect_kinds"`
		} `json:"summary"`
		Effects []struct {
			MatchName        string `json:"match_name"`
			Reproducibility  string `json:"reproducibility"`
			GenerationPolicy string `json:"generation_policy"`
		} `json:"effects"`
	}
	readFileJSON(t, outPath, &understanding)
	if understanding.SchemaVersion != 1 || understanding.Summary.EffectKinds != 1 || len(understanding.Effects) != 1 {
		t.Fatalf("understanding = %+v", understanding)
	}
	if understanding.Effects[0].Reproducibility != "third_party_plugin_required" || understanding.Effects[0].GenerationPolicy != "preserve_as_dependency" {
		t.Fatalf("effect = %+v", understanding.Effects[0])
	}
}

func TestRunPseudoControllerRebuildProofWritesArtifacts(t *testing.T) {
	root := t.TempDir()
	inventoryPath := filepath.Join(root, "inventory.json")
	understandingPath := filepath.Join(root, "understanding.json")
	outDir := filepath.Join(root, "pseudo-proof")
	writeMainJSONFile(t, inventoryPath, selfhostPseudoControllerInventoryFixture())
	writeMainJSONFile(t, understandingPath, selfhostPseudoControllerUnderstandingFixture())
	var stdout, stderr bytes.Buffer

	code := run([]string{"pseudo-controller-rebuild-proof", "-inventory", inventoryPath, "-understanding", understandingPath, "-out", outDir, "-max", "2"}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run pseudo-controller-rebuild-proof = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pseudo controller rebuild proof:") {
		t.Fatalf("stdout missing proof path:\n%s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "proof.json")); err != nil {
		t.Fatalf("proof.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "generated", "rig-a.aep")); err != nil {
		t.Fatalf("generated proof AEP missing: %v", err)
	}
}

func TestRunPseudoBehaviorWiringPlanWritesJSON(t *testing.T) {
	root := t.TempDir()
	proofPath := filepath.Join(root, "proof.json")
	outDir := filepath.Join(root, "behavior")
	writeMainJSONFile(t, proofPath, selfhostPseudoBehaviorProofFixture())
	var stdout, stderr bytes.Buffer

	code := run([]string{"pseudo-behavior-wiring-plan", "-proof", proofPath, "-out", outDir}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run pseudo-behavior-wiring-plan = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pseudo behavior wiring plan:") {
		t.Fatalf("stdout missing plan path:\n%s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "plan.json")); err != nil {
		t.Fatalf("plan.json missing: %v", err)
	}
}

func TestRunPseudoBehaviorPayloadExtractWritesJSON(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, "plan.json")
	sampleRoot := filepath.Join(root, "samples")
	outDir := filepath.Join(root, "payloads")
	writeMainJSONFile(t, planPath, map[string]any{
		"schema_version": 1,
		"families": []any{
			map[string]any{
				"match_name": "Pseudo/Rig A",
				"tasks": []any{
					map[string]any{
						"param_match_name": "Pseudo/Rig A-0002",
						"action":           "extract_keyframes_then_apply",
						"phase":            "keyframes",
					},
				},
			},
		},
	})
	if err := os.MkdirAll(sampleRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	var stdout, stderr bytes.Buffer

	code := run([]string{"pseudo-behavior-payload-extract", "-plan", planPath, "-root", sampleRoot, "-out", outDir}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run pseudo-behavior-payload-extract = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pseudo behavior payloads:") {
		t.Fatalf("stdout missing payload path:\n%s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "payloads.json")); err != nil {
		t.Fatalf("payloads.json missing: %v", err)
	}
}

func TestRunPseudoBehaviorApplicationWritesJSON(t *testing.T) {
	root := t.TempDir()
	proofPath := filepath.Join(root, "proof.json")
	payloadPath := filepath.Join(root, "payloads.json")
	outDir := filepath.Join(root, "application")
	writeMainJSONFile(t, proofPath, selfhostPseudoBehaviorApplicationProofFixture())
	writeMainJSONFile(t, payloadPath, selfhostPseudoBehaviorApplicationPayloadFixture())
	var stdout, stderr bytes.Buffer

	code := run([]string{"pseudo-behavior-application", "-proof", proofPath, "-payloads", payloadPath, "-out", outDir}, &stdout, &stderr, testPlatform())

	if code != 0 {
		t.Fatalf("run pseudo-behavior-application = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pseudo behavior application:") {
		t.Fatalf("stdout missing application path:\n%s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "application.json")); err != nil {
		t.Fatalf("application.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "generated", "rig-a.aep")); err != nil {
		t.Fatalf("generated application AEP missing: %v", err)
	}
}

func TestRunTechniqueReportWritesArtifacts(t *testing.T) {
	root := t.TempDir()
	outDir := filepath.Join(root, "report")
	platform := testPlatform()
	platform.Runner = fakeTechniqueReportRunner{t: t}
	var stdout, stderr bytes.Buffer

	code := run([]string{"technique-report", "-input", "demo", "-out", outDir, "-limit", "1", "-verify"}, &stdout, &stderr, platform)

	if code != 0 {
		t.Fatalf("run technique-report = %d, stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	for _, path := range []string{
		filepath.Join(outDir, "summary.json"),
		filepath.Join(outDir, "corpus.jsonl"),
		filepath.Join(outDir, "manifest.json"),
		filepath.Join(outDir, "report.html"),
		filepath.Join(outDir, "recipe_drafts.jsonl"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}
}

func selfhostPseudoControllerInventoryFixture() any {
	return map[string]any{
		"schema_version": 1,
		"root":           "data/samples",
		"summary": map[string]any{
			"project_count":      1,
			"effect_kinds":       1,
			"effect_occurrences": 4,
			"param_kinds":        4,
			"param_occurrences":  16,
		},
		"effects": []any{
			map[string]any{
				"match_name":        "Pseudo/Rig A",
				"class":             "pseudo",
				"occurrences":       4,
				"param_kinds":       4,
				"param_occurrences": 16,
				"params": []any{
					map[string]any{"match_name": "Pseudo/Rig A-0000", "value_types": []any{map[string]any{"name": "layer_ref", "count": 4}}, "occurrences": 4, "layer_ref_occurrences": 4},
					map[string]any{"match_name": "Pseudo/Rig A-0001", "value_types": []any{map[string]any{"name": "number", "count": 4}}, "occurrences": 4, "example_static_values": []any{50}},
					map[string]any{"match_name": "Pseudo/Rig A-0002", "value_types": []any{map[string]any{"name": "vector4", "count": 4}}, "occurrences": 4, "example_static_values": []any{[]any{0, 1, 0, 1}}},
					map[string]any{"match_name": "Pseudo/Rig A-0003", "value_types": []any{map[string]any{"name": "bool", "count": 4}}, "occurrences": 4, "example_static_values": []any{true}},
				},
			},
		},
	}
}

func selfhostPseudoControllerUnderstandingFixture() any {
	return map[string]any{
		"schema_version":    1,
		"source_inventory":  "inventory.json",
		"summary":           map[string]any{"effect_kinds": 1, "effect_occurrences": 4, "param_kinds": 4, "param_occurrences": 16},
		"generation_policy": "rebuild_pseudo_controls",
		"effects": []any{
			map[string]any{
				"match_name":          "Pseudo/Rig A",
				"class":               "pseudo",
				"occurrences":         4,
				"param_kinds":         4,
				"param_occurrences":   16,
				"reproducibility":     "pseudo_rebuildable",
				"generation_policy":   "rebuild_pseudo_controls",
				"field_understanding": "param_names_and_values_parseable",
				"study_priority":      100,
			},
		},
	}
}

func selfhostPseudoBehaviorProofFixture() any {
	return map[string]any{
		"schema_version": 1,
		"output_path":    "tmp/pseudo_controller_rebuild/proof.json",
		"summary": map[string]any{
			"pseudo_families":           1,
			"selected_families":         1,
			"generated_families":        1,
			"generated_controls":        3,
			"unsupported_control_count": 0,
		},
		"families": []any{
			map[string]any{
				"match_name":    "Pseudo/Rig A",
				"uid":           "Rig A",
				"status":        "generated",
				"generated_aep": "tmp/pseudo_controller_rebuild/generated/rig-a.aep",
				"controls": []any{
					map[string]any{"param_match_name": "Pseudo/Rig A-0001", "label": "Static", "control_kind": "slider", "value_type": "number"},
					map[string]any{"param_match_name": "Pseudo/Rig A-0002", "label": "Animated", "control_kind": "slider", "value_type": "number", "behavior_notes": []any{"keyframed"}},
					map[string]any{"param_match_name": "Pseudo/Rig A-0003", "label": "Driven", "control_kind": "slider", "value_type": "number", "behavior_notes": []any{"expression"}},
				},
			},
		},
	}
}

func selfhostPseudoBehaviorApplicationProofFixture() any {
	return map[string]any{
		"schema_version": 1,
		"families": []any{
			map[string]any{
				"match_name": "Pseudo/Rig A",
				"uid":        "Rig A",
				"status":     "generated",
				"controls": []any{
					map[string]any{"param_match_name": "Pseudo/Rig A-0001", "label": "Static", "control_kind": "slider", "value_type": "number", "default_value": 0},
					map[string]any{"param_match_name": "Pseudo/Rig A-0002", "label": "Animated", "control_kind": "slider", "value_type": "number", "default_value": 0, "behavior_notes": []any{"keyframed"}},
				},
			},
		},
	}
}

func selfhostPseudoBehaviorApplicationPayloadFixture() any {
	return map[string]any{
		"schema_version": 1,
		"families": []any{
			map[string]any{
				"match_name": "Pseudo/Rig A",
				"status":     "generated",
				"controls": []any{
					map[string]any{
						"param_match_name": "Pseudo/Rig A-0002",
						"action":           "extract_keyframes_then_apply",
						"phase":            "keyframes",
						"status":           "payload_found",
						"examples": []any{
							map[string]any{
								"effect_match_name": "Pseudo/Rig A",
								"param_match_name":  "Pseudo/Rig A-0002",
								"keyframes": []any{
									map[string]any{"time_seconds": 0, "value": 0},
									map[string]any{"time_seconds": 1, "value": 10},
								},
							},
						},
					},
				},
			},
		},
	}
}

func writeMainJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	writeFile(t, path, string(data)+"\n")
}

func TestRunVerifyReportReturnsValidationError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"verify-report", "-out-dir", t.TempDir()}, &stdout, &stderr, testPlatform())

	if code != 1 {
		t.Fatalf("run verify-report = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "missing report artifact") {
		t.Fatalf("stderr = %q, want missing artifact", stderr.String())
	}
}

func TestRunFinalizeRunReturnsValidationError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"finalize-run", "-out-root", t.TempDir(), "-run-root", t.TempDir()}, &stdout, &stderr, testPlatform())

	if code != 1 {
		t.Fatalf("run finalize-run = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "manifest.json") {
		t.Fatalf("stderr = %q, want missing manifest", stderr.String())
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
		summaryPath := argAfter(cmd.Args, "-summary-out")
		if outPath == "" {
			r.t.Fatalf("technique command missing -out: %+v", cmd.Args)
		}
		if summaryPath != "" {
			writeFile(r.t, summaryPath, `{"schema_version":1,"mode":"explain","project_count":1,"error_count":0,"totals":{"comp_count":1,"layer_count":1,"effect_count":1,"shape_operator_count":1,"text_animator_count":0,"dependency_count":1},"pattern_counts":{"shape_system":1}}`+"\n")
			writeFile(r.t, outPath, `{"path":"demo.aep","explanation":{"recreation_steps":[{"id":"structure","priority":10,"title":"Build","summary":"Build structure"}]}}`+"\n")
			break
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

type fakeVerifyRunner struct {
	t        *testing.T
	commands []host.Command
}

func (r *fakeVerifyRunner) Run(_ context.Context, cmd host.Command) host.Result {
	r.commands = append(r.commands, cmd)
	joined := strings.Join(append([]string{cmd.Name}, cmd.Args...), " ")
	switch {
	case strings.Contains(joined, "go test"):
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
		summaryPath := argAfter(cmd.Args, "-summary-out")
		if outPath == "" {
			r.t.Fatalf("technique command missing -out: %+v", cmd.Args)
		}
		if summaryPath != "" {
			writeFile(r.t, summaryPath, `{"schema_version":1,"mode":"explain","project_count":1,"error_count":0,"totals":{"comp_count":1,"layer_count":1,"effect_count":1,"shape_operator_count":1,"text_animator_count":0,"dependency_count":1},"pattern_counts":{"shape_system":1}}`+"\n")
			writeFile(r.t, outPath, `{"path":"demo.aep","explanation":{"recreation_steps":[{"id":"structure","priority":10,"title":"Build","summary":"Build structure"}]}}`+"\n")
			break
		}
		writeFile(r.t, outPath, `{"summary":{"comp_count":1,"layer_count":0}}`+"\n")
	default:
		r.t.Fatalf("unexpected verify command: %s", joined)
	}
	return host.Result{ExitCode: 0}
}

func (r *fakeVerifyRunner) Start(context.Context, host.Command) (host.Process, error) {
	r.t.Fatalf("Start should not be called")
	return nil, nil
}

type fakeTechniqueReportRunner struct {
	t *testing.T
}

func (r fakeTechniqueReportRunner) Run(_ context.Context, cmd host.Command) host.Result {
	joined := strings.Join(append([]string{cmd.Name}, cmd.Args...), " ")
	if !strings.Contains(joined, "./cmd/aeptechnique") {
		r.t.Fatalf("unexpected command: %s", joined)
	}
	outPath := argAfter(cmd.Args, "-out")
	summaryPath := argAfter(cmd.Args, "-summary-out")
	if outPath == "" || summaryPath == "" {
		r.t.Fatalf("missing output args: %+v", cmd.Args)
	}
	writeFile(r.t, summaryPath, `{"schema_version":1,"mode":"explain","project_count":1,"error_count":0,"totals":{"comp_count":1,"layer_count":1,"effect_count":1,"shape_operator_count":1,"text_animator_count":0,"dependency_count":1},"pattern_counts":{"shape_system":1}}`+"\n")
	writeFile(r.t, outPath, `{"path":"demo.aep","explanation":{"recreation_steps":[{"id":"structure","priority":10,"title":"Build","summary":"Build structure"}]}}`+"\n")
	return host.Result{ExitCode: 0}
}

func (r fakeTechniqueReportRunner) Start(context.Context, host.Command) (host.Process, error) {
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

func writeVerifyReportFixture(t *testing.T, outDir string, partial bool) {
	t.Helper()
	errorCount := 0
	patternCount := 1
	if partial {
		errorCount = 1
		patternCount = 1
	}
	writeFile(t, filepath.Join(outDir, "summary.json"), fmt.Sprintf(`{"project_count":1,"error_count":%d,"totals":{"comp_count":1,"layer_count":1,"effect_count":1,"shape_operator_count":1,"text_animator_count":0,"dependency_count":1}}`, errorCount))
	writeFile(t, filepath.Join(outDir, "digest.json"), fmt.Sprintf(`{"project_count":1,"patterns":[{"id":"shape","recreation_steps":[{"id":"structure"}]}],"pattern_count":%d}`, patternCount))
	writeFile(t, filepath.Join(outDir, "manifest.json"), fmt.Sprintf(`{"project_count":1,"error_count":%d,"pattern_count":%d,"git":{"commit":"abc"},"artifacts":["summary.json","digest.json","manifest.json","projects.csv","recipe_drafts.jsonl","report.html","report.md","learning.md"]}`, errorCount, patternCount))
	writeFile(t, filepath.Join(outDir, "corpus.jsonl"), `{"path":"demo.aep","explanation":{"recreation_steps":[{"id":"structure"}]}}`+"\n")
	writeFile(t, filepath.Join(outDir, "recipe_drafts.jsonl"), `{"project_path":"demo.aep","readiness":"analysis_ready","counts":{},"recipe":{"schema_version":1,"project":{},"comps":[{"name":"Main"}],"expected_profile":{"comp_count":1}},"gaps":[{"id":"layers"}]}`+"\n")
	writeFile(t, filepath.Join(outDir, "reconstruction_blueprints.jsonl"), `{"project_path":"demo.aep","readiness":"analysis_ready","counts":{},"phases":[{"id":"create_compositions"},{"id":"create_layers"},{"id":"apply_mechanisms"},{"id":"wire_dependencies"},{"id":"verify_recreation"}]}`+"\n")
	writeFile(t, filepath.Join(outDir, "learning.md"), "## Pattern Playbook\n## Study Queue\nrecreation steps\n## Plugin Risk Queue\n## Readiness Queue\n")
	writeFile(t, filepath.Join(outDir, "report.md"), "## Pattern Representatives\n## Study Queue\nStep 1:\n")
	writeFile(t, filepath.Join(outDir, "report.html"), "Technique Corpus Report Step 1: manifest.json learning.md projects.csv")
	writeCSVFile(t, filepath.Join(outDir, "projects.csv"), "project_path,readiness,readiness_blockers\n"+"demo.aep,analysis_ready,\n")
	writeCSVFile(t, filepath.Join(outDir, "study_queue.csv"), "rank,path,readiness,study_score\n1,demo.aep,analysis_ready,10\n")
	writeCSVFile(t, filepath.Join(outDir, "learning_actions.csv"), "priority,pattern,count,action,representative_project\n100,shape,1,study,demo.aep\n")
	writeCSVFile(t, filepath.Join(outDir, "coverage_scorecard.csv"), "artifact,metric,expected_count,actual_count,status,notes\nprojects.csv,rows,1,1,ok,\n")
	writeCSVFile(t, filepath.Join(outDir, "recreation_blockers.csv"), "project_path,readiness,blocker_type,blocker,action\n")
}

func writeCSVFile(t *testing.T, path, text string) {
	t.Helper()
	writeFile(t, path, text)
}

func writeMainTestAEPWithEffect(t *testing.T, path, effectMatchName string) {
	t.Helper()
	root := &rifx.Chunk{ID: rifx.IDRifx, FormType: rifx.IDEgg, Children: []*rifx.Chunk{
		{
			ID:       rifx.IDList,
			FormType: rifx.IDTdgp,
			Children: []*rifx.Chunk{
				{ID: rifx.IDTdmn, Data: []byte("ADBE Effect Parade")},
				{
					ID:       rifx.IDList,
					FormType: rifx.IDTdgp,
					Children: []*rifx.Chunk{
						{ID: rifx.IDTdmn, Data: []byte(effectMatchName)},
						{ID: rifx.IDList, FormType: rifx.IDSspc, Children: []*rifx.Chunk{
							{ID: rifx.IDFnam, Data: []byte("candidate")},
						}},
						{ID: rifx.IDTdmn, Data: []byte("ADBE Group End")},
					},
				},
				{ID: rifx.IDTdmn, Data: []byte("ADBE Group End")},
			},
		},
	}}
	var buf bytes.Buffer
	if err := root.Write(&buf); err != nil {
		t.Fatalf("write rifx: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write aep: %v", err)
	}
}

func writeMainTestEffectTemplateCandidate(t *testing.T, path string) {
	t.Helper()
	wrapper := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		{ID: rifx.IDTdmn, Data: []byte("ADBE Demo")},
		{ID: rifx.IDList, FormType: rifx.IDSspc, Children: []*rifx.Chunk{
			{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
				{ID: rifx.IDTdmn, Data: []byte("Param One")},
				{ID: rifx.IDList, FormType: rifx.IDTdbs, Children: []*rifx.Chunk{
					{ID: rifx.IDUtf8, Data: []byte("time")},
				}},
			}},
		}},
	}}
	var buf bytes.Buffer
	if err := wrapper.Write(&buf); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write candidate: %v", err)
	}
}
