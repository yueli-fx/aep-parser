// internal/aep/mg_effect_layerref_wave11_shipgate_test.go
//
// AE ship gate for the wave-11 LAYER-REFERENCE effect templates (Warp Stabilizer
// = ADBE SubspaceStabilizer, 3D Glasses, Timewarp, CC Particle World). These are
// the 4 foreign-tdpi effects wave 9 surfaced but deferred; their templates were
// materialized (re_effect_layerref2.aep) so each carries its layer-ref param(s)
// with a tdpi, exactly like the wave-8 Displacement Map family.
//
// On a 100% Go-built file: HOST + MAP solids, AddEffect all four on HOST, then
// SetEffectLayerParam every layer-ref param at MAP (3D Glasses / Timewarp expose
// TWO each). AE must open without corruption, read back all four parade
// match-names in order (DOM readback), and keep them across its own resave.
//
// These effects' visual "action surface" has no clean single-frame pixel proof
// (Warp Stabilizer = analysis-driven, Timewarp = time remap, CC Particle World =
// procedural, 3D Glasses = stereo composite) — same honest render-pixel deferral
// as wave-8 CC Vector Blur. The acceptance proof is ae-accept + readback + resave
// (delivery-contract: the layer-ref binding resolves and survives).
//
// Gated by AE_SHIP_GATE. Reuses verify_property_struct.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// wave11Effects in apply order; wave11LayerParams maps each to its layer-ref
// param match-name(s) (pointed at MAP).
var wave11Effects = []string{
	aep.EffectWarpStabilizer,
	aep.Effect3DGlasses,
	aep.EffectTimewarp,
	aep.EffectCCParticleWorld,
}

var wave11LayerParams = map[string][]string{
	aep.EffectWarpStabilizer:  {aep.EffectWarpStabilizerRefLayer},
	aep.Effect3DGlasses:       {aep.Effect3DGlassesLeftView, aep.Effect3DGlassesRightView},
	aep.EffectTimewarp:        {aep.EffectTimewarpMatteLayer, aep.EffectTimewarpSourceLayer},
	aep.EffectCCParticleWorld: {aep.EffectCCParticleWorldTexture},
}

func runWave11LayerRefGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/property_struct_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_property_struct.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "WAVE11", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "HOST", 1920, 1080, [3]float64{0.5, 0.5, 0.5}); err != nil {
		t.Fatalf("NewSolidLayer HOST: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "MAP", 1920, 1080, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer MAP: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]
	hl := c.LayerByName("HOST")
	ml := c.LayerByName("MAP")
	if hl == nil || ml == nil {
		t.Fatal("reopened: HOST/MAP missing")
	}
	for _, fxName := range wave11Effects {
		fx, err := aep.AddEffect(hl, fxName)
		if err != nil {
			t.Fatalf("AddEffect %q: %v", fxName, err)
		}
		for _, pm := range wave11LayerParams[fxName] {
			if err := aep.SetEffectLayerParam(hl, fx, pm, ml); err != nil {
				t.Fatalf("SetEffectLayerParam %q on %q: %v", pm, fxName, err)
			}
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "wave11_in.aep")
	resavedAEP := filepath.Join(tempDir, "wave11_resaved.aep")
	doneFile := filepath.Join(tempDir, "wave11.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(wave11Effects))
	for i, e := range wave11Effects {
		quoted[i] = fmt.Sprintf("%q", e)
	}
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expect":[%s]}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), strings.Join(quoted, ","))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("wave11 layer-ref %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("wave11 layer-ref %s ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("resaved: no layer with effects (AE dropped all wave-11 effects)")
	}
	if got := paradeChildNames(rl); !eq(got, wave11Effects) {
		t.Errorf("resaved parade = %v, want %v (AE reverted some wave-11 effects)", got, wave11Effects)
	}
}

func TestAddEffectWave11_AEShipGate_AE2020(t *testing.T) {
	runWave11LayerRefGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestAddEffectWave11_AEShipGate_AE2025(t *testing.T) {
	runWave11LayerRefGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
