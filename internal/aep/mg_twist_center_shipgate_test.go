// internal/aep/mg_twist_center_shipgate_test.go
//
// AE ship gate for Twist Center (priority-3 shape remaining,
// specs/2026-06-14-remaining-capability-roadmap.md). The `ADBE Vector Twist
// Center` Vec2 (default [0,0]) is AE-default-elided (no slot in the twist
// template); the serializer materializes it via synthesis-insert (splice the
// Vec2 leaf into the twist body via spliceShapeLeafBeforeGroupEnd, cdat[0:16] =
// two f64 BE — the first Vec2 synthesis-insert leaf, after the scalar Offset
// Copies and the Trim Type / ZigZag Points enums) when SetCenter offsets it.
//
// Verified at the capability's surface per delivery-contract red line 4 with a
// SINGLE frame carrying TWO identical twist cards (same Angle=180, white fill,
// dark BG) — left CENTERED (Center default [0,0]), right OFFSET (Center [150,0]).
// A twist about the path centre is point-symmetric, so the CENTERED card's white
// centroid stays at the card centre; offsetting the pivot makes the twist
// asymmetric, displacing the OFFSET card's centroid. The gate asserts the OFFSET
// centroid is displaced markedly more than the CENTERED one — proving the spliced
// Vec2 actually moved AE's twist pivot, not just that the value round-tripped.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// buildTwistCenterCard adds a ShapeLayer with a Rect + white Fill + Twist
// (Angle=180); Center is left default [0,0] or offset by the caller. Positioned
// at screen (cx,540).
func buildTwistCenterCard(t *testing.T, comp *aep.Composition, name string, cx float64, center *[2]float64) {
	t.Helper()
	card, err := aep.NewShapeLayer(comp, name)
	if err != nil {
		t.Fatalf("NewShapeLayer %s: %v", name, err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("%s AddRect: %v", name, err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("%s rect SetSize: %v", name, err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("%s AddFill: %v", name, err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("%s SetColor: %v", name, err)
	}
	tw, err := card.RootGroup().AddTwist()
	if err != nil {
		t.Fatalf("%s AddTwist: %v", name, err)
	}
	if err := tw.SetAngle(180); err != nil {
		t.Fatalf("%s SetAngle: %v", name, err)
	}
	if center != nil {
		if err := tw.SetCenter(*center); err != nil {
			t.Fatalf("%s SetCenter: %v", name, err)
		}
	}
	if err := card.Position().SetStaticValue([2]float64{cx, 540}); err != nil {
		t.Fatalf("%s Position: %v", name, err)
	}
}

func buildMGTwistCenterDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TWCENTER", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	bgRect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect BG: %v", err)
	}
	if err := bgRect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("BG rect SetSize: %v", err)
	}
	bgFill, err := bg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill BG: %v", err)
	}
	if err := bgFill.SetColor([4]float64{0.05, 0.05, 0.08, 1}); err != nil {
		t.Fatalf("BG SetColor: %v", err)
	}
	if err := bg.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("BG Position: %v", err)
	}

	// Left CENTERED (Center default [0,0], no leaf), right OFFSET (Center [150,0]).
	buildTwistCenterCard(t, comp, "CENTERED", 560, nil)
	off := [2]float64{150, 0}
	buildTwistCenterCard(t, comp, "OFFSET", 1360, &off)

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

// whiteCentroid returns the centroid (cx,cy) of white pixels in the box and the
// sample count (step 2). Used to measure how far a twist's mass is displaced from
// the card centre.
func whiteCentroid(img image.Image, xLo, xHi, yLo, yHi int) (cx, cy float64, n int) {
	var sx, sy, cnt int
	for y := yLo; y <= yHi; y += 2 {
		for x := xLo; x <= xHi; x += 2 {
			r, g, b, _ := img.At(x, y).RGBA()
			if r>>8 >= 200 && g>>8 >= 200 && b>>8 >= 200 {
				sx += x
				sy += y
				cnt++
			}
		}
	}
	if cnt == 0 {
		return 0, 0, 0
	}
	return float64(sx) / float64(cnt), float64(sy) / float64(cnt), cnt
}

func runMGTwistCenterGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_twist_center_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_twist_center.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGTwistCenterDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_twist_center_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_twist_center_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_twist_center.done")
	// Frame to a stable path (not tempDir) so the rendered gate frame can be
	// eyeballed after the run.
	framePNG := `e:/projects/tools/aep-parser/tmp_debug/mg_twist_center_` + ver + `.png`

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
	t.Logf("mg twist center %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg twist center %s ship gate FAIL:\n%s", ver, body)
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

	// CENTERED card centred x=560, OFFSET x=1360, both nominal y=540.
	cCx, cCy, cN := whiteCentroid(img, 560-260, 560+260, 280, 800)
	oCx, oCy, oN := whiteCentroid(img, 1360-260, 1360+260, 280, 800)
	cDisp := math.Hypot(cCx-560, cCy-540)
	oDisp := math.Hypot(oCx-1360, oCy-540)
	t.Logf("%s twist center: CENTERED centroid=(%.1f,%.1f) disp=%.1f n=%d  OFFSET centroid=(%.1f,%.1f) disp=%.1f n=%d",
		ver, cCx, cCy, cDisp, cN, oCx, oCy, oDisp, oN)

	// Both cards must actually render a twisted shape (mass present).
	if cN < 2000 {
		t.Errorf("%s CENTERED card barely rendered (n=%d)", ver, cN)
	}
	if oN < 2000 {
		t.Errorf("%s OFFSET card barely rendered (n=%d)", ver, oN)
	}
	// A twist about the path centre is point-symmetric → the CENTERED card's
	// centroid must stay put (proves the twist itself is symmetric, so any OFFSET
	// displacement is attributable to the Center leaf, not a general asymmetry).
	// AE2020 first run measured CENTERED disp=0.8, OFFSET disp=119.
	if cDisp >= 25 {
		t.Errorf("%s CENTERED centroid disp=%.1f too large — twist not symmetric about path centre", ver, cDisp)
	}
	if oDisp < 50 {
		t.Errorf("%s OFFSET centroid disp=%.1f < 50 — Center=[150,0] did not move the twist pivot", ver, oDisp)
	}

	// Resave proof: the spliced Center leaf survives AE's re-encode at [150,0].
	root := parseAEP(t, resavedAEP)
	cdat := streamCdat(root, "ADBE Vector Twist Center")
	if len(cdat) < 16 {
		t.Fatal("resaved: Twist Center cdat missing/short — spliced leaf dropped")
	}
	cx := math.Float64frombits(binary.BigEndian.Uint64(cdat[0:8]))
	cy := math.Float64frombits(binary.BigEndian.Uint64(cdat[8:16]))
	if math.Abs(cx-150) > 0.5 || math.Abs(cy-0) > 0.5 {
		t.Errorf("resaved Twist Center = [%g,%g], want [150,0]", cx, cy)
	}
}

func TestMGTwistCenter_AEShipGate_AE2020(t *testing.T) {
	runMGTwistCenterGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGTwistCenter_AEShipGate_AE2025(t *testing.T) {
	runMGTwistCenterGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
