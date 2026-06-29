// internal/aep/text_expressible_selector_shipgate_test.go
//
// AddTextExpressibleSelector — an expression-driven selector whose per-character
// selection percentage comes from an ExtendScript expression (reading textIndex /
// textTotal / time / selectorValue). Closes the last text-animator backlog item;
// the prior "evidence-defer" rationale (library expressions unverified) is
// obsolete — SetExpression is render-proven incl. effect params (S2 + the
// 2026-06-15 Utf8-ordering fix, see expression-enable-byte-pair.md).
//
// TestTextExpressibleSelector_RoundTrip (no AE): the selector + its Amount
// expression survive WriteAEP → Open.
//
// TestTextExpressibleSelector_AEShipGate_AE20{20,25} (red line 4): SPATIAL
// differential. Two comps, each "ABCDEFGH" + an Opacity-0 animator (empty range) +
// an Expressible Selector. compL's Amount = "textIndex <= 4 ? 100 : 0" selects
// (and thus hides) the LEFT half ABCD, so EFGH renders on the RIGHT; compR's
// "textIndex > 4 ? 100 : 0" hides the RIGHT half, so ABCD renders on the LEFT. The
// gate asserts the ink centroid of compL is well to the RIGHT of compR's — which
// can only happen if AE evaluated the expression to pick the glyphs. An inert /
// dropped expression would leave both comps fully visible (equal centroids).
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

// inkCentroidX returns the mean x of "ink" pixels (luma well above the dark BG)
// in the box, and the ink count.
func inkCentroidX(img image.Image, x0, y0, x1, y1, step int) (float64, int) {
	var sumX, n int
	for y := y0; y <= y1; y += step {
		for x := x0; x <= x1; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			luma := int((r>>8 + g>>8 + b>>8) / 3)
			if luma > 90 { // BG is ~0.06*255≈15; text fill is much brighter
				sumX += x
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0
	}
	return float64(sumX) / float64(n), n
}

// addExpressibleComp adds a comp named compName with a dark BG + a text layer
// "ABCDEFGH" carrying an Opacity-0 animator (empty range) and an Expressible
// Selector whose Amount is amountExpr.
func addExpressibleComp(t *testing.T, p *aep.Project, compName, amountExpr string) {
	t.Helper()
	comp, err := aep.NewComposition(p, compName, 1280, 720, 24, 5)
	if err != nil {
		t.Fatalf("NewComposition %s: %v", compName, err)
	}
	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	rect, _ := bg.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{1600, 1000})
	bgFill, _ := bg.RootGroup().AddFill()
	_ = bgFill.SetColor([4]float64{0.05, 0.05, 0.08, 1})
	_ = bg.Position().SetStaticValue([2]float64{640, 360})

	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("ABCDEFGH"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 0, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}
	if _, err := aep.AddTextExpressibleSelector(tl, amountExpr); err != nil {
		t.Fatalf("AddTextExpressibleSelector: %v", err)
	}
}

func buildExpressibleDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	addExpressibleComp(t, p, "TXEXP_L", "textIndex <= 4 ? 100 : 0") // hide ABCD → ink right
	addExpressibleComp(t, p, "TXEXP_R", "textIndex > 4 ? 100 : 0")  // hide EFGH → ink left
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	for _, c := range rp.Compositions {
		if bg := c.LayerByName("BG"); bg != nil {
			if err := aep.MoveToEnd(bg); err != nil {
				t.Fatalf("MoveToEnd BG: %v", err)
			}
		}
	}
	return rp
}

func TestTextExpressibleSelector_RoundTrip(t *testing.T) {
	rp := buildExpressibleDemo(t, aep.TargetAE2020)
	path := filepath.Join(t.TempDir(), "exp.aep")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(f); err != nil {
		f.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	f.Close()
	re, err := aep.Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	root := parseAEP(t, path)
	if !findShipChunkTdmnName(root, "ADBE Text Expressible Selector") {
		t.Errorf("Expressible Selector dropped after round-trip")
	}
	_ = re
}

func runTextExpressibleGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_expressible_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_expressible.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildExpressibleDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "exp_in.aep")
	resavedAEP := filepath.Join(tempDir, "exp_resaved.aep")
	doneFile := filepath.Join(tempDir, "exp.done")
	pngL := filepath.Join(tempDir, "exp_L.png")
	pngR := filepath.Join(tempDir, "exp_R.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"compL":"TXEXP_L","compR":"TXEXP_R","pngL":%q,"pngR":%q,"fontSize":140,"posX":120,"posY":420}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(pngL), toFwd(pngR))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	clearAEDiskCache(t)
	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 300)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("%s expressible AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s expressible ship gate FAIL:\n%s", ver, body)
	}

	decode := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("%s frame missing (%s): %v", ver, filepath.Base(path), err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("%s decode %s: %v", ver, filepath.Base(path), err)
		}
		return img
	}
	const x0, y0, x1, y1, step = 60, 150, 1220, 680, 2
	imL, imR := decode(pngL), decode(pngR)
	cxL, nL := inkCentroidX(imL, x0, y0, x1, y1, step)
	cxR, nR := inkCentroidX(imR, x0, y0, x1, y1, step)
	dx := cxL - cxR
	if dx < 0 {
		dx = -dx
	}
	t.Logf("%s expressible: L(hide-ABCD→EFGH) centroidX=%.0f ink=%d ; R(hide-EFGH→ABCD) centroidX=%.0f ink=%d ; |Δ|=%.0f", ver, cxL, nL, cxR, nR, dx)
	if nL < 200 || nR < 200 {
		t.Fatalf("%s text not rendered (inkL=%d inkR=%d)", ver, nL, nR)
	}
	// compL shows EFGH, compR shows ABCD — different glyph SETS occupy different
	// columns, so the ink centroids are far apart. (The absolute left/right sign is
	// a center-justify artifact and irrelevant.) An inert / dropped expression would
	// leave BOTH comps fully visible (identical "ABCDEFGH") → near-zero Δ. So a large
	// |Δ| can only mean AE evaluated each expression to pick a different glyph set.
	if dx < 100 {
		t.Errorf("%s expression did not steer selection: L centroidX=%.0f vs R centroidX=%.0f (|Δ|=%.0f, want >100) — both comps look identical, expression inert", ver, cxL, cxR, dx)
	}

	if _, err := aep.Open(resavedAEP); err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
}

func TestTextExpressibleSelector_AEShipGate_AE2020(t *testing.T) {
	runTextExpressibleGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextExpressibleSelector_AEShipGate_AE2025(t *testing.T) {
	runTextExpressibleGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
