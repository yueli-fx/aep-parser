package aep_test

import (
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
func runAeRunShipGate(t *testing.T, aeExe, jsxPath, doneFile string, timeoutSec int) {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	script := filepath.Join(repoRoot, "scripts", "ae_run.ps1")

	cmd := exec.Command("pwsh", "-NoProfile", "-File", script,
		"-AeExe", aeExe,
		"-Jsx", jsxPath,
		"-Done", doneFile,
		"-TimeoutSec", strconv.Itoa(timeoutSec))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("ae_run.ps1 failed: %v (check %s.fail/ for dump)", err, doneFile)
	}
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
