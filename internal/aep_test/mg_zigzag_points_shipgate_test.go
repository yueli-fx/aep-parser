// internal/aep/mg_zigzag_points_shipgate_test.go
//
// AE ship gate for ZigZag Points (priority-3 shape remaining,
// specs/2026-06-14-remaining-capability-roadmap.md). The `ADBE Vector Zigzag
// Points` enum (Corner=1 default / Smooth=2) is AE-default-elided (no slot in the
// zigzag template); the serializer materializes it via synthesis-insert (splice
// the leaf into the zigzag body via spliceShapeLeafBeforeGroupEnd, mirroring Trim
// Type / Offset Copies) when SetPoints selects Smooth.
//
// Verified at the capability's surface per delivery-contract red line 4 with a
// SINGLE frame carrying TWO identical zigzag cards (same Size/Detail, white fill,
// dark BG) — left is the default CORNER, right is SMOOTH. Holding everything but
// Points constant makes any pixel difference attributable to Points alone. The
// discriminator exploits the Corner-vs-Smooth geometry: a Corner ridge comes to a
// sharp narrow apex (few columns near the tooth's topmost extent), while a Smooth
// ridge is a rounded broad arc (many columns near the topmost extent). The gate
// asserts the Smooth card's apexes are markedly broader than the Corner card's —
// proving the spliced enum actually changed AE's render, not just that the value
// round-tripped.
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

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// buildZZPointsCard adds a ShapeLayer with a Rect + white Fill + ZigZag
// (Size/Detail set big so the teeth are large and clear); Points is left default
// (Corner) or set Smooth by the caller. Positioned at screen (cx,540).
func buildZZPointsCard(t *testing.T, comp *aep.Composition, name string, cx float64, smooth bool) {
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
	zz, err := card.RootGroup().AddZigZag()
	if err != nil {
		t.Fatalf("%s AddZigZag: %v", name, err)
	}
	if err := zz.SetSize(70); err != nil {
		t.Fatalf("%s SetSize: %v", name, err)
	}
	if err := zz.SetDetail(5); err != nil {
		t.Fatalf("%s SetDetail: %v", name, err)
	}
	if smooth {
		if err := zz.SetPoints(aep.ZigZagPointsSmooth); err != nil {
			t.Fatalf("%s SetPoints: %v", name, err)
		}
	}
	if err := card.Position().SetStaticValue([2]float64{cx, 540}); err != nil {
		t.Fatalf("%s Position: %v", name, err)
	}
}

func buildMGZigZagPointsDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "ZZPOINTS", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the white zigzag edges read unambiguously.
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

	// Left CORNER (default), right SMOOTH — identical otherwise.
	buildZZPointsCard(t, comp, "CORNER", 560, false)
	buildZZPointsCard(t, comp, "SMOOTH", 1360, true)

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

// apexBandWidth scans the top edge of a card centred at cx and returns
// (spread, flatCount): spread = (max-min) of topmost-white-y across the scanned
// columns (proves the zigzag is active), flatCount = number of columns whose
// topmost-white-y is within `band` px of the card's global topmost y (a measure
// of how broad the ridge apexes are — sharp Corner apexes are narrow → small
// count; rounded Smooth apexes are broad → large count). Scans the middle of the
// edge only, skipping the card corners where AE draws teardrop loops.
func apexBandWidth(img image.Image, cx int, band int) (spread, flatCount, samples int) {
	xLo, xHi := cx-140, cx+140 // middle 280px of the 400px edge
	type col struct{ x, y int }
	var cols []col
	for x := xLo; x <= xHi; x++ {
		y := topWhiteY(img, x, 180, 540)
		if y < 0 {
			continue
		}
		cols = append(cols, col{x, y})
	}
	if len(cols) == 0 {
		return 0, 0, 0
	}
	mn, mx := cols[0].y, cols[0].y
	for _, c := range cols {
		if c.y < mn {
			mn = c.y
		}
		if c.y > mx {
			mx = c.y
		}
	}
	flat := 0
	for _, c := range cols {
		if c.y <= mn+band {
			flat++
		}
	}
	return mx - mn, flat, len(cols)
}

func runMGZigZagPointsGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_zigzag_points_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_zigzag_points.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGZigZagPointsDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_zigzag_points_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_zigzag_points_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_zigzag_points.done")
	// Frame to a stable path (not tempDir, which is wiped on teardown) so the
	// rendered gate frame can be eyeballed after the run.
	framePNG := `e:/projects/tools/aep-parser/tmp_debug/mg_zigzag_points_` + ver + `.png`

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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("mg zigzag points %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg zigzag points %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. CORNER card centred x=560, SMOOTH x=1360.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}

	const band = 8
	cornerSpread, cornerFlat, cornerN := apexBandWidth(img, 560, band)
	smoothSpread, smoothFlat, smoothN := apexBandWidth(img, 1360, band)
	t.Logf("%s zigzag points: CORNER spread=%d flat=%d/%d  SMOOTH spread=%d flat=%d/%d",
		ver, cornerSpread, cornerFlat, cornerN, smoothSpread, smoothFlat, smoothN)

	// Both cards must show an active zigzag (edge oscillates).
	if cornerSpread < 40 {
		t.Errorf("%s CORNER spread=%d too small — zigzag not active", ver, cornerSpread)
	}
	if smoothSpread < 40 {
		t.Errorf("%s SMOOTH spread=%d too small — zigzag not active", ver, smoothSpread)
	}
	// Smooth apexes are rounded/broad → markedly more columns near the topmost
	// extent than the sharp Corner apexes. AE2020 first run measured 65 vs 22
	// (~3x); require a clear 2x margin so a marginal difference can't pass and a
	// dropped leaf (Smooth falling back to Corner → both ~22) fails.
	if smoothFlat < cornerFlat*2 {
		t.Errorf("%s SMOOTH apex flat=%d not >= 2x CORNER flat=%d — Points=Smooth did not round the ridges",
			ver, smoothFlat, cornerFlat)
	}

	// Resave proof: the spliced Points leaf survives AE's re-encode at value 2.
	root := parseAEP(t, resavedAEP)
	cdat := streamCdat(root, "ADBE Vector Zigzag Points")
	if len(cdat) < 8 {
		t.Fatal("resaved: Zigzag Points cdat missing — spliced leaf dropped")
	}
	if v := math.Float64frombits(binary.BigEndian.Uint64(cdat[:8])); math.Abs(v-2) > 0.01 {
		t.Errorf("resaved Zigzag Points = %g, want 2 (Smooth)", v)
	}
}

func TestMGZigZagPoints_AEShipGate_AE2020(t *testing.T) {
	runMGZigZagPointsGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGZigZagPoints_AEShipGate_AE2025(t *testing.T) {
	runMGZigZagPointsGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
