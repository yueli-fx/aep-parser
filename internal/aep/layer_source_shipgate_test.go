// internal/aep/layer_source_shipgate_test.go
//
// AE ship gate for the layer source-swap pair SetSource + ReplaceSource
// (verify=roundtrip). Two precomp layers in MAIN initially source comp "A";
// after Reopen one is swapped to "B" via ReplaceSource and the other via
// SetSource(id). AE reads back layer.source.name == "B" (A→B contrast proves
// the swap took effect, not a no-op).
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func buildLayerSourceDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	main, err := aep.NewComposition(p, "MAIN", 640, 360, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition MAIN: %v", err)
	}
	mkSrc := func(name string, rgb [3]float64) *aep.Composition {
		c, err := aep.NewComposition(p, name, 320, 180, 30, 4)
		if err != nil {
			t.Fatalf("NewComposition %s: %v", name, err)
		}
		if _, err := aep.NewSolidLayer(c, name+"s", 320, 180, rgb); err != nil {
			t.Fatalf("NewSolidLayer %s: %v", name, err)
		}
		return c
	}
	a := mkSrc("A", [3]float64{0.9, 0.2, 0.2})
	_ = mkSrc("B", [3]float64{0.2, 0.4, 0.9})
	if _, err := aep.NewPrecompLayer(main, a, "L1"); err != nil {
		t.Fatalf("NewPrecompLayer L1: %v", err)
	}
	if _, err := aep.NewPrecompLayer(main, a, "L2"); err != nil {
		t.Fatalf("NewPrecompLayer L2: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.CompositionByName("MAIN")
	bb := rp.CompositionByName("B")
	if cc == nil || bb == nil {
		t.Fatal("reopened MAIN/B missing")
	}
	if err := cc.LayerByName("L1").ReplaceSource(bb, false); err != nil {
		t.Fatalf("ReplaceSource: %v", err)
	}
	if err := cc.LayerByName("L2").SetSource(bb.ItemID()); err != nil {
		t.Fatalf("SetSource: %v", err)
	}
	return rp
}

func runLayerSourceGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/layer_source_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_layer_source.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildLayerSourceDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layer_source_in.aep")
	resavedAEP := filepath.Join(tempDir, "layer_source_resaved.aep")
	doneFile := filepath.Join(tempDir, "layer_source.done")

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

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("layer source %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layer source %s ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	if mc := re.CompositionByName("MAIN"); mc == nil || mc.LayerByName("L1") == nil {
		t.Fatalf("%s resaved: MAIN/L1 missing", ver)
	}
}

func TestLayerSource_AEShipGate_AE2020(t *testing.T) {
	runLayerSourceGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayerSource_AEShipGate_AE2025(t *testing.T) {
	runLayerSourceGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
