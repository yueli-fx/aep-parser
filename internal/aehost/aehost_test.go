package aehost

import (
	"context"
	"slices"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/host"
)

func TestUnavailableHostReportsUnavailable(t *testing.T) {
	h := NewUnavailableHost("")

	got := h.Available(context.Background())
	if got.Status != CapabilityUnavailable {
		t.Fatalf("Status = %q, want %q", got.Status, CapabilityUnavailable)
	}
	if got.Reason != DefaultUnavailableReason {
		t.Fatalf("Reason = %q", got.Reason)
	}
}

func TestUnavailableHostRunScriptFailsDeterministically(t *testing.T) {
	h := NewUnavailableHost("")

	got, err := h.RunScript(context.Background(), ScriptRequest{DonePath: "render.done"})
	if err == nil {
		t.Fatal("RunScript error = nil, want unavailable error")
	}
	if got.ExitCode != 1 || got.DonePath != "render.done" {
		t.Fatalf("RunScript result = %+v", got)
	}
}

func TestPowerShellHostRunScriptBuildsAERunnerCommand(t *testing.T) {
	runner := &recordingRunner{}
	h := NewPowerShellHost(runner, "scripts/ae-worker/ae_run.ps1")

	got, err := h.RunScript(context.Background(), ScriptRequest{
		AEPath:     "AfterFX.exe",
		JSXPath:    "render.jsx",
		DonePath:   "render.done",
		TimeoutSec: 42,
		WorkDir:    "work",
		Env:        map[string]string{"AEORACLE_REQUEST": "request.json"},
	})
	if err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	if got.ExitCode != 0 || got.DonePath != "render.done" {
		t.Fatalf("RunScript result = %+v", got)
	}
	if runner.cmd.Name != "pwsh" {
		t.Fatalf("command name = %q", runner.cmd.Name)
	}
	wantArgs := []string{"-NoProfile", "-File", "scripts/ae-worker/ae_run.ps1", "-AeExe", "AfterFX.exe", "-Jsx", "render.jsx", "-Done", "render.done", "-TimeoutSec", "42"}
	if !slices.Equal(runner.cmd.Args, wantArgs) {
		t.Fatalf("command args = %#v", runner.cmd.Args)
	}
	if runner.cmd.Dir != "work" {
		t.Fatalf("command dir = %q", runner.cmd.Dir)
	}
	if !slices.Contains(runner.cmd.Env, "AEORACLE_REQUEST=request.json") {
		t.Fatalf("command env missing AEORACLE_REQUEST: %#v", runner.cmd.Env)
	}
}

type recordingRunner struct {
	cmd host.Command
}

func (r *recordingRunner) Run(_ context.Context, cmd host.Command) host.Result {
	r.cmd = cmd
	return host.Result{ExitCode: 0}
}

func (r *recordingRunner) Start(context.Context, host.Command) (host.Process, error) {
	panic("Start should not be called")
}
