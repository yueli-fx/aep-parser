// internal/aep/mg_wiggle_modrt_shipgate_test.go
//
// AE round-trip ship gate for the Wiggle modulation noise-phase / coherence
// knobs that have no single-frame categorically-correct pixel result (a phase
// shift selects a statistically-equivalent alternate random instance — same
// class as RandomSeed), so they are verified by AE-acceptance + value
// round-trip rather than a pixel assertion (evidence-based, per the
// delivery-contract honesty principle):
//
//	ROUGH  · ADBE Vector Filter - Roughen : Temporal Phase, Spatial Phase
//	WTRANS · ADBE Vector Filter - Wiggler : Correlation, Temporal Phase,
//	                                        Spatial Phase
//
// All five are AE-default-elided and materialized via synthesis-insert. The gate
// proves AE accepts the from-scratch file (does not reject it as corrupt), reads
// each value back from its specific filter, resaves, and the spliced cdats
// survive the re-encode.
//
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

func buildMGWiggleModRTDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "WGMODRT", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// ROUGH: Wiggle Paths with non-default Temporal + Spatial Phase.
	rough, err := aep.NewShapeLayer(comp, "ROUGH")
	if err != nil {
		t.Fatalf("NewShapeLayer ROUGH: %v", err)
	}
	rRect, err := rough.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("ROUGH AddRect: %v", err)
	}
	if err := rRect.SetSize([2]float64{300, 300}); err != nil {
		t.Fatalf("ROUGH rect SetSize: %v", err)
	}
	rFill, err := rough.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("ROUGH AddFill: %v", err)
	}
	if err := rFill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("ROUGH SetColor: %v", err)
	}
	wp, err := rough.RootGroup().AddWigglePaths()
	if err != nil {
		t.Fatalf("AddWigglePaths: %v", err)
	}
	if err := wp.SetSize(40); err != nil {
		t.Fatalf("wp SetSize: %v", err)
	}
	if err := wp.SetDetail(10); err != nil {
		t.Fatalf("wp SetDetail: %v", err)
	}
	if err := wp.SetRandomSeed(3); err != nil {
		t.Fatalf("wp SetRandomSeed: %v", err)
	}
	if err := wp.SetTemporalPhase(45); err != nil {
		t.Fatalf("wp SetTemporalPhase: %v", err)
	}
	if err := wp.SetSpatialPhase(30); err != nil {
		t.Fatalf("wp SetSpatialPhase: %v", err)
	}
	if err := rough.Position().SetStaticValue([2]float64{600, 540}); err != nil {
		t.Fatalf("ROUGH Position: %v", err)
	}

	// WTRANS: Wiggle Transform with non-default Correlation + Temporal + Spatial
	// Phase (and a Position amplitude so the filter is meaningful).
	wtrans, err := aep.NewShapeLayer(comp, "WTRANS")
	if err != nil {
		t.Fatalf("NewShapeLayer WTRANS: %v", err)
	}
	wRect, err := wtrans.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("WTRANS AddRect: %v", err)
	}
	if err := wRect.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("WTRANS rect SetSize: %v", err)
	}
	wFill, err := wtrans.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("WTRANS AddFill: %v", err)
	}
	if err := wFill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("WTRANS SetColor: %v", err)
	}
	wt, err := wtrans.RootGroup().AddWiggleTransform()
	if err != nil {
		t.Fatalf("AddWiggleTransform: %v", err)
	}
	if err := wt.Transform().SetPosition([2]float64{80, 80}); err != nil {
		t.Fatalf("wt SetPosition: %v", err)
	}
	if err := wt.SetCorrelation(20); err != nil {
		t.Fatalf("wt SetCorrelation: %v", err)
	}
	if err := wt.SetTemporalPhase(40); err != nil {
		t.Fatalf("wt SetTemporalPhase: %v", err)
	}
	if err := wt.SetSpatialPhase(50); err != nil {
		t.Fatalf("wt SetSpatialPhase: %v", err)
	}
	if err := wtrans.Position().SetStaticValue([2]float64{1320, 540}); err != nil {
		t.Fatalf("WTRANS Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

func runMGWiggleModRTGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_wiggle_modrt_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_wiggle_modrt.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGWiggleModRTDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_wiggle_modrt_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_wiggle_modrt_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_wiggle_modrt.done")

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
	t.Logf("mg wiggle modrt %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg wiggle modrt %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: the spliced phase/correlation leaves survive AE's re-encode.
	// Both filters share the generic Temporal/Spatial Phase match-names, so two of
	// each must be present after resave (one Roughen, one Wiggler).
	root := parseAEP(t, resavedAEP)
	for _, c := range []struct {
		name string
		want int
	}{
		{"ADBE Vector Temporal Phase", 2},
		{"ADBE Vector Spatial Phase", 2},
	} {
		if got := countStream(root, c.name); got < c.want {
			t.Errorf("resaved: %s present %d times, want >= %d — a spliced leaf was dropped", c.name, got, c.want)
		}
	}
}

func TestMGWiggleModRT_AEShipGate_AE2020(t *testing.T) {
	runMGWiggleModRTGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGWiggleModRT_AEShipGate_AE2025(t *testing.T) {
	runMGWiggleModRTGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
