// internal/aep/mg_trimanim_shipgate_test.go
//
// AE ship gate for ANIMATED Trim Paths (`ADBE Vector Filter - Trim` with a
// keyframed `ADBE Vector Trim End`) — the canonical line-draw reveal. The
// static Trim gate (mg_trim_shipgate_test.go) proved the filter; the animated
// code path (lowerShapeScalar → injectAnimatedStream, shared with Rect
// Roundness / Stroke Opacity) was wired but never gated at its surface. Per
// delivery-contract red line 4 the gate renders THREE frames of the SAME layer
// and asserts on actual pixels that the stroked ring sweeps open over time:
//
//   - t=0  (End=0):   ring absent everywhere (trim cuts the whole stroke).
//   - t=2  (End=50):  AE's ellipse path starts at top and winds clockwise, so
//                     0..50% reveals the RIGHT half (top→right→bottom present,
//                     left absent).
//   - t=4  (End=100): a complete ring (left AND right present).
//
// The monotonic left-side reveal (absent→absent→present) across three frames
// of one layer is the animation proof — it rules out both a dropped layer and
// a static End=100. The resaved file is re-parsed to confirm Trim End survives
// as a 2-keyframe 1D non-spatial container (bpk-48).
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

func buildMGTrimAnimDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MGTRIMANIM", 1920, 1080, 30, 5)
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
	if err := rect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("rect SetSize: %v", err)
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

	// REVEAL: ellipse + white stroke + a Trim filter whose End sweeps 0→100
	// over the 4s [t=0,t=4] window. Add order [Ellipse, Stroke, Trim] matches
	// the static fixture (bottom-up path→stroke→trim, so the trim cuts the
	// stroked ring).
	rl, err := aep.NewShapeLayer(comp, "REVEAL")
	if err != nil {
		t.Fatalf("NewShapeLayer REVEAL: %v", err)
	}
	el, err := rl.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse: %v", err)
	}
	if err := el.SetSize([2]float64{320, 320}); err != nil {
		t.Fatalf("ellipse SetSize: %v", err)
	}
	st, err := rl.RootGroup().AddStroke()
	if err != nil {
		t.Fatalf("AddStroke: %v", err)
	}
	if err := st.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("stroke SetColor: %v", err)
	}
	if err := st.SetWidth(24); err != nil {
		t.Fatalf("stroke SetWidth: %v", err)
	}
	tr, err := rl.RootGroup().AddTrim()
	if err != nil {
		t.Fatalf("AddTrim: %v", err)
	}
	if err := tr.End().AddKeyframeLinear(0, 0); err != nil {
		t.Fatalf("Trim End kf0: %v", err)
	}
	if err := tr.End().AddKeyframeLinear(4, 100); err != nil {
		t.Fatalf("Trim End kf4: %v", err)
	}
	if err := rl.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("REVEAL Position: %v", err)
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

func runMGTrimAnimGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_trimanim_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_trimanim.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGTrimAnimDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_trimanim_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_trimanim_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_trimanim.done")
	pngEarly := filepath.Join(tempDir, "mg_trimanim_t0.png")
	pngMid := filepath.Join(tempDir, "mg_trimanim_t2.png")
	pngFull := filepath.Join(tempDir, "mg_trimanim_t4.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"pngEarly":%q,"pngMid":%q,"pngFull":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(pngEarly), toFwd(pngMid), toFwd(pngFull))
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
	t.Logf("mg trim-anim %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg trim-anim %s ship gate FAIL:\n%s", ver, body)
	}

	const R, win = 160, 10
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
	type ring struct{ top, right, bottom, left bool }
	sample := func(img image.Image) ring {
		return ring{
			top:    whiteNear(img, 960, 540-R, win),
			right:  whiteNear(img, 960+R, 540, win),
			bottom: whiteNear(img, 960, 540+R, win),
			left:   whiteNear(img, 960-R, 540, win),
		}
	}
	early := sample(decode(pngEarly))
	mid := sample(decode(pngMid))
	full := sample(decode(pngFull))
	t.Logf("%s rings: early{T%v R%v B%v L%v} mid{T%v R%v B%v L%v} full{T%v R%v B%v L%v}",
		ver, early.top, early.right, early.bottom, early.left,
		mid.top, mid.right, mid.bottom, mid.left,
		full.top, full.right, full.bottom, full.left)

	// t=4 End=100: a complete ring (both sides present).
	if !full.left || !full.right {
		t.Errorf("%s t=4 ring incomplete (left=%v right=%v) — stroke/ellipse not rendering or trim stuck cut", ver, full.left, full.right)
	}
	// t=0 End=0: ring cut away (left+right+bottom absent; the top start vertex
	// may keep a dot, so it is not asserted).
	if early.left || early.right || early.bottom {
		t.Errorf("%s t=0 ring still painted (left=%v right=%v bottom=%v) — Trim End did not start at 0", ver, early.left, early.right, early.bottom)
	}
	// t=2 End=50: right half revealed, left still absent.
	if !mid.right || !mid.top || !mid.bottom {
		t.Errorf("%s t=2 right half missing (top=%v right=%v bottom=%v) — sweep not reaching 50%%", ver, mid.top, mid.right, mid.bottom)
	}
	if mid.left {
		t.Errorf("%s t=2 left side already painted — End reached >50%% too early (not animating 0→100)", ver)
	}

	// Resave proof: Trim End survives AE's re-encode as a 2-keyframe bpk-48 1D
	// non-spatial container.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	if re.Compositions[0].LayerByName("REVEAL") == nil {
		t.Fatalf("resaved: REVEAL missing (layers=%d)", len(re.Compositions[0].Layers))
	}
	root := parseAEP(t, resavedAEP)
	kfl := findShipList(root, "ADBE Vector Trim End")
	if kfl == nil {
		t.Fatalf("resaved Trim End: keyframes dropped (not animated)")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Errorf("resaved Trim End numKf != 2")
	}
	if lhd3 != nil && binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != 48 {
		t.Errorf("resaved Trim End bpk = %d, want 48", binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	}
}

func TestMGTrimAnim_AEShipGate_AE2020(t *testing.T) {
	runMGTrimAnimGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGTrimAnim_AEShipGate_AE2025(t *testing.T) {
	runMGTrimAnimGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
