// internal/aep/add_effect_shipgate_test.go
//
// AE ship gate for AddEffect on the Effect Parade. Reuses the generic
// runPropStructShipGate harness (verify_property_struct.jsx): loads the
// AE-2020-native baseline (3 effects), Go-adds a 4th effect from an embedded
// template, WriteAEP, and has AE open the mutated file — proving AE ACCEPTS the
// Go-spliced effect (no data-loss / corrupt) and reads back the expected 4
// effects in order. Resaves so the Go side confirms AE kept the addition.
//
// Covers the full 12-template library — every embedded effect template is AE
// dual-version ship-gated (payload sizes 1.7KB Gaussian Blur … 20.6KB Pro
// Levels2), so the splice + bottom-up size recompute is proven per template,
// not per mechanism. Go-only round-trip remains in
// TestAddEffect_AllTemplates_RoundTrip.
//
// Gated by AE_SHIP_GATE. Baseline built by test_data/re_property_struct.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// addEffectSample is the ship-gated set: the full embedded template library.
// Baseline parade is [Gaussian Blur, Tint, Fill]; the added effect appends last.
var addEffectSample = []string{
	aep.EffectGaussianBlur,
	aep.EffectFill,
	aep.EffectTint,
	aep.EffectLevels,
	aep.EffectLevelsIndividual,
	aep.EffectBrightnessContrast,
	aep.EffectTritone,
	aep.EffectHueSaturation,
	aep.EffectBoxBlur,
	aep.EffectGlow,
	aep.EffectInvert,
	aep.EffectExposure,
	aep.EffectDropShadow,
	aep.EffectSharpen,
	aep.EffectMosaic,
	aep.EffectNoise,
	aep.EffectTransform,
	aep.EffectGradientRamp,
	aep.EffectFractalNoise,
	aep.EffectMotionTile,
	aep.EffectDirectionalBlur,
	aep.EffectLinearWipe,
	aep.EffectWaveWarp,
	aep.EffectCurves,
	aep.EffectSliderControl,
	aep.EffectPointControl,
	aep.EffectColorControl,
	aep.EffectAngleControl,
	aep.EffectCheckboxControl,
}

func runAddEffectGate(t *testing.T, aeExe, ver string) {
	for _, fxName := range addEffectSample {
		expect := []string{"ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill", fxName}
		mutate := func(l *aep.Layer) error {
			_, err := aep.AddEffect(l, fxName)
			return err
		}
		t.Run(fxName, func(t *testing.T) {
			runPropStructShipGate(t, aeExe, "addeffect-"+ver+"-"+fxName, mutate, expect)
		})
	}
}

func TestAddEffect_AEShipGate_AE2020(t *testing.T) { runAddEffectGate(t, ae2020(), "AE2020") }
func TestAddEffect_AEShipGate_AE2025(t *testing.T) { runAddEffectGate(t, ae2025(), "AE2025") }

// runAddEffectAutoParadeGate ship-gates the parade auto-create path end to end
// on a 100% Go-built file: NewProject → NewComposition → NewShapeLayer →
// Reopen (upgrade the built layer to a parsed one) → AddEffect, which must
// splice an empty Effect Parade before the Transform Group and the effect pair
// into it. AE opening the file without corruption + reading back the single
// effect + keeping it across its own resave is the acceptance proof.
func runAddEffectAutoParadeGate(t *testing.T, aeExe, label string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/property_struct_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_property_struct.jsx`
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
	if l.EffectsParade() != nil {
		t.Fatal("layer already has a parade; gate would not exercise auto-create")
	}
	if _, err := aep.AddEffect(l, aep.EffectGaussianBlur); err != nil {
		t.Fatalf("AddEffect: %v", err)
	}

	expect := []string{aep.EffectGaussianBlur}
	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "autoparade_in.aep")
	resavedAEP := filepath.Join(tempDir, "autoparade_resaved.aep")
	doneFile := filepath.Join(tempDir, "autoparade.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(expect))
	for i, e := range expect {
		quoted[i] = fmt.Sprintf("%q", e)
	}
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expect":[%s]}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), strings.Join(quoted, ","))
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

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("resaved: no layer with effects (AE dropped the auto-created parade)")
	}
	if got := paradeChildNames(rl); !eq(got, expect) {
		t.Errorf("resaved parade = %v, want %v (AE reverted the addition)", got, expect)
	}
}

func TestAddEffectAutoParade_AEShipGate_AE2020(t *testing.T) {
	runAddEffectAutoParadeGate(t, ae2020(), "autoparade-AE2020", aep.TargetAE2020)
}
func TestAddEffectAutoParade_AEShipGate_AE2025(t *testing.T) {
	runAddEffectAutoParadeGate(t, ae2025(), "autoparade-AE2025", aep.TargetAE2025)
}
