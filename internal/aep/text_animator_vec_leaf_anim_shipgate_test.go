// internal/aep/text_animator_vec_leaf_anim_shipgate_test.go
//
// AE ship gates for animating a 3D/4D text animator LEAF over time (Fill Color
// and Position 3D), vs sweeping the Range Selector. Per delivery-contract red
// line 4 each gate renders THREE frames of the SAME from-scratch text layer with
// the Range Offset LEFT STATIC (so every character shares the leaf's own value
// curve) and asserts on actual pixels:
//
//   - Fill Color leaf keyframed red→blue: the glyph ink mean colour migrates from
//     red (high R, low B) to blue (low R, high B). Distinct from the Opacity /
//     Rotation scalar-leaf gates and from the colour-wipe selector sweep.
//   - Position 3D leaf keyframed (0,0,0)→(0,250,0): the glyph ink vertical
//     centroid migrates downward.
//
// Color and Position exercise DIFFERENT animated keyframe blocks (4-channel
// marker-2 vs spatial 3D marker-3), both byte-verified against the AE-saved
// ground truth (re_text_animator_animatedvec) in the unit tests; these gates
// confirm AE acceptance + correct render. Each resaved file is re-parsed to
// confirm the leaf keeps 2 keyframes and the Range Offset stays static.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

// inkCentroid returns the centroid (cx, cy) of the ink pixels (luminance
// differing from bgLum by more than thresh) over the search box, subsampled by
// step, plus the ink pixel count. Centroid is 0,0 when no ink.
func inkCentroid(img image.Image, x0, y0, x1, y1, step, bgLum, thresh int) (float64, float64, int) {
	var sumX, sumY, n int
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			l := int((r>>8 + g>>8 + b>>8) / 3)
			if d := l - bgLum; d > thresh || d < -thresh {
				sumX += x
				sumY += y
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0, 0
	}
	return float64(sumX) / float64(n), float64(sumY) / float64(n), n
}

