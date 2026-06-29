// internal/aep/expr_vocab_shipgate_test.go
//
// AE ship gate for the common expression VOCABULARY beyond the single
// `time*90` case S2 first proved (MG roadmap S2 follow-up,
// specs/2026-06-12-from-scratch-mg-roadmap.md;
// incidents/expression-enable-byte-pair.md coverage-boundary note). The
// byte mechanism (tdb4 @0x77/@0x78 + Utf8 in tdbs) is expression-content
// agnostic, so what this gate adds is AE *evaluating* three structurally
// distinct idioms — and, for loopOut, the new combination of an expression
// layered on top of a KEYFRAMED property:
//
//   - LEAD: amber dot, static Position (480,250). The link target; no
//     expression.
//   - LINK: cyan dot, Position = `thisComp.layer("LEAD").transform.position
//   - [0,250]` → renders at (480,500). Proves cross-layer reference +
//     vector arithmetic resolve.
//   - LOOP: green dot, Position 2 linear keyframes x:300→1500 @ y=750 over
//     0..1s, `loopOut("cycle")`. At t=2.5s the cycle phase is 0.5 so the
//     looped x = 900 (linear midpoint); WITHOUT the loop AE would hold the
//     last keyframe at x=1500. Proves expression-on-keyframed-property.
//   - WIG: magenta dot, Position `wiggle(2,250)` around static (960,950).
//     Procedural/time-varying; JSX asserts valueAtTime moved off the anchor
//     and differs across time. Wiggle is non-deterministic per render so the
//     pixel check only confirms the dot rendered (not dropped).
//   - SLD: yellow dot, a Slider Control effect (value 880) on the layer drives
//     its own Position.x via `[effect(1)(1), 350]` → renders at (880,350).
//     Proves an expression resolving an effect-parameter reference (core MG
//     rig). Reference by BOTH indices — effect(1)(1) — not names: AddEffect
//     names the instance by its match-name "ADBE Slider Control" (not the
//     display "Slider Control"), and the materialized param's display name is
//     not "Slider". A name-based ref (effect("Slider Control")("Slider")) makes
//     AE silently fall back to the static value (a false-green trap).
//
// Render pixel proof (red line 4) is deterministic for LEAD/LINK/LOOP/SLD.
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

