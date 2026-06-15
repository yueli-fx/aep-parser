// internal/aep/text_animator_leaf_anim_shipgate_test.go
//
// AE ship gate for animating a text animator's DRIVEN LEAF over time (here the
// Rotation leaf), as opposed to sweeping the Range Selector Offset. Per delivery-
// contract red line 4 the gate renders THREE frames of the SAME from-scratch text
// layer and asserts on actual pixels that a single glyph's ink bounding box flips
// orientation as the leaf's own keyframed value sweeps:
//
//   - Rotation animator on a single big "L" with the Range Selector LEFT STATIC
//     (Start=0/End=100/Offset=0 → full select, no sweep); the Rotation leaf
//     itself is keyframed 0→90 over [t=0, t=2]. At t=0 the glyph is upright → its
//     ink bbox is TALL (w/h < 1); by t=2 the rotation reaches 90° → the "L" is
//     WIDE (w/h > 1). Every selected character shares the value curve.
//
// This is the inverse wiring of text_animator_rotation_shipgate_test.go (there
// the Offset sweeps and Rotation is static; here Rotation animates and the Offset
// is static), which is what proves AnimateTextRotation keyframes the leaf, not
// the selector. The resaved file is re-parsed to confirm the Rotation leaf keeps
// its 2 keyframes and the Range Offset stays static.
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

func buildTextRotLeafDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXROTLEAF", 1280, 720, 24, 5)
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
	if err := tl.SetText("L"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	// Full static selection; the Rotation LEAF (initial 0) is what we animate.
	if _, err := aep.AddTextRotationAnimator(tl, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextRotationAnimator: %v", err)
	}
	if err := aep.AnimateTextRotation(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 90}}); err != nil {
		t.Fatalf("AnimateTextRotation: %v", err)
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

func runTextRotLeafGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_animator_rotleaf_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_animator_rotleaf.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextRotLeafDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txrotleaf_in.aep")
	resavedAEP := filepath.Join(tempDir, "txrotleaf_resaved.aep")
	doneFile := filepath.Join(tempDir, "txrotleaf.done")
	png0 := filepath.Join(tempDir, "txrotleaf_t0.png")
	png1 := filepath.Join(tempDir, "txrotleaf_t1.png")
	png2 := filepath.Join(tempDir, "txrotleaf_t2.png")

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
	t.Logf("text-rotleaf %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("text-rotleaf %s ship gate FAIL:\n%s", ver, body)
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

	if n0 < 100 || n1 < 100 || n2 < 100 {
		t.Errorf("%s glyph not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
	}
	// Direction-agnostic orientation flip driven by the leaf keyframes.
	lo, hi := a0, a2
	if lo > hi {
		lo, hi = hi, lo
	}
	if lo <= 0 || hi < lo*1.5 {
		t.Errorf("%s glyph did not change orientation (t0=%.2f t2=%.2f) — leaf Rotation not animating", ver, a0, a2)
	}
	if a1 <= lo || a1 >= hi {
		t.Errorf("%s mid frame aspect not between ends (t0=%.2f t1=%.2f t2=%.2f) — leaf not sweeping monotonically", ver, a0, a1, a2)
	}

	// Resave proof: Rotation leaf keeps 2 keyframes + Range Offset stays static.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("TXT") == nil {
		t.Fatalf("resaved: TXT missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	rotKfl := findShipList(root, "ADBE Text Rotation")
	if rotKfl == nil {
		t.Fatalf("resaved: Rotation leaf dropped")
	}
	lhd3 := findShipChunk(rotKfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Rotation leaf numKf != 2")
	}
	if findShipList(root, "ADBE Text Percent Offset") != nil {
		t.Errorf("resaved: Range Offset unexpectedly animated (should stay static)")
	}
}

func TestTextRotLeafAnimator_AEShipGate_AE2020(t *testing.T) {
	runTextRotLeafGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextRotLeafAnimator_AEShipGate_AE2025(t *testing.T) {
	runTextRotLeafGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
