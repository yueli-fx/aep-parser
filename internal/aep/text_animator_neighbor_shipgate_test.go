// internal/aep/text_animator_neighbor_shipgate_test.go
//
// AE ship gate for the render-gateable "free-neighbor" Text Animator leaves —
// Fill Opacity, Stroke Opacity, Stroke Width, Stroke Color, Skew. Each shares
// the established splice + Range Selector machinery (1D scalar like Opacity, or
// 4-channel colour like Fill Color), so per delivery-contract red line 4 the
// gate's job is to render THREE frames of the SAME from-scratch text layer and
// assert on ACTUAL PIXELS that each leaf's visible surface changes as the Range
// Selector sweeps the per-character effect off the glyphs (offset 0→100 over
// t=0..2).
//
// One signature per leaf (its 作用面):
//   - Fill Opacity   → brightness spread (hidden fill → revealed text)
//   - Stroke Opacity → stroke ink area (hidden stroke → revealed stroke)
//   - Stroke Width   → stroke ink area (fat stroke → thin stroke)
//   - Stroke Color   → ink mean green channel (red stroke → white stroke)
//   - Skew           → glyph ink bbox aspect (sheared-wide → upright-narrow)
//
// Rotation X / Rotation Y are deliberately NOT here: their value is written and
// survives AE (see TestTextRotationXY_RoundTrip), but a per-character 3D rotation
// is visually inert in a plain 2D text layer — measured bbox identical across all
// three frames (160/160/160 height, 104/104/104 width) — because it needs
// Per-character 3D enabled (a separate 3D capability). Evidence-based defer; see
// incidents/text-animator-create-re.md.
//
// Each builds fill-only / stroke-only text as needed; the stroke leaves turn the
// fill off and a white stroke on (a sampling concern set in the verify JSX, not
// the tested write). Gated by AE_SHIP_GATE.
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

// brightnessSpread returns max-min mean-luminance over the subsampled box — a
// reveal signature (uniform BG → 0; visible glyphs → large).
func brightnessSpread(img image.Image, x0, y0, x1, y1, step int) int {
	mn, mx := 1<<30, -1
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			l := int((r>>8 + g>>8 + b>>8) / 3)
			if l < mn {
				mn = l
			}
			if l > mx {
				mx = l
			}
		}
	}
	if mx < 0 {
		return 0
	}
	return mx - mn
}

// inkMeanRGB returns the mean R/G/B (0-255) of "ink" pixels (luminance differing
// from bgLum by > thresh) over the subsampled box, plus the ink count.
func inkMeanRGB(img image.Image, x0, y0, x1, y1, step, bgLum, thresh int) (int, int, int, int) {
	var sr, sg, sb, n int
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(b>>8)
			l := (r8 + g8 + b8) / 3
			if d := l - bgLum; d > thresh || d < -thresh {
				sr += r8
				sg += g8
				sb += b8
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0, 0, 0
	}
	return sr / n, sg / n, sb / n, n
}

type neighborSpec struct {
	key   string // subtest name + comp name suffix
	text  string
	add   func(tl *aep.Layer) error
	args  string // extra JSON cosmetic fields, e.g. `,"fontSize":240,"posX":560,"posY":460`
	check func(t *testing.T, ver string, im0, im1, im2 image.Image)
}

