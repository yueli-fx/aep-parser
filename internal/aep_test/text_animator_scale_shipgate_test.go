// internal/aep/text_animator_scale_shipgate_test.go
//
// AE ship gate for the from-scratch Text Scale 3D animator (kinetic typography
// shrink-in). Per delivery-contract red line 4 the gate renders THREE frames of
// the SAME from-scratch text layer and asserts on actual pixels that the glyph
// ink AREA changes as the Range Selector sweeps the per-character scale off the
// text:
//
//   - Scale-(220,220,100) animator, Range Start=0/End=100, Offset keyframed
//     0→100 over [t=0, t=2]. At offset 0 the whole text is selected → every
//     glyph at 220% → LARGE ink area; as the offset sweeps to 100 the selection
//     window slides off, the scale releases to 100% → SMALLER ink area.
//
// The signature is the count of rendered "ink" pixels (differing from the flat
// dark BG): it migrates monotonically across the three frames of one layer.
// Unlike the Opacity gate the text stays visible throughout — this proves a SIZE
// change, not appearance. The resaved file is re-parsed to confirm the Scale
// value survives AE's re-encode and the Range Offset stays a 2-keyframe stream.
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

func buildTextScaleDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXSCALEGATE", 1280, 720, 24, 5)
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
	if _, err := aep.AddTextScaleAnimator(tl, 220, 220, 100, 0, 100, 0); err != nil {
		t.Fatalf("AddTextScaleAnimator: %v", err)
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

// inkCount returns the number of "ink" pixels (luminance differing from bgLum by
// more than thresh) over the box, subsampled by step.
func inkCount(img image.Image, x0, y0, x1, y1, step, bgLum, thresh int) int {
	n := 0
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			l := int((r>>8 + g>>8 + b>>8) / 3)
			if d := l - bgLum; d > thresh || d < -thresh {
				n++
			}
		}
	}
	return n
}

func runTextScaleGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_animator_scale_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_animator_scale.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextScaleDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txscale_in.aep")
	resavedAEP := filepath.Join(tempDir, "txscale_resaved.aep")
	doneFile := filepath.Join(tempDir, "txscale.done")
	png0 := filepath.Join(tempDir, "txscale_t0.png")
	png1 := filepath.Join(tempDir, "txscale_t1.png")
	png2 := filepath.Join(tempDir, "txscale_t2.png")

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
	t.Logf("text-scale %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("text-scale %s ship gate FAIL:\n%s", ver, body)
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
	count := func(img image.Image) int { return inkCount(img, 80, 40, 1200, 700, 4, 15, 40) }
	n0 := count(decode(png0))
	n1 := count(decode(png1))
	n2 := count(decode(png2))
	t.Logf("%s ink counts: t0=%d t1=%d t2=%d", ver, n0, n1, n2)

	// Text must render on every frame (shrink-in keeps it visible throughout).
	if n0 < 100 || n1 < 100 || n2 < 100 {
		t.Errorf("%s text not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
	}

	// Direction-agnostic area change: one extreme large, the other small, the
	// middle strictly between. >= 30% growth rules out a static block.
	lo, hi := n0, n2
	if lo > hi {
		lo, hi = hi, lo
	}
	if hi < lo*13/10 {
		t.Errorf("%s glyph ink area did not change (t0=%d t2=%d) — Scale animator not resizing", ver, n0, n2)
	}
	if n1 <= lo || n1 >= hi {
		t.Errorf("%s mid frame ink not between ends (t0=%d t1=%d t2=%d) — scale not sweeping monotonically", ver, n0, n1, n2)
	}

	// Resave proof: Scale value survives + Range Offset stays animated.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("TXT") == nil {
		t.Fatalf("resaved: TXT missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	scTdbs := followingList(root, "ADBE Text Scale 3D")
	if scTdbs == nil {
		t.Fatalf("resaved: Scale 3D dropped")
	}
	if cdat := findShipChunk(scTdbs, rifx.IDCdat); cdat == nil || len(cdat.Data) < 8 {
		t.Errorf("resaved: Scale cdat missing/short")
	} else if x := math.Float64frombits(binary.BigEndian.Uint64(cdat.Data[0:8])); math.Abs(x-220) > 0.5 {
		t.Errorf("resaved: Scale x = %g, want 220", x)
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

func TestTextScaleAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextScaleGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextScaleAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextScaleGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
