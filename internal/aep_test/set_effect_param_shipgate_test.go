// internal/aep/set_effect_param_shipgate_test.go
//
// AE ship gate for SetEffectParam (effect-parameter materialization,
// synthesis-lite): on a 100% Go-built file (NewProject → NewShapeLayer →
// Reopen → AddEffect Gaussian Blur → SetEffectParam ×3, all params
// default-elided before the call), AE must open without corruption, read back
// the three materialized values, and keep them across its own resave.
//
// Gated by AE_SHIP_GATE. Uses test_data/generators/verify_effect_param.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runSetEffectParamGate(t *testing.T, aeExe, label string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/effect_param_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_effect_param.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var l *aep.Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			if cl.Name == "S" {
				l = cl
			}
		}
	}
	if l == nil {
		t.Fatal("reopened project: layer S not found")
	}
	fx, err := aep.AddEffect(l, aep.EffectGaussianBlur)
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}
	if len(fx.Parameters) != 1 {
		t.Fatalf("precondition: default GB should expose 1 param, got %d", len(fx.Parameters))
	}
	// Drop Shadow params have no per-param template — they exercise the
	// generic per-control-type fallback (pard-patched tdmn/tdsn/tdum/tduM).
	ds, err := aep.AddEffect(l, aep.EffectDropShadow)
	if err != nil {
		t.Fatalf("AddEffect(DropShadow): %v", err)
	}
	set := func(e *aep.Effect, mn string, v any) {
		t.Helper()
		if _, err := aep.SetEffectParam(l, e, mn, v); err != nil {
			t.Fatalf("SetEffectParam(%s): %v", mn, err)
		}
	}
	set(fx, "ADBE Gaussian Blur 2-0001", 25.0)
	set(fx, "ADBE Gaussian Blur 2-0002", 2.0)
	set(fx, "ADBE Gaussian Blur 2-0003", 1.0)
	set(ds, "ADBE Drop Shadow-0001", []float64{255, 51, 102, 153}) // Shadow Color (color, generic)
	set(ds, "ADBE Drop Shadow-0003", 90.0)                         // Direction (angle, generic)
	set(ds, "ADBE Drop Shadow-0004", 20.0)                         // Distance (scalar, generic)
	set(ds, "ADBE Drop Shadow-0005", 10.0)                         // Softness (scalar, generic)
	set(ds, "ADBE Drop Shadow-0006", 1.0)                          // Shadow Only (boolean, generic)

	// Expression-control effects: per-param templates for the remaining control
	// types (angle / color / 2D / 3D / slider). Point values are fractions of
	// the layer's coordinate space — the COMP's 1920x1080 for a source-less
	// shape layer — so JSX reads them back in pixels.
	addAndSet := func(effectMN string, v any) {
		t.Helper()
		e, err := aep.AddEffect(l, effectMN)
		if err != nil {
			t.Fatalf("AddEffect(%s): %v", effectMN, err)
		}
		set(e, effectMN+"-0001", v)
	}
	addAndSet(aep.EffectAngleControl, 33.0)
	addAndSet(aep.EffectColorControl, []float64{255, 51, 102, 153})
	addAndSet(aep.EffectPointControl, []float64{0.25, 0.75})
	addAndSet(aep.EffectPoint3DControl, []float64{0.25, 0.75, 0.5})
	addAndSet(aep.EffectSliderControl, 12.25)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "setparam_in.aep")
	resavedAEP := filepath.Join(tempDir, "setparam_resaved.aep")
	doneFile := filepath.Join(tempDir, "setparam_"+label+".done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expect":[`+
		`{"effect":"ADBE Gaussian Blur 2","param":"ADBE Gaussian Blur 2-0001","value":25},`+
		`{"effect":"ADBE Gaussian Blur 2","param":"ADBE Gaussian Blur 2-0002","value":2},`+
		`{"effect":"ADBE Gaussian Blur 2","param":"ADBE Gaussian Blur 2-0003","value":1},`+
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0001","value":[0.2,0.4,0.6,1]},`+
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0003","value":90},`+
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0004","value":20},`+
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0005","value":10},`+
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0006","value":1},`+
		`{"effect":"ADBE Angle Control","param":"ADBE Angle Control-0001","value":33},`+
		`{"effect":"ADBE Color Control","param":"ADBE Color Control-0001","value":[0.2,0.4,0.6,1]},`+
		`{"effect":"ADBE Point Control","param":"ADBE Point Control-0001","value":[480,810]},`+
		`{"effect":"ADBE Point3D Control","param":"ADBE Point3D Control-0001","value":[480,810,540]},`+
		`{"effect":"ADBE Slider Control","param":"ADBE Slider Control-0001","value":12.25}]}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("%s AE readback:\n%s", label, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s ship gate FAIL:\n%s", label, body)
	}

	// Go side: AE's resave must keep the non-default materialized params.
	// -0003 is asserted by the JSX's reopen-the-resave pass instead: its set
	// value (true) EQUALS the AE 2025 default (flipped from AE 2020's false),
	// so that version legitimately re-elides the stream on save — the
	// effective value survives via the default (value-keyed elision,
	// incidents/effect-param-elision-synthesis-lite.md).
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("resaved: no layer with effects")
	}
	want := map[string]any{
		"ADBE Gaussian Blur 2-0001": 25.0,
		"ADBE Gaussian Blur 2-0002": 2.0,
		"ADBE Drop Shadow-0001":     []float64{255, 51, 102, 153},
		"ADBE Drop Shadow-0003":     90.0,
		"ADBE Drop Shadow-0004":     20.0,
		"ADBE Drop Shadow-0005":     10.0,
		"ADBE Drop Shadow-0006":     1.0,
	}
	got := map[string]any{}
	for _, e := range rl.Effects {
		for _, p := range e.Parameters {
			got[p.MatchName] = p.StaticValue
		}
	}
	for mn, v := range want {
		if !staticValueEq(got[mn], v) {
			t.Errorf("resaved %s = %v, want %v (AE dropped/reverted the materialized param)", mn, got[mn], v)
		}
	}
	for mn, v := range map[string]any{
		"ADBE Angle Control-0001":   33.0,
		"ADBE Color Control-0001":   []float64{255, 51, 102, 153},
		"ADBE Point Control-0001":   []float64{0.25, 0.75},
		"ADBE Point3D Control-0001": []float64{0.25, 0.75, 0.5},
		"ADBE Slider Control-0001":  12.25,
	} {
		if !staticValueEq(got[mn], v) {
			t.Errorf("resaved %s = %v, want %v (AE dropped/reverted the materialized param)", mn, got[mn], v)
		}
	}
	if f, ok := got["ADBE Gaussian Blur 2-0003"].(float64); ok && f != 1 {
		t.Errorf("resaved -0003 present with value %v, want 1", f)
	}
}

func TestSetEffectParam_AEShipGate_AE2020(t *testing.T) {
	runSetEffectParamGate(t, ae2020(), "ae2020", aep.TargetAE2020)
}
func TestSetEffectParam_AEShipGate_AE2025(t *testing.T) {
	runSetEffectParamGate(t, ae2025(), "ae2025", aep.TargetAE2025)
}
