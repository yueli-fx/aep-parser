// internal/aep/mg_offset_copies_shipgate_test.go
//
// AE ship gate for Offset Paths Copies (priority-3 shape remaining,
// specs/2026-06-14-remaining-capability-roadmap.md). The `ADBE Vector Offset
// Copies` sub-stream is AE-default-elided (no slot in the Amount-only template);
// the serializer materializes it via synthesis-insert (splice the leaf into the
// offset body in canonical order, mirroring SetMaterialOption) when SetCopies is
// called.
//
// Verified at the capability's surface per delivery-contract red line 4: a
// 400×400 white rect + an even-odd Fill + Offset Paths (Amount=40, Copies=3)
// stacks 3 progressively-offset copies that read as concentric even-odd ring
// bands — white centre (inside all 3), a dark gap (inside 2 → even), a white
// outer ring (inside 1 → odd). The gate asserts the outer white ring exists
// (only extra copies can paint white that far out) AND the dark gap sits between
// it and the centre (distinguishing concentric rings from one big solid square)
// — proving Copies actually multiplied the path, not just that the value
// round-tripped.
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

	aep "github.com/example/aep-parser/internal/aep"
)

func buildMGOffsetCopiesDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "OFFCOPIES", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	// Dark BG so the ring bands read unambiguously.
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

	// Rect 400×400 (half-width 200) + even-odd Fill + Offset Amount=40, Copies=3.
	// Copies stack at half-widths 240 / 280 / 320; even-odd fill → white centre
	// (<240), dark gap (240..280), white ring (280..320). Stack [Rect, Fill,
	// Offset]: Offset is a distort filter (paint below it).
	card, err := aep.NewShapeLayer(comp, "CARD")
	if err != nil {
		t.Fatalf("NewShapeLayer CARD: %v", err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect CARD: %v", err)
	}
	if err := rect.SetSize([2]float64{400, 400}); err != nil {
		t.Fatalf("CARD rect SetSize: %v", err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill CARD: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("CARD SetColor: %v", err)
	}
	if err := fill.SetFillRule(aep.FillRuleEvenOdd); err != nil {
		t.Fatalf("CARD SetFillRule: %v", err)
	}
	off, err := card.RootGroup().AddOffsetPaths()
	if err != nil {
		t.Fatalf("AddOffsetPaths: %v", err)
	}
	if err := off.SetAmount(40); err != nil {
		t.Fatalf("SetAmount: %v", err)
	}
	if err := off.SetCopies(3); err != nil {
		t.Fatalf("SetCopies: %v", err)
	}
	if err := card.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("CARD Position: %v", err)
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

func runMGOffsetCopiesGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_offset_copies_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_offset_copies.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGOffsetCopiesDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_offset_copies_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_offset_copies_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_offset_copies.done")
	framePNG := filepath.Join(tempDir, "mg_offset_copies_frame.png")

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
	t.Logf("mg offset copies %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg offset copies %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0 (centre 960,540): concentric even-odd ring
	// bands from 3 offset copies (half-widths 240/280/320).
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 6
	// Outer white ring (d≈300, band 280..320): only extra copies paint white
	// this far out. 4 mid-edge points.
	ringPts := [][2]int{{660, 540}, {1260, 540}, {960, 240}, {960, 840}}
	// Dark gap (d≈260, band 240..280): the even copy-count region — distinguishes
	// concentric rings from one big solid square (which would be white here).
	gapPts := [][2]int{{700, 540}, {1220, 540}, {960, 280}, {960, 800}}
	// White centre (inside all 3 copies).
	centrePts := [][2]int{{960, 540}, {760, 540}, {1160, 540}}

	ring := 0
	for _, pt := range ringPts {
		if whiteNear(img, pt[0], pt[1], win) {
			ring++
		}
	}
	gap := 0
	for _, pt := range gapPts {
		if !whiteNear(img, pt[0], pt[1], win) {
			gap++
		}
	}
	centre := 0
	for _, pt := range centrePts {
		if whiteNear(img, pt[0], pt[1], win) {
			centre++
		}
	}
	t.Logf("%s offset copies: outer-ring white=%d/4 gap dark=%d/4 centre white=%d/3", ver, ring, gap, centre)
	if ring != 4 {
		t.Errorf("%s only %d/4 outer-ring points white — extra offset copies not rendered", ver, ring)
	}
	if gap != 4 {
		t.Errorf("%s only %d/4 gap points dark — no concentric ring (solid square, not multi-copy)", ver, gap)
	}
	if centre != 3 {
		t.Errorf("%s only %d/3 centre points white — base shape not rendered", ver, centre)
	}

	// Resave proof: the spliced Copies leaf survives AE's re-encode at value 3.
	root := parseAEP(t, resavedAEP)
	cdat := streamCdat(root, "ADBE Vector Offset Copies")
	if len(cdat) < 8 {
		t.Fatal("resaved: Offset Copies cdat missing — spliced leaf dropped")
	}
	if v := math.Float64frombits(binary.BigEndian.Uint64(cdat[:8])); math.Abs(v-3) > 0.01 {
		t.Errorf("resaved Copies = %g, want 3", v)
	}
}

func TestMGOffsetCopies_AEShipGate_AE2020(t *testing.T) {
	runMGOffsetCopiesGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGOffsetCopies_AEShipGate_AE2025(t *testing.T) {
	runMGOffsetCopiesGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
