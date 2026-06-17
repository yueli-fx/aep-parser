// internal/aep/layer_xform_static_shipgate_test.go
//
// AE ship gate for layer-set transform statics + frame-time setters, capped at
// verify=roundtrip ("广泛被渲染 gate 间接覆盖" but no explicit value gate). One
// solid layer, every setter applied from scratch, AE DOM value readback:
//
//   - SetPosition / SetAnchorPoint / SetScale / SetOpacity (transform statics)
//   - SetFrameStartTime / SetFrameInPoint / SetFrameOutPoint (frame→time,
//     delegate to SetStartTime/SetInPoint/SetOutPoint)
//
// Units: SetScale 1.0=100% → AE %, SetOpacity 0..1 → AE %, position/anchor px.
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

func buildLayerXformDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "XFORM", 640, 360, 30, 6)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	// A from-scratch layer ELIDES default transform props (transform-group-
	// default-omission). The scene Layer.Set* setters mutate an *existing*
	// property, so materialize Position/Scale/Anchor/Opacity via the shape
	// create-path at NON-default placeholders first (≠ the targets below, so a
	// no-op mutate would be caught), then scene-mutate after Reopen.
	sl, err := aep.NewShapeLayer(comp, "L")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	el, err := sl.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse: %v", err)
	}
	if err := el.SetSize([2]float64{80, 80}); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	f, err := sl.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	if err := f.SetColor([4]float64{0.4, 0.6, 0.9, 1}); err != nil {
		t.Fatalf("SetColor: %v", err)
	}
	mustMat := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("materialize %s: %v", label, err)
		}
	}
	mustMat("Position", sl.Position().SetStaticValue([2]float64{111, 111}))
	mustMat("Scale", sl.Scale().SetStaticValue([2]float64{1.1, 1.1}))
	mustMat("AnchorPoint", sl.AnchorPoint().SetStaticValue([2]float64{10, 10}))
	mustMat("Opacity", sl.Opacity().SetStaticValue(0.9))

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	L := rp.Compositions[0].LayerByName("L")
	if L == nil {
		t.Fatal("reopened layer L missing")
	}
	mustMut := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	prop := func(p *aep.Property, name string) *aep.Property {
		t.Helper()
		if p == nil {
			t.Fatalf("L.%s nil (elided?)", name)
		}
		return p
	}
	// Build a value slice matching the property's component count (layer
	// transform props parse 3D / z=0 even on a 2D layer; AE DOM returns 2D).
	fit := func(p *aep.Property, x, y, z float64) []float64 {
		if p.Components >= 3 {
			return []float64{x, y, z}
		}
		return []float64{x, y}
	}

	pos := prop(L.Position(), "Position")
	anc := prop(L.AnchorPoint(), "AnchorPoint")
	scl := prop(L.Scale(), "Scale")

	mustMut("SetPosition", L.SetPosition(fit(pos, 400, 300, 0)))
	mustMut("SetAnchorPoint", L.SetAnchorPoint(fit(anc, 25, 25, 0)))
	mustMut("SetScale", L.SetScale(fit(scl, 1.25, 1.25, 1.0))) // 1.25 → AE 125%
	mustMut("SetOpacity", L.SetOpacity(0.6))                   // 0.6 → AE 60%
	// in/out are source-relative (trim); AE's comp-absolute inPoint/outPoint =
	// startTime + these. With startTime 0.5: AE shows in 1.5, out 5.5.
	mustMut("SetFrameStartTime", L.SetFrameStartTime(15))      // 15/30 = 0.5s
	mustMut("SetFrameInPoint", L.SetFrameInPoint(30))          // 30/30 = 1.0s rel
	mustMut("SetFrameOutPoint", L.SetFrameOutPoint(150))       // 150/30 = 5.0s rel

	return rp
}

func runLayerXformGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/layer_xform_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_layer_xform.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildLayerXformDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layer_xform_in.aep")
	resavedAEP := filepath.Join(tempDir, "layer_xform_resaved.aep")
	doneFile := filepath.Join(tempDir, "layer_xform.done")

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
	t.Logf("layer xform %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layer xform %s ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: the static transform values survive AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	L := re.Compositions[0].LayerByName("L")
	if L == nil {
		t.Fatal("resaved: L missing")
	}
	if op := L.Opacity(); op == nil || op.StaticValue == nil {
		t.Errorf("resaved: opacity static missing")
	}
}

func TestLayerXform_AEShipGate_AE2020(t *testing.T) {
	runLayerXformGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayerXform_AEShipGate_AE2025(t *testing.T) {
	runLayerXformGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
