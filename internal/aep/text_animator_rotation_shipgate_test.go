// internal/aep/text_animator_rotation_shipgate_test.go
//
// AE ship gate for the from-scratch Text Rotation animator (kinetic typography
// spin-in). Per delivery-contract red line 4 the gate renders THREE frames of
// the SAME from-scratch text layer and asserts on actual pixels that a single
// glyph's ink bounding box flips orientation as the Range Selector sweeps the
// per-character rotation off it:
//
//   - Rotation-90 animator on a single big "L", Range Start=0/End=100, Offset
//     keyframed 0→100 over [t=0, t=2]. At offset 0 the glyph is selected →
//     rotated 90° → its ink bbox is WIDE (w/h > 1); as the offset sweeps to 100
//     the selection slides off and the rotation releases to 0° → the "L" is TALL
//     (w/h < 1).
//
// The signature is the glyph's ink bounding-box aspect ratio (width/height): it
// migrates monotonically across the three frames. Rotation changes neither area
// nor luminance nor centroid, so a single asymmetric glyph's bbox aspect is the
// orientation-sensitive observable. The resaved file is re-parsed to confirm the
// Rotation value survives AE's re-encode and the Range Offset stays animated.
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

func buildTextRotDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXROTGATE", 1280, 720, 24, 5)
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
	// Single asymmetric glyph so its ink bbox aspect is a clean orientation cue.
	if err := tl.SetText("L"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if _, err := aep.AddTextRotationAnimator(tl, 90, 0, 100, 0); err != nil {
		t.Fatalf("AddTextRotationAnimator: %v", err)
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

// inkBBox returns the bounding box (w, h) and pixel count of the "ink" (pixels
// whose luminance differs from bgLum by more than thresh) over the search box,
// subsampled by step. w/h is 0,0,0 when no ink.
func inkBBox(img image.Image, x0, y0, x1, y1, step, bgLum, thresh int) (int, int, int) {
	minX, minY, maxX, maxY, n := 1<<30, 1<<30, -1, -1, 0
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			l := int((r>>8 + g>>8 + b>>8) / 3)
			if d := l - bgLum; d > thresh || d < -thresh {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0, 0
	}
	return maxX - minX, maxY - minY, n
}

func runTextRotGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_animator_rotation_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_animator_rotation.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextRotDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txrot_in.aep")
	resavedAEP := filepath.Join(tempDir, "txrot_resaved.aep")
	doneFile := filepath.Join(tempDir, "txrot.done")
	png0 := filepath.Join(tempDir, "txrot_t0.png")
	png1 := filepath.Join(tempDir, "txrot_t1.png")
	png2 := filepath.Join(tempDir, "txrot_t2.png")

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
	t.Logf("text-rot %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("text-rot %s ship gate FAIL:\n%s", ver, body)
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
	aspect := func(img image.Image) (float64, int) {
		w, h, n := inkBBox(img, 120, 60, 1160, 680, 2, 15, 40)
		if h == 0 {
			return 0, n
		}
		return float64(w) / float64(h), n
	}
	a0, n0 := aspect(decode(png0))
	a1, n1 := aspect(decode(png1))
	a2, n2 := aspect(decode(png2))
	t.Logf("%s ink bbox aspect (w/h): t0=%.2f(n=%d) t1=%.2f(n=%d) t2=%.2f(n=%d)", ver, a0, n0, a1, n1, a2, n2)

	// Glyph must render on every frame (rotation keeps it visible throughout).
	if n0 < 100 || n1 < 100 || n2 < 100 {
		t.Errorf("%s glyph not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
	}

	// Direction-agnostic orientation flip: one extreme wide (aspect high), the
	// other tall (aspect low), the middle strictly between. 1.5x ratio rules out
	// a static (unrotating) glyph.
	lo, hi := a0, a2
	if lo > hi {
		lo, hi = hi, lo
	}
	if lo <= 0 || hi < lo*1.5 {
		t.Errorf("%s glyph did not change orientation (t0=%.2f t2=%.2f) — Rotation animator not rotating", ver, a0, a2)
	}
	if a1 <= lo || a1 >= hi {
		t.Errorf("%s mid frame aspect not between ends (t0=%.2f t1=%.2f t2=%.2f) — rotation not sweeping monotonically", ver, a0, a1, a2)
	}

	// Resave proof: Rotation value survives + Range Offset stays animated.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("TXT") == nil {
		t.Fatalf("resaved: TXT missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	rotTdbs := followingList(root, "ADBE Text Rotation")
	if rotTdbs == nil {
		t.Fatalf("resaved: Rotation dropped")
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

func TestTextRotAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextRotGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextRotAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextRotGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
