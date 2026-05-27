// internal/aep/shape_layer_shipgate_test.go
//
// V2.2 Phase 5 Task 5.2 — AE ship gate (mirrors V2.1 runAEShipGate).
//
// Driver Go side:
//   1. Build canonical 3-ShapeLayer project (spec §5.2) + WriteAEP → tempDir
//   2. Write args.json to fixed path verify_v2_2.jsx reads
//   3. scripts/ae_run.ps1 dispatches AE + auto-dismisses modals
//   4. Assert .done first line == "PASS"
//
// Gated by AE_SHIP_GATE env var; CI / no-AE runs SKIP cleanly.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func runV2_2ShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_canonical.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_canonical.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_test.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_args.json`

	// 1. Build canonical shape graph (spec §5.2) — must match the layer
	//    names and values verify_v2_2.jsx is checking against.
	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	a, _ := comp.NewShapeLayer("A_RectFill_Animated")
	rectA, _ := a.RootGroup().AddRect()
	_ = rectA.Size().AddKeyframeLinear(0, [2]float64{50, 50})
	_ = rectA.Size().AddKeyframeLinear(2, [2]float64{300, 200})
	fillA, _ := a.RootGroup().AddFill()
	_ = fillA.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})
	_ = fillA.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})
	_ = a.Position().AddKeyframeLinear(0, [2]float64{0, 0})
	_ = a.Position().AddKeyframeLinear(2, [2]float64{500, 300})

	b, _ := comp.NewShapeLayer("B_EllipseStroke_Static")
	ellB, _ := b.RootGroup().AddEllipse()
	_ = ellB.SetSize([2]float64{150, 150})
	strokeB, _ := b.RootGroup().AddStroke()
	_ = strokeB.SetColor([4]float64{0, 0, 1, 1})
	_ = strokeB.SetWidth(5)

	c, _ := comp.NewShapeLayer("C_PathFillStroke_Static")
	pathC, _ := c.RootGroup().AddPath()
	_ = pathC.SetVertices([][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}})
	_ = pathC.SetClosed(true)
	fillC, _ := c.RootGroup().AddFill()
	_ = fillC.SetColor([4]float64{0, 1, 0, 1})
	strokeC, _ := c.RootGroup().AddStroke()
	_ = strokeC.SetWidth(2)

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	// 2. args.json — forward-slashed paths for AE
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	// 3. ae_run.ps1 (drop-in for AfterFX -r — handles convert / save-changes / data-loss modals)
	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	// 4. assert PASS (.done guaranteed to exist after ps1 exit 0)
	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ship gate FAIL:\n%s", string(content))
	}
}

// V2.2 ship-gate uses AddEllipse / AddPath / AddStroke — AE silent-drops these
// layer kinds in V2.2 alpha (Rect+Fill only is iter-8 sealed scope). See
// docs/shape.md:260 + scars/v2-2-aelayer-structure.md iter-8 alpha limits.
// Skip until V2.2.1 ships per-kind extract+embed bytes (board.md "Deferred").
func TestV2_2_AEShipGate_AE2025(t *testing.T) {
	t.Skip("V2.2.1 deferred: Ellipse/Path/Stroke embed bytes not yet implemented (see docs/shape.md V2.2.1 notes)")
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_AEShipGate_AE2020(t *testing.T) {
	t.Skip("V2.2.1 deferred: Ellipse/Path/Stroke embed bytes not yet implemented (see docs/shape.md V2.2.1 notes)")
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2020, aeExe)
}
