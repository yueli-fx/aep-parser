# Cross-Platform Architecture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `aep-parser` usable as a cross-platform parser/service while keeping AE verification as an optional platform-specific worker capability.

**Architecture:** Split the system by capability, not by operating system. Pure parsing/profile/recipe/search/report code must be Go-only and portable across Windows, macOS, and Linux; selfhost orchestration should move behind Go host interfaces; AE automation should become an optional worker adapter with Windows and macOS implementations and a Linux unavailable implementation.

**Tech Stack:** Go, build tags, `context.Context`, `os/exec`, `syscall`, existing `cmd/aepselfhost`, existing `cmd/aeoracle`, existing PowerShell scripts as compatibility wrappers during migration.

---

## State

This work package defines the cross-platform direction after the selfhost CLI
started moving from PowerShell into `cmd/aepselfhost`.

Authoritative decision:

- The long-term public service target is cross-platform.
- Linux and macOS service nodes must be able to parse, profile, diff, search,
  draft recipes, and generate learning reports without AE.
- AE validation and render comparison are optional worker capabilities, not
  requirements for the parser service.
- Linux does not provide AE validation. It should report that capability as
  unavailable, not fail unrelated parser workflows.
- Windows and macOS AE automation may differ internally, but callers should use
  the same Go capability interface.

## Platform Tiers

### Tier 1: Pure Go Core

Must run on Windows, macOS, and Linux without PowerShell or AE.

Includes:

- `internal/aep`
- `internal/scene`
- `internal/serializer`
- `internal/profile`
- `internal/profilediff`
- `internal/recipe`
- `internal/projectindex`
- `internal/technique`
- pure CLI modes in `cmd/aepsearch`, `cmd/aeprecipe`, `cmd/aeptechnique`,
  `cmd/aepdiff`, and non-render `cmd/aeoracle` modes.

Rules:

- No direct `pwsh`, `powershell.exe`, `tasklist`, `Start-Process`, `AfterFX.exe`,
  or hard-coded drive paths.
- No GUI automation.
- Use `filepath` for filesystem paths.
- If a field contains an AE-authored Windows path, preserve the data as project
  content; do not treat that as a host runtime path.

### Tier 2: Service Layer

Runs on Windows, macOS, or Linux.

Includes the future parser API service:

- upload `.aep`
- parse and profile
- produce JSON/profile/report artifacts
- search across one project or a corpus
- generate recipe drafts and learning facts
- expose capability status

Rules:

- The service must not require AE to start.
- The service must not execute AE from a request handler.
- The service can enqueue optional AE jobs only when an AE worker is registered.
- Uploaded projects and output artifacts must stay inside configured storage
  roots.

### Tier 3: Selfhost Orchestrator

Runs locally or in CI.

Current entrypoint:

- `cmd/aepselfhost`

Target:

- outcome/status/watch/start-watch/verify are Go commands.
- PowerShell scripts are compatibility wrappers only.
- report generation and comparison orchestration should move from
  `scripts/verify_technique_selfhost.ps1` into Go.
- PowerShell remains usable while migrating but is not the architectural layer.

### Tier 4: AE Worker

Optional worker process.

Capabilities:

- AE availability probe
- run JSX
- render frames
- read back AE script results
- pixel compare integration

Platform behavior:

- Windows: adapter may use existing `scripts/ae_run.ps1` during migration.
- macOS: adapter should use macOS AE application paths and platform launch
  mechanics.
- Linux: adapter returns `Unavailable` for AE execution and render validation.

AE workers should be behind queues or explicit local commands. They should not
be directly exposed as arbitrary public code execution surfaces.

## Current Dependency Inventory

Known platform-sensitive areas:

- `cmd/aepselfhost/main.go`
  - `verifySelfhost` shells out to `pwsh`.
  - `processRunning` uses Windows `tasklist`.
  - `startWatch` starts the current executable directly and should remain
    portable after process inspection is split.
- `scripts/verify_technique_selfhost.ps1`
  - still owns the main selfhost gate engine.
  - calls `technique_showcase_report.ps1`, `compare_technique_reports.ps1`,
    and the PS compatibility wrappers.
- `cmd/aeoracle/main.go`
  - `render` shells out to `pwsh scripts/ae_run.ps1`.
  - usage text currently says `AfterFX.exe`, which is Windows-specific wording.
- `scripts/ae_run.ps1` and `scripts/AeRun.Lib.ps1`
  - Windows AE launch, dialog handling, OCR, crash-state handling.
- `internal/aep_test/*_shipgate_test.go`
  - many tests hard-code `E:/adobe/Adobe After Effects .../AfterFX.exe`.
  - tests dispatch AE through `scripts/ae_run.ps1`.
