// internal/aep/shape_ellipse_shipgate_test.go
//
// AE ship gate: Ellipse + Fill focused gate.
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

	// Fill color round-trips with the V2.2.1 ARGB×255 encoding (fill set to
	// [0.5,0.5,0.5,1] → ARGB×255 [255,127.5,127.5,127.5]).
	root := parseAEP(t, resavedAEP)
	if fc := streamCdat(root, "ADBE Vector Fill Color"); len(fc) >= 32 {
		rd := func(off int) float64 { return math.Float64frombits(binary.BigEndian.Uint64(fc[off : off+8])) }
		want := []float64{255, 127.5, 127.5, 127.5}
		for i, w := range want {
			if got := rd(i * 8); math.Abs(got-w) > 1.0 {
				t.Errorf("resaved fill color[%d] = %.4g, want %.4g (ARGB×255)", i, got, w)
			}
		}
	} else {
		t.Errorf("resaved fill color cdat missing/short")
	}
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
