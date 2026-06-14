// internal/aep/mg_opacity_shipgate_test.go
//
// AE ship gate for shape Fill / Stroke Opacity (priority-3 shape remaining,
// secondary sub-properties; specs/2026-06-14-remaining-capability-roadmap.md).
// The opacity write paths already shipped (lowerFillNode / lowerStrokeNode via
// lowerShapeScalar); this gate render-verifies that AE actually composites them.
//
// Per delivery-contract red line 4, verified at the capability's surface: 4 cards
// over a dark BG — a white fill at 100% and 50% Fill Opacity, and a white stroked
// ring at 100% and 50% Stroke Opacity. The gate samples each card's luminance and
// asserts the 100% cards render full white (~255) while the 50% cards render
// ~half (~135 = 0.5·255 + 0.5·BG) — proving the opacity sub-stream changed AE's
// composite, not just that the value round-tripped.
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
)

type opacityCard struct {
	name     string
	cx, cy   int
	sx, sy   int     // luminance sample point
	stroke   bool    // stroke ring (true) vs solid fill (false)
	opacity  float64 // 100 or 50
	wantFull bool    // expect ~255 (true) vs ~half (false)
}

// Fill cards sample at centre; stroke cards sample on the ring top (cy-110, ring
// radius 110). Layer half-luminance over the dark BG ≈ 135.
var opacityCards = []opacityCard{
	{"FILLFULL", 560, 320, 560, 320, false, 100, true},
	{"FILLHALF", 1360, 320, 1360, 320, false, 50, false},
	{"STRKFULL", 560, 760, 560, 650, true, 100, true},
	{"STRKHALF", 1360, 760, 1360, 650, true, 50, false},
}

func buildMGOpacityDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "OPACITY", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

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

	for _, c := range opacityCards {
		card, err := aep.NewShapeLayer(comp, c.name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", c.name, err)
		}
		if c.stroke {
			e, err := card.RootGroup().AddEllipse()
			if err != nil {
				t.Fatalf("%s AddEllipse: %v", c.name, err)
			}
			if err := e.SetSize([2]float64{220, 220}); err != nil {
				t.Fatalf("%s ellipse size: %v", c.name, err)
			}
			st, err := card.RootGroup().AddStroke()
			if err != nil {
				t.Fatalf("%s AddStroke: %v", c.name, err)
			}
			if err := st.SetColor([4]float64{1, 1, 1, 1}); err != nil {
				t.Fatalf("%s stroke color: %v", c.name, err)
			}
			if err := st.SetWidth(60); err != nil {
				t.Fatalf("%s stroke width: %v", c.name, err)
			}
			if err := st.SetOpacity(c.opacity); err != nil {
				t.Fatalf("%s stroke opacity: %v", c.name, err)
			}
		} else {
			r, err := card.RootGroup().AddRect()
			if err != nil {
				t.Fatalf("%s AddRect: %v", c.name, err)
			}
			if err := r.SetSize([2]float64{260, 260}); err != nil {
				t.Fatalf("%s rect size: %v", c.name, err)
			}
			fill, err := card.RootGroup().AddFill()
			if err != nil {
				t.Fatalf("%s AddFill: %v", c.name, err)
			}
			if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
				t.Fatalf("%s fill color: %v", c.name, err)
			}
			if err := fill.SetOpacity(c.opacity); err != nil {
				t.Fatalf("%s fill opacity: %v", c.name, err)
			}
		}
		if err := card.Position().SetStaticValue([2]float64{float64(c.cx), float64(c.cy)}); err != nil {
			t.Fatalf("%s position: %v", c.name, err)
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

// avgLumNear returns the average luminance (0..255) in a (2*win+1)² window.
func avgLumNear(img image.Image, px, py, win int) int {
	b := img.Bounds()
	sum, n := 0, 0
	for y := py - win; y <= py+win; y++ {
		for x := px - win; x <= px+win; x++ {
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			r, g, bb, _ := img.At(x, y).RGBA()
			sum += int((r>>8 + g>>8 + bb>>8) / 3)
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / n
}

func runMGOpacityGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_opacity_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_opacity.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGOpacityDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_opacity_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_opacity_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_opacity.done")
	framePNG := filepath.Join(tempDir, "mg_opacity_frame.png")

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
	t.Logf("mg opacity %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg opacity %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 8
	for _, c := range opacityCards {
		lum := avgLumNear(img, c.sx, c.sy, win)
		t.Logf("%s %s (opacity %g): sample lum=%d (wantFull=%v)", ver, c.name, c.opacity, lum, c.wantFull)
		if c.wantFull {
			if lum < 220 {
				t.Errorf("%s %s: lum=%d, want full ~255 (>220)", ver, c.name, lum)
			}
		} else {
			if lum < 90 || lum > 190 {
				t.Errorf("%s %s: lum=%d, want ~half [90,190] — opacity %g not composited", ver, c.name, lum, c.opacity)
			}
		}
	}

	// Resave proof: every card survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	for _, c := range opacityCards {
		if re.Compositions[0].LayerByName(c.name) == nil {
			t.Errorf("resaved: %s layer missing", c.name)
		}
	}
}

func TestMGOpacity_AEShipGate_AE2020(t *testing.T) {
	runMGOpacityGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGOpacity_AEShipGate_AE2025(t *testing.T) {
	runMGOpacityGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