func neighborSpecs() []neighborSpec {
	// Sample box covers the centred text region in the 1280×720 comp.
	const x0, y0, x1, y1, step = 120, 60, 1160, 680, 2
	const bgLum = 12 // 0.05 linear ≈ dark

	aspect := func(im image.Image) (float64, int) {
		w, h, n := inkBBox(im, x0, y0, x1, y1, step, bgLum, 40)
		if h == 0 {
			return 0, n
		}
		return float64(w) / float64(h), n
	}
	return []neighborSpec{
		{
			key:  "FillOpacity",
			text: "ABCDEF",
			add:  func(tl *aep.Layer) error { _, e := aep.AddTextFillOpacityAnimator(tl, 0, 0, 100, 0); return e },
			args: `,"fontSize":140,"posX":150,"posY":420`,
			check: func(t *testing.T, ver string, im0, im1, im2 image.Image) {
				s0 := brightnessSpread(im0, x0, y0, x1, y1, step)
				s1 := brightnessSpread(im1, x0, y0, x1, y1, step)
				s2 := brightnessSpread(im2, x0, y0, x1, y1, step)
				t.Logf("%s FillOpacity spread: t0=%d t1=%d t2=%d", ver, s0, s1, s2)
				if s2 < 80 {
					t.Errorf("%s text never revealed (t2 spread=%d)", ver, s2)
				}
				if s0 >= s2 {
					t.Errorf("%s fill opacity not revealing (t0=%d >= t2=%d)", ver, s0, s2)
				}
				if s2-s0 < 40 {
					t.Errorf("%s reveal too weak (t0=%d t2=%d)", ver, s0, s2)
				}
				if s1 < s0 || s1 > s2 {
					t.Errorf("%s mid frame not between ends (t0=%d t1=%d t2=%d)", ver, s0, s1, s2)
				}
			},
		},
		{
			key:  "Skew",
			text: "L",
			add:  func(tl *aep.Layer) error { _, e := aep.AddTextSkewAnimator(tl, 60, 0, 100, 0); return e },
			args: `,"fontSize":240,"posX":560,"posY":460`,
			check: func(t *testing.T, ver string, im0, im1, im2 image.Image) {
				a0, n0 := aspect(im0)
				a1, n1 := aspect(im1)
				a2, n2 := aspect(im2)
				t.Logf("%s Skew aspect(w/h): t0=%.2f(n=%d) t1=%.2f(n=%d) t2=%.2f(n=%d)", ver, a0, n0, a1, n1, a2, n2)
				if n0 < 100 || n1 < 100 || n2 < 100 {
					t.Errorf("%s glyph not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
				}
				if a0 <= a2 || a0 < a2*1.3 {
					t.Errorf("%s skew not shearing (t0 aspect=%.2f should exceed t2=%.2f by 1.3x)", ver, a0, a2)
				}
				if a1 > a0 || a1 < a2 {
					t.Errorf("%s mid aspect not between ends (t0=%.2f t1=%.2f t2=%.2f)", ver, a0, a1, a2)
				}
			},
		},
		{
			key:  "StrokeWidth",
			text: "ABC",
			add:  func(tl *aep.Layer) error { _, e := aep.AddTextStrokeWidthAnimator(tl, 40, 0, 100, 0); return e },
			args: `,"fontSize":180,"posX":260,"posY":430,"applyFill":false,"applyStroke":true,"strokeWidth":3,"strokeWhite":true`,
			check: func(t *testing.T, ver string, im0, im1, im2 image.Image) {
				_, _, n0 := inkBBox(im0, x0, y0, x1, y1, step, bgLum, 40)
				_, _, n1 := inkBBox(im1, x0, y0, x1, y1, step, bgLum, 40)
				_, _, n2 := inkBBox(im2, x0, y0, x1, y1, step, bgLum, 40)
				t.Logf("%s StrokeWidth ink-area: t0=%d t1=%d t2=%d", ver, n0, n1, n2)
				if n2 < 50 {
					t.Errorf("%s base stroke not rendered at t2 (n=%d)", ver, n2)
				}
				if n0 <= n2 || n0 < int(float64(n2)*1.3) {
					t.Errorf("%s stroke width not thickening selected (t0 area=%d should exceed t2=%d by 1.3x)", ver, n0, n2)
				}
				if n1 > n0 || n1 < n2 {
					t.Errorf("%s mid area not between ends (t0=%d t1=%d t2=%d)", ver, n0, n1, n2)
				}
			},
		},
		{
			key:  "StrokeColor",
			text: "ABC",
			add:  func(tl *aep.Layer) error { _, e := aep.AddTextStrokeColorAnimator(tl, 1, 0, 0, 1, 0, 100, 0); return e },
			args: `,"fontSize":180,"posX":260,"posY":430,"applyFill":false,"applyStroke":true,"strokeWidth":10,"strokeWhite":true`,
			check: func(t *testing.T, ver string, im0, im1, im2 image.Image) {
				r0, g0, b0, q0 := inkMeanRGB(im0, x0, y0, x1, y1, step, bgLum, 40)
				r1, g1, b1, q1 := inkMeanRGB(im1, x0, y0, x1, y1, step, bgLum, 40)
				r2, g2, b2, q2 := inkMeanRGB(im2, x0, y0, x1, y1, step, bgLum, 40)
				t.Logf("%s StrokeColor ink mean RGB: t0=(%d,%d,%d,n=%d) t1=(%d,%d,%d,n=%d) t2=(%d,%d,%d,n=%d)",
					ver, r0, g0, b0, q0, r1, g1, b1, q1, r2, g2, b2, q2)
				if q0 < 50 || q2 < 50 {
					t.Errorf("%s stroke not rendered on some frame (n0=%d n2=%d)", ver, q0, q2)
				}
				// red stroke (t0) → white stroke (t2): green channel rises, red stays high.
				if g2 <= g0 || g2-g0 < 40 {
					t.Errorf("%s stroke colour not shifting red→white (green t0=%d t2=%d)", ver, g0, g2)
				}
				if g1 < g0 || g1 > g2 {
					t.Errorf("%s mid green not between ends (t0=%d t1=%d t2=%d)", ver, g0, g1, g2)
				}
				if r0 < 80 {
					t.Errorf("%s t0 stroke not red (mean R=%d)", ver, r0)
				}
			},
		},
		{
			key:  "StrokeOpacity",
			text: "ABC",
			add:  func(tl *aep.Layer) error { _, e := aep.AddTextStrokeOpacityAnimator(tl, 0, 0, 100, 0); return e },
			args: `,"fontSize":180,"posX":260,"posY":430,"applyFill":false,"applyStroke":true,"strokeWidth":10,"strokeWhite":true`,
			check: func(t *testing.T, ver string, im0, im1, im2 image.Image) {
				_, _, n0 := inkBBox(im0, x0, y0, x1, y1, step, bgLum, 40)
				_, _, n1 := inkBBox(im1, x0, y0, x1, y1, step, bgLum, 40)
				_, _, n2 := inkBBox(im2, x0, y0, x1, y1, step, bgLum, 40)
				t.Logf("%s StrokeOpacity ink-area: t0=%d t1=%d t2=%d", ver, n0, n1, n2)
				if n2 < 50 {
					t.Errorf("%s stroke never revealed (t2 area=%d)", ver, n2)
				}
				if n0 >= n2 || n2 < int(float64(n0+1)*1.3) {
					t.Errorf("%s stroke opacity not revealing (t0 area=%d should be < t2=%d by 1.3x)", ver, n0, n2)
				}
				if n1 < n0 || n1 > n2 {
					t.Errorf("%s mid area not between ends (t0=%d t1=%d t2=%d)", ver, n0, n1, n2)
				}
			},
		},
	}
}

func buildNeighborDemo(t *testing.T, target aep.AETarget, compName, text string, add func(*aep.Layer) error) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, compName, 1280, 720, 24, 5)
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
	if err := tl.SetText(text); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := add(tl); err != nil {
		t.Fatalf("add animator: %v", err)
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

func runOneNeighborGate(t *testing.T, aeExe, ver string, target aep.AETarget, sp neighborSpec) {
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_animator_neighbor_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_animator_neighbor.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	compName := "TXNB_" + sp.key
	p := buildNeighborDemo(t, target, compName, sp.text, sp.add)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "nb_in.aep")
	resavedAEP := filepath.Join(tempDir, "nb_resaved.aep")
	doneFile := filepath.Join(tempDir, "nb.done")
	png0 := filepath.Join(tempDir, "nb_t0.png")
	png1 := filepath.Join(tempDir, "nb_t1.png")
	png2 := filepath.Join(tempDir, "nb_t2.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png0":%q,"png1":%q,"png2":%q,"comp":%q%s}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png0), toFwd(png1), toFwd(png2), compName, sp.args)
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
	t.Logf("%s %s AE readback:\n%s", sp.key, ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s %s ship gate FAIL:\n%s", sp.key, ver, body)
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
	sp.check(t, ver, decode(png0), decode(png1), decode(png2))
}

func runNeighborSuite(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	for _, sp := range neighborSpecs() {
		t.Run(sp.key, func(t *testing.T) {
			runOneNeighborGate(t, aeExe, ver, target, sp)
		})
	}
}

func TestTextNeighborAnimators_AEShipGate_AE2020(t *testing.T) {
	runNeighborSuite(t, ae2020(), "AE2020", aep.TargetAE2020)
}

func TestTextNeighborAnimators_AEShipGate_AE2025(t *testing.T) {
	runNeighborSuite(t, ae2025(), "AE2025", aep.TargetAE2025)
}
