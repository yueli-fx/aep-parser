package aepmigrate

import (
	"path/filepath"
	"testing"
)

func TestVerifyRunsAEOpenGateWhenConfigured(t *testing.T) {
	source := writeTempVerifyProject(t, "Main")
	donePath := filepath.Join(t.TempDir(), "ae-open.done")
	host := &fakeAEOpenHost{doneBody: "PASS\nproject items.length=1\ncomp=Main layers.length=0\n"}

	report, err := Verify(VerifyOptions{
		SourcePath:    source,
		TargetPath:    source,
		TargetVersion: VersionAE2020,
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
		t.Fatalf("Verify: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Verification.AEOpenStatus != "pass" {
		t.Fatalf("AE open status = %q, verification=%+v", report.Verification.AEOpenStatus, report.Verification)
	}
	if !host.called {
		t.Fatal("fake AE host was not called")
	}
	if host.request.AEPath != "AfterFX.exe" || host.request.JSXPath != "verify_open.jsx" || host.request.DonePath != donePath || host.request.TimeoutSec != 12 {
		t.Fatalf("AE open request = %+v", host.request)
	}
}