// buildExprVocabDemo builds the EXPRVOCAB comp from scratch and attaches the
// three expression idioms after Reopen (expressions need parsed tdbs
// back-refs, same as the S2 gate). Shared by the round-trip and ship-gate
// tests.
func buildExprVocabDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "EXPRVOCAB", 1920, 1080, 30, 4)
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

	addDot := func(name string, color [4]float64) *aep.ShapeLayer {
		dl, err := aep.NewShapeLayer(comp, name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", name, err)
		}
		el, err := dl.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("AddEllipse %s: %v", name, err)
		}
		if err := el.SetSize([2]float64{72, 72}); err != nil {
			t.Fatalf("%s SetSize: %v", name, err)
		}
		f, err := dl.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("AddFill %s: %v", name, err)
		}
		if err := f.SetColor(color); err != nil {
			t.Fatalf("%s SetColor: %v", name, err)
		}
		return dl
	}

	lead := addDot("LEAD", [4]float64{1.0, 0.55, 0.1, 1}) // amber
	if err := lead.Position().SetStaticValue([2]float64{480, 250}); err != nil {
		t.Fatalf("LEAD Position: %v", err)
	}

	link := addDot("LINK", [4]float64{0.25, 0.85, 1.0, 1}) // cyan
	if err := link.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("LINK Position: %v", err)
	}

	loop := addDot("LOOP", [4]float64{0.25, 1.0, 0.5, 1}) // green
	if err := loop.Position().AddKeyframeLinear(0, [2]float64{300, 750}); err != nil {
		t.Fatalf("LOOP kf0: %v", err)
	}
	if err := loop.Position().AddKeyframeLinear(1, [2]float64{1500, 750}); err != nil {
		t.Fatalf("LOOP kf1: %v", err)
	}

	wig := addDot("WIG", [4]float64{1.0, 0.2, 0.78, 1}) // magenta
	if err := wig.Position().SetStaticValue([2]float64{960, 950}); err != nil {
		t.Fatalf("WIG Position: %v", err)
	}

	sld := addDot("SLD", [4]float64{1.0, 0.9, 0.16, 1}) // yellow
	if err := sld.Position().SetStaticValue([2]float64{960, 350}); err != nil {
		t.Fatalf("SLD Position: %v", err)
	}

	// Expressions and AddEffect both need parsed back-refs; do them after Reopen.
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	cc := rp.Compositions[0]

	// SLD: a Slider Control on SLD drives its own Position.x by expression —
	// proves an expression resolving an effect-parameter reference (a core MG
	// rig). Reference the effect by index (effect(1)); AE names the instance by
	// its match-name ("ADBE Slider Control"), so effect("Slider Control") would
	// miss.
	sldLayer := cc.LayerByName("SLD")
	fx, err := aep.AddEffect(sldLayer, aep.EffectSliderControl)
	if err != nil {
		t.Fatalf("AddEffect Slider Control: %v", err)
	}
	if _, err := aep.SetEffectParam(sldLayer, fx, "ADBE Slider Control-0001", 880.0); err != nil {
		t.Fatalf("SetEffectParam slider: %v", err)
	}

	for _, spec := range []struct {
		layer string
		expr  string
	}{
		{"LINK", `thisComp.layer("LEAD").transform.position + [0,250]`},
		{"LOOP", `loopOut("cycle")`},
		{"WIG", `wiggle(2,250)`},
		{"SLD", `[effect(1)(1), 350]`},
	} {
		pos := cc.LayerByName(spec.layer).Position()
		if pos == nil {
			t.Fatalf("%s: no position property", spec.layer)
		}
		if err := pos.SetExpression(spec.expr); err != nil {
			t.Fatalf("%s SetExpression: %v", spec.layer, err)
		}
		if err := pos.SetExpressionEnabled(true); err != nil {
			t.Fatalf("%s SetExpressionEnabled: %v", spec.layer, err)
		}
	}
	if err := aep.MoveToEnd(cc.LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

// countColor returns how many pixels in img are within tol of want.
func countColor(img image.Image, want [3]uint8, tol int) int {
	b := img.Bounds()
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := img.At(x, y).RGBA()
			got := [3]int{int(r >> 8), int(g >> 8), int(bb >> 8)}
			match := true
			for i := range 3 {
				d := got[i] - int(want[i])
				if d < -tol || d > tol {
					match = false
					break
				}
			}
			if match {
				n++
			}
		}
	}
	return n
}

// TestExprVocab_Roundtrip de-risks the Go side without AE: it builds the demo
// and confirms WriteAEP→Open preserves all three expression strings, their
// enabled flags, and LOOP's 2 keyframes. (AE *evaluation* is the ship gate's
// job; this only guards the byte round-trip.)
func TestExprVocab_Roundtrip(t *testing.T) {
	p := buildExprVocabDemo(t, aep.TargetAE2020)

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "exprvocab.aep")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	re, err := aep.Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	cc := re.Compositions[0]
	for _, want := range []struct {
		layer string
		expr  string
	}{
		{"LINK", `thisComp.layer("LEAD").transform.position + [0,250]`},
		{"LOOP", `loopOut("cycle")`},
		{"WIG", `wiggle(2,250)`},
		{"SLD", `[effect(1)(1), 350]`},
	} {
		l := cc.LayerByName(want.layer)
		if l == nil || l.Position() == nil {
			t.Errorf("%s: position stream gone after round-trip", want.layer)
			continue
		}
		pos := l.Position()
		if pos.Expression != want.expr {
			t.Errorf("%s: expr=%q, want %q", want.layer, pos.Expression, want.expr)
		}
		if !pos.ExpressionEnabled {
			t.Errorf("%s: expressionEnabled=false, want true", want.layer)
		}
	}
	loop := cc.LayerByName("LOOP")
	if loop == nil || loop.Position() == nil || len(loop.Position().Keyframes) != 2 {
		n := -1
		if loop != nil && loop.Position() != nil {
			n = len(loop.Position().Keyframes)
		}
		t.Errorf("LOOP position keyframes = %d, want 2 (expression + keyframes combo)", n)
	}
}

func runExprVocabGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/expr_vocab_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_expr_vocab.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildExprVocabDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "expr_vocab_in.aep")
	resavedAEP := filepath.Join(tempDir, "expr_vocab_resaved.aep")
	doneFile := filepath.Join(tempDir, "expr_vocab.done")
	framePNG := filepath.Join(tempDir, "expr_vocab_frame.png")

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
	t.Logf("expr vocab %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("expr vocab %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at t=2.5s.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const tol = 40
	amber := [3]uint8{255, 140, 26}
	cyan := [3]uint8{64, 217, 255}
	green := [3]uint8{64, 255, 128}
	magenta := [3]uint8{255, 51, 199}
	yellow := [3]uint8{255, 230, 41}

	leadX := scanRowForColor(img, 250, amber, tol)
	sldX := scanRowForColor(img, 350, yellow, tol)
	linkX := scanRowForColor(img, 500, cyan, tol)
	loopX := scanRowForColor(img, 750, green, tol)
	magN := countColor(img, magenta, tol)
	t.Logf("%s centres: LEAD x=%d SLD x=%d LINK x=%d LOOP x=%d  WIG px=%d", ver, leadX, sldX, linkX, loopX, magN)

	if leadX < 420 || leadX > 540 {
		t.Errorf("%s LEAD at x=%d, want ≈480 (static anchor)", ver, leadX)
	}
	if linkX < 420 || linkX > 540 {
		t.Errorf("%s LINK at x=%d, want ≈480 (cross-layer ref to LEAD.x)", ver, linkX)
	}
	if loopX < 840 || loopX > 960 {
		t.Errorf("%s LOOP at x=%d, want ≈900 (loopOut cycle phase 0.5); x≈1500 = loop not evaluating", ver, loopX)
	}
	if sldX < 820 || sldX > 940 {
		t.Errorf("%s SLD at x=%d, want ≈880 (Slider Control via expression); x≈960 = effect-param ref not evaluating", ver, sldX)
	}
	if magN < 500 {
		t.Errorf("%s WIG magenta pixels = %d, want >500 (wiggle dot rendered)", ver, magN)
	}

	// Resave proof: expressions + LOOP keyframes survive AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	cc := re.Compositions[0]
	for _, want := range []struct {
		layer string
		expr  string
	}{
		{"LINK", `thisComp.layer("LEAD").transform.position + [0,250]`},
		{"LOOP", `loopOut("cycle")`},
		{"WIG", `wiggle(2,250)`},
		{"SLD", `[effect(1)(1), 350]`},
	} {
		l := cc.LayerByName(want.layer)
		if l == nil || l.Position() == nil {
			t.Errorf("resaved %s: position stream gone (expression dropped on resave)", want.layer)
			continue
		}
		pos := l.Position()
		if pos.Expression != want.expr || !pos.ExpressionEnabled {
			t.Errorf("resaved %s: expr=%q enabled=%v, want %q/true", want.layer, pos.Expression, pos.ExpressionEnabled, want.expr)
		}
	}
	if loop := cc.LayerByName("LOOP"); loop != nil && loop.Position() != nil {
		if n := len(loop.Position().Keyframes); n != 2 {
			t.Errorf("resaved LOOP keyframes = %d, want 2", n)
		}
	}
}

func TestExprVocab_AEShipGate_AE2020(t *testing.T) {
	runExprVocabGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestExprVocab_AEShipGate_AE2025(t *testing.T) {
	runExprVocabGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
