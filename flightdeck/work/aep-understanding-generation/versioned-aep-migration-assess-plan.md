# Versioned AEP Migration Assess Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first `aepmigrate assess` slice: inspect a source `.aep`, normalize source/target AE versions, classify obvious version compatibility entries, and emit a JSON migration assessment report without writing a migrated `.aep`.

**Architecture:** Create a small `internal/aepmigrate` package that owns version parsing, report schema, a minimal compatibility ledger, and assess logic over `internal/profile`. Add `cmd/aepmigrate` as a thin CLI. This slice is read-only and Pure Go; `convert` and AE render verification stay out of scope.

**Tech Stack:** Go standard library, `internal/aep`, `internal/profile`, table-driven tests, JSON CLI output.

---

## Files

- Create: `internal/aepmigrate/model.go`
  - Report types, status constants, migration class constants.
- Create: `internal/aepmigrate/version.go`
  - Parse target flags and normalize source version strings.
- Create: `internal/aepmigrate/ledger.go`
  - First small compatibility ledger and path classification helpers.
- Create: `internal/aepmigrate/assess.go`
  - Open source AEP, build profile, classify entries, write report object.
- Create: `internal/aepmigrate/version_test.go`
- Create: `internal/aepmigrate/assess_test.go`
- Create: `cmd/aepmigrate/main.go`
- Create: `cmd/aepmigrate/main_test.go`
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
  - Mark assess slice as implemented after the code lands.
- Modify: `flightdeck/cockpit.md`
  - Replace the "next" entry with implementation status after the code lands.

## Task 1: Version Parser And Report Model

**Files:**
- Create: `internal/aepmigrate/model.go`
- Create: `internal/aepmigrate/version.go`
- Test: `internal/aepmigrate/version_test.go`

- [ ] **Step 1: Write failing target-version tests**

Create `internal/aepmigrate/version_test.go`:

```go
package aepmigrate

import "testing"

func TestParseVersionLabelAcceptsSupportedTargets(t *testing.T) {
	tests := map[string]VersionLabel{
		"AE2020": VersionAE2020,
		"2020":   VersionAE2020,
		"ae2022": VersionAE2022,
		"2022":   VersionAE2022,
		"AE2025": VersionAE2025,
		"2025":   VersionAE2025,
	}
	for input, want := range tests {
		got, err := ParseVersionLabel(input)
		if err != nil {
			t.Fatalf("ParseVersionLabel(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseVersionLabel(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseVersionLabelRejectsUnsupportedTarget(t *testing.T) {
	if _, err := ParseVersionLabel("AE2019"); err == nil {
		t.Fatal("ParseVersionLabel(AE2019) succeeded, want error")
	}
}

func TestNormalizeSourceVersionKeepsUnknownHonest(t *testing.T) {
	if got := NormalizeSourceVersion(""); got.Label != VersionUnknown {
		t.Fatalf("empty source label = %q, want unknown", got.Label)
	}
}
```

- [ ] **Step 2: Run red test**

Run:

```powershell
go test ./internal/aepmigrate -run TestParseVersionLabel -count=1
```

Expected: package does not compile because `internal/aepmigrate` does not exist.

- [ ] **Step 3: Add model and parser implementation**

Create `internal/aepmigrate/model.go`:

