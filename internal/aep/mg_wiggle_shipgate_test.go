// internal/aep/mg_wiggle_shipgate_test.go
//
// AE ship gate for the Wiggle Paths filter (`ADBE Vector Filter - Roughen`
// internally) from scratch (MG roadmap S5,
// specs/2026-06-12-from-scratch-mg-roadmap.md). Verified at the capability's
// surface per delivery-contract red line 4: a 400×400 white rectangle gets a
// Wiggle Paths filter (Size=60, Detail=30, Wiggles/Second=4, Seed=9). The gate
// renders the frame and asserts the originally straight top edge was roughened
// into a noisy boundary — the per-column topmost-white-y spreads far beyond the
// ~0 of a clean rect — proving the random displacement actually reshaped the
// path, not just that the values round-tripped.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildMGWiggleDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGWG", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the roughened silhouette reads unambiguously.
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

	// A 400×400 white rectangle + Wiggle Paths. Centered at 960,540 → originally
	// spans x∈[760,1160], y∈[340,740]. The wiggle roughens every straight edge
	// into a random jagged boundary.
	card, err := aep.NewShapeLayer(comp, "CARD")
	if err != nil {
		t.Fatalf("NewShapeLayer CARD: %v", err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect CARD: %v", err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("CARD rect SetSize: %v", err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill CARD: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("CARD SetColor: %v", err)
	}
	wg, err := card.RootGroup().AddWigglePaths()
	if err != nil {
		t.Fatalf("AddWigglePaths: %v", err)
	}
	if err := wg.SetSize(60); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	if err := wg.SetDetail(30); err != nil {
		t.Fatalf("SetDetail: %v", err)
	}
	if err := wg.SetWigglesPerSecond(4); err != nil {
		t.Fatalf("SetWigglesPerSecond: %v", err)
	}
	if err := wg.SetRandomSeed(9); err != nil {
		t.Fatalf("SetRandomSeed: %v", err)
	}
	if err := card.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("CARD Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

func runMGWiggleGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_wiggle_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_wiggle.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGWiggleDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_wiggle_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_wiggle_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_wiggle.done")
	framePNG := filepath.Join(tempDir, "mg_wiggle_frame.png")

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
	t.Logf("mg wiggle %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg wiggle %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. Original card spans x∈[760,1160], y∈[340,740].
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}

	// Diagnostic grid over the card region so the roughened silhouette can be
	// eyeballed in the log alongside the pixel assertions.
	var sb strings.Builder
	for y := 280; y <= 800; y += 30 {
		fmt.Fprintf(&sb, "y=%4d ", y)
		for x := 700; x <= 1220; x += 20 {
			if whiteNear(img, x, y, 4) {
				sb.WriteByte('#')
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteByte('\n')
	}
	t.Logf("%s wiggle silhouette:\n%s", ver, sb.String())

	// Center stays white (interior preserved under the roughen).
	if !whiteNear(img, 960, 540, 10) {
		t.Errorf("%s center not white — card body not rendered", ver)
	}

	// Edge-roughness: scan the top edge across the interior columns and find the
	// topmost white y per column. A clean rect's top is a flat line (spread ≈ 0);
	// the wiggle's random displacement makes the topmost-white-y vary widely.
	topMin, topMax := 1<<30, -1
	cols := 0
	for x := 800; x <= 1120; x += 8 {
		topY := -1
		for y := 240; y <= 540; y++ {
			if whiteNear(img, x, y, 1) {
				topY = y
				break
			}
		}
		if topY >= 0 {
			cols++
			if topY < topMin {
				topMin = topY
			}
			if topY > topMax {
				topMax = topY
			}
		}
	}
	spread := 0
	if cols > 0 {
		spread = topMax - topMin
	}
	t.Logf("%s wiggle: top-edge cols=%d topY∈[%d,%d] spread=%d", ver, cols, topMin, topMax, spread)
	// A clean (un-roughened) rect top is a flat line → spread ≈ 0 (only AA jitter
	// of a pixel or two). The wiggle's random displacement spreads the topmost
	// white y by tens of pixels (measured ~41 at Size=60/Seed=9). 25 sits well
	// above the no-op floor with margin below the observed value.
	if spread < 25 {
		t.Errorf("%s top-edge spread=%d too low — edge not roughened (wiggle no-op?)", ver, spread)
	}

	// Resave proof: the Wiggle Paths filter survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("CARD") == nil {
		t.Fatal("resaved: CARD layer missing")
	}
}

func TestMGWiggle_AEShipGate_AE2020(t *testing.T) {
	runMGWiggleGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGWiggle_AEShipGate_AE2025(t *testing.T) {
	runMGWiggleGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
