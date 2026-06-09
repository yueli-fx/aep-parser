// internal/aep/shape_taperwave_shipgate_test.go
//
// AE ship gate for Stroke Taper + Wave. Builds one ShapeLayer (rect + stroke)
// with the Taper %-mode controls and Wave Wavelength-mode controls set non-
// default, has AE open + resave it, and decodes the resaved cdats. Proves AE
// accepts the enriched stroke template (no silent-drop) AND honors the nested-
// group cdat overwrites. Gated by AE_SHIP_GATE.
//
// Reuses verify_v2_2_stroke.jsx (layer named "Stroke_Static" with rect+stroke;
// it just asserts no silent-drop and resaves). RE: Taper/Wave addendum in
// incident-reports/stroke-line-cap-join-miter-re.md.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runV2_2StrokeTaperWaveShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_taperwave.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_taperwave.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_taperwave.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_stroke_args.json`

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Stroke_Static") // name expected by verify_v2_2_stroke.jsx
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.SetWidth(12)
	tp := stroke.Taper()
	_ = tp.SetStartLength(30)
	_ = tp.SetEndLength(40)
	_ = tp.SetStartWidth(55)
	_ = tp.SetEndWidth(65)
	_ = tp.SetStartEase(35)
	_ = tp.SetEndEase(50)
	wv := stroke.Wave()
	_ = wv.SetAmount(20)
	_ = wv.SetWavelength(70)
	_ = wv.SetPhase(45)

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_stroke.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.SplitN(string(content), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("taper/wave ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	rdF64 := func(b []byte, off int) float64 {
		return math.Float64frombits(binary.BigEndian.Uint64(b[off : off+8]))
	}
	// All Taper/Wave sub-stream match-names are unique, so first-match streamCdat
	// reaches the nested-group leaves directly.
	checks := []struct {
		name string
		want float64
	}{
		{"ADBE Vector Taper Start Length", 30},
		{"ADBE Vector Taper End Length", 40},
		{"ADBE Vector Taper Start Width", 55},
		{"ADBE Vector Taper End Width", 65},
		{"ADBE Vector Taper Start Ease", 35},
		{"ADBE Vector Taper End Ease", 50},
		{"ADBE Vector Taper Wave Amount", 20},
		{"ADBE Vector Taper Wavelength", 70},
		{"ADBE Vector Taper Wave Phase", 45},
	}
	for _, c := range checks {
		cdat := streamCdat(root, c.name)
		if cdat == nil {
			t.Errorf("resaved %q cdat missing — AE dropped the slot?", c.name)
			continue
		}
		if got := rdF64(cdat, 0); math.Abs(got-c.want) > 0.01 {
			t.Errorf("resaved %q = %.3g, want %.3g", c.name, got, c.want)
		}
	}
}

func TestV2_2_StrokeTaperWave_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2StrokeTaperWaveShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_StrokeTaperWave_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2StrokeTaperWaveShipGate(t, aep.TargetAE2020, aeExe)
}
