// internal/aep/mg_mask_opacity_shipgate_test.go
//
// AE ship gate for Mask Opacity (priority-4 mask,
// specs/2026-06-14-remaining-capability-roadmap.md). The `ADBE Mask Opacity`
// leaf is AE-default-elided; SetOpacity materializes it via synthesis-insert into
// the mask atom group (mutate_mask_options.go — masks are eagerly decoded so the
// spliced leaf bytes are AE-native exact).
//
// Per delivery-contract red line 4, verified at the capability's surface: two
// white 600×600 shape rects, each with a coincident rectangular mask (Add mode),
// one mask at Opacity 100% and one at 50%, over a dark BG. The gate samples each
// revealed square's luminance and asserts the 100% mask reveals full white (~255)
// while the 50% mask dims to ~half (~135) — proving Mask Opacity actually scaled
// AE's reveal, not just that the value round-tripped.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

type maskOpCard struct {
	name     string
	cx, cy   int
	opacity  float64 // 1.0 or 0.5
	wantFull bool
}

var maskOpCards = []maskOpCard{
	{"MASKFULL", 560, 540, 1.0, true},
	{"MASKHALF", 1360, 540, 0.5, false},
}

func buildMGMaskOpacityDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MASKOP", 1920, 1080, 30, 5)
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

	for _, c := range maskOpCards {
		card, err := aep.NewShapeLayer(comp, c.name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", c.name, err)
		}
		r, err := card.RootGroup().AddRect()
		if err != nil {
			t.Fatalf("%s AddRect: %v", c.name, err)
		}
		if err := r.SetSize([2]float64{600, 600}); err != nil {
			t.Fatalf("%s rect size: %v", c.name, err)
		}
		fill, err := card.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("%s AddFill: %v", c.name, err)
		}
		if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
			t.Fatalf("%s fill color: %v", c.name, err)
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

	// Masks coincident with each 600×600 rect (local square −300..300).
	maskRect := aep.BezierPath{
		Vertices: [][2]float64{{-300, -300}, {300, -300}, {300, 300}, {-300, 300}},
		Closed:   true,
	}
	for _, c := range maskOpCards {
		l := rp.Compositions[0].LayerByName(c.name)
		if l == nil {
			t.Fatalf("%s layer missing after reopen", c.name)
		}
		m, err := aep.AddMask(l, "M", maskRect)
		if err != nil {
			t.Fatalf("%s AddMask: %v", c.name, err)
		}
		if err := m.SetOpacity(c.opacity); err != nil {
			t.Fatalf("%s SetOpacity: %v", c.name, err)
		}
	}
	return rp
}

func runMGMaskOpacityGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_mask_opacity_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_mask_opacity.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGMaskOpacityDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_mask_opacity_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_mask_opacity_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_mask_opacity.done")
	framePNG := filepath.Join(tempDir, "mg_mask_opacity_frame.png")

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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("mg mask opacity %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg mask opacity %s ship gate FAIL:\n%s", ver, body)
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
	for _, c := range maskOpCards {
		lum := avgLumNear(img, c.cx, c.cy, win)
		t.Logf("%s %s (mask opacity %g): sample lum=%d (wantFull=%v)", ver, c.name, c.opacity, lum, c.wantFull)
		if c.wantFull {
			if lum < 220 {
				t.Errorf("%s %s: lum=%d, want full ~255 (>220)", ver, c.name, lum)
			}
		} else {
			if lum < 90 || lum > 190 {
				t.Errorf("%s %s: lum=%d, want ~half [90,190] — mask opacity %g not composited", ver, c.name, lum, c.opacity)
			}
		}
	}

	// Resave proof: the spliced Mask Opacity leaf survives AE's re-encode at the
	// requested value (read back through our own parser).
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	for _, c := range maskOpCards {
		l := re.Compositions[0].LayerByName(c.name)
		if l == nil || len(l.Masks) != 1 {
			t.Errorf("resaved: %s mask missing", c.name)
			continue
		}
		if got := l.Masks[0].Opacity; math.Abs(got-c.opacity) > 0.01 {
			t.Errorf("resaved %s mask opacity = %v, want %v", c.name, got, c.opacity)
		}
	}
}

func TestMGMaskOpacity_AEShipGate_AE2020(t *testing.T) {
	runMGMaskOpacityGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGMaskOpacity_AEShipGate_AE2025(t *testing.T) {
	runMGMaskOpacityGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
