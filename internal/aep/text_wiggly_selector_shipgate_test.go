// internal/aep/text_wiggly_selector_shipgate_test.go
//
// AddTextWigglySelector — a selector whose selection wobbles over time, so the
// characters flicker in/out under the animator with no keyframes.
//
// TestTextWigglySelector_RoundTrip (no AE): the wiggly selector survives
// WriteAEP → Open.
//
// TestTextWigglySelector_AEShipGate_AE20{20,25} (red line 4): the wiggle's
// signature is TEMPORAL VARIATION — render three frames of the SAME layer and
// assert they differ from each other (the wiggle re-selects characters over
// time). A single text layer with an Opacity-0 animator whose range is empty
// (End=0, so it selects nothing) plus a Wiggly Selector: the wiggle alone drives
// which characters are hidden, and that set changes frame to frame.
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
	"github.com/example/aep-parser/internal/rifx"
)

// frameDiff counts pixels whose luminance differs by > thresh between two images
// over the subsampled box.
func frameDiff(a, b image.Image, x0, y0, x1, y1, step, thresh int) int {
	n := 0
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			ra, ga, ba, _ := a.At(x, y).RGBA()
			rb, gb, bb, _ := b.At(x, y).RGBA()
			la := int((ra>>8 + ga>>8 + ba>>8) / 3)
			lb := int((rb>>8 + gb>>8 + bb>>8) / 3)
			if d := la - lb; d > thresh || d < -thresh {
				n++
			}
		}
	}
	return n
}

func buildTextWigglyDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXWIG", 1280, 720, 24, 5)
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
	if err := tl.SetText("ABCDEFGH"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	// Opacity-0 animator with an EMPTY range (End=0 selects nothing); the Wiggly
	// Selector (Add) then drives the selection alone.
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 0, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}
	if _, err := aep.AddTextWigglySelector(tl); err != nil {
		t.Fatalf("AddTextWigglySelector: %v", err)
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

func TestTextWigglySelector_RoundTrip(t *testing.T) {
	rp := buildTextWigglyDemo(t, aep.TargetAE2020)
	path := filepath.Join(t.TempDir(), "wig.aep")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(f); err != nil {
		f.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	f.Close()
	if _, err := aep.Open(path); err != nil {
		t.Fatalf("reopen: %v", err)
	}
	root := parseAEP(t, path)
	if findShipChunkTdmnName(root, "ADBE Text Wiggly Selector") == false {
		t.Errorf("Wiggly Selector dropped after round-trip")
	}
}

// findShipChunkTdmnName reports whether any tdmn in the tree has the given name.
func findShipChunkTdmnName(root *rifx.Chunk, name string) bool {
	found := false
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for _, ch := range c.Children {
			if ch.ID == rifx.IDTdmn && trimShipNUL(string(ch.Data)) == name {
				found = true
				return
			}
			if ch.IsList() {
				walk(ch)
				if found {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func runTextWigglyGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_render3_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_render3.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextWigglyDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "wig_in.aep")
	resavedAEP := filepath.Join(tempDir, "wig_resaved.aep")
	doneFile := filepath.Join(tempDir, "wig.done")
	png0 := filepath.Join(tempDir, "wig_t0.png")
	png1 := filepath.Join(tempDir, "wig_t1.png")
	png2 := filepath.Join(tempDir, "wig_t2.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png0":%q,"png1":%q,"png2":%q,"comp":"TXWIG","fontSize":140,"posX":120,"posY":420}`,
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
	t.Logf("%s wiggly AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s wiggly ship gate FAIL:\n%s", ver, body)
	}

	decode := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("%s frame missing (%s): %v", ver, filepath.Base(path), err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("%s decode %s: %v", ver, filepath.Base(path), err)
		}
		return img
	}
	const x0, y0, x1, y1, step = 80, 150, 1200, 680, 2
	im0, im1, im2 := decode(png0), decode(png1), decode(png2)
	// Each frame must show SOME text (wiggle is partial, never all-hidden/all-shown).
	s0 := brightnessSpread(im0, x0, y0, x1, y1, step)
	s2 := brightnessSpread(im2, x0, y0, x1, y1, step)
	d01 := frameDiff(im0, im1, x0, y0, x1, y1, step, 40)
	d12 := frameDiff(im1, im2, x0, y0, x1, y1, step, 40)
	d02 := frameDiff(im0, im2, x0, y0, x1, y1, step, 40)
	t.Logf("%s wiggly: spread t0=%d t2=%d ; frameDiff 0-1=%d 1-2=%d 0-2=%d", ver, s0, s2, d01, d12, d02)
	if s0 < 40 || s2 < 40 {
		t.Errorf("%s text not rendered (spread t0=%d t2=%d)", ver, s0, s2)
	}
	// Temporal variation: the wiggle re-selects characters → frames differ.
	maxDiff := d01
	if d12 > maxDiff {
		maxDiff = d12
	}
	if d02 > maxDiff {
		maxDiff = d02
	}
	if maxDiff < 200 {
		t.Errorf("%s frames do not vary over time (max frameDiff=%d) — wiggly selector not animating", ver, maxDiff)
	}

	if _, err := aep.Open(resavedAEP); err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
}

func TestTextWigglySelector_AEShipGate_AE2020(t *testing.T) {
	runTextWigglyGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}

func TestTextWigglySelector_AEShipGate_AE2025(t *testing.T) {
	runTextWigglyGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
