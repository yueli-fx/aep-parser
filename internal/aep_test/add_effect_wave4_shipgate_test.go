// internal/aep/add_effect_wave4_shipgate_test.go
//
// AE ship gate for the effect-library wave 4 (MG distort/generate/stylize/
// transition). On a 100% Go-built file (NewProject → NewComposition →
// NewShapeLayer → Reopen → AddEffect ×10), AE must open without corruption and
// read back all ten spliced effects in order, then keep them across its own
// resave. One AE run per version (vs the per-effect baseline+1 in
// add_effect_shipgate_test.go) — sufficient acceptance proof for splice-portable
// built-ins, and far cheaper than 10 separate launches.
//
// Gated by AE_SHIP_GATE. Reuses test_data/generators/verify_property_struct.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

var wave4Effects = []string{
	aep.EffectTurbulentDisplace,
	aep.EffectRoughenEdges,
	aep.EffectEcho,
	aep.EffectRadialBlur,
	aep.EffectFourColorGradient,
	aep.EffectCheckerboard,
	aep.EffectGrid,
	aep.EffectStroke,
	aep.EffectCornerPin,
	aep.EffectVenetianBlinds,
}

func runAddEffectWave4Gate(t *testing.T, aeExe, label string, target aep.AETarget) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/property_struct_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_property_struct.jsx`
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
	for _, fxName := range wave4Effects {
		if _, err := aep.AddEffect(l, fxName); err != nil {
			t.Fatalf("AddEffect(%s): %v", fxName, err)
		}
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "wave4_in.aep")
	resavedAEP := filepath.Join(tempDir, "wave4_resaved.aep")
	doneFile := filepath.Join(tempDir, "wave4.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	quoted := make([]string, len(wave4Effects))
	for i, e := range wave4Effects {
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
		t.Fatal("resaved: no layer with effects (AE dropped the wave-4 effects)")
	}
	if got := paradeChildNames(rl); !eq(got, wave4Effects) {
		t.Errorf("resaved parade = %v, want %v (AE reverted some additions)", got, wave4Effects)
	}
}

func TestAddEffectWave4_AEShipGate_AE2020(t *testing.T) {
	runAddEffectWave4Gate(t, ae2020(), "wave4-AE2020", aep.TargetAE2020)
}
func TestAddEffectWave4_AEShipGate_AE2025(t *testing.T) {
	runAddEffectWave4Gate(t, ae2025(), "wave4-AE2025", aep.TargetAE2025)
}
