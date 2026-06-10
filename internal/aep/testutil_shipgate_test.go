package aep_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/example/aep-parser/internal/rifx"
)

// runAeRunShipGate dispatches AE via scripts/ae_run.ps1 (drop-in for AfterFX -r).
// ps1 owns timeout + .done polling + dialog dismissal. On non-zero exit ps1
// leaves a <doneFile>.fail/ dump dir with screenshot.png / ocr.txt / actions.log.
//
// Exit 1 (timeout) / 2 (unknown modal) get ONE automatic warm retry: AE
// cold-start (splash outliving the unknown-modal grace) manifests as exactly
// these codes and self-heals on relaunch, while a deterministic data reject
// fails the retry identically — so flakes vanish without masking real rejects.
// The first attempt's forensics dump is preserved as <doneFile>.fail.1/.
func runAeRunShipGate(t *testing.T, aeExe, jsxPath, doneFile string, timeoutSec int) {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	script := filepath.Join(repoRoot, "scripts", "ae_run.ps1")

	// Pull the gate's moving parts into the go test cache key. The cache
	// tracks files a test reads — without these reads, editing the verify JSX
	// (or a dialog rule) leaves a stale cached PASS on display (see
	// incidents/camera-light-layer-create-re.md gotcha). Reading them here
	// makes such edits invalidate the cache structurally instead of relying
	// on -count=1 discipline.
	for _, p := range []string{jsxPath, script, filepath.Join(repoRoot, "scripts", "ae_dialog_rules.json")} {
		if _, err := os.ReadFile(p); err != nil {
			t.Fatalf("gate input missing: %v", err)
		}
	}

	run := func() error {
		cmd := exec.Command("pwsh", "-NoProfile", "-File", script,
			"-AeExe", aeExe,
			"-Jsx", jsxPath,
			"-Done", doneFile,
			"-TimeoutSec", strconv.Itoa(timeoutSec))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	err := run()
	if err == nil {
		return
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if code := exitErr.ExitCode(); code == 1 || code == 2 {
			t.Logf("ae_run.ps1 exit %d — warm retry once (cold-start flake heals; a real reject fails again)", code)
			failDir := doneFile + ".fail"
			os.RemoveAll(failDir + ".1")
			os.Rename(failDir, failDir+".1")
			if err = run(); err == nil {
				return
			}
		}
	}
	t.Fatalf("ae_run.ps1 failed: %v (check %s.fail/ for dump; first attempt in %s.fail.1/ if retried)", err, doneFile, doneFile)
}

func trimShipNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}

func findShipChunk(root *rifx.Chunk, id rifx.ChunkID) *rifx.Chunk {
	for _, ch := range root.Children {
		if ch.ID == id {
			return ch
		}
		if ch.IsList() {
			if g := findShipChunk(ch, id); g != nil {
				return g
			}
		}
	}
	return nil
}

// findShipListByForm returns the first descendant LIST chunk whose FormType
// matches form (depth-first), or nil. Use this for container chunks like om-s
// that are LISTs (ID == "LIST", FormType == "om-s"), not leaf chunk IDs.
func findShipListByForm(root *rifx.Chunk, form rifx.ChunkID) *rifx.Chunk {
	for _, ch := range root.Children {
		if ch.IsList() {
			if ch.FormType == form {
				return ch
			}
			if g := findShipListByForm(ch, form); g != nil {
				return g
			}
		}
	}
	return nil
}

// findShipList finds the tdmn matching name, then returns the LIST(list/kfl)
// keyframe container inside the following LIST(tdbs), or nil (static stream).
func findShipList(root *rifx.Chunk, name string) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for i := 0; i+1 < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && trimShipNUL(string(ch.Data)) == name {
				tdbs := c.Children[i+1]
				if tdbs.IsList() && tdbs.FormType == rifx.IDTdbs {
					for _, cc := range tdbs.Children {
						if cc.IsList() && cc.FormType == rifx.IDkfl {
							found = cc
							return
						}
					}
				}
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func parseAEP(t *testing.T, path string) *rifx.Chunk {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return root
}

// streamCdat finds the tdmn matching name anywhere in the tree and returns the
// cdat inside the following LIST(tdbs), or nil.
func streamCdat(root *rifx.Chunk, name string) []byte {
	var found []byte
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && i+1 < len(c.Children) && trimShipNUL(string(ch.Data)) == name {
				tdbs := c.Children[i+1]
				if tdbs.IsList() && tdbs.FormType == rifx.IDTdbs {
					for _, cc := range tdbs.Children {
						if cc.ID == rifx.IDCdat {
							found = cc.Data
							return
						}
					}
				}
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}
