// internal/aep/mg_offset_extras_shipgate_test.go
//
// AE ship gate for the three remaining Offset Paths sub-streams (priority-3 shape
// remaining, specs/2026-06-14-remaining-capability-roadmap.md): `ADBE Vector
// Offset Line Join` (enum), `Offset Miter Limit` (scalar), `Offset Copy Offset`
// (scalar). All AE-default-elided; the serializer materializes each via
// synthesis-insert, spliced into the Amount-only offset body in canonical order
// (Line Join → Miter Limit → Copies → Copy Offset) when its setter is used.
//
// Verified at the capability's surface per delivery-contract red line 4 with one
// frame of five cards on a dark BG:
//   - MITER    : Rect200 + Offset Amount=60 (default Miter join, limit 4) — the
//     outward offset extends each 90° corner to a sharp MITER POINT, so
//     the corner-tip pixel just past the original corner is WHITE.
//   - BEVEL    : same + Line Join=Bevel — the corner is cut flat, tip DARK.
//   - MITERLIM : same Miter join + Miter Limit=1 — the miter exceeds the limit and
//     is clipped to a bevel, tip DARK.
//   - COPY1/2  : Rect160 + Amount=30 + Copies=3, Copy Offset 1 (default) vs 2 —
//     a larger Copy Offset widens the per-copy step, so the outermost
//     outline (white extent) reaches farther for COPY2.
//
// Asserting the corner-tip fill (MITER white, BEVEL+MITERLIM dark) and the offset
// extent (COPY2 > COPY1) proves the spliced leaves changed AE's render, not just
// that the values round-tripped.
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

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// cornerCard adds a Rect200 + white Fill + Offset(Amount=60) card at (cx,cy);
// the caller may set Line Join / Miter Limit on the returned node before reopen.
func offsetCornerCard(t *testing.T, comp *aep.Composition, name string, cx float64) *aep.OffsetPathsNode {
	t.Helper()
	card, err := aep.NewShapeLayer(comp, name)
	if err != nil {
		t.Fatalf("NewShapeLayer %s: %v", name, err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("%s AddRect: %v", name, err)
	}
	if err := rect.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("%s rect SetSize: %v", name, err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("%s AddFill: %v", name, err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("%s SetColor: %v", name, err)
	}
	off, err := card.RootGroup().AddOffsetPaths()
	if err != nil {
		t.Fatalf("%s AddOffsetPaths: %v", name, err)
	}
	if err := off.SetAmount(60); err != nil {
		t.Fatalf("%s SetAmount: %v", name, err)
	}
	if err := card.Position().SetStaticValue([2]float64{cx, 300}); err != nil {
		t.Fatalf("%s Position: %v", name, err)
	}
	return off
}

// offsetCopyCard adds a Rect160 + white Fill + Offset(Amount=30, Copies=3) card at
// (cx,780); caller may set Copy Offset on the returned node.
func offsetCopyCard(t *testing.T, comp *aep.Composition, name string, cx float64) *aep.OffsetPathsNode {
	t.Helper()
	card, err := aep.NewShapeLayer(comp, name)
	if err != nil {
		t.Fatalf("NewShapeLayer %s: %v", name, err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("%s AddRect: %v", name, err)
	}
	if err := rect.SetSize([2]float64{160, 160}); err != nil {
		t.Fatalf("%s rect SetSize: %v", name, err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("%s AddFill: %v", name, err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("%s SetColor: %v", name, err)
	}
	off, err := card.RootGroup().AddOffsetPaths()
	if err != nil {
		t.Fatalf("%s AddOffsetPaths: %v", name, err)
	}
	if err := off.SetAmount(30); err != nil {
		t.Fatalf("%s SetAmount: %v", name, err)
	}
	if err := off.SetCopies(3); err != nil {
		t.Fatalf("%s SetCopies: %v", name, err)
	}
	if err := card.Position().SetStaticValue([2]float64{cx, 780}); err != nil {
		t.Fatalf("%s Position: %v", name, err)
	}
	return off
}

func buildMGOffsetExtrasDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "OFFEXTRAS", 1920, 1080, 30, 5)
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

	// Corner cards (top): MITER default, BEVEL = Line Join Bevel, MITERLIM = Miter
	// Limit 1 (clips the miter to a bevel).
	_ = offsetCornerCard(t, comp, "MITER", 480)
	bevel := offsetCornerCard(t, comp, "BEVEL", 960)
	if err := bevel.SetLineJoin(aep.StrokeLineJoinBevel); err != nil {
		t.Fatalf("BEVEL SetLineJoin: %v", err)
	}
	miterlim := offsetCornerCard(t, comp, "MITERLIM", 1440)
	if err := miterlim.SetMiterLimit(1); err != nil {
		t.Fatalf("MITERLIM SetMiterLimit: %v", err)
	}

	// Copy cards (bottom): COPY1 default Copy Offset, COPY2 Copy Offset 2.
	_ = offsetCopyCard(t, comp, "COPY1", 640)
	copy2 := offsetCopyCard(t, comp, "COPY2", 1280)
	if err := copy2.SetCopyOffset(2); err != nil {
		t.Fatalf("COPY2 SetCopyOffset: %v", err)
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

// cornerFill counts white samples inside the miter-tip triangle just past the
// card's bottom-right corner (card centre cx,300; rect half 100; offset 60 → miter
// point at +160,+160). A sharp Miter join fills these; a Bevel / clipped corner
// cuts them.
func cornerFill(img image.Image, cx int) int {
	pts := [][2]int{{cx + 145, 445}, {cx + 150, 450}, {cx + 152, 448}, {cx + 148, 452}}
	n := 0
	for _, pt := range pts {
		if whiteNear(img, pt[0], pt[1], 3) {
			n++
		}
	}
	return n
}

// rightExtent returns how far right of cx the white region reaches along the
// card's centre row (y=780). Used to compare Copy Offset step widths.
func rightExtent(img image.Image, cx int) int {
	for dx := 360; dx >= 0; dx-- {
		if whiteNear(img, cx+dx, 780, 3) {
			return dx
		}
	}
	return -1
}

func runMGOffsetExtrasGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_offset_extras_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_offset_extras.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGOffsetExtrasDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_offset_extras_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_offset_extras_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_offset_extras.done")
	framePNG := `e:/projects/tools/aep-parser/tmp_debug/mg_offset_extras_` + ver + `.png`

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
	t.Logf("mg offset extras %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg offset extras %s ship gate FAIL:\n%s", ver, body)
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

	miterCorner := cornerFill(img, 480)
	bevelCorner := cornerFill(img, 960)
	miterlimCorner := cornerFill(img, 1440)
	copy1Ext := rightExtent(img, 640)
	copy2Ext := rightExtent(img, 1280)
	t.Logf("%s offset extras: corner-fill MITER=%d/4 BEVEL=%d/4 MITERLIM=%d/4  copy-extent COPY1=%d COPY2=%d",
		ver, miterCorner, bevelCorner, miterlimCorner, copy1Ext, copy2Ext)

	// Line Join: the default Miter fills its corner tip; Bevel cuts it.
	if miterCorner < 3 {
		t.Errorf("%s MITER corner-fill=%d/4 — default miter point not rendered (offset inactive?)", ver, miterCorner)
	}
	if bevelCorner > 1 {
		t.Errorf("%s BEVEL corner-fill=%d/4 — corner not cut (Line Join=Bevel did not change render)", ver, bevelCorner)
	}
	// Miter Limit=1 clips the same corner to a bevel.
	if miterlimCorner > 1 {
		t.Errorf("%s MITERLIM corner-fill=%d/4 — corner not clipped (Miter Limit=1 did not change render)", ver, miterlimCorner)
	}
	// Copy Offset: a larger step widens the outermost outline.
	if copy1Ext < 0 || copy2Ext < 0 {
		t.Errorf("%s copy cards not rendered (COPY1=%d COPY2=%d)", ver, copy1Ext, copy2Ext)
	}
	// AE2020 first run measured COPY1=172, COPY2=202 (+30 = one Amount step);
	// require a clear >=15px margin.
	if copy2Ext < copy1Ext+15 {
		t.Errorf("%s COPY2 extent=%d not >= COPY1=%d + 15 — Copy Offset=2 did not widen the step", ver, copy2Ext, copy1Ext)
	}
}

func TestMGOffsetExtras_AEShipGate_AE2020(t *testing.T) {
	runMGOffsetExtrasGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGOffsetExtras_AEShipGate_AE2025(t *testing.T) {
	runMGOffsetExtrasGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
