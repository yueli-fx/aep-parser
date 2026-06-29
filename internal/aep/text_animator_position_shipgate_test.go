// internal/aep/text_animator_position_shipgate_test.go
//
// AE ship gate for the from-scratch Text Position 3D animator (kinetic
// typography slide-in). Per delivery-contract red line 4 the gate renders THREE
// frames of the SAME from-scratch text layer and asserts on actual pixels that
// the glyph block slides VERTICALLY as the Range Selector sweeps the per-
// character displacement off the text:
//
//   - Position-(0,-260,0) animator, Range Start=0/End=100, Offset keyframed
//     0→100 over [t=0, t=2]. At offset 0 the whole text is selected → every
//     glyph lifted 260px → text HIGH; as the offset sweeps to 100 the selection
//     window slides off, the displacement releases, and the glyphs settle to
//     baseline → text LOW.
//
// The signature is the vertical centroid of the rendered ink (pixels differing
// from the flat dark BG): the centroid migrates monotonically down the frame
// across the three frames of one layer. That rules out a dropped layer (no ink
// at all) and a static frame (constant centroid), and unlike the Opacity gate
// the text stays visible throughout — this proves MOTION, not appearance. The
// resaved file is re-parsed to confirm the Position 3D offset survives AE's
// re-encode and the Range Offset stays a 2-keyframe stream.
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
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func buildTextPosDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXPOSGATE", 1280, 720, 24, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Full-frame uniform dark BG so rendered glyphs read as distinct "ink".
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

	// From-scratch text layer + Position-(0,-260,0) animator with a Range
	// Selector whose Offset sweeps 0→100 (the slide-in).
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if _, err := aep.AddTextPositionAnimator(tl, 0, -260, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextPositionAnimator: %v", err)
	}
	if err := aep.AnimateTextRangeOffset(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}}); err != nil {
		t.Fatalf("AnimateTextRangeOffset: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c2 := rp.Compositions[0]
	if err := aep.MoveToEnd(c2.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	// The text layer's Position is placed in-frame by the verify JSX.
	return rp
}

// inkCentroidY returns the y-centroid of the "ink" (pixels whose luminance
// differs from bgLum by more than thresh) over the box, plus the ink pixel
// count. step subsamples for speed.
func inkCentroidY(img image.Image, x0, y0, x1, y1, step, bgLum, thresh int) (int, int) {
	var sumY, n int
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			l := int((r>>8 + g>>8 + b>>8) / 3)
			if d := l - bgLum; d > thresh || d < -thresh {
				sumY += y
				n++
			}
		}
	}
	if n == 0 {
		return -1, 0
	}
	return sumY / n, n
}

func runTextPosGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_animator_position_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_animator_position.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextPosDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txpos_in.aep")
	resavedAEP := filepath.Join(tempDir, "txpos_resaved.aep")
	doneFile := filepath.Join(tempDir, "txpos.done")
	png0 := filepath.Join(tempDir, "txpos_t0.png")
	png1 := filepath.Join(tempDir, "txpos_t1.png")
	png2 := filepath.Join(tempDir, "txpos_t2.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png0":%q,"png1":%q,"png2":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png0), toFwd(png1), toFwd(png2))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("text-pos %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("text-pos %s ship gate FAIL:\n%s", ver, body)
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
	// BG luminance ≈ (0.05+0.05+0.08)/3 in 8-bit ≈ 15; ink = glyphs over it.
	centroid := func(img image.Image) (int, int) { return inkCentroidY(img, 80, 40, 1200, 700, 4, 15, 40) }
	c0, n0 := centroid(decode(png0))
	c1, n1 := centroid(decode(png1))
	c2, n2 := centroid(decode(png2))
	t.Logf("%s ink centroidY: t0=%d(n=%d) t1=%d(n=%d) t2=%d(n=%d)", ver, c0, n0, c1, n1, c2, n2)

	// Text must render (ink present) on every frame — slide-in keeps it visible
	// throughout, unlike the Opacity reveal.
	if n0 < 200 || n1 < 200 || n2 < 200 {
		t.Errorf("%s text not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
	}

	// Direction-agnostic vertical migration: one extreme high, the other low,
	// the middle strictly between. >= 60px separation rules out a static block.
	lo, hi := c0, c2
	if lo > hi {
		lo, hi = hi, lo
	}
	if hi-lo < 60 {
		t.Errorf("%s glyph block did not slide vertically (t0=%d t2=%d, Δ=%d) — Position animator not displacing", ver, c0, c2, hi-lo)
	}
	if c1 <= lo || c1 >= hi {
		t.Errorf("%s mid frame centroid not between ends (t0=%d t1=%d t2=%d) — displacement not sweeping monotonically", ver, c0, c1, c2)
	}

	// Resave proof: Position 3D offset survives + Range Offset stays animated.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("TXT") == nil {
		t.Fatalf("resaved: TXT missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	// Position 3D is a static value (cdat, no keyframe stream) — use the
	// tdmn→following-LIST walk, not findShipList (which only finds kfl containers).
	posTdbs := followingList(root, "ADBE Text Position 3D")
	if posTdbs == nil {
		t.Fatalf("resaved: Position 3D dropped")
	}
	if cdat := findShipChunk(posTdbs, rifx.IDCdat); cdat == nil || len(cdat.Data) < 24 {
		t.Errorf("resaved: Position cdat missing/short")
	} else if y := math.Float64frombits(binary.BigEndian.Uint64(cdat.Data[8:16])); math.Abs(y-(-260)) > 0.5 {
		t.Errorf("resaved: Position y = %g, want -260", y)
	}
	kfl := findShipList(root, "ADBE Text Percent Offset")
	if kfl == nil {
		t.Fatalf("resaved Range Offset: keyframes dropped (not animated)")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Range Offset numKf != 2")
	}
}

func TestTextPosAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextPosGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextPosAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextPosGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
