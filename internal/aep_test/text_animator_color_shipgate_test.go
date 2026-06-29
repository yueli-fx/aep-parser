// internal/aep/text_animator_color_shipgate_test.go
//
// AE ship gate for the from-scratch Text Fill Color animator (kinetic typography
// colour wipe). Per delivery-contract red line 4 the gate renders THREE frames
// of the SAME from-scratch text layer and asserts on actual pixels that the
// glyphs' ink colour sweeps from the animator's override (red) to the base text
// colour (white) as the Range Selector slides off the characters:
//
//   - Fill-Color-red animator on "ABCDEF", base text forced white, Range
//     Start=0/End=100, Offset keyframed 0→100 over [t=0, t=2]. At offset 0 every
//     character is selected → tinted RED; as the offset sweeps to 100 the
//     selection slides off and the characters resolve to the base WHITE colour.
//
// The signature is the ink-pixel mean colour: the green channel of the rendered
// glyphs migrates monotonically up (red → white) while red stays high
// throughout. A colour-specific observable (not luminance/area/centroid): a
// static red glyph keeps green low on every frame; a dropped layer renders no
// ink. The resaved file is re-parsed to confirm the Fill Color survives AE's
// re-encode and the Range Offset stays animated.
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

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func buildTextColorDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXCOLGATE", 1280, 720, 24, 5)
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
	// Opaque red override over the full character range.
	if _, err := aep.AddTextColorAnimator(tl, 1, 0, 0, 1, 0, 100, 0); err != nil {
		t.Fatalf("AddTextColorAnimator: %v", err)
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
	return rp
}

// inkMeanColor returns the mean R,G,B (0..255) of the "ink" pixels (luminance
// differing from bgLum by more than thresh) over the search box, subsampled by
// step, plus the ink pixel count. Means are 0 when no ink.
func inkMeanColor(img image.Image, x0, y0, x1, y1, step, bgLum, thresh int) (int, int, int, int) {
	var sumR, sumG, sumB, n int
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(b>>8)
			l := (r8 + g8 + b8) / 3
			if d := l - bgLum; d > thresh || d < -thresh {
				sumR += r8
				sumG += g8
				sumB += b8
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0, 0, 0
	}
	return sumR / n, sumG / n, sumB / n, n
}

func runTextColorGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_animator_color_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_animator_color.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextColorDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txcol_in.aep")
	resavedAEP := filepath.Join(tempDir, "txcol_resaved.aep")
	doneFile := filepath.Join(tempDir, "txcol.done")
	png0 := filepath.Join(tempDir, "txcol_t0.png")
	png1 := filepath.Join(tempDir, "txcol_t1.png")
	png2 := filepath.Join(tempDir, "txcol_t2.png")

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
	t.Logf("text-color %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("text-color %s ship gate FAIL:\n%s", ver, body)
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
	sample := func(img image.Image) (r, g, b, n int) {
		return inkMeanColor(img, 120, 250, 1180, 540, 2, 15, 40)
	}
	r0, g0, b0, n0 := sample(decode(png0))
	r1, g1, b1, n1 := sample(decode(png1))
	r2, g2, b2, n2 := sample(decode(png2))
	t.Logf("%s ink mean RGB: t0=(%d,%d,%d n=%d) t1=(%d,%d,%d n=%d) t2=(%d,%d,%d n=%d)",
		ver, r0, g0, b0, n0, r1, g1, b1, n1, r2, g2, b2, n2)

	// Glyphs must render on every frame (the colour change keeps them visible).
	if n0 < 100 || n1 < 100 || n2 < 100 {
		t.Errorf("%s glyphs not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
	}

	// Red throughout (the override and the base both have a high red channel),
	// so this is a fill-colour sweep, not an opacity fade.
	if r0 < 120 || r2 < 120 {
		t.Errorf("%s red channel unexpectedly low (r0=%d r2=%d) — not a fill render?", ver, r0, r2)
	}
	// t=0 selected → RED: green far below red. t=2 deselected → WHITE: green
	// near red. The green channel migrates monotonically up across the sweep —
	// a colour-specific signature (a static red glyph keeps green low always).
	if g0 >= r0-60 {
		t.Errorf("%s t0 not red enough (r0=%d g0=%d) — Fill Color override not applied", ver, r0, g0)
	}
	if g2 < r2-50 {
		t.Errorf("%s t2 not white enough (r2=%d g2=%d) — colour did not resolve to base", ver, r2, g2)
	}
	if !(g0 < g1 && g1 < g2) || g2-g0 < 60 {
		t.Errorf("%s green channel did not sweep monotonically (g0=%d g1=%d g2=%d) — colour not animating", ver, g0, g1, g2)
	}

	// Resave proof: Fill Color survives + Range Offset stays animated.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("TXT") == nil {
		t.Fatalf("resaved: TXT missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	if followingList(root, "ADBE Text Fill Color") == nil {
		t.Fatalf("resaved: Fill Color dropped")
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

func TestTextColorAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextColorGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextColorAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextColorGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
