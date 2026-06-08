// internal/aep/render_queue_remove_shipgate_test.go
//
// AE ship gate for RenderQueue.RemoveItem. Opens an AE-2020-native 2-item
// render-queue fixture, removes item index 1 via RemoveItem, WriteAEP, and has
// AE open the result: proves AE ACCEPTS the restructured LRdr (settings ldat +
// lhd3 count + Rout per-item block + LItm groups all shrunk in lock-step; no
// data-loss / corrupt) and reads back exactly one item with the right comp.
// Resaves so the Go side confirms the deletion stuck (NumItems == 1).
//
// Gated by AE_SHIP_GATE. Fixture built by test_data/re_rq_delete.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runRQRemoveShipGate(t *testing.T, aeExe, baseFixture string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/rq_delete_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_rq_delete.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	proj, err := aep.Open(baseFixture)
	if err != nil {
		t.Skipf("%s not present: %v", baseFixture, err)
	}
	rq := proj.RenderQueue
	if rq == nil || rq.NumItems() != 2 {
		t.Fatalf("%s: want 2 render queue items, got %d", baseFixture, rq.NumItems())
	}
	survivor := ""
	if c := rq.Items[0].Comp; c != nil {
		survivor = c.Name
	}
	if err := aep.RemoveItem(rq, 1); err != nil {
		t.Fatalf("RemoveItem(1): %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "rq_del_in.aep")
	resavedAEP := filepath.Join(tempDir, "rq_del_resaved.aep")
	doneFile := filepath.Join(tempDir, "rq_del.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expectItems":1,"survivorComp":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), survivor)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("RemoveItem AE readback:\n%s", body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("RQ remove ship gate FAIL:\n%s", body)
	}

	// Preservation: AE's resave still has exactly one item.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	if re.RenderQueue == nil || re.RenderQueue.NumItems() != 1 {
		t.Fatalf("resaved: NumItems = %d, want 1", re.RenderQueue.NumItems())
	}
}

func TestRQRemove_AEShipGate_AE2020(t *testing.T) {
	runRQRemoveShipGate(t, ae2020(), "../../test_data/re_rq_delete_before.aep")
}
func TestRQRemove_AEShipGate_AE2025(t *testing.T) {
	runRQRemoveShipGate(t, ae2025(), "../../test_data/re_rq_delete_before.aep")
}

// runRQAddShipGate clones the queue's lone item for comp "RQB" via AddItem and
// verifies AE accepts the appended item + reads RQB as item 2.
func runRQAddShipGate(t *testing.T, aeExe, baseFixture string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/rq_add_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_rq_add.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	proj, err := aep.Open(baseFixture)
	if err != nil {
		t.Skipf("%s not present: %v", baseFixture, err)
	}
	rq := proj.RenderQueue
	if rq == nil || rq.NumItems() != 1 {
		t.Fatalf("%s: want 1 render queue item, got %d", baseFixture, rq.NumItems())
	}
	var rqb *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RQB" {
			rqb = c
		}
	}
	if rqb == nil {
		t.Fatalf("%s: comp RQB not found", baseFixture)
	}
	if _, err := aep.AddItem(rq, rqb); err != nil {
		t.Fatalf("AddItem(RQB): %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "rq_add_in.aep")
	resavedAEP := filepath.Join(tempDir, "rq_add_resaved.aep")
	doneFile := filepath.Join(tempDir, "rq_add.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expectItems":2,"addedComp":"RQB"}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("AddItem AE readback:\n%s", body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("RQ add ship gate FAIL:\n%s", body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	if re.RenderQueue == nil || re.RenderQueue.NumItems() != 2 {
		t.Fatalf("resaved: NumItems = %d, want 2", re.RenderQueue.NumItems())
	}
}

func TestRQAdd_AEShipGate_AE2020(t *testing.T) {
	runRQAddShipGate(t, ae2020(), "../../test_data/re_rq_add_before.aep")
}
func TestRQAdd_AEShipGate_AE2025(t *testing.T) {
	runRQAddShipGate(t, ae2025(), "../../test_data/re_rq_add_before.aep")
}
