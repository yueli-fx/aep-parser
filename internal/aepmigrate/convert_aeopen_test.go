package aepmigrate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aehost"
)

func TestWriteAEOpenArgsUsesAbsolutePaths(t *testing.T) {
	argsPath := filepath.Join(t.TempDir(), "verify_open_args.json")
	if err := writeAEOpenArgs(argsPath, "relative/converted.aep", "relative/ae_open.done"); err != nil {
		t.Fatalf("writeAEOpenArgs: %v", err)
	}
	data, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var got struct {
		Input string `json:"input"`
		Done  string `json:"done"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !filepath.IsAbs(filepath.FromSlash(got.Input)) || !filepath.IsAbs(filepath.FromSlash(got.Done)) {
		t.Fatalf("args paths must be absolute, got input=%q done=%q", got.Input, got.Done)
	}
}

type fakeAEOpenHost struct {
	called   bool
	request  aehost.ScriptRequest
	doneBody string
	exitCode int
}

func (h *fakeAEOpenHost) Available(context.Context) aehost.Availability {
	return aehost.Availability{Status: aehost.CapabilityAvailable}
}

func (h *fakeAEOpenHost) RunScript(_ context.Context, req aehost.ScriptRequest) (aehost.ScriptResult, error) {
	h.called = true
	h.request = req
	if err := os.MkdirAll(filepath.Dir(req.DonePath), 0o755); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	if err := os.WriteFile(req.DonePath, []byte(h.doneBody), 0o644); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	return aehost.ScriptResult{ExitCode: h.exitCode, DonePath: req.DonePath}, nil
}
