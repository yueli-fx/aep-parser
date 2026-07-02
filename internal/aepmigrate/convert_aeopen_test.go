package aepmigrate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aehost"
	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestConvertRunsAEOpenGateWhenConfigured(t *testing.T) {
	source := writeTempProjectWithOneComp(t, aep.TargetAE2020)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	donePath := filepath.Join(t.TempDir(), "ae-open.done")
	host := &fakeAEOpenHost{doneBody: "PASS\nproject items.length=1\ncomp=Main layers.length=0\n"}

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2025,
		AEOpen: &AEOpenOptions{
			Host:       host,
			AEPath:     "AfterFX.exe",
			JSXPath:    "verify_open.jsx",
			ArgsPath:   filepath.Join(t.TempDir(), "verify_open_args.json"),
			DonePath:   donePath,
			TimeoutSec: 12,
		},
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.AEOpenStatus != "pass" {
		t.Fatalf("AE open status = %q, verification=%+v", report.Verification.AEOpenStatus, report.Verification)
	}
	if report.Verification.AEOpenLog == "" {
		t.Fatalf("AEOpenLog empty, verification=%+v", report.Verification)
	}
	if report.Verification.AEOpenExitCode == nil || *report.Verification.AEOpenExitCode != 0 {
		t.Fatalf("AEOpenExitCode = %v, want 0", report.Verification.AEOpenExitCode)
	}
	if !host.called {
		t.Fatal("fake AE host was not called")
	}
	if host.request.AEPath != "AfterFX.exe" || host.request.JSXPath != "verify_open.jsx" || host.request.DonePath != donePath || host.request.TimeoutSec != 12 {
		t.Fatalf("host request = %+v", host.request)
	}
}

func TestConvertBlocksWhenAEOpenGateFails(t *testing.T) {
	source := writeTempProjectWithOneComp(t, aep.TargetAE2020)
	outPath := filepath.Join(t.TempDir(), "converted.aep")
	host := &fakeAEOpenHost{doneBody: "FAIL\nERROR: cannot open\n", exitCode: 1}

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2025,
		AEOpen: &AEOpenOptions{
			Host:       host,
			AEPath:     "AfterFX.exe",
			JSXPath:    "verify_open.jsx",
			ArgsPath:   filepath.Join(t.TempDir(), "verify_open_args.json"),
			DonePath:   filepath.Join(t.TempDir(), "ae-open.done"),
			TimeoutSec: 12,
		},
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Summary.Status != StatusBlocked {
		t.Fatalf("status = %q, want blocked; entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.AEOpenStatus != "fail" {
		t.Fatalf("AE open status = %q, verification=%+v", report.Verification.AEOpenStatus, report.Verification)
	}
}

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
