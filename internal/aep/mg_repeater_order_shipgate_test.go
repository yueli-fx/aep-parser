// internal/aep/mg_repeater_order_shipgate_test.go
//
// AE ship gate for Repeater Order (priority-3 shape remaining,
// specs/2026-06-14-remaining-capability-roadmap.md). The `ADBE Vector Repeater
// Order` enum (Composite Below=1 default / Above=2) is AE-default-elided; the
// serializer materializes it via synthesis-insert, spliced BEFORE the Repeater
// Transform group (its canonical position, between Offset and Transform) when
// SetOrder selects Above.
//
// Order is a COMPOSITING-ORDER setting: it changes which copy stacks on top.
// With a single-color shape fill this has NO visible effect — compositing N
// same-color layers is commutative (for two layers of colour C, alphas a,b the
// result luminance C(a + b(1-a)) is symmetric in a,b). So there is no pixel to
// "discriminate"; the honest verification is (a) AE accepts the Above file and
// round-trips Order=2, and (b) Above renders IDENTICAL coverage to the default
// Below twin (confirming the spliced leaf neither corrupts the render nor — as
// theory predicts — changes it). This frame carries both an ABOVE card and a
// BELOW card with otherwise-identical repeaters (overlapping, scaled, opacity
// falloff) and asserts their white coverage matches within 2%.
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

// repeaterOrderCard adds a Rect150 + white Fill + Repeater (5 copies, overlapping
// + scaled + opacity falloff) card at (cx,540); order!=Below is applied if set.
func repeaterOrderCard(t *testing.T, comp *aep.Composition, name string, cx float64, order aep.RepeaterOrder) {
	t.Helper()
	card, err := aep.NewShapeLayer(comp, name)
	if err != nil {
		t.Fatalf("NewShapeLayer %s: %v", name, err)
	}
	rect, err := card.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("%s AddRect: %v", name, err)
	}
	if err := rect.SetSize([2]float64{150, 150}); err != nil {
		t.Fatalf("%s rect SetSize: %v", name, err)
	}
	fill, err := card.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("%s AddFill: %v", name, err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("%s SetColor: %v", name, err)
	}
	rep, err := card.RootGroup().AddRepeater()
	if err != nil {
		t.Fatalf("%s AddRepeater: %v", name, err)
	}
	if err := rep.Copies().SetStaticValue(5); err != nil {
		t.Fatalf("%s Copies: %v", name, err)
	}
	tr := rep.Transform()
	_ = tr.SetPosition([2]float64{90, 0})
	_ = tr.SetScale([2]float64{70, 70})
	_ = tr.SetStartOpacity(100)
	_ = tr.SetEndOpacity(20)
	if order != aep.RepeaterOrderBelow {
		if err := rep.SetOrder(order); err != nil {
			t.Fatalf("%s SetOrder: %v", name, err)
		}
	}
	if err := card.Position().SetStaticValue([2]float64{cx, 540}); err != nil {
		t.Fatalf("%s Position: %v", name, err)
	}
}

func buildMGRepeaterOrderDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "REPORDER", 1920, 1080, 30, 5)
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

	// BELOW (default, no leaf) at 560; ABOVE (Order=2, leaf spliced) at 1360 —
	// identical repeaters otherwise.
	repeaterOrderCard(t, comp, "BELOW", 560, aep.RepeaterOrderBelow)
	repeaterOrderCard(t, comp, "ABOVE", 1360, aep.RepeaterOrderAbove)

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

func runMGRepeaterOrderGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_repeater_order_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_repeater_order.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGRepeaterOrderDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_repeater_order_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_repeater_order_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_repeater_order.done")
	framePNG := `e:/projects/tools/aep-parser/tmp_debug/mg_repeater_order_` + ver + `.png`

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
	t.Logf("mg repeater order %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg repeater order %s ship gate FAIL:\n%s", ver, body)
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

	// Both cards span the same local region around their centre (repeater steps
	// right + shrinks). Compare white coverage: Order=Above must render the same
	// as the default Below (no visible effect — commutative compositing).
	_, _, belowN := whiteCentroid(img, 560-220, 560+320, 320, 760)
	_, _, aboveN := whiteCentroid(img, 1360-220, 1360+320, 320, 760)
	t.Logf("%s repeater order: BELOW white=%d ABOVE white=%d", ver, belowN, aboveN)
	if belowN < 1000 || aboveN < 1000 {
		t.Errorf("%s repeater not rendered (BELOW=%d ABOVE=%d)", ver, belowN, aboveN)
	}
	if belowN > 0 {
		diff := math.Abs(float64(aboveN-belowN)) / float64(belowN)
		if diff > 0.02 {
			t.Errorf("%s ABOVE coverage differs from BELOW by %.1f%% (>2%%) — Order=Above changed the render unexpectedly (n %d vs %d)",
				ver, diff*100, aboveN, belowN)
		}
	}

	// Resave proof: the spliced Order leaf survives AE's re-encode at value 2.
	root := parseAEP(t, resavedAEP)
	cdat := streamCdat(root, "ADBE Vector Repeater Order")
	if len(cdat) < 8 {
		t.Fatal("resaved: Repeater Order cdat missing — spliced leaf dropped")
	}
	if v := math.Float64frombits(binary.BigEndian.Uint64(cdat[:8])); math.Abs(v-2) > 0.01 {
		t.Errorf("resaved Repeater Order = %g, want 2 (Above)", v)
	}
}

func TestMGRepeaterOrder_AEShipGate_AE2020(t *testing.T) {
	runMGRepeaterOrderGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGRepeaterOrder_AEShipGate_AE2025(t *testing.T) {
	runMGRepeaterOrderGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
