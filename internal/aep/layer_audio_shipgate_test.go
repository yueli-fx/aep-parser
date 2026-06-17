// internal/aep/layer_audio_shipgate_test.go
//
// AE ship gate for the layer-set AUDIO setters, capped at verify=roundtrip until
// now ("无音轨属性 → 报错; 无专门 AE gate"):
//
//   - SetAudioEnabled  (ldta audio switch bit @0x27)
//   - SetAudioLevels   (Audio Levels property, [left,right] dB)
//
// Audio is NON-VISUAL — no render pixels to sample — so the acceptance proof is
// AE opening the mutated file and DOM-reading both values back (delivery-contract:
// ae-accept is the ceiling for a non-visual domain). Both setters mutate an
// EXISTING property/bit, and a default audio layer ELIDES Audio Levels
// (default-omission), so the gate starts from a carrier fixture
// (test_data/re_audio_levels.aep, authored by build_re_audio_levels.jsx with
// AE2020) whose audio layer already has Audio Levels materialized at a NON-default
// placeholder [3,3] — distinct from the gate's [-8,-8] target so a no-op mutate
// would be caught. The carrier embeds an mp3, so it is gitignored; the gate skips
// when it (or AE_SHIP_GATE) is absent. No resave in the verify JSX (idta-batch
// lesson). Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

const audioLevelsFixture = `e:/projects/tools/aep-parser/test_data/re_audio_levels.aep`

func runLayerAudioGate(t *testing.T, aeExe, ver string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	if _, err := os.Stat(audioLevelsFixture); err != nil {
		t.Skipf("audio-levels carrier missing (%s) — author it with build_re_audio_levels.jsx", audioLevelsFixture)
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/audio_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_audio.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	rp, err := aep.Open(audioLevelsFixture)
	if err != nil {
		t.Fatalf("Open carrier: %v", err)
	}
	if len(rp.Compositions) == 0 || len(rp.Compositions[0].Layers) == 0 {
		t.Fatal("carrier has no audio layer")
	}
	aud := rp.Compositions[0].Layers[0]
	if aud.AudioLevels() == nil {
		t.Fatal("carrier audio layer has no Audio Levels property (default-omission not materialized)")
	}
	if err := aud.SetAudioEnabled(false); err != nil {
		t.Fatalf("SetAudioEnabled: %v", err)
	}
	if err := aud.SetAudioLevels([]float64{-8, -8}); err != nil {
		t.Fatalf("SetAudioLevels: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "audio_in.aep")
	doneFile := filepath.Join(tempDir, "audio.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q}`, toFwd(inputAEP), toFwd(doneFile))
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
	t.Logf("layer audio %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("layer audio %s ship gate FAIL:\n%s", ver, body)
	}
}

func TestLayerAudio_AEShipGate_AE2020(t *testing.T) { runLayerAudioGate(t, ae2020(), "AE2020") }
func TestLayerAudio_AEShipGate_AE2025(t *testing.T) { runLayerAudioGate(t, ae2025(), "AE2025") }
