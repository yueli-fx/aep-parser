package aepmigrate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aehost"
)

func writeMinimalMatrixRecipe(t *testing.T, path string) {
	t.Helper()
	data := []byte(`{
  "schema_version": 1,
  "project": {
    "name": "Matrix fixture",
    "target_version": "AE2020"
  },
  "comps": [
    {
      "name": "Main",
      "width": 640,
      "height": 360,
      "frame_rate": 24,
      "duration": 2.5
    }
  ]
}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

type fakeMatrixHost struct {
	called   bool
	request  aehost.ScriptRequest
	requests []aehost.ScriptRequest
	doneBody string
}

func (h *fakeMatrixHost) Available(context.Context) aehost.Availability {
	return aehost.Availability{Status: aehost.CapabilityAvailable}
}

func (h *fakeMatrixHost) RunScript(_ context.Context, req aehost.ScriptRequest) (aehost.ScriptResult, error) {
	h.called = true
	h.request = req
	h.requests = append(h.requests, req)
	if err := os.MkdirAll(filepath.Dir(req.DonePath), 0o755); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	if err := os.WriteFile(req.DonePath, []byte(h.doneBody), 0o644); err != nil {
		return aehost.ScriptResult{ExitCode: 1, DonePath: req.DonePath}, err
	}
	return aehost.ScriptResult{ExitCode: 0, DonePath: req.DonePath}, nil
}

func assertMatrixCaseStatus(t *testing.T, cases []MatrixCase, target string, want MatrixStatus) {
	t.Helper()
	for _, c := range cases {
		if c.TargetVersion == target {
			if c.Status != want {
				t.Fatalf("%s status = %s, want %s; case=%+v", target, c.Status, want, c)
			}
			return
		}
	}
	t.Fatalf("case for target %s not found in %+v", target, cases)
}
