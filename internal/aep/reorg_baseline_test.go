package aep_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/example/aep-parser/internal/aep"
)

type reorgEntry struct {
	OpenOK bool   `json:"open_ok"`
	SHA    string `json:"sha,omitempty"`
	Err    string `json:"err,omitempty"`
}

// TestReorgRoundtripBaseline asserts each fixture's Open->WriteAEP output is
// byte-stable across the package reorg. The baseline is produced by the
// tmp_debug/reorg_baseline generator; when it is absent the test skips, so it
// only guards during the reorg window.
func TestReorgRoundtripBaseline(t *testing.T) {
	const baselinePath = "../../tmp_debug/reorg_baseline/baseline.json"
	raw, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Skipf("no reorg baseline (%v) — generate via `go run ./tmp_debug/reorg_baseline`", err)
	}
	var want map[string]reorgEntry
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("unmarshal baseline: %v", err)
	}
	for path, exp := range want {
		path, exp := path, exp
		t.Run(path, func(t *testing.T) {
			proj, oerr := aep.Open("../../" + path)
			if !exp.OpenOK {
				if oerr == nil {
					t.Fatalf("expected Open error %q, got nil", exp.Err)
				}
				return
			}
			if oerr != nil {
				t.Fatalf("Open: %v (baseline had OpenOK)", oerr)
			}
			var buf bytes.Buffer
			if werr := proj.WriteAEP(&buf); werr != nil {
				t.Fatalf("WriteAEP: %v", werr)
			}
			sum := sha256.Sum256(buf.Bytes())
			if got := hex.EncodeToString(sum[:]); got != exp.SHA {
				t.Fatalf("WriteAEP byte mismatch: baseline %s got %s", exp.SHA, got)
			}
		})
	}
}
