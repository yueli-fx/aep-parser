// internal/aep/comp_duplicate_shipgate_test.go
//
// Automated AE ship gate for DuplicateComposition. The cap was previously only
// MANUALLY JSX-gated (recorded in coverage.md) with no automated Go _AEShipGate
// test, so it was capped at verify=roundtrip. DuplicateComposition uses no
// version-specific feature, so it IS double-version reachable (AE2020+AE2025).
//
// Carrier: a from-scratch comp ORIG with two solids A,B and B parented to A.
// After DuplicateComposition(ORIG,"DUP") the surface AE must accept is:
//   - DUP exists as a distinct CompItem (own id), ORIG still intact
//   - both comps have layers [A,B] (deep-clone, not shared layer list)
//   - DUP's B.parent points at DUP's OWN A (intra-comp ref remap), NOT ORIG's A
//   - layer SOURCES are shared: DUP.A.source.id == ORIG.A.source.id (solids not
//     re-created — matches AE CompItem.duplicate())
//   - comp settings (width/height/frameRate/duration) deep-cloned from src
//
// The parent-remap + source-sharing checks are the meaningful correctness
// surface — a naive clone that left parent refs pointing back at the source
// comp's layers, or that re-created sources, would fail these.
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

func buildCompDuplicateDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	orig, err := aep.NewComposition(p, "ORIG", 640, 360, 30, 2)
	if err != nil {
		t.Fatalf("NewComposition ORIG: %v", err)
	}
	for _, n := range []string{"A", "B"} {
		if _, err := aep.NewSolidLayer(orig, n, 100, 100, [3]float64{0.8, 0.6, 0.2}); err != nil {
			t.Fatalf("NewSolidLayer %s: %v", n, err)
		}
	}

	// Reopen to materialize real layer IDs, then parent B→A so the duplicate
	// has an intra-comp ref to remap.
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	src := rp.CompositionByName("ORIG")
	if src == nil {
		t.Fatal("reopened ORIG missing")
	}
	a := src.LayerByName("A")
	b := src.LayerByName("B")
	if a == nil || b == nil {
		t.Fatalf("ORIG layers A/B missing (A=%v B=%v)", a, b)
	}
	if err := b.SetParent(a.ID); err != nil {
		t.Fatalf("SetParent B→A: %v", err)
	}

	dup, err := aep.DuplicateComposition(rp, src, "DUP")
	if err != nil {
		t.Fatalf("DuplicateComposition: %v", err)
	}
	if dup.Name != "DUP" || len(dup.Layers) != 2 {
		t.Fatalf("dup sanity: name=%q layers=%d", dup.Name, len(dup.Layers))
	}
	return rp
}

func runCompDuplicateGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/comp_duplicate_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_comp_duplicate.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildCompDuplicateDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "comp_duplicate_in.aep")
	resavedAEP := filepath.Join(tempDir, "comp_duplicate_resaved.aep")
	doneFile := filepath.Join(tempDir, "comp_duplicate.done")

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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("comp_duplicate %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("comp_duplicate %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: the duplicated comp survives AE's re-encode and the Go parser
	// reads back two comps each with layers A,B.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	for _, name := range []string{"ORIG", "DUP"} {
		c := re.CompositionByName(name)
		if c == nil {
			t.Errorf("resaved comp %s missing", name)
			continue
		}
		got := make([]string, len(c.Layers))
		for i, l := range c.Layers {
			got[i] = l.Name
		}
		if strings.Join(got, ",") != "A,B" {
			t.Errorf("resaved %s layers = %v, want [A B]", name, got)
		}
	}
}

func TestCompDuplicate_AEShipGate_AE2020(t *testing.T) {
	runCompDuplicateGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestCompDuplicate_AEShipGate_AE2025(t *testing.T) {
	runCompDuplicateGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
