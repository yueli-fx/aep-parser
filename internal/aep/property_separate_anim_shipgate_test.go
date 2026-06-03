// internal/aep/property_separate_anim_shipgate_test.go
//
// AE ship gate for Property.SetDimensionsSeparated on an ANIMATED 3D Position
// (both directions). Opens an AE-2020-native fixture, toggles Separate
// Dimensions, WriteAEP, and has AE open the mutated file: proves AE ACCEPTS the
// restructured keyframe streams (no data-loss / corrupt) and reads back the
// expected per-axis / merged keyframe values. Resaves so the Go side confirms
// AE kept the change.
//
// Pairs (× AE 2020 + AE 2025):
//   - separate animated (re_sepdim_anim_before.aep, merged animated 3D)
//   - merge    animated (re_sepdim_anim_after.aep,  separated animated 3D)
//
// Gated by AE_SHIP_GATE. Fixtures built by test_data/re_separate_dims_anim.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// runSepDimAnimShipGate toggles Separate Dimensions on baseFixture's animated
// Position (separate=true→split, false→merge), then verifies AE accepts + reads
// back the per-axis / merged keyframe values.
func runSepDimAnimShipGate(t *testing.T, aeExe, baseFixture string, separate bool) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/sepdim_anim_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_sepdim_anim.jsx`
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

	// Expected per-axis keyframe tables the way AE will read them after toggle.
	var times, vx, vy, vz []float64
	if separate {
		// Source leader carries the 3D keyframes → per-axis components.
		for _, kf := range pos.Keyframes {
			v := kf.Value.([]float64)
			times = append(times, kf.Time)
			vx = append(vx, v[0])
			vy = append(vy, v[1])
			vz = append(vz, v[2])
		}
	} else {
		// Source followers carry the per-axis keyframes → merged leader value.
		f0 := findProp(proj, aep.MatchNamePosition0)
		f1 := findProp(proj, aep.MatchNamePosition1)
		f2 := findProp(proj, aep.MatchNamePosition2)
		if f0 == nil || f1 == nil || f2 == nil {
			t.Fatalf("%s: separated followers missing", baseFixture)
		}
		for i := range f0.Keyframes {
			times = append(times, f0.Keyframes[i].Time)
			vx = append(vx, f0.Keyframes[i].Value.(float64))
			vy = append(vy, f1.Keyframes[i].Value.(float64))
			vz = append(vz, f2.Keyframes[i].Value.(float64))
		}
	}

	if err := pos.SetDimensionsSeparated(separate); err != nil {
		t.Fatalf("SetDimensionsSeparated(%v): %v", separate, err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "sepdim_anim_in.aep")
	resavedAEP := filepath.Join(tempDir, "sepdim_anim_resaved.aep")
	doneFile := filepath.Join(tempDir, "sepdim_anim.done")

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
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"mode":%q,"n":%d,"times":%s,"vx":%s,"vy":%s,"vz":%s}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), mode, len(times),
		jsonFloats(times), jsonFloats(vx), jsonFloats(vy), jsonFloats(vz))
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
	t.Logf("SetDimensionsSeparated(%v) animated AE readback:\n%s", separate, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("animated separate-dims ship gate FAIL:\n%s", body)
	}

	// Preservation proof: AE's own resave kept the toggle + animation.
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
	if separate {
		if rZ := findProp(re, aep.MatchNamePosition2); rZ == nil || !rZ.IsAnimated() {
			t.Errorf("resaved: animated Position_2 follower missing/not-animated")
		}
	} else if !rLeader.IsAnimated() {
		t.Errorf("resaved: merged leader not animated")
	}
}

func jsonFloats(fs []float64) string {
	parts := make([]string, len(fs))
	for i, f := range fs {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func TestSeparateDimsAnim_AEShipGate_AE2020(t *testing.T) {
	runSepDimAnimShipGate(t, ae2020(), "../../test_data/re_sepdim_anim_before.aep", true)
}
func TestSeparateDimsAnim_AEShipGate_AE2025(t *testing.T) {
	runSepDimAnimShipGate(t, ae2025(), "../../test_data/re_sepdim_anim_before.aep", true)
}

func TestMergeDimsAnim_AEShipGate_AE2020(t *testing.T) {
	runSepDimAnimShipGate(t, ae2020(), "../../test_data/re_sepdim_anim_after.aep", false)
}
func TestMergeDimsAnim_AEShipGate_AE2025(t *testing.T) {
	runSepDimAnimShipGate(t, ae2025(), "../../test_data/re_sepdim_anim_after.aep", false)
}
