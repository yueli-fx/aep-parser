// internal/aep/orbit_demo_shipgate_test.go
//
// AE ship gate for the from-scratch orbit animation demo (delivery-contract.md).
// The demo (tmp_debug/anim_demo) is an END-TO-END composition — a backdrop, a
// stroked ring, and three rotation-animated orbiting dots — built 100% in Go.
// Per the delivery contract, a composition/end-to-end artifact needs its OWN
// gate; point-gates on the individual New*/keyframe paths don't back the combo.
//
// This builds the exact demo bytes, has AE 2020 + 2025 open them, and proves:
//   - AE does not HANG (a hang means .done never lands → ae_run times out → FAIL)
//   - all 5 layers survive (no silent-drop)
//   - the dots' rotation animation is live (rotation differs across time)
//   - the keyframes survive AE's own resave (re-parsed back in Go)
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

// buildOrbitDemo constructs the demo project (mirrors tmp_debug/anim_demo): all
// motion via layer Rotation (2 keyframes, the gated pattern) + an anchor-offset
// circle, no spatial position keyframes, no effects.
func buildOrbitDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	const cx, cy, orbitR = 960.0, 540.0, 330.0

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "ORBIT", 1920, 1080, 30, 6)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	rect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
	}
	bgFill, err := bg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill BG: %v", err)
	}
	if err := bgFill.SetColor([4]float64{0.05, 0.06, 0.12, 1}); err != nil {
		t.Fatalf("BG SetColor: %v", err)
	}
	if err := bg.Position().SetStaticValue([2]float64{cx, cy}); err != nil {
		t.Fatalf("BG Position: %v", err)
	}

	ring, err := aep.NewShapeLayer(comp, "Ring")
	if err != nil {
		t.Fatalf("NewShapeLayer Ring: %v", err)
	}
	rel, err := ring.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse Ring: %v", err)
	}
	if err := rel.SetSize([2]float64{orbitR * 2, orbitR * 2}); err != nil {
		t.Fatalf("ring SetSize: %v", err)
	}
	rst, err := ring.RootGroup().AddStroke()
	if err != nil {
		t.Fatalf("AddStroke Ring: %v", err)
	}
	if err := rst.SetColor([4]float64{0.25, 0.85, 1.0, 1}); err != nil {
		t.Fatalf("ring stroke SetColor: %v", err)
	}
	if err := rst.SetWidth(6); err != nil {
		t.Fatalf("ring SetWidth: %v", err)
	}
	if err := ring.Position().SetStaticValue([2]float64{cx, cy}); err != nil {
		t.Fatalf("Ring Position: %v", err)
	}

	dots := []struct {
		name  string
		color [4]float64
		start float64
	}{
		{"Dot1", [4]float64{1.0, 0.55, 0.1, 1}, 0},
		{"Dot2", [4]float64{1.0, 0.2, 0.55, 1}, 120},
		{"Dot3", [4]float64{0.2, 1.0, 0.7, 1}, 240},
	}
	for _, d := range dots {
		dl, err := aep.NewShapeLayer(comp, d.name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", d.name, err)
		}
		el, err := dl.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("AddEllipse %s: %v", d.name, err)
		}
		if err := el.SetSize([2]float64{72, 72}); err != nil {
			t.Fatalf("%s SetSize: %v", d.name, err)
		}
		if err := el.SetPosition([2]float64{0, -orbitR}); err != nil {
			t.Fatalf("%s SetPosition: %v", d.name, err)
		}
		f, err := dl.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("AddFill %s: %v", d.name, err)
		}
		if err := f.SetColor(d.color); err != nil {
			t.Fatalf("%s SetColor: %v", d.name, err)
		}
		if err := dl.Position().SetStaticValue([2]float64{cx, cy}); err != nil {
			t.Fatalf("%s Position: %v", d.name, err)
		}
		if err := dl.Rotation().AddKeyframeLinear(0, d.start); err != nil {
			t.Fatalf("%s Rotation kf0: %v", d.name, err)
		}
		if err := dl.Rotation().AddKeyframeLinear(6, d.start+360); err != nil {
			t.Fatalf("%s Rotation kf1: %v", d.name, err)
		}
	}
	return p
}

func runOrbitDemoGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/orbit_demo_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_orbit_demo.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildOrbitDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "orbit_in.aep")
	resavedAEP := filepath.Join(tempDir, "orbit_resaved.aep")
	doneFile := filepath.Join(tempDir, "orbit.done")

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
	t.Logf("orbit demo %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("orbit demo %s ship gate FAIL:\n%s", ver, body)
	}

	// Acceptance proof: AE's own resave still carries 5 layers + the dots'
	// rotation keyframes.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	cc := re.Compositions[0]
	if len(cc.Layers) != 5 {
		t.Fatalf("resaved layers = %d, want 5", len(cc.Layers))
	}
	for _, name := range []string{"Dot1", "Dot2", "Dot3"} {
		l := cc.LayerByName(name)
		if l == nil {
			t.Errorf("resaved: %s missing", name)
			continue
		}
		if r := l.Rotation(); r == nil || len(r.Keyframes) != 2 {
			t.Errorf("resaved %s rotation keyframes lost", name)
		}
	}
}

func TestOrbitDemo_AEShipGate_AE2020(t *testing.T) {
	runOrbitDemoGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestOrbitDemo_AEShipGate_AE2025(t *testing.T) {
	runOrbitDemoGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
