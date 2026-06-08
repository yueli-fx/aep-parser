// internal/aep/shape_dashes_shipgate_test.go
//
// AE ship gate for Stroke Dashes. Builds one ShapeLayer (rect + stroke) with
// dashing enabled (Dash/Gap set non-default), has AE open + resave it, and
// decodes the resaved Dash 1 / Gap 1 cdats. Proves AE accepts the dashed
// stroke-body template (no silent-drop) AND honors the nested Dashes-group cdat
// overwrites. Gated by AE_SHIP_GATE.
//
// Reuses verify_v2_2_stroke.jsx (layer named "Stroke_Static" with rect+stroke;
// it asserts no silent-drop and resaves). Offset is intentionally not exercised
// — AE keeps it hidden / script-ungettable (RE: v2_2_stroke_dashed.done note,
// stroke-line-cap-join-miter-re.md Dashes addendum).
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

func runV2_2StrokeDashesShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_dashes.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_dashes.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_dashes.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_stroke_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
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
	d := stroke.Dashes()
	_ = d.SetDash(18)
	_ = d.SetGap(7)

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
		t.Errorf("dashes ship gate FAIL:\n%s", string(content))
	}

	root := parseAEP(t, resavedAEP)
	rdF64 := func(b []byte, off int) float64 {
		return math.Float64frombits(binary.BigEndian.Uint64(b[off : off+8]))
	}
	// Dash 1 / Gap 1 match-names are unique, so first-match streamCdat reaches
	// the nested Dashes-group leaves directly.
	checks := []struct {
		name string
		want float64
	}{
		{"ADBE Vector Stroke Dash 1", 18},
		{"ADBE Vector Stroke Gap 1", 7},
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

func TestV2_2_StrokeDashes_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2StrokeDashesShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_StrokeDashes_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2StrokeDashesShipGate(t, aep.TargetAE2020, aeExe)
}
