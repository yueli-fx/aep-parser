// internal/aep/add_effect_wave5_shipgate_test.go
//
// AE ship gate for the effect-library wave 5 (38 MG-common parameter-only
// built-ins: distort / stylize / perspective / color-correction / blur /
// channel / generate / time / matte). On a 100% Go-built file (NewProject →
// NewComposition → NewShapeLayer → Reopen → AddEffect ×38), AE must open
// without corruption and read back all 38 spliced effects in order, then keep
// them across its own resave. One AE run per version (vs the per-effect
// baseline+1 in add_effect_shipgate_test.go) — sufficient acceptance proof for
// splice-portable built-ins, and far cheaper than 38 separate launches.
//
// Gated by AE_SHIP_GATE. Reuses test_data/verify_property_struct.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// wave5Effects is in parade (insertion) order — the order verify_property_struct.jsx
// reads back and the gate asserts.
var wave5Effects = []string{
	aep.EffectTwirl,
	aep.EffectPolarCoordinates,
	aep.EffectSpherize,
	aep.EffectMagnify,
	aep.EffectRipple,
	aep.EffectOpticsCompensation,
	aep.EffectPosterize,
	aep.EffectThreshold,
	aep.EffectFindEdges,
	aep.EffectColorEmboss,
	aep.EffectEmboss,
	aep.EffectStrobeLight,
	aep.EffectBrushStrokes,
	aep.EffectBevelAlpha,
	aep.EffectBevelEdges,
	aep.EffectPhotoFilter,
	aep.EffectVibrance,
	aep.EffectColorBalance,
	aep.EffectColorBalanceHLS,
	aep.EffectBlackAndWhite,
	aep.EffectGammaPedestalGain,
	aep.EffectChannelBlur,
	aep.EffectBilateralBlur,
	aep.EffectSmartBlur,
	aep.EffectUnsharpMask,
	aep.EffectShiftChannels,
	aep.EffectSolidComposite,
	aep.EffectMinimax,
	aep.EffectArithmetic,
	aep.EffectCircle,
	aep.EffectLensFlare,
	aep.EffectCellPattern,
	aep.EffectAdvancedLightning,
	aep.EffectBeam,
	aep.EffectPaintBucket,
	aep.EffectPosterizeTime,
	aep.EffectSimpleChoker,
	aep.EffectMatteChoker,
}

func runAddEffectWave5Gate(t *testing.T, aeExe, label string, target aep.AETarget) {
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
	l := rp.Compositions[0].LayerByName("S")
	if l == nil {
		t.Fatal("reopened: layer S not found")
	}
	for _, fxName := range wave5Effects {
		if _, err := aep.AddEffect(l, fxName); err != nil {
			t.Fatalf("AddEffect(%s): %v", fxName, err)
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "wave5_in.aep")
	resavedAEP := filepath.Join(tempDir, "wave5_resaved.aep")
	doneFile := filepath.Join(tempDir, "wave5.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(wave5Effects))
	for i, e := range wave5Effects {
		quoted[i] = fmt.Sprintf("%q", e)
	}
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expect":[%s]}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), strings.Join(quoted, ","))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 360)

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
		t.Fatal("resaved: no layer with effects (AE dropped the wave-5 effects)")
	}
	if got := paradeChildNames(rl); !eq(got, wave5Effects) {
		t.Errorf("resaved parade = %v, want %v (AE reverted some additions)", got, wave5Effects)
	}
}

func TestAddEffectWave5_AEShipGate_AE2020(t *testing.T) {
	runAddEffectWave5Gate(t, ae2020(), "wave5-AE2020", aep.TargetAE2020)
}
func TestAddEffectWave5_AEShipGate_AE2025(t *testing.T) {
	runAddEffectWave5Gate(t, ae2025(), "wave5-AE2025", aep.TargetAE2025)
}
