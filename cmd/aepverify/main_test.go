package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/host"
)

func TestRunCrossPlatformUsesGoBuildTargets(t *testing.T) {
	var stdout, stderr bytes.Buffer
	runner := &recordingRunner{}

	code := runWithIO([]string{"cross-platform"}, &stdout, &stderr, runner)

	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
	}
	if len(runner.commands) != 3 {
		t.Fatalf("commands = %d, want 3", len(runner.commands))
	}
	if !slices.Contains(runner.commands[0].Env, "GOOS=windows") {
		t.Fatalf("first env = %#v", runner.commands[0].Env)
	}
	if !slices.Contains(runner.commands[1].Env, "GOOS=darwin") {
		t.Fatalf("second env = %#v", runner.commands[1].Env)
	}
	if !slices.Contains(runner.commands[2].Env, "GOOS=linux") {
		t.Fatalf("third env = %#v", runner.commands[2].Env)
	}
	if !slices.Contains(runner.commands[0].Args, "./cmd/aepserver") || !slices.Contains(runner.commands[0].Args, "./internal/server") {
		t.Fatalf("cross-platform packages missing server entries: %#v", runner.commands[0].Args)
	}
}

func TestRunRecipeProfilesEmitsJSONSummary(t *testing.T) {
	var stdout, stderr bytes.Buffer
	recipePath := filepath.Join("..", "..", "examples", "recipes", "minimal-comp-object-profile.json")
	outDir := t.TempDir()

	code := runWithIO([]string{"recipe-profiles", "-recipe", recipePath, "-out", outDir, "-json"}, &stdout, &stderr, host.ExecRunner{})

	if code != 0 {
		t.Fatalf("exit = %d, stderr = %s\nstdout=%s", code, stderr.String(), stdout.String())
	}
	var got RecipeProfileSummary
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal %s: %v", stdout.String(), err)
	}
	if got.Total != 1 || got.Passed != 1 || got.Failed != 0 {
		t.Fatalf("summary = %+v", got)
	}
	if !slices.Contains(got.CoveredProfilePaths, "expected_profile.background_color") {
		t.Fatalf("covered paths = %#v", got.CoveredProfilePaths)
	}
	if got.Recipes[0].OutputPath == "" {
		t.Fatalf("output path missing in %+v", got.Recipes[0])
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runWithIO([]string{"nope"}, &stdout, &stderr, host.ExecRunner{})

	if code != 2 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(stderr.String(), "usage: aepverify") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

type recordingRunner struct {
	commands []host.Command
}

func (r *recordingRunner) Run(_ context.Context, cmd host.Command) host.Result {
	r.commands = append(r.commands, cmd)
	return host.Result{ExitCode: 0}
}

func (r *recordingRunner) Start(context.Context, host.Command) (host.Process, error) {
	panic("Start should not be called")
}
