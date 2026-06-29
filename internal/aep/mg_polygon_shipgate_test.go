// internal/aep/mg_polygon_shipgate_test.go
//
// AE ship gate for PolyStar POLYGON type (`ADBE Vector Star Type` = 2) from
// scratch. A 6-point polygon (outer radius 200) filled white, centred in the
// comp. Per delivery-contract red line 4 the gate renders frame 0 and asserts
// the fill is CONVEX: a ring of 12 points at radius 140 (0.7·R, below the
// hexagon's edge-midpoint radius 200·cos30°≈173) are ALL white. A Star with the
// same outer radius (default inner radius 50) would leave the 6 notch directions
// at radius 140 on the background — which the all-filled check rejects. AE also
// reads back Star Type=2 / Points=6.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildMGPolygonDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGPOLY", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "POLY")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	star, err := l.RootGroup().AddStar()
	if err != nil {
		t.Fatalf("AddStar: %v", err)
	}
	if err := star.SetStarType(aep.StarTypePolygon); err != nil {
		t.Fatalf("SetStarType: %v", err)
	}
	if err := star.SetPoints(6); err != nil {
		t.Fatalf("SetPoints: %v", err)
	}
	if err := star.SetOuterRadius(200); err != nil {
		t.Fatalf("SetOuterRadius: %v", err)
	}
	if err := star.SetRotation(0); err != nil {
		t.Fatalf("SetRotation: %v", err)
	}
	fill, err := l.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("SetColor: %v", err)
	}
	if err := l.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	return rp
}

func runMGPolygonGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_polygon_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_polygon.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGPolygonDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_polygon_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_polygon_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_polygon.done")
	framePNG := filepath.Join(tempDir, "mg_polygon_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(framePNG))
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
	t.Logf("mg polygon %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg polygon %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const cx, cy = 960, 540
	white := func(r, g, b int) bool { return r > 180 && g > 180 && b > 180 }
	// Convex proof: 12 points on a ring at radius 140 must ALL be white. A
	// 6-point star (notches at inner radius 50) would leave the notch directions
	// on the background here.
	const ring = 140
	bg := 0
	for k := 0; k < 12; k++ {
		ang := float64(k) * math.Pi / 6
		px := cx + int(ring*math.Cos(ang))
		py := cy + int(ring*math.Sin(ang))
		r, g, b := avgRGB(img, px, py, 4)
		if !white(r, g, b) {
			bg++
			t.Logf("%s ring point k=%d (%d,%d) NOT white: r=%d g=%d b=%d", ver, k, px, py, r, g, b)
		}
	}
	if bg > 0 {
		t.Errorf("%s %d/12 ring points at r=140 not filled — shape is not a convex polygon (star notches?)", ver, bg)
	}
	// Bounded: a point well outside the outer radius (r=260) is background.
	oR, oG, oB := avgRGB(img, cx+260, cy, 4)
	if white(oR, oG, oB) {
		t.Errorf("%s point 60px past outer radius is white (r=%d g=%d b=%d) — shape unbounded", ver, oR, oG, oB)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("POLY") == nil {
		t.Fatal("resaved: POLY layer missing")
	}
}

func TestMGPolygon_AEShipGate_AE2020(t *testing.T) {
	runMGPolygonGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGPolygon_AEShipGate_AE2025(t *testing.T) {
	runMGPolygonGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