```go
package aepmigrate

const SchemaVersion = 1

type VersionLabel string

const (
	VersionUnknown VersionLabel = "unknown"
	VersionAE2020  VersionLabel = "AE2020"
	VersionAE2022  VersionLabel = "AE2022"
	VersionAE2025  VersionLabel = "AE2025"
)

type Status string

const (
	StatusPass    Status = "pass"
	StatusWarn    Status = "warn"
	StatusBlocked Status = "blocked"
	StatusError   Status = "error"
)

type MigrationClass string

const (
	ClassPreserved    MigrationClass = "preserved"
	ClassRetargeted   MigrationClass = "retargeted"
	ClassTranslated   MigrationClass = "translated"
	ClassApproximated MigrationClass = "approximated"
	ClassDropped      MigrationClass = "dropped"
	ClassBlocked      MigrationClass = "blocked"
	ClassUnknown      MigrationClass = "unknown"
)

type SourceInfo struct {
	Path       string       `json:"path"`
	Version   VersionLabel `json:"version_label"`
	VersionRaw string      `json:"version_raw,omitempty"`
}

type TargetInfo struct {
	Version VersionLabel `json:"version_label"`
	Path    string       `json:"path,omitempty"`
}

type Summary struct {
	Status       Status `json:"status"`
	Preserved    int    `json:"preserved"`
	Retargeted   int    `json:"retargeted"`
	Translated   int    `json:"translated"`
	Approximated int    `json:"approximated"`
	Dropped      int    `json:"dropped"`
	Blocked      int    `json:"blocked"`
	Unknown      int    `json:"unknown"`
}

type Entry struct {
	Path          string         `json:"path"`
	Class         MigrationClass `json:"class"`
	TargetVersion VersionLabel  `json:"target_version"`
	Reason        string        `json:"reason"`
	CapabilityKey string        `json:"capability_key,omitempty"`
}

type Verification struct {
	ProfileDiffStatus string `json:"profile_diff_status"`
	AEOpenStatus      string `json:"ae_open_status"`
	RenderStatus      string `json:"render_status"`
}

type Report struct {
	SchemaVersion int          `json:"schema_version"`
	Source        SourceInfo   `json:"source"`
	Target        TargetInfo   `json:"target"`
	Summary       Summary      `json:"summary"`
	Entries       []Entry      `json:"entries,omitempty"`
	Verification  Verification `json:"verification"`
}
```

Create `internal/aepmigrate/version.go`:

```go
package aepmigrate

import (
	"fmt"
	"strings"
)

type SourceVersion struct {
	Label VersionLabel
	Raw   string
}

func ParseVersionLabel(value string) (VersionLabel, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "AE2020", "2020":
		return VersionAE2020, nil
	case "AE2022", "2022":
		return VersionAE2022, nil
	case "AE2025", "2025":
		return VersionAE2025, nil
	default:
		return "", fmt.Errorf("target version must be AE2020, AE2022, or AE2025")
	}
}

func NormalizeSourceVersion(raw string) SourceVersion {
	trimmed := strings.TrimSpace(raw)
	upper := strings.ToUpper(trimmed)
	switch {
	case strings.Contains(upper, "2025") || strings.HasPrefix(upper, "25."):
		return SourceVersion{Label: VersionAE2025, Raw: trimmed}
	case strings.Contains(upper, "2022") || strings.HasPrefix(upper, "22."):
		return SourceVersion{Label: VersionAE2022, Raw: trimmed}
	case strings.Contains(upper, "2020") || strings.HasPrefix(upper, "17."):
		return SourceVersion{Label: VersionAE2020, Raw: trimmed}
	default:
		return SourceVersion{Label: VersionUnknown, Raw: trimmed}
	}
}
```

- [ ] **Step 4: Run green test**

Run:

```powershell
go test ./internal/aepmigrate -run 'TestParseVersionLabel|TestNormalizeSourceVersion' -count=1
```

Expected: tests pass.

- [ ] **Step 5: Commit**

Run:

```powershell
git add internal/aepmigrate/model.go internal/aepmigrate/version.go internal/aepmigrate/version_test.go
git commit -m "feat(aepmigrate): add version assessment model"
```

## Task 2: Assess Engine

**Files:**
- Create: `internal/aepmigrate/ledger.go`
- Create: `internal/aepmigrate/assess.go`
- Test: `internal/aepmigrate/assess_test.go`

- [ ] **Step 1: Write failing assess tests**

Create `internal/aepmigrate/assess_test.go`:

