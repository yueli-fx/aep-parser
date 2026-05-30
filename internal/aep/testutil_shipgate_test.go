package aep_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
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
