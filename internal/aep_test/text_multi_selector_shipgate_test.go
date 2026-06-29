// internal/aep/text_multi_selector_shipgate_test.go
//
// AddTextRangeSelector — add a 2nd+ Range Selector to a text animator (the "ADBE
// Text Selectors" indexed group). Multiple selectors combine per Mode (default
// Add = union).
//
// TestTextMultiSelector_RoundTrip (no AE): a 2-selector animator survives
// WriteAEP → Open with both selectors present.
//
// TestTextMultiSelector_AEShipGate_AE20{20,25} (red line 4): A/B differential in
// one frame proves the 2nd selector EXTENDS the selection. Both text layers carry
// an Opacity-0 animator with a first selector [0,50]; TXTB additionally gets a
// 2nd selector [50,100]. TXTA (one selector) hides only the first half → its
// second half stays visible (region spread high); TXTB (union of both) hides
// everything → its region is blank (spread ~0).
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
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// countSelectors returns the number of "ADBE Text Selector" children under the
// first "ADBE Text Selectors" group in the chunk tree.
func countSelectors(root *rifx.Chunk) int {
	sels := followingList(root, "ADBE Text Selectors")
	if sels == nil {
		return 0
	}
	n := 0
	for _, ch := range sels.Children {
		if ch.ID == rifx.IDTdmn && trimShipNUL(string(ch.Data)) == "ADBE Text Selector" {
			n++
		}
	}
	return n
}

func TestTextMultiSelector_RoundTrip(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "MSELRT", 1280, 720, 24, 5)
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
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 50, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}
	if _, err := aep.AddTextRangeSelector(tl, 50, 100, 0); err != nil {
		t.Fatalf("AddTextRangeSelector: %v", err)
	}

	path := filepath.Join(t.TempDir(), "mselrt.aep")
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
	if n := countSelectors(parseAEP(t, path)); n != 2 {
		t.Errorf("selector count after round-trip = %d, want 2", n)
	}
}

func buildTextMultiSelDemo(t *testing.T, target aep.AETarget) *aep.Project {
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

	for _, spec := range []struct {
		name   string
		second bool
	}{{"TXTA", false}, {"TXTB", true}} {
		tl, err := aep.NewTextLayer(comp, spec.name)
		if err != nil {
			t.Fatalf("NewTextLayer %s: %v", spec.name, err)
		}
		if err := tl.SetText("ABCDEF"); err != nil {
			t.Fatalf("SetText %s: %v", spec.name, err)
		}
		if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 50, 0); err != nil {
			t.Fatalf("AddTextOpacityAnimator %s: %v", spec.name, err)
		}
		if spec.second {
			if _, err := aep.AddTextRangeSelector(tl, 50, 100, 0); err != nil {
				t.Fatalf("AddTextRangeSelector %s: %v", spec.name, err)
			}
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

func runTextMultiSelGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_range_advanced_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_range_advanced.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildTextMultiSelDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "msel_in.aep")
	resavedAEP := filepath.Join(tempDir, "msel_resaved.aep")
	doneFile := filepath.Join(tempDir, "msel.done")
	png0 := filepath.Join(tempDir, "msel_t0.png")

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
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("%s multi-selector AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s multi-selector ship gate FAIL:\n%s", ver, body)
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
	// TXTA (1 selector [0,50]) at top → second half visible (spread high).
	// TXTB (selectors [0,50]+[50,100] union) at bottom → all hidden (spread ~0).
	topSpread := brightnessSpread(img, 120, 150, 1160, 360, 2)
	botSpread := brightnessSpread(img, 120, 420, 1160, 640, 2)
	t.Logf("%s region spread: top(A,1 selector)=%d bottom(B,2 selectors)=%d", ver, topSpread, botSpread)
	if topSpread < 80 {
		t.Errorf("%s TXTA (1 selector) should show its visible half but top spread=%d", ver, topSpread)
	}
	if botSpread > 60 {
		t.Errorf("%s TXTB (2 selectors union) should be fully hidden but bottom spread=%d", ver, botSpread)
	}
	if topSpread < botSpread+60 {
		t.Errorf("%s 2nd selector did not extend selection (top=%d bottom=%d should differ)", ver, topSpread, botSpread)
	}

	if _, err := aep.Open(resavedAEP); err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
}

func TestTextMultiSelector_AEShipGate_AE2020(t *testing.T) {
	runTextMultiSelGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}

func TestTextMultiSelector_AEShipGate_AE2025(t *testing.T) {
	runTextMultiSelGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