- `scripts/dump_effects_dict.ps1`, `scripts/run_ship_gates.ps1`,
  `scripts/regen_fixtures.ps1`
  - Windows AE paths and PowerShell orchestration.

This inventory is not a failure condition for Tier 1. Ship gates and AE workers
are allowed to stay platform-specific until their adapters exist.

## Target Interfaces

The first implementation slice should create small interfaces under
`internal/host`:

```go
package host

import (
	"context"
	"io"
)

type Command struct {
	Name    string
	Args    []string
	Dir     string
	Env     []string
	Stdout  io.Writer
	Stderr  io.Writer
}

type Result struct {
	ExitCode int
}

type Runner interface {
	Run(ctx context.Context, cmd Command) Result
	Start(ctx context.Context, cmd Command) (Process, error)
}

type Process interface {
	PID() int
	Release() error
}

type ProcessInspector interface {
	IsRunning(pid int) bool
}

type BrowserOpener interface {
	Open(ctx context.Context, path string) error
}

type Platform struct {
	Runner           Runner
	ProcessInspector ProcessInspector
	BrowserOpener    BrowserOpener
}
```

AE execution should be a separate package, tentatively `internal/aehost`, so
ordinary host process execution does not inherit AE-specific concepts:

```go
package aehost

import "context"

type CapabilityStatus string

const (
	CapabilityAvailable   CapabilityStatus = "available"
	CapabilityUnavailable CapabilityStatus = "unavailable"
)

type Availability struct {
	Status  CapabilityStatus `json:"status"`
	Reason  string           `json:"reason,omitempty"`
	Version string           `json:"version,omitempty"`
	Path    string           `json:"path,omitempty"`
}

type ScriptRequest struct {
	AEPath     string
	JSXPath    string
	DonePath   string
	TimeoutSec int
	Env        map[string]string
}

type ScriptResult struct {
	ExitCode int
	DonePath string
}

type Host interface {
	Available(ctx context.Context) Availability
	RunScript(ctx context.Context, req ScriptRequest) (ScriptResult, error)
}
```

## Implementation Plan

### Task 1: Add Host Interfaces And Process Inspection

**Files:**

- Create: `internal/host/host.go`
- Create: `internal/host/exec.go`
- Create: `internal/host/process_windows.go`
- Create: `internal/host/process_unix.go`
- Test: `internal/host/host_test.go`

- [ ] **Step 1: Write process inspector tests**

Create `internal/host/host_test.go`:

```go
package host

import (
	"context"
	"os"
	"testing"
)

func TestDefaultPlatformCanInspectCurrentProcess(t *testing.T) {
	platform := DefaultPlatform()
	if !platform.ProcessInspector.IsRunning(os.Getpid()) {
		t.Fatalf("current process pid %d should be running", os.Getpid())
	}
}

func TestRunnerReportsCommandExitCode(t *testing.T) {
	platform := DefaultPlatform()
	result := platform.Runner.Run(context.Background(), Command{
		Name:   os.Args[0],
		Args:   []string{"-test.run=TestHelperProcessExit3", "--"},
		Env:    append(os.Environ(), "GO_WANT_HELPER_PROCESS=1"),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if result.ExitCode != 3 {
		t.Fatalf("exit code = %d, want 3", result.ExitCode)
	}
}

func TestHelperProcessExit3(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(3)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/host -count=1
```

Expected: fail because `internal/host` does not exist.

- [ ] **Step 3: Add host interfaces and exec runner**

Create `internal/host/host.go` and `internal/host/exec.go` with the interfaces
from the Target Interfaces section. `DefaultPlatform()` should return an exec
runner, the platform process inspector, and a no-op browser opener that returns
a clear error until browser opening is needed.

- [ ] **Step 4: Add process inspectors**

Create `internal/host/process_windows.go`:

```go
//go:build windows

package host

import (
	"os/exec"
	"strconv"
	"strings"
)

type processInspector struct{}

func NewProcessInspector() ProcessInspector { return processInspector{} }

func (processInspector) IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), strconv.Itoa(pid))
}
```

Create `internal/host/process_unix.go`:

```go
//go:build !windows

package host

import (
	"os"
	"syscall"
)

type processInspector struct{}

func NewProcessInspector() ProcessInspector { return processInspector{} }

func (processInspector) IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
```

- [ ] **Step 5: Verify and commit**

Run:

```powershell
go test ./internal/host -count=1
go test ./cmd/aepselfhost -count=1
git diff --check
```

Commit:

```powershell
git add internal/host
git commit -m "feat: add cross-platform host primitives"
```

### Task 2: Move `aepselfhost` Process And Command Execution Behind Host

**Files:**

- Modify: `cmd/aepselfhost/main.go`
- Modify: `cmd/aepselfhost/main_test.go`
- Test: `cmd/aepselfhost/main_test.go`

