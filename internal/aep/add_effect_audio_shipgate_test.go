// internal/aep/add_effect_audio_shipgate_test.go
//
// AE ship gate for the wave-10 AUDIO-processing effect templates (Backwards,
// Bass & Treble, Delay, Flange & Chorus, High-Low Pass, Modulator, Parametric
// EQ, Reverb, Stereo Mixer, Tone). Unlike every other effect these can only be
// applied to a layer that HAS audio, so the gate starts from a base fixture with
// an imported-mp3 audio layer (test_data/fixtures/re_audio_base.aep, authored by
// re_audio_base.jsx) and Go-adds all ten audio effects to it, then has AE open
// the mutated file and read back the parade. Audio effects are NON-VISUAL — there
// are no render pixels to sample — so the acceptance proof is: AE opens without
// corruption, reads back all ten effect match-names in order (DOM readback), and
// keeps them across its own resave. (delivery-contract: ae-accept is the ceiling
// for a non-visual domain.)
//
// Reuses verify_property_struct.jsx (finds the layer carrying a non-empty Effect
// Parade, compares match-names to expect). Skips when the base fixture is absent
// (gitignored — it embeds an mp3) or AE_SHIP_GATE is unset.
//
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

// audioEffectSample is the full wave-10 audio template set, in apply order.
var audioEffectSample = []string{
	aep.EffectAudioBackwards,
	aep.EffectAudioBassTreble,
	aep.EffectAudioDelay,
	aep.EffectAudioFlangeChorus,
	aep.EffectAudioHighLowPass,
	aep.EffectAudioModulator,
	aep.EffectAudioParametricEQ,
	aep.EffectAudioReverb,
	aep.EffectAudioStereoMixer,
	aep.EffectAudioTone,
}

const audioBaseFixture = `e:/projects/tools/aep-parser/test_data/fixtures/re_audio_base.aep`

func runAudioEffectGate(t *testing.T, aeExe, ver string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	if _, err := os.Stat(audioBaseFixture); err != nil {
		t.Skipf("audio base fixture missing (%s) — author it with re_audio_base.jsx", audioBaseFixture)
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/property_struct_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_property_struct.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	rp, err := aep.Open(audioBaseFixture)
	if err != nil {
		t.Fatalf("Open base fixture: %v", err)
	}
	if len(rp.Compositions) == 0 || len(rp.Compositions[0].Layers) == 0 {
		t.Fatal("base fixture has no audio layer")
	}
	aud := rp.Compositions[0].Layers[0]
	if aud.EffectsParade() != nil {
		t.Fatal("base fixture audio layer already has effects; gate would not start clean")
	}
	for _, fxName := range audioEffectSample {
		if _, err := aep.AddEffect(aud, fxName); err != nil {
			t.Fatalf("AddEffect %q: %v", fxName, err)
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "audfx_in.aep")
	resavedAEP := filepath.Join(tempDir, "audfx_resaved.aep")
	doneFile := filepath.Join(tempDir, "audfx.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(audioEffectSample))
	for i, e := range audioEffectSample {
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
	t.Logf("audio-effects %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("audio-effects %s ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("resaved: no layer with effects (AE dropped all audio effects)")
	}
	if got := paradeChildNames(rl); !eq(got, audioEffectSample) {
		t.Errorf("resaved parade = %v, want %v (AE reverted some audio effects)", got, audioEffectSample)
	}
}

func TestAddEffectAudio_AEShipGate_AE2020(t *testing.T) { runAudioEffectGate(t, ae2020(), "AE2020") }
func TestAddEffectAudio_AEShipGate_AE2025(t *testing.T) { runAudioEffectGate(t, ae2025(), "AE2025") }
