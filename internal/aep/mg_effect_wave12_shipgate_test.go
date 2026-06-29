// internal/aep/mg_effect_wave12_shipgate_test.go
//
// AE ship gate for the wave-12 PARKED classic effects (Bezier Warp/Mesh Warp/
// Channel Mixer/Reshape/Vector Paint parameter-only + Texturize/Color Link/
// Compound Arithmetic/Set Channels layer-ref). These are the wave-6 effects whose
// camelCase match-names failed; probe11 found the correct ALL-CAPS names. Same
// flow as wave 11: AddEffect all nine on a HOST solid, point every layer-ref param
// at MAP (Set Channels has FOUR), then AE must open clean, read back all nine
// parade names in order, and keep them across resave. Non-render (these are
// distortion/channel effects without a clean single-frame layer-ref pixel proof):
// the acceptance proof is ae-accept + DOM readback + resave.
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

var wave12Effects = []string{
	aep.EffectBezierWarp, aep.EffectMeshWarp, aep.EffectChannelMixer,
	aep.EffectReshape, aep.EffectVectorPaint,
	aep.EffectTexturize, aep.EffectColorLink, aep.EffectCompoundArithmetic,
	aep.EffectSetChannels,
}

var wave12LayerParams = map[string][]string{
	aep.EffectTexturize:          {aep.EffectTexturizeLayer},
	aep.EffectColorLink:          {aep.EffectColorLinkSourceLayer},
	aep.EffectCompoundArithmetic: {aep.EffectCompoundArithmeticSecondSource},
	aep.EffectSetChannels: {
		aep.EffectSetChannelsSource1, aep.EffectSetChannelsSource2,
		aep.EffectSetChannelsSource3, aep.EffectSetChannelsSource4,
	},
}

func runWave12Gate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/property_struct_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_property_struct.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "WAVE12", 1920, 1080, 30, 5)
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
	for _, fxName := range wave12Effects {
		fx, err := aep.AddEffect(hl, fxName)
		if err != nil {
			t.Fatalf("AddEffect %q: %v", fxName, err)
		}
		for _, pm := range wave12LayerParams[fxName] {
			if err := aep.SetEffectLayerParam(hl, fx, pm, ml); err != nil {
				t.Fatalf("SetEffectLayerParam %q on %q: %v", pm, fxName, err)
			}
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "wave12_in.aep")
	resavedAEP := filepath.Join(tempDir, "wave12_resaved.aep")
	doneFile := filepath.Join(tempDir, "wave12.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(wave12Effects))
	for i, e := range wave12Effects {
		quoted[i] = fmt.Sprintf("%q", e)
	}
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expect":[%s]}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), strings.Join(quoted, ","))
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
	t.Logf("wave12 %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("wave12 %s ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("resaved: no layer with effects (AE dropped all wave-12 effects)")
	}
	if got := paradeChildNames(rl); !eq(got, wave12Effects) {
		t.Errorf("resaved parade = %v, want %v (AE reverted some wave-12 effects)", got, wave12Effects)
	}
}

func TestAddEffectWave12_AEShipGate_AE2020(t *testing.T) {
	runWave12Gate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestAddEffectWave12_AEShipGate_AE2025(t *testing.T) {
	runWave12Gate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