- [ ] **Step 1: Refactor `run` to receive a platform**

Change the CLI wiring so `main()` passes `host.DefaultPlatform()` and tests can
pass a fake platform:

```go
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, host.DefaultPlatform()))
}

func run(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	// existing subcommand dispatch
}
```

Update all tests in `cmd/aepselfhost/main_test.go` to call:

```go
code := run(args, &stdout, &stderr, testPlatform())
```

- [ ] **Step 2: Add fake platform in tests**

Add a local fake runner and inspector so tests do not call real `pwsh` or
`tasklist`:

```go
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

type fakeProcess struct{ pid int }

func (p fakeProcess) PID() int       { return p.pid }
func (p fakeProcess) Release() error { return nil }

type fakeInspector struct{}

func (fakeInspector) IsRunning(pid int) bool { return pid == os.Getpid() }
```

- [ ] **Step 3: Replace direct process checks**

Change `processRunning` to use `platform.ProcessInspector`. Remove direct
`runtime.GOOS`, `tasklist`, and `os.FindProcess` logic from `cmd/aepselfhost`.

- [ ] **Step 4: Replace direct command execution**

Change `runCommand`, `verifySelfhost`, and `startWatch` to use
`platform.Runner.Run` / `platform.Runner.Start`.

- [ ] **Step 5: Verify and commit**

Run:

```powershell
go test ./cmd/aepselfhost ./internal/host -count=1
go run ./cmd/aepselfhost verify -out-root tmp\technique_selfhost_gate -limit 1 -dry-run
go run ./cmd/aepselfhost status -out-root tmp\technique_selfhost_gate
git diff --check
```

Commit:

```powershell
git add cmd/aepselfhost internal/host
git commit -m "refactor: route selfhost host operations through platform"
```

### Task 3: Make Selfhost Verification Go-Owned

**Files:**

- Modify: `cmd/aepselfhost/main.go`
- Modify: `cmd/aepselfhost/main_test.go`
- Create or modify: `internal/selfhost/*`
- Keep as wrapper: `scripts/verify_technique_selfhost.ps1`

- [ ] **Step 1: Extract a Go selfhost package**

Create `internal/selfhost` with explicit operations for:

- running full and partial technique reports
- comparing reports
- running recipe draft smoke checks
- writing acceptance/effectiveness/outcome artifacts
- appending history

Do not port the entire PowerShell file as one large Go file. Split by output:

- `gate.go`
- `reports.go`
- `outcome.go`
- `history.go`
- `html.go`

- [ ] **Step 2: Add a dry-run contract test**

Add a test that verifies `aepselfhost verify -dry-run` reports the Go plan and
does not mention `pwsh`:

```go
func TestRunVerifyDryRunUsesGoGate(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"verify", "-out-root", t.TempDir(), "-dry-run"}, &stdout, &stderr, testPlatform())
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "pwsh") {
		t.Fatalf("dry-run should not depend on pwsh:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "DRY RUN technique selfhost verify") {
		t.Fatalf("missing dry-run marker:\n%s", stdout.String())
	}
}
```

- [ ] **Step 3: Port only artifact orchestration first**

Keep existing Go CLIs for report generation where available. If a report still
only exists as a PS script, call the eventual Go equivalent from the new
package only after adding that equivalent. Do not call PS from the new
`internal/selfhost` package.

- [ ] **Step 4: Convert `scripts/verify_technique_selfhost.ps1` to wrapper**

After the Go gate produces the same `latest_outcome.*`,
`latest_effectiveness.*`, `latest_index.html`, `history.jsonl`, and
`history.csv`, reduce the script to:

```powershell
param(
    [string]$OutRoot = "tmp\technique_selfhost_gate",
    [int]$Limit = 0,
    [switch]$Open
)

$argsList = @("run", "./cmd/aepselfhost", "verify", "-out-root", $OutRoot)
if ($Limit -gt 0) { $argsList += @("-limit", [string]$Limit) }
if ($Open) { $argsList += "-open" }

& go @argsList
exit $LASTEXITCODE
```

- [ ] **Step 5: Verify and commit**

Run:

```powershell
go test ./cmd/aepselfhost ./internal/selfhost ./cmd/aeptechnique ./internal/technique -count=1
go run ./cmd/aepselfhost verify -out-root tmp\technique_selfhost_gate
pwsh -NoProfile -File scripts\verify_technique_selfhost.ps1 -OutRoot tmp\technique_selfhost_gate
git diff --check
```

Commit:

```powershell
git add cmd/aepselfhost internal/selfhost scripts/verify_technique_selfhost.ps1
git commit -m "feat: move selfhost verification gate to Go"
```

### Task 4: Define AE Host Capability

**Files:**

