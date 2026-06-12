// internal/aep/mg_ease_shipgate_test.go
//
// AE ship gate for temporal-ease keyframes + >2-keyframe streams, both
// from-scratch (MG roadmap S1, specs/2026-06-12-from-scratch-mg-roadmap.md).
// Verified at the capability's surface per delivery-contract red line 4: the
// gate renders the MID-FRAME and asserts the eased layer's position deviates
// from linear interpolation on actual pixels.
//
//   - LIN: dot, Position 2 linear keyframes (200,300)→(1700,300) over 4s —
//     the linear baseline; at t=2s its centre must sit at x≈950.
//   - EAS: dot, same span at y=700 with asymmetric ease (slow-out from kf0,
//     influence 0.9) — at t=2s it must lag far behind x=950.
//   - SCL: dot, Position 6 linear keyframes (zigzag) — breaks the historic
//     2-keyframe coverage boundary; at t=2s it sits at the interpolated
//     midpoint of segment 3, and AE's resave must retain all 6 keyframes.
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

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/codec"
)

func buildMGEaseDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGEASE", 1920, 1080, 30, 4)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	rect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
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

	addDot := func(name string, color [4]float64) *aep.ShapeLayer {
		dl, err := aep.NewShapeLayer(comp, name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", name, err)
		}
		el, err := dl.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("AddEllipse %s: %v", name, err)
		}
		if err := el.SetSize([2]float64{72, 72}); err != nil {
			t.Fatalf("%s SetSize: %v", name, err)
		}
		f, err := dl.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("AddFill %s: %v", name, err)
		}
		if err := f.SetColor(color); err != nil {
			t.Fatalf("%s SetColor: %v", name, err)
		}
		return dl
	}

	lin := addDot("LIN", [4]float64{1.0, 0.55, 0.1, 1}) // amber
	if err := lin.Position().AddKeyframeLinear(0, [2]float64{200, 300}); err != nil {
		t.Fatalf("LIN kf0: %v", err)
	}
	if err := lin.Position().AddKeyframeLinear(4, [2]float64{1700, 300}); err != nil {
		t.Fatalf("LIN kf1: %v", err)
	}

	eas := addDot("EAS", [4]float64{0.25, 0.85, 1.0, 1}) // cyan
	slowOut := codec.TemporalEase{Speed: 0, Influence: 0.9}
	fastIn := codec.TemporalEase{Speed: 0, Influence: 0.1}
	if err := eas.Position().AddKeyframeWithEase(0, [2]float64{200, 700}, codec.TemporalEase{}, slowOut); err != nil {
		t.Fatalf("EAS kf0: %v", err)
	}
	if err := eas.Position().AddKeyframeWithEase(4, [2]float64{1700, 700}, fastIn, codec.TemporalEase{}); err != nil {
		t.Fatalf("EAS kf1: %v", err)
	}

	scl := addDot("SCL", [4]float64{1.0, 0.2, 0.55, 1}) // magenta
	sclKfs := []struct {
		t float64
		v [2]float64
	}{
		{0, [2]float64{200, 850}},
		{0.8, [2]float64{500, 950}},
		{1.6, [2]float64{800, 850}},
		{2.4, [2]float64{1100, 950}},
		{3.2, [2]float64{1400, 850}},
		{4, [2]float64{1700, 950}},
	}
	for _, kf := range sclKfs {
		if err := scl.Position().AddKeyframeLinear(kf.t, kf.v); err != nil {
			t.Fatalf("SCL kf@%g: %v", kf.t, err)
		}
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

// scanRowForColor walks row y and returns the centre x of the widest run of
// pixels within tol of want, or -1 when the colour never appears.
func scanRowForColor(img image.Image, y int, want [3]uint8, tol int) int {
	b := img.Bounds()
	bestStart, bestLen, runStart, runLen := -1, 0, -1, 0
	for x := b.Min.X; x < b.Max.X; x++ {
		r, g, bb, _ := img.At(x, y).RGBA()
		got := [3]int{int(r >> 8), int(g >> 8), int(bb >> 8)}
		match := true
		for i := range 3 {
			d := got[i] - int(want[i])
			if d < -tol || d > tol {
				match = false
				break
			}
		}
		if match {
			if runStart < 0 {
				runStart = x
			}
			runLen++
		} else {
			if runLen > bestLen {
				bestStart, bestLen = runStart, runLen
			}
			runStart, runLen = -1, 0
		}
	}
	if runLen > bestLen {
		bestStart, bestLen = runStart, runLen
	}
	if bestLen < 10 { // a 72px dot crossed mid-row must span far more than noise
		return -1
	}
	return bestStart + bestLen/2
}

func runMGEaseGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_ease_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_ease.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGEaseDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_ease_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_ease_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_ease.done")
	framePNG := filepath.Join(tempDir, "mg_ease_frame.png")

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
	t.Logf("mg ease %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg ease %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at the mid-frame (t=2s).
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const tol = 40
	linX := scanRowForColor(img, 300, [3]uint8{255, 140, 26}, tol)
	easX := scanRowForColor(img, 700, [3]uint8{64, 217, 255}, tol)
	sclX := scanRowForColor(img, 900, [3]uint8{255, 51, 140}, tol)
	t.Logf("%s mid-frame centres: LIN x=%d EAS x=%d SCL x=%d", ver, linX, easX, sclX)
	if linX < 0 || easX < 0 || sclX < 0 {
		t.Fatalf("%s: a dot was not found on its row (LIN %d / EAS %d / SCL %d)", ver, linX, easX, sclX)
	}
	if linX < 890 || linX > 1010 {
		t.Errorf("%s LIN at x=%d, want ≈950 (linear midpoint)", ver, linX)
	}
	if easX >= linX-150 {
		t.Errorf("%s EAS at x=%d vs LIN x=%d — slow-out ease did not lag ≥150px; ease not rendering", ver, easX, linX)
	}
	// SCL t=2.0 sits mid-segment between kf@1.6 (800,850) and kf@2.4 (1100,950):
	// x=950, y=900 — row 900 crosses the dot centre.
	if sclX < 890 || sclX > 1010 {
		t.Errorf("%s SCL at x=%d, want ≈950 (6-keyframe interpolation)", ver, sclX)
	}

	// Resave proof: ease values + all 6 keyframes survive AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	cc := re.Compositions[0]
	easL := cc.LayerByName("EAS")
	if easL == nil {
		t.Fatal("resaved: EAS missing")
	}
	pos := easL.Position()
	if pos == nil || len(pos.Keyframes) != 2 {
		t.Fatalf("resaved EAS position keyframes = %v", pos)
	}
	k0 := pos.Keyframes[0]
	if len(k0.OutTemporalEase) == 0 || k0.OutTemporalEase[0].Influence < 0.8 {
		t.Errorf("resaved EAS kf0 out ease lost: %+v", k0.OutTemporalEase)
	}
	if k0.OutInterp != aep.InterpBezier {
		t.Errorf("resaved EAS kf0 OutInterp = %v, want Bezier", k0.OutInterp)
	}
	sclL := cc.LayerByName("SCL")
	if sclL == nil {
		t.Fatal("resaved: SCL missing")
	}
	if sp := sclL.Position(); sp == nil || len(sp.Keyframes) != 6 {
		n := -1
		if sp != nil {
			n = len(sp.Keyframes)
		}
		t.Errorf("resaved SCL position keyframes = %d, want 6", n)
	}
}

func TestMGEase_AEShipGate_AE2020(t *testing.T) {
	runMGEaseGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGEase_AEShipGate_AE2025(t *testing.T) {
	runMGEaseGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
