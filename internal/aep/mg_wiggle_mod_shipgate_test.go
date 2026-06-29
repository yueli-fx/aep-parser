// internal/aep/mg_wiggle_mod_shipgate_test.go
//
// AE ship gate for the Wiggle modulation render-pixel knobs — the shape
// vector-filter family's last visually-gatable sub-streams:
//   - Wiggle Paths `ADBE Vector Roughen Points` (Corner=1 default / Smooth=2)
//   - Wiggle Paths `ADBE Vector Correlation` (0 = jagged independent jitter /
//     100 = coherent smooth boil)
// Both are AE-default-elided and materialized via synthesis-insert
// (spliceShapeLeafBefore, mirroring ZigZag Points / Twist Center).
//
// One frame, four cards, all sharing Size/Detail/Seed so any pixel difference is
// attributable to the one varied knob (red line 4: render the capability's
// surface). Fixed RandomSeed → frame 0 is deterministic and cross-version
// reproducible (proven by TestMGWiggle). CORNER vs SMOOTH share the random
// pattern but render sharp spikes vs rounded bumps; CORRLOW vs CORRHIGH change
// the spatial coherence of the displacement.
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

// buildWiggleModCard adds a ShapeLayer: Rect 300 + white Fill + Wiggle Paths
// (Size/Detail/Seed fixed). The caller varies exactly one modulation knob.
func buildWiggleModCard(t *testing.T, comp *aep.Composition, name string, cx float64, apply func(n *aep.WigglePathsNode)) {
	t.Helper()
	card, err := aep.NewShapeLayer(comp, name)
	if err != nil {
		t.Fatalf("NewShapeLayer %s: %v", name, err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("%s AddRect: %v", name, err)
	}
	if err := rect.SetSize([2]float64{300, 300}); err != nil {
		t.Fatalf("%s rect SetSize: %v", name, err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("%s AddFill: %v", name, err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("%s SetColor: %v", name, err)
	}
	wg, err := card.RootGroup().AddWigglePaths()
	if err != nil {
		t.Fatalf("%s AddWigglePaths: %v", name, err)
	}
	if err := wg.SetSize(60); err != nil {
		t.Fatalf("%s SetSize: %v", name, err)
	}
	if err := wg.SetDetail(8); err != nil {
		t.Fatalf("%s SetDetail: %v", name, err)
	}
	if err := wg.SetRandomSeed(5); err != nil {
		t.Fatalf("%s SetRandomSeed: %v", name, err)
	}
	apply(wg)
	if err := card.Position().SetStaticValue([2]float64{cx, 540}); err != nil {
		t.Fatalf("%s Position: %v", name, err)
	}
}

func buildMGWiggleModDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "WGMOD", 1920, 1080, 30, 5)
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

	buildWiggleModCard(t, comp, "CORNER", 360, func(n *aep.WigglePathsNode) {})
	buildWiggleModCard(t, comp, "SMOOTH", 760, func(n *aep.WigglePathsNode) {
		if err := n.SetPoints(aep.RoughenPointsSmooth); err != nil {
			t.Fatalf("SetPoints: %v", err)
		}
	})
	buildWiggleModCard(t, comp, "CORRLOW", 1160, func(n *aep.WigglePathsNode) {
		if err := n.SetCorrelation(0); err != nil {
			t.Fatalf("SetCorrelation 0: %v", err)
		}
	})
	buildWiggleModCard(t, comp, "CORRHIGH", 1560, func(n *aep.WigglePathsNode) {
		if err := n.SetCorrelation(100); err != nil {
			t.Fatalf("SetCorrelation 100: %v", err)
		}
	})

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

// wiggleEdgeStats scans the top edge of a card centred at cx (Rect 300 → edge
// spans cx±150; scan the middle cx±90 to skip the corner loops). Returns:
//   spread = (max-min) of topmost-white-y across columns — proves wiggle active,
//            and collapses to ~0 under full Correlation (rigid offset).
//   jag    = mean |Δy| between adjacent columns — high-frequency roughness.
//   curv   = mean chord deviation |y[i]-(y[i-d]+y[i+d])/2| at baseline d — low on
//            straight Corner ramps (points on the chord), high on curved Smooth
//            arcs (points bow off the chord). The Corner/Smooth discriminator.
func wiggleEdgeStats(img image.Image, cx int) (spread int, jag, curv float64, samples int) {
	xLo, xHi := cx-90, cx+90
	var ys []int
	for x := xLo; x <= xHi; x++ {
		y := topWhiteY(img, x, 220, 560)
		if y < 0 {
			continue
		}
		ys = append(ys, y)
	}
	if len(ys) == 0 {
		return 0, 0, 0, 0
	}
	mn, mx := ys[0], ys[0]
	for _, y := range ys {
		if y < mn {
			mn = y
		}
		if y > mx {
			mx = y
		}
	}
	var jsum float64
	for i := 1; i < len(ys); i++ {
		jsum += math.Abs(float64(ys[i] - ys[i-1]))
	}
	if len(ys) > 1 {
		jag = jsum / float64(len(ys)-1)
	}
	// curv = mean chord deviation at baseline d: |y[i] - (y[i-d]+y[i+d])/2|.
	// On a straight Corner ramp the sampled point lies on the chord (dev≈0); on a
	// curved Smooth arc it bows away from the chord (dev>0). A wide baseline (d>1)
	// amplifies the straight-vs-curved distinction that the 1-step 2nd difference
	// washes out.
	const d = 10
	var csum float64
	cn := 0
	for i := d; i+d < len(ys); i++ {
		csum += math.Abs(float64(ys[i]) - float64(ys[i-d]+ys[i+d])/2)
		cn++
	}
	if cn > 0 {
		curv = csum / float64(cn)
	}
	return mx - mn, jag, curv, len(ys)
}

// wiggleWhiteArea counts white pixels in the card's bounding box (cx±200,
// y 250..830). A sharp Corner roughen produces thin pointed spikes (less
// enclosed area); a Smooth roughen produces fat rounded lobes (more area).
func wiggleWhiteArea(img image.Image, cx int) int {
	b := img.Bounds()
	n := 0
	for y := 250; y <= 830; y++ {
		if y < b.Min.Y || y >= b.Max.Y {
			continue
		}
		for x := cx - 200; x <= cx+200; x++ {
			if x < b.Min.X || x >= b.Max.X {
				continue
			}
			r, g, bb, _ := img.At(x, y).RGBA()
			if r>>8 >= 200 && g>>8 >= 200 && bb>>8 >= 200 {
				n++
			}
		}
	}
	return n
}

func runMGWiggleModGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_wiggle_mod_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_wiggle_mod.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGWiggleModDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_wiggle_mod_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_wiggle_mod_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_wiggle_mod.done")
	framePNG := `e:/projects/tools/aep-parser/tmp_debug/mg_wiggle_mod_` + ver + `.png`

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
	t.Logf("mg wiggle mod %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg wiggle mod %s ship gate FAIL:\n%s", ver, body)
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

	cornerSpread, cornerJag, cornerCurv, cornerN := wiggleEdgeStats(img, 360)
	smoothSpread, smoothJag, smoothCurv, smoothN := wiggleEdgeStats(img, 760)
	lowSpread, lowJag, lowCurv, lowN := wiggleEdgeStats(img, 1160)
	highSpread, highJag, highCurv, highN := wiggleEdgeStats(img, 1560)
	cornerArea := wiggleWhiteArea(img, 360)
	smoothArea := wiggleWhiteArea(img, 760)
	lowArea := wiggleWhiteArea(img, 1160)
	highArea := wiggleWhiteArea(img, 1560)
	t.Logf("%s CORNER  spread=%d jag=%.2f curv=%.2f area=%d n=%d", ver, cornerSpread, cornerJag, cornerCurv, cornerArea, cornerN)
	t.Logf("%s SMOOTH  spread=%d jag=%.2f curv=%.2f area=%d n=%d", ver, smoothSpread, smoothJag, smoothCurv, smoothArea, smoothN)
	t.Logf("%s CORRLOW spread=%d jag=%.2f curv=%.2f area=%d n=%d", ver, lowSpread, lowJag, lowCurv, lowArea, lowN)
	t.Logf("%s CORRHGH spread=%d jag=%.2f curv=%.2f area=%d n=%d", ver, highSpread, highJag, highCurv, highArea, highN)

	// Points cards (CORNER/SMOOTH) and the jagged CORRLOW must show an active,
	// roughened edge. CORRHIGH is deliberately NOT checked here — full coherence
	// flattens the edge (that IS the correct render; see below).
	for _, s := range []struct {
		name string
		v    int
	}{{"CORNER", cornerSpread}, {"SMOOTH", smoothSpread}, {"CORRLOW", lowSpread}} {
		if s.v < 20 {
			t.Errorf("%s %s spread=%d too small — wiggle not active", ver, s.name, s.v)
		}
	}

	// Points discriminator (curv = mean chord deviation, d=10): a sharp Corner
	// sawtooth connects its displaced vertices with straight ramps, so sampled
	// points sit ON the chord → low deviation; a Smooth bump arcs between the same
	// vertices, bowing OFF the chord → high deviation. So SMOOTH curv must sit
	// clearly above CORNER curv. AE2020 render: CORNER≈1.99, SMOOTH≈3.14 (1.58x);
	// require >=1.3x so a dropped leaf (Smooth falling back to Corner → ratio 1.0)
	// fails. Both cards carry an identical seed/Size/Detail, so the curvature
	// difference is attributable to Points alone.
	if smoothCurv < cornerCurv*1.3 {
		t.Errorf("%s SMOOTH curv=%.2f not >= 1.3x CORNER curv=%.2f — Points=Smooth did not round the apexes", ver, smoothCurv, cornerCurv)
	}

	// Correlation discriminator: low correlation = each point jitters
	// independently → jagged edge (large spread); high correlation = neighbours
	// move together → the displacement becomes a near-rigid offset, leaving the
	// edge essentially flat (small spread). The categorical signal is the spread
	// collapse: CORRHIGH must be near-flat AND markedly flatter than CORRLOW.
	if highSpread >= 15 {
		t.Errorf("%s CORRHIGH spread=%d not near-flat — Correlation=100 should make the displacement a rigid offset", ver, highSpread)
	}
	if lowSpread < highSpread*3 {
		t.Errorf("%s CORRLOW spread=%d not >= 3x CORRHIGH spread=%d — Correlation did not change edge coherence", ver, lowSpread, highSpread)
	}

	// Resave proof: the spliced leaves survive AE's re-encode.
	root := parseAEP(t, resavedAEP)
	for _, nm := range []string{"ADBE Vector Roughen Points", "ADBE Vector Correlation"} {
		cdat := streamCdat(root, nm)
		if len(cdat) < 8 {
			t.Errorf("resaved: %s cdat missing — spliced leaf dropped", nm)
			continue
		}
		v := math.Float64frombits(binary.BigEndian.Uint64(cdat[:8]))
		t.Logf("%s resaved %s = %g", ver, nm, v)
	}
}

func TestMGWiggleMod_AEShipGate_AE2020(t *testing.T) {
	runMGWiggleModGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGWiggleMod_AEShipGate_AE2025(t *testing.T) {
	runMGWiggleModGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
