// internal/aep/add_effect_wave7_shipgate_test.go
//
// AE ship gate for the effect-library wave 7 (Lumetri Color + Lightning + 43
// Cycore "CC" effects). On a 100% Go-built file (NewProject → NewComposition →
// NewShapeLayer → Reopen → AddEffect ×45), AE must open without corruption and
// read back all 45 spliced effects in order, then keep them across its own
// resave. One AE run per version.
//
// NOTE 4 CC effects store a "CS …" internal match-name (CrossBlur / Threads /
// HexTile / Vignette) — the const values use the STORED name, which is what AE
// reads back, so paradeChildNames matches.
//
// Gated by AE_SHIP_GATE. Reuses test_data/verify_property_struct.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// wave7Effects is in parade (insertion) order.
var wave7Effects = []string{
	aep.EffectLumetri,
	aep.EffectLightning,
	aep.EffectCCRadialFastBlur,
	aep.EffectCCRadialBlur,
	aep.EffectCCCrossBlur,
	aep.EffectCCBendIt,
	aep.EffectCCBender,
	aep.EffectCCBlobbylize,
	aep.EffectCCFloMotion,
	aep.EffectCCGriddler,
	aep.EffectCCLens,
	aep.EffectCCPageTurn,
	aep.EffectCCPowerPin,
	aep.EffectCCRipplePulse,
	aep.EffectCCSlant,
	aep.EffectCCSmear,
	aep.EffectCCSplit,
	aep.EffectCCSplit2,
	aep.EffectCCTiler,
	aep.EffectCCWarpoMatic,
	aep.EffectCCLightBurst,
	aep.EffectCCLightRays,
	aep.EffectCCLightSweep,
	aep.EffectCCThreads,
	aep.EffectCCCylinder,
	aep.EffectCCSphere,
	aep.EffectCCSpotlight,
	aep.EffectCCGlass,
	aep.EffectCCHexTile,
	aep.EffectCCKaleida,
	aep.EffectCCMrSmoothie,
	aep.EffectCCPlastic,
	aep.EffectCCRepeTile,
	aep.EffectCCThreshold,
	aep.EffectCCThresholdRGB,
	aep.EffectCCPixelPolly,
	aep.EffectCCScatterize,
	aep.EffectCCStarBurst,
	aep.EffectCCForceMotionBlur,
	aep.EffectCCWideTime,
	aep.EffectCCColorOffset,
	aep.EffectCCToner,
	aep.EffectCCBurnFilm,
	aep.EffectCCVignette,
	aep.EffectCCSimpleWireRemoval,
}

func runAddEffectWave7Gate(t *testing.T, aeExe, label string, target aep.AETarget) {
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
	for _, fxName := range wave7Effects {
		if _, err := aep.AddEffect(l, fxName); err != nil {
			t.Fatalf("AddEffect(%s): %v", fxName, err)
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "wave7_in.aep")
	resavedAEP := filepath.Join(tempDir, "wave7_resaved.aep")
	doneFile := filepath.Join(tempDir, "wave7.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(wave7Effects))
	for i, e := range wave7Effects {
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
		t.Fatal("resaved: no layer with effects (AE dropped the wave-7 effects)")
	}
	if got := paradeChildNames(rl); !eq(got, wave7Effects) {
		t.Errorf("resaved parade = %v, want %v (AE reverted some additions)", got, wave7Effects)
	}
}

func TestAddEffectWave7_AEShipGate_AE2020(t *testing.T) {
	runAddEffectWave7Gate(t, ae2020(), "wave7-AE2020", aep.TargetAE2020)
}
func TestAddEffectWave7_AEShipGate_AE2025(t *testing.T) {
	runAddEffectWave7Gate(t, ae2025(), "wave7-AE2025", aep.TargetAE2025)
}
