// internal/aep/layer_transform_shipgate_test.go
//
// AE ship gate for SetLayerTransform (gap #1: animated transform on TEMPLATED
// layers). Builds a fresh project with one Go-created text layer "GLITCH", then
// SetLayerTransform installs a full canonical Transform Group carrying a
// non-default Anchor + 3-kf Position + 3-kf Opacity onto a layer whose template
// ELIDED those channels. AE must accept the file, type it as text, and read the
// materialized animated transform back via valueAtTime. Resaves for the Go-side
// preservation check.
//
// Gated by AE_SHIP_GATE. Verified PASS on AE 2020 + AE 2025 (2026-06-21).
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runLayerTransformGate(t *testing.T, target aep.AETarget, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/layer_transform_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_layer_transform.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "XForm", 1920, 1080, 30, 6)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewTextLayer(comp, "GLITCH"); err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	layer := rp.Compositions[0].Layers[0]

	tr := aep.NewLayerTransform()
	if err := tr.AnchorPoint().SetStaticValue([2]float64{9, -436}); err != nil {
		t.Fatal(err)
	}
	for _, kf := range []struct {
		tm   float64
		x, y float64
	}{{0, 600, 540}, {1, 1320, 540}, {2, 960, 300}} {
		if err := tr.Position().AddKeyframeLinear(kf.tm, [2]float64{kf.x, kf.y}); err != nil {
			t.Fatal(err)
		}
	}
	for _, kf := range []struct{ tm, v float64 }{{0, 10}, {1, 100}, {2, 40}} {
		if err := tr.Opacity().AddKeyframeLinear(kf.tm, kf.v); err != nil {
			t.Fatal(err)
		}
	}
	if err := aep.SetLayerTransform(layer, tr); err != nil {
		t.Fatalf("SetLayerTransform: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "xform_in.aep")
	resavedAEP := filepath.Join(tempDir, "xform_resaved.aep")
	doneFile := filepath.Join(tempDir, "xform.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
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

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 200)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("%s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s layer-transform ship gate FAIL:\n%s", ver, body)
	}
}

func TestSetLayerTransform_AEShipGate_AE2020(t *testing.T) {
	runLayerTransformGate(t, aep.TargetAE2020, ae2020(), "AE2020")
}

func TestSetLayerTransform_AEShipGate_AE2025(t *testing.T) {
	runLayerTransformGate(t, aep.TargetAE2025, ae2025(), "AE2025")
}
