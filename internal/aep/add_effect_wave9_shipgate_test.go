// internal/aep/add_effect_wave9_shipgate_test.go
//
// AE ship gate for effect-library wave 9 (43 more built-ins from the big probe
// sweep: keying / simulation / utility / time / transition / more Cycore CC).
// 100% Go-built file, AddEffect ×43, AE opens clean + reads back all 43 in order
// + keeps them across resave. One AE run per version.
//
// NOTE several CC effects store a "CS …" internal match-name (BlockLoad /
// Color Neutralizer / Kernel / LineSweep / Rainfall / Snowfall) — consts use the
// STORED name (what AE reads back), so paradeChildNames matches.
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

var wave9Effects = []string{
	aep.EffectBasic3D, aep.EffectBroadcastColors, aep.EffectChannelCombiner,
	aep.EffectCineonConverter, aep.EffectColorKey, aep.EffectColorRange,
	aep.EffectExtract, aep.EffectGeometryLegacy, aep.EffectGradientWipe,
	aep.EffectGrowBounds, aep.EffectKeyCleaner, aep.EffectLayerControl,
	aep.EffectLumaKey, aep.EffectMedian, aep.EffectNoiseHLSAuto,
	aep.EffectColorProfileConverter, aep.EffectTimeDisplacement, aep.EffectTimecode,
	aep.EffectCCBallAction, aep.EffectCCBubbles, aep.EffectCCComposite,
	aep.EffectCCDrizzle, aep.EffectCCEnvironment, aep.EffectCCGlassWipe,
	aep.EffectCCGlueGun, aep.EffectCCGridWipe, aep.EffectCCHair,
	aep.EffectCCImageWipe, aep.EffectCCJaws, aep.EffectCCLightWipe,
	aep.EffectCCMrMercury, aep.EffectCCParticleSystemsII, aep.EffectCCRadialScaleWipe,
	aep.EffectCCRain, aep.EffectCCScaleWipe, aep.EffectCCSnow,
	aep.EffectCCTwister, aep.EffectCCBlockLoad, aep.EffectCCColorNeutralizer,
	aep.EffectCCKernel, aep.EffectCCLineSweep, aep.EffectCCRainfall,
	aep.EffectCCSnowfall,
}

func runAddEffectWave9Gate(t *testing.T, aeExe, label string, target aep.AETarget) {
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
	for _, fxName := range wave9Effects {
		if _, err := aep.AddEffect(l, fxName); err != nil {
			t.Fatalf("AddEffect(%s): %v", fxName, err)
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "wave9_in.aep")
	resavedAEP := filepath.Join(tempDir, "wave9_resaved.aep")
	doneFile := filepath.Join(tempDir, "wave9.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(wave9Effects))
	for i, e := range wave9Effects {
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
		t.Fatal("resaved: no layer with effects (AE dropped the wave-9 effects)")
	}
	if got := paradeChildNames(rl); !eq(got, wave9Effects) {
		t.Errorf("resaved parade = %v, want %v (AE reverted some additions)", got, wave9Effects)
	}
}

func TestAddEffectWave9_AEShipGate_AE2020(t *testing.T) {
	runAddEffectWave9Gate(t, ae2020(), "wave9-AE2020", aep.TargetAE2020)
}
func TestAddEffectWave9_AEShipGate_AE2025(t *testing.T) {
	runAddEffectWave9Gate(t, ae2025(), "wave9-AE2025", aep.TargetAE2025)
}
