// internal/aep/text_range_advanced_shipgate_test.go
//
// SetTextRangeAdvanced — the Range Selector "Advanced" sub-params (Units / Based
// On / Mode / Amount / Shape / Smoothness / Ease High·Low / Randomize / Seed).
//
// TestTextRangeAdvanced_RoundTrip (no AE): all 10 params survive a WriteAEP →
// Open round-trip with the values set.
//
// TestTextRangeAdvancedAmount_AEShipGate_AE20{20,25} (red line 4): Amount is the
// param with a clean visual signature, gated by an A/B differential in ONE
// rendered frame — two text layers identical except Amount (both carry an
// Opacity-0 animator over the full range): TXTA Amount=100 (effect full → text
// hidden), TXTB Amount=20 (effect at 20% → text mostly visible). The frame must
// show TXTA's region dark and TXTB's region bright. The other params round-trip
// but are not individually render-gated (Mode needs multiple selectors,
// Smoothness only affects Shape=Square, the rest reshape selection internally).
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
)

func TestTextRangeAdvanced_RoundTrip(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "ADVRT", 1280, 720, 24, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}
	adv := aep.DefaultTextRangeAdvanced()
	adv.Units = 2
	adv.BasedOn = 3
	adv.Mode = 2
	adv.Amount = 40
	adv.Shape = 4
	adv.Smoothness = 60
	adv.EaseHigh = 25
	adv.EaseLow = 15
	adv.RandomizeOrder = 1
	adv.RandomSeed = 7
	if err := aep.SetTextRangeAdvanced(tl, adv); err != nil {
		t.Fatalf("SetTextRangeAdvanced: %v", err)
	}

	path := filepath.Join(t.TempDir(), "advrt.aep")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(f); err != nil {
		f.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	f.Close()
	if _, err := aep.Open(path); err != nil {
		t.Fatalf("reopen: %v", err)
	}

	root := parseAEP(t, path)
	want := []struct {
		mn  string
		val float64
	}{
		{"ADBE Text Range Units", 2},
		{"ADBE Text Range Type2", 3},
		{"ADBE Text Selector Mode", 2},
		{"ADBE Text Selector Max Amount", 40},
		{"ADBE Text Range Shape", 4},
		{"ADBE Text Selector Smoothness", 60},
		{"ADBE Text Levels Max Ease", 25},
		{"ADBE Text Levels Min Ease", 15},
		{"ADBE Text Randomize Order", 1},
		{"ADBE Text Random Seed", 7},
	}
	for _, w := range want {
		cdat := streamCdat(root, w.mn)
		if len(cdat) < 8 {
			t.Errorf("%s: cdat missing after round-trip", w.mn)
			continue
		}
		if got := math.Float64frombits(binary.BigEndian.Uint64(cdat[0:8])); math.Abs(got-w.val) > 0.01 {
			t.Errorf("%s = %v, want %v", w.mn, got, w.val)
		}
	}
}

func buildTextAmountDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXADV", 1280, 720, 24, 5)
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

	// Two text layers, opacity-0 animator over the full range, differing only in
	// Amount: A=100 (hidden), B=20 (mostly visible).
	for _, spec := range []struct {
		name   string
		amount float64
	}{{"TXTA", 100}, {"TXTB", 20}} {
		tl, err := aep.NewTextLayer(comp, spec.name)
		if err != nil {
			t.Fatalf("NewTextLayer %s: %v", spec.name, err)
		}
		if err := tl.SetText("ABCDEF"); err != nil {
			t.Fatalf("SetText %s: %v", spec.name, err)
		}
		if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 100, 0); err != nil {
			t.Fatalf("AddTextOpacityAnimator %s: %v", spec.name, err)
		}
		adv := aep.DefaultTextRangeAdvanced()
		adv.Amount = spec.amount
		if err := aep.SetTextRangeAdvanced(tl, adv); err != nil {
			t.Fatalf("SetTextRangeAdvanced %s: %v", spec.name, err)
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

func runTextAmountGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_range_advanced_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_range_advanced.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextAmountDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "txadv_in.aep")
	resavedAEP := filepath.Join(tempDir, "txadv_resaved.aep")
	doneFile := filepath.Join(tempDir, "txadv.done")
	png0 := filepath.Join(tempDir, "txadv_t0.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png0":%q,"comp":"TXADV"}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png0))
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
	t.Logf("%s Amount AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s Amount ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(png0)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("%s decode: %v", ver, err)
	}
	// TXTA at y~250 (top), TXTB at y~520 (bottom). Sample each region's spread.
	topSpread := brightnessSpread(img, 120, 150, 1160, 360, 2)
	botSpread := brightnessSpread(img, 120, 420, 1160, 640, 2)
	t.Logf("%s region spread: top(A,Amount=100)=%d bottom(B,Amount=20)=%d", ver, topSpread, botSpread)
	// A hidden (effect full) → top near-uniform; B mostly visible (effect 20%) → bottom high spread.
	if topSpread > 60 {
		t.Errorf("%s TXTA (Amount=100) should be hidden but top spread=%d", ver, topSpread)
	}
	if botSpread < 80 {
		t.Errorf("%s TXTB (Amount=20) should be visible but bottom spread=%d", ver, botSpread)
	}
	if botSpread < topSpread+60 {
		t.Errorf("%s Amount not scaling effect (top=%d bottom=%d should differ)", ver, topSpread, botSpread)
	}

	if _, err := aep.Open(resavedAEP); err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
}

func TestTextRangeAdvancedAmount_AEShipGate_AE2020(t *testing.T) {
	runTextAmountGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}

func TestTextRangeAdvancedAmount_AEShipGate_AE2025(t *testing.T) {
	runTextAmountGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