func newTextVecLeafComp(t *testing.T, target aep.AETarget, name string) (*aep.Project, *aep.Layer) {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, name, 1280, 720, 24, 5)
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
	if err := rect.SetSize([2]float64{1600, 1000}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
	}
	bgFill, err := bg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	if err := bgFill.SetColor([4]float64{0.05, 0.05, 0.08, 1}); err != nil {
		t.Fatalf("BG SetColor: %v", err)
	}
	if err := bg.Position().SetStaticValue([2]float64{640, 360}); err != nil {
		t.Fatalf("BG Position: %v", err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	return p, tl
}

func reopenMoveBG(t *testing.T, p *aep.Project) *aep.Project {
	t.Helper()
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

// runVecLeafGate drives the shared verify jsx + render harness; the per-leaf
// assertions on the decoded frames are supplied by checkFrames.
func runVecLeafGate(t *testing.T, aeExe, ver, compName, leaf string, baseWhite bool, p *aep.Project,
	checkFrames func(t *testing.T, ver string, f0, f1, f2 image.Image)) {

	const argsPath = `e:/projects/tools/aep-parser/test_data/text_animator_vecleaf_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_animator_vecleaf.jsx`
	toFwd := func(s string) string { return strings.ReplaceAll(s, `\`, `/`) }

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "in.aep")
	resavedAEP := filepath.Join(tempDir, "resaved.aep")
	doneFile := filepath.Join(tempDir, "vl.done")
	png0 := filepath.Join(tempDir, "t0.png")
	png1 := filepath.Join(tempDir, "t1.png")
	png2 := filepath.Join(tempDir, "t2.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png0":%q,"png1":%q,"png2":%q,"comp":%q,"leaf":%q,"basewhite":%t}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png0), toFwd(png1), toFwd(png2), compName, leaf, baseWhite)
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 300)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("vec-leaf %s (%s) AE readback:\n%s", ver, leaf, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("vec-leaf %s (%s) ship gate FAIL:\n%s", ver, leaf, body)
	}

	decode := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("%s rendered frame missing (%s): %v", ver, filepath.Base(path), err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("%s decode %s: %v", ver, filepath.Base(path), err)
		}
		return img
	}
	checkFrames(t, ver, decode(png0), decode(png1), decode(png2))

	// Resave proof: leaf keeps 2 keyframes + Range Offset stays static.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("TXT") == nil {
		t.Fatalf("resaved: TXT missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, leaf)
	if kfl == nil {
		t.Fatalf("resaved: %s leaf not animated", leaf)
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved %s leaf numKf != 2", leaf)
	}
	if findShipList(root, "ADBE Text Percent Offset") != nil {
		t.Errorf("resaved: Range Offset unexpectedly animated (should stay static)")
	}
}

// ---- Fill Color leaf gate ----

func buildTextColorLeafDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p, tl := newTextVecLeafComp(t, target, "TXCOLLEAF")
	if _, err := aep.AddTextColorAnimator(tl, 1, 0, 0, 1, 0, 100, 0); err != nil {
		t.Fatalf("AddTextColorAnimator: %v", err)
	}
	// Keyframe the colour itself red → blue (full static selection).
	if err := aep.AnimateTextColor(tl, 0, []aep.VectorKeyframe{
		{Time: 0, Value: []float64{1, 0, 0, 1}},
		{Time: 2, Value: []float64{0, 0, 1, 1}},
	}); err != nil {
		t.Fatalf("AnimateTextColor: %v", err)
	}
	return reopenMoveBG(t, p)
}

func checkColorLeafFrames(t *testing.T, ver string, f0, f1, f2 image.Image) {
	sample := func(img image.Image) (r, g, b, n int) {
		return inkMeanColor(img, 120, 250, 1180, 540, 2, 15, 40)
	}
	r0, _, b0, n0 := sample(f0)
	r1, _, b1, n1 := sample(f1)
	r2, _, b2, n2 := sample(f2)
	t.Logf("%s ink mean R/B: t0=(R%d,B%d n=%d) t1=(R%d,B%d n=%d) t2=(R%d,B%d n=%d)", ver, r0, b0, n0, r1, b1, n1, r2, b2, n2)
	if n0 < 100 || n1 < 100 || n2 < 100 {
		t.Errorf("%s glyphs not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
	}
	// Red→blue: red channel falls, blue channel rises, both monotone with a gap.
	if !(r0 > r1 && r1 > r2) || r0-r2 < 60 {
		t.Errorf("%s red channel did not fall monotonically (r0=%d r1=%d r2=%d) — colour leaf not animating", ver, r0, r1, r2)
	}
	if !(b0 < b1 && b1 < b2) || b2-b0 < 60 {
		t.Errorf("%s blue channel did not rise monotonically (b0=%d b1=%d b2=%d) — colour leaf not animating", ver, b0, b1, b2)
	}
}

func runTextColorLeafGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	p := buildTextColorLeafDemo(t, target)
	runVecLeafGate(t, aeExe, ver, "TXCOLLEAF", "ADBE Text Fill Color", true, p, checkColorLeafFrames)
}

func TestTextColorLeafAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextColorLeafGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextColorLeafAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextColorLeafGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}

// ---- Position 3D leaf gate ----

func buildTextPosLeafDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p, tl := newTextVecLeafComp(t, target, "TXPOSLEAF")
	if _, err := aep.AddTextPositionAnimator(tl, 0, 0, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextPositionAnimator: %v", err)
	}
	// Keyframe the offset itself, gliding the glyphs downward.
	if err := aep.AnimateTextPosition(tl, 0, []aep.VectorKeyframe{
		{Time: 0, Value: []float64{0, 0, 0}},
		{Time: 2, Value: []float64{0, 250, 0}},
	}); err != nil {
		t.Fatalf("AnimateTextPosition: %v", err)
	}
	return reopenMoveBG(t, p)
}

func checkPosLeafFrames(t *testing.T, ver string, f0, f1, f2 image.Image) {
	cen := func(img image.Image) (float64, int) {
		_, cy, n := inkCentroid(img, 80, 60, 1200, 700, 2, 15, 40)
		return cy, n
	}
	y0, n0 := cen(f0)
	y1, n1 := cen(f1)
	y2, n2 := cen(f2)
	t.Logf("%s ink centroid Y: t0=%.0f(n=%d) t1=%.0f(n=%d) t2=%.0f(n=%d)", ver, y0, n0, y1, n1, y2, n2)
	if n0 < 100 || n1 < 100 || n2 < 100 {
		t.Errorf("%s glyphs not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
	}
	// Glyphs glide downward: vertical centroid migrates monotonically down with a
	// meaningful displacement (rules out a static / dropped leaf).
	if !(y0 < y1 && y1 < y2) || y2-y0 < 80 {
		t.Errorf("%s vertical centroid did not migrate down (y0=%.0f y1=%.0f y2=%.0f) — Position leaf not animating", ver, y0, y1, y2)
	}
}

func runTextPosLeafGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	p := buildTextPosLeafDemo(t, target)
	runVecLeafGate(t, aeExe, ver, "TXPOSLEAF", "ADBE Text Position 3D", false, p, checkPosLeafFrames)
}

func TestTextPosLeafAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextPosLeafGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextPosLeafAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextPosLeafGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
