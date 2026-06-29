// internal/aep/structural_ops_shipgate_test.go
//
// Automated AE ship gate for the layer-list structural ops (DeleteLayer /
// DuplicateLayer / InsertLayer / MoveLayer + the 4 MoveTo*/Move* wrappers).
// These were previously only MANUALLY JSX-gated (recorded in coverage.md) with
// no automated Go _AEShipGate test, so the caps were capped at verify=roundtrip.
// This gate drives every op from scratch and verifies AE's accepted layer order
// (the ops' surface) via DOM readback, one isolated comp per op so a single bad
// op never masks the rest.
//
//   - DEL  A,B,C → DeleteLayer(1)        → A,C
//   - DUP  A,B,C → DuplicateLayer(1,"Bd")→ A,Bd,B,C  (clone above source)
//   - INS  A,B   → InsertLayer(SRC.S, 1) → A,S,B      (same-project cross-comp)
//   - MOVL A,B,C → MoveLayer(0, 2)       → B,C,A
//   - MTOB A,B,C → MoveToBeginning(C)    → C,A,B
//   - MTOE A,B,C → MoveToEnd(A)          → B,C,A
//   - MAFT A,B,C → MoveAfter(A, C)       → B,C,A
//   - MBEF A,B,C → MoveBefore(C, A)      → C,A,B
//
// Layers are solids (AV) — the ops refuse non-AV. NewSolidLayer appends, so
// creation order == AE top-to-bottom order.
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildStructuralDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	mk := func(name string) *aep.Composition {
		c, err := aep.NewComposition(p, name, 640, 360, 30, 2)
		if err != nil {
			t.Fatalf("NewComposition %s: %v", name, err)
		}
		return c
	}
	// Solid layers (LayerTypeAV) — Delete/Duplicate/InsertLayer refuse non-AV
	// (shape/text/camera/light) layers, so the structural ops need AV carriers.
	addSolid := func(c *aep.Composition, name string) {
		if _, err := aep.NewSolidLayer(c, name, 100, 100, [3]float64{0.8, 0.6, 0.2}); err != nil {
			t.Fatalf("NewSolidLayer %s/%s: %v", c.Name, name, err)
		}
	}
	dots := func(c *aep.Composition, names ...string) {
		for _, n := range names {
			addSolid(c, n)
		}
	}

	src := mk("SRC")
	addSolid(src, "S")
	dots(mk("DEL"), "A", "B", "C")
	dots(mk("DUP"), "A", "B", "C")
	dots(mk("INS"), "A", "B")
	dots(mk("MOVL"), "A", "B", "C")
	dots(mk("MTOB"), "A", "B", "C")
	dots(mk("MTOE"), "A", "B", "C")
	dots(mk("MAFT"), "A", "B", "C")
	dots(mk("MBEF"), "A", "B", "C")

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}

	comp := func(name string) *aep.Composition {
		c := rp.CompositionByName(name)
		if c == nil {
			t.Fatalf("reopened comp %s missing", name)
		}
		return c
	}
	order := func(c *aep.Composition) []string {
		out := make([]string, len(c.Layers))
		for i, l := range c.Layers {
			out[i] = l.Name
		}
		return out
	}
	assertOrder := func(label string, c *aep.Composition, want ...string) {
		t.Helper()
		got := order(c)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("%s baseline order = %v, want %v", label, got, want)
		}
	}
	idxOf := func(c *aep.Composition, name string) int {
		for i, l := range c.Layers {
			if l.Name == name {
				return i
			}
		}
		t.Fatalf("layer %s not found in %s", name, c.Name)
		return -1
	}

	// DEL: A,B,C → DeleteLayer(B) → A,C
	del := comp("DEL")
	assertOrder("DEL", del, "A", "B", "C")
	if err := aep.DeleteLayer(del, idxOf(del, "B")); err != nil {
		t.Fatalf("DeleteLayer: %v", err)
	}

	// DUP: A,B,C → DuplicateLayer(B,"Bd") → A,Bd,B,C
	dup := comp("DUP")
	assertOrder("DUP", dup, "A", "B", "C")
	if _, err := aep.DuplicateLayer(dup, idxOf(dup, "B"), "Bd"); err != nil {
		t.Fatalf("DuplicateLayer: %v", err)
	}

	// INS: A,B + InsertLayer(SRC.S at 1) → A,S,B
	ins := comp("INS")
	assertOrder("INS", ins, "A", "B")
	srcLayer := comp("SRC").LayerByName("S")
	if srcLayer == nil {
		t.Fatal("SRC.S missing")
	}
	if _, err := aep.InsertLayer(ins, srcLayer, 1); err != nil {
		t.Fatalf("InsertLayer: %v", err)
	}

	// MOVL: A,B,C → MoveLayer(0,2) → B,C,A
	movl := comp("MOVL")
	assertOrder("MOVL", movl, "A", "B", "C")
	if err := aep.MoveLayer(movl, 0, 2); err != nil {
		t.Fatalf("MoveLayer: %v", err)
	}

	// MTOB: A,B,C → MoveToBeginning(C) → C,A,B
	mtob := comp("MTOB")
	assertOrder("MTOB", mtob, "A", "B", "C")
	if err := aep.MoveToBeginning(mtob.LayerByName("C")); err != nil {
		t.Fatalf("MoveToBeginning: %v", err)
	}

	// MTOE: A,B,C → MoveToEnd(A) → B,C,A
	mtoe := comp("MTOE")
	assertOrder("MTOE", mtoe, "A", "B", "C")
	if err := aep.MoveToEnd(mtoe.LayerByName("A")); err != nil {
		t.Fatalf("MoveToEnd: %v", err)
	}

	// MAFT: A,B,C → MoveAfter(A, C) → B,C,A
	maft := comp("MAFT")
	assertOrder("MAFT", maft, "A", "B", "C")
	if err := aep.MoveAfter(maft.LayerByName("A"), maft.LayerByName("C")); err != nil {
		t.Fatalf("MoveAfter: %v", err)
	}

	// MBEF: A,B,C → MoveBefore(C, A) → C,A,B
	mbef := comp("MBEF")
	assertOrder("MBEF", mbef, "A", "B", "C")
	if err := aep.MoveBefore(mbef.LayerByName("C"), mbef.LayerByName("A")); err != nil {
		t.Fatalf("MoveBefore: %v", err)
	}

	return rp
}

func runStructuralGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/structural_ops_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_structural_ops.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildStructuralDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "structural_ops_in.aep")
	resavedAEP := filepath.Join(tempDir, "structural_ops_resaved.aep")
	doneFile := filepath.Join(tempDir, "structural_ops.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("structural %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("structural %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: the mutated layer lists survive AE's re-encode and the Go
	// parser reads back the post-op order.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	check := func(name string, want ...string) {
		c := re.CompositionByName(name)
		if c == nil {
			t.Errorf("resaved comp %s missing", name)
			return
		}
		got := make([]string, len(c.Layers))
		for i, l := range c.Layers {
			got[i] = l.Name
		}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("resaved %s order = %v, want %v", name, got, want)
		}
	}
	check("DEL", "A", "C")
	check("DUP", "A", "Bd", "B", "C")
	check("INS", "A", "S", "B")
	check("MOVL", "B", "C", "A")
	check("MTOB", "C", "A", "B")
	check("MTOE", "B", "C", "A")
	check("MAFT", "B", "C", "A")
	check("MBEF", "C", "A", "B")
}

func TestStructuralOps_AEShipGate_AE2020(t *testing.T) {
	runStructuralGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestStructuralOps_AEShipGate_AE2025(t *testing.T) {
	runStructuralGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
