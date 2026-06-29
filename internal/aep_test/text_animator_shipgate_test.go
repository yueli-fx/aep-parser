// internal/aep/text_animator_shipgate_test.go
//
// AE ship gate for from-scratch Text Animators (kinetic typography) — the first
// time a text animator is created entirely in Go (AddTextOpacityAnimator) and
// its Range Selector Offset is keyframed (AnimateTextRangeOffset) into a reveal
// sweep. Per delivery-contract red line 4 the gate renders THREE frames of the
// SAME from-scratch text layer and asserts on actual pixels that the text
// sweeps from hidden → revealed over time:
//
//   - Opacity-0 animator, Range Start=0/End=100, Offset keyframed 0→100 over
//     [t=0, t=2]. At offset 0 the whole text is selected → opacity 0 → INVISIBLE;
//     as the offset sweeps to 100 the selection window slides off, revealing the
//     characters → VISIBLE.
//
// The signature is COLOR- and POSITION-agnostic: a full-frame uniform BG means
// at the hidden end the frame is flat (luminance spread ≈ 0) and at the revealed
// end the glyphs introduce a large spread (max−min luminance), whatever the text
// colour or exact placement. The monotonic spread growth across three frames of
// one layer is the animation proof — it rules out a dropped layer (revealed end
// shows glyphs) and a static frame (hidden end is flat). The resaved file is
// re-parsed to confirm the Range Offset survives as a 2-keyframe bpk-48 1D
// non-spatial container.
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

func buildTextAnimDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXANIMGATE", 1280, 720, 24, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Full-frame uniform BG so the hidden frame reads as flat luminance.
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

	// From-scratch text layer + Opacity-0 animator with a Range Selector whose
	// Offset sweeps 0→100 (the reveal).
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
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
	// The text layer's Position is placed in-frame by the verify JSX (the Go
	// text-layer accessor doesn't expose Position; positioning is only a pixel-
	// sampling fixture concern, not the animator capability under test).
	return rp
}

// lumSpread returns the max−min average luminance over a grid of windows across
// [x0,x1]×[y0,y1]; large for a frame with glyphs over a flat BG, ~0 for a flat
// frame.
func lumSpread(img image.Image, x0, y0, x1, y1, step int) int {
	lo, hi := 1<<30, -1
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			l := avgLumNear(img, x, y, 3)
			if l < lo {
				lo = l
			}
			if l > hi {
				hi = l
			}
		}
	}
	if hi < 0 {
		return 0
	}
	return hi - lo
}

func runTextAnimGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_animator_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_animator.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextAnimDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txanim_in.aep")
	resavedAEP := filepath.Join(tempDir, "txanim_resaved.aep")
	doneFile := filepath.Join(tempDir, "txanim.done")
	png0 := filepath.Join(tempDir, "txanim_t0.png")
	png1 := filepath.Join(tempDir, "txanim_t1.png")
	png2 := filepath.Join(tempDir, "txanim_t2.png")

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
	t.Logf("text-anim %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("text-anim %s ship gate FAIL:\n%s", ver, body)
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
	// Scan most of the frame (uniform full-frame BG + text); the JSX places the
	// text in-frame at [300,380].
	spread := func(img image.Image) int { return lumSpread(img, 60, 60, 1220, 660, 12) }
	s0 := spread(decode(png0))
	s1 := spread(decode(png1))
	s2 := spread(decode(png2))
	t.Logf("%s luminance spreads: t0=%d t1=%d t2=%d", ver, s0, s1, s2)

	// One end hidden (flat BG), the other revealed (glyphs), and the middle in
	// between — direction-agnostic so the gate passes whichever way AE sweeps the
	// selection.
	lo, hi := s0, s2
	if lo > hi {
		lo, hi = hi, lo
	}
	if lo > 18 {
		t.Errorf("%s neither end is hidden (t0=%d t2=%d) — Opacity-0 animator not hiding text", ver, s0, s2)
	}
	if hi < 55 {
		t.Errorf("%s neither end is revealed (t0=%d t2=%d) — text never renders or animator hides it permanently", ver, s0, s2)
	}
	if s1 <= lo || s1 >= hi {
		t.Errorf("%s mid frame not between ends (t0=%d t1=%d t2=%d) — Range Offset not animating monotonically", ver, s0, s1, s2)
	}

	// Resave proof: Range Offset survives AE's re-encode as a 2-keyframe bpk-48
	// 1D non-spatial container.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("TXT") == nil {
		t.Fatalf("resaved: TXT missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Text Percent Offset")
	if kfl == nil {
		t.Fatalf("resaved Range Offset: keyframes dropped (not animated)")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Range Offset numKf != 2")
	}
	if lhd3 != nil && binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != 48 {
		t.Errorf("resaved Range Offset bpk = %d, want 48", binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	}
}

func TestTextAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextAnimGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextAnimGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
