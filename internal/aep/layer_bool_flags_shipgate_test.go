// internal/aep/layer_bool_flags_shipgate_test.go
//
// AE ship gate for layer-set ldta flag setters that need no special source
// (so a from-scratch shape layer carries them), capped at verify=roundtrip:
//
//   - SetAutoOrient      → AE layer.autoOrient (enum; AlongPath needs a path,
//                          so the carrier gets position keyframes)
//   - SetCollapseTransform → AE layer.collapseTransformation (shape =
//                          continuously rasterize)
//
// These are ldta-byte/bit writes (not property-tree), so unlike the transform
// statics they don't need a materialized property — a bare layer works.
// (SetStretch belongs to this family byte-wise but is a confirmed false-green —
// see the note in buildLayerBoolDemo.)
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

func buildLayerBoolDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "BOOL", 640, 360, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	addShape := func(name string) *aep.ShapeLayer {
		s, err := aep.NewShapeLayer(comp, name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", name, err)
		}
		el, err := s.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("AddEllipse %s: %v", name, err)
		}
		if err := el.SetSize([2]float64{80, 80}); err != nil {
			t.Fatalf("SetSize %s: %v", name, err)
		}
		f, err := s.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("AddFill %s: %v", name, err)
		}
		if err := f.SetColor([4]float64{0.6, 0.4, 0.9, 1}); err != nil {
			t.Fatalf("SetColor %s: %v", name, err)
		}
		return s
	}

	// NOTE: SetStretch is intentionally NOT gated here — it's a confirmed
	// false-green. SetStretch(0.5) writes the @0x08/@0x6C dividend/divisor and
	// Go round-trips it, but AE reads layer.stretch=100 regardless (even with a
	// precomp source): AE recomputes stretch from the in/out span, which the
	// setter does not adjust. See incident layer-setstretch-ae-recomputes-span.
	aut := addShape("AUT")
	// AlongPath needs a motion path — give AUT position keyframes.
	if err := aut.Position().AddKeyframeLinear(0, [2]float64{120, 180}); err != nil {
		t.Fatalf("AUT kf0: %v", err)
	}
	if err := aut.Position().AddKeyframeLinear(3, [2]float64{520, 180}); err != nil {
		t.Fatalf("AUT kf1: %v", err)
	}
	_ = addShape("COL")

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.Compositions[0]
	mustMut := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustMut("SetAutoOrient", cc.LayerByName("AUT").SetAutoOrient(aep.AutoOrientAlongPath))
	mustMut("SetCollapseTransform", cc.LayerByName("COL").SetCollapseTransform(true))

	return rp
}

func runLayerBoolGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/layer_bool_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_layer_bool.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildLayerBoolDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "layer_bool_in.aep")
	resavedAEP := filepath.Join(tempDir, "layer_bool_resaved.aep")
	doneFile := filepath.Join(tempDir, "layer_bool.done")

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
	t.Logf("layer bool %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layer bool %s ship gate FAIL:\n%s", ver, body)
	}
}

func TestLayerBool_AEShipGate_AE2020(t *testing.T) {
	runLayerBoolGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayerBool_AEShipGate_AE2025(t *testing.T) {
	runLayerBoolGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