```go
package aepmigrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestAssessReportsPassForEmptyCompatibleProject(t *testing.T) {
	source := writeTempProject(t, aep.TargetAE2020)

	report, err := Assess(Options{InputPath: source, Target: VersionAE2020})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if report.SchemaVersion != SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", report.SchemaVersion, SchemaVersion)
	}
	if report.Target.Version != VersionAE2020 {
		t.Fatalf("target = %q, want AE2020", report.Target.Version)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.ProfileDiffStatus != "not_run" {
		t.Fatalf("ProfileDiffStatus = %q, want not_run", report.Verification.ProfileDiffStatus)
	}
}

func TestAssessBlocksExplicitMatteDowngrade(t *testing.T) {
	source := writeTempProjectWithExplicitMatte(t)

	report, err := Assess(Options{InputPath: source, Target: VersionAE2020})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if report.Summary.Status != StatusBlocked {
		t.Fatalf("status = %q, want blocked; entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Summary.Blocked == 0 {
		t.Fatalf("blocked count = 0, entries=%+v", report.Entries)
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

func writeTempProjectWithExplicitMatte(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2025)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	matte, err := aep.NewSolidLayer(comp, "Matte", 640, 360, [3]uint8{255, 255, 255})
	if err != nil {
		t.Fatalf("NewSolidLayer matte: %v", err)
	}
	fill, err := aep.NewSolidLayer(comp, "Fill", 640, 360, [3]uint8{255, 0, 0})
	if err != nil {
		t.Fatalf("NewSolidLayer fill: %v", err)
	}
	if err := fill.SetTrackMatte(aep.TrackMatteAlpha); err != nil {
		t.Fatalf("SetTrackMatte: %v", err)
	}
	if err := fill.SetTrackMatteSource(matte); err != nil {
		t.Fatalf("SetTrackMatteSource: %v", err)
	}
	path := filepath.Join(t.TempDir(), "explicit-matte.aep")
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
```

- [ ] **Step 2: Run red test**

Run:

```powershell
go test ./internal/aepmigrate -run TestAssess -count=1
```

Expected: compile fails because `Options` and `Assess` do not exist.

- [ ] **Step 3: Add assess implementation**

Create `internal/aepmigrate/ledger.go`:

```go
package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func classifyProfile(target VersionLabel, prof profile.Profile) []Entry {
	var entries []Entry
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if layer.MatteRef != nil && target != VersionAE2025 {
				entries = append(entries, Entry{
					Path:          layer.Path.String() + ".matte_ref",
					Class:         ClassBlocked,
					TargetVersion: target,
					Reason:        "Explicit matte source requires AE2025 in the current writer contract.",
					CapabilityKey: "layer.set_track_matte_source",
				})
			}
		}
	}
	if len(entries) == 0 {
		entries = append(entries, Entry{
			Path:          "project",
			Class:         ClassPreserved,
			TargetVersion: target,
			Reason:        "No version-specific blockers detected in the first assess slice.",
		})
	}
	return entries
}
```

Create `internal/aepmigrate/assess.go`:

```go
package aepmigrate

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

type Options struct {
	InputPath string
	Target    VersionLabel
}

func Assess(opts Options) (Report, error) {
	project, err := aep.Open(opts.InputPath)
	if err != nil {
		return Report{}, err
	}
	prof, err := profile.Build(project, profile.Options{Path: opts.InputPath})
	if err != nil {
		return Report{}, err
	}
	sourceVersion := NormalizeSourceVersion(aep.NewApplication(project).Version())
	entries := classifyProfile(opts.Target, prof)
	report := Report{
		SchemaVersion: SchemaVersion,
		Source: SourceInfo{
			Path:       opts.InputPath,
			Version:    sourceVersion.Label,
			VersionRaw: sourceVersion.Raw,
		},
		Target: TargetInfo{Version: opts.Target},
		Entries: entries,
		Verification: Verification{
			ProfileDiffStatus: "not_run",
			AEOpenStatus:      "not_run",
			RenderStatus:      "not_run",
		},
	}
	report.Summary = summarize(entries)
	return report, nil
}

func summarize(entries []Entry) Summary {
	summary := Summary{Status: StatusPass}
	for _, entry := range entries {
		switch entry.Class {
		case ClassPreserved:
			summary.Preserved++
		case ClassRetargeted:
			summary.Retargeted++
		case ClassTranslated:
			summary.Translated++
		case ClassApproximated:
			summary.Approximated++
			if summary.Status == StatusPass {
				summary.Status = StatusWarn
			}
		case ClassDropped:
			summary.Dropped++
			if summary.Status == StatusPass {
				summary.Status = StatusWarn
			}
		case ClassBlocked:
			summary.Blocked++
			summary.Status = StatusBlocked
		default:
			summary.Unknown++
			if summary.Status != StatusBlocked {
				summary.Status = StatusWarn
			}
		}
	}
	return summary
}
```

- [ ] **Step 4: Run green test**

Run:

```powershell
go test ./internal/aepmigrate -run TestAssess -count=1
```

Expected: tests pass.

- [ ] **Step 5: Commit**

