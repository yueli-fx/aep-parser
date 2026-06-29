// internal/aep/track_matte_shipgate_test.go
//
// AE ship gate for the track-matte family. Two paths, version-split because the
// explicit-source matte (ldta @0xA0) is an AE23+ feature absent in AE2020/2022:
//
//   - CLASSIC (SetTrackMatte, mode byte @0x6B) → AE2020 + AE2025. Matte = the
//     layer immediately above; the setter only flips the mode byte.
//   - EXPLICIT (SetTrackMatteSource / SetTrackMatteLayer / ClearTrackMatteLayer
//     / RemoveTrackMatte, @0xA0 source id) → AE2024 + AE2025 (AE2020 lacks the
//     slot).
//
// Carriers are two solids each ("M" matte above "T" target). Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// buildTrackMatteClassic: comp with matte "M" above target "T"; SetTrackMatte
// flips T's mode byte to Alpha. NewSolidLayer appends, so create M first (top).
func buildTrackMatteClassic(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MATTE", 640, 360, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "M", 200, 200, [3]float64{1, 1, 1}); err != nil {
		t.Fatalf("NewSolidLayer M: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "T", 640, 360, [3]float64{0.2, 0.5, 0.9}); err != nil {
		t.Fatalf("NewSolidLayer T: %v", err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.Compositions[0]
	if cc.Layers[0].Name != "M" || cc.Layers[1].Name != "T" {
		t.Fatalf("layer order = [%s,%s], want [M,T] (matte above target)", cc.Layers[0].Name, cc.Layers[1].Name)
	}
	if err := cc.LayerByName("T").SetTrackMatte(aep.TrackMatteAlpha); err != nil {
		t.Fatalf("SetTrackMatte: %v", err)
	}
	return rp
}

// buildTrackMatteExplicit: AE23+ explicit-source matte (ldta @0xA0). Four
// matte/target pairs cover all 4 setters: A=SetTrackMatteSource, B=
// SetTrackMatteLayer, C=ClearTrackMatteLayer, D=RemoveTrackMatte.
func buildTrackMatteExplicit(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "EMATTE", 640, 360, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	for _, n := range []string{"MA", "TA", "MB", "TB", "MC", "TC", "MD", "TD"} {
		if _, err := aep.NewSolidLayer(comp, n, 200, 200, [3]float64{0.8, 0.8, 0.8}); err != nil {
			t.Fatalf("NewSolidLayer %s: %v", n, err)
		}
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.Compositions[0]
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	lyr := func(name string) *aep.Layer {
		l := cc.LayerByName(name)
		if l == nil {
			t.Fatalf("layer %s missing", name)
		}
		return l
	}

	// A: SetTrackMatteSource → LUMA, source = MA.
	must("SetTrackMatteSource", lyr("TA").SetTrackMatteSource(lyr("MA"), aep.TrackMatteLuma))
	// B: SetTrackMatteLayer(id) → ALPHA, source = MB.
	must("SetTrackMatteLayer", lyr("TB").SetTrackMatteLayer(lyr("MB").ID, aep.TrackMatteAlpha))
	// C: set then ClearTrackMatteLayer → NONE.
	must("TC set", lyr("TC").SetTrackMatteSource(lyr("MC"), aep.TrackMatteLuma))
	must("ClearTrackMatteLayer", lyr("TC").ClearTrackMatteLayer())
	// D: set then RemoveTrackMatte → NONE.
	must("TD set", lyr("TD").SetTrackMatteSource(lyr("MD"), aep.TrackMatteLuma))
	must("RemoveTrackMatte", lyr("TD").RemoveTrackMatte())

	return rp
}

func runTrackMatteGate(t *testing.T, aeExe, ver, kind, jsxName string, build func(*testing.T, aep.AETarget) *aep.Project, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	argsPath := `e:/projects/tools/aep-parser/test_data/` + kind + `_args.json`
	jsxPath := `E:/projects/tools/aep-parser/test_data/` + jsxName
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := build(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, kind+"_in.aep")
	resavedAEP := filepath.Join(tempDir, kind+"_resaved.aep")
	doneFile := filepath.Join(tempDir, kind+".done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
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
	t.Logf("%s %s AE readback:\n%s", kind, ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s %s ship gate FAIL:\n%s", kind, ver, body)
	}
}

func TestTrackMatteClassic_AEShipGate_AE2020(t *testing.T) {
	runTrackMatteGate(t, ae2020(), "AE2020", "track_matte_classic", "verify_track_matte_classic.jsx", buildTrackMatteClassic, aep.TargetAE2020)
}
func TestTrackMatteClassic_AEShipGate_AE2025(t *testing.T) {
	runTrackMatteGate(t, ae2025(), "AE2025", "track_matte_classic", "verify_track_matte_classic.jsx", buildTrackMatteClassic, aep.TargetAE2025)
}

// Explicit-source matte is AE23+ (ldta @0xA0). Single-version AE2025 only:
// the @0xA0 slot is produced only by TargetAE2025, whose fingerprint AE2024
// refuses to open (forward-compat: "文件使用 25.1 ... 无法打开"), and there is no
// intermediate ≥AE23 target. All 4 explicit setters are DOM-verified here; the
// caps stay roundtrip per the no-single-version-ae-accept convention.
func TestTrackMatteExplicit_AEShipGate_AE2025(t *testing.T) {
	runTrackMatteGate(t, ae2025(), "AE2025", "track_matte_explicit", "verify_track_matte_explicit.jsx", buildTrackMatteExplicit, aep.TargetAE2025)
}