- Create: `internal/aehost/aehost.go`
- Create: `internal/aehost/unavailable.go`
- Create: `internal/aehost/windows.go`
- Create: `internal/aehost/darwin.go`
- Modify: `cmd/aeoracle/main.go`
- Test: `internal/aehost/aehost_test.go`

- [ ] **Step 1: Add AE host interfaces**

Use the `internal/aehost` interfaces from the Target Interfaces section.

- [ ] **Step 2: Add unavailable host behavior**

Linux and unknown configurations must return:

```json
{"status":"unavailable","reason":"after effects automation is not configured on this platform"}
```

- [ ] **Step 3: Route `aeoracle render` through `aehost.Host`**

Replace direct `exec.Command("pwsh", ...)` in `cmd/aeoracle/main.go` with an
AE host call. Keep the Windows adapter allowed to call the existing PS runner
as an internal implementation detail during migration.

- [ ] **Step 4: Update user-facing wording**

Change CLI usage from `AfterFX.exe path` to `After Effects executable path`.

- [ ] **Step 5: Verify and commit**

Run:

```powershell
go test ./internal/aehost ./cmd/aeoracle -count=1
go run ./cmd/aeoracle render -request tmp\nonexistent.json -dry-run
git diff --check
```

Commit:

```powershell
git add internal/aehost cmd/aeoracle
git commit -m "refactor: route AE rendering through host adapter"
```

### Task 5: Add Cross-Platform Build Gate

**Files:**

- Create: `scripts/verify_cross_platform.ps1`
- Modify: `README.md`

- [ ] **Step 1: Add build script**

Create `scripts/verify_cross_platform.ps1`:

```powershell
$ErrorActionPreference = "Stop"

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64" },
    @{ GOOS = "darwin"; GOARCH = "arm64" },
    @{ GOOS = "linux"; GOARCH = "amd64" }
)

$packages = @(
    "./cmd/aepsearch",
    "./cmd/aeprecipe",
    "./cmd/aeptechnique",
    "./cmd/aepselfhost",
    "./internal/host"
)

foreach ($target in $targets) {
    Write-Host "building GOOS=$($target.GOOS) GOARCH=$($target.GOARCH)"
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    go build $packages
}

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
```

- [ ] **Step 2: Add README command**

Add:

```powershell
pwsh -NoProfile -File scripts\verify_cross_platform.ps1
```

- [ ] **Step 3: Verify and commit**

Run:

```powershell
pwsh -NoProfile -File scripts\verify_cross_platform.ps1
go test ./cmd/aepselfhost ./internal/host -count=1
git diff --check
```

Commit:

```powershell
git add scripts/verify_cross_platform.ps1 README.md
git commit -m "test: add cross-platform build gate"
```

## Service Direction

After Tasks 1-5, create a service package only over Tier 1 features.

Recommended first service slice:

- `cmd/aepserver`
- `internal/server`
- endpoints:
  - `GET /health`
  - `GET /capabilities`
  - `POST /parse`
  - `POST /profile`

Capability response should include:

```json
{
  "parser": "available",
  "profile": "available",
  "recipe_draft": "available",
  "ae_readback": "unavailable",
  "render": "unavailable",
  "pixel_diff": "available"
}
```

`pixel_diff` can be available without AE because PNG comparison is pure Go.
`render` is unavailable unless an AE worker is configured.

## Acceptance Criteria

- `go test ./...` continues to pass on Windows.
- `GOOS=darwin GOARCH=arm64 go test` build checks pass for Tier 1 packages.
- `GOOS=linux GOARCH=amd64 go test` build checks pass for Tier 1 packages.
- `cmd/aepselfhost` has no direct `runtime.GOOS == "windows"` branch.
- `cmd/aepselfhost` has no direct `exec.Command("tasklist", ...)`.
- `cmd/aepselfhost verify -dry-run` does not mention `pwsh` after Task 3.
- `cmd/aeoracle render` reports AE unavailable on Linux instead of trying to
  run `pwsh`.
- PowerShell scripts that remain are either compatibility wrappers or explicit
  Windows AE adapter internals.

## Non-Goals

- Do not make Linux run AE validation.
- Do not expose AE script execution directly over the public parser service.
- Do not rewrite all ship gate tests in the first cross-platform slice.
- Do not remove Windows AE automation before the Windows adapter preserves the
  current gate behavior.
- Do not create a service that requires AE at startup.

## Progress

- [x] Direction accepted: cross-platform core/service, optional AE worker.
- [x] Initial platform dependency inventory recorded.
- [x] Add host primitives.
- [x] Route `aepselfhost` through host primitives.
- [x] Move report compare and recipe draft smoke orchestration into Go.
- [ ] Move selfhost verification gate from PowerShell to Go.
- [ ] Define AE host adapter.
- [ ] Add cross-platform build gate.
- [ ] Start pure parser service.