Run:

```powershell
git add internal/aepmigrate/ledger.go internal/aepmigrate/assess.go internal/aepmigrate/assess_test.go
git commit -m "feat(aepmigrate): assess target version compatibility"
```

## Task 3: CLI

**Files:**
- Create: `cmd/aepmigrate/main.go`
- Test: `cmd/aepmigrate/main_test.go`

- [ ] **Step 1: Write failing CLI tests**

Create `cmd/aepmigrate/main_test.go`:

```go
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
		Summary struct {
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
```

- [ ] **Step 2: Run red test**

Run:

```powershell
go test ./cmd/aepmigrate -count=1
```

Expected: package does not exist.

- [ ] **Step 3: Add CLI implementation**

Create `cmd/aepmigrate/main.go`:

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aepmigrate"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: aepmigrate assess -in source.aep -target AE2020 [-out assess.json]")
		return 2
	}
	switch args[0] {
	case "assess":
		return runAssess(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		return 2
	}
}

func runAssess(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepmigrate assess", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("in", "", "source .aep path")
	targetRaw := fs.String("target", "", "target AE version: AE2020, AE2022, or AE2025")
	outPath := fs.String("out", "", "optional JSON output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" || *targetRaw == "" {
		fmt.Fprintln(stderr, "usage: aepmigrate assess -in source.aep -target AE2020 [-out assess.json]")
		return 2
	}
	target, err := aepmigrate.ParseVersionLabel(*targetRaw)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report, err := aepmigrate.Assess(aepmigrate.Options{InputPath: *input, Target: target})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data = append(data, '\n')
	if *outPath != "" {
		if err := os.WriteFile(*outPath, data, 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "migration assess: %s\n", *outPath)
		return statusCode(report.Summary.Status)
	}
	fmt.Fprint(stdout, string(data))
	return statusCode(report.Summary.Status)
}

func statusCode(status aepmigrate.Status) int {
	if status == aepmigrate.StatusBlocked || status == aepmigrate.StatusError {
		return 1
	}
	return 0
}
```

- [ ] **Step 4: Run green test**

Run:

```powershell
go test ./cmd/aepmigrate -count=1
```

Expected: tests pass.

- [ ] **Step 5: Commit**

Run:

```powershell
git add cmd/aepmigrate/main.go cmd/aepmigrate/main_test.go
git commit -m "feat(aepmigrate): add assess command"
```

## Task 4: Docs, Verification, And Final Commit

**Files:**
- Modify: `flightdeck/work/aep-understanding-generation/index.md`
- Modify: `flightdeck/cockpit.md`

- [ ] **Step 1: Update work docs**

In `flightdeck/work/aep-understanding-generation/index.md`, append under `Progress`:

```markdown
- Versioned AEP migration assess slice is implemented as `internal/aepmigrate`
  and `cmd/aepmigrate assess`. It detects/normalizes source and target version
  labels, builds a stable profile, reports first-slice compatibility entries,
  and blocks AE2025 explicit matte downgrades to AE2020/AE2022.
```

In `flightdeck/cockpit.md`, change the `versioned AEP migration` next entry to:

```markdown
- **versioned AEP migration**：`cmd/aepmigrate assess` 已落地为第一片，只做版本识别与兼容评估，不写目标 AEP；下一步再考虑 conservative convert。
```

- [ ] **Step 2: Run full verification**

Run:

```powershell
go test ./internal/aepmigrate ./cmd/aepmigrate -count=1
go test ./... -count=1
go vet ./...
git diff --check
```

Expected: all commands exit `0`.

- [ ] **Step 3: Commit docs**

Run:

```powershell
git add flightdeck/work/aep-understanding-generation/index.md flightdeck/cockpit.md
git commit -m "docs: record versioned migration assess slice"
```

## Self-Review

- Spec coverage: this plan covers the first useful slice from
  `versioned-aep-migration-spec.md`: assess only, version parsing, profile build,
  a small compatibility ledger, JSON report, and tests.
- Scope control: no convert, no blind chunk copying, no AE automation dependency.
- Type consistency: `VersionLabel`, `Status`, `MigrationClass`, `Report`,
  `Options`, and `Assess` are defined before CLI use.
- Verification: package tests, full Go tests, vet, and diff whitespace checks are
  required before claiming completion.
