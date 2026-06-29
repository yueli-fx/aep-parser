// internal/aep/layer_3d_enable_shipgate_test.go
//
// AE ship gate for the 3D-enable foundation (roadmap priority 2): flip Is3D
// (ldta 0x26 bit2) on a from-scratch shape layer and confirm AE not only
// ACCEPTS it but materializes a full 3D layer from the single bit — threeDLayer
// true, Position expanded 2D→3D (z=0), and the Orientation / X·Y rotation /
// Material Options channels synthesized from defaults. So no channel synthesis
// is needed for ENABLE; AE re-creates the 3D transform tree on open.
//
// This is a DOM-readback gate (not pixel): a 3D layer with a DEFAULT transform
// (z=0, no rotation) renders identically to its 2D self, so the enable's
// surface is the 3D-ness AE reports, not pixels. The VISIBLE 3D transform
// (Z-depth / rotation under a camera) is a separate capability with its own
// render-pixel gate.
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

func runLayer3DEnableProbe(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/3d_enable_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_3d_enable.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "3DENABLE", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	box, err := aep.NewShapeLayer(comp, "BOX")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, err := box.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{300, 300}); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	fill, err := box.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("SetColor: %v", err)
	}
	if err := box.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("Position: %v", err)
	}

	// Reopen (parse-the-clone) so the layer carries an ldta backref; SetIs3D
	// flips the ldta 0x26 bit2 length-preservingly on the parsed layer.
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	rbox := rp.Compositions[0].LayerByName("BOX")
	if rbox == nil {
		t.Fatal("reopened BOX missing")
	}
	if err := rbox.SetIs3D(true); err != nil {
		t.Fatalf("SetIs3D: %v", err)
	}
	p = rp

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "3d_enable_in.aep")
	resavedAEP := filepath.Join(tempDir, "3d_enable_resaved.aep")
	doneFile := filepath.Join(tempDir, "3d_enable.done")

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
	t.Logf("3d-enable %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d-enable %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: the layer survives AE's re-encode as a 3D layer (the
	// materialized Orientation / X·Y rotation channels are default → AE elides
	// them on disk, so they read back nil; the Is3D bit is the durable signal).
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	reBox := re.Compositions[0].LayerByName("BOX")
	if reBox == nil {
		t.Fatalf("resaved BOX missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	if !reBox.Is3D {
		t.Errorf("resaved BOX Is3D=false — AE dropped the 3D switch")
	}
}

func TestLayer3DEnable_AEShipGate_AE2020(t *testing.T) {
	runLayer3DEnableProbe(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayer3DEnable_AEShipGate_AE2025(t *testing.T) {
	runLayer3DEnableProbe(t, ae2025(), "AE2025", aep.TargetAE2025)
}
