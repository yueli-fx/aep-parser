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
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
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

// The 3-layer canonical ship-gate exercises Path + Stroke, still deferred in
// V2.2.1 (Ellipse shipped separately — see TestV2_2_Ellipse_AEShipGate_*).
// Skip until per-kind extract+embed bytes land for Path/Stroke.
func TestV2_2_AEShipGate_AE2025(t *testing.T) {
	t.Skip("Path/Stroke embed bytes still deferred (Ellipse ships via TestV2_2_Ellipse_AEShipGate_*)")
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_AEShipGate_AE2020(t *testing.T) {
	t.Skip("Path/Stroke embed bytes still deferred (Ellipse ships via TestV2_2_Ellipse_AEShipGate_*)")
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2ShipGate(t, aep.TargetAE2020, aeExe)
}

// runV2_2EllipseShipGate is the V2.2.1 focused gate: one ShapeLayer with an
// Ellipse + Fill. Asserts AE accepts the Ellipse (no silent-drop) and
// re-derives the Size/Position values Go injected via embed+overwrite. Values
// differ from the tolerance fixture so a no-op overwrite is caught.
func runV2_2EllipseShipGate(t *testing.T, target aep.AETarget, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "v2_2_ellipse.aep")
	resavedAEP := filepath.Join(tempDir, "v2_2_ellipse.resaved.aep")
	doneFile := filepath.Join(tempDir, "v2_2_ellipse.done")
	argsPath := `e:/projects/tools/aep-parser/test_data/v2_2_ellipse_args.json`

	p := aep.NewProject(target)
	comp, err := p.NewComposition("Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := comp.NewShapeLayer("Ellipse_Static")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	ell, err := l.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse: %v", err)
	}
	_ = ell.SetSize([2]float64{260, 140})
	_ = ell.SetPosition([2]float64{70, 90})
	fill, err := l.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	_ = fill.SetColor([4]float64{0.5, 0.5, 0.5, 1})

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

	jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_2_ellipse.jsx`
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(content), "\n", 2)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("ellipse ship gate FAIL:\n%s", string(content))
	}

	// Prove AE retained the overwritten Size/Position by parsing the file AE
	// re-saved (sidesteps the .value ExtendScript quirk). AE writes the value
	// at cdat offset 0 (BE f64), same as the native tolerance fixture.
	assertResavedEllipse(t, resavedAEP, [2]float64{260, 140}, [2]float64{70, 90})
}

// assertResavedEllipse parses an AE-resaved .aep, locates the Ellipse Size /
// Position cdat, and asserts the leading f64 BE values match want.
func assertResavedEllipse(t *testing.T, path string, wantSize, wantPos [2]float64) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open resaved: %v", err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		t.Fatalf("parse resaved: %v", err)
	}
	gotSize := readShapeStreamVec2(t, root, "ADBE Vector Ellipse Size")
	gotPos := readShapeStreamVec2(t, root, "ADBE Vector Ellipse Position")
	if math.Abs(gotSize[0]-wantSize[0]) > 0.5 || math.Abs(gotSize[1]-wantSize[1]) > 0.5 {
		t.Errorf("resaved Ellipse Size = %v, want %v", gotSize, wantSize)
	}
	if math.Abs(gotPos[0]-wantPos[0]) > 0.5 || math.Abs(gotPos[1]-wantPos[1]) > 0.5 {
		t.Errorf("resaved Ellipse Position = %v, want %v", gotPos, wantPos)
	}
}

// readShapeStreamVec2 finds the tdmn matching name anywhere in the tree, reads
// the following tdbs cdat, and returns its first two f64 BE values.
func readShapeStreamVec2(t *testing.T, root *rifx.Chunk, name string) [2]float64 {
	t.Helper()
	var out [2]float64
	var found bool
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found {
			return
		}
		for i := 0; i < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && i+1 < len(c.Children) && trimShipNUL(string(ch.Data)) == name {
				tdbs := c.Children[i+1]
				if tdbs.IsList() && tdbs.FormType == rifx.IDTdbs {
					for _, cc := range tdbs.Children {
						if cc.ID == rifx.IDCdat && len(cc.Data) >= 16 {
							out[0] = math.Float64frombits(binary.BigEndian.Uint64(cc.Data[0:8]))
							out[1] = math.Float64frombits(binary.BigEndian.Uint64(cc.Data[8:16]))
							found = true
							return
						}
					}
				}
			}
			if ch.IsList() {
				walk(ch)
				if found {
					return
				}
			}
		}
	}
	walk(root)
	if !found {
		t.Fatalf("%s: not found in resaved aep", name)
	}
	return out
}

func trimShipNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}

func TestV2_2_Ellipse_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runV2_2EllipseShipGate(t, aep.TargetAE2025, aeExe)
}

func TestV2_2_Ellipse_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runV2_2EllipseShipGate(t, aep.TargetAE2020, aeExe)
}
