// internal/aep/set_effect_param_shipgate_test.go
//
// AE ship gate for SetEffectParam (effect-parameter materialization,
// synthesis-lite): on a 100% Go-built file (NewProject → NewShapeLayer →
// Reopen → AddEffect Gaussian Blur → SetEffectParam ×3, all params
// default-elided before the call), AE must open without corruption, read back
// the three materialized values, and keep them across its own resave.
//
// Gated by AE_SHIP_GATE. Uses test_data/verify_effect_param.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runSetEffectParamGate(t *testing.T, aeExe, label string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/effect_param_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_effect_param.jsx`
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
	set := func(e *aep.Effect, mn string, v float64) {
		t.Helper()
		if _, err := aep.SetEffectParam(l, e, mn, v); err != nil {
			t.Fatalf("SetEffectParam(%s): %v", mn, err)
		}
	}
	set(fx, "ADBE Gaussian Blur 2-0001", 25)
	set(fx, "ADBE Gaussian Blur 2-0002", 2)
	set(fx, "ADBE Gaussian Blur 2-0003", 1)
	set(ds, "ADBE Drop Shadow-0004", 20) // Distance (scalar, generic)
	set(ds, "ADBE Drop Shadow-0005", 10) // Softness (scalar, generic)
	set(ds, "ADBE Drop Shadow-0006", 1)  // Shadow Only (boolean, generic)

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
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0004","value":20},`+
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0005","value":10},`+
		`{"effect":"ADBE Drop Shadow","param":"ADBE Drop Shadow-0006","value":1}]}`,
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
	want := map[string]float64{
		"ADBE Gaussian Blur 2-0001": 25,
		"ADBE Gaussian Blur 2-0002": 2,
		"ADBE Drop Shadow-0004":     20,
		"ADBE Drop Shadow-0005":     10,
		"ADBE Drop Shadow-0006":     1,
	}
	got := map[string]float64{}
	for _, e := range rl.Effects {
		for _, p := range e.Parameters {
			if f, ok := p.StaticValue.(float64); ok {
				got[p.MatchName] = f
			}
		}
	}
	for mn, v := range want {
		if got[mn] != v {
			t.Errorf("resaved %s = %v, want %v (AE dropped/reverted the materialized param)", mn, got[mn], v)
		}
	}
	if f, present := got["ADBE Gaussian Blur 2-0003"]; present && f != 1 {
		t.Errorf("resaved -0003 present with value %v, want 1", f)
	}
}

func TestSetEffectParam_AEShipGate_AE2020(t *testing.T) {
	runSetEffectParamGate(t, ae2020(), "ae2020", aep.TargetAE2020)
}
func TestSetEffectParam_AEShipGate_AE2025(t *testing.T) {
	runSetEffectParamGate(t, ae2025(), "ae2025", aep.TargetAE2025)
}
