// internal/aep/property_structural_shipgate_test.go
//
// AE ship gate for PropertyBase structural ops (Remove / MoveTo) on the Effect
// Parade. Loads the AE-2020-native baseline (3 effects), applies the Go
// mutation, WriteAEP, and has AE open the mutated file: proves AE ACCEPTS the
// restructured Effect Parade (no data-loss / corrupt) and reads back the
// expected effect order. Resaves so the Go side confirms AE kept the change.
//
// Pairs (× AE 2020 + AE 2025):
//   - remove middle  (Tint)  -> [Gaussian Blur, Fill]
//   - move last to front (Fill -> 0) -> [Fill, Gaussian Blur, Tint]
//
// Gated by AE_SHIP_GATE. Baseline built by test_data/generators/re_property_struct.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runPropStructShipGate(t *testing.T, aeExe, label string, mutate func(*aep.Layer) error, expect []string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/property_struct_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_property_struct.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("re_property_struct_baseline.aep not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	if err := mutate(l); err != nil {
		t.Fatalf("%s: mutate: %v", label, err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "propstruct_in.aep")
	resavedAEP := filepath.Join(tempDir, "propstruct_resaved.aep")
	doneFile := filepath.Join(tempDir, "propstruct.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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

	// Preservation proof: AE's own resave kept the new effect order.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rl := layerWithEffects(re)
	if rl == nil {
		t.Fatal("resaved: no layer with effects")
	}
	if got := paradeChildNames(rl); !eq(got, expect) {
		t.Errorf("resaved parade = %v, want %v (AE reverted the mutation)", got, expect)
	}
}

func removeMiddleEffect(l *aep.Layer) error {
	return aep.RemovePropertyGroup(l.EffectsParade().ChildByIndex(1).(*aep.AEPropertyGroup))
}

func moveLastEffectToFront(l *aep.Layer) error {
	return aep.MovePropertyGroup(l.EffectsParade().ChildByIndex(2).(*aep.AEPropertyGroup), 0)
}

func TestPropStructRemove_AEShipGate_AE2020(t *testing.T) {
	runPropStructShipGate(t, ae2020(), "remove-AE2020", removeMiddleEffect,
		[]string{"ADBE Gaussian Blur 2", "ADBE Fill"})
}
func TestPropStructRemove_AEShipGate_AE2025(t *testing.T) {
	runPropStructShipGate(t, ae2025(), "remove-AE2025", removeMiddleEffect,
		[]string{"ADBE Gaussian Blur 2", "ADBE Fill"})
}

func TestPropStructMove_AEShipGate_AE2020(t *testing.T) {
	runPropStructShipGate(t, ae2020(), "move-AE2020", moveLastEffectToFront,
		[]string{"ADBE Fill", "ADBE Gaussian Blur 2", "ADBE Tint"})
}
func TestPropStructMove_AEShipGate_AE2025(t *testing.T) {
	runPropStructShipGate(t, ae2025(), "move-AE2025", moveLastEffectToFront,
		[]string{"ADBE Fill", "ADBE Gaussian Blur 2", "ADBE Tint"})
}

func duplicateFirstEffect(l *aep.Layer) error {
	_, err := aep.DuplicatePropertyGroup(l.EffectsParade().ChildByIndex(0).(*aep.AEPropertyGroup))
	return err
}

func TestPropStructDuplicate_AEShipGate_AE2020(t *testing.T) {
	runPropStructShipGate(t, ae2020(), "duplicate-AE2020", duplicateFirstEffect,
		[]string{"ADBE Gaussian Blur 2", "ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"})
}
func TestPropStructDuplicate_AEShipGate_AE2025(t *testing.T) {
	runPropStructShipGate(t, ae2025(), "duplicate-AE2025", duplicateFirstEffect,
		[]string{"ADBE Gaussian Blur 2", "ADBE Gaussian Blur 2", "ADBE Tint", "ADBE Fill"})
}
