// internal/aep/property_separate_shipgate_test.go
//
// AE ship gate for Property.SetDimensionsSeparated (both directions). Opens an
// AE-2020-native fixture, toggles Separate Dimensions, WriteAEP, and has AE open
// the mutated file: proves AE ACCEPTS the restructured transform group (no
// data-loss / corrupt) and reads back the expected dimensionsSeparated state +
// per-axis / merged values. Resaves so the Go side confirms AE kept the change.
//
// Three pairs (× AE 2020 + AE 2025):
//   - separate 3D  (re_separate_dims_before.aep, merged 3D)
//   - merge        (re_sepdim_merge_before.aep, separated 3D)
//   - separate 2D  (re_sepdim_2d_before.aep, merged 2D — X/Y only, no Position_2)
//
// Gated by AE_SHIP_GATE. Fixtures built by test_data/re_separate_dims*.jsx.
package aep_test

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// runSepDimShipGate toggles Separate Dimensions on baseFixture's Position
// (separate=true→split, false→merge) and verifies AE accepts + reads back the
// result. is3D drives whether a Position_2 follower is expected post-separate.
func runSepDimShipGate(t *testing.T, aeExe, baseFixture string, separate, is3D bool) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/separate_dims_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_separate_dims.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	proj, err := aep.Open(baseFixture)
	if err != nil {
		t.Skipf("%s not present: %v", baseFixture, err)
	}
	pos := findProp(proj, aep.MatchNamePosition)
	if pos == nil {
		t.Fatalf("%s: Position leader not found", baseFixture)
	}
	if pos.DimensionsSeparated() == separate {
		t.Fatalf("%s: base already in target state (separated=%v)", baseFixture, separate)
	}

	// Determine expected X/Y/Z the way AE will read them after the toggle.
	var wantX, wantY, wantZ float64
	if separate {
		xyz := pos.StaticValue.([]float64)
		wantX, wantY, wantZ = xyz[0], xyz[1], xyz[2]
	} else {
		// Merged value = the per-axis follower values (leader is at default
		// while separated, so it is not the source of truth here).
		wantX, _ = scalar(findProp(proj, aep.MatchNamePosition0))
		wantY, _ = scalar(findProp(proj, aep.MatchNamePosition1))
		if f := findProp(proj, aep.MatchNamePosition2); f != nil {
			wantZ, _ = scalar(f)
		}
	}

	if err := aep.SetDimensionsSeparated(pos, separate); err != nil {
		t.Fatalf("SetDimensionsSeparated(%v): %v", separate, err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "sepdim_in.aep")
	resavedAEP := filepath.Join(tempDir, "sepdim_resaved.aep")
	doneFile := filepath.Join(tempDir, "sepdim.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	mode := "sep"
	if !separate {
		mode = "merge"
	}
	var argsJSON string
	if separate && !is3D {
		// 2D separate: omit z so the verify JSX skips the (synthesized) Position_2.
		argsJSON = fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"mode":%q,"x":%g,"y":%g}`,
			toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), mode, wantX, wantY)
	} else {
		argsJSON = fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"mode":%q,"x":%g,"y":%g,"z":%g}`,
			toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), mode, wantX, wantY, wantZ)
	}
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("SetDimensionsSeparated(%v) AE readback:\n%s", separate, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("separate-dims ship gate FAIL:\n%s", body)
	}

	// Preservation proof: AE's own resave kept the toggle.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rLeader := findProp(re, aep.MatchNamePosition)
	if rLeader == nil {
		t.Fatal("resaved: Position leader missing")
	}
	if rLeader.DimensionsSeparated() != separate {
		t.Fatalf("resaved: leader separated=%v, want %v (AE reverted the toggle)", rLeader.DimensionsSeparated(), separate)
	}
	if separate && is3D {
		rZ := findProp(re, aep.MatchNamePosition2)
		if rZ == nil {
			t.Fatal("resaved: 3D Position_2 follower missing")
		}
		if z, ok := rZ.StaticValue.(float64); !ok || math.Abs(z-wantZ) > 0.001 {
			t.Errorf("resaved: Position_2 = %v, want %g", rZ.StaticValue, wantZ)
		}
	}
	if !separate {
		// AE re-introduces the pre-allocated (zeroed) Position_0/1 on its own
		// resave — that is the canonical merged form, equivalent to our
		// leader-only output. The discriminator that the merge truly took is
		// the absence of the Z follower (Position_2 exists only while separated)
		// plus dimensionsSeparated=false (asserted above).
		if findProp(re, aep.MatchNamePosition2) != nil {
			t.Errorf("resaved: Position_2 still present after merge (AE kept it separated)")
		}
	}
}

func scalar(p *aep.Property) (float64, bool) {
	if p == nil {
		return 0, false
	}
	f, ok := p.StaticValue.(float64)
	return f, ok
}

func ae2020() string {
	if e := os.Getenv("AE2020_EXE"); e != "" {
		return e
	}
	return `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
}

func ae2025() string {
	if e := os.Getenv("AE2025_EXE"); e != "" {
		return e
	}
	return `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
}

func ae2024() string {
	if e := os.Getenv("AE2024_EXE"); e != "" {
		return e
	}
	return `E:/adobe/Adobe After Effects 2024/Support Files/AfterFX.exe`
}

func TestSeparateDims_AEShipGate_AE2020(t *testing.T) {
	runSepDimShipGate(t, ae2020(), "../../test_data/re_separate_dims_before.aep", true, true)
}
func TestSeparateDims_AEShipGate_AE2025(t *testing.T) {
	runSepDimShipGate(t, ae2025(), "../../test_data/re_separate_dims_before.aep", true, true)
}

func TestMergeDims_AEShipGate_AE2020(t *testing.T) {
	runSepDimShipGate(t, ae2020(), "../../test_data/re_sepdim_merge_before.aep", false, true)
}
func TestMergeDims_AEShipGate_AE2025(t *testing.T) {
	runSepDimShipGate(t, ae2025(), "../../test_data/re_sepdim_merge_before.aep", false, true)
}

func TestSeparate2D_AEShipGate_AE2020(t *testing.T) {
	runSepDimShipGate(t, ae2020(), "../../test_data/re_sepdim_2d_before.aep", true, false)
}
func TestSeparate2D_AEShipGate_AE2025(t *testing.T) {
	runSepDimShipGate(t, ae2025(), "../../test_data/re_sepdim_2d_before.aep", true, false)
}
