// internal/aep/property_separate_shipgate_test.go
//
// AE ship gate for Property.SetDimensionsSeparated (merge→separate, static 3D
// Position). Opens an AE-2020-native fixture with a merged 3D Position, flips
// Separate Dimensions via SetDimensionsSeparated, WriteAEP, and has AE open the
// mutated file: proves AE ACCEPTS the restructured transform group (synthesized
// Position_2 + flipped tdsb flags + migrated values; no data-loss / corrupt) and
// reads back dimensionsSeparated=true with the per-axis X/Y/Z. Resaves so the Go
// side confirms AE kept the separation rather than silently merging.
//
// Gated by AE_SHIP_GATE. Base fixture built by test_data/re_separate_dims.jsx
// (re_separate_dims_before.aep = the merged state).
package aep_test

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runSeparateDimsShipGate(t *testing.T, aeExe, baseFixture string) {
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
		t.Fatalf("%s: merged Position leader not found", baseFixture)
	}
	if pos.DimensionsSeparated() {
		t.Fatalf("%s: base Position already separated (want merged)", baseFixture)
	}
	xyz, ok := pos.StaticValue.([]float64)
	if !ok || len(xyz) < 3 {
		t.Fatalf("%s: Position not static 3D: %v", baseFixture, pos.StaticValue)
	}
	wantX, wantY, wantZ := xyz[0], xyz[1], xyz[2]

	if err := pos.SetDimensionsSeparated(true); err != nil {
		t.Fatalf("SetDimensionsSeparated: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "separate_dims_in.aep")
	resavedAEP := filepath.Join(tempDir, "separate_dims_resaved.aep")
	doneFile := filepath.Join(tempDir, "separate_dims.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"x":%g,"y":%g,"z":%g}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), wantX, wantY, wantZ)
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
	t.Logf("SetDimensionsSeparated AE readback:\n%s", body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("separate-dims ship gate FAIL:\n%s", body)
	}

	// Preservation proof: AE's own resave still reports the leader separated and
	// keeps a Position_2 follower carrying the migrated Z — i.e. AE accepted the
	// restructure rather than silently merging it back.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rLeader := findProp(re, aep.MatchNamePosition)
	if rLeader == nil || !rLeader.DimensionsSeparated() {
		t.Fatalf("resaved: Position leader not separated (AE merged it back)")
	}
	rZ := findProp(re, aep.MatchNamePosition2)
	if rZ == nil {
		t.Fatalf("resaved: Position_2 follower missing")
	}
	if z, ok := rZ.StaticValue.(float64); !ok || math.Abs(z-wantZ) > 0.001 {
		t.Errorf("resaved: Position_2 value = %v, want %g", rZ.StaticValue, wantZ)
	}
}

func TestSeparateDims_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runSeparateDimsShipGate(t, aeExe, "../../test_data/re_separate_dims_before.aep")
}

func TestSeparateDims_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runSeparateDimsShipGate(t, aeExe, "../../test_data/re_separate_dims_before.aep")
}
